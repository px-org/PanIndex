package googledrive

import (
	"bytes"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/pan/onedrive"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var GoogleDrives = map[string]module.GoogleDriveAuthInfo{}

const (
	PartSize = 1 * 1024 * 1024
)

func init() {
	base.RegisterPan("googledrive", &GoogleDrive{})
}

type GoogleDrive struct{}

func (g GoogleDrive) IsLogin(account *module.Account) bool {
	return true
}

func g_req() *resty.Request {
	req := resty.New().R()
	if module.GloablConfig.Proxy != "" {
		req = resty.New().SetProxy(module.GloablConfig.Proxy).R()
	}
	return req
}

func (g GoogleDrive) AuthLogin(account *module.Account) (string, error) {
	var auth module.GoogleDriveAuthInfo
	_, err := g_req().
		SetResult(&auth).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     account.User,
			"client_secret": account.Password,
			"refresh_token": account.RefreshToken,
			"redirect_uri":  account.RedirectUri,
			"grant_type":    "refresh_token",
			"access_type":   "offline",
		}).
		Post("https://www.googleapis.com/oauth2/v4/token")
	if err != nil {
		log.Error(err)
		return "", err
	}
	GoogleDrives[account.Id] = auth
	return account.RefreshToken, err
}

func (g GoogleDrive) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	gd := GoogleDrives[account.Id]
	pageToken := ""
	for {
		var filesResp GoogleFilesResp
		_, err := g_req().
			SetResult(&filesResp).
			SetAuthToken(gd.AccessToken).
			SetQueryParams(map[string]string{
				"pageSize":                  "100",
				"supportsAllDrives":         "true",
				"includeItemsFromAllDrives": "true",
				"fields":                    "nextPageToken, files(id,name,mimeType,parents,size,fileExtension,thumbnailLink,modifiedTime,createdTime,md5Checksum)",
				"q":                         "trashed = false and '" + fileId + "' in parents",
				"orderBy":                   "folder,modifiedTime asc,name",
				"pageToken":                 pageToken,
			}).
			Get("https://www.googleapis.com/drive/v3/files")
		if err != nil {
			return fileNodes, err
		}
		for _, f := range filesResp.Files {
			fn := g.ToFileNode(f)
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
		if filesResp.NextPageToken != "" {
			pageToken = filesResp.NextPageToken
		} else {
			break
		}
	}
	return fileNodes, nil
}

func (g GoogleDrive) File(account module.Account, fileId, path string) (module.FileNode, error) {
	gd := GoogleDrives[account.Id]
	item := Files{}
	fn := module.FileNode{}
	if fileId == "" {
		return fn, nil
	}
	_, err := g_req().
		SetResult(&item).
		SetAuthToken(gd.AccessToken).
		SetQueryParam("fields", "*").
		SetQueryParam("supportsAllDrives", "true").
		SetPathParam("fileId", fileId).
		Get("https://www.googleapis.com/drive/v3/files/{fileId}")
	if err != nil {
		log.Errorln(err)
		return fn, err
	}
	fn = g.ToFileNode(item)
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	fn.AccountId = account.Id
	return fn, nil
}

func (g GoogleDrive) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	gd := GoogleDrives[account.Id]
	for _, file := range files {
		t1 := time.Now()
		resp, err := g_req().
			SetAuthToken(gd.AccessToken).
			SetHeader("Content-Type", "application/json; charset=UTF-8").
			SetBody(base.KV{
				"name":    file.FileName,
				"parents": []string{parentFileId},
			}).
			SetQueryParam("supportsAllDrives", "true").
			Post("https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable")
		uploadUrl := resp.Header().Get("location")
		if uploadUrl == "" {
			return false, "File upload failed", err
		}
		bfs := onedrive.ReadBlock(16384000, file) //15.625MB
		for _, bf := range bfs {
			httpClient := util.GetClient(0)
			r, _ := http.NewRequest("PUT", uploadUrl, bytes.NewReader(bf.Content))
			r.Header.Add("Content-Length", strconv.FormatInt(file.FileSize, 10))
			r.Header.Add("Content-Range", bf.Name)
			res, _ := httpClient.Do(r)
			defer res.Body.Close()
		}
		log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "all files uploaded", nil
}

func (g GoogleDrive) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	gd := GoogleDrives[account.Id]
	resp, err := g_req().
		SetAuthToken(gd.AccessToken).
		SetPathParam("fileId", fileId).
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetBody(base.KV{"name": name}).
		SetQueryParam("supportsAllDrives", "true").
		Patch("https://www.googleapis.com/drive/v3/files/{fileId}")
	if err != nil {
		return false, "File rename failed", err
	}
	log.Debug("File rename: ", resp.String())
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		return true, resp.String(), err
	}
	return true, resp.String(), err
}

func (g GoogleDrive) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	gd := GoogleDrives[account.Id]
	resp, err := g_req().
		SetAuthToken(gd.AccessToken).
		SetPathParam("fileId", fileId).
		SetQueryParam("supportsAllDrives", "true").
		Delete("https://www.googleapis.com/drive/v3/files/{fileId}")
	if err != nil {
		return false, "File remove failed", err
	}
	log.Debug("File remove: ", resp.String())
	if resp.StatusCode() == http.StatusOK {
		return true, resp.String(), err
	}
	return true, resp.String(), err
}

