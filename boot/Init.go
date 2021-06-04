package boot

import (
	"PanIndex/jobs"
	"PanIndex/model"
	"PanIndex/service"
	"encoding/json"
	"fmt"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	VERSION        string
	GO_VERSION     string
	BUILD_TIME     string
	GIT_COMMIT_SHA string
)

func Start(host, port string, debug bool) {
	//初始化日志设置
	InitLog(debug)
	//打印asc
	PrintAsc()
	//打印版本信息
	PrintVersion()
	//检查新版本
	go CheckUpdate()
	//初始化数据库
	model.InitDb(host, port, debug)
	//初始化配置
	//从环境变量写入到config
	service.EnvToConfig()
	service.GetConfig()
	//定时任务初始化
	jobs.Run()
	//刷新cookie和目录缓存
	//go jobs.StartInit()
}

func PrintAsc() {
	fmt.Println(`
 ____   __    _  _  ____  _  _  ____  ____  _  _ 
(  _ \ /__\  ( \( )(_  _)( \( )(  _ \( ___)( \/ )
 )___//(__)\  )  (  _)(_  )  (  )(_) ))__)  )  ( 
(__) (__)(__)(_)\_)(____)(_)\_)(____/(____)(_/\_)
`)
}

// boot logrus
func InitLog(debug bool) {
	if os.Getenv("PAN_INDEX_DEBUG") != "" {
		if os.Getenv("PAN_INDEX_DEBUG") == "true" {
			debug = true
		} else if os.Getenv("PAN_INDEX_DEBUG") == "false" {
			debug = false
		}
	}
	if debug {
		log.SetLevel(log.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	formatter := runtime.Formatter{ChildFormatter: &log.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05",
		FullTimestamp:             true,
	}}
	formatter.Line = true
	log.SetFormatter(&formatter)
}

func PrintVersion() {
	GO_VERSION = strings.ReplaceAll(GO_VERSION, "go version ", "")
	log.Printf("Git Commit Hash: %s \n", GIT_COMMIT_SHA)
	log.Printf("Version: %s \n", VERSION)
	log.Printf("Go Version: %s \n", GO_VERSION)
	log.Printf("Build TimeStamp: %s \n", BUILD_TIME)
}

// check updtae
func CheckUpdate() {
	log.Infof("检查更新...")
	url := "https://api.github.com/repos/libsgh/PanIndex/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		log.Warnf("检查更新失败:%s", err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("读取更新内容失败:%s", err.Error())
		return
	}
	if strings.Contains(string(body), "API rate limit") {
		log.Warnf("检查更新失败: API rate limit")
		return
	}
	var release GithubRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		log.Warnf("解析更新失败:%s", err.Error())
		return
	}
	lasted := release.TagName[1:]
	now := VERSION
	if IsLastVersion(lasted, now) {
		log.Infof("当前已是最新版本:%s", VERSION)
	} else {
		log.Infof("发现新版本:%s", release.TagName)
		log.Infof("请至'%s'获取更新.", release.HtmlUrl)
	}
}

func IsLastVersion(lasted string, now string) bool {
	if now != "" {
		lasted = strings.ReplaceAll(lasted, "v", "")
		lasted = strings.ReplaceAll(lasted, ".", "")
		lasted = strings.ReplaceAll(lasted, "BETA", "")
		lastedV, _ := strconv.Atoi(lasted)
		now = strings.ReplaceAll(now, "v", "")
		now = strings.ReplaceAll(now, ".", "")
		now = strings.ReplaceAll(now, "BETA", "")
		nowV, _ := strconv.Atoi(now)
		if lastedV > nowV {
			return false
		}
	}
	return true
}

type GithubRelease struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
	Body    string `json:"body"`
}
