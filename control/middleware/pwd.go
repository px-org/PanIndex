package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/px-org/PanIndex/dao"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/util"
	"net/http"
	"net/url"
	"strings"
)

func Check(c *gin.Context) {
	fullPath := c.Request.URL.Path
	if strings.HasPrefix(fullPath, "/api/v3/public/raw") {
		fullPath = strings.ReplaceAll(fullPath, "/api/v3/public/raw", "")
	}
	account, fullPath, path, bypassName := util.ParseFullPath(fullPath, c.Request.Host)
	c.Set("account", account)
	c.Set("path", path)
	c.Set("full_path", fullPath)
	c.Set("bypass_name", bypassName)
	PwdCheck(c, fullPath)
	SortCheck(c)
	ThemeCheck(c)
	LayoutCheck(c)
	if len(module.GloablConfig.Accounts) == 0 {
		c.Redirect(http.StatusFound, module.GloablConfig.PathPrefix+module.GloablConfig.AdminPath)
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
	pwds, filePath, ok := dao.GetPwdFromPath(fullPath)
	if ok {
		pwd := c.Query("pwd")
		if pwd != "" && util.In(pwd, pwds) {
			c.Set("pwd_err_msg", "")
			c.Set("has_pwd", false)
		} else {
			pwdCookie, err := c.Request.Cookie("file_pwd")
			if err != nil {
				//redirect input pwd
				c.Set("pwd_err_msg", "")
				c.Set("has_pwd", true)
			} else {
				result, msg := VerifyPwd(pwds, filePath, pwdCookie.Value)
				c.Set("pwd_err_msg", msg)
				c.Set("has_pwd", result)
			}
		}
		c.Set("pwd_path", filePath)
	} else {
		c.Set("pwd_err_msg", "")
		c.Set("has_pwd", false)
		c.Set("pwd_path", "")
	}
}

func VerifyPwd(pwds []string, path string, cookiepwd string) (bool, string) {
	inputPwd := GetPwdFromCookie(cookiepwd, path)
	if inputPwd == "" {
		return true, ""
	} else if !util.In(inputPwd, pwds) {
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
