package teambition

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
	"net/http"
	"strconv"
	"strings"
	"time"
)

var TeambitionSessions = map[string]module.Teambition{}
var t_zones = map[string]module.Zone{
	"teambition": module.Zone{
		Login:  "https://account.teambition.com",
		Api:    "https://www.teambition.com",
		Upload: "https://tcs.teambition.net/upload",
		Desc:   "国内版",
	},
	"teambition-us": module.Zone{
		Login:  "https://us-account.teambition.com",
		Api:    "https://us.teambition.com",
		Upload: "https://us-tcs.teambition.net/upload",
		Desc:   "国外版",
	},
}

func init() {
	base.RegisterPan("teambition", &Teambition{})
	base.RegisterPan("teambition-us", &Teambition{})
}

type Teambition struct{}

func (t Teambition) IsLogin(account *module.Account) bool {
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		Get(t_zones[account.Mode].Login + "/api/account")
	if err != nil {
		return false
	}
	if resp.StatusCode() == 401 {
		return false
	} else if resp.StatusCode() == 200 {
		id := jsoniter.Get(resp.Body(), "_id").ToString()
		if id != "" {
			return true
		}
	}
	return false
}

func (t Teambition) AuthLogin(account *module.Account) (string, error) {
	req := resty.New().R()
	teambition := module.Teambition{}
	auth := ""
	resp, err := req.Get(t_zones[account.Mode].Login + "/login/password")
	if err != nil {
		return "", err
	}
	token := util.GetBetweenStr(resp.String(), "TOKEN\":\"", "\"")
	clientId := util.GetBetweenStr(resp.String(), "CLIENT_ID\":\"", "\"")
	params := base.KV{
		"client_id":     clientId,
		"token":         token,
		"password":      account.Password,
		"response_type": "session"}
	if strings.Contains(account.User, "@") {
		params["email"] = account.User
		resp, err = req.SetBody(params).
			Post(t_zones[account.Mode].Login + "/api/login/email")
	} else {
		params["phone"] = account.User
		resp, err = req.SetBody(params).
			Post(t_zones[account.Mode].Login + "/api/login/phone")
	}
	u := jsoniter.Get(resp.Body(), "user")
	if u == nil || u.Get("_id").ToString() == "" {
		//login failed
		log.Error("[Teambition] login failed: ", resp.String())
		return "", err
	} else {
		auth = jsoniter.Get(resp.Body(), "token").ToString()
	}
	//2. get orgId, memberId
	resp, err = req.Get(t_zones[account.Mode].Api + "/api/projects/" + account.RootId)
	if err != nil {
		return "", err
	}
	auth = req.Header.Get("Cookie")
	//teambition.GloablRootId = jsoniter.Get(resp.Body(), "_rootCollectionId").ToString()
	teambition.GloablProjectId = account.SiteId
	teambition.GloablRootId = account.RootId
	TeambitionSessions[account.Id] = teambition
	return auth, err
}

