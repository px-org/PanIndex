package model

import (
	"PanIndex/entity"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
)

var SqliteDb *gorm.DB

func init() {
	var err error
	SqliteDb, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		log.Println("[程序启动]Sqlite数据库 >> 连接成功")
	}
	SqliteDb.SingularTable(true)
	SqliteDb.AutoMigrate(&entity.FileNode{})
	//打印sql语句
	SqliteDb.LogMode(true)
}
