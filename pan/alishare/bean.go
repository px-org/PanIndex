package _alishare

import "time"

type ShareTokenResp struct {
	ShareToken  string    `json:"share_token"`
	AccessToken string    `json:"access_token"`
	ExpireTime  time.Time `json:"expire_time"`
	ExpiresIn   int       `json:"expires_in"`
	//local use
	Signature      string `json:"signature"`
	Nonce          int    `json:"nonce"`
	PrivateKeyHex  string `json:"private_key_hex"`
	DeviceId       string `json:"device_id"`
	UserId         string `json:"user_id"`
	DefaultDriveId string `json:"default_drive_id"`
}

type FilesResp struct {
	Items      []Items `json:"items"`
	NextMarker string  `json:"next_marker"`
}
type Items struct {
	DriveID       string `json:"drive_id"`
	DomainID      string `json:"domain_id"`
	FileID        string `json:"file_id"`
	ShareID       string `json:"share_id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	FileExtension string `json:"file_extension"`
	MimeType      string `json:"mime_type"`
	MimeExtension string `json:"mime_extension"`
	Size          int    `json:"size"`
	ParentFileID  string `json:"parent_file_id"`
	Category      string `json:"category"`
	PunishFlag    int    `json:"punish_flag"`
	Thumbnail     string `json:"thumbnail,omitempty"`
}

type DownloadResp struct {
	DownloadURL string `json:"download_url"`
	URL         string `json:"url"`
	Thumbnail   string `json:"thumbnail"`
}
