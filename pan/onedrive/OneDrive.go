package onedrive

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

var OneDrives = map[string]module.OneDriveAuthInfo{}
var zones = map[string]module.Zone{
	"onedrive": module.Zone{
		Login: "https://login.microsoftonline.com",
		Api:   "https://graph.microsoft.com",
		Desc:  "国际版",
	},
	"onedrive-cn": module.Zone{
		Login: "https://login.chinacloudapi.cn",
		Api:   "https://microsoftgraph.chinacloudapi.cn",
		Desc:  "世纪互联",
	},
}

func init() {
	base.RegisterPan("onedrive", &OneDrive{})
	base.RegisterPan("onedrive-cn", &OneDrive{})
}

type OneDrive struct{}

func (o OneDrive) IsLogin(account *module.Account) bool {
	return true
}

func (o OneDrive) AuthLogin(account *module.Account) (string, error) {
	var auth module.OneDriveAuthInfo
	_, err := base.Client.R().
		SetResult(&auth).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     account.User,
			"redirect_uri":  account.RedirectUri,
			"client_secret": account.Password,
			"refresh_token": account.RefreshToken,
			"grant_type":    "refresh_token",
		}).
		Post(zones[account.Mode].Login + "/common/oauth2/v2.0/token")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	auth.Mode = account.Mode
	driveId := o.GetDriveId(auth)
	auth.DriveId = driveId
	OneDrives[account.Id] = auth
	if account.SiteId != "" {
		auth.Sharepoint = o.SharePointId(account)
	}
	OneDrives[account.Id] = auth
	return auth.RefreshToken, nil
}

func (o OneDrive) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	od := OneDrives[account.Id]
	url := BuildODRequestUrl(od.Mode, SitePath(od), fileId, "/children?select=id,name,size,folder,@microsoft.graph.downloadUrl,lastModifiedDateTime,file,createdDateTime")
	var err error
	for {
		var filesResp OnedriveFilesResp
		_, err = base.Client.R().
			SetResult(&filesResp).
			SetAuthToken(od.AccessToken).
			SetHeader("Host", "graph.microsoft.com").
			Get(url)
		for _, f := range filesResp.Value {
			fn, _ := o.ToFileNode(f)
			if path == "/" {
				fn.Path = path + fn.FileName
			} else {
				fn.Path = path + "/" + fn.FileName
			}
			fn.AccountId = account.Id
			if fileId == "/" {
				fn.FileId = fileId + fn.FileName
			} else {
				fn.FileId = fileId + "/" + fn.FileName
			}
			fn.ParentId = fileId
			fn.ParentPath = path
			fileNodes = append(fileNodes, fn)
		}
		if filesResp.OdataNextLink != "" {
			//load next page
			url = filesResp.OdataNextLink
		} else {
			break
		}
	}
	return fileNodes, err
}

