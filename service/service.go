package service

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"strings"
)

func GetFilesByPath(path, pwd string) map[string]interface{} {
	result := make(map[string]interface{})
	list := []entity.FileNode{}
	model.SqliteDb.Raw("select * from file_node where parent_path=? and hide = 0", path).Find(&list)
	result["isFile"] = false
	if len(list) == 0 {
		result["isFile"] = true
		model.SqliteDb.Raw("select * from file_node where path = ? and is_folder = 0 and hide = 0", path).Find(&list)
	}
	result["HasPwd"] = false
	fileNode := entity.FileNode{}
	model.SqliteDb.Raw("select * from file_node where path = ? and is_folder = 1", path).First(&fileNode)
	PwdDirIds := config.Config189.PwdDirId
	for _, pdi := range PwdDirIds {
		if pdi.Id == fileNode.FileId && pwd != pdi.Pwd {
			result["HasPwd"] = true
			result["FileId"] = fileNode.FileId
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
	return result
}

func GetDownlaodUrl(fileIdDigest string) string {
	return Util.GetDownlaodUrl(fileIdDigest)
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
