package model

import (
	"PanIndex/entity"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var SqliteDb *gorm.DB

func InitDb(host, port, dataPath string, debug bool) {
	if os.Getenv("PAN_INDEX_DATA_PATH") != "" {
		dataPath = os.Getenv("PAN_INDEX_DATA_PATH")
	}
	if dataPath == "" {
		dataPath = "data"
	}
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		os.Mkdir(dataPath, os.ModePerm)
	}
	var err error
	LogLevel := logger.Silent
	if debug {
		LogLevel = logger.Info
	}
	SqliteDb, err = gorm.Open(sqlite.Open(dataPath+"/data.db"), &gorm.Config{
		Logger: logger.Default.LogMode(LogLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		log.Info("[程序启动]Sqlite数据库 >> 连接成功")
	}
	//SqliteDb.SingularTable(true)
	//打印sql语句
	//SqliteDb.Logger.Info()
	//创建表
	SqliteDb.AutoMigrate(&entity.FileNode{})
	SqliteDb.AutoMigrate(&entity.Config{})
	SqliteDb.AutoMigrate(&entity.Account{})
	SqliteDb.AutoMigrate(&entity.Damagou{})
	//初始化数据
	c := entity.Config{}
	SqliteDb.Raw("select * from config where 1=1").Find(&c)
	if c.Host == "" {
		rand.Seed(time.Now().UnixNano())
		ApiToken := strconv.Itoa(rand.Intn(10000))
		SqliteDb.Create(&entity.Config{"0.0.0.0", 5238, nil, "", "", "", ApiToken, "", "default", "mdui", "PanIndex", entity.Damagou{}, "", "0 0 8 1/1 * ?", "", "", "", ""})
	}
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	if host != "" {
		//启动时指定了host/port
		SqliteDb.Table("config").Where("1=1").Update("host", host)
	}
	if port != "" {
		//启动时指定了host/port
		SqliteDb.Table("config").Where("1=1").Update("port", port)
	}
}
