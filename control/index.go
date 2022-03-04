package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libsgh/PanIndex/control/middleware"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/pan"
	"github.com/libsgh/PanIndex/service"
	"github.com/libsgh/PanIndex/util"
	"net/http"
	"net/url"
	"strings"
)

func index(c *gin.Context) {
	var fns []module.FileNode
	var isFile bool
	tmpFile := strings.Join([]string{"templates/pan/", "/index.html"}, module.GloablConfig.Theme)
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
	if isSearch {
		fns = service.Search(searchKey)
	} else {
		if module.GloablConfig.AccountChoose == "display" && fullPath == "/" {
			//返回账号列表
			fns = service.AccountsToNodes()
		} else {
			fns, isFile, lastFile, nextFile = service.Index(ac, path, fullPath, sortColumn, sortOrder, isView)
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
	c.HTML(http.StatusOK, tmpFile, gin.H{
		"title":        CurrentTitle(ac, module.GloablConfig, bypassName),
		"path":         path,
		"full_path":    fullPath,
		"account":      ac,
		"accounts":     service.GetAccounts(),
		"config":       module.GloablConfig,
		"pwd_err_msg":  c.GetString("pwd_err_msg"),
		"has_pwd":      c.GetBool("has_pwd"),
		"pwd_path":     pwdPath,
		"has_parent":   hasParent,
		"parent_path":  parentPath,
		"account_path": CurrentAccountPath(ac.Name, bypassName),
		"search_key":   searchKey,
		"pre_paths":    util.GetPrePath(fullPath),
		"fns":          fns,
		"theme":        theme,
		"version":      module.VERSION,
		"layout":       layout,
		"last_file":    lastFile,
		"next_file":    nextFile,
	})
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
			c.FileAttachment(downUrl, fileNode.FileName)
		} else if ac.Mode == "ftp" {
			service.FtpDownload(ac, downUrl, fileNode, c)
		} else if ac.Mode == "webdav" {
			service.WebDavDownload(ac, downUrl, fileNode, c)
		}
	}
}

func DataRroxy(ac module.Account, downUrl, fileName string, c *gin.Context) {
	client := util.GetClient(20)
	req, err := http.NewRequest("GET", downUrl, nil)
	req.Header.Add("Range", c.GetHeader("Range"))
	if ac.Mode == "googledrive" {
		req.Header.Add("Authorization", "Bearer "+pan.GoogleDrives[ac.Id].AccessToken)
	}
	response, err := client.Do(req)
	defer response.Body.Close()
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
	theme := strings.ReplaceAll(module.GloablConfig.Theme, "-dark", "")
	theme = strings.ReplaceAll(theme, "-light", "")
	*tmpFile = fmt.Sprintf("templates/pan/%s/view-%s.html", theme, viewType)
}

func ShortRedirect(c *gin.Context) {
	pathName := c.Request.URL.Path
	//_, isView := c.GetQuery("v")
	if pathName != "/" && pathName[len(pathName)-1:] == "/" {
		pathName = pathName[0 : len(pathName)-1]
	}
	paths := strings.Split(pathName, "/")
	if len(paths) == 3 && paths[1] == "s" {
		redirectUri := service.GetRedirectUri(paths[2])
		c.Redirect(http.StatusFound, redirectUri)
	}
	c.Abort()
}
