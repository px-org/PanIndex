package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"bytes"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/nic"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
	"time"
)

//var GloablOrgId string
//var GloablDriveId string
//var GloablSpaceId string
//var GloablRootId string
//var GloablProjectId string
//var IsPorject bool = false
//var TeambitionSession nic.Session
var TeambitionSessions = map[string]entity.Teambition{}

//Teambition网盘登录
func TeambitionLogin(accountId, user, password string) string {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
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
	u := jsoniter.Get(resp.Bytes, "user")
	if u == nil || u.Get("_id").ToString() == "" {
		//登录失败
		Teambition.TeambitionSession = TeambitionSession
		TeambitionSessions[accountId] = Teambition
		return "4"
	}
	//2. 获orgId, memberId
	resp, err = TeambitionSession.Get("https://www.teambition.com/api/organizations/personal", nil)
	if err != nil {
		panic(err.Error())
	}
	Teambition.GloablOrgId = jsoniter.Get(resp.Bytes, "_id").ToString()
	memberId := jsoniter.Get(resp.Bytes, "_creatorId").ToString()
	//3.获取rootId、spaceId
	resp, err = TeambitionSession.Get(fmt.Sprintf("https://pan.teambition.com/pan/api/spaces?orgId=%s&memberId=%s", Teambition.GloablOrgId, memberId), nil)
	if err != nil {
		panic(err.Error())
	}
	Teambition.GloablRootId = jsoniter.Get(resp.Bytes, 0, "rootId").ToString()
	Teambition.GloablSpaceId = jsoniter.Get(resp.Bytes, 0, "spaceId").ToString()
	//4.获取driverId
	resp, err = TeambitionSession.Get(fmt.Sprintf("https://pan.teambition.com/pan/api/orgs/%s?orgId=%s", Teambition.GloablOrgId, Teambition.GloablOrgId), nil)
	if err != nil {
		panic(err.Error())
	}
	Teambition.GloablDriveId = jsoniter.Get(resp.Bytes, "data").Get("driveId").ToString()
	Teambition.TeambitionSession = TeambitionSession
	TeambitionSessions[accountId] = Teambition
	return "success"
}

//Teambition网盘登录
func TeambitionUSLogin(accountId, user, password string) string {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	//0.登录-获取token
	resp, err := TeambitionSession.Get("https://us-account.teambition.com/login/password", nil)
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
		resp, err = TeambitionSession.Post("https://us-account.teambition.com/api/login/email", nic.H{
			JSON:          param,
			AllowRedirect: false,
		})
		if err != nil {
			panic(err.Error())
		}
	} else {
		//手机号登录
		param["phone"] = user
		resp, err = TeambitionSession.Post("https://us-account.teambition.com/api/login/phone", nic.H{
			JSON: param,
		})
		if err != nil {
			panic(err.Error())
		}
	}
	u := jsoniter.Get(resp.Bytes, "user")
	if u != nil && u.Get("_id").ToString() != "" {
		//登录成功
		Teambition.TeambitionSession = TeambitionSession
		TeambitionSessions[accountId] = Teambition
		return "success"
	}
	return ""
}

func ProjectIdCheck(server, accountId, rootId string) string {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	resp, err := TeambitionSession.Get(fmt.Sprintf("https://%s.teambition.com/api/projects/%s", server, rootId), nil)
	if err != nil {
		panic(err.Error())
	}
	if resp.StatusCode == 404 {
		//项目id查询失败，可能是个人文件
		Teambition.IsPorject = false
		return ""
	}
	Teambition.IsPorject = true
	Teambition.GloablRootId = jsoniter.Get(resp.Bytes, "_rootCollectionId").ToString()
	Teambition.GloablProjectId = rootId
	Teambition.TeambitionSession = TeambitionSession
	TeambitionSessions[accountId] = Teambition
	return Teambition.GloablRootId
}

