package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"github.com/jlaffaye/ftp"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"time"
)

func FtpLogin(account entity.Account, isCheck bool) (*ftp.ServerConn, error) {
	c, err := ftp.Connect(account.ApiUrl)
	if err != nil {
		log.Errorf("FTP服务器[%s]连接失败:%s", account.ApiUrl, err)
	} else {
		if account.User != "" && account.Password != "" {
			err = c.Login(account.User, account.Password)
		} else {
			err = c.Login("anonymous", "anonymous")
		}
		if err != nil {
			log.Errorf("FTP服务器[%s]登录失败，请检查用户名密码是否正确:%s", account.ApiUrl, err)
		} else {
			if isCheck {
				err = c.Logout()
				if err != nil {
					log.Errorf("FTP服务器[%s]退出登录异常:%s", account.ApiUrl, err)
					err = nil
				}
				err = c.Quit()
				if err != nil {
					log.Errorf("FTP服务器[%s]退出异常:%s", account.ApiUrl, err)
					err = nil
				}
			}
		}
	}
	return c, err
}

func FtpGetFiles(account entity.Account, fileId, path string, hide, hasPwd int, syncChild bool) {
	c, er := FtpLogin(account, false)
	fileNodes := []entity.FileNode{}
	if er != nil {
		log.Errorf("FTP服务器[%s]连接或登录异常:%s", account.ApiUrl, er)
	} else {
		entries, err := c.List(fileId)
		if er != nil {
			log.Errorf("FTP服务器[%s]目录列表[%s]加载失败:%s", account.ApiUrl, path, er)
		} else {
			for _, entry := range entries {
				fn := entity.FileNode{}
				fn.AccountId = account.Id
				fn.Id = uuid.NewV4().String()
				fn.FileName = entry.Name
				fileSize, _ := strconv.ParseInt(strconv.FormatUint(entry.Size, 10), 10, 64)
				fn.FileSize = fileSize
				fn.SizeFmt = FormatFileSize(fileSize)
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
				fn.FileIdDigest = entry.Target
				fn.FileType = GetFileType(fn.FileName)
				fn.MediaType = GetMimeType(fn.FileName)
				if entry.Type == ftp.EntryTypeFolder {
					fn.IsFolder = true
				} else {
					fn.IsFolder = false
				}
				fn.IsStarred = true
				fn.LastOpTime = time.Unix(entry.Time.Unix(), 0).Format("2006-01-02 15:04:05")
				fn.CreateTime = fn.LastOpTime
				fn.ParentPath = path
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
				if fn.IsFolder == true {
					if syncChild {
						FtpGetFiles(account, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
					}
				}
				fn.CacheTime = time.Now().UnixNano()
				fn.Delete = 1
				model.SqliteDb.Create(fn)
				fileNodes = append(fileNodes, fn)
			}
		}
		if err = c.Quit(); err != nil {
			log.Fatal(err)
		}
	}
}

func FtpReadFileToBytes(account entity.Account, fileId string) []byte {
	c, er := FtpLogin(account, false)
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if er != nil {
		log.Errorf("FTP服务器[%s]连接或登录异常:%s", account.ApiUrl, er)
	} else {
		r, err := c.Retr(fileId)
		if err != nil {
			panic(err)
		}
		defer r.Close()
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		return buf
	}
	return nil
}

func FtpUpload(account entity.Account, fileId string, files []*multipart.FileHeader) bool {
	c, er := FtpLogin(account, false)
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if er != nil {
		log.Errorf("FTP服务器[%s]连接或登录异常:%s", account.ApiUrl, er)
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
			err := c.Stor(filePath, fileContent)
			if err != nil {
				panic(err)
			}
			log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
		}
		return true
	}
	return false
}
