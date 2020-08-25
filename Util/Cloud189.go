package Util

import (
	"PanIndex/entity"
	"PanIndex/model"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/eddieivan01/nic"
	jsoniter "github.com/json-iterator/go"
	"log"
	math_rand "math/rand"
	"regexp"
	"strings"
	"time"
)

var CLoud189Session = nic.Session{}

//获取文件列表
func Cloud189GetFiles(rootId, fileId string) {
	pageNum := 1
	for {
		url := fmt.Sprintf("https://cloud.189.cn/v2/listFiles.action?fileId=%s&mediaType=&keyword=&inGroupSpace=false&orderBy=3&order=DESC&pageNum=%d&pageSize=100&noCache=%s", fileId, pageNum, random())
		resp, err := CLoud189Session.Get(url, nil)
		if err != nil {
			log.Fatal(err.Error())
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
	dRedirectRep, _ := CLoud189Session.Get("https://cloud.189.cn/downloadFile.action?fileStr="+fileIdDigest+"&downloadType=1", nic.H{
		AllowRedirect: false,
	})
	redirectUrl := dRedirectRep.Header.Get("Location")
	dRedirectRep, _ = CLoud189Session.Get(redirectUrl, nic.H{
		AllowRedirect: false,
	})
	return dRedirectRep.Header.Get("Location")
}
func GetDownlaodMultiFiles(fileId string) string {
	dRedirectRep, _ := CLoud189Session.Get(fmt.Sprintf("https://cloud.189.cn/downloadMultiFiles.action?fileIdS=%s&downloadType=1&recursive=1", fileId), nic.H{
		AllowRedirect: false,
	})
	redirectUrl := dRedirectRep.Header.Get("Location")
	return redirectUrl
}

//天翼云网盘登录
func Cloud189Login(user, password string) string {
	url := "https://cloud.189.cn/udb/udb_login.jsp?pageId=1&redirectURL=/main.action"
	res, _ := CLoud189Session.Get(url, nil)
	b := res.Text
	lt := regexp.MustCompile(`lt = "(.+?)"`).FindStringSubmatch(b)[1]
	captchaToken := regexp.MustCompile(`captchaToken' value='(.+?)'`).FindStringSubmatch(b)[1]
	returnUrl := regexp.MustCompile(`returnUrl = '(.+?)'`).FindStringSubmatch(b)[1]
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(b)[1]
	//reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(b)[1]
	jRsakey := regexp.MustCompile(`j_rsaKey" value="(\S+)"`).FindStringSubmatch(b)[1]
	userRsa := RsaEncode([]byte(user), jRsakey)
	passwordRsa := RsaEncode([]byte(password), jRsakey)
	url = "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
	loginResp, _ := CLoud189Session.Post(url, nic.H{
		Data: nic.KV{
			"appKey":       "cloud",
			"accountType":  "01",
			"userName":     "{RSA}" + userRsa,
			"password":     "{RSA}" + passwordRsa,
			"validateCode": "",
			"captchaToken": captchaToken,
			"returnUrl":    returnUrl,
			"mailSuffix":   "@189.cn",
			"paramId":      paramId,
			"clientType":   "10010",
			"dynamicCheck": "FALSE",
			"cb_SaveName":  "1",
			"isOauth2":     "false",
		},
		Headers: nic.KV{
			"lt":         lt,
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/76.0",
			"Referer":    "https://open.e.189.cn/",
		},
	})
	restCode := jsoniter.Get([]byte(loginResp.Text), "result").ToInt()
	//0登录成功，-2，需要获取验证码，-5 app info获取失败
	if restCode == 0 {
		toUrl := jsoniter.Get([]byte(loginResp.Text), "toUrl").ToString()
		res, _ := CLoud189Session.Get(toUrl, nil)
		return res.Cookies()[0].Value
	}
	return ""
}

// 加密
func RsaEncode(origData []byte, j_rsakey string) string {
	publicKey := []byte("-----BEGIN PUBLIC KEY-----\n" + j_rsakey + "\n-----END PUBLIC KEY-----")
	block, _ := pem.Decode(publicKey)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	pub := pubInterface.(*rsa.PublicKey)
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		fmt.Println("err: " + err.Error())
	}
	return b64tohex(base64.StdEncoding.EncodeToString(b))
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
