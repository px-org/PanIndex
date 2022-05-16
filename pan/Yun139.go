package pan

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/util"
	"github.com/libsgh/nic"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

func init() {
	RegisterPan("yun139", &Yun139{})
}

type Yun139 struct{}

func (y Yun139) IsLogin(account *module.Account) bool {
	body := KV{
		"qryUserExternInfoReq": KV{
			"commonAccountInfo": KV{"account": account.User, "accountType": 1},
		},
	}
	resp, err := resty.New().R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/user/v1.0/qryUserExternInfo")
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if err != nil || status {
		return true
	}
	return false
}

func (y Yun139) AuthLogin(account *module.Account) (string, error) {
	isLogin, err := y.LoginCheck(*account)
	if isLogin {
		return account.Password, err
	}
	return "", err
}

func (y Yun139) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	size := 200
	offset := 0
	var err error
	for {
		body := KV{
			"catalogID":         fileId,
			"sortDirection":     1,
			"filterType":        0,
			"catalogSortType":   0,
			"contentSortType":   0,
			"startNumber":       offset + 1,
			"endNumber":         offset + size,
			"commonAccountInfo": nic.KV{"account": account.User, "accountType": 1},
		}
		var filesResp Yun139FilesResp
		_, err = resty.New().R().
			SetResult(&filesResp).
			SetHeaders(y.CreateHeaders(body, account.Password)).
			SetBody(body).
			Post("https://yun.139.com/orchestration/personalCloud/catalog/v1.0/getDisk")
		if err == nil && filesResp.Success {
			for _, folder := range filesResp.Data.GetDiskResult.CatalogList {
				fn, _ := y.ToFolderNode(folder)
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
			for _, file := range filesResp.Data.GetDiskResult.ContentList {
				fn, _ := y.ToFileNode(file)
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
			if filesResp.Data.GetDiskResult.IsCompleted == 1 {
				break
			} else {
				offset = offset + size
			}
		} else {
			break
		}
	}
	return fileNodes, err
}

func (y Yun139) ToFolderNode(item CatalogList) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = item.CatalogID
	fn.FileName = item.CatalogName
	fn.CreateTime = TimeFormat139(item.CreateTime)
	fn.LastOpTime = TimeFormat139(item.UpdateTime)
	fn.ParentId = item.ParentCatalogID
	fn.IsDelete = 1
	fn.IsFolder = true
	fn.FileType = ""
	fn.FileSize = 0
	fn.SizeFmt = "-"
	return fn, nil
}

func (y Yun139) ToFileNode(item ContentList) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = item.ContentID
	fn.FileName = item.ContentName
	fn.CreateTime = TimeFormat139(item.UploadTime)
	fn.LastOpTime = TimeFormat139(item.UpdateTime)
	fn.IsDelete = 1
	fn.IsFolder = false
	fn.FileType = util.GetExt(item.ContentName)
	fn.ViewType = util.GetViewType(fn.FileType)
	fn.FileSize = int64(item.ContentSize)
	fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	fn.Thumbnail = item.ThumbnailURL
	return fn, nil
}

func (y Yun139) File(account module.Account, fileId, path string) (module.FileNode, error) {
	_, _, p, _ := util.ParseFullPath(path, "")
	if fileId == account.RootId {
		return module.FileNode{
			Id:         uuid.NewV4().String(),
			FileId:     fileId,
			FileName:   "root",
			FileSize:   0,
			IsFolder:   true,
			Path:       "/",
			ParentPath: "",
			ParentId:   "",
			LastOpTime: account.LastOpTime,
		}, nil
	}
	fn := y.GetFileFromApi(account, p)
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	return fn, nil
}

func (y Yun139) LoopGetFileId(ac module.Account, fileId, path, filePath string) (module.FileNode, bool) {
	fileName := util.GetFileName(path)
	fns, _ := y.Files(ac, fileId, util.GetParentPath(path), "", "")
	fn := y.GetCurrentFile(fileName, fns)
	if fn.FileId != "" {
		if fn.Path == filePath {
			return fn, true
		}
		return fn, false
	} else {
		return module.FileNode{}, false
	}
}

func (y Yun139) GetFileFromApi(ac module.Account, path string) module.FileNode {
	fileId := ac.RootId
	paths := util.GetPrePath(path)
	for _, pathMap := range paths {
		fn, ok := y.LoopGetFileId(ac, fileId, pathMap["PathUrl"], path)
		fileId = fn.FileId
		if ok {
			return fn
		}
	}
	return module.FileNode{}
}

