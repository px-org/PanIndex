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
	"math"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
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
		path = fmt.Sprintf(":%s:/", path)
	}
	return "https://graph.microsoft.com/v1.0" + "/me/drive/root" + path + query
}
func OndriveGetFiles(url, accountId, fileId, p string, hide, hasPwd int, syncChild bool) {
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
				OndriveGetFiles("", accountId, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
			}
		}
		fn.Id = uuid.NewV4().String()
		fn.CacheTime = time.Now().UnixNano()
		model.SqliteDb.Create(fn)
	}
	nextLink := jsoniter.Get(byteFiles, "@odata.nextLink").ToString()
	if nextLink != "" {
		OndriveGetFiles(nextLink, accountId, fileId, p, hide, hasPwd, syncChild)
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
	for _, file := range files {
		t1 := time.Now()
		//bfs := ReadBlock(file, 16384000)327680
		bfs := ReadBlock(file)
		uploadUrl := CreateUploadSession(accountId, filepath.Join(parentId, file.Filename))
		for _, bf := range bfs {
			r, _ := http.NewRequest("PUT", uploadUrl, bytes.NewReader(bf.Content))
			r.Header.Add("Content-Length", strconv.FormatInt(file.Size, 10))
			r.Header.Add("Content-Range", bf.Name)
			res, _ := http.DefaultClient.Do(r)
			defer res.Body.Close()
			body, _ := ioutil.ReadAll(res.Body)
			log.Debugf("上传接口返回：%s", body)
		}
		log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
		log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
	}
	return false
}

func ReadBlock(file *multipart.FileHeader) []BlockFile {
	bfs := []BlockFile{}
	FileHandle, err := file.Open()
	if err != nil {
		log.Error(err)
		return bfs
	}
	defer FileHandle.Close()
	const fileChunk = 16384000 //15.625MB
	totalPartsNum := uint64(math.Ceil(float64(file.Size) / float64(fileChunk)))
	log.Debugf("Spliting to %d pieces.\n", totalPartsNum)
	totalSize := file.Size
	off := int64(0) //起始点
	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int64(math.Min(fileChunk, float64(file.Size-int64(i*fileChunk))))
		contentRange := fmt.Sprintf("bytes %d-%d/%d", off, off+partSize-1, totalSize)
		partBuffer := make([]byte, partSize)
		FileHandle.Read(partBuffer)
		bfs = append(bfs, BlockFile{partBuffer, contentRange})
		off += partSize
	}
	return bfs
}

type BlockFile struct {
	Content []byte
	Name    string
}
type SplitFile struct {
	File      *multipart.FileHeader
	FileName  string
	Length    int64
	Size      int64
	BlockSize int64
	blockpath []string
}

func CreateUploadSession(accountId, filePath string) string {
	od := OneDrives[accountId]
	auth := od.TokenType + " " + od.AccessToken
	resp, err := nic.Post("https://graph.microsoft.com/v1.0/me/drive/root:"+filePath+":/createUploadSession", nic.H{
		Headers: nic.KV{
			"Authorization": auth,
			"Host":          "graph.microsoft.com",
		},
	})
	if err != nil {
		log.Error(err)
		return ""
	}
	if resp.StatusCode == 409 {
		log.Debugf("被请求的资源的当前状态之间存在冲突，请求无法完成，重新提交上传请求")
		return ""
	}
	log.Debugf("[OneDrive]获取上传url:%s", resp.Text)
	return jsoniter.Get(resp.Bytes, "uploadUrl").ToString()
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
