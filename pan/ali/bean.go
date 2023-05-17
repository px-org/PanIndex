package ali

import "time"

type Sessions struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenResp struct {
	RespError
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	//local use
	Signature     string `json:"signature"`
	Nonce         int    `json:"nonce"`
	PrivateKeyHex string `json:"private_key_hex"`

	UserInfo

	DefaultSboxDriveId string        `json:"default_sbox_drive_id"`
	ExpireTime         *time.Time    `json:"expire_time"`
	State              string        `json:"state"`
	ExistLink          []interface{} `json:"exist_link"`
	NeedLink           bool          `json:"need_link"`
	PinSetup           bool          `json:"pin_setup"`
	IsFirstLogin       bool          `json:"is_first_login"`
	NeedRpVerify       bool          `json:"need_rp_verify"`
	DeviceId           string        `json:"device_id"`
}

type UserInfo struct {
	RespError
	DomainId       string                 `json:"domain_id"`
	UserId         string                 `json:"user_id"`
	Avatar         string                 `json:"avatar"`
	CreatedAt      int64                  `json:"created_at"`
	UpdatedAt      int64                  `json:"updated_at"`
	Email          string                 `json:"email"`
	NickName       string                 `json:"nick_name"`
	Phone          string                 `json:"phone"`
	Role           string                 `json:"role"`
	Status         string                 `json:"status"`
	UserName       string                 `json:"user_name"`
	Description    string                 `json:"description"`
	DefaultDriveId string                 `json:"default_drive_id"`
	UserData       map[string]interface{} `json:"user_data"`
}

type RespError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// file api response
type AliFilesResp struct {
	Items             []Items `json:"items"`
	NextMarker        string  `json:"next_marker"`
	PunishedFileCount int     `json:"punished_file_count"`
}

// file api file
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

// remove api response
type AliRemoveResp struct {
	DomainID    string `json:"domain_id"`
	DriveID     string `json:"drive_id"`
	FileID      string `json:"file_id"`
	AsyncTaskID string `json:"async_task_id"`
}

// mkdir api response
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

// path api response
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
