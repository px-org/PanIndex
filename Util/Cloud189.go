package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/eddieivan01/nic"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	math_rand "math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var CLoud189Session nic.Session

//获取文件列表
func Cloud189GetFiles(rootId, fileId string) {
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	pageNum := 1
	for {
		url := fmt.Sprintf("https://cloud.pan.cn/v2/listFiles.action?fileId=%s&mediaType=&keyword=&inGroupSpace=false&orderBy=3&order=DESC&pageNum=%d&pageSize=100&noCache=%s", fileId, pageNum, random())
		resp, err := CLoud189Session.Get(url, nil)
		if err != nil {
			panic(err.Error())
		}
		byteFiles := []byte(resp.Text)
		totalCount := jsoniter.Get(byteFiles, "recordCount").ToInt()
		d := jsoniter.Get(byteFiles, "data")
		paths := jsoniter.Get(byteFiles, "path")
		ps := []entity.Paths{}
		err = jsoniter.Unmarshal([]byte(paths.ToString()), &ps)
		p := ""
		flag := false
		if err == nil {
			for _, item := range ps {
				if flag == true && item.FileId != rootId {
					if strings.HasSuffix(p, "/") != true {
						p += "/" + item.FileName
					} else {
						p += item.FileName
					}
				}
				if item.FileId == rootId {
					flag = true
				}
				if flag == true && item.FileId == rootId {
					p += "/"
				}
			}
		}
		if d != nil {
			m := []entity.FileNode{}
			err = jsoniter.Unmarshal([]byte(d.ToString()), &m)
			if err == nil {
				for _, item := range m {
					if p == "/" {
						item.Path = "/" + item.FileName
					} else {
						item.Path = p + "/" + item.FileName
					}
					item.ParentPath = p
					item.SizeFmt = FormatFileSize(item.FileSize)
					if item.IsFolder == true {
						Cloud189GetFiles(rootId, item.FileId)
					} else {
						//如果是文件，解析下载直链
						/*dRedirectRep, _ := CLoud189Session.Get("https://cloud.189.cn/downloadFile.action?fileStr="+item.FileIdDigest+"&downloadType=1", nic.H{
							AllowRedirect: false,
						})
						redirectUrl := dRedirectRep.Header.Get("Location")
						dRedirectRep, _ = CLoud189Session.Get(redirectUrl, nic.H{
							AllowRedirect: false,
						})
						item.DownloadUrl = dRedirectRep.Header.Get("Location")*/
					}
					item.Delete = 0
					if config.GloablConfig.HideFileId != "" {
						listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
						sort.Strings(listSTring)
						i := sort.SearchStrings(listSTring, item.FileId)
						if i < len(listSTring) && listSTring[i] == item.FileId {
							item.Hide = 1
						}
					}
					model.SqliteDb.Save(item)
				}
			}
		}
		if pageNum*100 < totalCount {
			pageNum++
		} else {
			break
		}
	}
}
func GetDownlaodUrl(fileIdDigest string) string {
	dRedirectRep, _ := CLoud189Session.Get("https://cloud.pan.cn/downloadFile.action?fileStr="+fileIdDigest+"&downloadType=1", nic.H{
		AllowRedirect: false,
	})
	redirectUrl := dRedirectRep.Header.Get("Location")
	dRedirectRep, _ = CLoud189Session.Get(redirectUrl, nic.H{
		AllowRedirect: false,
	})
	return dRedirectRep.Header.Get("Location")
}
func GetDownlaodMultiFiles(fileId string) string {
	dRedirectRep, _ := CLoud189Session.Get(fmt.Sprintf("https://cloud.pan.cn/downloadMultiFiles.action?fileIdS=%s&downloadType=1&recursive=1", fileId), nic.H{
		AllowRedirect: false,
	})
	redirectUrl := dRedirectRep.Header.Get("Location")
	return redirectUrl
}

