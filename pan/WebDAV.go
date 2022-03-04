package pan

import (
	"bytes"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

func init() {
	RegisterPan("webdav", &WebDav{})
}

type WebDav struct{}

func (w WebDav) IsLogin(account *module.Account) bool {
	return true
}

func (w WebDav) AuthLogin(account *module.Account) (string, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err != nil {
		log.Errorf("WebDav server [%s] connected failed:%s", account.ApiUrl, err)
	} else {
		return "WebDav server connect success", nil
	}
	return "", err
}

func (w WebDav) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err == nil {
		var files []os.FileInfo
		files, err = c.ReadDir(fileId)
		if err == nil {
			for _, file := range files {
				fn := w.ToFileNode(file, file.Name())
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
	return fileNodes, err
}

func (w WebDav) File(account module.Account, fileId, path string) (module.FileNode, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err == nil {
		var file os.FileInfo
		statId := ""
		ext := filepath.Ext(fileId)
		if ext != "" {
			statId = fileId
		} else {
			statId = fileId + "/"
		}
		file, err = c.Stat(statId)
		if err == nil {
			fn := w.ToFileNode(file, util.GetFileName(fileId))
			fn.FileName = util.GetFileName(fileId)
			fn.Path = path
			fn.FileId = fileId
			fn.AccountId = account.Id
			fn.ParentId = util.GetParentPath(fileId)
			fn.ParentPath = util.GetParentPath(path)
			return fn, nil
		}
	}
	return module.FileNode{}, nil
}

func (w WebDav) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		p := path.Join(parentFileId, file.FileName)
		reader := bytes.NewReader(file.Content)
		err = c.WriteStream(p, reader, 0664)
		if err == nil {
			log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
		} else {
			log.Debugf("file：%s，upload failed：%s，timespan：%s", file.FileName, err, util.ShortDur(time.Now().Sub(t1)))
		}
	}
	return true, "all files uploaded", err
}

func (w WebDav) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err == nil {
		oldPath := fileId
		newPath := path.Join(filepath.Dir(fileId), name)
		err = c.Rename(PathClean(oldPath), PathClean(newPath), true)
		if err != nil {
			return true, "File rename success", nil
		}
	}
	return false, "File rename failed", err
}

func (w WebDav) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err == nil {
		err = c.Remove(PathClean(fileId))
		if err != nil {
			return true, "File remove success", nil
		}
	}
	return false, "File remove failed", err
}

func (w WebDav) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err == nil {
		newPath := path.Join(parentFileId, name)
		err = c.Mkdir(PathClean(newPath), 0664)
		if err != nil {
			return true, "Dir create success", nil
		}
	}
	return false, "Dir create failed", err
}

func (w WebDav) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err == nil {
		oldPath := fileId
		newPath := path.Join(targetFileId, util.GetFileName(fileId))
		err = c.Copy(PathClean(oldPath), PathClean(newPath), overwrite)
		if err != nil {
			return true, "File move success", nil
		}
		c.Remove(fileId)
	}
	return false, "File move failed", err
}

func (w WebDav) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err == nil {
		oldPath := fileId
		newPath := path.Join(targetFileId, util.GetFileName(fileId))
		err = c.Copy(PathClean(oldPath), PathClean(newPath), overwrite)
		if err != nil {
			return true, "File copy success", nil
		}
	}
	return false, "File copy failed", err
}

func (w WebDav) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	return fileId, nil
}

func (w WebDav) GetSpaceSzie(account module.Account) (int64, int64) {
	return 0, 0
}

func (w WebDav) ToFileNode(file os.FileInfo, fileName string) module.FileNode {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileName = fileName
	fn.CreateTime = time.Unix(file.ModTime().Unix(), 0).Format("2006-01-02 15:04:05")
	fn.LastOpTime = time.Unix(file.ModTime().Unix(), 0).Format("2006-01-02 15:04:05")
	fn.IsDelete = 1
	if file.IsDir() {
		fn.IsFolder = file.IsDir()
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	} else {
		fn.IsFolder = file.IsDir()
		fn.FileType = util.GetExt(fileName)
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = file.Size()
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	}
	return fn
}

func (w WebDav) ReadFileReader(account module.Account, fileId string, offset, fileSize int64) (io.ReadCloser, error) {
	c := gowebdav.NewClient(account.ApiUrl, account.User, account.Password)
	err := c.Connect()
	if err != nil {
		log.Errorf("WebDac server[%s]connect error:%s", account.ApiUrl, err)
	} else {
		return c.ReadStreamRange(fileId, offset, fileSize)
	}
	return nil, err
}

func PathClean(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		path = path + "/"
	}
	return path
}
