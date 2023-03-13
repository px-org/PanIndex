package pikpak

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

func init() {
	base.RegisterPan("pikpak", &Pikpak{})
}

type Pikpak struct {
}

var Pikpaks = map[string]module.PikpakToken{}

func (p Pikpak) AuthLogin(account *module.Account) (string, error) {
	var tokenResp module.PikpakToken
	if account.RefreshToken != "" {
		refreshToken, accessToken := p.AuthToken(account)
		if refreshToken != "" || accessToken != "" {
			tokenResp.AccessToken = accessToken
			tokenResp.RefreshToken = refreshToken
			Pikpaks[account.Id] = tokenResp
			return refreshToken, nil
		}
	}
	_, err := base.Client.R().
		SetResult(&tokenResp).
		SetBody(
			base.KV{
				"username":  account.User,
				"password":  account.Password,
				"client_id": "YNxT9w7GMdWvEOKa",
			},
		).
		Post("https://user.mypikpak.com/v1/auth/signin")
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	Pikpaks[account.Id] = tokenResp
	return tokenResp.RefreshToken, nil
}

func (p Pikpak) AuthToken(account *module.Account) (string, string) {
	var e RespErr
	resp, err := base.Client.R().
		SetError(e).
		SetBody(base.KV{
			"client_id":     "YNxT9w7GMdWvEOKa",
			"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
			"grant_type":    "refresh_token",
			"refresh_token": account.RefreshToken,
		}).Post("https://user.mypikpak.com/v1/auth/token")
	if err != nil {
		log.Errorf("auth token: %v", err)
	}
	if e.ErrorCode != 0 {
		if e.ErrorCode == 4126 {
			return "", ""
		}
	}
	refreshToken := jsoniter.Get(resp.Body(), "refresh_token").ToString()
	accessToken := jsoniter.Get(resp.Body(), "access_token").ToString()
	return refreshToken, accessToken
}

type RespErr struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

func (p Pikpak) IsLogin(account *module.Account) bool {
	var e RespErr
	resp, err := base.Client.R().
		SetError(e).Get("https://user.mypikpak.com/v1/user/me")
	if err != nil {
		return false
	}
	sub := jsoniter.Get(resp.Body(), "sub").ToString()
	if sub != "" {
		return true
	}
	return false
}

func (p Pikpak) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	pikPakToken := Pikpaks[account.Id]
	nextPageToken := ""
	for {
		var filesResp PikpakFilesResp
		_, err := base.Client.R().
			SetAuthToken(pikPakToken.AccessToken).
			SetResult(&filesResp).
			SetQueryParams(map[string]string{
				"thumbnail_size": "SIZE_MEDIUM",
				"limit":          "100",
				"parent_id":      fileId,
				"with_audit":     "true",
				"page_token":     nextPageToken,
				"filters":        `{"phase":{"eq":"PHASE_TYPE_COMPLETE"},"trashed":{"eq":false}}`,
			}).
			Get("https://api-drive.mypikpak.com/drive/v1/files")
		if err != nil {
			log.Errorln(err)
			return nil, err
		}
		nextPageToken = filesResp.NextPageToken
		for _, f := range filesResp.Files {
			fn, _ := p.ToFileNode(f)
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
		if filesResp.NextPageToken == "" {
			break
		}
	}
	return fileNodes, nil
}

func (p Pikpak) File(account module.Account, fileId, path string) (module.FileNode, error) {
	fn := module.FileNode{}
	pikPakToken := Pikpaks[account.Id]
	if fileId == "" {
		return module.FileNode{
			FileId:     "",
			FileName:   "",
			FileSize:   0,
			IsFolder:   true,
			Path:       "/",
			LastOpTime: time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	}
	var file PikpakFile
	_, err := base.Client.R().
		SetAuthToken(pikPakToken.AccessToken).
		SetResult(&file).
		Get("https://api-drive.mypikpak.com/drive/v1/files/" + fileId)
	if err != nil {
		log.Errorln(err)
		return fn, err
	}
	fn, _ = p.ToFileNode(file)
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	fn.AccountId = account.Id
	return fn, err
}

func (a Pikpak) ToFileNode(item PikpakFile) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = item.ID
	fn.FileName = item.Name
	fn.CreateTime = util.UTCTimeFormat(item.CreatedTime)
	fn.LastOpTime = util.UTCTimeFormat(item.ModifiedTime)
	fn.ParentId = item.ParentID
	fn.IsDelete = 1
	if item.Kind == "drive#file" {
		fn.IsFolder = false
		fn.FileType = strings.ToLower(item.FileExtension[1:len(item.FileExtension)])
		fn.ViewType = util.GetViewType(fn.FileType)
		size, _ := strconv.ParseInt(item.Size, 10, 64)
		fn.FileSize = size
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
		fn.Thumbnail = item.ThumbnailLink
		if len(item.Medias) > 0 && item.Medias[0].Link.URL != "" {
			fn.DownloadUrl = item.Medias[0].Link.URL
		} else {
			fn.DownloadUrl = item.WebContentLink
		}
	} else {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	}
	return fn, nil
}

