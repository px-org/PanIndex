package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/px-org/PanIndex/control/middleware"
	"github.com/px-org/PanIndex/module"
	_115 "github.com/px-org/PanIndex/pan/115"
	"github.com/px-org/PanIndex/pan/googledrive"
	"github.com/px-org/PanIndex/service"
	"github.com/px-org/PanIndex/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
)

func index(c *gin.Context, isAdminLogin bool) {
	var fns []module.FileNode
	var isFile bool
	tmpFile := strings.Join([]string{"templates/pan/", "/index.html"}, util.GetCurrentTheme(module.GloablConfig.Theme))
	account, _ := c.Get("account")
	sortColumn := c.GetString("sort_column")
	sortOrder := c.GetString("sort_order")
	ac := account.(module.Account)
	path := c.GetString("path")
	fullPath := c.GetString("full_path")
	theme := c.GetString("theme")
	bypassName := c.GetString("bypass_name")
	pwdPath := c.GetString("pwd_path")
	layout := c.GetString("layout")
	_, isView := c.GetQuery("v")
	searchKey, isSearch := c.GetQuery("search")
	var lastFile, nextFile = "", ""
	status := http.StatusOK
	if isSearch {
		fns = service.Search(searchKey)
		t := Redirect404(c, false, isAdminLogin)
		if t != "" {
			tmpFile = t
		}
	} else {
		if module.GloablConfig.AccountChoose == "display" && fullPath == "/" {
			//返回账号列表
			fns = service.AccountsToNodes(c.Request.Host)
		} else {
			fns, isFile, lastFile, nextFile = service.Index(ac, path, fullPath, sortColumn, sortOrder, isView)
		}
		t := Redirect404(c, isFile, isAdminLogin)
		if t != "" {
			tmpFile = t
		}
		if isFile {
			if isView {
				view(&tmpFile, fns[0].ViewType)
			} else {
				download(ac, fns[0], c)
				c.Abort()
				return
			}
		}
	}
	hasParent, parentPath := service.HasParent(fullPath)
	c.HTML(status, tmpFile, gin.H{
		"title":          CurrentTitle(ac, module.GloablConfig, bypassName),
		"path":           path,
		"full_path":      fullPath,
		"account":        ac,
		"accounts":       service.GetAccounts(),
		"config":         module.GloablConfig,
		"pwd_err_msg":    c.GetString("pwd_err_msg"),
		"has_pwd":        c.GetBool("has_pwd"),
		"pwd_path":       pwdPath,
		"has_parent":     hasParent,
		"parent_path":    parentPath,
		"account_path":   CurrentAccountPath(ac.Name, bypassName),
		"search_key":     searchKey,
		"pre_paths":      util.GetPrePath(fullPath),
		"fns":            fns,
		"theme":          theme,
		"version":        module.VERSION,
		"layout":         layout,
		"last_file":      lastFile,
		"next_file":      nextFile,
		"is_admin_login": isAdminLogin,
	})
}

func Redirect404(c *gin.Context, flag bool, isAdminLogin bool) string {
	_, isView := c.GetQuery("v")
	_, isSearch := c.GetQuery("search")
	t := "templates/pan/admin/404.html"
	if module.GloablConfig.Access == "0" {
		//公开
		c.Next()
	} else if module.GloablConfig.Access == "1" {
		//仅直链
		if !isAdminLogin && (isView || isSearch || !flag) {
			c.Abort()
			return t
		}
	} else if !isAdminLogin && (module.GloablConfig.Access == "2") {
		//直链 + 预览
		if isSearch || !flag {
			c.Abort()
			return t
		}
	} else if module.GloablConfig.Access == "3" {
		//登录
	}
	return ""
}

func CurrentTitle(ac module.Account, config module.Config, bypassName string) string {
	if config.SiteName != "" {
		return config.SiteName
	}
	if bypassName != "" {
		return bypassName
	}
	return ac.Name
}

func CurrentAccountPath(accountName, bypassName string) string {
	if bypassName != "" {
		return "/" + bypassName
	}
	return "/" + accountName
}

func download(ac module.Account, fileNode module.FileNode, c *gin.Context) {
	isForbidden := middleware.CheckReferer(c)
	if isForbidden {
		c.String(http.StatusForbidden, "403 Hotlink Forbidden")
		c.Abort()
		return
	}
	if c.GetBool("has_pwd") {
		c.String(http.StatusForbidden, "401 Unauthorized")
		return
	}
	downUrl := service.GetDownloadUrl(ac, fileNode.FileId)
	if strings.HasPrefix(downUrl, "http") {
		if ac.DownTransfer == 1 {
			if ac.TransferDomain != "" {
				u, _ := url.Parse(downUrl)
				domain := util.GetTransferDomain(ac.TransferDomain, u.Host)
				u.Host = domain
				downUrl = u.String()
				c.Redirect(http.StatusFound, downUrl)
			} else {
				DataRroxy(ac, downUrl, fileNode.FileName, c)
			}
		} else {
			c.Redirect(http.StatusFound, downUrl)
		}
	} else {
		if ac.Mode == "native" {
			c.FileAttachment(downUrl, url.QueryEscape(fileNode.FileName))
		} else if ac.Mode == "ftp" {
			service.FtpDownload(ac, downUrl, fileNode, c)
		} else if ac.Mode == "webdav" {
			service.WebDavDownload(ac, downUrl, fileNode, c)
		}
	}
}

func DataRroxy(ac module.Account, downUrl, fileName string, c *gin.Context) {
	client := util.GetClient(0)
	req, err := http.NewRequest("GET", downUrl, nil)
	reqRange := c.GetHeader("Range")
	if reqRange != "" {
		req.Header.Add("Range", c.GetHeader("Range"))
	}
	if ac.Mode == "googledrive" {
		req.Header.Add("Authorization", "Bearer "+googledrive.GoogleDrives[ac.Id].AccessToken)
	} else if ac.Mode == "115" {
		req.Header.Add("Cookie", ac.Password)
		req.Header.Add("User-Agent", _115.UA)
	}
	response, err := client.Do(req)
	defer func() {
		err = response.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	reader := response.Body
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")
	fileName = url.QueryEscape(fileName)
	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="` + fileName + `"`,
		"Accept-Ranges":       "bytes",
		"Content-Range":       response.Header.Get("Content-Range"),
	}
	c.DataFromReader(response.StatusCode, contentLength, contentType, reader, extraHeaders)
}

func view(tmpFile *string, viewType string) {
	if !strings.Contains(*tmpFile, "404.html") {
		*tmpFile = fmt.Sprintf("templates/pan/%s/view-%s.html", util.GetCurrentTheme(module.GloablConfig.Theme), viewType)
	}
}

func ShortRedirect(c *gin.Context, r *gin.Engine) {
	pathName := c.Request.URL.Path
	if pathName != "/" && pathName[len(pathName)-1:] == "/" {
		pathName = pathName[0 : len(pathName)-1]
	}
	paths := strings.Split(pathName, "/")
	if len(paths) == 3 && paths[1] == "s" {
		redirectUri, v := service.GetRedirectUri(paths[2])
		if redirectUri == "" {
			t := "templates/pan/admin/404.html"
			c.HTML(http.StatusOK, t, gin.H{})
		} else {
			c.Request.URL.Path = redirectUri
			q := c.Request.URL.Query()
			q.Add(v, "")
			c.Request.URL.RawQuery = q.Encode()
			r.HandleContext(c)
		}
	}
	c.Abort()
}
