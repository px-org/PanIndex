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
		resp, err := nic.Get("https://pan.noki.top/", nil)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("[定时任务]heroku防休眠 >> " + resp.Status)
	})
	c.AddFunc("0 0 8 1/1 * ?", func() {
		Util.Cloud189Login(os.Getenv("USER"), os.Getenv("PASSWORD"))
		log.Println("[定时任务]cookie更新 >> 登录成功")
	})
	c.AddFunc("0 0 0/1 * * ?", func() {
		model.SqliteDb.Delete(&entity.FileNode{})
		Util.Cloud189GetFiles(config.Config189.RootId, config.Config189.RootId)
		log.Println("[定时任务]目录缓存刷新 >> 刷新成功")
	})
	c.Start()
}
func StartInit() {
	config.LoadCloud189Config()
	if config.Config189.User != "" {
		log.Println("[程序启动]配置加载 >> 获取成功")
	}
	cookie := Util.Cloud189Login(config.Config189.User, config.Config189.Password)
	if cookie != "" {
		log.Println("[程序启动]cookie更新 >> 登录成功")
		log.Println(cookie)
	}
	model.SqliteDb.Delete(&entity.FileNode{})
	Util.Cloud189GetFiles(config.Config189.RootId, config.Config189.RootId)
}
