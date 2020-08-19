package main

import (
	"github.com/eddieivan01/nic"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"log"
	"net/http"
	"os"
)

func main() {
	//gin.SetMode(gin.ReleaseMode)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("必须设置 $PORT")
	}
	//首先先生成一个gin实例
	r := gin.New()
	r.Use(gin.Logger())
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "static")
	r.StaticFile("/favicon.ico", "./static/img/favicon.ico")
	//声明一个路由
	r.GET("/", index)
	c := cron.New()
	c.AddFunc("0 0/5 * * * ?", func() {
		resp, err := nic.Get("https://pan.noki.top/", nil)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("heroku防休眠请求成功：" + resp.Status)
	})
	r.Run(":" + port) // 监听并在 0.0.0.0:8080 上启动服务

}

/**
首页
*/
func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{"name": "Libs"})
}
