package control

import (
	"github.com/gin-gonic/gin"
	"github.com/px-org/PanIndex/control/webdav"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/util"
	"net/http"
)

func WebDAVAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if module.GloablConfig.EnableDav == "0" {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		if module.GloablConfig.DavMode == "0" && (c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodDelete ||
			c.Request.Method == "COPY" ||
			c.Request.Method == "MOVE") {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Writer.Header()["WWW-Authenticate"] = []string{`Basic realm="PanIndex"`}
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		if username != module.GloablConfig.DavUser || password != module.GloablConfig.DavPassword {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

func ServeWebDAV(c *gin.Context) {
	//not support bypass
	p := c.Param("path")
	account, fullPath, path, _ := util.ParseFullPath(p, "")
	handler := &webdav.Handler{
		Prefix:     "/",
		FileSystem: webdav.FileSystem{},
		LockSystem: webdav.NewMemLS(),
		Account:    account,
		FullPath:   fullPath,
		Path:       path,
	}
	handler.ServeHTTP(c.Writer, c.Request)
}
