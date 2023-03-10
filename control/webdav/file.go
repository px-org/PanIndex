// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

import (
	"fmt"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/service"
	"github.com/px-org/PanIndex/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type FileSystem struct{}

func (fs *FileSystem) File(account module.Account, path, fullPath string) (module.FileNode, bool) {
	if strings.HasPrefix(util.GetFileName(fullPath), "._") {
		return module.FileNode{}, false
	}
	if fullPath == "/" {
		//accounts list
		return module.FileNode{
			FileId:     "root",
			FileName:   "root",
			FileSize:   0,
			IsFolder:   true,
			Path:       "/",
			LastOpTime: time.Now().Format("2006-01-02 15:04:05"),
		}, true
	} else {
		fn, err := service.File(account, path, fullPath)
		if err != nil {
			log.Error(err)
		}
		if fn.FileId != "" || (fn.FileId == "" && path == "/") {
			return fn, true
		}
		return module.FileNode{}, false
	}
}

func (fs *FileSystem) Files(account module.Account, path, fullPath string) []module.FileNode {
	fns := []module.FileNode{}
	if strings.HasPrefix(util.GetFileName(fullPath), "._") {
		return fns
	}
	fileNamePath := module.GloablConfig.DavPath
	if fullPath != "/" {
		fileNamePath = fileNamePath + fullPath + "/"
	} else {
		fileNamePath = fileNamePath + "/"
	}
	if fullPath == "/" {
		for _, ac := range module.GloablConfig.Accounts {
			fn := module.FileNode{
				FileId:     fmt.Sprintf("/%s", ac.Name),
				IsFolder:   true,
				FileName:   fileNamePath + ac.Name,
				FileSize:   int64(ac.FilesCount),
				SizeFmt:    "-",
				FileType:   "",
				Path:       fmt.Sprintf("/%s", ac.Name),
				ViewType:   "",
				LastOpTime: ac.LastOpTime,
				ParentId:   "",
			}
			fns = append(fns, fn)
		}
		return fns
	} else {
		fns0 := service.Files(account, path, fullPath)
		for _, fn := range fns0 {
			fn.FileName = fileNamePath + fn.FileName
			fns = append(fns, fn)
		}
		return fns
	}
}

func (fs *FileSystem) Delete(account module.Account, path, fullPath string) error {
	return service.DeleteFile(account, path, fullPath)
}

func (fs *FileSystem) Upload(account module.Account, req *http.Request, path, fullPath, fileId string, overwrite bool) error {
	_, fileName := util.ParsePath(path)
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	files := []*module.UploadInfo{{
		FileName:    fileName,
		FileSize:    req.ContentLength,
		ContentType: req.Header.Get("Content-Type"),
		Content:     content,
	}}
	p, _ := base.GetPan(account.Mode)
	ok, _, err := p.UploadFiles(account, fileId, files, overwrite)
	log.Debugf("[webdav] Upload '%s', result: %v", fullPath, ok)
	if ok && err == nil && account.CachePolicy != "nc" {
		go UploadCall(account, path, fullPath, overwrite)
	}
	return err
}

func UploadCall(account module.Account, path, fullPath string, overwrite bool) {
	//upload success
	p, _ := base.GetPan(account.Mode)
	fId := service.GetFileIdByPath(account, path, fullPath)
	log.Debug("Upload file id: ", fId, " start callback:")
	fn, _ := p.File(account, fId, fullPath)
	service.UploadCall(account, fn, overwrite)
}

func (fs *FileSystem) Mkdir(account module.Account, path string, fullPath string) error {
	if account.Id == "" {
		return ErrNotImplemented
	}
	p, _ := base.GetPan(account.Mode)
	filePath, fileName := util.ParsePath(path)
	fileFullPath, _ := util.ParsePath(fullPath)
	parentFileId := service.GetFileIdByPath(account, filePath, fileFullPath)
	ok, _, err := p.Mkdir(account, parentFileId, fileName)
	log.Debugf("[webdav] Mkdir '%s', result: %v", fullPath, ok)
	if ok && err == nil && account.CachePolicy != "nc" {
		fileId := service.GetFileIdFromApi(account, path)
		fn, _ := p.File(account, fileId, fullPath)
		service.MkdirCall(account, fn)
	}
	return err
}

func (fs *FileSystem) Copy(src string, dst string, overwrite bool) (int, error) {
	srcAccount, srcFullPath, srcPath, _ := util.ParseFullPath(src, "")
	dstAccount, dstFullPath, dstPath, _ := util.ParseFullPath(dst, "")
	if srcAccount.Id != dstAccount.Id {
		return http.StatusMethodNotAllowed, ErrNotImplemented
	}
	srcFileId := service.GetFileIdByPath(srcAccount, srcPath, srcFullPath)
	targetFileId := service.GetFileIdByPath(dstAccount, dstPath, dstFullPath)
	p, _ := base.GetPan(srcAccount.Mode)
	ok, _, err := p.Copy(srcAccount, srcFileId, targetFileId, overwrite)
	if ok {
		return http.StatusCreated, nil
	} else {
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

func (fs *FileSystem) Move(src string, dst string, overwrite bool) (int, error) {
	i := strings.Index(src, module.GloablConfig.DavPath) + len(module.GloablConfig.DavPath)
	src = src[i:]
	dst = dst[i:]
	srcAccount, srcFullPath, srcPath, _ := util.ParseFullPath(src, "")
	dstAccount, dstFullPath, dstPath, _ := util.ParseFullPath(dst, "")
	if srcAccount.Id != dstAccount.Id {
		return http.StatusMethodNotAllowed, ErrNotImplemented
	}
	srcParentPath := util.GetParentPath(srcPath)
	dstParentPath, fileName := util.ParsePath(dstPath)
	p, _ := base.GetPan(srcAccount.Mode)
	srcFileId := service.GetFileIdByPath(srcAccount, srcPath, srcFullPath)
	var ok bool
	var err error
	if srcParentPath == dstParentPath {
		//rename
		ok, _, err = p.Rename(srcAccount, srcFileId, fileName)
	} else {
		targetFileId := service.GetFileIdByPath(dstAccount, dstParentPath, util.GetParentPath(dstFullPath))
		ok, _, err = p.Move(srcAccount, srcFileId, targetFileId, overwrite)
	}
	if ok {
		service.MoveCall(srcAccount, srcFileId, srcFullPath, dstFullPath)
	}
	return http.StatusNoContent, err

}

func (fs *FileSystem) GetSpace(account module.Account, fullPath string) (int64, int64) {
	accounts := module.GloablConfig.Accounts
	if len(accounts) > 0 && module.GloablConfig.AccountChoose == "display" && fullPath == "/" {
		var total int64
		var used int64
		for _, ac := range accounts {
			disk, _ := base.GetPan(ac.Mode)
			oneTotal, oneUsed := disk.GetSpaceSzie(ac)
			total += oneTotal
			used += oneUsed
		}
		return total, used
	} else {
		disk, _ := base.GetPan(account.Mode)
		return disk.GetSpaceSzie(account)
	}
}

type WalkFunc func(path string, account module.Account, p, fullPath string, info module.FileNode, err error) error

// walkFS traverses filesystem fs starting at name up to depth levels.
//
// Allowed values for depth are 0, 1 or infiniteDepth. For each visited node,
// walkFS calls walkFn. If a visited file system node is a directory and
// walkFn returns filepath.SkipDir, walkFS will skip traversal of this node.
func walkFS(ctx context.Context, fs FileSystem, depth int, name string, account module.Account, p, fullPath string, info module.FileNode, walkFn WalkFunc) error {
	// This implementation is based on Walk's code in the standard path/filepath package.
	err := walkFn(name, account, p, fullPath, info, nil)
	if err != nil {
		if info.IsFolder && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsFolder || depth == 0 {
		return nil
	}
	if depth == 1 {
		depth = 0
	}
	files := fs.Files(account, p, fullPath)
	for _, file := range files {
		pathName := path.Join(p, file.FileName)
		fullPathName := path.Join(fullPath, file.FileName)
		err = walkFS(ctx, fs, depth, file.FileName, account, pathName, fullPathName, file, walkFn)
		if err != nil {
			if !file.IsFolder || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
