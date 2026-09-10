package cloud189

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const (
	cloud189AppId      = "9317140619"
	cloud189ClientType = "10020"
	cloud189ReturnURL  = "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html"
	cloud189API        = "https://api.cloud.189.cn"
	cloud189APIVersion = "7.1.8.0"
	cloud189ChannelId  = "web_cloud.189.cn"
)

var CLoud189s = map[string]module.Cloud189{}

func init() {
	base.RegisterPan("cloud189", &Cloud189{})
}

type Cloud189 struct{}

func (c Cloud189) IsLogin(account *module.Account) bool {
	cloud189, ok := CLoud189s[account.Id]
	if !ok || cloud189.Cloud189Session == nil || cloud189.SessionKey == "" {
		return false
	}
	resp, err := c.ReqlReq(*account).Get(cloud189API + "/getUserInfo.action")
	if err != nil || resp == nil || !jsoniter.Valid(resp.Body()) {
		return false
	}
	errCode := jsoniter.Get(resp.Body(), "errorCode").ToString()
	if errCode == "InvalidSessionKey" || errCode == "InvalidAccessToken" || errCode == "InvalidSignature" {
		return false
	}
	if jsoniter.Get(resp.Body(), "res_code").ToInt() != 0 && errCode != "" {
		return false
	}
	return true
}

