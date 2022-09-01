package module

import (
	"github.com/go-resty/resty/v2"
	"github.com/smallnest/weighted"
	"time"
)

var (
	VERSION        string
	GO_VERSION     string
	BUILD_TIME     string
	GIT_COMMIT_SHA string
)

var GloablConfig Config

type FileNode struct {
	Id          string `json:"id"`                              //数据库唯一主键
	AccountId   string `json:"account_id" gorm:"index:idx_aid"` //文件所属账号
	FileId      string `json:"file_id" gorm:"index:idx_fid`     //网盘中的文件id
	FileName    string `json:"file_name" gorm:"index:idx_fn"`   //文件名称
	FileSize    int64  `json:"file_size"`                       //文件大小
	SizeFmt     string `json:"size_fmt"`                        //文件大小（格式化）
	FileType    string `json:"file_type"`                       //文件类型
	IsFolder    bool   `json:"is_folder"`                       //是否是目录
	LastOpTime  string `json:"last_op_time"`                    //最近一次操作时间
	ParentId    string `json:"parent_id"`                       //父目录id
	Path        string `json:"path" gorm:"index:idx_p"`         //文件路径
	ParentPath  string `json:"parent_path" gorm:"index:idx_pp"` //文件上层目录
	Thumbnail   string `json:"thumbnail" gorm:"-"`              //缩略图
	DownloadUrl string `json:"download_url" gorm:"-"`           //下载地址
	ViewType    string `json:"view_type"`                       //预览类型，取决于全局配置
	CreateTime  string `json:"create_time"`                     //创建时间（目录信息入库时间）
	IsDelete    int    `json:"is_delete" gorm:"index:idx_del"`  //删除标记（便于做缓存）
	Hide        int    `json:"hide"`                            //是否隐藏
	HasPwd      int    `json:"has_pwd"`                         //是否是密码文件（包含）
}
type ShareInfo struct {
	FilePath  string `json:"file_path"`  //PanIndex文件路径
	ShortCode string `json:"short_code"` //短链接code
	//IsFile    bool   `json:"is_file"`    //是否是文件（文件根据配置跳转预览或下载，目录直接打开）
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
	Accounts         []Account         `json:"accounts" gorm:"-"`
	SiteName         string            `json:"site_name"`
	AccountChoose    string            `json:"account_choose"`
	Theme            string            `json:"theme"`
	PathPrefix       string            `json:"path_prefix""` //路径前缀
	AdminUser        string            `json:"admin_user""`
	AdminPassword    string            `json:"admin_password""`
	OnlyReferrer     string            `json:"only_referrer"`
	EnableSafetyLink string            `json:"enable_safety_link"`
	IsNullReferrer   string            `json:"is_null_referrer"`
	FaviconUrl       string            `json:"favicon_url"`     //网站图标
	Footer           string            `json:"footer"`          //网站底部信息
	Css              string            `json:"css"`             //自定义css
	Js               string            `json:"js"`              //自定义js
	EnablePreview    string            `json:"enable_preview"`  //是否开启文件预览
	Image            string            `json:"image"`           //图片
	Audio            string            `json:"audio"`           //音频
	Video            string            `json:"video"`           //视频
	Code             string            `json:"code"`            //代码
	Doc              string            `json:"doc"`             //文档
	Other            string            `json:"other"`           //other
	EnableLrc        string            `json:"enable_lrc"`      //是否开启歌词
	LrcPath          string            `json:"lrc_path"`        //歌词路径
	Subtitle         string            `json:"subtitle"`        //字幕
	SubtitlePath     string            `json:"subtitle_path"`   //字幕路径
	Danmuku          string            `json:"danmuku"`         //弹幕
	DanmukuPath      string            `json:"danmuku_path"`    //弹幕路径
	SColumn          string            `json:"s_column"`        //排序字段
	SOrder           string            `json:"s_order"`         //排序顺序
	PwdFiles         []PwdFiles        `json:"pwd_files"`       //密码文件列表
	HideFiles        map[string]string `json:"hide_files"`      //隐藏文件
	AdminPath        string            `json:"admin_path"`      //后台管理路径前缀
	Cdn              string            `json:"cdn"`             //cdn
	CdnFiles         map[string]string `json:"-"`               //cdn files
	BypassList       []Bypass          `json:"bypass_list"`     //分流列表
	EnableDav        string            `json:"enable_dav"`      //dav enabled 1 disabled 0
	DavPath          string            `json:"dav_path"`        //dav path
	DavMode          string            `json:"dav_mode"`        //0 read-only, 1 read-write
	DavDownMode      string            `json:"dav_down_mode"`   //0 302 downloadurl, 1 proxy
	DavUser          string            `json:"dav_user"`        //dav user
	DavPassword      string            `json:"dav_password"`    //dav password
	Proxy            string            `json:"proxy"`           //google api prxoy
	Readme           string            `json:"readme"`          //show or hide readme
	Head             string            `json:"head"`            //show or hide head
	ShareInfoList    []ShareInfo       `json:"share_info_list"` //分享信息列表
	Access           string            `json:"access"`          //access
	ShortAction      string            `json:"short_action"`    //短链行为
}
type ConfigItem struct {
	K string `json:"k" gorm:"unique;not null"` //配置项key
	V string `json:"v"`                        //配置项值
	G string `json:"g"`                        //配置项分组
}

