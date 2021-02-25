package config

import (
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var GloablConfig CommonConfig
var Debug bool = false

func LoadConfig(path string) {
	//配置文件读取优先级,自定义路径->当前路径->环境变量
	//配置优先级，环境变量最高
	b, err := PathExists(path)
	if err != nil {
		log.Fatal("PathExists(%s),err(%v)\n", path, err)
	}
	config := os.Getenv("CONFIG")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	mode := os.Getenv("MODE")
	user := os.Getenv("CLOUD_USER")
	pwd := os.Getenv("CLOUD_PASSWORD")
	ri := os.Getenv("ROOT_ID")
	pdi := os.Getenv("PWD_DIR_ID")
	hfi := os.Getenv("HIDE_FILE_ID")
	hau := os.Getenv("HEROKU_APP_URL")
	apitk := os.Getenv("API_TOKEN")
	theme := os.Getenv("THEME")
	dmg_usr := os.Getenv("DMG_USER")
	dmg_pwd := os.Getenv("DMG_PASS")
	only_Referer := os.Getenv("ONLY_REFERER")
	cron_refresh_cookie := os.Getenv("CRON_REFRESH_COOKIE")
	cron_update_folder_cache := os.Getenv("CRON_UPDATE_FOLDER_CACHE")
	cron_heroku_keep_alive := os.Getenv("CRON_HEROKU_KEEP_ALIVE")
	if b && config == "" {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fd, err := ioutil.ReadAll(file)
		config = string(fd)
	}
	err = jsoniter.Unmarshal([]byte(config), &GloablConfig)
	if err != nil {
		log.Warnln("未发现配置文件，将从环境变量读取配置")
	}
	if host != "" {
		GloablConfig.Host = host
	}
	if port != "" {
		portInt, _ := strconv.Atoi(port)
		GloablConfig.Port = portInt
	}
	if mode != "" {
		GloablConfig.Mode = mode
	}
	if user != "" {
		GloablConfig.User = user
	}
	if pwd != "" {
		GloablConfig.Password = pwd
	}
	if ri != "" {
		GloablConfig.RootId = ri
	}

	if pdi != "" {
		s := []PwdDirId{}
		pdiArr := strings.Split(pdi, ";")
		for _, a := range pdiArr {
			pwdDirId := PwdDirId{strings.Split(a, ":")[0], strings.Split(a, ":")[1]}
			s = append(s, pwdDirId)
		}
		GloablConfig.PwdDirId = s
		//	GloablConfig.Password = pwd
	}
	if hfi != "" {
		GloablConfig.HideFileId = hfi
	}
	if hau != "" {
		GloablConfig.HerokuAppUrl = hau
	}
	if apitk != "" {
		GloablConfig.ApiToken = apitk
	}
	if theme != "" {
		GloablConfig.Theme = theme
	}
	if dmg_usr != "" {
		GloablConfig.Damagou.Username = dmg_usr
	}
	if dmg_pwd != "" {
		GloablConfig.Damagou.Password = dmg_pwd
	}
	if only_Referer != "" {
		GloablConfig.OnlyReferer = strings.Split(only_Referer, ",")
	}
	if cron_refresh_cookie != "" {
		GloablConfig.CronExps.RefreshCookie = cron_refresh_cookie
	}
	if cron_update_folder_cache != "" {
		GloablConfig.CronExps.UpdateFolderCache = cron_update_folder_cache
	}
	if cron_heroku_keep_alive != "" {
		GloablConfig.CronExps.HerokuKeepAlive = cron_heroku_keep_alive
	}
	//设置默认值
	if GloablConfig.Theme == "" {
		GloablConfig.Theme = "classic"
	}
	if GloablConfig.Mode == "" {
		GloablConfig.Mode = "cloud189"
	}
	if GloablConfig.CronExps.RefreshCookie == "" {
		GloablConfig.CronExps.RefreshCookie = "0 0 8 1/1 * ?"
	}
	if GloablConfig.CronExps.UpdateFolderCache == "" {
		GloablConfig.CronExps.UpdateFolderCache = "0 0 0/1 * * ?"
	}
	if GloablConfig.CronExps.HerokuKeepAlive == "" {
		GloablConfig.CronExps.HerokuKeepAlive = "0 0/5 * * * ?"
	}
	log.Infoln("[程序启动]配置加载 >> 获取成功")
}
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
