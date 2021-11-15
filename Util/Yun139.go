package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/model"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/nic"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var Yun139Credentials = map[string]entity.Yun139{}

func Yun139Login(account entity.Account) string {
	if Yun139LoginCheck(account) {
		Yun139Credentials[account.Id] = entity.Yun139{
			account.Password,
			account.User,
		}
		return account.Password
	}
	return ""
}
func Yun139LoginCheck(account entity.Account) bool {
	Yun139Credentials[account.Id] = entity.Yun139{
		account.Password,
		account.User,
	}
	body := nic.KV{
		"qryUserExternInfoReq": nic.KV{
			"commonAccountInfo": nic.KV{
				"account":     account.User,
				"accountType": 1,
			},
		},
	}
	resp, _ := nic.Post("https://yun.139.com/orchestration/personalCloud/user/v1.0/qryUserExternInfo", nic.H{
		Headers: createHeaders(body, account.Password),
		JSON:    body,
	})
	if resp != nil && jsoniter.Get(resp.Bytes, "success").ToBool() {
		nicName := jsoniter.Get(resp.Bytes, "data").Get("qryUserExternInfoRsp").Get("nickName").ToString()
		if nicName != "" {
			return true
		}
	}
	return false
}

func Yun139GetFiles(accountId, fileId, p string, hide, hasPwd int, syncChild bool) {
	yun139Credential := Yun139Credentials[accountId]
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	offset := 0
	size := 200
	for {
		body := nic.KV{
			"catalogID":         fileId,
			"sortDirection":     1,
			"filterType":        0,
			"catalogSortType":   0,
			"contentSortType":   0,
			"startNumber":       offset + 1,
			"endNumber":         offset + size,
			"commonAccountInfo": nic.KV{"account": yun139Credential.Mobile, "accountType": 1},
		}
		resp, _ := nic.Post("https://yun.139.com/orchestration/personalCloud/catalog/v1.0/getDisk", nic.H{
			Headers: createHeaders(body, yun139Credential.Cookie),
			JSON:    body,
		})
		status := jsoniter.Get(resp.Bytes, "success").ToBool()
		isCompleted := jsoniter.Get(resp.Bytes, "data").Get("getDiskResult").Get("isCompleted").ToInt()
		fns := []entity.FileNode{}
		if status {
			diskResults := jsoniter.Get(resp.Bytes, "data").Get("getDiskResult")
			floderList := diskResults.Get("catalogList")
			if floderList != nil {
				var ls []map[string]interface{}
				json.Unmarshal([]byte(floderList.ToString()), &ls)
				for _, item := range ls {
					fn := entity.FileNode{
						Id:           uuid.NewV4().String(),
						AccountId:    accountId,
						Delete:       1,
						CacheTime:    time.Now().UnixNano(),
						FileIdDigest: "",
						IsFolder:     true,
						FileId:       item["catalogID"].(string),
						FileName:     item["catalogName"].(string),
						FileSize:     0,
						SizeFmt:      "-",
						FileType:     "",
						MediaType:    0,
						DownloadUrl:  "",
						CreateTime:   TimeFormat139(item["createTime"].(string)),
						LastOpTime:   TimeFormat139(item["updateTime"].(string)),
						ParentId:     fileId,
						ParentPath:   p,
					}
					if p == "/" {
						fn.Path = p + fn.FileName
					} else {
						fn.Path = p + "/" + fn.FileName
					}
					FileNodeAuth(&fn, hide, hasPwd)
					if fn.IsFolder == true {
						//同步子目录&&子目录不为空
						if syncChild {
							Yun139GetFiles(accountId, fn.FileId, fn.Path, fn.Hide, fn.HasPwd, syncChild)
						}
					}
					fns = append(fns, fn)
				}
			}
			fileList := diskResults.Get("contentList")
			if fileList != nil {
				var fs []map[string]interface{}
				json.Unmarshal([]byte(fileList.ToString()), &fs)
				for _, item := range fs {
					fn := entity.FileNode{
						Id:           uuid.NewV4().String(),
						AccountId:    accountId,
						Delete:       1,
						CacheTime:    time.Now().UnixNano(),
						FileIdDigest: "",
						IsFolder:     false,
						FileId:       item["contentID"].(string),
						FileName:     item["contentName"].(string),
						FileSize:     int64(item["contentSize"].(float64)),
						SizeFmt:      FormatFileSize(int64(item["contentSize"].(float64))),
						FileType:     item["contentSuffix"].(string),
						MediaType:    int(item["contentType"].(float64)),
						DownloadUrl:  "",
						CreateTime:   TimeFormat139(item["uploadTime"].(string)),
						LastOpTime:   TimeFormat139(item["updateTime"].(string)),
						ParentId:     fileId,
						ParentPath:   p,
					}
					if p == "/" {
						fn.Path = p + fn.FileName
					} else {
						fn.Path = p + "/" + fn.FileName
					}
					FileNodeAuth(&fn, hide, hasPwd)
					fns = append(fns, fn)
				}
			}
			model.SqliteDb.Create(&fns)
		}
		if isCompleted != 1 {
			//获取下一页
			offset = offset + size
		} else {
			break
		}
	}
}
func FileNodeAuth(fn *entity.FileNode, hide, hasPwd int) {
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
}
func GetYun139DownUrl(accountId, fileId string) string {
	yun139Credential := Yun139Credentials[accountId]
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	body := nic.KV{
		"appName":           "",
		"contentID":         fileId,
		"commonAccountInfo": nic.KV{"account": yun139Credential.Mobile, "accountType": 1},
	}
	resp, _ := nic.Post("https://yun.139.com/orchestration/personalCloud/uploadAndDownload/v1.0/downloadRequest", nic.H{
		Headers: createHeaders(body, yun139Credential.Cookie),
		JSON:    body,
	})
	status := jsoniter.Get(resp.Bytes, "success").ToBool()
	if status {
		return jsoniter.Get(resp.Bytes, "data").Get("downloadURL").ToString()
	}
	return ""
}
func Yun139Upload(accountId, parentId string, files []*multipart.FileHeader) bool {
	yun139Credential := Yun139Credentials[accountId]
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
		f, _ := file.Open()
		fileBytes, _ := ioutil.ReadAll(f)
		textQuoted := strconv.QuoteToASCII(file.Filename)
		textUnquoted := textQuoted[1 : len(textQuoted)-1]
		digest := fmt.Sprintf("%x", md5.Sum(fileBytes))
		body := nic.KV{
			"fileCount":         1,
			"parentCatalogID":   parentId,
			"manualRename":      2,
			"newCatalogName":    "",
			"operation":         0,
			"totalSize":         file.Size,
			"commonAccountInfo": nic.KV{"account": yun139Credential.Mobile, "accountType": 1},
			"uploadContentList": []nic.KV{nic.KV{
				"contentName": file.Filename,
				"contentSize": file.Size,
				"digest":      digest,
			}},
		}
		resp, _ := nic.Post("https://yun.139.com/orchestration/personalCloud/uploadAndDownload/v1.0/pcUploadFileRequest", nic.H{
			Headers: createHeaders(body, yun139Credential.Cookie),
			JSON:    body,
		})
		status := jsoniter.Get(resp.Bytes, "success").ToBool()
		if status {
			isNeedUpload := jsoniter.Get(resp.Bytes, "data").
				Get("uploadResult").
				Get("newContentIDList").
				Get(0).
				Get("isNeedUpload").ToInt()
			if isNeedUpload == 1 {
				//需要上传
				uploadUrl := jsoniter.Get(resp.Bytes, "data").
					Get("uploadResult").
					Get("redirectionUrl").ToString()
				uploadTaskID := jsoniter.Get(resp.Bytes, "data").
					Get("uploadResult").
					Get("uploadTaskID").ToString()
				bfs := ReadBlock(file)
				for _, bf := range bfs {
					r, _ := http.NewRequest("POST", uploadUrl, bytes.NewReader(bf.Content))
					r.Header.Add("uploadtaskID", uploadTaskID)
					r.Header.Add("rangeType", "0")
					r.Header.Add("Range", bf.Range2)
					r.Header.Add("Content-Type", "text/plain;name="+textUnquoted)
					r.Header.Add("contentSize", strconv.FormatInt(file.Size, 10))
					r.Header.Add("Content-Length", strconv.FormatInt(file.Size, 10))
					r.Header.Add("Referer", "https://yun.139.com/")
					r.Header.Add("x-SvcType", "1")
					r.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
					res, _ := http.DefaultClient.Do(r)
					defer res.Body.Close()
					b, _ := ioutil.ReadAll(res.Body)
					log.Debugf("上传接口返回：%s", b)
				}
			}
		}
		log.Debugf("上传请求接口返回：%s", resp.Text)
		log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, ShortDur(time.Now().Sub(t1)))
	}
	return true
}
func createHeaders(body nic.KV, cookie string) nic.KV {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	key := getRandomStr(16)
	json, _ := jsoniter.MarshalToString(body)
	sign := getSign(timestamp, key, json)
	headers := nic.KV{
		"x-huawei-channelSrc": "10000034",
		"x-inner-ntwk":        "2",
		"mcloud-channel":      "1000101",
		"mcloud-client":       "10701",
		"mcloud-sign":         fmt.Sprintf("%s,%s,%s", timestamp, key, sign),
		"content-type":        "application/json;charset=UTF-8",
		"caller":              "web",
		"CMS-DEVICE":          "default",
		"x-DeviceInfo":        "||9|6.5.2|chrome|95.0.4638.17|||linux unknow||zh-CN|||",
		"x-SvcType":           "1",
		"referer":             "https://yun.139.com/w/",
		"Cookie":              cookie,
	}
	return headers
}
func getSign(timestamp, key, json string) string {
	//去除多余空格
	json = strings.TrimSpace(json)
	json = encodeURIComponent(json)
	c := strings.Split(json, "")
	sort.Strings(c)
	json = strings.Join(c, "")
	s1 := Md5(base64.StdEncoding.EncodeToString([]byte(json)))
	s2 := Md5(timestamp + ":" + key)
	return strings.ToUpper(Md5(s1 + s2))
}
func getRandomStr(n int) string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func Md5(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}
func TimeFormat139(timeStr string) string {
	t, _ := time.ParseInLocation("20060102150405", timeStr, time.Local)
	timeFormat := time.Unix(t.Unix(), 0).Format("2006-01-02 15:04:05")
	return timeFormat
}
func encodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	r = strings.Replace(r, "%21", "!", -1)
	r = strings.Replace(r, "%27", "'", -1)
	r = strings.Replace(r, "%28", "(", -1)
	r = strings.Replace(r, "%29", ")", -1)
	r = strings.Replace(r, "%2A", "*", -1)
	return r
}
