package _alishare

import (
	"encoding/hex"
	"fmt"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/ali"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"net/http"
	"strings"
)

func init() {
	base.RegisterPan("aliyundrive-share", &AliShare{})
}

type AliShare struct{}

func (a AliShare) AuthLogin(account *module.Account) (string, error) {
	var tokenResp ali.TokenResp
	var shareTokenResp ShareTokenResp
	_, err := base.Client.R().
		SetResult(&tokenResp).
		SetBody(base.KV{"refresh_token": account.RefreshToken, "grant_type": "refresh_token"}).
		Post("https://auth.aliyundrive.com/v2/account/token")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	shareTokenResp.AccessToken = tokenResp.AccessToken
	shareTokenResp.DeviceId = tokenResp.DeviceId
	shareTokenResp.UserId = tokenResp.UserId
	shareTokenResp.DefaultDriveId = tokenResp.DefaultDriveId
	_, err = base.Client.R().
		SetResult(&shareTokenResp).
		SetBody(base.KV{"share_id": account.SiteId, "share_pwd": account.Password}).
		Post("https://api.aliyundrive.com/v2/share_link/get_share_token")
	if err != nil {
		log.Error(err)
		return "", err
	}
	Sessions[account.Id] = shareTokenResp
	return tokenResp.RefreshToken, nil
}

func (a AliShare) IsLogin(account *module.Account) bool {
	//TODO implement me
	panic("implement me")
}

func (a AliShare) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	var fsResp FilesResp
	fileNodes := make([]module.FileNode, 0)
	limit := 20
	nextMarker := ""
	for {
		body, err := a.request(&account, "https://api.aliyundrive.com/adrive/v2/file/list_by_share", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.KV{
				"limit":                   limit,
				"order_by":                "name",
				"order_direction":         "DESC",
				"parent_file_id":          fileId,
				"share_id":                account.SiteId,
				"image_thumbnail_process": "image/resize,w_400/format,jpeg",
				"image_url_process":       "image/resize,w_1920/format,jpeg",
				"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
				"marker":                  nextMarker,
			})
		}, &fsResp)
		if err != nil {
			log.Errorln(err)
			return fileNodes, err
		}
		nextMarker = fsResp.NextMarker
		if len(fsResp.Items) == 0 {
			code := jsoniter.Get(body, "code").ToString()
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

func (a AliShare) ToFileNode(item Items) (module.FileNode, error) {
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

func (a AliShare) File(account module.Account, fileId, path string) (module.FileNode, error) {
	var item Items
	fn := module.FileNode{}
	_, err := a.request(&account, "https://api.aliyundrive.com/adrive/v2/file/get_by_share", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.KV{
			"fields":                  "*",
			"file_id":                 fileId,
			"image_thumbnail_process": "image/resize,w_400/format,jpeg",
			"image_url_process":       "image/resize,w_375/format,jpeg",
			"share_id":                account.SiteId,
			"video_thumbnail_process": "video/snapshot,t_1000,f_jpg,ar_auto,w_375",
		})
	}, &item)
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

func (a AliShare) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (a AliShare) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (a AliShare) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (a AliShare) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (a AliShare) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (a AliShare) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (a AliShare) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	accessToken := Sessions[account.Id].AccessToken
	var resp DownloadResp
	_, err := a.request(&account, "https://api.aliyundrive.com/v2/file/get_share_link_download_url", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.KV{
			//"drive_id":        "337954",
			"expire_sec":      600,
			"file_id":         fileId,
			"get_streams_url": true,
			"share_id":        account.SiteId,
		}).SetAuthToken(accessToken)
	}, &resp)
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	return resp.DownloadURL, err
}

func (a AliShare) GetSpaceSzie(account module.Account) (int64, int64) {
	return 0, 0
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

func (a AliShare) CreateSession(account module.Account) (string, error) {
	if SignCache.Has(account.Id) {
		signature, err := SignCache.Get(account.Id)
		log.Debugf("get signature from cache：%s", signature)
		return signature.(string), err
	}
	tokenResp := Sessions[account.Id]
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
	Sessions[account.Id] = tokenResp
	return signature, nil
}

// transcode api return (ok, string, err)
func (a AliShare) Transcode(account module.Account, fileId string) (string, error) {
	tokenResp := Sessions[account.Id]
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