func (c Cloud189) AuthLogin(account *module.Account) (string, error) {
	client := resty.New()
	referer := ""
	lt := ""
	reqId := ""
	appKey := ""
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		referer = req.URL.String()
		q := req.URL.Query()
		if v := q.Get("lt"); v != "" {
			lt = v
		}
		if v := q.Get("reqId"); v != "" {
			reqId = v
		}
		if v := q.Get("appId"); v != "" {
			appKey = v
		}
		return nil
	}))

	_, err := client.R().
		SetQueryParams(map[string]string{
			"appId":      cloud189AppId,
			"clientType": cloud189ClientType,
			"timeStamp":  strconv.FormatInt(time.Now().UnixMilli(), 10),
			"returnURL":  cloud189ReturnURL,
		}).
		Get("https://cloud.189.cn/unifyLoginForPC.action")
	if err != nil {
		log.Error(err)
		return "", err
	}
	if referer == "" || lt == "" || reqId == "" {
		return "", fmt.Errorf("cloud189: failed to resolve login params")
	}
	if appKey == "" {
		appKey = cloud189AppId
	}

	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))

	appConfResp, err := client.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Origin":       "https://open.e.189.cn",
			"Referer":      referer,
			"Reqid":        reqId,
			"lt":           lt,
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		}).
		SetFormData(map[string]string{
			"version": "2.0",
			"appKey":  appKey,
		}).
		Post("https://open.e.189.cn/api/logbox/oauth2/appConf.do")
	if err != nil {
		log.Error(err)
		return "", err
	}
	if jsoniter.Get(appConfResp.Body(), "result").ToString() != "0" {
		msg := jsoniter.Get(appConfResp.Body(), "msg").ToString()
		return "", fmt.Errorf("cloud189 appConf failed: %s", msg)
	}
	appConfAppKey := jsoniter.Get(appConfResp.Body(), "data", "appKey").ToString()
	accountType := jsoniter.Get(appConfResp.Body(), "data", "accountType").ToString()
	clientType := jsoniter.Get(appConfResp.Body(), "data", "clientType").ToString()
	paramId := jsoniter.Get(appConfResp.Body(), "data", "paramId").ToString()
	mailSuffix := jsoniter.Get(appConfResp.Body(), "data", "mailSuffix").ToString()
	isOauth2 := jsoniter.Get(appConfResp.Body(), "data", "isOauth2").ToBool()

	encryptConfResp, err := client.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Referer":      referer,
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		}).
		SetFormData(map[string]string{
			"appId": "cloud",
		}).
		Post("https://open.e.189.cn/api/logbox/config/encryptConf.do")
	if err != nil {
		log.Error(err)
		return "", err
	}
	if jsoniter.Get(encryptConfResp.Body(), "result").ToInt() != 0 {
		return "", fmt.Errorf("cloud189: failed to get encrypt config")
	}
	pubKey := jsoniter.Get(encryptConfResp.Body(), "data", "pubKey").ToString()
	pre := jsoniter.Get(encryptConfResp.Body(), "data", "pre").ToString()

	userRsa, err := util.RsaEncryptHex([]byte(account.User), pubKey)
	if err != nil {
		return "", err
	}
	passwordRsa, err := util.RsaEncryptHex([]byte(account.Password), pubKey)
	if err != nil {
		return "", err
	}

	loginResp, err := client.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Referer":      referer,
			"Reqid":        reqId,
			"lt":           lt,
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		}).
		SetFormData(map[string]string{
			"version":         "v2.0",
			"appKey":          appConfAppKey,
			"accountType":     accountType,
			"userName":        pre + userRsa,
			"epd":             pre + passwordRsa,
			"captchaType":     "",
			"validateCode":    "",
			"smsValidateCode": "",
			"captchaToken":    "",
			"returnUrl":       referer,
			"mailSuffix":      mailSuffix,
			"dynamicCheck":    "FALSE",
			"clientType":      clientType,
			"cb_SaveName":     "0",
			"isOauth2":        strconv.FormatBool(isOauth2),
			"state":           "",
			"paramId":         paramId,
		}).
		Post("https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do")
	if err != nil {
		log.Error(err)
		return "", err
	}
	restCode := jsoniter.Get(loginResp.Body(), "result").ToInt()
	if restCode == -2 {
		return "", base.LoginCaptcha
	}
	if restCode != 0 {
		msg := jsoniter.Get(loginResp.Body(), "msg").ToString()
		log.Errorf("cloud189 login failed: %s", msg)
		return "", fmt.Errorf("cloud189 login failed: %s", msg)
	}
	toUrl := jsoniter.Get(loginResp.Body(), "toUrl").ToString()
	if toUrl == "" {
		return "", fmt.Errorf("cloud189 login failed: empty toUrl")
	}

	sessionClient := resty.New()
	sessionClient.SetHeader("Accept", "application/json;charset=UTF-8")
	sessionClient.SetHeader("user-agent", "desktop")
	sessionResp, err := sessionClient.R().
		SetFormData(map[string]string{
			"redirectURL": toUrl,
		}).
		SetQueryParams(map[string]string{
			"rand":       strconv.FormatInt(time.Now().UnixMilli(), 10),
			"clientType": "TELEPC",
			"version":    cloud189APIVersion,
			"channelId":  cloud189ChannelId,
		}).
		Post(cloud189API + "/getSessionForPC.action")
	if err != nil {
		log.Error(err)
		return "", err
	}
	sessionKey := jsoniter.Get(sessionResp.Body(), "sessionKey").ToString()
	sessionSecret := jsoniter.Get(sessionResp.Body(), "sessionSecret").ToString()
	accessToken := jsoniter.Get(sessionResp.Body(), "accessToken").ToString()
	if sessionKey == "" || sessionSecret == "" {
		errCode := jsoniter.Get(sessionResp.Body(), "errorCode").ToString()
		errMsg := jsoniter.Get(sessionResp.Body(), "errorMsg").ToString()
		if errCode == "" {
			errCode = jsoniter.Get(sessionResp.Body(), "res_code").ToString()
			errMsg = jsoniter.Get(sessionResp.Body(), "res_message").ToString()
		}
		return "", fmt.Errorf("cloud189 getSessionForPC failed: %s %s", errCode, errMsg)
	}

	CLoud189s[account.Id] = module.Cloud189{
		Cloud189Session: sessionClient,
		SessionKey:      sessionKey,
		SessionSecret:   sessionSecret,
		AccessToken:     accessToken,
		RootId:          account.RootId,
		FamilyId:        "",
	}
	c.attachSigner(account.Id)
	return fmt.Sprintf("cloud189 login success [%s]", time.Now()), nil
}