func (y Yun139) GetCurrentFile(pathName string, fns []module.FileNode) module.FileNode {
	for _, fn := range fns {
		if fn.FileName == pathName {
			return fn
		}
	}
	return module.FileNode{}
}

func (y Yun139) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		textQuoted := strconv.QuoteToASCII(file.FileName)
		textUnquoted := textQuoted[1 : len(textQuoted)-1]
		digest := fmt.Sprintf("%x", md5.Sum(file.Content))
		body := KV{
			"fileCount":         1,
			"parentCatalogID":   parentFileId,
			"manualRename":      2,
			"newCatalogName":    "",
			"operation":         0,
			"totalSize":         file.FileSize,
			"commonAccountInfo": KV{"account": account.User, "accountType": 1},
			"uploadContentList": []KV{{
				"contentName": file.FileName,
				"contentSize": file.FileSize,
				"digest":      digest,
			}},
		}
		resp, err := resty.New().R().
			SetHeaders(y.CreateHeaders(body, account.Password)).
			SetBody(body).
			Post("https://yun.139.com/orchestration/personalCloud/uploadAndDownload/v1.0/pcUploadFileRequest")
		if err != nil {
			return false, "File upload failed", err
		}
		status := jsoniter.Get(resp.Body(), "success").ToBool()
		if status {
			isNeedUpload := jsoniter.Get(resp.Body(), "data").
				Get("uploadResult").
				Get("newContentIDList").
				Get(0).
				Get("isNeedUpload").ToInt()
			if isNeedUpload == 1 {
				//need upload
				uploadUrl := jsoniter.Get(resp.Body(), "data").
					Get("uploadResult").
					Get("redirectionUrl").ToString()
				uploadTaskID := jsoniter.Get(resp.Body(), "data").
					Get("uploadResult").
					Get("uploadTaskID").ToString()
				bfs := ReadBlock(16384000, file)
				for _, bf := range bfs {
					r, _ := http.NewRequest("POST", uploadUrl, bytes.NewReader(bf.Content))
					r.Header.Add("uploadtaskID", uploadTaskID)
					r.Header.Add("rangeType", "0")
					r.Header.Add("Range", bf.Range2)
					r.Header.Add("Content-Type", "text/plain;name="+textUnquoted)
					r.Header.Add("contentSize", fmt.Sprintf("%d", len(bf.Content)))
					r.Header.Add("Content-Length", fmt.Sprintf("%d", len(bf.Content)))
					r.Header.Add("Referer", "https://yun.139.com/")
					r.Header.Add("x-SvcType", "1")
					r.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
					res, _ := http.DefaultClient.Do(r)
					defer res.Body.Close()
				}
			} else {
				log.Debug("[yun139] fast upload success")
			}
		}
		log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "all files uploaded", nil
}

func (y Yun139) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	body := KV{
		"catalogName":       name,
		"catalogID":         fileId,
		"commonAccountInfo": KV{"account": account.User, "accountType": 1},
	}
	resp, err := resty.New().R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/catalog/v1.0/updateCatalogInfo")
	if err != nil {
		return false, "Dir rename failed", err
	}
	log.Debug("Dir rename: ", resp.String())
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if !status {
		body = KV{
			"contentName":       name,
			"contentID":         fileId,
			"commonAccountInfo": KV{"account": account.User, "accountType": 1},
		}
		resp, err = resty.New().R().
			SetHeaders(y.CreateHeaders(body, account.Password)).
			SetBody(body).
			Post("https://yun.139.com/orchestration/personalCloud/content/v1.0/updateContentInfo")
		if err != nil {
			return false, "File rename failed", err
		}
		log.Debug("File rename: ", resp.String())
		status = jsoniter.Get(resp.Body(), "success").ToBool()
	}
	if status {
		return true, resp.String(), err
	}
	return true, resp.String(), err
}

func (y Yun139) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	resp, err := y.CommonReq(account, 201, 2, []string{fileId}, []string{}, "")
	resp, err = y.CommonReq(account, 201, 2, []string{}, []string{fileId}, "")
	if err != nil {
		return false, "File remove failed", err
	}
	log.Debug("File remove: ", resp.String())
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		return true, resp.String(), err
	}
	return true, resp.String(), err
}

func (y Yun139) QueryTaskInfo(account module.Account, taskId string) {
	body := KV{
		"queryBatchOprTaskDetailReq": KV{
			"taskID":            taskId,
			"commonAccountInfo": KV{"account": account.User, "accountType": 1},
		},
	}
	resp, _ := resty.New().R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/batchOprTask/v1.0/queryBatchOprTaskDetail")
	log.Debug("Task query:", resp.String())
}

