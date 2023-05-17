package control

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/control/middleware"
	"github.com/px-org/PanIndex/dao"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/ali"
	_alishare "github.com/px-org/PanIndex/pan/alishare"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/service"
	"github.com/px-org/PanIndex/util"
	"net/http"
	"strings"
)

func ExchangeToken(c *gin.Context) {
	/*clientId := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")
	code := c.PostForm("code")
	redirectUri := c.PostForm("redirect_uri")
	zone := c.PostForm("zone")
	tokenInfo := Util.OneExchangeToken(zone, clientId, redirectUri, clientSecret, code)
	c.String(http.StatusOK, tokenInfo)*/
}

// aliyundrive transcode
func AliTranscode(c *gin.Context) {
	accountId := c.Query("accountId")
	fileId := c.Query("fileId")
	account := dao.GetAccountById(accountId)
	p, _ := base.GetPan(account.Mode)
	var result string
	if pan, ok := p.(*ali.Ali); ok {
		result, _ = pan.Transcode(account, fileId)
	} else {
		result, _ = p.(*_alishare.AliShare).Transcode(account, fileId)
	}
	c.String(http.StatusOK, result)
	c.Abort()
}

func Raw(c *gin.Context) {
	p := c.Param("path")
	if strings.HasPrefix(p, module.GloablConfig.PathPrefix) {
		p = strings.TrimPrefix(p, module.GloablConfig.PathPrefix)
	}
	hasPwd := c.GetBool("has_pwd")
	if hasPwd {
		CommonResp(c, "unauthorized", 401, nil)
		return
	}
	account, fullpath, path, _ := util.ParseFullPath(p, "")
	fileName := util.GetFileName(fullpath)
	fileId := service.GetFileIdByPath(account, path, fullpath)
	downloadUrl := service.GetDownloadUrl(account, fileId)
	if fileId == "" || downloadUrl == "" {
		CommonResp(c, "file not found", 404, nil)
		return
	}
	if strings.HasPrefix(downloadUrl, "http") {
		DataRroxy(account, downloadUrl, fileName, c)
	} else {
		pa, _ := base.GetPan(account.Mode)
		fileNode, _ := pa.File(account, fileId, fullpath)
		if account.Mode == "ftp" {
			service.FtpDownload(account, downloadUrl, fileNode, c)
		} else if account.Mode == "webdav" {
			service.WebDavDownload(account, downloadUrl, fileNode, c)
		} else {
			c.FileAttachment(downloadUrl, fileName)
		}
	}
}

func ConfigJS(c *gin.Context) {
	config, _ := jsoniter.MarshalToString(gin.H{
		"path_prefix": module.GloablConfig.PathPrefix,
		"admin_path":  module.GloablConfig.AdminPath,
	})
	c.String(http.StatusOK, `var $config=%s;`, config)
}

func Files(c *gin.Context) {
	hasPwd := c.GetBool("has_pwd")
	if hasPwd {
		CommonResp(c, "unauthorized", 401, nil)
		return
	}
	path := c.PostForm("path")
	if strings.HasPrefix(path, module.GloablConfig.PathPrefix) {
		path = strings.TrimPrefix(path, module.GloablConfig.PathPrefix)
	}
	viewType := c.PostForm("viewType")
	sColumn := c.PostForm("sortColumn")
	sOrder := c.PostForm("sortOrder")
	account, fullPath, p, _ := util.ParseFullPath(path, "")
	files := service.GetFiles(account, p, fullPath, sColumn, sOrder, viewType)
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"status":  0,
		"data":    files,
	})
}

func CommonResp(c *gin.Context, msg string, code int, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": code,
		"msg":    msg,
		"data":   data,
	})
	c.Abort()
}

func CommonSuccessResp(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    msg,
		"data":   data,
	})
	c.Abort()
}

func ConfigJson(c *gin.Context) {
	CommonSuccessResp(c, "success", gin.H{
		"site_name":      module.GloablConfig.SiteName,
		"theme":          module.GloablConfig.Theme,
		"account_choose": module.GloablConfig.AccountChoose,
		"s_column":       module.GloablConfig.SColumn,
		"s_order":        module.GloablConfig.SOrder,
		"path_prefix":    module.GloablConfig.PathPrefix,
		"favicon_url":    module.GloablConfig.FaviconUrl,
		"footer":         module.GloablConfig.Footer,
		"css":            module.GloablConfig.Css,
		"js":             module.GloablConfig.Js,
		"readme":         module.GloablConfig.Readme,
		"head":           module.GloablConfig.Head,
		"image":          module.GloablConfig.Image,
		"video":          module.GloablConfig.Video,
		"audio":          module.GloablConfig.Audio,
		"doc":            module.GloablConfig.Doc,
		"code":           module.GloablConfig.Code,
		"short_action":   module.GloablConfig.ShortAction,
	})
}

func AccountList(c *gin.Context) {
	accounts := module.GloablConfig.Accounts
	acs := []gin.H{}
	for _, account := range accounts {
		acs = append(acs, gin.H{
			"name": account.Name,
			"path": "/" + account.Name,
			"mode": account.Mode,
		})
	}
	CommonSuccessResp(c, "success", acs)
}

func IndexData(c *gin.Context) {
	var fns []module.FileNode
	var isFile bool
	var lastFile, nextFile = "", ""
	path := c.PostForm("path")
	sortBy := c.PostForm("sort_by")
	order := c.PostForm("order")
	if strings.HasPrefix(path, module.GloablConfig.PathPrefix) {
		path = strings.TrimPrefix(path, module.GloablConfig.PathPrefix)
	}
	ac, fullPath, path, _ := util.ParseFullPath(path, "")
	middleware.PwdCheck(c, fullPath)
	if module.GloablConfig.AccountChoose == "display" && fullPath == "/" {
		//return account list
		fns = service.AccountsToNodes(c.Request.Host)
	} else {
		fns, isFile, lastFile, nextFile = service.Index(ac, path, fullPath, sortBy, order, true)
	}
	noReferrer := false
	if ac.Mode == "aliyundrive" || ac.Mode == "aliyundrive-share" {
		noReferrer = true
	}
	if c.GetBool("has_pwd") {
		c.JSON(http.StatusOK, gin.H{
			"status": 403,
			"msg":    c.GetString("pwd_err_msg"),
			"data": gin.H{
				"is_folder":   !isFile,
				"content":     []module.FileNode{},
				"no_referrer": noReferrer,
				"last_file":   lastFile,
				"next_file":   nextFile,
				"pwd_path":    c.GetString("pwd_path"),
			},
		})
		c.Abort()
		return
	}
	CommonSuccessResp(c, "success", gin.H{
		"is_folder":   !isFile,
		"content":     fns,
		"no_referrer": noReferrer,
		"last_file":   lastFile,
		"next_file":   nextFile,
		"page_no":     1,
		"page_size":   10,
		"pages":       1,
	})
}
func SearchData(c *gin.Context) {
	key := c.PostForm("key")
	fns := service.Search(key)
	CommonSuccessResp(c, "success", gin.H{
		"content": fns,
	})
}

func Info(c *gin.Context) {
	CommonSuccessResp(c, "success", gin.H{
		"name":       "PanIndex",
		"version":    module.VERSION,
		"commit_sha": module.GIT_COMMIT_SHA,
		"author":     "Libs",
	})
}

func ShortRedirectInfo(c *gin.Context) {
	shortCode := c.PostForm("short_code")
	redirectUri, v := service.GetRedirectUri(shortCode)
	CommonSuccessResp(c, "success", gin.H{
		"redirectUri": redirectUri,
		"v":           v,
	})
}
