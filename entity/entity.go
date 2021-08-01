package entity

import (
	"github.com/libsgh/nic"
	"time"
)

type FileNode struct {
	Id           string `json:"id"`
	AccountId    string `json:"account_id"`
	FileId       string `json:"fileId"`
	FileIdDigest string `json:"fileIdDigest"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	SizeFmt      string `json:"sizeFmt"`
	FileType     string `json:"fileType"`
	IsFolder     bool   `json:"isFolder"`
	IsStarred    bool   `json:"isStarred"`
	LastOpTime   string `json:"lastOpTime"`
	ParentId     string `json:"parentId"`
	Path         string `json:"path"`
	ParentPath   string `json:"parentPath"`
	DownloadUrl  string `json:"downloadUrl"`
	MediaType    int    `json:"mediaType"` //1图片，2音频，3视频，4文本文档，0其他类型
	LargeUrl     string `json:"largeUrl"`
	SmallUrl     string `json:"smallUrl"`
	CreateTime   string `json:"create_time"`
	Delete       int    `json:"delete"`
	Hide         int    `json:"hide"`
	CacheTime    int64  `json:"cache_time"`
}
type SearchNode struct {
	FileNode
	Dx        string
	AccountId string
}
type Paths struct {
	FileId    string `json:"fileId"`
	FileName  string `json:"fileName"`
	IsCoShare int    `json:"isCoShare"`
}
type Config struct {
	Host              string    `json:"host" gorm:"default:'0.0.0.0'"`
	Port              int       `json:"port" gorm:"default:5238"`
	Accounts          []Account `json:"accounts" gorm:"-"`
	PwdDirId          string    `json:"pwd_dir_id"`
	HideFileId        string    `json:"hide_file_id"`
	HerokuAppUrl      string    `json:"heroku_app_url"`
	ApiToken          string    `json:"api_token"`
	SiteName          string    `json:"site_name"`
	AccountChoose     string    `json:"account_choose"`
	Theme             string    `json:"theme" gorm:"default:'mdui'"`
	AdminPassword     string    `json:"admin_password" gorm:"default:'PanIndex'"`
	Damagou           Damagou   `json:"damagou" gorm:"-"`
	OnlyReferrer      string    `json:"only_referrer"`
	RefreshCookie     string    `json:"refresh_cookie" gorm:"default:'0 0 8 1/1 * ?'"`
	UpdateFolderCache string    `json:"update_folder_cache"`
	HerokuKeepAlive   string    `json:"heroku_keep_alive"`
	FaviconUrl        string    `json:"favicon_url"` //网站图标
	Footer            string    `json:"footer"`      //网站底部信息
}
type Account struct {
	Id           string `json:"id"`            //网盘空间id
	Name         string `json:"name"`          //网盘空间名称
	Mode         string `json:"mode"`          //网盘模式，native（本地模式），cloud189(默认，天翼云网盘)，teambition（阿里teambition网盘），aliyundrive
	User         string `json:"user"`          //网盘账号用户名，邮箱或手机号
	Password     string `json:"password"`      //网盘账号密码
	RefreshToken string `json:"refresh_token"` //刷新token
	AccessToken  string `json:"access_token"`  //授权token
	RedirectUri  string `json:"redirect_uri"`  //重定向url（onedrive）
	RootId       string `json:"root_id"`       //目录id
	Default      int    `json:"default"`       //是否默认
	FilesCount   int    `json:"files_count"`   //文件总数
	Status       int    `json:"status"`        //状态：-1，缓存中 1，未缓存，2缓存成功，3缓存失败
	CookieStatus int    `json:"cookie_status"` //cookie状态：-1刷新中， 1未刷新，2正常，3失效
	TimeSpan     string `json:"time_span"`
	LastOpTime   string `json:"last_op_time"` //最近一次更新时间
	SyncDir      string `json:"sync_dir"`     //定时缓存指定目录
	SyncChild    int    `json:"sync_child"`   //是否缓存指定目录的子目录
}
type Damagou struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Teambition struct {
	TeambitionSession nic.Session
	GloablOrgId       string
	GloablDriveId     string
	GloablSpaceId     string
	GloablRootId      string
	GloablProjectId   string
	IsPorject         bool
}
type Ali struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type TokenResp struct {
	RespError
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`

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
type PartInfo struct {
	PartNumber int    `json:"partNumber"`
	UploadUrl  string `json:"uploadUrl"`
}
type AliPartInfo struct {
	PartNumber        int    `json:"part_number"`
	UploadUrl         string `json:"upload_url"`
	InternalUploadUrl string `json:"internal_upload_url"`
	ContentType       string `json:"content_type"`
}
type OneDriveAuthInfo struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
