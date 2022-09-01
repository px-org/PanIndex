package pan

import (
	"bytes"
	"encoding/base64"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"net/http"
	"strings"
	"time"
)

var Alis = map[string]module.TokenResp{}

func init() {
	RegisterPan("aliyundrive", &Ali{})
}

type Ali struct{}

func (a Ali) IsLogin(account *module.Account) bool {
	total, _ := a.GetSpaceSzie(*account)
	if total > 0 {
		return true
	} else {
		return false
	}
}

//auth login api return (refresh_token, err)
func (a Ali) AuthLogin(account *module.Account) (string, error) {
	var tokenResp module.TokenResp
	_, err := client.R().
		SetResult(&tokenResp).
		SetBody(KV{"refresh_token": account.RefreshToken, "grant_type": "refresh_token"}).
		Post("https://auth.aliyundrive.com/v2/account/token")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	Alis[account.Id] = tokenResp
	return tokenResp.RefreshToken, nil
}

func (a Ali) GetSpaceSzie(account module.Account) (int64, int64) {
	tokenResp := Alis[account.Id]
	resp, err := client.R().
		SetAuthToken(tokenResp.AccessToken).
		Post("https://auth.aliyundrive.com/v2/databox/get_personal_info")
	if err != nil {
		log.Errorln(err)
		return 0, 0
	}
	totalSize := jsoniter.Get(resp.Body(), "personal_space_info").Get("total_size").ToInt64()
	usedSize := jsoniter.Get(resp.Body(), "personal_space_info").Get("used_size").ToInt64()
	return totalSize, usedSize
}

//files api return (entity.FileNode, err)
func (a Ali) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	tokenResp := Alis[account.Id]
	sortColumn, sortOrder = a.GetOrderFiled(sortColumn, sortOrder)
	limit := 100
	nextMarker := ""
	for {
		var fsResp AliFilesResp
		re, err := client.R().
			SetResult(&fsResp).
			SetAuthToken(tokenResp.AccessToken).
			SetBody(KV{
				"drive_id":                tokenResp.DefaultDriveId,
				"parent_file_id":          fileId,
				"all":                     false,
				"limit":                   limit,
				"url_expire_sec":          1600,
				"image_thumbnail_process": "image/resize,w_400/format,jpeg",
				"image_url_process":       "image/resize,w_1920/format,jpeg",
				"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
				"fields":                  "*",
				"order_by":                sortColumn, //default updated_at
				"order_direction":         sortOrder,  //default DESC
				"marker":                  nextMarker,
			}).
			Post("https://api.aliyundrive.com/adrive/v3/file/list")
		if err != nil {
			log.Errorln(err)
			return nil, err
		}
		nextMarker = fsResp.NextMarker
		if len(fsResp.Items) == 0 {
			code := jsoniter.Get(re.Body(), "code").ToString()
			if code == "ParamFlowException" {
				return nil, FlowLimit
			}
		}
		for _, f := range fsResp.Items {
			fn, _ := a.ToFileNode(f)
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
		if fsResp.NextMarker == "" {
			break
		}
	}
	return fileNodes, nil
}

func (a Ali) ToFileNode(item Items) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = item.FileID
	fn.FileName = item.Name
	fn.CreateTime = util.UTCTimeFormat(item.CreatedAt)
	fn.LastOpTime = util.UTCTimeFormat(item.UpdatedAt)
	fn.ParentId = item.ParentFileID
	fn.IsDelete = 1
	kind := item.Type
	if kind == "file" {
		fn.IsFolder = false
		fn.FileType = strings.ToLower(item.FileExtension)
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = int64(item.Size)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
		fn.Thumbnail = item.Thumbnail
	} else {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	}
	return fn, nil
}

//file api return (entity.FileNode, err)
func (a Ali) File(account module.Account, fileId, path string) (module.FileNode, error) {
	tokenResp := Alis[account.Id]
	item := Items{}
	fn := module.FileNode{}
	_, err := client.R().
		SetResult(&item).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"drive_id": tokenResp.DefaultDriveId,
			"file_id":  fileId,
		}).
		Post("https://api.aliyundrive.com/v2/file/get")
	if err != nil {
		log.Errorln(err)
		return fn, err
	}
	fn, _ = a.ToFileNode(item)
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	fn.AccountId = account.Id
	return fn, nil
}

