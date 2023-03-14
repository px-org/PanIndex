package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/px-org/PanIndex/boot"
	"github.com/px-org/PanIndex/control"
	"github.com/px-org/PanIndex/control/middleware"
	_ "github.com/px-org/PanIndex/pan/115"
	_ "github.com/px-org/PanIndex/pan/123"
	_ "github.com/px-org/PanIndex/pan/ali"
	_ "github.com/px-org/PanIndex/pan/alishare"
	_ "github.com/px-org/PanIndex/pan/cloud189"
	_ "github.com/px-org/PanIndex/pan/ftp"
	_ "github.com/px-org/PanIndex/pan/googledrive"
	_ "github.com/px-org/PanIndex/pan/native"
	_ "github.com/px-org/PanIndex/pan/onedrive"
	_ "github.com/px-org/PanIndex/pan/pikpak"
	_ "github.com/px-org/PanIndex/pan/s3"
	_ "github.com/px-org/PanIndex/pan/teambition"
	_ "github.com/px-org/PanIndex/pan/webdav"
	_ "github.com/px-org/PanIndex/pan/yun139"
)

//go:embed templates static
var fs embed.FS

func main() {
	//boot init
	config, result := boot.Init()
	if result {
		return
	}
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.Use(middleware.Cors())
	r.Use(gin.Logger(), middleware.RequestCancelRecover())
	//set html templates
	r.SetHTMLTemplate(boot.Templates(fs, config))
	//set static box
	boot.InitStaticBox(r, fs)
	//set all routers
	control.SetRouters(r)
	if config.CertFile != "" && config.KeyFile != "" {
		//enable https
		r.Use(boot.TlsHandler(config.Port))
		r.RunTLS(fmt.Sprintf("%s:%d", config.Host, config.Port), config.CertFile, config.KeyFile)
	} else {
		r.Run(fmt.Sprintf("%s:%d", config.Host, config.Port))
	}
}
