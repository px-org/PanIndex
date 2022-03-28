package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func Check(c *gin.Context) {
	fullPath := c.Request.URL.Path
	account, fullPath, path, bypassName := ParseFullPath(fullPath)
	c.Set("account", account)
	c.Set("path", path)
	c.Set("full_path", fullPath)
	c.Set("bypass_name", bypassName)
	PwdCheck(c, fullPath)
	SortCheck(c)
	ThemeCheck(c)
	LayoutCheck(c)
	if len(module.GloablConfig.Accounts) == 0 {
		c.Redirect(http.StatusFound, module.GloablConfig.AdminPath)
	}
	c.Next()
}

func SortCheck(c *gin.Context) {
	sColumn := ""
	sColumnCookie, err := c.Request.Cookie("sort_column")
	if err != nil {
		sColumn = module.GloablConfig.SColumn
	} else {
		sColumn = sColumnCookie.Value
	}
	c.Set("sort_column", sColumn)
	sOrder := ""
	sOrderCookie, err := c.Request.Cookie("sort_order")
	if err != nil {
		sOrder = module.GloablConfig.SOrder
	} else {
		sOrder = sOrderCookie.Value
	}
	c.Set("sort_order", sOrder)
}

func PwdCheck(c *gin.Context, fullPath string) {
	filePwd, ok := util.GetPwdFromPath(fullPath)
	if ok {
		pwdCookie, err := c.Request.Cookie("file_pwd")
		if err != nil {
			//redirect input pwd
			c.Set("pwd_err_msg", "")
			c.Set("has_pwd", true)
		} else {
			result, msg := VerifyPwd(filePwd.Password, filePwd.FilePath, pwdCookie.Value)
			c.Set("pwd_err_msg", msg)
			c.Set("has_pwd", result)
		}
		c.Set("pwd_path", filePwd.FilePath)
	} else {
		c.Set("pwd_err_msg", "")
		c.Set("has_pwd", false)
		c.Set("pwd_path", "")
	}
}

func VerifyPwd(pwd, path, cookiepwd string) (bool, string) {
	inputPwd := GetPwdFromCookie(cookiepwd, path)
	if inputPwd == "" {
		return true, ""
	} else if inputPwd != pwd {
		return true, "密码错误"
	} else {
		return false, ""
	}
}

func ThemeCheck(c *gin.Context) {
	theme, err := c.Request.Cookie("theme")
	if err != nil {
		c.Set("theme", module.GloablConfig.Theme)
	} else {
		c.Set("theme", theme.Value)
	}
}

func AdminThemeCheck(c *gin.Context) {
	theme, err := c.Request.Cookie("theme")
	if err != nil {
		if strings.HasPrefix(module.GloablConfig.Theme, "mdui") {
			c.Set("theme", module.GloablConfig.Theme)
		} else {
			c.Set("theme", "mdui")
		}
	} else {
		if strings.HasPrefix(theme.Value, "mdui") {
			c.Set("theme", theme.Value)
		} else {
			c.Set("theme", "mdui")
		}
	}
}

func LayoutCheck(c *gin.Context) {
	layout, err := c.Request.Cookie("layout")
	if err != nil {
		c.Set("layout", "view_comfy")
	} else {
		c.Set("layout", layout.Value)
	}
}

func GetPwdFromCookie(pwd, fullPath string) string {
	pathMd5 := util.Md5(fullPath)
	decodedValue, _ := url.QueryUnescape(pwd)
	s := strings.Split(decodedValue, ",")
	if len(s) > 0 {
		for _, v := range s {
			if strings.Split(v, ":")[0] == pathMd5 {
				return strings.Split(v, ":")[1]
			}
		}
	}
	return ""
}

func ParseFullPath(path string) (module.Account, string, string, string) {
	if strings.HasPrefix(path, "/d_") {
		//old path
		path = OldParseFullPath(path)
	}
	if path == "" {
		path = "/"
	}
	if path == "/" && module.GloablConfig.AccountChoose == "default" && len(module.GloablConfig.BypassList) > 0 {
		path = "/" + module.GloablConfig.BypassList[0].Name
	} else {
		if path == "/" && module.GloablConfig.AccountChoose == "default" && len(module.GloablConfig.Accounts) > 0 {
			path = "/" + module.GloablConfig.Accounts[0].Name
		}
	}
	fullPath := path
	if path != "/" && path[len(path)-1:] == "/" {
		path = path[0 : len(path)-1]
	}
	paths := strings.Split(path, "/")
	accountName := paths[1]
	account, bypassName := GetCurrentAccount(accountName)
	path = strings.Join(paths[2:], "/")
	path = "/" + path
	if fullPath != "/" && fullPath[len(fullPath)-1:] == "/" {
		fullPath = fullPath[0 : len(fullPath)-1]
	}
	return account, fullPath, path, bypassName
}

func OldParseFullPath(path string) string {
	iStr := util.GetBetweenStr(path, "_", "/")
	index, _ := strconv.Atoi(iStr)
	account := module.GloablConfig.Accounts[index]
	return strings.ReplaceAll(path, "/d_"+iStr, "/"+account.Name)
}

func GetCurrentAccount(name string) (module.Account, string) {
	//get account from bypass
	var bypass module.Bypass
	if len(module.GloablConfig.BypassList) > 0 {
		if name == "" {
			bypass = module.GloablConfig.BypassList[0]
		} else {
			for _, item := range module.GloablConfig.BypassList {
				if item.Name == name {
					bypass = item
					break
				}
			}
		}
		if bypass.Name != "" {
			//get round robin account
			bypassAccount := bypass.Rw.Next().(module.Account)
			log.Debugf("access bypass account: %s", bypassAccount.Name)
			return bypassAccount, bypass.Name
		}
	}
	//get account from accounts
	var account module.Account
	if name == "" {
		if len(module.GloablConfig.Accounts) > 0 {
			account = module.GloablConfig.Accounts[0]
		} else {
			account = module.Account{}
		}
	} else {
		for _, ac := range module.GloablConfig.Accounts {
			if ac.Name == name {
				account = ac
				break
			}
		}
	}
	return account, bypass.Name
}
