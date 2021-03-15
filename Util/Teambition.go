package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"encoding/json"
	"fmt"
	"github.com/eddieivan01/nic"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

var GloablOrgId string
var GloablDriveId string
var GloablSpaceId string
var GloablRootId string
var GloablProjectId string
var IsPorject bool = false
var TeambitionSession nic.Session

//Teambition网盘登录
func TeambitionLogin(user, password string) string {
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	//0.登录-获取token
	resp, err := TeambitionSession.Get("https://account.teambition.com/login/password", nil)
	if err != nil {
		panic(err.Error())
	}
	token := GetBetweenStr(resp.Text, "TOKEN\":\"", "\"")
	clientId := GetBetweenStr(resp.Text, "CLIENT_ID\":\"", "\"")
	param := nic.KV{
		"client_id":     clientId,
		"token":         token,
		"password":      password,
		"response_type": "session",
	}
	//1.登录-用户名密码登录,获取cookie
	if strings.Contains(user, "@") {
		//邮箱登录
		param["email"] = user
		resp, err = TeambitionSession.Post("https://account.teambition.com/api/login/email", nic.H{
			JSON:          param,
			AllowRedirect: false,
		})
		if err != nil {
			panic(err.Error())
		}
	} else {
		//手机号登录
		param["phone"] = user
		resp, err = TeambitionSession.Post("https://account.teambition.com/api/login/phone", nic.H{
			JSON: param,
		})
		if err != nil {
			panic(err.Error())
		}
	}
	//2. 获orgId, memberId
	resp, err = TeambitionSession.Get("https://www.teambition.com/api/organizations/personal", nil)
	if err != nil {
		panic(err.Error())
	}
	GloablOrgId = jsoniter.Get(resp.Bytes, "_id").ToString()
	memberId := jsoniter.Get(resp.Bytes, "_creatorId").ToString()
	//3.获取rootId、spaceId
	resp, err = TeambitionSession.Get(fmt.Sprintf("https://pan.teambition.com/pan/api/spaces?orgId=%s&memberId=%s", GloablOrgId, memberId), nil)
	if err != nil {
		panic(err.Error())
	}
	GloablRootId = jsoniter.Get(resp.Bytes, 0, "rootId").ToString()
	GloablSpaceId = jsoniter.Get(resp.Bytes, 0, "spaceId").ToString()
	//4.获取driverId
	resp, err = TeambitionSession.Get(fmt.Sprintf("https://pan.teambition.com/pan/api/orgs/%s?orgId=%s", GloablOrgId, GloablOrgId), nil)
	if err != nil {
		panic(err.Error())
	}
	GloablDriveId = jsoniter.Get(resp.Bytes, "data.driveId").ToString()
	return "success"
}

func ProjectIdCheck(rootId string) string {
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	resp, err := TeambitionSession.Get("https://www.teambition.com/api/projects/"+rootId, nil)
	if err != nil {
		panic(err.Error())
	}
	if resp.StatusCode == 404 {
		//项目id查询失败，可能是个人文件
		IsPorject = false
		return ""
	}
	IsPorject = true
	GloablRootId = jsoniter.Get(resp.Bytes, "_rootCollectionId").ToString()
	GloablProjectId = rootId
	return GloablRootId
}

//获取个人文件列表
func TeambitionGetFiles(rootId, fileId, p string) {
	if rootId == "" {
		//如果没有设置rootId,这里使用全局的rootId
		rootId = GloablRootId
		fileId = GloablRootId
	}
	defer func() {
		if p := recover(); p != nil {
			log.Warningln(p)
		}
	}()
	limit := 100
	pageNum := 0
	for {
		offset := pageNum * limit
		url := fmt.Sprintf("https://pan.teambition.com/pan/api/nodes?orgId=%s&offset=%d&limit=%d&orderBy=updateTime&orderDirection=desc&driveId=%s&spaceId=%s&parentId=%s", GloablOrgId, offset, limit, GloablDriveId, GloablSpaceId, fileId)
		resp, err := TeambitionSession.Get(url, nil)
		if err != nil {
			panic(err.Error())
		}
		byteFiles := []byte(resp.Text)
		d := jsoniter.Get(byteFiles, "data")
		nextMarker := jsoniter.Get(byteFiles, "nextMarker").ToString()
		var m []map[string]interface{}
		json.Unmarshal([]byte(d.ToString()), &m)
		for _, item := range m {
			fn := entity.FileNode{}
			fn.FileId = item["nodeId"].(string)
			fn.FileName = item["name"].(string)
			fn.FileIdDigest = ""
			fn.CreateTime = UTCTimeFormat(item["created"].(string))
			fn.LastOpTime = UTCTimeFormat(item["updated"].(string))
			kind := item["kind"].(string)
			if kind == "file" {
				fn.FileType = item["ext"].(string)
				fn.IsFolder = false
				fn.FileSize = int64(item["size"].(float64))
				fn.SizeFmt = FormatFileSize(fn.FileSize)
				category := item["category"].(string)
				if category == "image" {
					//图片
					fn.MediaType = 1
				} else if category == "doc" {
					//文本
					fn.MediaType = 4
				} else if category == "video" {
					//视频
					fn.MediaType = 3
				} else if category == "audio" {
					//音频
					fn.MediaType = 2
				} else {
					//其他类型
					fn.MediaType = 0
				}
				fn.DownloadUrl = item["downloadUrl"].(string)
			} else {
				fn.FileType = ""
				fn.IsFolder = true
				fn.FileSize = 0
				fn.SizeFmt = "-"
				fn.MediaType = 0
				fn.DownloadUrl = ""
			}
			fn.Delete = 0
			//天翼云网盘独有，这里随便定义一个
			fn.IsStarred = true
			fn.ParentId = item["parentId"].(string)
			fn.Hide = 0
			if config.GloablConfig.HideFileId != "" {
				listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
				sort.Strings(listSTring)
				i := sort.SearchStrings(listSTring, fn.FileId)
				if i < len(listSTring) && listSTring[i] == fn.FileId {
					fn.Hide = 1
				}
			}
			fn.ParentPath = p
			if fn.ParentId == rootId {
				fn.Path = p + fn.FileName
			} else {
				fn.Path = p + "/" + fn.FileName
			}
			if fn.IsFolder == true {
				TeambitionGetFiles(rootId, fn.FileId, fn.Path)
			}
			model.SqliteDb.Save(fn)
		}
		if nextMarker != "" {
			pageNum++
		} else {
			break
		}
	}
}

