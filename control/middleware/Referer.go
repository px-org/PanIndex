package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/libsgh/PanIndex/module"
	log "github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"strings"
)

//防盗链检测
func CheckReferer(c *gin.Context) bool {
	isForbidden := true
	if module.GloablConfig.EnableSafetyLink == "0" {
		//开启了防盗链
		isForbidden = false
	} else {
		host := c.Request.Host
		referer, err := url.Parse(c.Request.Referer())
		if err != nil {
			log.Warningln(err)
		}
		if referer != nil && referer.Host != "" {
			if referer.Host == host {
				//站内，自动通过
				isForbidden = false
			} else if referer.Host != host && len(module.GloablConfig.OnlyReferrer) > 0 {
				//外部引用，并且设置了防盗链，需要进行判断
				for _, rf := range strings.Split(module.GloablConfig.OnlyReferrer, ",") {
					match, _ := regexp.MatchString(rf, referer.Host)
					if rf == referer.Host || match {
						isForbidden = false
						break
					}
				}
			} else {
				isForbidden = false
			}
		} else {
			if module.GloablConfig.IsNullReferrer == "1" {
				isForbidden = false
			} else {
				isForbidden = true
			}
		}
	}
	return isForbidden
}

func RequestCancelRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Debug("client cancel the request...")
				log.Error(err)
				c.Request.Context().Done()
			}
		}()
		c.Next()
	}
}
