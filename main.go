package main

import (
	"github.com/eddieivan01/nic"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"log"
	"net/http"
	"os"
	"strings"
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
	//r.GET("/:xx/:s", index)
	//r.GET("/:second/:third", index)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/test") {
			test(c)
		} else {
			c.HTML(http.StatusOK, "index.html", GetFilesByPath(path))
		}
	})
	c := cron.New()
	c.AddFunc("0 0/5 * * * ?", func() {
		resp, err := nic.Get("https://pan.noki.top/", nil)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("heroku防休眠请求成功：" + resp.Status)
	})
	login(os.Getenv("USER"), os.Getenv("PASSWORD"))
	r.Run(":" + port) // 监听并在 0.0.0.0:8080 上启动服务

}

func index(c *gin.Context) {
	//path := c.Param("path")
	log.Println(c.Request.URL.Path)
	c.HTML(http.StatusOK, "index.html", gin.H{"name": c.Request.URL.Path})
}

func test(c *gin.Context) {
	GetFiles("-11", "-11")
	c.JSON(http.StatusOK, "success")
}
