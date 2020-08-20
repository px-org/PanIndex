package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"strings"
)

var SqliteDb *gorm.DB

func init() {
	var err error
	SqliteDb, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		fmt.Println("Sqlite数据库连接成功")
	}
	SqliteDb.SingularTable(true)
	SqliteDb.AutoMigrate(&FileNode{})
	//打印sql语句
	//SqliteDb.LogMode(true)
}

func GetFilesByPath(path string) map[string]interface{} {
	result := make(map[string]interface{})
	list := []FileNode{}
	SqliteDb.Raw("select * from file_node where parent_path = ?", path).Find(&list)
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