//天翼云网盘登录
func Cloud189Login(user, password string) string {
	CLoud189Session = nic.Session{}
	url := "https://cloud.pan.cn/udb/udb_login.jsp?pageId=1&redirectURL=/main.action"
	res, _ := CLoud189Session.Get(url, nil)
	b := res.Text
	lt := regexp.MustCompile(`lt = "(.+?)"`).FindStringSubmatch(b)[1]
	captchaToken := regexp.MustCompile(`captchaToken' value='(.+?)'`).FindStringSubmatch(b)[1]
	returnUrl := regexp.MustCompile(`returnUrl = '(.+?)'`).FindStringSubmatch(b)[1]
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(b)[1]
	//reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(b)[1]
	jRsakey := regexp.MustCompile(`j_rsaKey" value="(\S+)"`).FindStringSubmatch(b)[1]
	vCodeRS := ""
	if config.GloablConfig.Damagou.Username != "" {
		vCodeID := regexp.MustCompile(`picCaptcha\.do\?token\=([A-Za-z0-9\&\=]+)`).FindStringSubmatch(b)[1]
		vCodeRS = GetValidateCode(vCodeID)
		log.Warningln("[登录接口]得到验证码：" + vCodeRS)
	}
	userRsa := RsaEncode([]byte(user), jRsakey)
	passwordRsa := RsaEncode([]byte(password), jRsakey)
	url = "https://open.e.pan.cn/api/logbox/oauth2/loginSubmit.do"
	loginResp, _ := CLoud189Session.Post(url, nic.H{
		Data: nic.KV{
			"appKey":       "cloud",
			"accountType":  "01",
			"userName":     "{RSA}" + userRsa,
			"password":     "{RSA}" + passwordRsa,
			"validateCode": vCodeRS,
			"captchaToken": captchaToken,
			"returnUrl":    returnUrl,
			"mailSuffix":   "@pan.cn",
			"paramId":      paramId,
			"clientType":   "10010",
			"dynamicCheck": "FALSE",
			"cb_SaveName":  "1",
			"isOauth2":     "false",
		},
		Headers: nic.KV{
			"lt":         lt,
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/76.0",
			"Referer":    "https://open.e.pan.cn/",
		},
	})
	restCode := jsoniter.Get([]byte(loginResp.Text), "result").ToInt()
	//0登录成功，-2，需要获取验证码，-5 app info获取失败
	if restCode == 0 {
		toUrl := jsoniter.Get([]byte(loginResp.Text), "toUrl").ToString()
		res, _ := CLoud189Session.Get(toUrl, nil)
		return res.Cookies()[0].Value
	}
	errorReason := jsoniter.Get([]byte(loginResp.Text), "msg").ToString()
	if errorReason == "" {
		switch restCode {
		case -2:
			errorReason = "需要验证码"
		case -5:
			errorReason = "App Info 获取失败"
		default:
			errorReason = "未知错误"
		}
	}
	log.Warningln("[登录接口]登录失败，错误代码：" + strconv.Itoa(restCode) + " (" + errorReason + ")")
	return ""
}

//分享链接跳转下载
func Cloud189shareToDown(url, passCode, fileId, subFileId string) string {
	subIndex := strings.LastIndex(url, "/") + 1
	shortCode := url[subIndex:]
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if fileId != "" && subFileId != "" {
		if passCode == "" {
			passCode = "undefined"
		}
		floderFileDownUrlRep, _ := CLoud189Session.Get(fmt.Sprintf("https://cloud.pan.cn/v2/getFileDownloadUrl.action?"+
			"shortCode=%s&fileId=%s&accessCode=%s&subFileId=%s", shortCode, fileId, passCode, subFileId), nil)
		longDownloadUrl := GetBetweenStr(floderFileDownUrlRep.Text, "\"", "\"")
		longDownloadUrl = "http:" + strings.ReplaceAll(longDownloadUrl, "\\/", "/")
		floderFileDownUrlRep, _ = CLoud189Session.Get(longDownloadUrl, nic.H{
			AllowRedirect: false,
		})
		redirectUrl := floderFileDownUrlRep.Header.Get("Location")
		floderFileDownUrlRep, _ = CLoud189Session.Get(redirectUrl, nic.H{
			AllowRedirect: false,
		})
		return floderFileDownUrlRep.Header.Get("Location")
	}
	resp, err := CLoud189Session.Get(url, nil)
	if err != nil {
		panic(err.Error())
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err == nil {
		shareId, exists := doc.Find(".shareId").Attr("value")
		if !exists || shareId == "" {
			//文件夹
			verifyCode := GetBetweenStr(resp.Text, "_verifyCode = '", "'")
			url := fmt.Sprintf("https://cloud.pan.cn/v2/listShareDirByShareIdAndFileId.action?"+
				"shortCode=%s&accessCode=%s&verifyCode=%s&"+
				"orderBy=1&order=ASC&pageNum=1&pageSize=60",
				shortCode, passCode, verifyCode)
			resp, _ = CLoud189Session.Get(url, nil)
			return resp.Text
		} else {
			fileId = GetBetweenStr(resp.Text, "window.fileId = \"", "\"")
			if fileId == "" {
				//需要访问码，需要将访问码加入到cookie中再次请求获取fileId
				resp, _ = CLoud189Session.Get(url, nic.H{
					Cookies: nic.KV{
						"shareId_" + shareId: passCode,
					}})
				fileId = GetBetweenStr(resp.Text, "window.fileId = \"", "\"")
			}
			dRedirectRep, _ := CLoud189Session.Get(fmt.Sprintf("https://cloud.pan.cn/v2/getFileDownloadUrl.action?"+
				"shortCode=%s&fileId=%s", shortCode, fileId), nil)
			longDownloadUrl := GetBetweenStr(dRedirectRep.Text, "\"", "\"")
			longDownloadUrl = "http:" + strings.ReplaceAll(longDownloadUrl, "\\/", "/")
			dRedirectRep, _ = CLoud189Session.Get(longDownloadUrl, nic.H{
				AllowRedirect: false,
			})
			redirectUrl := dRedirectRep.Header.Get("Location")
			dRedirectRep, _ = CLoud189Session.Get(redirectUrl, nic.H{
				AllowRedirect: false,
			})
			return dRedirectRep.Header.Get("Location")
		}
	}
	return "https://cloud.pan.cn/"
}

// 加密
func RsaEncode(origData []byte, j_rsakey string) string {
	publicKey := []byte("-----BEGIN PUBLIC KEY-----\n" + j_rsakey + "\n-----END PUBLIC KEY-----")
	block, _ := pem.Decode(publicKey)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	pub := pubInterface.(*rsa.PublicKey)
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		log.Errorf("err: %s", err.Error())
	}
	return b64tohex(base64.StdEncoding.EncodeToString(b))
}

