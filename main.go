package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libsgh/PanIndex/boot"
	"github.com/libsgh/PanIndex/control"
	"github.com/libsgh/PanIndex/control/middleware"
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
