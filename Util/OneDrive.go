package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/nic"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"mime/multipart"
	"sort"
	"strings"
	"time"
)

var OneDrives = map[string]entity.OneDriveAuthInfo{}

func OneDriveRefreshToken(account entity.Account) string {
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	resp, _ := nic.Post("https://login.microsoftonline.com/common/oauth2/v2.0/token", nic.H{
		Headers: nic.KV{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Data: nic.KV{
			"client_id":     account.User,
			"redirect_uri":  account.RedirectUri,
			"client_secret": account.Password,
			"refresh_token": account.RefreshToken,
			"grant_type":    "refresh_token",
		},
	})
	var auth entity.OneDriveAuthInfo
	err := jsoniter.Unmarshal(resp.Bytes, &auth)
	if err != nil {
		panic(err.Error())
		return ""
	}
	OneDrives[account.Id] = auth
	return auth.RefreshToken
}
func BuildODRequestUrl(path, query string) string {
	if path != "" && path != "/" {
		path = fmt.Sprintf(":/%s:/", path)
	}
	return "https://graph.microsoft.com/v1.0" + "/me/drive/root" + path + query
}
func OndriveGetFiles(url, accountId, fileId, p string) {
	od := OneDrives[accountId]
	auth := od.TokenType + " " + od.AccessToken
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if url == "" {
		url = BuildODRequestUrl(fileId, "children?select=id,name,size,folder,@microsoft.graph.downloadUrl,lastModifiedDateTime,file")
	}
	//limit := 100
	//nextMarker := ""
	resp, err := nic.Get(url, nic.H{
		Headers: nic.KV{
			"Authorization": auth,
			"Host":          "graph.microsoft.com",
		},
	})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(resp.Text)
	byteFiles := []byte(resp.Text)
	d := jsoniter.Get(byteFiles, "value")
	//nextMarker = jsoniter.Get(byteFiles, "next_marker").ToString()
	var m []map[string]interface{}
	json.Unmarshal([]byte(d.ToString()), &m)
	for _, item := range m {
		fn := entity.FileNode{}
		fn.AccountId = accountId
		if fileId == "/" {
			fn.FileId = fileId + item["name"].(string)
		} else {
			fn.FileId = fileId + "/" + item["name"].(string)
		}
		fn.FileName = item["name"].(string)
		fn.FileIdDigest = ""
		fn.CreateTime = ""
		fn.LastOpTime = UTCTimeFormat(item["lastModifiedDateTime"].(string))
		fn.Delete = 1
		if item["folder"] == nil {
			fn.FileType = GetFileType(fn.FileName)
			fn.IsFolder = false
			fn.FileSize = int64(item["size"].(float64))
			fn.SizeFmt = FormatFileSize(fn.FileSize)
			category := GetCategory(item["file"].(map[string]interface{})["mimeType"].(string))
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
			fn.DownloadUrl = item["@microsoft.graph.downloadUrl"].(string)
		} else {
			fn.FileType = ""
			fn.IsFolder = true
			fn.FileSize = 0
			fn.SizeFmt = "-"
			fn.MediaType = 0
			fn.DownloadUrl = ""
		}
		fn.IsStarred = false
		fn.ParentId = fileId
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
		if p == "/" {
			fn.Path = p + fn.FileName
		} else {
			fn.Path = p + "/" + fn.FileName
		}
		if fn.IsFolder == true {
			OndriveGetFiles("", accountId, fn.FileId, fn.Path)
		}
		fn.Id = uuid.NewV4().String()
		model.SqliteDb.Create(fn)
	}
	nextLink := jsoniter.Get(byteFiles, "@odata").Get("nextLink").ToString()
	if nextLink != "" {
		OndriveGetFiles(url, accountId, fileId, p)
	}
}
func GetFileType(name string) string {
	arr := strings.Split(name, ".")
	if len(arr) > 1 {
		return arr[len(arr)-1]
	} else {
		return ""
	}
}
func GetCategory(mimeType string) string {
	arr := strings.Split(mimeType, "/")
	return arr[0]
}
func GetOneDriveDownloadUrl(accountId, fileId string) string {
	od := OneDrives[accountId]
	auth := od.TokenType + " " + od.AccessToken
	resp, _ := nic.Get("https://graph.microsoft.com/v1.0/me/drive/root:"+fileId, nic.H{
		Headers: nic.KV{
			"Authorization": auth,
			"Host":          "graph.microsoft.com",
		},
	})
	return jsoniter.Get(resp.Bytes, "@microsoft.graph.downloadUrl").ToString()
}
func OneDriveUpload(accountId, parentId string, files []*multipart.FileHeader) bool {
	//od := OneDrives[accountId]
	//auth := od.TokenType + " " + od.AccessToken
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
		log.Debugf("上传接口返回：%s", "1")
		log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
	}
	return false
}
func OneExchangeToken(clientId, redirectUri, clientSecret, code string) string {
	resp, err := nic.Post("https://login.microsoftonline.com/common/oauth2/v2.0/token", nic.H{
		Headers: nic.KV{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Data: nic.KV{
			"client_id":     clientId,
			"redirect_uri":  redirectUri,
			"client_secret": clientSecret,
			"code":          code,
			"grant_type":    "authorization_code",
		},
	})
	if err != nil {
		log.Error(err)
		return ""
	}
	return resp.Text
}
func OneGetRefreshToken(clientId, redirectUri, clientSecret, refreshToken string) string {
	resp, err := nic.Post("https://login.microsoftonline.com/common/oauth2/v2.0/token", nic.H{
		Headers: nic.KV{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Data: nic.KV{
			"client_id":     clientId,
			"redirect_uri":  redirectUri,
			"client_secret": clientSecret,
			"refresh_token": refreshToken,
			"grant_type":    "refresh_token",
		},
	})
	if err != nil {
		log.Error(err)
		return ""
	}
	return resp.Text
}
