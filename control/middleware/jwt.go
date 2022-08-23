package middleware

import (
	"errors"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/libsgh/PanIndex/module"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	LoginTimeOut = 24 * 365
	identityKey  = "id"
)

func JWTMiddlewar() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "PanIndex Zone",
		Key:         []byte("PanIndex"),
		Timeout:     (time.Duration(LoginTimeOut)) * time.Hour,
		MaxRefresh:  (time.Duration(LoginTimeOut)) * time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				UserName: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals Login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			password := loginVals.Password
			user := loginVals.User
			if user == module.GloablConfig.AdminUser &&
				password == module.GloablConfig.AdminPassword {
				return &User{
					UserName: module.GloablConfig.AdminUser,
				}, nil
			}

			return nil, errors.New("密码错误！请重试")
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*User); ok && v.UserName == module.GloablConfig.AdminUser {
				return true
			}
			return false
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			//c.Redirect(http.StatusFound, module.GloablConfig.AdminPath+"/common")
			referer := c.Request.Header.Get("Referer")
			u, _ := url.Parse(referer)
			if strings.HasPrefix(u.Path, module.GloablConfig.AdminPath) {
				c.Redirect(http.StatusFound, module.GloablConfig.AdminPath+"/common")
			} else {
				c.Redirect(http.StatusFound, c.Request.Header.Get("Referer"))
			}
		},
		LogoutResponse: func(c *gin.Context, code int) {
			ThemeCheck(c)
			theme := c.GetString("theme")
			c.HTML(http.StatusOK, "templates/pan/admin/login.html", gin.H{
				"error":        true,
				"msg":          "退出成功",
				"redirect_url": "login",
				"config":       module.GloablConfig,
				"theme":        theme,
			})
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			path := c.Request.RequestURI
			if strings.HasPrefix(path, "/api") {
				//api return json
				c.JSON(code, gin.H{
					"status": code,
					"msg":    message,
				})
			} else {
				//return to login
				data := gin.H{
					"error":        true,
					"msg":          message,
					"redirect_url": "login",
					"config":       module.GloablConfig,
					"theme":        module.GloablConfig.Theme,
				}
				if message == "cookie token is empty" {
					data["error"] = false
					data["msg"] = ""
				}
				c.HTML(http.StatusOK, "templates/pan/admin/login.html", data)
			}
		},
		SendCookie:     true,
		SecureCookie:   false,   //non HTTPS dev environments
		CookieHTTPOnly: true,    // JS can't modify
		CookieName:     "token", // default jwt
		TokenLookup:    "header: Authorization, cookie: token",
		CookieSameSite: http.SameSiteDefaultMode, //SameSiteDefaultMode, SameSiteLaxMode, SameSiteStrictMode, SameSiteNoneMode
		TokenHeadName:  "Bearer",
		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})
	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	return authMiddleware, err
}

type Login struct {
	User     string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
type User struct {
	UserName string
}
