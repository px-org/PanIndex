package control

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/PanIndex/dao"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/pan"
	"github.com/libsgh/PanIndex/service"
	"github.com/libsgh/PanIndex/util"
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

// short url & qrcode
func ShortInfo(c *gin.Context) {
	path := c.PostForm("path")
	prefix := c.PostForm("prefix")
	url, qrCode, msg := service.ShortInfo(path, prefix)
	c.JSON(http.StatusOK, gin.H{
		"short_url": url,
		"qr_code":   qrCode,
		"msg":       msg,
	})
}

// aliyundrive transcode
func AliTranscode(c *gin.Context) {
	accountId := c.Query("accountId")
	fileId := c.Query("fileId")
	account := dao.GetAccountById(accountId)
	p, _ := pan.GetPan(account.Mode)
	result, _ := p.(*pan.Ali).Transcode(account, fileId)
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
		pa, _ := pan.GetPan(account.Mode)
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
	config, _ := jsoniter.MarshalToString(gin.H{"path_prefix": module.GloablConfig.PathPrefix})
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