func (o OneDrive) File(account module.Account, fileId, path string) (module.FileNode, error) {
	od := OneDrives[account.Id]
	//var err error
	var fileResp Value
	_, err := base.Client.R().
		SetResult(&fileResp).
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		Get(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive/root:" + fileId + "?select=id,name,size,folder,@microsoft.graph.downloadUrl,lastModifiedDateTime,file,createdDateTime")
	if err != nil {
		log.Errorln(err)
	}
	fn, _ := o.ToFileNode(fileResp)
	fn.FileId = fileId
	fn.Path = path
	fn.AccountId = account.Id
	fn.ParentId = util.GetParentPath(fileId)
	fn.ParentPath = util.GetParentPath(path)
	return fn, err
}

func (o OneDrive) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	for _, file := range files {
		t1 := time.Now()
		bfs := ReadBlock(16384000, file) //15.625MB
		uploadUrl := CreateUploadSession(account.Id, filepath.Join(parentFileId, file.FileName))
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		for _, bf := range bfs {
			r, _ := http.NewRequest("PUT", uploadUrl, bytes.NewReader(bf.Content))
			r.Header.Add("Content-Length", strconv.FormatInt(file.FileSize, 10))
			r.Header.Add("Content-Range", bf.Name)
			res, _ := http.DefaultClient.Do(r)
			defer res.Body.Close()
		}
		log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "all files uploaded", nil
}

func (o OneDrive) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	od := OneDrives[account.Id]
	parentId := util.GetParentPath(fileId)
	targetItemId, _ := o.GetItemId(account, parentId)
	itemId, _ := o.GetItemId(account, fileId)
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		SetHeader("Content-Type", "application/json").
		SetBody(base.KV{
			"parentReference": base.KV{
				"id": targetItemId,
			},
			"name": name,
		}).
		Patch(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive/items/" + itemId)
	if err != nil {
		log.Errorln(err)
	}
	log.Debug("File rename: ", resp.String())
	if resp.StatusCode() == http.StatusOK {
		return true, "File rename success", nil
	}
	return false, "File rename error", err
}

func (o OneDrive) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	od := OneDrives[account.Id]
	itemId, _ := o.GetItemId(account, fileId)
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		Delete(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive/items/" + itemId)
	if err != nil {
		log.Errorln(err)
	}
	log.Debug("File remove: ", resp.String())
	if resp.StatusCode() == http.StatusNoContent {
		return true, "File remove success", nil
	}
	return false, "File remove error", err
}

func (o OneDrive) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	od := OneDrives[account.Id]
	itemId, _ := o.GetItemId(account, parentFileId)
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		SetHeader("Content-Type", "application/json").
		SetBody(base.KV{
			"name":                              name,
			"folder":                            base.KV{},
			"@microsoft.graph.conflictBehavior": "rename",
		}).
		Post(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive/items/" + itemId + "/children")
	if err != nil {
		log.Errorln(err)
	}
	log.Debug("Dir create: ", resp.String())
	if resp.StatusCode() == http.StatusCreated {
		return true, "Dir create success", nil
	}
	return false, "Dir create error", err
}

func (o OneDrive) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	od := OneDrives[account.Id]
	fileName := util.GetFileName(fileId)
	itemId, _ := o.GetItemId(account, fileId)
	targetItemId, _ := o.GetItemId(account, targetFileId)
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		SetHeader("Content-Type", "application/json").
		SetBody(base.KV{
			"parentReference": base.KV{
				"id": targetItemId,
			},
			"name": fileName,
		}).
		Patch(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive/items/" + itemId)
	if err != nil {
		log.Errorln(err)
	}
	log.Debug("File move: ", resp.String())
	if resp.StatusCode() == http.StatusOK {
		return true, "File move success", nil
	}
	return false, "File move error", err
}

func (o OneDrive) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	od := OneDrives[account.Id]
	fileName := util.GetFileName(fileId)
	itemId, _ := o.GetItemId(account, fileId)
	targetItemId, _ := o.GetItemId(account, targetFileId)
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		SetHeader("Content-Type", "application/json").
		SetBody(base.KV{
			"parentReference": base.KV{
				"driveId": od.DriveId,
				"id":      targetItemId,
			},
			"name": fileName,
		}).
		Post(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive/items/" + itemId + "/copy")
	if err != nil {
		log.Errorln(err)
	}
	log.Debug("File copy: ", resp.String())
	if resp.StatusCode() == http.StatusOK {
		return true, "File copy success", nil
	}
	return false, "File copy error", err
}

func (o OneDrive) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	fn, err := o.File(account, fileId, fileId)
	if err == nil {
		return fn.DownloadUrl, err
	} else {
		return "", err
	}
}

