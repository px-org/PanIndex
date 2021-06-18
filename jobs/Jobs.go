package jobs

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"github.com/bluele/gcache"
	"github.com/libsgh/nic"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"time"
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
			if config.GloablConfig.HerokuAppUrl != "" {
				resp, err := nic.Get(config.GloablConfig.HerokuAppUrl+"/api/admin/envToConfig?token="+config.GloablConfig.ApiToken, nil)
				if err != nil {
					log.Infoln(err.Error())
				} else {
					log.Infoln("[定时任务]heroku配置防丢失 >> " + resp.Status)
				}
			}
			for _, account := range config.GloablConfig.Accounts {
				AccountLogin(account)
			}
		})
	}
	if config.GloablConfig.UpdateFolderCache != "" {
		c.AddFunc(config.GloablConfig.UpdateFolderCache, func() {
			Util.GC = gcache.New(10).LRU().Build()
			for _, account := range config.GloablConfig.Accounts {
				SyncOneAccount(account)
			}
		})
	}
	//阿里云盘 accesstoken过期时间为2小时，这里采取固定cron刷新的方式
	c.AddFunc("0 0 0/1 * * ?", func() {
		for k, v := range Util.Alis {
			account := entity.Account{}
			account.Id = k
			account.RefreshToken = v.RefreshToken
			Util.AliRefreshToken(account)
		}
	})
	//cookie有效性检测
	c.AddFunc("0 0/1 * * * ?", func() {
		for _, account := range config.GloablConfig.Accounts {
			if account.Mode != "native" || account.Mode != "aliyundrive" {
				cookieValid := true
				if account.Mode == "cloud189" {
					cookieValid = Util.Cloud189IsLogin(account.Id)
				}
				if account.Mode == "teambition" {
					cookieValid = Util.TeambitionIsLogin(account.Id, false)
				}
				if account.Mode == "teambition-us" {
					cookieValid = Util.TeambitionIsLogin(account.Id, true)
				}
				if cookieValid == false {
					log.Infof("[COOKIE定时检查][%s]>>%s>>失效", account.Name, account.Mode)
					log.Infof("开始刷新[%s]的COOKIE...", account.Name)
					AccountLogin(account)
				} else {
					log.Debugf("[COOKIE定时检查][%s]>>%s>>有效", account.Name, account.Mode)
				}
			}
		}
	})
	c.Start()
}
func StartInit() {
	for _, account := range config.GloablConfig.Accounts {
		AccountLogin(account)
		//SyncOneAccount(account)
	}
}

func SyncInit(account entity.Account) {
	AccountLogin(account)
	SyncOneAccount(account)
}

func AccountLogin(account entity.Account) {
	cookie := ""
	msg := ""
	model.SqliteDb.Table("account").Where("id=?", account.Id).Update("cookie_status", -1)
	if account.Mode == "cloud189" {
		msg = "[网盘模式][" + account.Name + "] >> 天翼云网盘"
		cookie = Util.Cloud189Login(account.Id, account.User, account.Password)
	} else if account.Mode == "teambition" {
		cookie = Util.TeambitionLogin(account.Id, account.User, account.Password)
		Util.ProjectIdCheck("www", account.Id, account.RootId)
		if Util.TeambitionSessions[account.Id].IsPorject {
			msg = "[" + account.Name + "] >> teambition网盘-项目"
		} else {
			msg = "[" + account.Name + "] >> teambition网盘-个人"
		}
	} else if account.Mode == "teambition-us" {
		cookie = Util.TeambitionUSLogin(account.Id, account.User, account.Password)
		Util.ProjectIdCheck("us", account.Id, account.RootId)
		if Util.TeambitionSessions[account.Id].IsPorject {
			msg = "[" + account.Name + "] >> teambition国际盘-项目"
		}
	} else if account.Mode == "aliyundrive" {
		cookie = Util.AliRefreshToken(account)
		msg = "[" + account.Name + "] >> 阿里云盘"
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("refresh_token", cookie)
	} else if account.Mode == "native" {
		msg = "[" + account.Name + "] >> 本地模式"
	}
	if cookie != "" {
		log.Infoln(msg + " >> COOKIE更新 >> 登录成功")
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("cookie_status", 2)
	} else if cookie == "" && account.Mode != "native" {
		log.Infoln(msg + "COOKIE更新 >> 登录失败，请检查用户名,密码(token)是否正确")
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("cookie_status", 3)
	}
}
func SyncOneAccount(account entity.Account) {
	t1 := time.Now()
	model.SqliteDb.Table("account").Where("id=?", account.Id).Update("status", -1)
	if account.Mode == "cloud189" {
		Util.Cloud189GetFiles(account.Id, account.RootId, account.RootId, "")
	} else if account.Mode == "teambition" {
		rootId := Util.ProjectIdCheck("www", account.Id, account.RootId)
		if Util.TeambitionSessions[account.Id].IsPorject {
			Util.TeambitionGetProjectFiles("www", account.Id, rootId, "/")
		} else {
			Util.TeambitionGetFiles(account.Id, account.RootId, account.RootId, "/")
		}
	} else if account.Mode == "teambition-us" {
		rootId := Util.ProjectIdCheck("us", account.Id, account.RootId)
		if Util.TeambitionSessions[account.Id].IsPorject {
			Util.TeambitionGetProjectFiles("us", account.Id, rootId, "/")
		} else {
		}
	} else if account.Mode == "aliyundrive" {
		Util.AliGetFiles(account.Id, account.RootId, account.RootId, "/")
	} else if account.Mode == "native" {
	}
	//删除旧数据
	model.SqliteDb.Where("account_id=? and `delete`=0", account.Id).Delete(entity.FileNode{})
	//暴露新数据
	model.SqliteDb.Table("file_node").Where("account_id=?", account.Id).Update("delete", 0)
	var fileNodeCount int64
	model.SqliteDb.Model(&entity.FileNode{}).Where("account_id=?", account.Id).Count(&fileNodeCount)
	status := 3
	if int(fileNodeCount) > 0 {
		status = 2
		log.Infoln("[目录缓存][" + account.Name + "]缓存刷新 >> 刷新成功")
	}
	t2 := time.Now()
	d := t2.Sub(t1)
	//timespan := fmt.Sprintf("%v分%v秒", d.Minutes(), d.Seconds())
	now := time.Now().UTC().Add(8 * time.Hour)
	//更新同步记录
	model.SqliteDb.Table("account").Where("id=?", account.Id).Updates(map[string]interface{}{
		"status": status, "files_count": int(fileNodeCount), "last_op_time": now.Format("2006-01-02 15:04:05"),
		"time_span": Util.ShortDur(d),
	})
}
