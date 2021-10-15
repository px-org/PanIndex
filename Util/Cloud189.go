package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/nic"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	math_rand "math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//var CLoud189Session nic.Session
var CLoud189Sessions = map[string]entity.Cloud189{}

//获取文件列表2.0
func Cloud189GetFiles(accountId, rootId, fileId, p string, hide, hasPwd int, syncChild bool) {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	pageNum := 1
	for {
		url := fmt.Sprintf("https://cloud.189.cn/api/open/file/listFiles.action?noCache=%s&pageSize=100&pageNum=%d&mediaType=0&folderId=%s&iconOption=5&orderBy=lastOpTime&descending=true", random(), pageNum, fileId)
		resp, err := CLoud189Session.Get(url, nic.H{
			Headers: nic.KV{
				"accept": "application/json;charset=UTF-8",
			},
		})
		if err != nil {
			panic(err.Error())
		}
		byteFiles := []byte(resp.Text)
		totalCount := jsoniter.Get(byteFiles, "fileListAO").Get("count").ToInt()
		if totalCount == 0 {
			break
		}
		d := jsoniter.Get(byteFiles, "fileListAO").Get("folderList")
		var folderList []Folder
		json.Unmarshal([]byte(d.ToString()), &folderList)
		//同步文件夹
		for _, item := range folderList {
			fn := entity.FileNode{}
			fn.FileId = fmt.Sprintf("%d", item.Id)
			fn.AccountId = accountId
			fn.FileName = item.Name
			fn.CreateTime = item.CreateDate
			fn.LastOpTime = item.LastOpTime
			fn.FileType = ""
			fn.IsFolder = true
			fn.FileSize = 0
			fn.SizeFmt = "-"
			fn.MediaType = 0
			fn.DownloadUrl = ""
			fn.ParentId = fmt.Sprintf("%d", item.ParentId)
			fn.ParentPath = p
			fn.Hide = 0
			fn.HasPwd = 0
			if hide == 1 {
				fn.Hide = hide
			} else {
				if config.GloablConfig.HideFileId != "" {
					listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && listSTring[i] == fn.FileId {
						fn.Hide = 1
					}
				}
			}
			if hasPwd == 1 {
				fn.HasPwd = hasPwd
			} else {
				if config.GloablConfig.PwdDirId != "" {
					listSTring := strings.Split(config.GloablConfig.PwdDirId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && strings.Split(listSTring[i], ":")[0] == fn.FileId {
						fn.HasPwd = 1
					}
				}
			}
			if p == "/" {
				fn.Path = p + fn.FileName
			} else {
				fn.Path = p + "/" + fn.FileName
			}
			if fn.IsFolder == true {
				//同步子目录&&子目录不为空
				if syncChild && item.FileCount > 0 {
					Cloud189GetFiles(accountId, rootId, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
				}
			}
			fn.Delete = 1
			fn.Id = uuid.NewV4().String()
			fn.CacheTime = time.Now().UnixNano()
			model.SqliteDb.Create(fn)
			/**/
		}
		//同步文件
		d = jsoniter.Get(byteFiles, "fileListAO").Get("fileList")
		var fileList []File
		json.Unmarshal([]byte(d.ToString()), &fileList)
		//同步文件夹
		for _, item := range fileList {
			fn := entity.FileNode{}
			fn.AccountId = accountId
			fn.FileId = fmt.Sprintf("%d", item.Id)
			fn.FileName = item.Name
			fn.FileIdDigest = ""
			fn.CreateTime = item.CreateDate
			fn.LastOpTime = item.LastOpTime
			fn.FileSize = item.Size
			fn.SizeFmt = FormatFileSize(fn.FileSize)
			fn.Id = uuid.NewV4().String()
			fn.CacheTime = time.Now().UnixNano()
			fn.IsFolder = false
			fn.IsStarred = false
			fn.MediaType = item.MediaType
			fn.FileType = GetFileType(fn.FileName)
			fn.ParentId = fileId
			fn.ParentPath = p
			if p == "/" {
				fn.Path = p + fn.FileName
			} else {
				fn.Path = p + "/" + fn.FileName
			}
			fn.Hide = 0
			fn.HasPwd = 0
			if hide == 1 {
				fn.Hide = hide
			} else {
				if config.GloablConfig.HideFileId != "" {
					listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && listSTring[i] == fn.FileId {
						fn.Hide = 1
					}
				}
			}
			if hasPwd == 1 {
				fn.HasPwd = hasPwd
			} else {
				if config.GloablConfig.PwdDirId != "" {
					listSTring := strings.Split(config.GloablConfig.PwdDirId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fn.FileId)
					if i < len(listSTring) && strings.Split(listSTring[i], ":")[0] == fn.FileId {
						fn.HasPwd = 1
					}
				}
			}
			fn.Delete = 1
			fn.Id = uuid.NewV4().String()
			fn.CacheTime = time.Now().UnixNano()
			model.SqliteDb.Create(fn)
		}
		pageNum++
	}
}

type Folder struct {
	CreateDate   string `json:"createDate"`
	FileCata     int    `json:"fileCata"`
	FileCount    int    `json:"fileCount"`
	FileListSize int    `json:"fileListSize"`
	Id           int64  `json:"id"`
	LastOpTime   string `json:"lastOpTime"`
	Name         string `json:"name"`
	ParentId     int64  `json:"parentId"`
	Rev          string `json:"rev"`
	StarLabel    int    `json:"starLabel"`
}
type File struct {
	CreateDate string `json:"createDate"`
	FileCata   int    `json:"fileCata"`
	Id         int64  `json:"id"`
	LastOpTime string `json:"lastOpTime"`
	Md5        string `json:"md5"`
	MediaType  int    `json:"mediaType"`
	Name       string `json:"name"`
	Rev        string `json:"rev"`
	StarLabel  int    `json:"starLabel"`
	Size       int64  `json:"size"`
}

func GetDownlaodUrlNew(accountId, fileId string) string {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	defer CLoud189Session.Client.CloseIdleConnections()
	dRedirectRep, err := CLoud189Session.Get(fmt.Sprintf("https://cloud.189.cn/api/open/file/getFileDownloadUrl.action?noCache=%s&fileId=%s", random(), fileId), nic.H{
		Headers: nic.KV{
			"accept": "application/json;charset=UTF-8	",
		},
	})
	if dRedirectRep != nil {
		defer dRedirectRep.Body.Close()
	}
	if err != nil {
		log.Error(err)
		return ""
	}
	resCode := jsoniter.Get(dRedirectRep.Bytes, "res_code").ToInt()
	if resCode == 0 {
		fileDownloadUrl := jsoniter.Get(dRedirectRep.Bytes, "fileDownloadUrl").ToString()
		dRedirectRep, err = CLoud189Session.Get(fileDownloadUrl, nic.H{
			AllowRedirect:     false,
			Timeout:           20,
			DisableKeepAlives: true,
		})
		if dRedirectRep != nil {
			defer dRedirectRep.Body.Close()
		}
		if err != nil {
			log.Error(err)
			return ""
		}
		return dRedirectRep.Header.Get("location")
	} else {
		return ""
	}
}
func GetDownlaodMultiFiles(accountId, fileId string) string {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	dRedirectRep, _ := CLoud189Session.Get(fmt.Sprintf("https://cloud.189.cn/downloadMultiFiles.action?fileIdS=%s&downloadType=1&recursive=1", fileId), nic.H{
		AllowRedirect: false,
	})
	redirectUrl := dRedirectRep.Header.Get("Location")
	return redirectUrl
}

//天翼云网盘登录
func Cloud189Login(accountId, user, password string) string {
	CLoud189Session := nic.Session{}
	url := "https://cloud.189.cn/api/portal/loginUrl.action?redirectURL=https%3A%2F%2Fcloud.189.cn%2Fmain.action"
	res, err := CLoud189Session.Get(url, nil)
	if err != nil {
		log.Errorln(err)
		return "5"
	}
	log.Debugf("登录页面接口：%s", res.Status)
	b := res.Text
	lt := ""
	ltText := regexp.MustCompile(`lt = "(.+?)"`)
	ltTextArr := ltText.FindStringSubmatch(b)
	if len(ltTextArr) > 0 {
		lt = ltTextArr[1]
	} else {
		return ""
	}
	captchaToken := regexp.MustCompile(`captchaToken' value='(.+?)'`).FindStringSubmatch(b)[1]
	returnUrl := regexp.MustCompile(`returnUrl = '(.+?)'`).FindStringSubmatch(b)[1]
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(b)[1]
	//reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(b)[1]
	jRsakey := regexp.MustCompile(`j_rsaKey" value="(\S+)"`).FindStringSubmatch(b)[1]
	vCodeID := regexp.MustCompile(`picCaptcha\.do\?token\=([A-Za-z0-9\&\=]+)`).FindStringSubmatch(b)[1]
	vCodeRS := ""
	if vCodeID != "" {
		//vCodeRS = GetValidateCode(accountId, vCodeID)
		//log.Warningln("[登录接口]得到验证码：" + vCodeRS)
		//log.Warningln("[登录接口]需要输入验证码")
		//return "4"
	}
	userRsa := RsaEncode([]byte(user), jRsakey)
	passwordRsa := RsaEncode([]byte(password), jRsakey)
	url = "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
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
			"Referer":    "https://open.e.189.cn/",
		},
	})
	restCode := jsoniter.Get([]byte(loginResp.Text), "result").ToInt()
	//0登录成功，-2，需要获取验证码，-5 app info获取失败
	if restCode == 0 {
		toUrl := jsoniter.Get([]byte(loginResp.Text), "toUrl").ToString()
		res, err := CLoud189Session.Get(toUrl, nic.H{
			AllowRedirect: false,
		})
		if err != nil {
			log.Warningln(err.Error())
			return "4"
		}
		sessionKey := GetSessionKey(CLoud189Session)
		CLoud189Sessions[accountId] = entity.Cloud189{CLoud189Session, sessionKey}
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
	return "4"
}

