package jobs

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"github.com/eddieivan01/nic"
	"github.com/robfig/cron"
	"log"
	"os"
)

func Run() {
	c := cron.New()
	c.AddFunc("0 0/5 * * * ?", func() {
		if config.Config189.HerokuAppUrl != "" {
			resp, err := nic.Get(config.Config189.HerokuAppUrl, nil)
			if err != nil {
				log.Println(err.Error())
			} else {
				log.Println("[定时任务]heroku防休眠 >> " + resp.Status)
			}
		}
	})
	c.AddFunc("0 0 8 1/1 * ?", func() {
		Util.Cloud189Login(os.Getenv("USER"), os.Getenv("PASSWORD"))
		log.Println("[定时任务]cookie更新 >> 登录成功")
	})
	c.AddFunc("0 0 0/1 * * ?", func() {
		model.SqliteDb.Model(&entity.FileNode{}).Update("`delete`", "1")
		Util.Cloud189GetFiles(config.Config189.RootId, config.Config189.RootId)
		model.SqliteDb.Delete(entity.FileNode{}, "`delete` = 1")
		log.Println("[定时任务]目录缓存刷新 >> 刷新成功")
	})
	c.Start()
}
func StartInit() {
	cookie := Util.Cloud189Login(config.Config189.User, config.Config189.Password)
	if cookie != "" {
		log.Println("[程序启动]cookie更新 >> 登录成功")
	} else {
		log.Println("[程序启动]cookie更新 >> 登录失败，请检查用户名和密码是否正确")
	}
	model.SqliteDb.Model(&entity.FileNode{}).Update("`delete`", "1")
	Util.Cloud189GetFiles(config.Config189.RootId, config.Config189.RootId)
	model.SqliteDb.Delete(entity.FileNode{}, "`delete` = 1")
	log.Println("[程序启动]目录缓存刷新 >> 刷新成功")
}