func (c Cloud189) attachSigner(accountId string) {
	cloud189 := CLoud189s[accountId]
	if cloud189.Cloud189Session == nil {
		return
	}
	cloud189.Cloud189Session.OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
		s := CLoud189s[accountId]
		if s.SessionKey == "" || s.SessionSecret == "" {
			return fmt.Errorf("cloud189 session invalid, please re-login")
		}
		u, err := url.Parse(req.URL)
		if err != nil {
			return err
		}
		now := time.Now()
		req.QueryParam.Set("rand", strconv.FormatInt(now.UnixMilli(), 10))
		req.QueryParam.Set("clientType", "TELEPC")
		req.QueryParam.Set("version", cloud189APIVersion)
		req.QueryParam.Set("channelId", cloud189ChannelId)

		if strings.Contains(u.Path, "/getSessionForPC.action") {
			return nil
		}
		date := now.Format(time.RFC1123)
		signData := fmt.Sprintf("SessionKey=%s&Operate=%s&RequestURI=%s&Date=%s",
			s.SessionKey, req.Method, u.Path, date)
		req.SetHeaders(map[string]string{
			"Date":         date,
			"user-agent":   "desktop",
			"SessionKey":   s.SessionKey,
			"Signature":    util.HmacSha1(signData, s.SessionSecret),
			"X-Request-ID": uuid.NewV4().String(),
			"Accept":       "application/json;charset=UTF-8",
		})
		return nil
	})
}

func (c Cloud189) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	pageSize := 100
	pageNum := 1
	var err error
	var resp *resty.Response
	for {
		var filesResp Cloud189FilesResp
		resp, err = c.ReqlReq(account).
			//SetResult(&filesResp).
			SetQueryParams(map[string]string{
				"noCache":    util.Random(),
				"pageSize":   fmt.Sprintf("%d", pageSize),
				"pageNum":    fmt.Sprintf("%d", pageNum),
				"mediaType":  "0",
				"folderId":   fileId,
				"iconOption": "5",
				"orderBy":    "lastOpTime",
				"descending": "true",
			}).
			SetHeader("accept", "application/json;charset=UTF-8").
			Get(c.RealUrl(account, cloud189API+"/listFiles.action"))
		jsoniter.Unmarshal(resp.Body(), &filesResp)
		err := jsoniter.Unmarshal([]byte(resp.String()), &filesResp)
		if err != nil {
			break
		}
		if filesResp.ResCode == 0 {
			if len(filesResp.FileListAO.FolderList) == 0 && len(filesResp.FileListAO.FileList) == 0 {
				break
			}
			for _, folder := range filesResp.FileListAO.FolderList {
				fn := module.FileNode{}
				fn.Id = uuid.NewV4().String()
				fn.FileId = fmt.Sprintf("%d", folder.ID)
				fn.FileName = folder.Name
				fn.CreateTime = folder.CreateDate
				fn.LastOpTime = folder.LastOpTime
				fn.ParentId = fmt.Sprintf("%d", folder.ParentID)
				fn.IsDelete = 1
				fn.IsFolder = true
				fn.FileType = ""
				fn.FileSize = 0
				fn.SizeFmt = "-"
				if path == "/" {
					fn.Path = path + fn.FileName
				} else {
					fn.Path = path + "/" + fn.FileName
				}
				fn.AccountId = account.Id
				fn.ParentId = fileId
				fn.ParentPath = path
				fileNodes = append(fileNodes, fn)
			}
			for _, file := range filesResp.FileListAO.FileList {
				fn, _ := c.ToFileNode(file)
				if path == "/" {
					fn.Path = path + fn.FileName
				} else {
					fn.Path = path + "/" + fn.FileName
				}
				fn.AccountId = account.Id
				fn.ParentId = fileId
				fn.ParentPath = path
				fileNodes = append(fileNodes, fn)
			}
			pageNum++
		} else {
			break
		}
	}
	return fileNodes, err
}

func (c Cloud189) ToFileNode(item FileList) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = fmt.Sprintf("%d", item.ID)
	fn.FileName = item.Name
	fn.CreateTime = item.CreateDate
	fn.LastOpTime = item.LastOpTime
	fn.IsDelete = 1
	fn.IsFolder = false
	fn.FileType = util.GetExt(item.Name)
	fn.ViewType = util.GetViewType(fn.FileType)
	fn.FileSize = int64(item.Size)
	fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	fn.Thumbnail = item.Icon.SmallURL
	return fn, nil
}