func GetSessionKey(session nic.Session) string {
	resp, error := session.Get("https://cloud.189.cn/v2/getUserBriefInfo.action?noCache="+random(), nil)
	if error != nil {
		return ""
	}
	sessionKey := jsoniter.Get(resp.Bytes, "sessionKey").ToString()
	return sessionKey
}

func GetRsaKey(accountId string) (string, string) {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	resp, error := CLoud189Session.Get("https://cloud.189.cn/api/security/generateRsaKey.action?noCache="+random(), nic.H{
		Headers: nic.KV{
			"accept": "application/json;charset=UTF-8",
		},
	})
	if error != nil {
		return "", ""
	}
	pubKey := jsoniter.Get(resp.Bytes, "pubKey").ToString()
	pkId := jsoniter.Get(resp.Bytes, "pkId").ToString()
	return pubKey, pkId
}

func Cloud189IsLogin(accountId string) bool {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	if _, ok := CLoud189Sessions[accountId]; ok {
		resp, err := CLoud189Session.Get("https://cloud.189.cn/v2/getLoginedInfos.action?showPC=true", nic.H{
			Timeout:           20,
			DisableKeepAlives: true,
		})
		if resp != nil {
			defer resp.Body.Close()
		}
		if err == nil && resp != nil && resp.Text != "" && jsoniter.Valid(resp.Bytes) && jsoniter.Get(resp.Bytes, "errorMsg").ToString() == "" {
			return true
		} else {
			if jsoniter.Get(resp.Bytes, "errorCode").ToString() == "InvalidSessionKey" {
				return false
			} else {
				return true
			}
		}
	}
	return false
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
func LoginDamagou(accountId string) string {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	url := "http://www.damagou.top/apiv1/login-bak.html?username=" + config.GloablConfig.Damagou.Username + "&password=" + config.GloablConfig.Damagou.Password
	res, _ := CLoud189Session.Get(url, nil)
	rsText := regexp.MustCompile(`([A-Za-z0-9]+)`).FindStringSubmatch(res.Text)[1]
	return rsText
}

// 调用打码狗获取验证码结果
func GetValidateCode(accountId, params string) string {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	timeStamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	url := "https://open.e.189.cn/api/logbox/oauth2/picCaptcha.do?token=" + params + timeStamp
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
		damagouKey := LoginDamagou(accountId)
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

func Cloud189UploadFilesNew(accountId, parentId string, files []*multipart.FileHeader) bool {
	CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	sessionKey := CLoud189Sessions[accountId].SessionKey
	UPLOAD_PART_SIZE := 10 * 1024 * 1024
	date := time.Now().Unix()
	rid := "8aab8fc8-99ae-458e-89dd-85e0d7b7d4f6"
	pk := strings.ReplaceAll(rid, "-", "")
	pubKey, pId := GetRsaKey(accountId)
	fmt.Println(qs(nic.KV{
		"parentFolderId": parentId,
		"fileName":       files[0].Filename,
		"fileSize":       files[0].Size,
		"sliceSize":      UPLOAD_PART_SIZE,
		"lazyCheck":      1,
	}))
	re, _ := nic.Post("https://www.devglan.com/online-tools/aes-encryption", nic.H{
		Data: nic.KV{
			"file": "undefined",
			"data": nic.KV{
				"textToEncrypt": qs(nic.KV{
					"parentFolderId": parentId,
					"fileName":       files[0].Filename,
					"fileSize":       files[0].Size,
					"sliceSize":      UPLOAD_PART_SIZE,
					"lazyCheck":      1,
				}),
				"secretKey":  pk[0:16],
				"mode":       "ECB",
				"keySize":    "128",
				"dataFormat": "Hex",
			},
		},
	})
	params := jsoniter.Get(re.Bytes, "output").ToString()
	signature := hmacSha1(fmt.Sprintf("SessionKey=%s&Operate=GET&RequestURI=%s&Date=%s&params=%s", sessionKey, "/person/initMultiUpload", date, params), pk)
	encryptiontext := RsaEncode([]byte(pk), pubKey)
	headers := nic.KV{
		"encryptiontext": encryptiontext,
		"pkid":           pId,
		"signature":      signature,
		"sessionkey":     sessionKey,
		"x-request-id":   rid,
		"x-request-date": date,
		"origin":         "https://cloud.189.cn",
		"referer":        "https://cloud.189.cn/",
	}
	response, _ := CLoud189Session.Get("https://upload.cloud.189.cn"+"/person/initMultiUpload?params="+params, nic.H{
		Headers: headers,
	})
	fmt.Println(response.Text)
	return true
}
func hmacSha1(data string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
func Cloud189UploadFiles(accountId, parentId string, files []*multipart.FileHeader) bool {
	//CLoud189Session := CLoud189Sessions[accountId].Cloud189Session
	//response, _ := CLoud189Session.Get("https://cloud.189.cn/main.action#home", nil)
	//sessionKey := GetCurBetweenStr(response.Text, "window.edrive.sessionKey = '", "';")
	sessionKey := CLoud189Sessions[accountId].SessionKey
	log.Debug(sessionKey)
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
		fileContent, _ := file.Open()
		byteContent, _ := ioutil.ReadAll(fileContent)
		reader := bytes.NewReader(byteContent)
		b := &bytes.Buffer{}
		writer := multipart.NewWriter(b)
		writer.WriteField("parentId", parentId)
		writer.WriteField("sessionKey", sessionKey)
		writer.WriteField("opertype", "1")
		writer.WriteField("fname", file.Filename)
		part, _ := writer.CreateFormFile("Filedata", file.Filename)
		io.Copy(part, reader)
		writer.Close()
		r, _ := http.NewRequest("POST", "https://hb02.upload.cloud.189.cn/v1/DCIWebUploadAction", b)
		r.Header.Add("Content-Type", writer.FormDataContentType())
		res, _ := http.DefaultClient.Do(r)
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		log.Debugf("上传接口返回：%s", string(body))
		log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
	}
	return true
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
		n = 0
	} else {
		n = n + len(start)
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}

func qs(params nic.KV) string {
	var dataParams string
	//ksort
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("key:", k, "Value:", Strval(params[k]))
		dataParams = dataParams + k + "=" + Strval(params[k]) + "&"
	}
	ff := dataParams[0 : len(dataParams)-1]
	return ff
}
