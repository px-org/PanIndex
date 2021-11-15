package Util

import (
	"PanIndex/entity"
	"PanIndex/model"
	"bytes"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/nic"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var GoogleDrives = map[string]entity.GoogleDriveAuthInfo{}

const (
	Proxy    = "socks5://127.0.0.1:1089"
	PartSize = 1 * 1024 * 1024
)

func GDfreshToken(account entity.Account) string {
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	resp, _ := nic.Post("https://oauth2.googleapis.com/token", nic.H{
		Proxy: Proxy,
		Headers: nic.KV{
			"content-type": "application/x-www-form-urlencoded",
		},
		Data: nic.KV{
			"client_id":     account.User,
			"client_secret": account.Password,
			"refresh_token": account.RefreshToken,
			"redirect_uri":  account.RedirectUri,
			"grant_type":    "refresh_token",
		},
	})
	var auth entity.GoogleDriveAuthInfo
	err := jsoniter.Unmarshal(resp.Bytes, &auth)
	if err != nil {
		panic(err.Error())
		return ""
	}
	GoogleDrives[account.Id] = auth
	return account.RefreshToken
}

//api: https://developers.google.com/drive/api/v3/reference/files/list
func GDGetFiles(accountId, fileId, p string, hide, hasPwd int, syncChild bool) {
	gd := GoogleDrives[accountId]
	auth := gd.TokenType + " " + gd.AccessToken
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	pageToken := ""
	for {
		fns := []entity.FileNode{}
		url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files?pageSize=%d&supportsAllDrives=true&includeItemsFromAllDrives=true&fields=%s&q=trashed = false and '%s' in parents&orderBy=%s&pageToken=%s",
			2, "nextPageToken, files(id,name,mimeType,parents,size,fileExtension,thumbnailLink,modifiedTime,createdTime,md5Checksum)", fileId, "folder,modifiedTime asc,name", pageToken)
		resp, _ := nic.Get(url, nic.H{
			Proxy: Proxy,
			Headers: nic.KV{
				"Authorization": auth,
			},
		})
		pageToken = jsoniter.Get(resp.Bytes, "nextPageToken").ToString()
		files := jsoniter.Get(resp.Bytes, "files").ToString()
		var fs []map[string]interface{}
		json.Unmarshal([]byte(files), &fs)
		for _, item := range fs {
			fn := entity.FileNode{
				Id:           uuid.NewV4().String(),
				AccountId:    accountId,
				Delete:       1,
				CacheTime:    time.Now().UnixNano(),
				FileIdDigest: "",
				IsFolder:     true,
				FileId:       item["id"].(string),
				FileName:     item["name"].(string),
				FileSize:     0,
				SizeFmt:      "-",
				FileType:     "",
				MediaType:    0,
				DownloadUrl:  "",
				CreateTime:   UTCTimeFormat(item["createdTime"].(string)),
				LastOpTime:   UTCTimeFormat(item["modifiedTime"].(string)),
				ParentId:     fileId,
				ParentPath:   p,
			}
			if p == "/" {
				fn.Path = p + fn.FileName
			} else {
				fn.Path = p + "/" + fn.FileName
			}
			FileNodeAuth(&fn, hide, hasPwd)
			if item["mimeType"].(string) == "application/vnd.google-apps.folder" {
				fn.IsFolder = true
			} else {
				size, _ := strconv.ParseInt(item["size"].(string), 10, 64)
				fn.IsFolder = false
				fn.FileType = item["fileExtension"].(string)
				fn.FileSize = size
				fn.SizeFmt = FormatFileSize(size)
				fn.MediaType = GetMimeType(fn.FileName)
			}
			if fn.IsFolder == true {
				//同步子目录&&子目录不为空
				if syncChild {
					GDGetFiles(accountId, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
				}
			}
			fns = append(fns, fn)
		}
		model.SqliteDb.Create(&fns)
		if pageToken == "" {
			break
		}
	}
}
func GDGetDownUrl(accountId, fileId string) string {
	gd := GoogleDrives[accountId]
	auth := gd.TokenType + " " + gd.AccessToken
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	resp, _ := nic.Get("https://www.googleapis.com/drive/v3/files/"+fileId+"?fields=*", nic.H{
		Proxy: Proxy,
		Headers: nic.KV{
			"Authorization": auth,
		},
	})
	if resp != nil {
		downUrl := jsoniter.Get(resp.Bytes, "webContentLink").ToString()
		return downUrl
	}
	return ""
}
func GDUpload(accountId, parentId string, files []*multipart.FileHeader) bool {
	gd := GoogleDrives[accountId]
	auth := gd.TokenType + " " + gd.AccessToken
	fmt.Println(auth)
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	for _, file := range files {
		resp, _ := nic.Post("https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable", nic.H{
			Proxy: Proxy,
			Headers: nic.KV{
				"Authorization": auth,
				"Content-Type":  "application/json; charset=UTF-8",
			},
			JSON: nic.KV{
				"name":    file.Filename,
				"parents": []string{parentId},
			},
		})
		uploadUrl := resp.Header.Get("location")
		if uploadUrl == "" {
			log.Debugf("文件上传失败：%s，上传地址为空", file.Filename)
			return false
		}
		bfs := ReadBlock(file)
		log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
		for _, bf := range bfs {
			r, _ := http.NewRequest("PUT", uploadUrl, bytes.NewReader(bf.Content))
			r.Header.Add("Content-Length", strconv.FormatInt(file.Size, 10))
			r.Header.Add("Content-Range", bf.Name)
			res, _ := http.DefaultClient.Do(r)
			defer res.Body.Close()
			body, _ := ioutil.ReadAll(res.Body)
			log.Debugf("上传接口返回：%s", body)
		}
	}
	return true
}
func GetGDContentReader(accountId, fileId, r string) (io.Reader, string) {
	gd := GoogleDrives[accountId]
	auth := gd.TokenType + " " + gd.AccessToken
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	rr := ""
	cr := ""
	if r != "" {
		startStr := strings.ReplaceAll(strings.ReplaceAll(r, "bytes=", ""), "-", "")
		start, _ := strconv.ParseInt(startStr, 10, 64)
		if start+PartSize > 11269310 {
			rr = fmt.Sprintf("bytes=%d-%d", start, 11269310-1)
			cr = fmt.Sprintf("bytes %d-%d/%d", start, 11269310-1, 11269310)
		} else {
			rr = fmt.Sprintf("bytes=%d-%d", start, start+PartSize-1)
			cr = fmt.Sprintf("bytes %d-%d/%d", start, start+PartSize-1, 11269310)
		}
	}
	resp, _ := nic.Get("https://www.googleapis.com/drive/v3/files/"+fileId+"?alt=media", nic.H{
		Proxy: Proxy,
		Headers: nic.KV{
			"Authorization": auth,
			"Range":         rr,
		},
	})
	if resp != nil {
		return bytes.NewReader(resp.Bytes), cr
	}
	return nil, ""
}
