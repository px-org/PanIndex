package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func WebDavLogin(account entity.Account) string {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err != nil {
		log.Errorf("WebDav服务器[%s]连接失败:%s", account.ApiUrl, err)
	} else {
		return "webdav server connect success"
	}
	return ""
}

func WebDavGetFiles(account entity.Account, fileId, path string, hide, hasPwd int, syncChild bool) {
	fileNodes := []entity.FileNode{}
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err != nil {
		log.Errorf("WebDav服务器[%s]连接失败:%s", account.ApiUrl, err)
	} else {
		files, err := c.ReadDir(fileId)
		if err == nil {
			for _, fileInfo := range files {
				fileType := GetMimeType(fileInfo.Name())
				fn := entity.FileNode{
					Id:         uuid.NewV4().String(),
					AccountId:  account.Id,
					IsFolder:   fileInfo.IsDir(),
					FileName:   fileInfo.Name(),
					FileSize:   int64(fileInfo.Size()),
					SizeFmt:    FormatFileSize(int64(fileInfo.Size())),
					FileType:   strings.TrimLeft(filepath.Ext(fileInfo.Name()), "."),
					Path:       path,
					MediaType:  fileType,
					LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
					Hide:       0,
					HasPwd:     0,
				}
				if path == "/" {
					fn.Path = path + fn.FileName
				} else {
					fn.Path = path + "/" + fn.FileName
				}
				if fileId == "/" {
					fn.FileId = fileId + fn.FileName
				} else {
					fn.FileId = fileId + "/" + fn.FileName
				}
				fn.ParentPath = path
				fn.ParentId = fileId
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
				if fn.IsFolder == true {
					if syncChild {
						WebDavGetFiles(account, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
					}
				}
				fn.CacheTime = time.Now().UnixNano()
				fn.Delete = 1
				model.SqliteDb.Create(fn)
				fileNodes = append(fileNodes, fn)
			}
		}
	}
}
func WebDavReadFileToBytes(account entity.Account, fileId string) []byte {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if err != nil {
		log.Errorf("WebDav服务器[%s]连接失败:%s", account.ApiUrl, err)
	} else {
		buf, err := c.Read(fileId)
		if err != nil {
			panic(err)
		}
		return buf
	}
	return nil
}
func WebDavUpload(account entity.Account, fileId string, files []*multipart.FileHeader) bool {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if err != nil {
		log.Errorf("WebDav服务器[%s]连接失败:%s", account.ApiUrl, err)
	} else {
		for _, file := range files {
			log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
			t1 := time.Now()
			fileContent, _ := file.Open()
			defer fileContent.Close()
			filePath := ""
			if fileId == "/" {
				filePath = fileId + file.Filename
			} else {
				filePath = fileId + "/" + file.Filename
			}
			err := c.WriteStream(filePath, fileContent, 0644)
			if err != nil {
				panic(err)
			}
			log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
		}
		return true
	}
	return false
}
