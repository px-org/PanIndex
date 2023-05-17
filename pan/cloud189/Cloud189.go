package cloud189

import (
	"bytes"
	"fmt"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

var CLoud189s = map[string]module.Cloud189{}

func init() {
	base.RegisterPan("cloud189", &Cloud189{})
}

type Cloud189 struct{}

func (c Cloud189) IsLogin(account *module.Account) bool {
	resp, err := c.ReqlReq(*account).Get("https://cloud.189.cn/v2/getLoginedInfos.action?showPC=true")
	if err == nil && resp != nil && resp.String() != "" && jsoniter.Valid(resp.Body()) && jsoniter.Get(resp.Body(), "errorMsg").ToString() == "" {
		return true
	} else {
		if jsoniter.Get(resp.Body(), "errorCode").ToString() == "InvalidSessionKey" {
			return false
		} else {
			return true
		}
	}
	return false
}

func (c Cloud189) AuthLogin(account *module.Account) (string, error) {
	client := resty.New()
	tempUrl := "https://cloud.189.cn/api/portal/loginUrl.action?redirectURL=https%3A%2F%2Fcloud.189.cn%2Fmain.action"
	resp := &resty.Response{}
	var err error
	lt := ""
	reqId := ""
	resp, err = client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		if req.URL.Query().Get("lt") != "" {
			lt = req.URL.Query().Get("lt")
		}
		if req.URL.Query().Get("reqId") != "" {
			reqId = req.URL.Query().Get("reqId")
		}
		return nil
	})).R().Get(tempUrl)
	if err != nil {
		log.Error(err)
		return "", err
	}
	cookies := ""
	for _, cookie := range resp.Cookies() {
		cookies += cookie.Name + "=" + cookie.Value + ";"
	}
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))
	appConfResp, err := client.R().
		SetHeaders(map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:76.0) Gecko/20100101 Firefox/74.0",
			"Referer":    resp.Request.URL,
			"Cookie":     cookies,
			"lt":         lt,
			"reqId":      reqId,
		}).
		SetFormData(map[string]string{
			"appKey": "cloud",
			"version": "2.0",
		}).
		Post("https://open.e.189.cn/api/logbox/oauth2/appConf.do")
	if err != nil {
		log.Error(err)
		return "", err
	}
	accountType := jsoniter.Get(appConfResp.Body(), "data", "accountType").ToString()
	clientType := jsoniter.Get(appConfResp.Body(), "data", "clientType").ToString()
	paramId := jsoniter.Get(appConfResp.Body(), "data", "paramId").ToString()
	mailSuffix := jsoniter.Get(appConfResp.Body(), "data", "mailSuffix").ToString()
	returnUrl := jsoniter.Get(appConfResp.Body(), "data", "returnUrl").ToString()
	encryptConfResp, err := client.R().
		SetHeaders(map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:76.0) Gecko/20100101 Firefox/74.0",
			"Referer":    "https://open.e.189.cn/api/logbox/separate/web/index.html",
			"Cookie":     cookies,
		}).
		SetFormData(map[string]string{
			"appId":      "cloud",
		}).
		Post("https://open.e.189.cn/api/logbox/config/encryptConf.do")
	if err != nil {
		log.Error(err)
		return "", err
	}
	resCode := jsoniter.Get(encryptConfResp.Body(), "result").ToInt()
	if resCode != 0 {
		log.Error("Failed to get encrypt config")
		return "", fmt.Errorf("Failed to get encrypt config")
	}
	pubKey := jsoniter.Get(encryptConfResp.Body(), "data", "pubKey").ToString()
	pre := jsoniter.Get(encryptConfResp.Body(), "data", "pre").ToString()
	vCodeRS := ""
	userRsa := util.RsaEncode([]byte(account.User), pubKey)
	passwordRsa := util.RsaEncode([]byte(account.Password), pubKey)
	loginResp, _ := client.R().
		SetHeaders(map[string]string{
			"lt":         lt,
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/76.0",
			"Referer":    "https://open.e.189.cn/",
		}).
		SetFormData(map[string]string{
			"version":      "v2.0",
			"appKey":       "cloud",
			"accountType":  accountType,
			"userName":     pre + userRsa,
			"epd":          pre + passwordRsa,
			"validateCode": vCodeRS,
			"captchaToken": "",
			"returnUrl":    returnUrl,
			"mailSuffix":   mailSuffix,
			"paramId":      paramId,
			"clientType":   clientType,
			"dynamicCheck": "FALSE",
			"cb_SaveName":  "1",
			"isOauth2":     "false",
		}).
		Post("https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do")
	restCode := jsoniter.Get([]byte(loginResp.String()), "result").ToInt()
	if restCode == 0 {
		toUrl := jsoniter.Get([]byte(loginResp.String()), "toUrl").ToString()
		resp, err = client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(3)).R().Get(toUrl)
		if err != nil {
			return "", err
		}
		resp, err = client.R().Get("https://cloud.189.cn/v2/getUserBriefInfo.action?noCache=" + util.Random())
		if err != nil {
			return "", err
		}
		sessionKey := jsoniter.Get(resp.Body(), "sessionKey").ToString()
		resp, err = client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(3)).R().Get("https://api.cloud.189.cn/open/oauth2/ssoH5.action")
		v, _ := url.ParseQuery(resp.Header().Get("location"))
		accessToken := v.Get("https://h5.cloud.189.cn/index.html?accessToken")
		CLoud189s[account.Id] = module.Cloud189{client, sessionKey, accessToken, account.RootId, ""}
		return fmt.Sprintf("cloud189 login success [%s]", time.Now()), nil
	} else if restCode == -2 {
		return "", base.LoginCaptcha
	} else {
		return "", nil
	}
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
			Get(c.RealUrl(account, "https://cloud.189.cn/api/open/file/listFiles.action"))
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
	fn.CreateTime = time.Unix(0, item.CreateTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
	fn.LastOpTime = time.Unix(0, item.LastOpTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
	fn.IsDelete = 1
	fn.IsFolder = item.IsFolder
	if !fn.IsFolder {
		fn.FileType = util.GetExt(item.FileName)
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = int64(item.FileSize)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
		fn.DownloadUrl = item.DownloadURL
	} else {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	}
	return fn, nil
}

func (c Cloud189) File(account module.Account, fileId, path string) (module.FileNode, error) {
	cloud189 := CLoud189s[account.Id]
	item := Cloud189FileResp{}
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
	_, err := cloud189.Cloud189Session.R().
		SetResult(&item).
		SetQueryParams(map[string]string{
			"noCache": util.Random(),
			"fileId":  fileId,
		}).
		Get("https://cloud.189.cn/api/portal/getFileInfo.action")
	if err != nil {
		log.Errorln(err)
		return fn, err
	}
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
			SetQueryParam("noCache", util.Random()).
			SetHeader("content-type", "application/json;charset=UTF-8").
			SetFormData(map[string]string{
				"folderId":       fileId,
				"destFolderName": name,
			}).
			Post("https://cloud.189.cn/api/open/file/renameFolder.action")
	} else {
		resp, err = req.
			SetQueryParam("noCache", util.Random()).
			SetHeader("content-type", "application/json;charset=UTF-8").
			SetFormData(map[string]string{
				"fileId":       fileId,
				"destFileName": name,
			}).
			Post("https://cloud.189.cn/api/open/file/renameFile.action")

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
		SetQueryParam("noCache", util.Random()).
		SetHeader("content-type", "application/json;charset=UTF-8").
		SetFormData(map[string]string{
			"type":           "DELETE",
			"taskInfos":      infoStr,
			"targetFolderId": "",
		}).
		Post("https://cloud.189.cn/api/open/batch/createBatchTask.action")
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
		SetQueryParam("noCache", util.Random()).
		SetHeader("content-type", "application/json;charset=UTF-8").
		SetFormData(map[string]string{
			"parentFolderId": parentFileId,
			"folderName":     name,
		}).
		Post("https://cloud.189.cn/api/open/file/createFolder.action")
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
		SetQueryParam("noCache", util.Random()).
		SetHeader("content-type", "application/json;charset=UTF-8").
		SetFormData(map[string]string{
			"type":           "MOVE",
			"taskInfos":      infoStr,
			"targetFolderId": targetFileId,
		}).
		Post("https://cloud.189.cn/api/open/batch/createBatchTask.action")
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
		SetQueryParam("noCache", util.Random()).
		SetHeader("content-type", "application/json;charset=UTF-8").
		SetFormData(map[string]string{
			"type":           "COPY",
			"taskInfos":      infoStr,
			"targetFolderId": targetFileId,
		}).
		Post("https://cloud.189.cn/api/open/batch/createBatchTask.action")
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
	timestamp := fmt.Sprintf("%d", int(time.Now().UTC().UnixNano()/1e6))
	dRedirectRep, err := resty.New().R().
		SetHeaders(map[string]string{
			"accept":      "application/json;charset=UTF-8",
			"accesstoken": cloud189.AccessToken,
			"sign-type":   "1",
			"signature": util.Md5Params(map[string]string{
				"AccessToken": cloud189.AccessToken,
				"fileId":      fileId,
				"Timestamp":   timestamp,
			}),
			"timestamp": timestamp,
		}).
		SetQueryParam("fileId", fileId).
		Get("https://api.cloud.189.cn/open/file/getFileDownloadUrl.action")
	if err != nil {
		log.Error(err)
		return "", err
	}
	resCode := jsoniter.Get(dRedirectRep.Body(), "res_code").ToInt()
	if resCode == 0 {
		fileDownloadUrl := jsoniter.Get(dRedirectRep.Body(), "fileDownloadUrl").ToString()
		dRedirectRep, _ = session.SetRedirectPolicy(resty.FlexibleRedirectPolicy(0)).R().Get(fileDownloadUrl)
		if dRedirectRep.StatusCode() == 302 {
			return dRedirectRep.Header().Get("Location"), err
		}
	}
	return "", err
}

func (c Cloud189) GetSpaceSzie(account module.Account) (int64, int64) {
	session := CLoud189s[account.Id].Cloud189Session
	req := session.R()
	resp, err := req.
		SetQueryParam("noCache", util.Random()).
		SetHeader("content-type", "application/json;charset=UTF-8").
		Get(c.RealUrl(account, "https://cloud.189.cn/api/open/user/getUserInfoForPortal.action"))
	if err != nil {
		return 0, 0
	} else {
		available := jsoniter.Get(resp.Body(), "available").ToInt64()
		capacity := jsoniter.Get(resp.Body(), "capacity").ToInt64()
		return capacity, (capacity - available)
	}
}

func (c Cloud189) RealUrl(ac module.Account, url string) string {
	/*if ac.SiteId != "" {
		url = strings.Replace(url, "/api/open", "/api/open/family", 1)
	}*/
	return url
}

func (c Cloud189) ReqlReq(ac module.Account) *resty.Request {
	cloud189 := CLoud189s[ac.Id]
	r := cloud189.Cloud189Session.R()
	/*if ac.SiteId != "" {
		r.SetQueryParam("familyId", ac.SiteId)
		r.SetFormData(map[string]string{
			"familyId": ac.SiteId,
		})
	}*/
	return r
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
	DownloadURL   string `json:"downloadUrl"`
	FileID        string `json:"fileId"`
	FileIDDigest  string `json:"fileIdDigest"`
	FileName      string `json:"fileName"`
	FileSize      int    `json:"fileSize"`
	FileType      string `json:"fileType"`
	IsFolder      bool   `json:"isFolder"`
	LastOpTime    int64  `json:"lastOpTime"`
	MediaType     int    `json:"mediaType"`
	ParentID      string `json:"parentId"`
}
