package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

var version string = "1.3.0"

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
	r.Run(":" + port) // 监听并在 0.0.0.0:8080 上启动服务

}

/**
首页
*/
func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{"name": "Libs"})
}