func (c Cloud189) ToFileNode2(item Cloud189FileResp) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = item.FileID
	fn.ParentId = item.ParentID
	fn.FileName = item.FileName
	if item.CreateTime > 0 {
		fn.CreateTime = time.Unix(0, item.CreateTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
	} else {
		fn.CreateTime = item.CreateDate
	}
	if item.LastOpTime > 0 {
		fn.LastOpTime = time.Unix(0, item.LastOpTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
	} else {
		fn.LastOpTime = item.LastOpTimeStr
	}
	fn.IsDelete = 1
	fn.IsFolder = item.IsFolder || item.Md5 == ""
	if !fn.IsFolder {
		fn.FileType = util.GetExt(item.FileName)
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = int64(item.FileSize)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
		fn.DownloadUrl = item.DownloadURL
	} else {
		fn.IsFolder = true
		fn.FileType = ""
		fn.FileSize = 0
		fn.SizeFmt = "-"
	}
	return fn, nil
}

func (c Cloud189) File(account module.Account, fileId, path string) (module.FileNode, error) {
	fn := module.FileNode{}
	if fileId == "-11" {
		return module.FileNode{
			FileId:     "-11",
			FileName:   "root",
			FileSize:   0,
			IsFolder:   true,
			Path:       "/",
			LastOpTime: time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	}
	resp, err := c.ReqlReq(account).
		SetQueryParams(map[string]string{
			"fileId":     fileId,
			"filePath":   "",
			"pathList":   "1",
			"iconOption": "0",
		}).
		Get(cloud189API + "/getFileInfo.action")
	if err != nil {
		log.Errorln(err)
		return fn, err
	}
	item := Cloud189FileResp{}
	_ = jsoniter.Unmarshal(resp.Body(), &item)
	// PC API uses id/name/size fields
	if item.FileID == "" {
		item.FileID = jsoniter.Get(resp.Body(), "id").ToString()
	}
	if item.FileName == "" {
		item.FileName = jsoniter.Get(resp.Body(), "name").ToString()
	}
	if item.FileSize == 0 {
		item.FileSize = jsoniter.Get(resp.Body(), "fileSize").ToInt()
		if item.FileSize == 0 {
			item.FileSize = jsoniter.Get(resp.Body(), "size").ToInt()
		}
	}
	if item.ParentID == "" {
		item.ParentID = jsoniter.Get(resp.Body(), "parentId").ToString()
	}
	if item.CreateDate == "" {
		item.CreateDate = jsoniter.Get(resp.Body(), "createDate").ToString()
	}
	if item.LastOpTimeStr == "" {
		item.LastOpTimeStr = jsoniter.Get(resp.Body(), "lastOpTime").ToString()
	}
	item.Md5 = jsoniter.Get(resp.Body(), "md5").ToString()
	item.IsFolder = item.Md5 == ""
	fn, _ = c.ToFileNode2(item)
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	fn.AccountId = account.Id
	if account.RootId == fileId {
		fn.IsFolder = true
	}
	return fn, err
}

func (c Cloud189) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	sessionKey := CLoud189s[account.Id].SessionKey
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		reader := bytes.NewReader(file.Content)
		b := &bytes.Buffer{}
		writer := multipart.NewWriter(b)
		writer.WriteField("parentId", parentFileId)
		writer.WriteField("sessionKey", sessionKey)
		writer.WriteField("opertype", "1")
		writer.WriteField("fname", file.FileName)
		part, _ := writer.CreateFormFile("Filedata", file.FileName)
		io.Copy(part, reader)
		writer.Close()
		r, _ := http.NewRequest("POST", "https://hb02.upload.cloud.189.cn/v1/DCIWebUploadAction", b)
		r.Header.Add("Content-Type", writer.FormDataContentType())
		res, _ := http.DefaultClient.Do(r)
		defer res.Body.Close()
		log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "all files uploaded", nil
}

func (c Cloud189) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	session := CLoud189s[account.Id].Cloud189Session
	req := session.R()
	var err error
	var resp *resty.Response
	fn, err := c.File(account, fileId, "/")
	if err != nil {
		return false, "File remove error", err
	}
	if fn.IsFolder {
		resp, err = req.
			SetHeader("content-type", "application/x-www-form-urlencoded").
			SetFormData(map[string]string{
				"folderId":       fileId,
				"destFolderName": name,
			}).
			Post(cloud189API + "/renameFolder.action")
	} else {
		resp, err = req.
			SetHeader("content-type", "application/x-www-form-urlencoded").
			SetFormData(map[string]string{
				"fileId":       fileId,
				"destFileName": name,
			}).
			Post(cloud189API + "/renameFile.action")

	}
	log.Debug("File rename: ", resp.String())
	resCode := jsoniter.Get(resp.Body(), "res_code").ToInt()
	if resCode == 0 {
		return true, "File rename success", nil
	}
	return false, "File rename error", err
}

