package ali

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"net/http"
	"strings"
	"time"
)

func init() {
	base.RegisterPan("aliyundrive", &Ali{})
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

// auth login api return (refresh_token, err)
func (a Ali) AuthLogin(account *module.Account) (string, error) {
	var tokenResp TokenResp
	_, err := base.Client.R().
		SetResult(&tokenResp).
		SetBody(base.KV{"refresh_token": account.RefreshToken, "grant_type": "refresh_token"}).
		Post("https://auth.aliyundrive.com/v2/account/token")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	Alis[account.Id] = tokenResp
	signature, _ := a.CreateSession(*account)
	if signature == "" {
		return "", err
	}
	return tokenResp.RefreshToken, nil
}

func (a Ali) CreateSession(account module.Account) (string, error) {
	if SignCache.Has(account.Id) {
		signature, err := SignCache.Get(account.Id)
		log.Debugf("get signature from cache：%s", signature)
		return signature.(string), err
	}
	tokenResp := Alis[account.Id]
	privateKey, pubKey := genKeys()
	tokenResp.Nonce = NonceMin
	signature := genSignature(privateKey, tokenResp.DeviceId, tokenResp.UserId, tokenResp.Nonce)
	resp, err := base.Client.R().
		SetBody(base.KV{"deviceName": "Chrome浏览器", "modelName": "Windows网页版", "pubKey": pubKey}).
		SetAuthToken(tokenResp.AccessToken).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/users/v1/users/device/create_session")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	success := jsoniter.Get(resp.Body(), "success").ToBool()
	result := jsoniter.Get(resp.Body(), "result").ToBool()
	if result && success {
		SignCache.Set(account.Id, signature)
		log.Debugf("CreateSession success, signature：%s", signature)
	} else {
		log.Error(resp.String())
	}
	tokenResp.Signature = signature
	tokenResp.PrivateKeyHex = hex.EncodeToString(privateKey.Bytes())
	Alis[account.Id] = tokenResp
	return signature, nil
}

func (a Ali) RenewSession(account module.Account) (string, error) {
	tokenResp := Alis[account.Id]
	privateKey, _ := hex.DecodeString(tokenResp.PrivateKeyHex)
	nonce := getNextNonce(tokenResp.Nonce)
	signature := genSignature(privateKey, tokenResp.DeviceId, tokenResp.UserId, nonce)
	resp, err := base.Client.R().
		SetBody(base.KV{}).
		SetAuthToken(tokenResp.AccessToken).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/users/v1/users/device/renew_session")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	success := jsoniter.Get(resp.Body(), "success").ToBool()
	result := jsoniter.Get(resp.Body(), "result").ToBool()
	if result && success {
		log.Debugf("RenewSession success, signature：%s", signature)
	} else {
		return "", err
	}
	tokenResp.Signature = signature
	tokenResp.Nonce = nonce
	Alis[account.Id] = tokenResp
	return signature, nil
}

func genSignature(privKey secp256k1.PrivKey, deviceId, userId string, nonce int) string {
	str := "%s:%s:%s:%d"
	message := fmt.Sprintf(str, APPID, deviceId, userId, nonce)
	signature, _ := privKey.Sign([]byte(message))
	sign := hex.EncodeToString(signature) + "00"
	return sign
}

func genKeys() (secp256k1.PrivKey, string) {
	privateKey := secp256k1.GenPrivKey()
	return privateKey, hex.EncodeToString(privateKey.PubKey().Bytes())
}

func getNextNonce(nonce int) int {
	if nonce > NonceMax {
		return NonceMin
	}
	return nonce + 1
}

func (a Ali) GetSpaceSzie(account module.Account) (int64, int64) {
	tokenResp := Alis[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(tokenResp.AccessToken).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", tokenResp.Signature).
		Post("https://auth.aliyundrive.com/v2/databox/get_personal_info")
	if err != nil {
		log.Errorln(err)
		return 0, 0
	}
	totalSize := jsoniter.Get(resp.Body(), "personal_space_info").Get("total_size").ToInt64()
	usedSize := jsoniter.Get(resp.Body(), "personal_space_info").Get("used_size").ToInt64()
	return totalSize, usedSize
}

// files api return (entity.FileNode, err)
func (a Ali) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	tokenResp := Alis[account.Id]
	sortColumn, sortOrder = a.GetOrderFiled(sortColumn, sortOrder)
	limit := 100
	nextMarker := ""
	for {
		var fsResp AliFilesResp
		re, err := base.Client.R().
			SetResult(&fsResp).
			SetAuthToken(tokenResp.AccessToken).
			SetBody(base.KV{
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
			/*SetHeader("x-device-id", tokenResp.DeviceId).
			SetHeader("x-signature", tokenResp.Signature).*/
			Post("https://api.aliyundrive.com/adrive/v3/file/list")
		if err != nil {
			log.Errorln(err)
			return nil, err
		}
		nextMarker = fsResp.NextMarker
		if len(fsResp.Items) == 0 {
			code := jsoniter.Get(re.Body(), "code").ToString()
			if code == "ParamFlowException" {
				return nil, base.FlowLimit
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

// file api return (entity.FileNode, err)
func (a Ali) File(account module.Account, fileId, path string) (module.FileNode, error) {
	tokenResp := Alis[account.Id]
	item := Items{}
	fn := module.FileNode{}
	signature, _ := a.CreateSession(account)
	resp, err := base.Client.R().
		SetResult(&item).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"drive_id": tokenResp.DefaultDriveId,
			"file_id":  fileId,
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/v2/file/get")
	if err != nil {
		log.Errorln(err)
		return fn, err
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.File(account, fileId, path)
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
		signature, _ := a.CreateSession(account)
		var aliMkdirResp AliMkdirResp
		resp, err := base.Client.R().
			SetResult(&aliMkdirResp).
			SetAuthToken(tokenResp.AccessToken).
			SetBody(base.KV{
				"drive_id": tokenResp.DefaultDriveId,
				"part_info_list": []base.KV{base.KV{
					"part_number": 1,
				},
				},
				"parent_file_id":  parentFileId,
				"name":            file.FileName,
				"type":            "file",
				"check_name_mode": checkNameMode,
				"size":            file.FileSize,
			}).
			SetHeader("x-device-id", tokenResp.DeviceId).
			SetHeader("x-signature", signature).
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
		resp, e := base.Client.R().
			SetAuthToken(tokenResp.AccessToken).
			SetBody(base.KV{
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

// rename api return (ok, AliRenameResp, err)
func (a Ali) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	signature, _ := a.CreateSession(account)
	var result AliRenameResp
	resp, err := base.Client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"drive_id":        tokenResp.DefaultDriveId,
			"file_id":         fileId,
			"name":            name,
			"check_name_mode": "refuse",
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/v3/file/update")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.Rename(account, fileId, name)
	}
	return true, result, err
}

// remove api return (ok, AliRemoveResp, err)
func (a Ali) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	var result AliRemoveResp
	signature, _ := a.CreateSession(account)
	resp, err := base.Client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"drive_id": tokenResp.DefaultDriveId,
			"file_id":  fileId,
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/v2/recyclebin/trash")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.Remove(account, fileId)
	}
	return true, result, err
}

// mkdir api return (ok, AliMkdirResp, err)
func (a Ali) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	signature, _ := a.CreateSession(account)
	var result AliMkdirResp
	resp, err := base.Client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"drive_id":        tokenResp.DefaultDriveId,
			"parent_file_id":  parentFileId,
			"name":            name,
			"check_name_mode": "refuse",
			"type":            "folder",
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.Mkdir(account, parentFileId, name)
	}
	return true, result, err
}

// move api return (ok, BatchApiResp, err)
func (a Ali) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	tokenResp := Alis[account.Id]
	signature, _ := a.CreateSession(account)
	var result BatchApiResp
	resp, err := base.Client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"requests": []base.KV{
				{
					"body": base.KV{"drive_id": tokenResp.DefaultDriveId,
						"file_id":           fileId,
						"to_drive_id":       tokenResp.DefaultDriveId,
						"to_parent_file_id": targetFileId},
					"headers": base.KV{
						"Content-Type": "application/json",
					},
					"id":     fileId,
					"method": "POST",
					"url":    "/file/move",
				},
			},
			"resource": "file",
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/v3/batch")
	if err != nil {
		log.Errorln(err)
		return false, result, err
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.Move(account, fileId, targetFileId, overwrite)
	}
	return true, result, err
}

// copy api return (ok, BatchApiResp, err)
func (a Ali) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	return false, nil, base.ErrNotImplement
}

func (a Ali) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	tokenResp := Alis[account.Id]
	signature, _ := a.CreateSession(account)
	var result AliDownResp
	resp, err := base.Client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"drive_id":   tokenResp.UserInfo.DefaultDriveId,
			"file_id":    fileId,
			"expire_sec": 14400,
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/v2/file/get_download_url")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.GetDownloadUrl(account, fileId)
	}
	if result.Ratelimit.PartSpeed != 0 {
		log.Warningf("该文件限速：%d", result.Ratelimit.PartSpeed)
	}
	return result.URL, err
}

// Get Paths by fileId
func (a Ali) GetPaths(account module.Account, fileId string) ([]module.FileNode, error) {
	fns := make([]module.FileNode, 0)
	signature, _ := a.CreateSession(account)
	tokenResp := Alis[account.Id]
	var items []Items
	resp, err := base.Client.R().
		SetResult(&items).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"drive_id": tokenResp.UserInfo.DefaultDriveId,
			"file_id":  fileId,
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/adrive/v1/file/get_path")
	if err != nil {
		log.Errorln(err)
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.GetPaths(account, fileId)
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
	signature, _ := a.CreateSession(account)
	resp, err := base.Client.R().
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"category":    "live_transcoding",
			"drive_id":    tokenResp.DefaultDriveId,
			"file_id":     fileId,
			"template_id": "",
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
		Post("https://api.aliyundrive.com/v2/file/get_video_preview_play_info")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	if strings.Contains(resp.String(), "DeviceSessionSignatureInvalid") {
		SignCache.Remove(account.Id)
		log.Debugf("signature expired create and retry:%s", signature)
		return a.Transcode(account, fileId)
	}
	return resp.String(), err
}

// path api return (ok, AliPathResp, err)
func (a Ali) GetPath(account module.Account, fileId string) (AliPathResp, error) {
	tokenResp := Alis[account.Id]
	signature, _ := a.CreateSession(account)
	var result AliPathResp
	_, err := base.Client.R().
		SetResult(&result).
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{
			"drive_id": tokenResp.DefaultDriveId,
			"file_id":  fileId,
		}).
		SetHeader("x-device-id", tokenResp.DeviceId).
		SetHeader("x-signature", signature).
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
		_, err := base.Client.R().
			SetResult(&result).
			SetAuthToken(tokenResp.AccessToken).
			SetBody(base.KV{
				"drive_id":                tokenResp.DefaultDriveId,
				"limit":                   100,
				"query":                   "name match \"" + key + "\"",
				"image_thumbnail_process": "image/resize,w_200/format,jpeg",
				"image_url_process":       "image/resize,w_1920/format,jpeg",
				"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
				"order_by":                "updated_at DESC",
				"marker":                  marker,
			}).
			SetHeader("x-device-id", tokenResp.DeviceId).
			SetHeader("x-signature", tokenResp.Signature).
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

// ali qrcode gen
func QrcodeGen() (string, string) {
	resp, err := base.Client.R().
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

// ali qrcode check
func QrcodeCheck(t, codeContent, ck, resultCode string) (string, string) {
	resp, _ := base.Client.R().
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

// ali sign activity
func (a Ali) SignActivity(account module.Account) {
	tokenResp := Alis[account.Id]
	_, err := base.Client.R().
		SetAuthToken(tokenResp.AccessToken).
		SetBody(base.KV{"refresh_token": account.RefreshToken, "grant_type": "refresh_token"}).
		Post("https://member.aliyundrive.com/v1/activity/sign_in_list")
	if err != nil {
		log.Errorln(err)
	}
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