func (p Pikpak) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	pikPakToken := Pikpaks[account.Id]
	for _, file := range files {
		t1 := time.Now()
		sha1 := sha1.Sum(file.Content)
		resp, err := base.Client.R().
			SetAuthToken(pikPakToken.AccessToken).
			SetResult(&file).
			SetBody(base.KV{
				"kind":        "drive#file",
				"name":        file.FileName,
				"size":        file.FileSize,
				"hash":        strings.ToUpper(fmt.Sprintf("%x", sha1)),
				"upload_type": "UPLOAD_TYPE_RESUMABLE",
				"objProvider": base.KV{"provider": "UPLOAD_TYPE_UNKNOWN"},
				"parent_id":   parentFileId}).
			Post("https://api-drive.mypikpak.com/drive/v1/files")
		params := jsoniter.Get(resp.Body(), "resumable").Get("params")
		endpoint := params.Get("endpoint").ToString()
		endpointS := strings.Split(endpoint, ".")
		endpoint = strings.Join(endpointS[1:], ".")
		accessKeyId := params.Get("access_key_id").ToString()
		accessKeySecret := params.Get("access_key_secret").ToString()
		securityToken := params.Get("security_token").ToString()
		key := params.Get("key").ToString()
		bucket := params.Get("bucket").ToString()
		cfg := &aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, securityToken),
			Region:      aws.String("pikpak"),
			Endpoint:    &endpoint,
		}
		s, err := session.NewSession(cfg)
		if err != nil {
			log.Errorf("Unable to create session:%s, err:%v", file.FileName, err)
		}
		uploader := s3manager.NewUploader(s)
		input := &s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(file.Content),
		}
		_, err = uploader.Upload(input)
		if err != nil {
			log.Errorf("Unable to upload file:%s, err:%v", file.FileName, err)
		}
		log.Debugf("file:%s，upload success,timespan:%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "all files uploaded", nil
}

func (p Pikpak) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	pikPakToken := Pikpaks[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(pikPakToken.AccessToken).
		SetBody(base.KV{"name": name}).
		Patch("https://api-drive.mypikpak.com/drive/v1/files/" + fileId)
	if err != nil {
		log.Errorf("File renaming failed: %v", err)
		return false, resp.String(), err
	}
	log.Debug("File renamed successfully")
	return true, resp.String(), err
}

func (p Pikpak) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	pikPakToken := Pikpaks[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(pikPakToken.AccessToken).
		SetBody(base.KV{"ids": []string{fileId}}).
		Post("https://api-drive.mypikpak.com/drive/v1/files:batchTrash")
	log.Debug("File remove success: ", resp.String())
	if err != nil {
		log.Errorf("Error occurred while waiting for object %s to be deleted, %v", fileId, err)
		return false, "File remove error", err
	}
	return true, resp.String(), err
}

func (p Pikpak) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	pikPakToken := Pikpaks[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(pikPakToken.AccessToken).
		SetBody(base.KV{"kind": "drive#folder", "parent_id": parentFileId, "name": name}).
		Post("https://api-drive.mypikpak.com/drive/v1/files")
	log.Debug("Dir create: ", resp.String())
	if err == nil {
		return true, "Dir create success", nil
	}
	return false, "Dir create error", err
}