func (g GoogleDrive) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	gd := GoogleDrives[account.Id]
	resp, err := g_req().
		SetAuthToken(gd.AccessToken).
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetBody(base.KV{
			"name":     name,
			"parents":  []string{parentFileId},
			"mimeType": "application/vnd.google-apps.folder",
		}).
		SetQueryParam("supportsAllDrives", "true").
		Post("https://www.googleapis.com/drive/v3/files")
	if err != nil {
		return false, "Dir create failed", err
	}
	log.Debug("Dir create: ", resp.String())
	dirId := jsoniter.Get(resp.Body(), "id")
	status := jsoniter.Get(resp.Body(), "success").ToBool()
	if status {
		return true, dirId, err
	}
	return true, dirId, err
}

func (g GoogleDrive) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	gd := GoogleDrives[account.Id]
	f, _ := g.File(account, fileId, "")
	resp, err := g_req().
		SetAuthToken(gd.AccessToken).
		SetPathParam("fileId", fileId).
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetQueryParam("addParents", targetFileId).
		SetQueryParam("removeParents", f.ParentId).
		SetQueryParam("supportsAllDrives", "true").
		Patch("https://www.googleapis.com/drive/v3/files/{fileId}")
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

func (g GoogleDrive) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	f, err := g.File(account, fileId, "")
	if err != nil {
		if f.IsFolder {
			//only copy file
			resp, er := g.CopyOneFile(account, fileId, targetFileId, overwrite)
			if er != nil {
				return false, "File copy failed", er
			}
			log.Debug("File copy: ", resp.String())
			status := jsoniter.Get(resp.Body(), "success").ToBool()
			if status {
				return true, resp.String(), er
			}
		} else {
			//copy file and folder
			_, target, _ := g.Mkdir(account, targetFileId, f.FileName)
			files, _ := g.Files(account, fileId, "/", "", "")
			for _, file := range files {
				resp, er := g.CopyOneFile(account, file.FileId, target.(string), true)
				log.Debug("File copy: ", resp.String())
				status := jsoniter.Get(resp.Body(), "success").ToBool()
				if status {
					return true, resp.String(), er
				}
			}
		}
	}
	return false, "File copy failed", err
}

func (g GoogleDrive) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	return "https://www.googleapis.com/drive/v3/files/" + fileId + "?alt=media", nil
}

func (g GoogleDrive) GetSpaceSzie(account module.Account) (int64, int64) {
	gd := GoogleDrives[account.Id]
	resp, err := g_req().
		SetAuthToken(gd.AccessToken).
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetQueryParam("fields", "storageQuota").
		SetQueryParam("supportsAllDrives", "true").
		Get("https://www.googleapis.com/drive/v3/about")
	if err != nil {
		return 0, 0
	}
	total := jsoniter.Get(resp.Body(), "storageQuota", "limit").ToInt64()
	used := jsoniter.Get(resp.Body(), "storageQuota", "usage").ToInt64()
	return total, used
}

func (g GoogleDrive) ToFileNode(f Files) module.FileNode {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = f.ID
	fn.FileName = f.Name
	fn.CreateTime = util.UTCTimeFormat(f.CreatedTime)
	fn.LastOpTime = util.UTCTimeFormat(f.ModifiedTime)
	if f.Parents != nil {
		fn.ParentId = f.Parents[0]
	}
	fn.IsDelete = 1
	if strings.Contains(f.MimeType, "folder") {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	} else {
		fn.IsFolder = false
		fn.FileType = util.GetExt(fn.FileName)
		fn.ViewType = util.GetViewType(fn.FileType)
		size, _ := strconv.ParseInt(f.Size, 10, 64)
		fn.FileSize = int64(size)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
		fn.Thumbnail = f.ThumbnailLink
		fn.DownloadUrl = f.WebContentLink
	}
	return fn
}

func (g GoogleDrive) CopyOneFile(account module.Account, fileId string, targetId string, overwrite bool) (*resty.Response, error) {
	gd := GoogleDrives[account.Id]
	resp, er := g_req().
		SetAuthToken(gd.AccessToken).
		SetPathParam("fileId", fileId).
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetBody(base.KV{
			"parents": []string{targetId},
		}).
		SetQueryParam("supportsAllDrives", "true").
		Post("https://www.googleapis.com/drive/v3/files/{fileId}/copy")
	return resp, er
}

type GoogleFilesResp struct {
	Files         []Files `json:"files"`
	NextPageToken string  `json:"nextPageToken"`
}

type Files struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	MimeType       string   `json:"mimeType"`
	Parents        []string `json:"parents"`
	CreatedTime    string   `json:"createdTime"`
	ModifiedTime   string   `json:"modifiedTime"`
	ThumbnailLink  string   `json:"thumbnailLink,omitempty"`
	FileExtension  string   `json:"fileExtension,omitempty"`
	Md5Checksum    string   `json:"md5Checksum,omitempty"`
	Size           string   `json:"size,omitempty"`
	WebContentLink string   `json:"webContentLink,omitempty"`
}
