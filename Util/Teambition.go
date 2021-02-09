package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"fmt"
	"github.com/eddieivan01/nic"
	jsoniter "github.com/json-iterator/go"
	"log"
	"sort"
	"strings"
)

var GloablOrgId string
var GloablDriveId string
var GloablSpaceId string
var GloablRootId string
var TeambitionSession nic.Session

//Teambition网盘登录
func TeambitionLogin(user, password string) {
	defer func() {
		if p := recover(); p != nil {
			log.Println(p)
		}
	}()
	//0.登录-获取token
	resp, err := TeambitionSession.Get("https://account.teambition.com/login/password", nil)
	if err != nil {
		panic(err.Error())
	}
	token := GetBetweenStr(resp.Text, "TOKEN\":\"", "\"")
	clientId := GetBetweenStr(resp.Text, "CLIENT_ID\":\"", "\"")
	fmt.Println("t: " + token)
	fmt.Println("c: " + clientId)
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
	//5.获取文件列表
	resp, err = TeambitionSession.Get(fmt.Sprintf("https://pan.teambition.com/pan/api/nodes?orgId=%s&offset=0&limit=100&orderBy=updateTime&orderDirection=desc&driveId=%s&spaceId=%s&parentId=%s", GloablOrgId, GloablDriveId, GloablSpaceId, GloablRootId), nil)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(resp.Text)
}

//获取文件列表
func TeambitionGetFiles(rootId, fileId string) {
	defer func() {
		if p := recover(); p != nil {
			log.Println(p)
		}
	}()
	pageNum := 1
	for {
		url := fmt.Sprintf(fmt.Sprintf("https://pan.teambition.com/pan/api/nodes?orgId=%s&offset=0&limit=100&orderBy=updateTime&orderDirection=desc&driveId=%s&spaceId=%s&parentId=%s", GloablOrgId, GloablDriveId, GloablSpaceId, fileId), nil)
		resp, err := TeambitionSession.Get(url, nil)
		if err != nil {
			panic(err.Error())
		}
		byteFiles := []byte(resp.Text)
		totalCount := jsoniter.Get(byteFiles, "recordCount").ToInt()
		d := jsoniter.Get(byteFiles, "data")
		paths := jsoniter.Get(byteFiles, "path")
		ps := []entity.Paths{}
		err = jsoniter.Unmarshal([]byte(paths.ToString()), &ps)
		p := ""
		flag := false
		if err == nil {
			for _, item := range ps {
				if flag == true && item.FileId != rootId {
					if strings.HasSuffix(p, "/") != true {
						p += "/" + item.FileName
					} else {
						p += item.FileName
					}
				}
				if item.FileId == rootId {
					flag = true
				}
				if flag == true && item.FileId == rootId {
					p += "/"
				}
			}
		}
		if d != nil {
			m := []entity.FileNode{}
			err = jsoniter.Unmarshal([]byte(d.ToString()), &m)
			if err == nil {
				for _, item := range m {
					if p == "/" {
						item.Path = "/" + item.FileName
					} else {
						item.Path = p + "/" + item.FileName
					}
					item.ParentPath = p
					item.SizeFmt = FormatFileSize(item.FileSize)
					if item.IsFolder == true {
						Cloud189GetFiles(rootId, item.FileId)
					} else {
						//如果是文件，解析下载直链
						/*dRedirectRep, _ := CLoud189Session.Get("https://cloud.189.cn/downloadFile.action?fileStr="+item.FileIdDigest+"&downloadType=1", nic.H{
							AllowRedirect: false,
						})
						redirectUrl := dRedirectRep.Header.Get("Location")
						dRedirectRep, _ = CLoud189Session.Get(redirectUrl, nic.H{
							AllowRedirect: false,
						})
						item.DownloadUrl = dRedirectRep.Header.Get("Location")*/
					}
					item.Delete = 0
					if config.GloablConfig.HideFileId != "" {
						listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
						sort.Strings(listSTring)
						i := sort.SearchStrings(listSTring, item.FileId)
						if i < len(listSTring) && listSTring[i] == item.FileId {
							item.Hide = 1
						}
					}
					log.Println(item.FileName)
					model.SqliteDb.Save(item)
				}
			}
		}
		if pageNum*100 < totalCount {
			pageNum++
		} else {
			break
		}
	}
}