func (c Cloud189) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	session := CLoud189s[account.Id].Cloud189Session
	req := session.R()
	fn, err := c.File(account, fileId, "/")
	isFolder := 0
	if fn.IsFolder {
		isFolder = 1
	}
	info := []base.KV{
		{
			"fileId":   fileId,
			"fileName": fn.FileName,
			"isFolder": isFolder,
		},
	}
	if err != nil {
		return false, "File remove error", err
	}
	infoStr, err := jsoniter.MarshalToString(info)
	resp, err := req.
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"type":           "DELETE",
			"taskInfos":      infoStr,
			"targetFolderId": "",
		}).
		Post(cloud189API + "/batch/createBatchTask.action")
	log.Debug("File remove: ", resp.String())
	resCode := jsoniter.Get(resp.Body(), "res_code").ToInt()
	if resCode == 0 {
		return true, resp.String(), nil
	}
	return false, resp.String(), err
}

func (c Cloud189) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	session := CLoud189s[account.Id].Cloud189Session
	req := session.R()
	resp, err := req.
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"parentFolderId": parentFileId,
			"folderName":     name,
		}).
		Post(cloud189API + "/createFolder.action")
	log.Debug("Dir create: ", resp.String())
	resCode := jsoniter.Get(resp.Body(), "res_code").ToInt()
	if resCode == 0 {
		return true, resp.String(), nil
	}
	return false, resp.String(), err
}

func (c Cloud189) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	session := CLoud189s[account.Id].Cloud189Session
	req := session.R()
	fn, err := c.File(account, fileId, "/")
	isFolder := 0
	if fn.IsFolder {
		isFolder = 1
	}
	info := []base.KV{
		{
			"fileId":   fileId,
			"fileName": fn.FileName,
			"isFolder": isFolder,
		},
	}
	if err != nil {
		return false, "File copy error", err
	}
	infoStr, err := jsoniter.MarshalToString(info)
	resp, err := req.
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"type":           "MOVE",
			"taskInfos":      infoStr,
			"targetFolderId": targetFileId,
		}).
		Post(cloud189API + "/batch/createBatchTask.action")
	log.Debug("File move: ", resp.String())
	resCode := jsoniter.Get(resp.Body(), "res_code").ToInt()
	if resCode == 0 {
		return true, resp.String(), nil
	}
	return false, resp.String(), err
}

func (c Cloud189) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	session := CLoud189s[account.Id].Cloud189Session
	req := session.R()
	fn, err := c.File(account, fileId, "/")
	isFolder := 0
	if fn.IsFolder {
		isFolder = 1
	}
	info := []base.KV{
		{
			"fileId":   fileId,
			"fileName": fn.FileName,
			"isFolder": isFolder,
		},
	}
	if err != nil {
		return false, "File copy error", err
	}
	infoStr, err := jsoniter.MarshalToString(info)
	resp, err := req.
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"type":           "COPY",
			"taskInfos":      infoStr,
			"targetFolderId": targetFileId,
		}).
		Post(cloud189API + "/batch/createBatchTask.action")
	log.Debug("File copy: ", resp.String())
	resCode := jsoniter.Get(resp.Body(), "res_code").ToInt()
	if resCode == 0 {
		return true, resp.String(), nil
	}
	return false, resp.String(), err
}

