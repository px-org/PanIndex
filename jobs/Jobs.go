package jobs

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"github.com/eddieivan01/nic"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

func Run() {
	c := cron.New()
	c.AddFunc(config.GloablConfig.CronExps.HerokuKeepAlive, func() {
		if config.GloablConfig.HerokuAppUrl != "" && config.GloablConfig.CronExps.HerokuKeepAlive != "" {
			resp, err := nic.Get(config.GloablConfig.HerokuAppUrl, nil)
			if err != nil {
				log.Infoln(err.Error())
			} else {
				log.Infoln("[定时任务]heroku防休眠 >> " + resp.Status)
			}
		}
	})
	c.AddFunc(config.GloablConfig.CronExps.RefreshCookie, func() {
		cookie := ""
		if config.GloablConfig.Mode == "cloud189" {
			cookie = Util.Cloud189Login(config.GloablConfig.User, config.GloablConfig.Password)
		} else if config.GloablConfig.Mode == "teambition" {
			cookie = Util.TeambitionLogin(config.GloablConfig.User, config.GloablConfig.Password)
		}
		if cookie != "" {
			log.Infoln("[定时任务]cookie更新 >> 登录成功")
		} else {
			log.Infoln("[定时任务]cookie更新 >> 登录失败，请检查用户名和密码是否正确")
		}
	})
	c.AddFunc(config.GloablConfig.CronExps.UpdateFolderCache, func() {
		model.SqliteDb.Delete(entity.FileNode{})
		if config.GloablConfig.Mode == "cloud189" {
			Util.Cloud189GetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId)
		} else if config.GloablConfig.Mode == "teambition" {
			rootId := Util.ProjectIdCheck(config.GloablConfig.RootId)
			if Util.IsPorject {
				Util.TeambitionGetProjectFiles(rootId, "/")
			} else {
				Util.TeambitionGetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId, "/")
			}
		}
		var fileNodeCount int
		model.SqliteDb.Model(&entity.FileNode{}).Where("1=1").Count(&fileNodeCount)
		if fileNodeCount > 0 {
			log.Infoln("[定时任务]目录缓存刷新 >> 刷新成功")
		}
	})
	c.Start()
}
func StartInit() {
	cookie := ""
	model.SqliteDb.Delete(entity.FileNode{})
	if config.GloablConfig.Mode == "cloud189" {
		log.Infoln("[网盘模式] >> 天翼云网盘")
		cookie = Util.Cloud189Login(config.GloablConfig.User, config.GloablConfig.Password)
		Util.Cloud189GetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId)
	} else if config.GloablConfig.Mode == "teambition" {
		cookie = Util.TeambitionLogin(config.GloablConfig.User, config.GloablConfig.Password)
		rootId := Util.ProjectIdCheck(config.GloablConfig.RootId)
		if Util.IsPorject {
			log.Infoln("[网盘模式] >> teambition网盘-项目")
			Util.TeambitionGetProjectFiles(rootId, "/")
		} else {
			log.Infoln("[网盘模式] >> teambition网盘-个人")
			Util.TeambitionGetFiles(config.GloablConfig.RootId, config.GloablConfig.RootId, "/")
		}
	} else if config.GloablConfig.Mode == "native" {
		log.Infoln("[网盘模式] >> 本地模式")
	}
	if cookie != "" {
		log.Infoln("[程序启动]cookie更新 >> 登录成功")
	} else if cookie == "" && config.GloablConfig.Mode != "native" {
		log.Infoln("[程序启动]cookie更新 >> 登录失败，请检查用户名和密码是否正确")
	}
	var fileNodeCount int
	model.SqliteDb.Model(&entity.FileNode{}).Where("1=1").Count(&fileNodeCount)
	if fileNodeCount > 0 {
		log.Infoln("[程序启动]目录缓存刷新 >> 刷新成功")
	}
}
