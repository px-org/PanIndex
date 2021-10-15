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

const InitSql string = `
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('host', '0.0.0.0', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('port', '5238', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('site_name', '', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('account_choose', 'default', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('admin_password', 'PanIndex', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('api_token', '', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('s_column', 'default', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('s_order', 'asc', 'common');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('favicon_url', '', 'appearance');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('footer', '', 'appearance');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('css', '', 'appearance');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('js', '', 'appearance');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('enable_preview', '1', 'view');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('image', 'png,gif,jpg,bmp,jpeg,ico', 'view');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('video', 'mp4,mkv,m3u8,ts,avi', 'view');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('audio', 'mp3,wav,flac,ape', 'view');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('code', 'txt,go,html,js,java,json,css,lua,sh,sql,py,cpp,xml,jsp,properties,yaml,ini', 'view');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('doc', 'doc,docx,dotx,ppt,pptx,xls,xlsx', 'view');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('other', '*', 'view');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('pwd_dir_id', '', 'pdi');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('hide_file_id', '', 'hfi');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('theme', 'mdui', 'appearance');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('refresh_cookie', '', 'cron');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('update_folder_cache', '', 'cron');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('heroku_keep_alive', '', 'cron');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('heroku_app_url', '', 'cron');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('enable_safety_link', '0', 'safety');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('only_referrer', '', 'safety');
INSERT OR IGNORE INTO "main"."config_item"("k", "v", "g") VALUES ('is_null_referrer', '0', 'safety');
`

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
	SqliteDb.AutoMigrate(&entity.ShareInfo{})
	SqliteDb.AutoMigrate(&entity.ConfigItem{})
	SqliteDb.AutoMigrate(&entity.Account{})
	//初始化数据
	var count int64
	err = SqliteDb.Model(entity.ConfigItem{}).Count(&count).Error
	if err != nil {
		panic(err)
	} else if count == 0 {
		rand.Seed(time.Now().UnixNano())
		ApiToken := strconv.Itoa(rand.Intn(10000000))
		configItem := entity.ConfigItem{K: "api_token", V: ApiToken, G: "common"}
		SqliteDb.Create(configItem)
	}
	SqliteDb.Model(entity.ConfigItem{}).Exec(InitSql)
	//同步旧版配置数据
	syncOldConfig()
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	if host != "" {
		//启动时指定了host/port
		SqliteDb.Table("config_item").Where("k='host'").Update("v", host)
	}
	if port != "" {
		//启动时指定了host/port
		SqliteDb.Table("config_item").Where("k='port'").Update("v", port)
	}
}

func syncOldConfig() {
	old := entity.Config{}
	var count int64
	SqliteDb.Model(entity.Config{}).Count(&count)
	if count > 0 {
		SqliteDb.Model(entity.Config{}).Raw("select * from config where 1=1").First(&old)
		if old.Host != "" {
			log.Info("检测到旧版本配置表存在，开始同步基础配置...")
			SqliteDb.Table("config_item").Where("k='host'").Update("v", old.Host)
			SqliteDb.Table("config_item").Where("k='port'").Update("v", old.Port)
			SqliteDb.Table("config_item").Where("k='admin_password'").Update("v", old.AdminPassword)
			SqliteDb.Table("config_item").Where("k='site_name'").Update("v", old.SiteName)
			SqliteDb.Table("config_item").Where("k='theme'").Update("v", old.Theme)
			SqliteDb.Table("config_item").Where("k='account_choose'").Update("v", old.AccountChoose)
			SqliteDb.Table("config_item").Where("k='api_token'").Update("v", old.ApiToken)
			SqliteDb.Table("config_item").Where("k='pwd_dir_id'").Update("v", old.PwdDirId)
			SqliteDb.Table("config_item").Where("k='hide_file_id'").Update("v", old.HideFileId)
			SqliteDb.Table("config_item").Where("k='only_referrer'").Update("v", old.OnlyReferrer)
			SqliteDb.Table("config_item").Where("k='favicon_url'").Update("v", old.FaviconUrl)
			SqliteDb.Table("config_item").Where("k='footer'").Update("v", old.Footer)
			SqliteDb.Table("account").Where("1=1").Update("sync_cron", old.UpdateFolderCache)
			accounts := []entity.Account{}
			SqliteDb.Table("account").Raw("select * from account where 1=1 order by `default` desc").Find(&accounts)
			for i, account := range accounts {
				i++
				SqliteDb.Table("account").Where("id=?", account.Id).Update("seq", i)
			}
			log.Info("旧版本配置同步完成，开始删除旧表...")
			SqliteDb.Table("config_item").Exec("drop table main.config")
			SqliteDb.Table("config_item").Exec("drop table main.damagou")
		}
	}
}