func (c Cloud189) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	cloud189 := CLoud189s[account.Id]
	session := cloud189.Cloud189Session
	dRedirectRep, err := c.ReqlReq(account).
		SetQueryParam("fileId", fileId).
		Get(cloud189API + "/getFileDownloadUrl.action")
	if err != nil {
		log.Error(err)
		return "", err
	}
	resCode := jsoniter.Get(dRedirectRep.Body(), "res_code").ToInt()
	fileDownloadUrl := jsoniter.Get(dRedirectRep.Body(), "fileDownloadUrl").ToString()
	if fileDownloadUrl == "" {
		fileDownloadUrl = jsoniter.Get(dRedirectRep.Body(), "downloadUrl").ToString()
	}
	if resCode == 0 && fileDownloadUrl != "" {
		dRedirectRep, _ = session.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		})).R().Get(fileDownloadUrl)
		if dRedirectRep.StatusCode() == 302 || dRedirectRep.StatusCode() == 301 {
			return dRedirectRep.Header().Get("Location"), nil
		}
		return fileDownloadUrl, nil
	}
	errMsg := jsoniter.Get(dRedirectRep.Body(), "errorMsg").ToString()
	if errMsg == "" {
		errMsg = jsoniter.Get(dRedirectRep.Body(), "res_message").ToString()
	}
	if errMsg != "" {
		return "", fmt.Errorf(errMsg)
	}
	return "", err
}

func (c Cloud189) GetSpaceSzie(account module.Account) (int64, int64) {
	resp, err := c.ReqlReq(account).
		Get(cloud189API + "/getUserInfo.action")
	if err != nil {
		return 0, 0
	}
	available := jsoniter.Get(resp.Body(), "available").ToInt64()
	capacity := jsoniter.Get(resp.Body(), "capacity").ToInt64()
	return capacity, (capacity - available)
}

func (c Cloud189) RealUrl(ac module.Account, url string) string {
	/*if ac.SiteId != "" {
		url = strings.Replace(url, "/api/open", "/api/open/family", 1)
	}*/
	return url
}

func (c Cloud189) ReqlReq(ac module.Account) *resty.Request {
	cloud189, ok := CLoud189s[ac.Id]
	if !ok || cloud189.Cloud189Session == nil {
		return resty.New().R()
	}
	return cloud189.Cloud189Session.R()
}

type Cloud189FilesResp struct {
	ResCode    int        `json:"res_code"`
	ResMessage string     `json:"res_message"`
	FileListAO FileListAO `json:"fileListAO"`
	LastRev    int64      `json:"lastRev"`
}
type Icon struct {
	LargeURL string `json:"largeUrl"`
	SmallURL string `json:"smallUrl"`
}
type FileList struct {
	CreateDate  string `json:"createDate"`
	FileCata    int    `json:"fileCata"`
	Icon        Icon   `json:"icon,omitempty"`
	ID          int64  `json:"id"`
	LastOpTime  string `json:"lastOpTime"`
	Md5         string `json:"md5"`
	MediaType   int    `json:"mediaType"`
	Name        string `json:"name"`
	Rev         string `json:"rev"`
	Size        int    `json:"size"`
	StarLabel   int    `json:"starLabel"`
	Orientation int    `json:"orientation,omitempty"`
}
type FolderList struct {
	CreateDate   string `json:"createDate"`
	FileCata     int    `json:"fileCata"`
	FileCount    int    `json:"fileCount"`
	FileListSize int    `json:"fileListSize"`
	ID           int64  `json:"id"`
	LastOpTime   string `json:"lastOpTime"`
	Name         string `json:"name"`
	ParentID     int64  `json:"parentId"`
	Rev          string `json:"rev"`
	StarLabel    int    `json:"starLabel"`
}
type FileListAO struct {
	Count        int          `json:"count"`
	FileList     []FileList   `json:"fileList"`
	FileListSize int          `json:"fileListSize"`
	FolderList   []FolderList `json:"folderList"`
}

type Cloud189FileResp struct {
	ResCode       int    `json:"res_code"`
	ResMessage    string `json:"res_message"`
	CreateAccount string `json:"createAccount"`
	CreateTime    int64  `json:"createTime"`
	CreateDate    string `json:"createDate"`
	DownloadURL   string `json:"downloadUrl"`
	FileID        string `json:"fileId"`
	FileIDDigest  string `json:"fileIdDigest"`
	FileName      string `json:"fileName"`
	FileSize      int    `json:"fileSize"`
	FileType      string `json:"fileType"`
	IsFolder      bool   `json:"isFolder"`
	LastOpTime    int64  `json:"lastOpTime"`
	LastOpTimeStr string `json:"lastOpTimeStr"`
	MediaType     int    `json:"mediaType"`
	Md5           string `json:"md5"`
	ParentID      string `json:"parentId"`
}