func (t Teambition) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := []module.FileNode{}
	teambition := TeambitionSessions[account.Id]
	//req := teambition.Req
	limit := 100
	pageNum := 1
	if fileId == teambition.GloablProjectId {
		fileId = teambition.GloablRootId
	}
	for {
		//1. read dir
		var dirsResp []TeambitionDirResp
		resty.New().R().
			SetResult(&dirsResp).
			SetHeader("Cookie", account.RefreshToken).
			SetQueryParams(map[string]string{
				"_projectId": teambition.GloablProjectId,
				"_parentId":  fileId,
				"order":      "updatedAsc",
				"count":      strconv.Itoa(limit),
				"page":       strconv.Itoa(pageNum),
				"uuid":       uuid.NewV4().String(),
			}).Get(t_zones[account.Mode].Api + "/api/collections")
		for _, f := range dirsResp {
			if f.CollectionType == "default" {
				continue
			}
			fn := t.ToDirNode(f)
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
		//2. read files
		var filesResp []TeambitionFileResp
		resty.New().R().
			SetResult(&filesResp).
			SetHeader("Cookie", account.RefreshToken).
			SetQueryParams(map[string]string{
				"_projectId": teambition.GloablProjectId,
				"_parentId":  fileId,
				"order":      "updatedAsc",
				"count":      strconv.Itoa(limit),
				"page":       strconv.Itoa(pageNum),
			}).Get(t_zones[account.Mode].Api + "/api/works")
		for _, f := range filesResp {
			fn := t.ToFileNode(f)
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
		if len(filesResp) == 0 && len(dirsResp) == 0 {
			break
		}
		pageNum++
	}
	return fileNodes, nil
}

func (t Teambition) File(account module.Account, fileId, path string) (module.FileNode, error) {
	fn := module.FileNode{}
	var dirResp TeambitionDirResp
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		SetResult(&dirResp).
		Get(t_zones[account.Mode].Api + "/api/collections/" + fileId)
	if resp.StatusCode() == http.StatusNotFound {
		var fileResp TeambitionFileResp
		resp, err = resty.New().R().
			SetHeader("Cookie", account.RefreshToken).
			SetResult(&fileResp).
			Get(t_zones[account.Mode].Api + "/api/works/" + fileId)
		if resp.StatusCode() != http.StatusNotFound {
			fn = t.ToFileNode(fileResp)
			fn.AccountId = account.Id
			fn.Path = path
			fn.ParentPath = util.GetParentPath(path)
		}
	} else {
		fn = t.ToDirNode(dirResp)
		fn.AccountId = account.Id
		fn.Path = path
		fn.ParentPath = util.GetParentPath(path)
	}
	return fn, err
}

func (t Teambition) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	for _, file := range files {
		if file.FileSize <= 20971520 {
			t.UploadFile(account, parentFileId, file)
		} else {
			t.UploadChunkFile(account, parentFileId, file)
		}
	}
	return true, "all files uploaded", nil
}

func (t Teambition) UploadChunkFile(account module.Account, parentFileId string, file *module.UploadInfo) (bool, error) {
	t1 := time.Now()
	log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		Get(t_zones[account.Mode].Api + "/projects")
	jwt := util.GetBetweenStr(resp.String(), "strikerAuth&quot;:&quot;", "&quot;,&quot;phoneForLogin")
	var teambitionUploadResp TeambitionUploadResp
	resp, err = resty.New().R().
		SetResult(&teambitionUploadResp).
		SetHeader("Authorization", jwt).
		SetBody(base.KV{
			"fileName":    file.FileName,
			"fileSize":    file.FileSize,
			"lastUpdated": time.Now(),
		}).Post(t_zones[account.Mode].Upload + "/chunk")
	if resp.StatusCode() != 200 {
		log.Error(err)
		return false, err
	}
	//chunks := jsoniter.Get(resp.Body(), "chunks").ToInt()
	chunkSize := jsoniter.Get(resp.Body(), "chunkSize").ToInt()
	bfs := util.ChunkBytes(file.Content, chunkSize)
	for i, bf := range bfs {
		uploadUrl := fmt.Sprintf(t_zones[account.Mode].Upload+"/chunk/"+teambitionUploadResp.FileKey+"?chunk=%d&chunks=%d", i+1, len(bfs))
		r, _ := http.NewRequest("POST", uploadUrl, bytes.NewReader(bf))
		r.Header.Add("Content-Length", strconv.FormatInt(file.FileSize, 10))
		r.Header.Add("Authorization", jwt)
		r.Header.Add("Content-Length", fmt.Sprintf("%d", len(bf)))
		r.Header.Add("Referer", t_zones[account.Mode].Api)
		r.Header.Add("Content-Type", "application/octet-stream")
		res, er := http.DefaultClient.Do(r)
		defer res.Body.Close()
		if er != nil {
			log.Error(er)
			return false, err
		}
	}
	resp, err = resty.New().R().SetHeader("Authorization", jwt).Post(t_zones[account.Mode].Upload + "/chunk/" + teambitionUploadResp.FileKey)
	_, err = t.UploadWorks(account, parentFileId, teambitionUploadResp)
	if err != nil {
		log.Errorln(err)
		return false, err
	}
	log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	return true, nil
}