//获取个人文件列表
func TeambitionGetFiles(accountId, rootId, fileId, p string, hide, hasPwd int, syncChild bool) {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
	if rootId == "" {
		//如果没有设置rootId,这里使用全局的rootId
		rootId = Teambition.GloablRootId
		fileId = Teambition.GloablRootId
	}
	defer func() {
		if p := recover(); p != nil {
			log.Warningln(p)
		}
	}()
	limit := 100
	nextMarker := ""
	for {
		url := fmt.Sprintf("https://pan.teambition.com/pan/api/nodes?orgId=%s&from=%s&limit=%d&orderBy=updated_at&orderDirection=DESC&driveId=%s&parentId=%s", Teambition.GloablOrgId, nextMarker, limit, Teambition.GloablDriveId, fileId)
		resp, err := TeambitionSession.Get(url, nil)
		if err != nil {
			panic(err.Error())
		}
		byteFiles := []byte(resp.Text)
		d := jsoniter.Get(byteFiles, "data")
		nextMarker = jsoniter.Get(byteFiles, "nextMarker").ToString()
		var m []map[string]interface{}
		json.Unmarshal([]byte(d.ToString()), &m)
		for _, item := range m {
			fn := entity.FileNode{}
			fn.AccountId = accountId
			fn.FileId = item["nodeId"].(string)
			fn.FileName = item["name"].(string)
			fn.FileIdDigest = ""
			fn.CreateTime = UTCTimeFormat(item["created"].(string))
			fn.LastOpTime = UTCTimeFormat(item["updated"].(string))
			fn.Delete = 1
			kind := item["kind"].(string)
			if kind == "file" {
				if item["ext"] == nil {
					fn.FileType = ""
				} else {
					fn.FileType = item["ext"].(string)
				}
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
			//天翼云网盘独有，这里随便定义一个
			fn.IsStarred = true
			fn.ParentId = item["parentId"].(string)
			fn.Hide = 0
			fn.HasPwd = 0
			if hide == 1 {
				fn.Hide = hide
			} else {
				if config.GloablConfig.HideFileId != "" {
					listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && listSTring[i] == fn.FileId {
						fn.Hide = 1
					}
				}
			}
			if hasPwd == 1 {
				fn.HasPwd = hasPwd
			} else {
				if config.GloablConfig.PwdDirId != "" {
					listSTring := strings.Split(config.GloablConfig.PwdDirId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && strings.Split(listSTring[i], ":")[0] == fn.FileId {
						fn.HasPwd = 1
					}
				}
			}
			fn.ParentPath = p
			if p == "/" {
				fn.Path = p + fn.FileName
			} else {
				fn.Path = p + "/" + fn.FileName
			}
			if fn.IsFolder == true {
				if syncChild {
					TeambitionGetFiles(accountId, rootId, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
				}
			}
			fn.Id = uuid.NewV4().String()
			fn.CacheTime = time.Now().UnixNano()
			model.SqliteDb.Create(fn)
		}
		if nextMarker == "" {
			break
		}
	}
}

func TeambitionGetProjectFiles(server, accountId, rootId, p string, hide, hasPwd int, syncChild bool) {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
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
		url := fmt.Sprintf("https://%s.teambition.com/api/collections?_parentId=%s&_projectId=%s&order=updatedDesc&count=%d&page=%d", server, rootId, Teambition.GloablProjectId, limit, pageNum)
		resp, err := TeambitionSession.Get(url, nil)
		if err != nil {
			panic(err.Error())
		}
		json.Unmarshal(resp.Bytes, &m)
		url = fmt.Sprintf("https://%s.teambition.com/api/works?_parentId=%s&_projectId=%s&order=updatedDesc&count=%d&page=%d", server, rootId, Teambition.GloablProjectId, limit, pageNum)
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
			fn.AccountId = accountId
			fn.FileId = item["_id"].(string)
			fn.FileIdDigest = ""
			fn.CreateTime = UTCTimeFormat(item["created"].(string))
			fn.LastOpTime = UTCTimeFormat(item["updated"].(string))
			fn.Delete = 1
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
			//天翼云网盘独有，这里随便定义一个
			fn.IsStarred = true
			fn.ParentId = item["_parentId"].(string)
			fn.Hide = 0
			fn.HasPwd = 0
			if hide == 1 {
				fn.Hide = hide
			} else {
				if config.GloablConfig.HideFileId != "" {
					listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && listSTring[i] == fn.FileId {
						fn.Hide = 1
					}
				}
			}
			if hasPwd == 1 {
				fn.HasPwd = hasPwd
			} else {
				if config.GloablConfig.PwdDirId != "" {
					listSTring := strings.Split(config.GloablConfig.PwdDirId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && strings.Split(listSTring[i], ":")[0] == fn.FileId {
						fn.HasPwd = 1
					}
				}
			}
			fn.ParentPath = p
			if p == "/" {
				fn.Path = p + fn.FileName
			} else {
				fn.Path = p + "/" + fn.FileName
			}
			if syncChild {
				TeambitionGetProjectFiles(server, accountId, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
			}
			if fn.FileName != "" {
				fn.Id = uuid.NewV4().String()
				fn.CacheTime = time.Now().UnixNano()
				model.SqliteDb.Create(fn)
			}
		}
		pageNum++
	}
}

func GetTeambitionDownUrl(accountId, nodeId string) string {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
	url := fmt.Sprintf("https://pan.teambition.com/pan/api/nodes/%s?orgId=%s&driveId=%s", nodeId, Teambition.GloablOrgId, Teambition.GloablDriveId)
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
func GetTeambitionProDownUrl(server, accountId, nodeId string) string {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
	url := fmt.Sprintf("https://%s.teambition.com/api/works/%s", server, nodeId)
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

func TeambitionUpload(accountId, parentId string, files []*multipart.FileHeader) bool {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
		fs := []nic.KV{nic.KV{
			"driveId":     Teambition.GloablDriveId,
			"chunkCount":  1,
			"name":        file.Filename,
			"ccpParentId": parentId,
			"contentType": "",
			"size":        file.Size,
			"type":        "file",
		}}
		resp, _ := TeambitionSession.Post("https://pan.teambition.com/pan/api/nodes/file", nic.H{
			JSON: nic.KV{
				"orgId":         Teambition.GloablOrgId,
				"spaceId":       Teambition.GloablSpaceId,
				"parentId":      parentId,
				"checkNameMode": "autoRename",
				"infos":         fs,
			},
		})
		nodeId := jsoniter.Get(resp.Bytes, 0).Get("nodeId").ToString()
		uploadId := jsoniter.Get(resp.Bytes, 0).Get("uploadId").ToString()
		resp, _ = TeambitionSession.Post(fmt.Sprintf("https://pan.teambition.com/pan/api/nodes/%s/uploadUrl", nodeId), nic.H{
			JSON: nic.KV{
				"orgId":           Teambition.GloablOrgId,
				"driveId":         Teambition.GloablDriveId,
				"uploadId":        uploadId,
				"startPartNumber": 1,
				"endPartNumber":   1,
			},
		})
		fileId := jsoniter.Get(resp.Bytes, "fileId").ToString()
		partInfoListString := jsoniter.Get(resp.Bytes, "partInfoList").ToString()
		partInfoList := []entity.PartInfo{}
		jsoniter.UnmarshalFromString(partInfoListString, &partInfoList)
		log.Debugf("文件分片数：%d", len(partInfoList))
		for _, partInfo := range partInfoList {
			fileContent, _ := file.Open()
			byteContent, _ := ioutil.ReadAll(fileContent)
			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPut, partInfo.UploadUrl, bytes.NewBuffer(byteContent))
			if err != nil {
				log.Error("上传失败")
				return false
			}
			client.Do(req)
		}
		resp, _ = TeambitionSession.Post("https://pan.teambition.com/pan/api/nodes/complete", nic.H{
			JSON: nic.KV{
				"orgId":     Teambition.GloablOrgId,
				"driveId":   Teambition.GloablDriveId,
				"uploadId":  uploadId,
				"nodeId":    nodeId,
				"ccpFileId": fileId,
			},
		})
		log.Debugf("上传接口返回：%s", resp.Text)
		log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
	}
	return true
}

func TeambitionProUpload(server, accountId, parentId string, files []*multipart.FileHeader) bool {
	Teambition := TeambitionSessions[accountId]
	TeambitionSession := Teambition.TeambitionSession
	prefix := ""
	if server == "us" {
		prefix = "us"
	} else {
		prefix = "www"
	}
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
		resp, _ := TeambitionSession.Get(fmt.Sprintf("https://%s.teambition.com/projects", prefix), nil)
		//0.准备文件
		fileContent, _ := file.Open()
		byteContent, _ := ioutil.ReadAll(fileContent)
		//1.获取jwt
		jwt := GetCurBetweenStr(resp.Text, "strikerAuth&quot;:&quot;", "&quot;,&quot;phoneForLogin")
		//2.上传文件
		if server == "us" {
			prefix = "us-"
		} else {
			prefix = ""
		}
		resp, _ = nic.Post(fmt.Sprintf("https://%stcs.teambition.net/upload", prefix), nic.H{
			Files: nic.KV{
				"file": nic.File(
					file.Filename, byteContent),
			},
			Headers: nic.KV{
				"Authorization": jwt,
			},
		})
		fmt.Println(resp.Text)
		fileKey := jsoniter.Get(resp.Bytes, "fileKey").ToString()
		fileName := jsoniter.Get(resp.Bytes, "fileName").ToString()
		fileType := jsoniter.Get(resp.Bytes, "fileType").ToString()
		fileSize := jsoniter.Get(resp.Bytes, "fileSize").ToInt64()
		fileCategory := jsoniter.Get(resp.Bytes, "fileCategory").ToString()
		//imageWidth := jsoniter.Get(resp.Bytes, "imageWidth").ToString()
		//imageHeight := jsoniter.Get(resp.Bytes, "imageHeight").ToString()
		//3.完成上传
		if server == "us" {
			prefix = "us"
		} else {
			prefix = "www"
		}
		resp, _ = TeambitionSession.Post(fmt.Sprintf("https://%s.teambition.com/api/works", prefix), nic.H{
			JSON: nic.KV{
				"works": []nic.KV{nic.KV{
					"fileKey":      fileKey,
					"fileName":     fileName,
					"fileType":     fileType,
					"fileSize":     fileSize,
					"fileCategory": fileCategory,
					/*"imageWidth":   imageWidth,
					"imageHeight":  imageHeight,*/
					"source":    "tcs",
					"visible":   "members",
					"_parentId": parentId,
				}},
				"_parentId": parentId,
			},
		})
		log.Debugf("上传接口返回：%s", resp.Text)
		log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
	}
	return true
}
func TeambitionIsLogin(accountId string, isUs bool) bool {
	if _, ok := TeambitionSessions[accountId]; ok {
		TeambitionSession := TeambitionSessions[accountId].TeambitionSession
		d := ""
		if isUs {
			d = "us-"
		}
		resp, _ := TeambitionSession.Get(fmt.Sprintf("https://%saccount.teambition.com/api/account", d), nil)
		if resp.StatusCode == 401 {
			return false
		} else if resp.StatusCode == 200 {
			id := jsoniter.Get(resp.Bytes, "_id").ToString()
			if id != "" {
				return true
			}
		}
	}
	return false
}
func UTCTimeFormat(timeStr string) string {
	t, _ := time.Parse(time.RFC3339, timeStr)
	timeUint := t.In(time.Local).Unix()
	return time.Unix(timeUint, 0).Format("2006-01-02 15:04:05")
}
