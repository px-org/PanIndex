package jobs

import (
	"github.com/px-org/PanIndex/dao"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/ali"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

func Run() {
	util.Cron = cron.New(cron.WithSeconds())
	//aliyundrive onedrive googledrive refresh token
	util.Cron.AddFunc("0 59 * * * ?", func() {
		for _, account := range module.GloablConfig.Accounts {
			if account.Mode == "aliyundrive" || account.Mode == "aliyundrive-share" || account.Mode == "onedrive" || account.Mode == "onedrive-cn" || account.Mode == "googledrive" || account.Mode == "pikpak" {
				dao.SyncAccountStatus(account)
			}
		}
	})
	//cookie有效性检测
	util.Cron.AddFunc("0 0/1 * * * ?", func() {
		for _, account := range module.GloablConfig.Accounts {
			p, _ := base.GetPan(account.Mode)
			if account.Mode == "cloud189" || account.Mode == "yun139" ||
				account.Mode == "teambition-us" || account.Mode == "teambition" || account.Mode == "123" || account.Mode == "115" {
				status := p.IsLogin(&account)
				if !status {
					log.Debugf("[cron] account:%s, logout, start login...", account.Name)
					dao.SyncAccountStatus(account)
				} else {
					//log.Debugf("[cron] account:%s, status ok", account.Name)
				}
			}
		}
	})
	util.Cron.AddFunc("0 0 0 4 * ?", func() {
		//extra jobs, eg:sign
		for _, account := range module.GloablConfig.Accounts {
			p, _ := base.GetPan(account.Mode)
			if account.Mode == "aliyundrive" {
				//aliyundrive sign
				p.(ali.Ali).SignActivity(account)
			}
		}
	})
	accounts := []module.Account{}
	dao.DB.Raw("select * from account where cache_policy ='dc' and sync_cron!='' and sync_dir!=''").Find(&accounts)
	for _, ac := range accounts {
		entryId, er := util.Cron.AddFunc(ac.SyncCron, func() {
			dao.SyncFilesCache(ac)
			//log.Debugf("[%s] [%s] [%s] [%s] [%s]", ac.Name, ac.Mode, ac.CachePolicy, ac.SyncCron, ac.SyncDir)
		})
		if er == nil {
			util.CacheCronMap[ac.Id] = entryId
		}
	}
	util.Cron.Start()
}
