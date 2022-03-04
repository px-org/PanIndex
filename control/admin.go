package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libsgh/PanIndex/control/middleware"
	"github.com/libsgh/PanIndex/dao"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/pan"
	"github.com/libsgh/PanIndex/service"
	"github.com/libsgh/PanIndex/util"
	"net/http"
	"net/url"
	"strings"
)

func AdminIndex(c *gin.Context) {
	c.Redirect(http.StatusFound, module.GloablConfig.AdminPath+"/common")
}

//admin config managent
func ConfigManagent(c *gin.Context) {
	middleware.ThemeCheck(c)
	theme := c.GetString("theme")
	fullPath := c.Request.URL.Path
	adminModule := strings.Split(fullPath, "/")[2]
	var cacheData []module.Cache
	searchKey := ""
	if adminModule == "cache" {
		path := c.Query("path")
		pathEsc, _ := url.QueryUnescape(path)
		cacheData = service.GetCacheData(pathEsc)
		searchKey = path
	}
	template := fmt.Sprintf("templates/pan/admin/%s.html", adminModule)
	configData := module.GloablConfig
	c.HTML(http.StatusOK, template, gin.H{
		"config":       configData,
		"cache":        cacheData,
		"search_key":   searchKey,
		"redirect_url": adminModule,
		"version":      module.VERSION,
		"theme":        theme,
	})
}

// admin save config
func SaveConfig(c *gin.Context) {
	configMap := make(map[string]string)
	c.BindJSON(&configMap)
	dao.UpdateConfig(configMap)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "配置已更新，部分配置重启后生效！"})
}

// admin get config
func GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, module.GloablConfig)
}

// admin get account
func GetAccount(c *gin.Context) {
	id := c.Query("id")
	c.JSON(http.StatusOK, dao.GetAccountById(id))
}

// admin del accounts
func DeleteAccounts(c *gin.Context) {
	delIds := []string{}
	c.BindJSON(&delIds)
	dao.DeleteAccounts(delIds)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "删除成功！"})
}

// admin sort accounts
func SortAccounts(c *gin.Context) {
	sortIds := []string{}
	c.BindJSON(&sortIds)
	dao.SortAccounts(sortIds)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "排序完成！"})
}

// admin update cache
func UpdateCache(c *gin.Context) {
	accountId := c.PostForm("accountId")
	cachePath := c.PostForm("cachePath")
	account := dao.GetAccountById(accountId)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": service.UpdateCache(account, cachePath)})
}

// admin refresh login status
func RefreshLogin(c *gin.Context) {
	id := c.Query("id")
	account := dao.GetAccountById(id)
	if account.CookieStatus != -1 {
		go dao.SyncAccountStatus(account)
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "登录状态刷新中，请稍后查看结果"})
}

// admin upload file
func Upload(c *gin.Context) {
	accountId := c.PostForm("uploadAccount")
	path := c.PostForm("uploadPath")
	t := c.PostForm("type")
	msg := ""
	if t == "0" {
		msg = service.Upload(accountId, path, c)
	} else if t == "1" {
		service.Async(accountId, path)
		msg = "刷新缓存成功"
	} else if t == "2" {
		service.Upload(accountId, path, c)
		service.Async(accountId, path)
		msg = "上传并刷新成功"
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": msg})
}

//add hide file
func Hide(c *gin.Context) {
	data := gin.H{}
	c.BindJSON(&data)
	dao.SaveHideFile(data["hide_path"].(string))
	parentPath := util.GetParentPath(data["hide_path"].(string))
	service.ClearFileCache(parentPath)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "添加成功！"})
}

//del hide file by path
func DelHide(c *gin.Context) {
	delPaths := []string{}
	c.BindJSON(&delPaths)
	dao.DeleteHideFiles(delPaths)
	for _, p := range delPaths {
		parentPath := util.GetParentPath(p)
		service.ClearFileCache(parentPath)
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "删除成功！"})
}

//save pwd file
func PwdFile(c *gin.Context) {
	pwdFiles := module.PwdFiles{}
	c.BindJSON(&pwdFiles)
	dao.SavePwdFile(pwdFiles)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "保存成功！"})
}

//del hide file by path
func DelPwdFile(c *gin.Context) {
	delPaths := []string{}
	c.BindJSON(&delPaths)
	dao.DeletePwdFiles(delPaths)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "删除成功！"})
}

//save bypass
func Bypass(c *gin.Context) {
	bypass := module.Bypass{}
	c.BindJSON(&bypass)
	msg := dao.SaveBypass(bypass)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": msg})
}

//del bypass
func DelBypass(c *gin.Context) {
	delIds := []string{}
	c.BindJSON(&delIds)
	dao.DeleteBypass(delIds)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "删除成功！"})
}

//get file cache
func GetCache(c *gin.Context) {
	path := c.Query("path")
	pathEsc, _ := url.QueryUnescape(path)
	if path == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "缺少路径参数！"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": 0, "msg": service.GetCacheByPath(pathEsc)})
	}
}

//get file cache
func CacheClear(c *gin.Context) {
	data := gin.H{}
	c.BindJSON(&data)
	path := data["path"].(string)
	isLoopChildren := data["is_loop_children"].(string)
	pathEsc, _ := url.QueryUnescape(path)
	if path == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "缺少路径参数！"})
	} else {
		service.CacheClear(pathEsc, isLoopChildren)
		c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "清理成功"})
	}
}

//get bypass by account
func GetBypass(c *gin.Context) {
	accountId := c.Query("account_id")
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "成功", "data": service.GetBypassByAccountId(accountId)})
}

func CacheConfig(c *gin.Context) {
	data := module.Account{}
	c.BindJSON(&data)
	dao.UpdateCacheConfig(data)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "配置成功！"})
}

func AliQrcode(c *gin.Context) {
	qr, data := pan.QrcodeGen()
	c.JSON(http.StatusOK, gin.H{
		"qr":    qr,
		"param": data,
	})
}

func AliQrcodeCheck(c *gin.Context) {
	t := c.PostForm("t")
	codeContent := c.PostForm("codeContent")
	ck := c.PostForm("ck")
	resultCode := c.PostForm("resultCode")
	qrCodeStatus, refreshToken := pan.QrcodeCheck(t, codeContent, ck, resultCode)
	c.JSON(http.StatusOK, gin.H{
		"qrCodeStatus": qrCodeStatus,
		"refreshToken": refreshToken,
	})
}

func SaveAccount(c *gin.Context) {
	data := module.Account{}
	c.BindJSON(&data)
	msg := dao.SaveAccount(data)
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": msg})
}
