package ftp

import (
	"bytes"
	"github.com/jlaffaye/ftp"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
	"strconv"
	"time"
)

func init() {
	base.RegisterPan("ftp", &FTP{})
}

type FTP struct{}

func (F FTP) IsLogin(account *module.Account) bool {
	_, err := F.AuthLogin(account)
	if err == nil {
		return true
	} else {
		return false
	}
}

func (F FTP) AuthLogin(account *module.Account) (string, error) {
	_, err := F.FtpLogin(account, true)
	if err == nil {
		return "ftp server login success", err
	} else {
		return "", err
	}
}

func (F FTP) FtpLogin(account *module.Account, isCheck bool) (*ftp.ServerConn, error) {
	c, err := ftp.Dial(account.ApiUrl, ftp.DialWithDisabledEPSV(true))
	if err != nil {
		log.Errorf("FTP server [%s] connected failed:%s", account.ApiUrl, err)
	} else {
		if account.User != "" && account.Password != "" {
			err = c.Login(account.User, account.Password)
		} else {
			err = c.Login("anonymous", "anonymous")
		}
		if err != nil {
			log.Errorf("FTP server [%s] login failed，please check your user and password:%s", account.ApiUrl, err)
		} else {
			if isCheck {
				err = c.Logout()
				if err != nil {
					//log.Errorf("FTP server [%s] logout error:%s", account.ApiUrl, err)
					err = nil
				}
				err = c.Quit()
				if err != nil {
					log.Errorf("FTP server [%s] quit error:%s", account.ApiUrl, err)
					err = nil
				}
			}
		}
	}
	return c, err
}

func (F FTP) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	c, err := F.FtpLogin(&account, false)
	if err == nil {
		var entries []*ftp.Entry
		entries, err = c.List(fileId)
		if err == nil {
			for _, entry := range entries {
				fn := F.ToFileNode(entry)
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
				fn.AccountId = account.Id
				fn.ParentId = fileId
				fn.ParentPath = path
				fileNodes = append(fileNodes, fn)
			}
		}
	}
	if err = c.Quit(); err != nil {
		log.Error(err)
	}
	return fileNodes, err
}

func (F FTP) File(account module.Account, fileId, path string) (module.FileNode, error) {
	fn := module.FileNode{}
	if fileId == "/" {
		return module.FileNode{
			FileId:     "/",
			FileName:   "root",
			FileSize:   0,
			IsFolder:   true,
			Path:       "/",
			LastOpTime: time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	} else {
		parentId := util.GetParentPath(fileId)
		parentPath := util.GetParentPath(path)
		files, err := F.Files(account, parentId, parentPath, "", "")
		for _, file := range files {
			if file.FileId == fileId {
				return file, err
			}
		}
	}
	return fn, nil
}

func (F FTP) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	c, er := F.FtpLogin(&account, false)
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		p := path.Join(parentFileId, file.FileName)
		reader := bytes.NewReader(file.Content)
		er = c.Stor(p, reader)
		if er == nil {
			log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
		} else {
			log.Debugf("file：%s，upload failed：%s，timespan：%s", file.FileName, er, util.ShortDur(time.Now().Sub(t1)))
		}
	}
	return true, "all files uploaded", er
}

func (F FTP) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	c, err := F.FtpLogin(&account, false)
	if err == nil {
		err = c.Rename(fileId, path.Join(util.GetParentPath(fileId), name))
		if err != nil {
			return true, "File rename success", nil
		}
	}
	if err = c.Quit(); err != nil {
		log.Error(err)
	}
	return false, "File rename failed", err
}

func (F FTP) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	c, err := F.FtpLogin(&account, false)
	if err == nil {
		err = c.Delete(fileId)
		if err != nil {
			return true, "File remove success", nil
		}
	}
	if err = c.Quit(); err != nil {
		log.Error(err)
	}
	return false, "File remove failed", err
}

func (F FTP) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	c, err := F.FtpLogin(&account, false)
	if err == nil {
		err = c.MakeDir(path.Join(parentFileId, name))
		if err != nil {
			return true, "Dir create success", nil
		}
	}
	if err = c.Quit(); err != nil {
		log.Error(err)
	}
	return false, "Dir create failed", err
}

func (F FTP) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	c, err := F.FtpLogin(&account, false)
	if err == nil {
		r, er := c.Retr(fileId)
		fileName := util.GetFileName(fileId)
		content, _ := ioutil.ReadAll(r)
		r.Close()
		er = c.Stor(path.Join(targetFileId, fileName), bytes.NewReader(content))
		if er == nil {
			c.Delete(fileId)
			return true, "File move success", nil
		}
	}
	if err = c.Quit(); err != nil {
		log.Error(err)
	}
	return false, "File move failed", err
}

func (F FTP) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	c, err := F.FtpLogin(&account, false)
	if err == nil {
		r, er := c.Retr(fileId)
		fileName := util.GetFileName(fileId)
		r.Close()
		er = c.Stor(path.Join(targetFileId, fileName), r)
		if er == nil {
			return true, "File copy success", nil
		}
	}
	if err = c.Quit(); err != nil {
		log.Error(err)
	}
	return false, "File copy failed", err
}

func (F FTP) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	return fileId, nil
}

func (F FTP) GetSpaceSzie(account module.Account) (int64, int64) {
	return 0, 0
}

func (F FTP) ToFileNode(entry *ftp.Entry) module.FileNode {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileName = entry.Name
	fn.CreateTime = time.Unix(entry.Time.Unix(), 0).Format("2006-01-02 15:04:05")
	fn.LastOpTime = time.Unix(entry.Time.Unix(), 0).Format("2006-01-02 15:04:05")
	fn.IsDelete = 1
	if entry.Type == ftp.EntryTypeFolder {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	} else {
		fn.IsFolder = false
		fn.FileType = util.GetExt(entry.Name)
		fn.ViewType = util.GetViewType(fn.FileType)
		fileSize, _ := strconv.ParseInt(strconv.FormatUint(entry.Size, 10), 10, 64)
		fn.FileSize = int64(fileSize)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	}
	return fn
}

func (F FTP) ReadFileReader(account module.Account, fileId string, offset uint64) (*ftp.Response, error) {
	c, er := F.FtpLogin(&account, false)
	if er != nil {
		log.Errorf("FTP server[%s]connect or login error:%s", account.ApiUrl, er)
	} else {
		r, err := c.RetrFrom(fileId, offset)
		return r, err
	}
	return nil, nil
}