// upload api return (ok, result, error)
func (a Ali) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	checkNameMode := "auto_rename"
	if overwrite == true {
		checkNameMode = "overwrite"
	}
	for _, file := range files {
		t1 := time.Now()
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		var aliMkdirResp AliMkdirResp
		resp, err := client.R().
			SetResult(&aliMkdirResp).
			SetAuthToken(tokenResp.AccessToken).
			SetBody(KV{
				"drive_id": tokenResp.DefaultDriveId,
				"part_info_list": []KV{KV{
					"part_number": 1,
				},
				},
				"parent_file_id":  parentFileId,
				"name":            file.FileName,
				"type":            "file",
				"check_name_mode": checkNameMode,
				"size":            file.FileSize,
			}).
			Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
		if err != nil {
			log.Errorln(err)
			return false, aliMkdirResp, err
		}
		if aliMkdirResp.RapidUpload {
			log.Debugf("File：%s，fast upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
			return true, aliMkdirResp, nil
		}
		log.Debugf("File Parts Count：%d", len(aliMkdirResp.PartInfoList))
		for _, partInfo := range aliMkdirResp.PartInfoList {
			byteContent := file.Content
			client := &http.Client{}
			req, er := http.NewRequest(http.MethodPut, partInfo.UploadURL, bytes.NewBuffer(byteContent))
			if er != nil {
				log.Error(er)
				return false, "File part upload failed", er
			}
			client.Do(req)
		}
		resp, e := client.R().
			SetAuthToken(tokenResp.AccessToken).
			SetBody(KV{
				"drive_id":  tokenResp.DefaultDriveId,
				"file_id":   aliMkdirResp.FileID,
				"upload_id": aliMkdirResp.UploadID,
			}).
			Post("https://api.aliyundrive.com/v2/file/complete")
		if e != nil {
			log.Errorln(e)
			return false, resp.String(), e
		}
		file.FileId = aliMkdirResp.FileID
		log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "All files uploaded", nil
}

//rename api return (ok, AliRenameResp, err)
func (a Ali) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	var result AliRenameResp
	_, err := client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"drive_id":        tokenResp.DefaultDriveId,
			"file_id":         fileId,
			"name":            name,
			"check_name_mode": "refuse",
		}).
		Post("https://api.aliyundrive.com/v3/file/update")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	return true, result, err
}

// remove api return (ok, AliRemoveResp, err)
func (a Ali) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	var result AliRemoveResp
	_, err := client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"drive_id": tokenResp.DefaultDriveId,
			"file_id":  fileId,
		}).
		Post("https://api.aliyundrive.com/v2/recyclebin/trash")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	return true, result, err
}

// mkdir api return (ok, AliMkdirResp, err)
func (a Ali) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	var result AliMkdirResp
	_, err := client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"drive_id":        tokenResp.DefaultDriveId,
			"parent_file_id":  parentFileId,
			"name":            name,
			"check_name_mode": "refuse",
			"type":            "folder",
		}).
		Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	return true, result, err
}

// move api return (ok, BatchApiResp, err)
func (a Ali) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	var result BatchApiResp
	_, err := client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"requests": []KV{
				{
					"body": KV{"drive_id": tokenResp.DefaultDriveId,
						"file_id":           fileId,
						"to_drive_id":       tokenResp.DefaultDriveId,
						"to_parent_file_id": targetFileId},
					"headers": KV{
						"Content-Type": "application/json",
					},
					"id":     fileId,
					"method": "POST",
					"url":    "/file/move",
				},
			},
			"resource": "file",
		}).
		Post("https://api.aliyundrive.com/v3/batch")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	return true, result, err
}

// copy api return (ok, BatchApiResp, err)
func (a Ali) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	return false, nil, ErrNotImplement
}

func (a Ali) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	tokenResp := Alis[account.Id]
	var result AliDownResp
	_, err := client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"drive_id":   tokenResp.UserInfo.DefaultDriveId,
			"file_id":    fileId,
			"expire_sec": 14400,
		}).
		Post("https://api.aliyundrive.com/v2/file/get_download_url")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	if result.Ratelimit.PartSpeed != -1 {
		log.Warningf("该文件限速：%d", result.Ratelimit.PartSpeed)
	}
	return result.URL, err
}