func (y Yun139) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	body := KV{
		"createCatalogExtReq": KV{
			"newCatalogName":    name,
			"parentCatalogID":   parentFileId,
			"commonAccountInfo": KV{"account": account.User, "accountType": 1},
		},
	}
	resp, err := resty.New().R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/catalog/v1.0/createCatalogExt")
	if err != nil {
		return false, "Dir create failed", err
	}
	log.Debug("Dir create: ", resp.String())
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		return true, resp.String(), err
	}
	return true, resp.String(), err
}

func (y Yun139) CommonReq(account module.Account, actionType, taskType int, catalogIds, contentIds []string, newCatalogID string) (*resty.Response, error) {
	body := KV{
		"createBatchOprTaskReq": KV{
			"actionType":        actionType,
			"taskType":          taskType,
			"commonAccountInfo": KV{"account": account.User, "accountType": 1},
			"taskInfo": KV{
				"catalogInfoList": catalogIds,
				"contentInfoList": contentIds,
				"newCatalogID":    newCatalogID,
			},
		},
	}
	resp, err := resty.New().R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/batchOprTask/v1.0/createBatchOprTask")
	return resp, err
}

func (y Yun139) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	resp, err := y.CommonReq(account, 304, 3, []string{}, []string{fileId}, targetFileId)
	resp, err = y.CommonReq(account, 304, 3, []string{fileId}, []string{}, targetFileId)
	if err != nil {
		return false, "File move failed", err
	}
	log.Debug("File move: ", resp.String())
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		return true, resp.String(), err
	}
	return true, resp.String(), err
}

func (y Yun139) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	resp, err := y.CommonReq(account, 309, 1, []string{}, []string{fileId}, targetFileId)
	resp, err = y.CommonReq(account, 309, 1, []string{fileId}, []string{}, targetFileId)
	if err != nil {
		return false, "File copy failed", err
	}
	log.Debug("File copy: ", resp.String())
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		return true, resp.String(), err
	}
	return true, resp.String(), err
}

func (y Yun139) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	body := KV{
		"appName":           "",
		"contentID":         fileId,
		"commonAccountInfo": KV{"account": account.User, "accountType": 1},
	}
	resp, err := resty.New().R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/uploadAndDownload/v1.0/downloadRequest")
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		return jsoniter.Get(resp.Body(), "data").Get("downloadURL").ToString(), err
	}
	return "", err
}

func (y Yun139) GetSpaceSzie(account module.Account) (int64, int64) {
	body := KV{
		"account":           account.User,
		"commonAccountInfo": KV{"account": account.User, "accountType": 1},
	}
	resp, err := resty.New().R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/user/v1.0/getDiskInfo")
	if err != nil {
		return 0, 0
	}
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		//MB
		total := jsoniter.Get(resp.Body(), "data").Get("diskInfo").Get("diskSize").ToInt64()
		free := jsoniter.Get(resp.Body(), "data").Get("diskInfo").Get("freeDiskSize").ToInt64()
		return total * 1024 * 1024, (total - free) * 1024 * 1024
	}
	return 0, 0
}

func (y Yun139) LoginCheck(account module.Account) (bool, error) {
	body := KV{
		"qryUserExternInfoReq": KV{
			"commonAccountInfo": KV{
				"account":     account.User,
				"accountType": 1,
			},
		},
	}
	resp, err := client.R().
		SetHeaders(y.CreateHeaders(body, account.Password)).
		SetBody(body).
		Post("https://yun.139.com/orchestration/personalCloud/user/v1.0/qryUserExternInfo")
	if err == nil && jsoniter.Get(resp.Body(), "success").ToBool() {
		return true, err
	}
	return false, err
}