func (t Teambition) UploadFile(account module.Account, parentFileId string, file *module.UploadInfo) (bool, error) {
	t1 := time.Now()
	log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		Get(t_zones[account.Mode].Api + "/projects")
	jwt := util.GetBetweenStr(resp.String(), "strikerAuth&quot;:&quot;", "&quot;,&quot;phoneForLogin")
	reader := bytes.NewReader(file.Content)
	var teambitionUploadResp TeambitionUploadResp
	resp, err = resty.New().R().
		SetResult(&teambitionUploadResp).
		SetHeader("Authorization", jwt).
		SetMultipartField("file", file.FileName, file.ContentType, reader).
		SetMultipartFormData(map[string]string{
			"name": file.FileName,
			"type": file.ContentType,
			"size": strconv.FormatInt(file.FileSize, 10),
		}).Post(t_zones[account.Mode].Upload)
	if resp.StatusCode() != 200 {
		log.Error(err)
		return false, err
	}
	if err != nil {
		log.Error(err)
		return false, err
	} else {
		_, err = t.UploadWorks(account, parentFileId, teambitionUploadResp)
		if err != nil {
			log.Errorln(err)
			return false, err
		}
		log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
		return true, nil
	}
}

func (t Teambition) UploadWorks(account module.Account, parentFileId string, teambitionUploadResp TeambitionUploadResp) (string, error) {
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		SetBody(base.KV{
			"works": []base.KV{{
				"fileKey":      teambitionUploadResp.FileKey,
				"fileName":     teambitionUploadResp.FileName,
				"fileType":     teambitionUploadResp.FileType,
				"fileSize":     teambitionUploadResp.FileSize,
				"fileCategory": teambitionUploadResp.FileCategory,
				"source":       "tcs",
				"visible":      "members",
				"_parentId":    parentFileId,
			}},
			"_parentId": parentFileId,
		}).
		Post(t_zones[account.Mode].Api + "/api/works")
	log.Debug("[Teambition] upload finish:", resp.String())
	return resp.String(), err
}

func (t Teambition) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	var err error
	var resp *resty.Response
	f, err := t.File(account, fileId, "/")
	if err == nil {
		if f.IsFolder {
			resp, err = resty.New().R().
				SetHeader("Cookie", account.RefreshToken).
				SetBody(base.KV{
					"title": name,
				}).
				Put(t_zones[account.Mode].Api + "/api/collections/" + fileId)
		} else {
			resp, err = resty.New().R().
				SetHeader("Cookie", account.RefreshToken).
				SetBody(base.KV{
					"fileName": name,
				}).
				Put(t_zones[account.Mode].Api + "/api/works/" + fileId)
		}
		log.Debug("File rename: ", resp.String())
		if resp.StatusCode() == http.StatusOK {
			return true, "File rename success", err
		}
	}
	return false, "File rename error", err
}

func (t Teambition) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	updated := time.Now().UTC().String()
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		SetBody(base.KV{
			"_id":     fileId,
			"updated": updated,
		}).
		Delete(t_zones[account.Mode].Api + "/api/works/" + fileId)
	resp, err = resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		SetBody(base.KV{
			"_id":     fileId,
			"updated": updated,
		}).
		Delete(t_zones[account.Mode].Api + "/api/collections/" + fileId)
	log.Debug("File remove: ", resp.String())
	if resp.StatusCode() == http.StatusOK {
		return true, "File remove success", nil
	}
	return false, "File remove error", err
}

func (t Teambition) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		SetBody(base.KV{
			"collectionType": "",
			"color":          "blue",
			"created":        "",
			"description":    "",
			"objectType":     "collection",
			"recentWorks":    []base.KV{},
			"title":          name,
			"updated":        "",
			"workCount":      0,
			"_creatorId":     "",
			"_parentId":      parentFileId,
			"_projectId":     account.RootId,
		}).
		Post(t_zones[account.Mode].Api + "/api/collections")
	log.Debug("Dir create: ", resp.String())
	if resp.StatusCode() == http.StatusOK {
		return true, "Dir create success", nil
	}
	return false, "Dir create error", err
}

