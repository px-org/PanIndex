package jobs

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"github.com/eddieivan01/nic"
	"github.com/robfig/cron"
	"log"
)

func Run() {
	c := cron.New()
	c.AddFunc(config.GloablConfig.CronExps.HerokuKeepAlive, func() {
		if config.GloablConfig.HerokuAppUrl != "" && config.GloablConfig.CronExps.HerokuKeepAlive != "" {
			resp, err := nic.Get(config.GloablConfig.HerokuAppUrl, nil)
			if err != nil {
				log.Println(err.Error())
			} else {
				log.Println("[定时任务]heroku防休眠 >> " + resp.Status)
			}
		}
	})
	c.AddFunc(config.GloablConfig.CronExps.RefreshCookie, func() {
		Util.Cloud189Login(config.GloablConfig.User, config.GloablConfig.Password)
		log.Println("[定时任务]cookie更新 >> 登录成功")
	})
	c.AddFunc(config.GloablConfig.CronExps.UpdateFolderCache, func() {
		model.SqliteDb.Model(&entity.FileNode{}).Update("`delete`", "1")
		Util.Cloud189GetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId)
		model.SqliteDb.Delete(entity.FileNode{}, "`delete` = 1")
		log.Println("[定时任务]目录缓存刷新 >> 刷新成功")
	})
	c.Start()
}
func StartInit() {
	cookie := Util.Cloud189Login(config.GloablConfig.User, config.GloablConfig.Password)
	if cookie != "" {
		log.Println("[程序启动]cookie更新 >> 登录成功")
	} else {
		log.Println("[程序启动]cookie更新 >> 登录失败，请检查用户名和密码是否正确")
	}
	model.SqliteDb.Delete(entity.FileNode{})
	Util.Cloud189GetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId)
	log.Println("[程序启动]目录缓存刷新 >> 刷新成功")
}
