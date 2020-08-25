package main

import (
	"PanIndex/entity"
	"PanIndex/jobs"
	"PanIndex/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("必须设置 $PORT")
	}
	//首先先生成一个gin实例
	r := gin.New()
	r.Use(gin.Logger())
	r.LoadHTMLGlob("templates/189/classic/*.html")
	r.Static("/static", "static")
	r.StaticFile("/favicon.ico", "./static/img/favicon.ico")
	//声明一个路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		log.Println(path)
		if strings.HasPrefix(path, "/api") {
			//文件夹下载
			downloadMultiFiles(c)
		} else {
			index(c)
		}
	})
	jobs.Run()
	jobs.StartInit()
	r.Run(":" + port) // 监听并在 0.0.0.0:8080 上启动服务

}

func index(c *gin.Context) {
	result := service.GetFilesByPath(c.Request.URL.Path)
	fs, ok := result["List"].([]entity.FileNode)
	if ok {
		if len(fs) == 1 && !fs[0].IsFolder && result["isFile"].(bool) {
			//文件
			downUrl := service.GetDownlaodUrl(fs[0].FileIdDigest)
			c.Redirect(http.StatusFound, downUrl)
			/*if fs[0].MediaType == 1{
				//图片
			}else if fs[0].MediaType == 2{
				//音频
			}else if fs[0].MediaType == 3{
				//视频
			}else if fs[0].MediaType == 4{
				//文本
			}else{
				//其他类型，直接下载
			}*/
		}
	}
	c.HTML(http.StatusOK, "index.html", result)
}
func downloadMultiFiles(c *gin.Context) {
	fileId := c.Query("fileId")
	downUrl := service.GetDownlaodMultiFiles(fileId)
	c.Redirect(http.StatusFound, downUrl)
}