func (t Teambition) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	var err error
	var resp *resty.Response
	f, err := t.File(account, fileId, "/")
	if err == nil {
		if f.IsFolder {
			resp, err = resty.New().R().
				SetHeader("Cookie", account.RefreshToken).
				SetBody(base.KV{
					"_parentId": targetFileId,
				}).
				Put(t_zones[account.Mode].Api + "/api/collections/" + fileId + "/move")
		} else {
			resp, err = resty.New().R().
				SetHeader("Cookie", account.RefreshToken).
				SetBody(base.KV{
					"_parentId": targetFileId,
				}).
				Put(t_zones[account.Mode].Api + "/api/works/" + fileId + "/move")
		}
		log.Debug("File move: ", resp.String())
		if resp.StatusCode() == http.StatusOK {
			return true, "File move success", err
		}
	}
	return false, "File move error", err
}

func (t Teambition) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	var err error
	var resp *resty.Response
	f, err := t.File(account, fileId, "/")
	if err == nil {
		if f.IsFolder {
			resp, err = resty.New().R().
				SetHeader("Cookie", account.RefreshToken).
				SetBody(base.KV{
					"_parentId":  targetFileId,
					"_projectId": account.RootId,
				}).
				Put(t_zones[account.Mode].Api + "/api/collections/" + fileId + "/fork")
		} else {
			resp, err = resty.New().R().
				SetHeader("Cookie", account.RefreshToken).
				SetBody(base.KV{
					"_parentId": targetFileId,
				}).
				Put(t_zones[account.Mode].Api + "/api/works/" + fileId + "/fork")
		}
		log.Debug("File copy: ", resp.String())
		if resp.StatusCode() == http.StatusOK {
			return true, "File copy success", err
		}

	}
	return false, "File copy error", err
}

func (t Teambition) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	if fileId == "" {
		return "", nil
	}
	fn := module.FileNode{}
	var fileResp TeambitionFileResp
	resp, err := resty.New().R().
		SetHeader("Cookie", account.RefreshToken).
		SetResult(&fileResp).
		Get(t_zones[account.Mode].Api + "/api/works/" + fileId)
	if resp.StatusCode() != http.StatusNotFound {
		fn = t.ToFileNode(fileResp)
		rs, _ := resty.New().SetRedirectPolicy(resty.FlexibleRedirectPolicy(1)).R().Get(fn.DownloadUrl)
		return rs.Header().Get("Location"), nil
	}
	return "", err
}

func (t Teambition) GetSpaceSzie(account module.Account) (int64, int64) {
	//no limit
	return 0, 0
}

func (t Teambition) ToDirNode(f TeambitionDirResp) module.FileNode {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = f.ID
	fn.FileName = f.Title
	fn.CreateTime = util.UTCTimeFormat(f.Created)
	fn.LastOpTime = util.UTCTimeFormat(f.Updated)
	fn.IsDelete = 1
	fn.IsFolder = true
	fn.FileType = ""
	fn.IsFolder = true
	fn.FileSize = 0
	fn.SizeFmt = "-"
	fn.ParentId = f.ParentID
	fn.FileId = f.ID
	return fn
}

func (t Teambition) ToFileNode(f TeambitionFileResp) module.FileNode {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = f.ID
	fn.FileName = f.FileName
	fn.CreateTime = util.UTCTimeFormat(f.Created)
	fn.LastOpTime = util.UTCTimeFormat(f.Updated)
	fn.IsDelete = 1
	fn.IsFolder = false
	fn.FileType = util.GetExt(fn.FileName)
	fn.ViewType = util.GetViewType(fn.FileType)
	fn.FileSize = int64(f.FileSize)
	fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	fn.Thumbnail = f.ThumbnailURL
	fn.DownloadUrl = f.DownloadURL
	fn.ParentId = f.ParentID
	fn.FileId = f.ID
	return fn
}

