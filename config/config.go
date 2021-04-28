package config

import (
	"PanIndex/entity"
	"os"
)

var GloablConfig entity.Config

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

type CommonConfig struct {
	Host         string     `json:"host"`
	Port         int        `json:"port"`
	Mode         string     `json:"mode"` //网盘模式，native（本地模式），cloud189(默认，天翼云网盘)，teambition（阿里teambition网盘）
	User         string     `json:"user"`
	Password     string     `json:"password"`
	RootId       string     `json:"root_id"`
	PwdDirId     []PwdDirId `json:"pwd_dir_id"`
	HideFileId   string     `json:"hide_file_id"`
	HerokuAppUrl string     `json:"heroku_app_url"`
	ApiToken     string     `json:"api_token"`
	Theme        string     `json:"theme"`
	Damagou      Damagou    `json:"damagou"`
	OnlyReferer  []string   `json:"only_referrer"`
	CronExps     CronExps   `json:"cron_exps"`
}

type CronExps struct {
	RefreshCookie     string `json:"refresh_cookie"`
	UpdateFolderCache string `json:"update_folder_cache"`
	HerokuKeepAlive   string `json:"heroku_keep_alive"`
}

type PwdDirId struct {
	Id  string `json:"id"`
	Pwd string `json:"pwd"`
}

type Damagou struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