func TeambitionGetProjectFiles(rootId, p string) {
	defer func() {
		if p := recover(); p != nil {
			log.Warningln(p)
		}
	}()
	limit := 100
	pageNum := 1
	for {
		var m []map[string]interface{}
		var n []map[string]interface{}
		//先查询目录
		url := fmt.Sprintf("https://www.teambition.com/api/collections?_parentId=%s&_projectId=%s&order=updatedDesc&count=%d&page=%d", rootId, GloablProjectId, limit, pageNum)
		resp, err := TeambitionSession.Get(url, nil)
		if err != nil {
			panic(err.Error())
		}
		json.Unmarshal(resp.Bytes, &m)
		url = fmt.Sprintf("https://www.teambition.com/api/works?_parentId=%s&_projectId=%s&order=updatedDesc&count=%d&page=%d", rootId, GloablProjectId, limit, pageNum)
		resp, err = TeambitionSession.Get(url, nil)
		if err != nil {
			panic(err.Error())
		}
		json.Unmarshal(resp.Bytes, &n)
		m = append(m, n...)
		//再查询文件
		if len(m) == 0 {
			break
		}
		for _, item := range m {
			fn := entity.FileNode{}
			fn.FileId = item["_id"].(string)
			fn.FileIdDigest = ""
			fn.CreateTime = UTCTimeFormat(item["created"].(string))
			fn.LastOpTime = UTCTimeFormat(item["updated"].(string))
			if item["title"] == nil {
				fn.FileName = item["fileName"].(string)
				fn.FileType = item["fileType"].(string)
				fn.IsFolder = false
				fn.FileSize = int64(item["fileSize"].(float64))
				fn.SizeFmt = FormatFileSize(fn.FileSize)
				category := item["fileCategory"].(string)
				if category == "image" {
					//图片
					fn.MediaType = 1
				} else if category == "doc" {
					//文本
					fn.MediaType = 4
				} else if category == "video" {
					//视频
					fn.MediaType = 3
				} else if category == "audio" {
					//音频
					fn.MediaType = 2
				} else {
					//其他类型
					fn.MediaType = 0
				}
				fn.DownloadUrl = item["downloadUrl"].(string)
			} else {
				fn.FileName = item["title"].(string)
				fn.FileType = ""
				fn.IsFolder = true
				fn.FileSize = 0
				fn.SizeFmt = "-"
				fn.MediaType = 0
				fn.DownloadUrl = ""
			}
			fn.Delete = 0
			//天翼云网盘独有，这里随便定义一个
			fn.IsStarred = true
			fn.ParentId = item["_parentId"].(string)
			fn.Hide = 0
			if config.GloablConfig.HideFileId != "" {
				listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
				sort.Strings(listSTring)
				i := sort.SearchStrings(listSTring, fn.FileId)
				if i < len(listSTring) && listSTring[i] == fn.FileId {
					fn.Hide = 1
				}
			}
			fn.ParentPath = p
			if fn.ParentId == rootId {
				fn.Path = p + fn.FileName
			} else {
				fn.Path = p + "/" + fn.FileName
			}
			TeambitionGetProjectFiles(fn.FileId, fn.Path)
			if fn.FileName != "" {
				model.SqliteDb.Save(fn)
			}
		}
		pageNum++
	}
}

func GetTeambitionDownUrl(nodeId string) string {
	url := fmt.Sprintf("https://pan.teambition.com/pan/api/nodes/%s?orgId=%s&driveId=%s", nodeId, GloablOrgId, GloablDriveId)
	resp, err := TeambitionSession.Get(url, nil)
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if err != nil {
		panic(err.Error())
	}
	downUrl := jsoniter.Get(resp.Bytes, "downloadUrl").ToString()
	if downUrl == "" {
		log.Warningln("Teambition盘下载地址获取失败")
	}
	return downUrl
}
func GetTeambitionProDownUrl(nodeId string) string {
	url := fmt.Sprintf("https://www.teambition.com/api/works/%s", nodeId)
	resp, err := TeambitionSession.Get(url, nil)
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if err != nil {
		panic(err.Error())
	}
	downUrl := jsoniter.Get(resp.Bytes, "downloadUrl").ToString()

	if downUrl == "" {
		log.Warningln("Teambition盘下载地址获取失败")
	}
	rs, _ := nic.Get(downUrl, nic.H{
		AllowRedirect: false,
	})
	return rs.Header.Get("Location")
}

func UTCTimeFormat(timeStr string) string {
	t, _ := time.Parse(time.RFC3339, timeStr)
	timeUint := t.In(time.Local).Unix()
	return time.Unix(timeUint, 0).Format("2006-01-02 15:04:05")
}
