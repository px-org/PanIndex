package jobs

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"github.com/bluele/gcache"
	"github.com/eddieivan01/nic"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

func Run() {
	c := cron.New()
	if config.GloablConfig.HerokuKeepAlive != "" {
		c.AddFunc(config.GloablConfig.HerokuKeepAlive, func() {
			if config.GloablConfig.HerokuAppUrl != "" && config.GloablConfig.HerokuKeepAlive != "" {
				resp, err := nic.Get(config.GloablConfig.HerokuAppUrl, nil)
				if err != nil {
					log.Infoln(err.Error())
				} else {
					log.Infoln("[定时任务]heroku防休眠 >> " + resp.Status)
				}
			}
		})
	}
	if config.GloablConfig.RefreshCookie != "" {
		c.AddFunc(config.GloablConfig.RefreshCookie, func() {
			cookie := ""
			for _, account := range config.GloablConfig.Accounts {
				if account.Mode == "cloud189" {
					cookie = Util.Cloud189Login(account.User, account.Password)
				} else if account.Mode == "teambition" {
					cookie = Util.TeambitionLogin(account.User, account.Password)
				}
				if cookie != "" {
					log.Infoln("[定时任务][" + account.Name + "]cookie更新 >> 登录成功")
				} else if cookie == "" && account.Mode != "native" {
					log.Infoln("[定时任务][" + account.Name + "]cookie更新 >> 登录失败，请检查用户名和密码是否正确")
				}
			}
		})
	}
	if config.GloablConfig.UpdateFolderCache != "" {
		c.AddFunc(config.GloablConfig.UpdateFolderCache, func() {
			Util.GC = gcache.New(10).LRU().Build()
			model.SqliteDb.Where("1=1").Delete(entity.FileNode{})
			for _, account := range config.GloablConfig.Accounts {
				if account.Mode == "cloud189" {
					Util.Cloud189GetFiles(account.Id, account.RootId, account.RootId)
				} else if account.Mode == "teambition" {
					rootId := Util.ProjectIdCheck(account.RootId)
					if Util.IsPorject {
						Util.TeambitionGetProjectFiles(account.Id, rootId, "/")
					} else {
						Util.TeambitionGetFiles(account.Id, account.RootId, account.RootId, "/")
					}
				}
				var fileNodeCount int64
				model.SqliteDb.Model(&entity.FileNode{}).Where("1=1").Count(&fileNodeCount)
				if fileNodeCount > 0 {
					log.Infoln("[定时任务][" + account.Name + "]目录缓存刷新 >> 刷新成功")
				}
			}
		})
	}
	c.Start()
}
func StartInit() {
	for _, account := range config.GloablConfig.Accounts {
		cookie := ""
		model.SqliteDb.Where("account_id=?", account.Id).Delete(entity.FileNode{})
		if account.Mode == "cloud189" {
			log.Infoln("[网盘模式][" + account.Name + "] >> 天翼云网盘")
			cookie = Util.Cloud189Login(account.User, account.Password)
			Util.Cloud189GetFiles(account.Id, account.RootId, account.RootId)
		} else if account.Mode == "teambition" {
			cookie = Util.TeambitionLogin(account.User, account.Password)
			rootId := Util.ProjectIdCheck(account.RootId)
			if Util.IsPorject {
				log.Infoln("[网盘模式][" + account.Name + "] >> teambition网盘-项目")
				Util.TeambitionGetProjectFiles(account.Id, rootId, "/")
			} else {
				log.Infoln("[网盘模式][" + account.Name + "] >> teambition网盘-个人")
				Util.TeambitionGetFiles(account.Id, account.RootId, account.RootId, "/")
			}
		} else if account.Mode == "native" {
			log.Infoln("[网盘模式][" + account.Name + "] >> 本地模式")
		}
		if cookie != "" {
			log.Infoln("[程序启动][" + account.Name + "]cookie更新 >> 登录成功")
		} else if cookie == "" && account.Mode != "native" {
			log.Infoln("[程序启动][" + account.Name + "]cookie更新 >> 登录失败，请检查用户名和密码是否正确")
		}
		var fileNodeCount int64
		model.SqliteDb.Model(&entity.FileNode{}).Where("1=1").Count(&fileNodeCount)
		if fileNodeCount > 0 {
			log.Infoln("[程序启动][" + account.Name + "]目录缓存刷新 >> 刷新成功")
		}
	}
}