func (y Yun139) CreateHeaders(body KV, cookie string) map[string]string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	key := util.GetRandomStr(16)
	json, _ := jsoniter.MarshalToString(body)
	sign := util.Yun139Sign(timestamp, key, json)
	headers := map[string]string{
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

func TimeFormat139(timeStr string) string {
	t, _ := time.ParseInLocation("20060102150405", timeStr, time.Local)
	timeFormat := time.Unix(t.Unix(), 0).Format("2006-01-02 15:04:05")
	return timeFormat
}

type Yun139FilesResp struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}
type Result struct {
	ResultCode string      `json:"resultCode"`
	ResultDesc interface{} `json:"resultDesc"`
}
type CatalogList struct {
	CatalogID       string      `json:"catalogID"`
	CatalogName     string      `json:"catalogName"`
	CatalogType     int         `json:"catalogType"`
	CreateTime      string      `json:"createTime"`
	UpdateTime      string      `json:"updateTime"`
	IsShared        bool        `json:"isShared"`
	CatalogLevel    int         `json:"catalogLevel"`
	ShareDoneeCount int         `json:"shareDoneeCount"`
	OpenType        int         `json:"openType"`
	ParentCatalogID string      `json:"parentCatalogId"`
	DirEtag         int         `json:"dirEtag"`
	Tombstoned      int         `json:"tombstoned"`
	ProxyID         interface{} `json:"proxyID"`
	Moved           int         `json:"moved"`
	IsFixedDir      int         `json:"isFixedDir"`
	IsSynced        interface{} `json:"isSynced"`
	Owner           string      `json:"owner"`
	Modifier        string      `json:"modifier"`
	Path            string      `json:"path"`
	ShareType       int         `json:"shareType"`
	SoftLink        interface{} `json:"softLink"`
	ExtProp1        interface{} `json:"extProp1"`
	ExtProp2        interface{} `json:"extProp2"`
	ExtProp3        string      `json:"extProp3"`
	ExtProp4        interface{} `json:"extProp4"`
	ExtProp5        interface{} `json:"extProp5"`
	ETagOprType     int         `json:"ETagOprType"`
}
type ExtInfo struct {
	Uploader string `json:"uploader"`
}
type Exif struct {
	CreateTime    string      `json:"createTime"`
	Longitude     interface{} `json:"longitude"`
	Latitude      interface{} `json:"latitude"`
	LocalSaveTime interface{} `json:"localSaveTime"`
}
type ContentList struct {
	ContentID       string      `json:"contentID"`
	ContentName     string      `json:"contentName"`
	ContentSuffix   string      `json:"contentSuffix"`
	ContentSize     int         `json:"contentSize"`
	ContentDesc     string      `json:"contentDesc"`
	ContentType     int         `json:"contentType"`
	ContentOrigin   int         `json:"contentOrigin"`
	UpdateTime      string      `json:"updateTime"`
	CommentCount    int         `json:"commentCount"`
	ThumbnailURL    string      `json:"thumbnailURL"`
	BigthumbnailURL string      `json:"bigthumbnailURL"`
	PresentURL      string      `json:"presentURL"`
	PresentLURL     string      `json:"presentLURL"`
	PresentHURL     string      `json:"presentHURL"`
	ContentTAGList  interface{} `json:"contentTAGList"`
	ShareDoneeCount int         `json:"shareDoneeCount"`
	Safestate       int         `json:"safestate"`
	Transferstate   int         `json:"transferstate"`
	IsFocusContent  int         `json:"isFocusContent"`
	UpdateShareTime interface{} `json:"updateShareTime"`
	UploadTime      string      `json:"uploadTime"`
	OpenType        int         `json:"openType"`
	AuditResult     int         `json:"auditResult"`
	ParentCatalogID string      `json:"parentCatalogId"`
	Channel         string      `json:"channel"`
	GeoLocFlag      string      `json:"geoLocFlag"`
	Digest          string      `json:"digest"`
	Version         string      `json:"version"`
	FileEtag        string      `json:"fileEtag"`
	FileVersion     string      `json:"fileVersion"`
	Tombstoned      int         `json:"tombstoned"`
	ProxyID         string      `json:"proxyID"`
	Moved           int         `json:"moved"`
	MidthumbnailURL string      `json:"midthumbnailURL"`
	Owner           string      `json:"owner"`
	Modifier        string      `json:"modifier"`
	ShareType       int         `json:"shareType"`
	ExtInfo         ExtInfo     `json:"extInfo"`
	Exif            Exif        `json:"exif"`
	CollectionFlag  interface{} `json:"collectionFlag"`
	TreeInfo        interface{} `json:"treeInfo"`
	IsShared        bool        `json:"isShared"`
	ETagOprType     int         `json:"ETagOprType"`
}
type GetDiskResult struct {
	ParentCatalogID string        `json:"parentCatalogID"`
	NodeCount       int           `json:"nodeCount"`
	CatalogList     []CatalogList `json:"catalogList"`
	ContentList     []ContentList `json:"contentList"`
	IsCompleted     int           `json:"isCompleted"`
}
type Data struct {
	Result        Result        `json:"result"`
	GetDiskResult GetDiskResult `json:"getDiskResult"`
}