type TeambitionDirResp struct {
	ID                      string        `json:"_id"`
	ParentID                string        `json:"_parentId"`
	CollectionType          string        `json:"collectionType"`
	CreatorID               string        `json:"_creatorId"`
	ProjectID               string        `json:"_projectId"`
	OrganizationID          interface{}   `json:"_organizationId"`
	Description             string        `json:"description"`
	Title                   string        `json:"title"`
	Pinyin                  string        `json:"pinyin"`
	Py                      string        `json:"py"`
	Updated                 string        `json:"updated"`
	Created                 string        `json:"created"`
	IsArchived              bool          `json:"isArchived"`
	WorkCount               int           `json:"workCount"`
	CollectionCount         int           `json:"collectionCount"`
	Color                   string        `json:"color"`
	ObjectType              string        `json:"objectType"`
	Visible                 string        `json:"visible"`
	InvolveMembers          []interface{} `json:"involveMembers"`
	IsConfigurable          bool          `json:"isConfigurable"`
	LockedConfigurabilityBy string        `json:"lockedConfigurabilityBy"`
	Creator                 Creator       `json:"creator"`
	Involvers               []interface{} `json:"involvers"`
}

type TeambitionFileResp struct {
	ID                   string        `json:"_id"`
	FileName             string        `json:"fileName"`
	Pinyin               string        `json:"pinyin"`
	Py                   string        `json:"py"`
	FileType             string        `json:"fileType"`
	FileSize             int           `json:"fileSize"`
	FileKey              string        `json:"fileKey"`
	FileCategory         string        `json:"fileCategory"`
	ImageWidth           interface{}   `json:"imageWidth"`
	ImageHeight          interface{}   `json:"imageHeight"`
	ParentID             string        `json:"_parentId"`
	ProjectID            string        `json:"_projectId"`
	OrganizationID       interface{}   `json:"_organizationId"`
	CreatorID            string        `json:"_creatorId"`
	TagIds               []interface{} `json:"tagIds"`
	Visible              string        `json:"visible"`
	DownloadURL          string        `json:"downloadUrl"`
	ThumbnailURL         string        `json:"thumbnailUrl"`
	Thumbnail            string        `json:"thumbnail"`
	Description          string        `json:"description"`
	Source               string        `json:"source"`
	InvolveMembers       []string      `json:"involveMembers"`
	Created              string        `json:"created"`
	Updated              string        `json:"updated"`
	LastVersionTime      time.Time     `json:"lastVersionTime"`
	LastUploaderID       string        `json:"lastUploaderId"`
	IsArchived           bool          `json:"isArchived"`
	ObjectType           string        `json:"objectType"`
	RawData              RawData       `json:"rawData"`
	PreviewURL           interface{}   `json:"previewUrl"`
	Creator              Creator       `json:"creator"`
	CommentsCount        int           `json:"commentsCount"`
	AttachmentsCount     int           `json:"attachmentsCount"`
	ParentVisible        string        `json:"parentVisible"`
	ParentInvolveMembers []interface{} `json:"parentInvolveMembers"`
	ImmPreviewURL        string        `json:"immPreviewUrl,omitempty"`
}

type RawData struct {
}

type Creator struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type TeambitionUploadResp struct {
	FileKey       string `json:"fileKey"`
	FileName      string `json:"fileName"`
	FileType      string `json:"fileType"`
	FileCategory  string `json:"fileCategory"`
	MimeType      string `json:"mimeType"`
	FileSize      int    `json:"fileSize"`
	ImageWidth    int    `json:"imageWidth"`
	ImageHeight   int    `json:"imageHeight"`
	Source        string `json:"source"`
	DownloadURL   string `json:"downloadUrl"`
	ThumbnailURL  string `json:"thumbnailUrl"`
	PreviewURL    string `json:"previewUrl"`
	ImmPreviewURL string `json:"immPreviewUrl"`
}