// 打码狗平台登录
func LoginDamagou() string {
	url := "http://www.damagou.top/apiv1/login.html?username=" + config.GloablConfig.Damagou.Username + "&password=" + config.GloablConfig.Damagou.Password
	res, _ := CLoud189Session.Get(url, nil)
	rsText := regexp.MustCompile(`([A-Za-z0-9]+)`).FindStringSubmatch(res.Text)[1]
	return rsText
}

// 调用打码狗获取验证码结果
func GetValidateCode(params string) string {
	timeStamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	url := "https://open.e.pan.cn/api/logbox/oauth2/picCaptcha.do?token=" + params + timeStamp
	log.Warningln("[登录接口]正在尝试获取验证码")
	res, err := CLoud189Session.Get(url, nic.H{
		Headers: nic.KV{
			"User-Agent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/76.0",
			"Referer":        "https://open.e.pan.cn/",
			"Sec-Fetch-Dest": "image",
			"Sec-Fetch-Mode": "no-cors",
			"Sec-Fetch-Site": "same-origin",
		},
	})
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	if err != nil {
		panic(err.Error())
	} else {
		f, err := os.OpenFile("validateCode.png", os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		f.Write(res.Bytes)
		damagouKey := LoginDamagou()
		base64Str := base64.StdEncoding.EncodeToString(res.Bytes)
		base64Str = "data:image/png;base64," + base64Str
		url := "http://www.damagou.top/apiv1/recognize.html"
		vres, _ := CLoud189Session.Post(url, nic.H{
			Data: nic.KV{
				"userkey": damagouKey,
				"image":   base64Str,
			},
		})
		return vres.Text
	}
	return ""
}

var b64map = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var BI_RM = "0123456789abcdefghijklmnopqrstuvwxyz"

func int2char(a int) string {
	return strings.Split(BI_RM, "")[a]
}

func b64tohex(a string) string {
	d := ""
	e := 0
	c := 0
	for i := 0; i < len(a); i++ {
		m := strings.Split(a, "")[i]
		if m != "=" {
			v := strings.Index(b64map, m)
			if 0 == e {
				e = 1
				d += int2char(v >> 2)
				c = 3 & v
			} else if 1 == e {
				e = 2
				d += int2char(c<<2 | v>>4)
				c = 15 & v
			} else if 2 == e {
				e = 3
				d += int2char(c)
				d += int2char(v >> 2)
				c = 3 & v
			} else {
				e = 0
				d += int2char(c<<2 | v>>4)
				d += int2char(15 & v)
			}
		}
	}
	if e == 1 {
		d += int2char(c << 2)
	}
	return d
}

//获取随机数
func random() string {
	return fmt.Sprintf("0.%17v", math_rand.New(math_rand.NewSource(time.Now().UnixNano())).Int63n(100000000000000000))
}

func FormatFileSize(fileSize int64) (size string) {
	if fileSize == 0 {
		return "-"
	} else if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2f B", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2f KB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f MB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f GB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f TB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2f EB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		return ""
	}
	n = n + len(start)
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		return ""
	}
	str = string([]byte(str)[:m])
	return str
}
