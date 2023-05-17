package _123

import "time"

// LoginResp /**
type LoginResp struct {
	Code int `json:"code"`
	Data struct {
		RefreshTokenExpireTime int    `json:"refresh_token_expire_time"`
		LoginType              int    `json:"login_type"`
		Expire                 string `json:"expire"`
		Token                  string `json:"token"`
	} `json:"data"`
	Message string `json:"message"`
}

// file list
type FilesResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Next     string `json:"Next"`
		Len      int    `json:"Len"`
		IsFirst  bool   `json:"IsFirst"`
		InfoList []Item `json:"InfoList"`
	} `json:"data"`
}

// file item
type Item struct {
	FileId       int    `json:"FileId"`
	FileName     string `json:"FileName"`
	Type         int    `json:"Type"`
	Size         int    `json:"Size"`
	ContentType  string `json:"ContentType"`
	S3KeyFlag    string `json:"S3KeyFlag"`
	CreateAt     string `json:"CreateAt"`
	UpdateAt     string `json:"UpdateAt"`
	Hidden       bool   `json:"Hidden"`
	Etag         string `json:"Etag"`
	Status       int    `json:"Status"`
	ParentFileId int    `json:"ParentFileId"`
	Category     int    `json:"Category"`
	PunishFlag   int    `json:"PunishFlag"`
	ParentName   string `json:"ParentName"`
	DownloadUrl  string `json:"DownloadUrl"`
	EnableAppeal int    `json:"EnableAppeal"`
	ToolTip      string `json:"ToolTip"`
}

// file resp
type FileResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		InfoList []Item `json:"InfoList"`
	} `json:"data"`
}

// download resp
type DownloadResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessKeyID     interface{} `json:"AccessKeyId"`
		SecretAccessKey interface{} `json:"SecretAccessKey"`
		SessionToken    interface{} `json:"SessionToken"`
		Expiration      interface{} `json:"Expiration"`
		Key             string      `json:"Key"`
		Bucket          string      `json:"Bucket"`
		FileID          int         `json:"FileId"`
		Reuse           bool        `json:"Reuse"`
		Info            interface{} `json:"Info"`
		UploadID        string      `json:"UploadId"`
		DownloadURL     string      `json:"DownloadUrl"`
	} `json:"data"`
}

// user info
type UserInfoResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		UID                       int       `json:"UID"`
		Nickname                  string    `json:"Nickname"`
		SpaceUsed                 int64     `json:"SpaceUsed"`
		SpacePermanent            int64     `json:"SpacePermanent"`
		SpaceTemp                 int       `json:"SpaceTemp"`
		FileCount                 int       `json:"FileCount"`
		SpaceTempExpr             time.Time `json:"SpaceTempExpr"`
		Mail                      string    `json:"Mail"`
		Passport                  int64     `json:"Passport"`
		HeadImage                 string    `json:"HeadImage"`
		BindWechat                bool      `json:"BindWechat"`
		StraightLink              bool      `json:"StraightLink"`
		OpenLink                  int       `json:"OpenLink"`
		Vip                       bool      `json:"Vip"`
		VipExpire                 string    `json:"VipExpire"`
		SpaceBuy                  bool      `json:"SpaceBuy"`
		VipExplain                string    `json:"VipExplain"`
		SignType                  int       `json:"SignType"`
		ContinuousPayment         bool      `json:"ContinuousPayment"`
		ContinuousPaymentDate     string    `json:"ContinuousPaymentDate"`
		ContinuousPaymentAmount   int       `json:"ContinuousPaymentAmount"`
		ContinuousPaymentDuration int       `json:"ContinuousPaymentDuration"`
	} `json:"data"`
}

type TrashResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		InfoList []struct {
			FileID int `json:"FileId"`
		} `json:"InfoList"`
	} `json:"data"`
}

type CommonResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UploadRequestResp upload request response
type UploadRequestResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessKeyID     string      `json:"AccessKeyId"`
		SecretAccessKey string      `json:"SecretAccessKey"`
		SessionToken    string      `json:"SessionToken"`
		Expiration      time.Time   `json:"Expiration"`
		Key             string      `json:"Key"`
		Bucket          string      `json:"Bucket"`
		FileID          int         `json:"FileId"`
		Reuse           bool        `json:"Reuse"`
		Info            interface{} `json:"Info"`
		UploadID        string      `json:"UploadId"`
		DownloadURL     string      `json:"DownloadUrl"`
	} `json:"data"`
}

type RepareResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		PresignedUrls map[int]string `json:"presignedUrls"`
	} `json:"data"`
}

type ListUploadResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Parts []struct {
			PartNumber string `json:"PartNumber"`
			Size       string `json:"Size"`
			ETag       string `json:"ETag"`
		} `json:"Parts"`
	} `json:"data"`
}

type DownloadUrlResp struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    struct {
		RedirectURL string `json:"redirect_url"`
	} `json:"data"`
}
