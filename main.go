package main

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/jobs"
	"PanIndex/service"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var configPath = flag.String("config.path", "", "配置文件config.json的路径")

func main() {
	flag.Parse()

	// 配置文件应该最先加载，因为要读取模板名字
	config.LoadCloud189Config(*configPath)
	if config.Config189.User != "" {
		log.Println("[程序启动]配置加载 >> 获取成功")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	//	staticBox := packr.NewBox("./static")
	r.SetHTMLTemplate(initTemplates())
	//	r.LoadHTMLFiles("templates/**")
	//	r.Static("/static", "./static")
	//	r.StaticFS("/static", staticBox)
	r.StaticFile("/favicon.ico", "./static/img/favicon.ico")
	staticBox := packr.New("static", "./static")
	r.StaticFS("/static", staticBox)
	//声明一个路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		if method == "POST" && path == "/api/downloadMultiFiles" {
			//文件夹下载
			downloadMultiFiles(c)
		} else if method == "GET" && path == "/api/updateFolderCache" {
			requestToken := c.Query("token")
			if requestToken == config.Config189.ApiToken {
				go updateCaches()
				log.Println("[API请求]目录缓存刷新 >> 请求刷新")
				message := "Cache update successful"
				c.String(http.StatusOK, message)
			} else {
				message := "Invalid api token"
				c.String(http.StatusOK, message)
			}
		} else if method == "GET" && path == "/api/shareToDown" {
			shareToDown(c)
		} else {
			index(c)
		}
	})
	jobs.Run()
	go jobs.StartInit()
	r.Run(fmt.Sprintf("%s:%d", config.Config189.Host, config.Config189.Port)) // 监听并在 0.0.0.0:8080 上启动服务

}

func initTemplates() *template.Template {
	tmpFile := strings.Join([]string{"189/", "/index.html"}, config.Config189.Theme)
	box := packr.New("templates", "./templates")
	t := template.New("")
	tmpl := t.New(tmpFile)
	data, _ := box.FindString(tmpFile)
	tmpl.Parse(data)
	return t
}

func index(c *gin.Context) {
	tmpFile := strings.Join([]string{"189/", "/index.html"}, config.Config189.Theme)
	pwd := ""
	pwdCookie, err := c.Request.Cookie("dir_pwd")
	if err == nil {
		decodePwd, err := url.QueryUnescape(pwdCookie.Value)
		if err != nil {
			log.Println(err)
		}
		pwd = decodePwd
	}
	pathName := c.Request.URL.Path
	if pathName != "/" && pathName[len(pathName)-1:] == "/" {
		pathName = pathName[0 : len(pathName)-1]
	}
	result := service.GetFilesByPath(pathName, pwd)
	result["HerokuappUrl"] = config.Config189.HerokuAppUrl
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
	c.HTML(http.StatusOK, tmpFile, result)
}

func downloadMultiFiles(c *gin.Context) {
	fileId := c.Query("fileId")
	downUrl := service.GetDownlaodMultiFiles(fileId)
	c.JSON(http.StatusOK, gin.H{"redirect_url": downUrl})
}

func updateCaches() {
	service.UpdateFolderCache()
	log.Println("[API请求]目录缓存刷新 >> 刷新成功")
}
func shareToDown(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET")
	c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")
	url := c.Query("url")
	passCode := c.Query("passCode")
	fileId := c.Query("fileId")
	fileIdDigest := c.Query("fileIdDigest")
	downUrl := ""
	if fileIdDigest != "" {
		downUrl = service.GetDownlaodUrl(fileIdDigest)
	} else {
		downUrl = Util.Cloud189shareToDown(url, passCode, fileId)
	}
	c.String(http.StatusOK, downUrl)
	/*if jsoniter.Valid([]byte(downUrl)) == true {
		c.String(http.StatusOK, downUrl)
	} else {
		c.Redirect(http.StatusFound, downUrl)
	}*/
}
