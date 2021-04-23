package entity

type FileNode struct {
	FileId       string `json:"fileId" gorm:"primary_key"`
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
	Icon         Icon   `json:"icon"`
	CreateTime   string `json:"create_time"`
	Delete       int    `json:"delete"`
	Hide         int    `json:"hide"`
}
type Icon struct {
	LargeUrl string `json:"largeUrl"`
	SmallUrl string `json:"smallUrl"`
}
type Paths struct {
	FileId    string `json:"fileId"`
	FileName  string `json:"fileName"`
	IsCoShare int    `json:"isCoShare"`
}
type Config struct {
	Host          string     `json:"host" gorm:"default:'0.0.0.0'"`
	Port          int        `json:"port" gorm:"default:8080"`
	Accounts      []Account  `json:"accounts" gorm:"-"`
	PwdDirId      []PwdDirId `json:"pwd_dir_id" gorm:"-"`
	HideFileId    string     `json:"hide_file_id"`
	HerokuAppUrl  string     `json:"heroku_app_url"`
	ApiToken      string     `json:"api_token"`
	Theme         string     `json:"theme" gorm:"default:'mdui'"`
	AdminPassword string     `json:"admin_password" gorm:"default:'PanIndex'"`
	Damagou       Damagou    `json:"damagou" gorm:"-"`
	OnlyReferer   []string   `json:"only_referrer" gorm:"-"`
	CronExps      CronExps   `json:"cron_exps" gorm:"-"`
	Footer        string     `json:"footer"` //网站底部信息
}

type Account struct {
	Id           string `json:"id"`            //网盘空间id
	Name         string `json:"name"`          //网盘空间名称
	Mode         string `json:"mode"`          //网盘模式，native（本地模式），cloud189(默认，天翼云网盘)，teambition（阿里teambition网盘）
	User         string `json:"user"`          //网盘账号用户名，邮箱或手机号
	Password     string `json:"password"`      //网盘账号密码
	RefreshToken string `json:"refresh_token"` //刷新token
	AccessToken  string `json:"access_token"`  //授权token
	RootId       string `json:"root_id"`       //目录id
	Default      int    `json:"default"`       //是否默认
}
type CronExps struct {
	RefreshCookie     string `json:"refresh_cookie" gorm:"default:'0 0 8 1/1 * ?'"`
	UpdateFolderCache string `json:"update_folder_cache" gorm:"default:'0 0 0/1 * * ?'"`
	HerokuKeepAlive   string `json:"heroku_keep_alive" gorm:"default:'0 0/5 * * * ?'"`
}
type PwdDirId struct {
	Id  string `json:"id"`
	Pwd string `json:"pwd"`
}
type Damagou struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