//Get Paths by fileId
func (a Ali) GetPaths(account module.Account, fileId string) ([]module.FileNode, error) {
	fns := make([]module.FileNode, 0)
	tokenResp := Alis[account.Id]
	var items []Items
	_, err := client.R().
		SetResult(&items).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"drive_id": tokenResp.UserInfo.DefaultDriveId,
			"file_id":  fileId,
		}).
		Post("https://api.aliyundrive.com/adrive/v1/file/get_path")
	if err != nil {
		log.Errorln(err)
	}
	for _, f := range items {
		fn, _ := a.ToFileNode(f)
		fns = append(fns, fn)
	}
	return fns, err
}

// transcode api return (ok, string, err)
func (a Ali) Transcode(account module.Account, fileId string) (string, error) {
	tokenResp := Alis[account.Id]
	resp, err := client.R().
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"category":    "live_transcoding",
			"drive_id":    tokenResp.DefaultDriveId,
			"file_id":     fileId,
			"template_id": "",
		}).
		Post("https://api.aliyundrive.com/v2/file/get_video_preview_play_info")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	return resp.String(), err
}

// path api return (ok, AliPathResp, err)
func (a Ali) GetPath(account module.Account, fileId string) (AliPathResp, error) {
	tokenResp := Alis[account.Id]
	var result AliPathResp
	_, err := client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(KV{
			"drive_id": tokenResp.DefaultDriveId,
			"file_id":  fileId,
		}).
		Post("https://api.aliyundrive.com/adrive/v1/file/get_path")
	if err != nil {
		log.Errorln(err)
		return result, err
	}
	return result, err
}

// search api return (items, err)
func (a Ali) Search(account module.Account, key string) ([]Items, error) {
	items := []Items{}
	tokenResp := Alis[account.Id]
	marker := ""
	for {
		var result AliFilesResp
		_, err := client.R().
			SetResult(&result).
			SetAuthToken(tokenResp.AccessToken).
			SetBody(KV{
				"drive_id":                tokenResp.DefaultDriveId,
				"limit":                   100,
				"query":                   "name match \"" + key + "\"",
				"image_thumbnail_process": "image/resize,w_200/format,jpeg",
				"image_url_process":       "image/resize,w_1920/format,jpeg",
				"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
				"order_by":                "updated_at DESC",
				"marker":                  marker,
			}).
			Post("https://api.aliyundrive.com/adrive/v3/file/search")
		if err != nil {
			log.Errorln(err)
			return items, err
		}
		for _, item := range result.Items {
			items = append(items, item)
		}
		if result.NextMarker != "" {
			marker = result.NextMarker
		} else {
			break
		}
	}
	return items, nil

}

//ali qrcode gen
func QrcodeGen() (string, string) {
	resp, err := client.R().
		Get("https://passport.aliyundrive.com/newlogin/qrcode/generate.do?appName=aliyun_drive")
	if err != nil {
		log.Errorln(err)
		return "", ""
	}
	data := jsoniter.Get(resp.Body(), "content").Get("data")
	codeUrl := data.Get("codeContent").ToString()
	png, err := qrcode.Encode(codeUrl, qrcode.Medium, 256)
	if err != nil {
		panic(err)
	}
	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(png))
	return dataURI, data.ToString()
}

//ali qrcode check
func QrcodeCheck(t, codeContent, ck, resultCode string) (string, string) {
	resp, _ := client.R().
		SetQueryParam("t", t).
		SetQueryParam("ck", ck).
		SetQueryParam("appName", "aliyun_drive").
		Post("https://passport.aliyundrive.com/newlogin/qrcode/query.do?appName=aliyun_drive")
	//NEW 新建二维码 SCANED 已扫码 CONFIRMED确认登录
	data := jsoniter.Get(resp.Body(), "content").Get("data")
	qrCodeStatus := data.Get("qrCodeStatus").ToString()
	refreshToken := ""
	if qrCodeStatus == "CONFIRMED" {
		bizExt := data.Get("bizExt").ToString()
		loginResult, _ := util.Base64Decode(bizExt)
		refreshToken = jsoniter.Get([]byte(loginResult), "pds_login_result").Get("refreshToken").ToString()
	}
	return qrCodeStatus, refreshToken
}