// table pwd_files
type PwdFiles struct {
	Id       string `json:"id"`        //主键id
	FilePath string `json:"file_path"` //文件路径
	Password string `json:"password"`  //密码
	ExpireAt int64  `json:"expire_at"` //失效时间
	Info     string `json:"info"`      //密码备注
}

// table hide_files
type HideFiles struct {
	FilePath string `json:"file_path"` //文件路径
}

type Account struct {
	Id             string `json:"id"`            //网盘账号id
	Name           string `json:"name"`          //网盘账号名称
	Mode           string `json:"mode"`          //网盘模式，native（本地模式），cloud189(默认，天翼云网盘)，teambition（阿里teambition网盘），aliyundrive
	User           string `json:"user"`          //网盘账号用户名，邮箱或手机号
	Password       string `json:"password"`      //网盘账号密码
	RefreshToken   string `json:"refresh_token"` //刷新token
	AccessToken    string `json:"access_token"`  //授权token
	RedirectUri    string `json:"redirect_uri"`  //重定向url（onedrive）
	ApiUrl         string `json:"api_url"`       //api地址
	RootId         string `json:"root_id"`       //目录id
	SiteId         string `json:"site_id"`       //网站id
	Seq            int    `json:"seq"`           //排序序号
	FilesCount     int    `json:"files_count"`   //文件总数
	Status         int    `json:"status"`        //状态：-1，缓存中 1，未缓存，2缓存成功，3缓存失败
	CookieStatus   int    `json:"cookie_status"` //cookie状态：-1刷新中， 1未刷新，2正常，3失效
	TimeSpan       string `json:"time_span"`
	LastOpTime     string `json:"last_op_time"`                      //最近一次更新时间
	SyncDir        string `json:"sync_dir"`                          //定时缓存指定目录
	SyncChild      int    `json:"sync_child"`                        //是否缓存指定目录的子目录
	SyncCron       string `json:"sync_cron"`                         //定时任务-缓存
	DownTransfer   int    `json:"down_transfer"`                     //是否开启下载中转
	TransferDomain string `json:"transfer_domain"`                   //中转地址，为空将使用本地服务器中转
	CachePolicy    string `json:"cache_policy" gorm:"default:nc"`    //缓存策略：nc（No Cache）,mc（Memory Cache）, dc（Database Cache）
	ExpireTimeSpan int    `json:"expire_time_span" gorm:"default:1"` //缓存时间单位是小时
	Host           string `json:"host"`                              //绑定host
	PathStyle      string `json:"path_style"`                        //S3 PathStyle :Path, Virtual Hosting
	Info           string `json:"info"`                              //备注
}

type Bypass struct {
	Id       string          `json:"id"`   //主键id
	Name     string          `json:"name"` //分流名称
	Accounts []Account       `json:"accounts" gorm:"-"`
	Rw       *weighted.RandW `json:"rw" gorm:"-"`
}

type BypassAccounts struct {
	BypassId  string `json:"bypass_id"`  //分流id
	AccountId string `json:"account_id"` //账号id
}

//cache data struct
type Cache struct {
	FilePath    string      `json:"file_path"`
	CacheTime   string      `json:"cache_time"`
	ExpireTime  string      `json:"expire_time"`
	CachePolicy string      `json:"cache_policy"`
	Data        interface{} `json:"data"`
}

type Ali struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type Cloud189 struct {
	Cloud189Session *resty.Client
	SessionKey      string `json:"session_key"`
	AccessToken     string `json:"access_token"`
	RootId          string `json:"root_id"`
	FamilyId        string `json:"family_id"`
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
	Sharepoint   string `json:"sharepoint"`
	DriveId      string `json:"drive_id"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Mode         string `json:"mode"`
}
type GoogleDriveAuthInfo struct {
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
type User struct {
	UserName string
}
type Zone struct {
	Login  string
	Api    string
	Upload string
	Desc   string
}
type Yun139 struct {
	Cookie string `json:"cookie"`
	Mobile string `json:"mobile"`
}

type UploadInfo struct {
	FileId      string
	FileName    string
	FileSize    int64
	ContentType string
	Content     []byte
}

type Teambition struct {
	GloablRootId    string
	GloablProjectId string
}

type PikpakToken struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	Sub          string `json:"sub"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