func (o OneDrive) GetSpaceSzie(account module.Account) (int64, int64) {
	od := OneDrives[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		Get(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive")
	if err != nil {
		log.Errorln(err)
	}
	total := jsoniter.Get(resp.Body(), "quota").Get("total").ToInt64()
	remaining := jsoniter.Get(resp.Body(), "quota").Get("remaining").ToInt64()
	return total, total - remaining
}

func (o OneDrive) ToFileNode(f Value) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = f.ID
	fn.FileName = f.Name
	fn.CreateTime = util.UTCTimeFormat(f.CreatedDateTime)
	fn.LastOpTime = util.UTCTimeFormat(f.LastModifiedDateTime)
	fn.IsDelete = 1
	if f.MicrosoftGraphDownloadURL == "" {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	} else {
		fn.IsFolder = false
		fn.FileType = util.GetExt(fn.FileName)
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = int64(f.Size)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
		fn.Thumbnail = ""
		fn.DownloadUrl = f.MicrosoftGraphDownloadURL
	}
	return fn, nil
}

func (o OneDrive) GetDriveId(od module.OneDriveAuthInfo) string {
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		Get(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive")
	if err != nil {
		log.Errorln(err)
	}
	id := jsoniter.Get(resp.Body(), "id").ToString()
	return id
}

func (o OneDrive) GetItemId(account module.Account, fileId string) (string, error) {
	od := OneDrives[account.Id]
	//var err error
	var fileResp Value
	_, err := base.Client.R().
		SetResult(&fileResp).
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		Get(zones[od.Mode].Api + "/v1.0/" + SitePath(od) + "/drive/root:" + fileId + "?select=id,name,size,folder,@microsoft.graph.downloadUrl,lastModifiedDateTime,file,createdDateTime")
	if err != nil {
		log.Errorln(err)
	}
	return fileResp.ID, err
}

func (o OneDrive) SharePointId(account *module.Account) string {
	od := OneDrives[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Content-Type", "application/json").
		Get(zones[account.Mode].Api + "/v1.0/sites/" + account.SiteId)
	if err != nil {
		log.Errorln(err)
		return ""
	}
	id := jsoniter.Get(resp.Body(), "id").ToString()
	return id
}

func SitePath(od module.OneDriveAuthInfo) string {
	if od.Sharepoint != "" {
		return "/sites/" + od.Sharepoint
	}
	return "/me"
}

func OneExchangeToken(zone, clientId, redirectUri, clientSecret, code string) string {
	if zone == "" {
		zone = "onedrive"
	}
	resp, err := base.Client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     clientId,
			"redirect_uri":  redirectUri,
			"client_secret": clientSecret,
			"code":          code,
			"grant_type":    "authorization_code",
		}).
		Post(zones[zone].Login + "/common/oauth2/v2.0/token")
	if err != nil {
		log.Error(err)
		return ""
	}
	return resp.String()
}
func OneGetRefreshToken(zone, clientId, redirectUri, clientSecret, refreshToken string) string {
	if zone == "" {
		zone = "onedrive"
	}
	resp, err := base.Client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     clientId,
			"redirect_uri":  redirectUri,
			"client_secret": clientSecret,
			"refresh_token": refreshToken,
			"grant_type":    "refresh_token",
		}).
		Post(zones[zone].Login + "/common/oauth2/v2.0/token")
	if err != nil {
		log.Error(err)
		return ""
	}
	return resp.String()
}

func CreateUploadSession(accountId, filePath string) string {
	od := OneDrives[accountId]
	resp, err := base.Client.R().
		SetAuthToken(od.AccessToken).
		SetHeader("Host", "graph.microsoft.com").
		Post(zones[od.Mode].Api + "/v1.0" + SitePath(od) + "/drive/root:" + filePath + ":/createUploadSession")
	if err != nil {
		log.Error(err)
	}
	if resp.StatusCode() == 409 {
		log.Debugf("被请求的资源的当前状态之间存在冲突，请求无法完成，重新提交上传请求")
		return ""
	}
	log.Debugf("[OneDrive]get upload session: %s", resp.String())
	return jsoniter.Get(resp.Body(), "uploadUrl").ToString()
}

func ReadBlock(fileChunkSize int, file *module.UploadInfo) []BlockFile {
	bfs := []BlockFile{}
	cbs := util.ChunkBytes(file.Content, fileChunkSize)
	off := int64(0)
	for i, cb := range cbs {
		partSize := int64(math.Min(float64(fileChunkSize), float64(file.FileSize-int64(i*fileChunkSize))))
		contentRange := fmt.Sprintf("bytes %d-%d/%d", off, off+partSize-1, file.FileSize)
		contentRange2 := fmt.Sprintf("bytes=%d-%d", off, off+partSize-1)
		bfs = append(bfs, BlockFile{cb, contentRange, contentRange2})
		off += partSize
	}
	return bfs
}

func BuildODRequestUrl(mode, sitePath, path, query string) string {
	if path != "" && path != "/" {
		return zones[mode].Api + "/v1.0" + sitePath + "/drive/root:" + path + ":" + query
	}
	return zones[mode].Api + "/v1.0" + sitePath + "/drive/root" + query
}

type OnedriveFilesResp struct {
	OdataContext  string  `json:"@odata.context"`
	OdataNextLink string  `json:"@odata.nextLink"`
	Value         []Value `json:"value"`
}
type Folder struct {
	ChildCount int `json:"childCount"`
}
type Hashes struct {
	QuickXorHash string `json:"quickXorHash"`
}
type File struct {
	MimeType string `json:"mimeType"`
	Hashes   Hashes `json:"hashes"`
}
type Value struct {
	ID                        string `json:"id"`
	LastModifiedDateTime      string `json:"lastModifiedDateTime"`
	CreatedDateTime           string `json:"createdDateTime"`
	Name                      string `json:"name"`
	Size                      int    `json:"size"`
	Folder                    Folder `json:"folder,omitempty"`
	MicrosoftGraphDownloadURL string `json:"@microsoft.graph.downloadUrl,omitempty"`
	File                      File   `json:"file,omitempty"`
}
type BlockFile struct {
	Content []byte
	Name    string
	Range2  string
}
