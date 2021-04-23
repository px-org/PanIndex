package model

import (
	"PanIndex/entity"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var SqliteDb *gorm.DB

func InitDb() {
	path := "data"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	var err error
	SqliteDb, err = gorm.Open("sqlite3", "data/data.db")
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		log.Println("[程序启动]Sqlite数据库 >> 连接成功")
	}
	SqliteDb.SingularTable(true)
	//打印sql语句
	SqliteDb.LogMode(true)
	//创建表
	SqliteDb.AutoMigrate(&entity.FileNode{})
	SqliteDb.AutoMigrate(&entity.Config{})
	SqliteDb.AutoMigrate(&entity.Account{})
	SqliteDb.AutoMigrate(&entity.CronExps{})
	SqliteDb.AutoMigrate(&entity.PwdDirId{})
	SqliteDb.AutoMigrate(&entity.Damagou{})
	//初始化数据
	c := entity.Config{}
	SqliteDb.Raw("select * from config where 1=1").Find(&c)
	if c.Host == "" {
		rand.Seed(time.Now().UnixNano())
		ApiToken := strconv.Itoa(rand.Intn(10000))
		SqliteDb.Save(&entity.Config{"0.0.0.0", 8080, nil, nil, "", "", ApiToken, "mdui", "PanIndex", entity.Damagou{}, nil, entity.CronExps{}, ""})
	}
	crons := entity.CronExps{}
	SqliteDb.Raw("select * from cron_exps where 1=1").Find(&crons)
	if crons.RefreshCookie == "" {
		SqliteDb.Save(&entity.CronExps{"0 0 8 1/1 * ?", "0 0 0/1 * * ?", ""})
	}
}
