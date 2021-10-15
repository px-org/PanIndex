package jobs

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"errors"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

var CacheCron *cron.Cron

func AutoCacheRun() {
	CacheCron = cron.New()
	accounts := []entity.Account{}
	model.SqliteDb.Raw("select * from account where mode !=?", "native").Find(&accounts)
	for _, ac := range accounts {
		if ac.SyncCron != "" {
			CacheCron.AddFunc(ac.SyncCron, func() {
				SyncOneAccount(ac)
			})
		}
	}
	CacheCron.Start()
}
func Run() {
	c := cron.New()
	/*if config.GloablConfig.UpdateFolderCache != "" {
		c.AddFunc(config.GloablConfig.UpdateFolderCache, func() {
			Util.GC = gcache.New(10).LRU().Build()
			for _, account := range SelectAllAccounts() {
				SyncOneAccount(account)
			}
		})
	}*/
	//阿里云盘 accesstoken过期时间为2小时，这里采取固定cron刷新的方式
	c.AddFunc("0 0 0/1 * * ?", func() {
		for k, v := range Util.Alis {
			account := entity.Account{}
			account.Id = k
			account.RefreshToken = v.RefreshToken
			Util.AliRefreshToken(account)
		}
		for k, v := range Util.OneDrives {
			account := entity.Account{}
			model.SqliteDb.Raw("select * from account where id=?", k).First(&account)
			account.RefreshToken = v.RefreshToken
			Util.OneDriveRefreshToken(account)
		}
	})
	//cookie有效性检测
	c.AddFunc("0 0/1 * * * ?", func() {
		for _, account := range SelectAllAccounts() {
			if account.Mode != "native" && account.Mode != "aliyundrive" && account.Mode != "onedrive" {
				if account.CookieStatus == 4 {
					//频繁登录或用户名密码错误导致的失败
					//跳过验证
					log.Infof("[COOKIE定时检查][%s]>>%s>>由频繁登录或用户名密码错误导致的失败不再检测", account.Name, account.Mode)
					continue
				}
				cookieValid := true
				if account.Mode == "cloud189" {
					if _, ok := Util.CLoud189Sessions[account.Id]; ok {
						cookieValid = Util.Cloud189IsLogin(account.Id)
					}
				}
				if _, ok := Util.TeambitionSessions[account.Id]; ok {
					if account.Mode == "teambition" {
						cookieValid = Util.TeambitionIsLogin(account.Id, false)
					}
					if account.Mode == "teambition-us" {
						cookieValid = Util.TeambitionIsLogin(account.Id, true)
					}
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
	} else if account.Mode == "onedrive" {
		cookie = Util.OneDriveRefreshToken(account)
		msg = "[" + account.Name + "] >> 微软云盘"
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("refresh_token", cookie)
	} else if account.Mode == "native" {
		msg = "[" + account.Name + "] >> 本地模式"
	}
	if cookie != "" && cookie != "4" {
		log.Infoln(msg + " >> COOKIE更新 >> 登录成功")
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("cookie_status", 2)
	} else if cookie == "4" && account.Mode != "native" {
		log.Infoln(msg + "COOKIE更新 >> 登录失败，请检查用户名,密码(token)是否正确")
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("cookie_status", 4)
	} else if cookie == "" && account.Mode != "native" {
		log.Infoln(msg + "COOKIE更新 >> 登录失败，原因未知")
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("cookie_status", 3)
	}
}
func SyncOneAccount(account entity.Account) {
	t1 := time.Now()
	syncDir := ""
	syncChild := true
	fileId := account.RootId
	if account.SyncDir == "" {
		syncDir = "/"
	} else {
		syncDir = account.SyncDir
	}
	if account.SyncChild == 0 {
		syncChild = true
	} else {
		syncChild = false
	}
	if syncDir != "/" {
		//查询fileId
		dbFile := entity.FileNode{}
		result := model.SqliteDb.Raw("select * from file_node where path=? and `delete`=0 and account_id=?", syncDir, account.Id).Take(&dbFile)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fileId = dbFile.FileId
		}
	}
	model.SqliteDb.Table("account").Where("id=?", account.Id).Update("status", -1)
	if account.Mode == "cloud189" {
		Util.Cloud189GetFiles(account.Id, fileId, fileId, syncDir, 0, 0, syncChild)
	} else if account.Mode == "teambition" {
		rootId := Util.ProjectIdCheck("www", account.Id, account.RootId)
		if Util.TeambitionSessions[account.Id].IsPorject {
			if syncDir == "/" {
				fileId = rootId
			}
			Util.TeambitionGetProjectFiles("www", account.Id, fileId, syncDir, 0, 0, syncChild)
		} else {
			Util.TeambitionGetFiles(account.Id, account.RootId, fileId, syncDir, 0, 0, syncChild)
		}
	} else if account.Mode == "teambition-us" {
		rootId := Util.ProjectIdCheck("us", account.Id, account.RootId)
		if Util.TeambitionSessions[account.Id].IsPorject {
			if syncDir == "/" {
				fileId = rootId
			}
			Util.TeambitionGetProjectFiles("us", account.Id, fileId, syncDir, 0, 0, syncChild)
		} else {
		}
	} else if account.Mode == "aliyundrive" {
		cookie := Util.AliRefreshToken(account)
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("refresh_token", cookie)
		Util.AliGetFiles(account.Id, account.RootId, fileId, syncDir, 0, 0, syncChild)
	} else if account.Mode == "onedrive" {
		cookie := Util.OneDriveRefreshToken(account)
		model.SqliteDb.Table("account").Where("id=?", account.Id).Update("refresh_token", cookie)
		Util.OndriveGetFiles("", account.Id, fileId, syncDir, 0, 0, syncChild)
	} else if account.Mode == "native" {
	}
	var fileNodeCount int64
	model.SqliteDb.Model(&entity.FileNode{}).Where("account_id=? and `delete`=1", account.Id).Count(&fileNodeCount)
	status := 3
	if int(fileNodeCount) > 0 {
		status = 2
		if syncDir == "/" || syncDir == "" {
			//删除旧数据
			model.SqliteDb.Where("account_id=? and `delete`=0", account.Id).Delete(entity.FileNode{})
			//暴露新数据
			model.SqliteDb.Table("file_node").Where("account_id=?", account.Id).Update("delete", 0)
		} else {
			RefreshFileNodes(account.Id, fileId)
		}
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
func SelectAllAccounts() []entity.Account {
	var list []entity.Account
	model.SqliteDb.Where("1=1").Find(&list)
	return list
}
func RefreshFileNodes(accountId, fileId string) {
	tmpList := []entity.FileNode{}
	list := []entity.FileNode{}
	model.SqliteDb.Raw("select * from file_node where parent_id=? and `delete`=0 and account_id=?", fileId, accountId).Find(&tmpList)
	getAllNodes(&tmpList, &list)
	for _, fn := range list {
		model.SqliteDb.Where("id=?", fn.Id).Delete(entity.FileNode{})
	}
	model.SqliteDb.Table("file_node").Where("account_id=?", accountId).Update("delete", 0)
}

func getAllNodes(tmpList, list *[]entity.FileNode) {
	for _, fn := range *tmpList {
		tmpList = &[]entity.FileNode{}
		model.SqliteDb.Raw("select * from file_node where parent_id=? and `delete`=0", fn.FileId).Find(&tmpList)
		*list = append(*list, fn)
		if len(*tmpList) != 0 {
			getAllNodes(tmpList, list)
		}
	}
}
