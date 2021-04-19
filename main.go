package main

import (
	"PanIndex/Util"
	"PanIndex/boot"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/service"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var configPath = flag.String("config", "config.json", "配置文件config.json的路径")

func main() {
	flag.Parse()
	boot.Start(*configPath)
	r := gin.New()
	r.Use(gin.Logger())
	//	staticBox := packr.NewBox("./static")
	r.SetHTMLTemplate(initTemplates())
	//	r.LoadHTMLFiles("templates/**")
	//	r.Static("/static", "./static")
	//	r.StaticFS("/static", staticBox)
	//r.StaticFile("/favicon-cloud189.ico", "./static/img/favicon-cloud189.ico")
	initStaticBox(r)
	//声明一个路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		if method == "POST" && path == "/api/downloadMultiFiles" {
			//文件夹下载
			downloadMultiFiles(c)
		} else if method == "GET" && path == "/api/updateFolderCache" {
			requestToken := c.Query("token")
			if requestToken == config.GloablConfig.ApiToken {
				message := ""
				if config.GloablConfig.Mode == "native" {
					log.Infoln("[API请求]目录缓存刷新 >> 当前为本地模式，无需刷新")
					message = "Native mode does not need to be refreshed"
				} else {
					go updateCaches()
					log.Infoln("[API请求]目录缓存刷新 >> 请求刷新")
					message = "Cache update successful"
				}
				c.String(http.StatusOK, message)
			} else {
				message := "Invalid api token"
				c.String(http.StatusOK, message)
			}
		} else if method == "GET" && path == "/api/refreshCookie" {
			requestToken := c.Query("token")
			if requestToken == config.GloablConfig.ApiToken {
				message := ""
				if config.GloablConfig.Mode == "native" {
					log.Infoln("[API请求]目录缓存刷新 >> 当前为本地模式，无需刷新")
					message = "Native mode does not need to be refreshed"
				} else {
					go refreshCookie()
					log.Infoln("[API请求]cookie刷新 >> 请求刷新")
					message = "Cookie refresh successful"
				}
				c.String(http.StatusOK, message)
			} else {
				message := "Invalid api token"
				c.String(http.StatusOK, message)
			}
		} else if method == "GET" && path == "/api/shareToDown" {
			shareToDown(c)
		} else {
			isForbidden := true
			host := c.Request.Host
			referer, err := url.Parse(c.Request.Referer())
			if err != nil {
				log.Warningln(err)
			}
			if referer != nil && referer.Host != "" {
				if referer.Host == host {
					//站内，自动通过
					isForbidden = false
				} else if referer.Host != host && len(config.GloablConfig.OnlyReferer) > 0 {
					//外部引用，并且设置了防盗链，需要进行判断
					for _, rf := range config.GloablConfig.OnlyReferer {
						if rf == referer.Host {
							isForbidden = false
							break
						}
					}
				} else {
					isForbidden = false
				}
			} else {
				isForbidden = false
			}
			if isForbidden == true {
				c.String(http.StatusForbidden, "403 Hotlink Forbidden")
				return
			} else {
				index(c)
			}
		}
	})
	r.Run(fmt.Sprintf("%s:%d", config.GloablConfig.Host, config.GloablConfig.Port)) // 监听并在 0.0.0.0:8080 上启动服务

}

func initTemplates() *template.Template {
	tmpFile := strings.Join([]string{"pan/", "/index.html"}, config.GloablConfig.Theme)
	box := packr.New("templates", "./templates")
	data, _ := box.FindString(tmpFile)
	if Util.FileExist("./templates/" + tmpFile) {
		s, _ := ioutil.ReadFile("./templates/" + tmpFile)
		data = string(s)
	}
	tmpl := template.New(tmpFile)
	tmpl.Parse(data)
	return tmpl
}
func initStaticBox(r *gin.Engine) {
	staticBox := packr.New("static", "./static")
	if Util.FileExist("./static") {
		r.Static("/static", "./static")
	} else {
		r.StaticFS("/static", staticBox)
	}
}
func index(c *gin.Context) {
	tmpFile := strings.Join([]string{"pan/", "/index.html"}, config.GloablConfig.Theme)
	pwd := ""
	pwdCookie, err := c.Request.Cookie("dir_pwd")
	if err == nil {
		decodePwd, err := url.QueryUnescape(pwdCookie.Value)
		if err != nil {
			log.Warningln(err)
		}
		pwd = decodePwd
	}
	pathName := c.Request.URL.Path
	if pathName != "/" && pathName[len(pathName)-1:] == "/" {
		pathName = pathName[0 : len(pathName)-1]
	}
	result := service.GetFilesByPath(pathName, pwd)
	result["HerokuappUrl"] = config.GloablConfig.HerokuAppUrl
	result["Mode"] = config.GloablConfig.Mode
	result["PrePaths"] = Util.GetPrePath(result["Path"].(string))
	fs, ok := result["List"].([]entity.FileNode)
	if ok {
		if len(fs) == 1 && !fs[0].IsFolder && result["isFile"].(bool) {
			//文件
			if config.GloablConfig.Mode == "native" {
				c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fs[0].FileName))
				c.Writer.Header().Add("Content-Type", "application/octet-stream")
				c.File(fs[0].FileId)
				return
			} else {
				downUrl := service.GetDownlaodUrl(fs[0])
				c.Redirect(http.StatusFound, downUrl)
			}
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
	log.Infoln("[API请求]目录缓存刷新 >> 刷新成功")
}

func refreshCookie() {
	service.RefreshCookie()
	log.Infoln("[API请求]cookie刷新 >> 刷新成功")
}

func shareToDown(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET,HEAD,POST")
	c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")
	url := c.Query("url")
	passCode := c.Query("passCode")
	fileId := c.Query("fileId")
	subFileId := c.Query("subFileId")
	downUrl := Util.Cloud189shareToDown(url, passCode, fileId, subFileId)
	c.String(http.StatusOK, downUrl)
}
