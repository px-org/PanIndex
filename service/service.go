package service

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func GetFilesByPath(path, pwd string) map[string]interface{} {
	result := make(map[string]interface{})
	list := []entity.FileNode{}
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if config.GloablConfig.Mode == "native" {
		//列出文件夹相对路径
		rootPath := config.GloablConfig.RootId
		fullPath := filepath.Join(rootPath, path)
		if Util.FileExist(fullPath) {
			if Util.IsDirectory(fullPath) {
				//是目录
				// 读取该文件夹下所有文件
				fileInfos, err := ioutil.ReadDir(fullPath)
				//默认按照目录，时间倒序排列
				sort.Slice(fileInfos, func(i, j int) bool {
					d1 := 0
					if fileInfos[i].IsDir() {
						d1 = 1
					}
					d2 := 0
					if fileInfos[j].IsDir() {
						d2 = 1
					}
					if d1 > d2 {
						return true
					} else if d1 == d2 {
						return fileInfos[i].ModTime().After(fileInfos[j].ModTime())
					} else {
						return false
					}
				})
				if err != nil {
					panic(err.Error())
				} else {
					for _, fileInfo := range fileInfos {
						fileId := filepath.Join(fullPath, fileInfo.Name())
						// 当前文件是隐藏文件(以.开头)则不显示
						if Util.IsHiddenFile(fileInfo.Name()) {
							continue
						}
						//指定隐藏的文件或目录过滤
						if config.GloablConfig.HideFileId != "" {
							listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
							sort.Strings(listSTring)
							i := sort.SearchStrings(listSTring, fileId)
							if i < len(listSTring) && listSTring[i] == fileId {
								continue
							}
						}
						fileType := Util.GetMimeType(fileInfo)
						// 实例化FileNode
						file := entity.FileNode{
							FileId:     fileId,
							IsFolder:   fileInfo.IsDir(),
							FileName:   fileInfo.Name(),
							FileSize:   int64(fileInfo.Size()),
							SizeFmt:    Util.FormatFileSize(int64(fileInfo.Size())),
							FileType:   strings.TrimLeft(filepath.Ext(fileInfo.Name()), "."),
							Path:       filepath.Join(path, fileInfo.Name()),
							MediaType:  fileType,
							LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
						}
						// 添加到切片中等待json序列化
						list = append(list, file)
					}
				}
				result["isFile"] = false
				result["HasPwd"] = false
				PwdDirIds := config.GloablConfig.PwdDirId
				for _, pdi := range PwdDirIds {
					if pdi.Id == fullPath && pwd != pdi.Pwd {
						result["HasPwd"] = true
						result["FileId"] = fullPath
					}
				}
			} else {
				//是文件
				fileInfo, err := os.Stat(fullPath)
				if err != nil {
					panic(err.Error())
				} else {
					fileType := Util.GetMimeType(fileInfo)
					file := entity.FileNode{
						FileId:     fullPath,
						IsFolder:   fileInfo.IsDir(),
						FileName:   fileInfo.Name(),
						FileSize:   int64(fileInfo.Size()),
						SizeFmt:    Util.FormatFileSize(int64(fileInfo.Size())),
						FileType:   strings.TrimLeft(filepath.Ext(fileInfo.Name()), "."),
						Path:       filepath.Join(path, fileInfo.Name()),
						MediaType:  fileType,
						LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
					}
					// 添加到切片中等待json序列化
					list = append(list, file)
				}
				result["isFile"] = true
			}
		}
	} else {
		model.SqliteDb.Raw("select * from file_node where parent_path=? and hide = 0", path).Find(&list)
		result["isFile"] = false
		if len(list) == 0 {
			result["isFile"] = true
			model.SqliteDb.Raw("select * from file_node where path = ? and is_folder = 0 and hide = 0", path).Find(&list)
		}
		result["HasPwd"] = false
		fileNode := entity.FileNode{}
		model.SqliteDb.Raw("select * from file_node where path = ? and is_folder = 1", path).First(&fileNode)
		PwdDirIds := config.GloablConfig.PwdDirId
		for _, pdi := range PwdDirIds {
			if pdi.Id == fileNode.FileId && pwd != pdi.Pwd {
				result["HasPwd"] = true
				result["FileId"] = fileNode.FileId
			}
		}
	}
	result["List"] = list
	result["Path"] = path
	if path == "/" {
		result["HasParent"] = false
	} else {
		result["HasParent"] = true
	}
	result["ParentPath"] = PetParentPath(path)
	if config.GloablConfig.Mode == "teambition" {
		result["SurportFolderDown"] = false
	} else {
		result["SurportFolderDown"] = true
	}
	return result
}

func GetDownlaodUrl(fileNode entity.FileNode) string {
	if config.GloablConfig.Mode == "cloud189" {
		return Util.GetDownlaodUrl(fileNode.FileIdDigest)
	} else if config.GloablConfig.Mode == "teambition" {
		return Util.GetTeambitionDownUrl(fileNode.FileId)
	} else if config.GloablConfig.Mode == "native" {
	}
	return ""
}

func GetDownlaodMultiFiles(fileId string) string {
	return Util.GetDownlaodMultiFiles(fileId)
}

func PetParentPath(p string) string {
	if p == "/" {
		return ""
	} else {
		s := ""
		ss := strings.Split(p, "/")
		for i := 0; i < len(ss)-1; i++ {
			if ss[i] != "" {
				s += "/" + ss[i]
			}
		}
		if s == "" {
			s = "/"
		}
		return s
	}
}

//获取查询游标start
func GetPageStart(pageNo, pageSize int) int {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 0
	}
	return (pageNo - 1) * pageSize
}

//获取总页数
func GetTotalPage(totalCount, pageSize int) int {
	if pageSize == 0 {
		return 0
	}
	if totalCount%pageSize == 0 {
		return totalCount / pageSize
	} else {
		return totalCount/pageSize + 1
	}
}

//刷新目录缓存
func UpdateFolderCache() {
	model.SqliteDb.Delete(entity.FileNode{})
	if config.GloablConfig.Mode == "cloud189" {
		Util.Cloud189GetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId)
	} else if config.GloablConfig.Mode == "teambition" {
		Util.TeambitionGetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId, "/")
	} else if config.GloablConfig.Mode == "native" {
	}
}

//刷新登录cookie
func RefreshCookie() {
	if config.GloablConfig.Mode == "cloud189" {
		Util.Cloud189Login(config.GloablConfig.User, config.GloablConfig.Password)
	} else if config.GloablConfig.Mode == "teambition" {
		Util.TeambitionLogin(config.GloablConfig.User, config.GloablConfig.Password)
	} else if config.GloablConfig.Mode == "native" {
	}

}
func IsDirectory(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
