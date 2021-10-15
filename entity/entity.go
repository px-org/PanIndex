package entity

import (
	"github.com/libsgh/nic"
	"time"
)

type FileNode struct {
	Id           string `json:"id"`                              //数据库唯一主键
	AccountId    string `json:"account_id" gorm:"index:idx_aid"` //文件所属账号
	FileId       string `json:"fileId" gorm:"index:idx_fid`      //网盘中的文件id
	FileIdDigest string `json:"fileIdDigest"`
	FileName     string `json:"fileName" gorm:"index:idx_fn"`   //文件名称
	FileSize     int64  `json:"fileSize"`                       //文件大小
	SizeFmt      string `json:"sizeFmt"`                        //文件大小（格式化）
	FileType     string `json:"fileType"`                       //文件类型
	IsFolder     bool   `json:"isFolder"`                       //是否是目录
	IsStarred    bool   `json:"isStarred"`                      //是否收藏
	LastOpTime   string `json:"lastOpTime"`                     //最近一次操作时间
	ParentId     string `json:"parentId"`                       //父目录id
	Path         string `json:"path" gorm:"index:idx_p"`        //文件路径
	ParentPath   string `json:"parentPath" gorm:"index:idx_pp"` //文件上层目录
	DownloadUrl  string `json:"downloadUrl"`                    //下载地址
	MediaType    int    `json:"mediaType"`                      //1图片，2音频，3视频，4文本文档，0其他类型
	LargeUrl     string `json:"largeUrl"`                       //大图预览
	SmallUrl     string `json:"smallUrl"`                       //小图预览
	CreateTime   string `json:"create_time"`                    //创建时间（目录信息入库时间）
	Delete       int    `json:"delete" gorm:"index:idx_del"`    //删除标记（便于做缓存）
	Hide         int    `json:"hide"`                           //是否隐藏
	HasPwd       int    `json:"has_pwd"`                        //是否是密码文件（包含）
	CacheTime    int64  `json:"cache_time"`                     //缓存时间
}
type ShareInfo struct {
	AccountId string `json:"account_id"` //文件所属账号
	FilePath  string `json:"file_path"`  //PanIndex文件路径
	ShortCode string `json:"short_code"` //短链接code
}
type SearchNode struct {
	FileNode
	AccountId string
}
type Paths struct {
	FileId    string `json:"fileId"`
	FileName  string `json:"fileName"`
	IsCoShare int    `json:"isCoShare"`
}
type Config struct {
	Host              string    `json:"host" gorm:"default:'0.0.0.0'"`
	Port              string    `json:"port" gorm:"default:5238"`
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
	EnableSafetyLink  string    `json:"enable_safety_link"`
	IsNullReferrer    string    `json:"is_null_referrer"`
	RefreshCookie     string    `json:"refresh_cookie" gorm:"default:'0 0 8 1/1 * ?'"`
	UpdateFolderCache string    `json:"update_folder_cache"`
	HerokuKeepAlive   string    `json:"heroku_keep_alive"`
	FaviconUrl        string    `json:"favicon_url"`    //网站图标
	Footer            string    `json:"footer"`         //网站底部信息
	Css               string    `json:"css"`            //自定义css
	Js                string    `json:"js"`             //自定义js
	EnablePreview     string    `json:"enable_preview"` //是否开启文件预览
	Image             string    `json:"image"`          //图片
	Audio             string    `json:"audio"`          //音频
	Video             string    `json:"video"`          //视频
	Code              string    `json:"code"`           //代码
	Doc               string    `json:"doc"`            //文档
	Other             string    `json:"other"`          //other
	SColumn           string    `json:"s_column"`       //排序字段
	SOrder            string    `json:"s_order"`        //排序顺序
}
type ConfigItem struct {
	K string `json:"k" gorm:"unique;not null"` //配置项key
	V string `json:"v"`                        //配置项值
	G string `json:"g"`                        //配置项分组
}
type PwdDirs struct {
	FileId   string `json:"file_id"`  //文件id
	Password string `json:"password"` //密码
}
type Account struct {
	Id           string `json:"id"`            //网盘账号id
	Name         string `json:"name"`          //网盘账号名称
	Mode         string `json:"mode"`          //网盘模式，native（本地模式），cloud189(默认，天翼云网盘)，teambition（阿里teambition网盘），aliyundrive
	User         string `json:"user"`          //网盘账号用户名，邮箱或手机号
	Password     string `json:"password"`      //网盘账号密码
	RefreshToken string `json:"refresh_token"` //刷新token
	AccessToken  string `json:"access_token"`  //授权token
	RedirectUri  string `json:"redirect_uri"`  //重定向url（onedrive）
	RootId       string `json:"root_id"`       //目录id
	Default      int    `json:"default"`       //是否默认
	Seq          int    `json:"seq"`           //排序序号
	FilesCount   int    `json:"files_count"`   //文件总数
	Status       int    `json:"status"`        //状态：-1，缓存中 1，未缓存，2缓存成功，3缓存失败
	CookieStatus int    `json:"cookie_status"` //cookie状态：-1刷新中， 1未刷新，2正常，3失效
	TimeSpan     string `json:"time_span"`
	LastOpTime   string `json:"last_op_time"` //最近一次更新时间
	SyncDir      string `json:"sync_dir"`     //定时缓存指定目录
	SyncChild    int    `json:"sync_child"`   //是否缓存指定目录的子目录
	SyncCron     string `json:"sync_cron"`    //定时任务-缓存
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
type Cloud189 struct {
	Cloud189Session nic.Session
	SessionKey      string `json:"session_key"`
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
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
type User struct {
	UserName string
}
