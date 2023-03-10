package native

import (
	"fmt"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/v3/disk"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

func init() {
	base.RegisterPan("native", &Native{})
}

type Native struct{}

func (n Native) IsLogin(account *module.Account) bool {
	return true
}

func (n Native) AuthLogin(account *module.Account) (string, error) {
	return "no need login", nil
}

func (n Native) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fns := []module.FileNode{}
	if util.FileExist(fileId) && IsDirectory(fileId) {
		fileInfos, err := ioutil.ReadDir(fileId)
		if err != nil {
			log.Error(err)
			return fns, err
		} else {
			for _, fileInfo := range fileInfos {
				file := module.FileNode{
					Id:         uuid.NewV4().String(),
					AccountId:  account.Id,
					FileId:     filepath.Join(fileId, fileInfo.Name()),
					IsFolder:   fileInfo.IsDir(),
					FileName:   fileInfo.Name(),
					FileSize:   fileInfo.Size(),
					SizeFmt:    util.FormatFileSize(fileInfo.Size()),
					FileType:   util.GetExt(fileInfo.Name()),
					Path:       PathJoin(path, fileInfo.Name()),
					ViewType:   util.GetViewType(util.GetExt(fileInfo.Name())),
					LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
					ParentId:   fileId,
					ParentPath: path,
					IsDelete:   1,
				}
				fns = append(fns, file)
			}
		}
	}
	return fns, nil
}

func (n Native) File(account module.Account, fileId, path string) (module.FileNode, error) {
	fn := module.FileNode{}
	fileInfo, err := os.Stat(fileId)
	if err != nil {
		log.Error(err)
		return fn, err
	} else {
		fn = module.FileNode{
			Id:         uuid.NewV4().String(),
			AccountId:  account.Id,
			FileId:     fileId,
			IsFolder:   fileInfo.IsDir(),
			FileName:   fileInfo.Name(),
			FileSize:   fileInfo.Size(),
			SizeFmt:    util.FormatFileSize(fileInfo.Size()),
			FileType:   util.GetExt(fileInfo.Name()),
			Path:       path,
			ViewType:   util.GetViewType(util.GetExt(fileInfo.Name())),
			LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
			ParentId:   filepath.Dir(fileId),
			ParentPath: util.GetParentPath(path),
			IsDelete:   1,
		}
	}
	return fn, nil
}

func (n Native) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		fileName := filepath.Join(parentFileId, file.FileName)
		err := ioutil.WriteFile(fileName, file.Content, 0644)
		if err != nil {
			log.Error(err)
		} else {
			log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
		}
	}
	return true, "all files uploaded", nil
}

func (n Native) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	oldPath := fileId
	newPath := filepath.Join(filepath.Dir(oldPath), name)
	err := os.Rename(oldPath, newPath)
	if err == nil {
		return true, newPath, nil
	} else {
		return false, "Rename failed", err
	}
}

func (n Native) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	err := os.Remove(fileId)
	if err != nil {
		log.Error(err)
		return false, "File remove Error", err
	} else {
		return true, "File remove success", nil
	}
}

func (n Native) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	err := os.MkdirAll(filepath.Join(parentFileId, name), os.ModePerm)
	if err != nil {
		log.Error(err)
		return false, "Dir create Error", err
	} else {
		return true, "Dir create success", nil
	}
}

func (n Native) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	n.Copy(account, fileId, targetFileId, overwrite)
	return n.Remove(account, fileId)
}

func (n Native) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	fileInfo, err := os.Stat(fileId)
	if err != nil {
		log.Error(err)
		return false, "Failed to get source file", err
	} else {
		if fileInfo.IsDir() {
			err = DirCopy(fileId, targetFileId)
		} else {
			_, fileName := util.ParsePath(fileId)
			targetFile := filepath.Join(targetFileId, fileName)
			err = FileCopy(fileId, targetFile)
		}
		if err != nil {
			return false, "File copy failed", err
		} else {
			return true, "File copy success", nil
		}
	}
}

func (n Native) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	return fileId, nil
}

func (n Native) GetSpaceSzie(account module.Account) (int64, int64) {
	usageStat, err := disk.Usage(account.RootId)
	if err == nil {
		return int64(usageStat.Total), int64(usageStat.Used)
	}
	return 0, 0
}

func IsDirectory(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func PathJoin(path, fileName string) string {
	if path == "/" {
		return fmt.Sprintf("%s%s", path, fileName)
	} else {
		return fmt.Sprintf("%s/%s", path, fileName)
	}
}

func FileCopy(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Dir copies a whole directory recursively
func DirCopy(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = DirCopy(srcfp, dstfp); err != nil {
				log.Error(err)
			}
		} else {
			if err = FileCopy(srcfp, dstfp); err != nil {
				log.Error(err)
			}
		}
	}
	return nil
}
