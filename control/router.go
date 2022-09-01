package control

import (
	"github.com/gin-gonic/gin"
	"github.com/libsgh/PanIndex/control/middleware"
	"github.com/libsgh/PanIndex/module"
)

func SetRouters(r *gin.Engine) {
	jwt, _ := middleware.JWTMiddlewar()
	api := r.Group("/api/v3")
	public := api.Group("/public")
	{
		public.GET("/download/folder")                 //download folder
		public.POST("/files", middleware.Check, Files) //flie list by path
		public.GET("/file", func(c *gin.Context) {
			c.String(200, "api file")
		}) //flie by path
		public.POST("/transcode", AliTranscode)               //ali transcode stream
		public.POST("/shortInfo", ShortInfo)                  //short url && qrcode
		public.POST("/onedrive/exchangeToken", ExchangeToken) //onedrive exchange token
		public.POST("/onedrive/refreshToken")                 //onedrive refresh token
		public.GET("/raw/*path", middleware.Check, Raw)       //file original content
		public.GET("/config.js", ConfigJS)                    //file original content
	}
	adminApi := api.Group(module.GloablConfig.AdminPath, jwt.MiddlewareFunc())
	{
		adminApi.GET("/refresh_token", jwt.RefreshHandler)
		adminApi.POST("/config/upload", UploadConfig)          //upload config
		adminApi.POST("/config", SaveConfig)                   //save config
		adminApi.GET("/config", GetConfig)                     //get config info
		adminApi.GET("/account", GetAccount)                   //get account info
		adminApi.POST("/account", SaveAccount)                 //save account info
		adminApi.DELETE("/accounts", DeleteAccounts)           //del accounts
		adminApi.POST("/accounts/sort", SortAccounts)          //sort accounts
		adminApi.POST("/cache/update", UpdateCache)            //update cache
		adminApi.POST("/cache/update/batch", BatchUpdateCache) //batch update cache
		adminApi.POST("/refresh/login", RefreshLogin)          //refresh login status
		adminApi.POST("/upload", Upload)                       //simple upload file
		adminApi.POST("/hide/file", Hide)                      //add hide file
		adminApi.DELETE("/hide/file", DelHide)                 //del hide file by path
		adminApi.POST("/password/file", PwdFile)               //add pwd file
		adminApi.DELETE("/password/file", DelPwdFile)          //del pwd file by path
		adminApi.POST("/password/file/upload", UploadPwdFile)  //upload pwd file
		adminApi.POST("/password/file/share/info", ShareInfo)  //upload pwd file
		adminApi.POST("/bypass", Bypass)                       //save bypass config
		adminApi.DELETE("/bypass", DelBypass)                  //del bypass config
		adminApi.GET("/bypass", GetBypass)                     //get bypass by account
		adminApi.GET("/cache", GetCache)                       //get file cache data
		adminApi.POST("/cache/clear", CacheClear)              //clear file cache
		adminApi.POST("/cache/config", CacheConfig)            //save cache config
		adminApi.GET("/ali/qrcode", AliQrcode)                 //ali qrcode
		adminApi.POST("/ali/qrcode/check", AliQrcodeCheck)     //ali qrcode check
	}

	admin := r.Group(module.GloablConfig.AdminPath)
	{
		admin.POST("/login", jwt.LoginHandler)  //login
		admin.GET("/logout", jwt.LogoutHandler) //logout
		auth := admin.Use(jwt.MiddlewareFunc())
		auth.GET("", AdminIndex)
		auth.GET("/", AdminIndex)
		auth.GET("/common", ConfigManagent)     //base config
		auth.GET("/appearance", ConfigManagent) //appearance
		auth.GET("/view", ConfigManagent)       //view config
		auth.GET("/pwd", ConfigManagent)        //pwd file config
		auth.GET("/hide", ConfigManagent)       //hide file config
		auth.GET("/safety", ConfigManagent)     //safety
		auth.GET("/disk", ConfigManagent)       //bind net disk
		auth.GET("/bypass", ConfigManagent)     //bypass download
		auth.GET("/cache", ConfigManagent)      //cache
		auth.GET("/webdav", ConfigManagent)     //webdav
		auth.GET("/access", ConfigManagent)     //access
	}
	r.GET("/s/*shortCode", func(context *gin.Context) {
		ShortRedirect(context, r)
	})
	dav := r.Group(module.GloablConfig.DavPath, WebDAVAuth())
	{
		dav.Any("/*path", ServeWebDAV)
		dav.Handle("PROPFIND", "/*path", ServeWebDAV)
		dav.Handle("PROPFIND", "", ServeWebDAV)
		dav.Handle("MKCOL", "/*path", ServeWebDAV)
		dav.Handle("LOCK", "/*path", ServeWebDAV)
		dav.Handle("UNLOCK", "/*path", ServeWebDAV)
		dav.Handle("PROPPATCH", "/*path", ServeWebDAV)
		dav.Handle("COPY", "/*path", ServeWebDAV)
		dav.Handle("MOVE", "/*path", ServeWebDAV)
	}
	if module.GloablConfig.Access == "3" {
		r.NoRoute(middleware.Check, jwt.MiddlewareFunc(), func(c *gin.Context) {
			claim, err := jwt.CheckIfTokenExpire(c)
			isAdminLogin := false
			if err != nil && claim["id"] == module.GloablConfig.AdminUser {
				isAdminLogin = true
			}
			index(c, isAdminLogin)
		})
	} else {
		r.NoRoute(middleware.Check, func(c *gin.Context) {
			claim, err := jwt.CheckIfTokenExpire(c)
			isAdminLogin := false
			if err == nil && claim["id"] == module.GloablConfig.AdminUser {
				isAdminLogin = true
			}
			index(c, isAdminLogin)
		})
	}
}