func (a Ali) GetOrderFiled(sortColumn, sortOrder string) (string, string) {
	if sortColumn == "default" {
		sortColumn = "updated_at"
	} else if sortColumn == "file_name" {
		sortColumn = "name"
	} else if sortColumn == "file_size" {
		sortColumn = "size"
	} else if sortColumn == "last_op_time" {
		sortColumn = "updated_at"
	}
	if sortOrder == "null" {
		sortOrder = "asc"
	}
	sortOrder = strings.ToUpper(sortOrder)
	return sortColumn, sortOrder
}

//file api response
type AliFilesResp struct {
	Items             []Items `json:"items"`
	NextMarker        string  `json:"next_marker"`
	PunishedFileCount int     `json:"punished_file_count"`
}

//file api file
type Items struct {
	DriveID         string `json:"drive_id"`
	FileID          string `json:"file_id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	Hidden          bool   `json:"hidden"`
	Status          string `json:"status"`
	ParentFileID    string `json:"parent_file_id"`
	FileExtension   string `json:"file_extension,omitempty"`
	MimeType        string `json:"mime_type,omitempty"`
	Size            int    `json:"size,omitempty"`
	ContentHash     string `json:"content_hash,omitempty"`
	ContentHashName string `json:"content_hash_name,omitempty"`
	Category        string `json:"category,omitempty"`
	Thumbnail       string `json:"thumbnail,omitempty"`
}

//remove api response
type AliRemoveResp struct {
	DomainID    string `json:"domain_id"`
	DriveID     string `json:"drive_id"`
	FileID      string `json:"file_id"`
	AsyncTaskID string `json:"async_task_id"`
}

//mkdir api response
type AliMkdirResp struct {
	UploadID     string      `json:"upload_id"`
	ParentFileID string      `json:"parent_file_id"`
	Type         string      `json:"type"`
	FileID       string      `json:"file_id"`
	DomainID     string      `json:"domain_id"`
	DriveID      string      `json:"drive_id"`
	FileName     string      `json:"file_name"`
	EncryptMode  string      `json:"encrypt_mode"`
	RapidUpload  bool        `json:"rapid_upload"`
	PartInfoList []*PartInfo `json:"part_info_list"`
}

// batch api(/file/move) response
type BatchApiResp struct {
	Responses []Responses `json:"responses"`
}

type Body struct {
	DomainID string `json:"domain_id"`
	DriveID  string `json:"drive_id"`
	FileID   string `json:"file_id"`
}

type Responses struct {
	Body   Body   `json:"body"`
	ID     string `json:"id"`
	Status int    `json:"status"`
}

// rename api response
type AliRenameResp struct {
	DriveID          string    `json:"drive_id"`
	DomainID         string    `json:"domain_id"`
	FileID           string    `json:"file_id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Hidden           bool      `json:"hidden"`
	Starred          bool      `json:"starred"`
	Status           string    `json:"status"`
	UserMeta         string    `json:"user_meta"`
	ParentFileID     string    `json:"parent_file_id"`
	EncryptMode      string    `json:"encrypt_mode"`
	CreatorType      string    `json:"creator_type"`
	CreatorID        string    `json:"creator_id"`
	CreatorName      string    `json:"creator_name"`
	LastModifierType string    `json:"last_modifier_type"`
	LastModifierID   string    `json:"last_modifier_id"`
	LastModifierName string    `json:"last_modifier_name"`
	RevisionID       string    `json:"revision_id"`
	Trashed          bool      `json:"trashed"`
}

type CreateFileWithProofResp struct {
	UploadID     string      `json:"upload_id"`
	FileID       string      `json:"file_id"`
	RapidUpload  bool        `json:"rapid_upload"`
	PartInfoList []*PartInfo `json:"part_info_list"`
}

type PartInfo struct {
	PartNumber int    `json:"part_number"`
	UploadURL  string `json:"upload_url"`
}

//path api response
type AliPathResp struct {
	Items []Items `json:"items"`
}

// Ali down response
type AliDownResp struct {
	Method          string    `json:"method"`
	URL             string    `json:"url"`
	InternalURL     string    `json:"internal_url"`
	Expiration      time.Time `json:"expiration"`
	Size            int       `json:"size"`
	Ratelimit       Ratelimit `json:"ratelimit"`
	Crc64Hash       string    `json:"crc64_hash"`
	ContentHash     string    `json:"content_hash"`
	ContentHashName string    `json:"content_hash_name"`
}
type Ratelimit struct {
	PartSpeed int `json:"part_speed"`
	PartSize  int `json:"part_size"`
}