func (p Pikpak) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	pikPakToken := Pikpaks[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(pikPakToken.AccessToken).
		SetBody(base.KV{"ids": []string{fileId}, "to": base.KV{"parent_id": targetFileId}}).
		Post("https://api-drive.mypikpak.com/drive/v1/files:batchMove")
	if err != nil {
		log.Errorf("File moved failed: %v", err)
		return false, resp.String(), err
	}
	log.Debug("File moved successfully")
	return true, resp.String(), err
}

func (p Pikpak) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	pikPakToken := Pikpaks[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(pikPakToken.AccessToken).
		SetBody(base.KV{"ids": []string{fileId}, "to": base.KV{"parent_id": targetFileId}}).
		Post(" https://api-drive.mypikpak.com/drive/v1/files:batchCopy")
	if err != nil {
		log.Errorf("File copy failed: %v", err)
		return false, resp.String(), err
	}
	log.Debug("File copy successfully")
	return true, resp.String(), err
}

func (p Pikpak) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	fn, err := p.File(account, fileId, "")
	if err != nil {
		return "", err
	}
	return fn.DownloadUrl, err
}

func (p Pikpak) GetSpaceSzie(account module.Account) (int64, int64) {
	pikPakToken := Pikpaks[account.Id]
	resp, err := base.Client.R().
		SetAuthToken(pikPakToken.AccessToken).
		Get("https://api-drive.mypikpak.com/drive/v1/about")
	if err != nil {
		return 0, 0
	}
	totalSizeStr := jsoniter.Get(resp.Body(), "quota").Get("limit").ToString()
	usedSizeStr := jsoniter.Get(resp.Body(), "quota").Get("usage").ToString()
	totalSize, _ := strconv.ParseInt(totalSizeStr, 10, 64)
	usedSize, _ := strconv.ParseInt(usedSizeStr, 10, 64)
	return totalSize, usedSize
}

type PikpakFilesResp struct {
	NextPageToken string       `json:"next_page_token"`
	Files         []PikpakFile `json:"files"`
}

type PikpakFile struct {
	Kind              string        `json:"kind"`      //文件类型
	ID                string        `json:"id"`        //文件ID
	ParentID          string        `json:"parent_id"` //父目录ID
	Name              string        `json:"name"`
	UserID            string        `json:"user_id"`
	Size              string        `json:"size"`
	Revision          string        `json:"revision"`
	FileExtension     string        `json:"file_extension"`
	MimeType          string        `json:"mime_type"`
	Starred           bool          `json:"starred"`
	WebContentLink    string        `json:"web_content_link"`
	CreatedTime       string        `json:"created_time"`
	ModifiedTime      string        `json:"modified_time"`
	IconLink          string        `json:"icon_link"`
	ThumbnailLink     string        `json:"thumbnail_link"`
	Md5Checksum       string        `json:"md5_checksum"`
	Hash              string        `json:"hash"`
	Phase             string        `json:"phase"`
	Audit             interface{}   `json:"audit"`
	Medias            Medias        `json:"medias"`
	Trashed           bool          `json:"trashed"`
	DeleteTime        string        `json:"delete_time"`
	OriginalURL       string        `json:"original_url"`
	OriginalFileIndex int           `json:"original_file_index"`
	Space             string        `json:"space"`
	Apps              []interface{} `json:"apps"`
	Writable          bool          `json:"writable"`
	FolderType        string        `json:"folder_type"`
	Collection        interface{}   `json:"collection"`
}
type Medias []struct {
	VipTypes       []interface{} `json:"vip_types"`
	IconLink       string        `json:"icon_link"`
	IsOrigin       bool          `json:"is_origin"`
	Category       string        `json:"category"`
	MediaID        string        `json:"media_id"`
	NeedMoreQuota  bool          `json:"need_more_quota"`
	RedirectLink   string        `json:"redirect_link"`
	ResolutionName string        `json:"resolution_name"`
	IsVisible      bool          `json:"is_visible"`
	MediaName      string        `json:"media_name"`
	Video          Video         `json:"video"`
	Link           Link          `json:"link"`
	IsDefault      bool          `json:"is_default"`
	Priority       int           `json:"priority"`
}
type Video struct {
	FrameRate  int    `json:"frame_rate"`
	VideoCodec string `json:"video_codec"`
	AudioCodec string `json:"audio_codec"`
	VideoType  string `json:"video_type"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Duration   int    `json:"duration"`
	BitRate    int    `json:"bit_rate"`
}
type Link struct {
	URL    string    `json:"url"`
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}
