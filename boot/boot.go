package boot

import (
	"embed"
	"flag"
	"fmt"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/PanIndex/dao"
	"github.com/libsgh/PanIndex/jobs"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/util"
	log "github.com/sirupsen/logrus"
	"github.com/unrolled/secure"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func Init() (BootConfig, bool) {
	//load config
	config := LoadConfig()
	//Print config
	configStr, _ := jsoniter.MarshalToString(config)
	result := PrintConfig(config.ConfigQuery, configStr)
	if result {
		return config, true
	}
	//Create data dir
	config = CreatDataDir(config)
	//int log level
	InitLog(config.LogLevel)
	//print asc
	PrintAsc()
	//print version
	PrintVersion()
	//init dao
	InitDb(config)
	// init global config
	dao.InitGlobalConfig()
	configStr, _ = jsoniter.MarshalToString(module.GloablConfig)
	result = PrintConfig(config.ConfigQuery, configStr)
	//init accounts auth login
	for _, account := range module.GloablConfig.Accounts {
		dao.SyncAccountStatus(account)
	}
	//init all jobs
	jobs.Run()
	return config, result
}

func PrintConfig(query string, config string) bool {
	if query != "" {
		v := jsoniter.Get([]byte(config), query).ToString()
		if v != "" {
			fmt.Print(v)
			return true
		}
	}
	return false
}

func CreatDataDir(config BootConfig) BootConfig {
	dataPath := config.DataPath
	if config.DataPath == "" {
		ex, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		dataPath = filepath.Join(filepath.Dir(ex), "data")
	}
	err := os.MkdirAll(dataPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.DataPath = dataPath
	return config
}

func InitDb(config BootConfig) {
	driver := "sqlite"
	dsn := filepath.FromSlash(path.Join(config.DataPath, "data.db"))
	if config.DbType != "" {
		driver = config.DbType
	}
	if config.Dsn != "" {
		dsn = config.Dsn
	}
	d, _ := dao.GetDb(driver)
	d.CreateDb(dsn)
}

//config.json > env > flag
func LoadConfig() BootConfig {
	var Config = flag.String("c", "config.json", "config.json")
	var Host = flag.String("host", "", "bind host, default 0.0.0.0")
	var Port = flag.String("port", "", "bind port, default 5238")
	var LogLevel = flag.String("log_level", "", "log level: debug, info")
	var DataPath = flag.String("data_path", "", "data storage directory, default program sibling directory")
	var CertFile = flag.String("cert_file", "", "https cert file, /path/to/test.pem")
	var KeyFile = flag.String("key_file", "", "https key file, /path/to/test.key")
	var ConfigQueryOld = flag.String("cq", "", "config query old version, e.g. port")
	var ConfigQuery = flag.String("config_query", "", "config query new version, e.g. port")
	var DbType = flag.String("db_type", "", "dao type, e.g. sqlite,mysql,postgres...")
	var Dsn = flag.String("dsn", "", "database connection url")
	flag.Parse()
	config, _ := LoadFromFile(*Config)
	config.Host = LoadFromEnv("HOST", *Host, config.Host)
	if config.Host == "" {
		config.Host = "0.0.0.0"
	}
	portStr := LoadFromEnv("PORT", *Port, strconv.Itoa(config.Port))
	if portStr == "" || portStr == "0" {
		config.Port = 5238
	} else {
		port, _ := strconv.Atoi(portStr)
		config.Port = port
	}
	config.LogLevel = LoadFromEnv("LOG_LEVEL", *LogLevel, config.LogLevel)
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	config.DataPath = LoadFromEnv("DATA_PATH", *DataPath, config.DataPath)
	config.CertFile = LoadFromEnv("CERT_FILE", *CertFile, config.CertFile)
	config.KeyFile = LoadFromEnv("KEY_FILE", *KeyFile, config.KeyFile)
	config.DbType = LoadFromEnv("DB_TYPE", *DbType, config.DbType)
	config.Dsn = LoadFromEnv("DSN", *Dsn, config.Dsn)
	config.ConfigQuery = LoadFromEnv("CONFIG_QUERY", *ConfigQuery, config.ConfigQuery)
	if *ConfigQueryOld != "" {
		config.ConfigQuery = *ConfigQueryOld
	}
	return *config
}

func LoadFromEnv(key string, def, cv string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	if def != "" {
		return def
	}
	return cv
}

func LoadFromFile(path string) (*BootConfig, error) {
	config := new(BootConfig)
	configFile, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("Unable to read configuration file %s", path)
	}
	decoder := jsoniter.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("Unable to parse configuration file %s", path)
	}
	return config, nil
}
func PrintAsc() {
	fmt.Println(`
 ____   __    _  _  ____  _  _  ____  ____  _  _ 
(  _ \ /__\  ( \( )(_  _)( \( )(  _ \( ___)( \/ )
 )___//(__)\  )  (  _)(_  )  (  )(_) ))__)  )  ( 
(__) (__)(__)(_)\_)(____)(_)\_)(____/(____)(_/\_)
`)
}

func PrintVersion() {
	module.GO_VERSION = strings.ReplaceAll(module.GO_VERSION, "go version ", "")
	log.Printf("Git Commit Hash: %s", module.GIT_COMMIT_SHA)
	log.Printf("Version: %s", module.VERSION)
	log.Printf("Go Version: %s", module.GO_VERSION)
	log.Printf("Build TimeStamp: %s", module.BUILD_TIME)
}

func TlsHandler(port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     ":" + strconv.Itoa(port),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			return
		}
		c.Next()
	}
}

// logrus config
func InitLog(lvl string) error {
	level, err := log.ParseLevel(lvl)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	if lvl == "debug" {
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
	return nil
}

func InitStaticBox(r *gin.Engine, fs embed.FS) {
	if util.FileExist("./static") {
		r.StaticFS("/static", http.Dir("./static"))
	} else {
		r.Any("/static/*filepath", func(c *gin.Context) {
			staticServer := http.FileServer(http.FS(fs))
			staticServer.ServeHTTP(c.Writer, c.Request)
		})
	}
}

func Templates(fs embed.FS) *template.Template {
	themes := [3]string{"mdui", "classic", "bootstrap"}
	tmpl := template.New("")
	templatesFileNames := []string{"base", "appearance", "common", "disk", "hide", "login", "pwd", "safety", "view", "bypass", "cache", "webdav"}
	addTemplatesFromFolder("admin", tmpl, fs, templatesFileNames)
	for _, theme := range themes {
		theme = util.GetCurrentTheme(theme)
		tmpFile := strings.Join([]string{"templates/pan/", "/index.html"}, theme)
		dataBuf, _ := fs.ReadFile(tmpFile)
		data := string(dataBuf)
		if util.FileExist("./" + tmpFile) {
			s, _ := ioutil.ReadFile("./" + tmpFile)
			data = string(s)

		}
		tmpl.New(tmpFile).Funcs(template.FuncMap{
			"unescaped":    unescaped,
			"contains":     strings.Contains,
			"iconclass":    iconclass,
			"FormateName":  FormateName,
			"TruncateName": TruncateName,
		}).Parse(data)
	}
	//添加详情模板
	viewTemplates := [10]string{"base", "img", "audio", "video", "code", "office", "ns", "pdf", "md", "epub"}
	for _, vt := range viewTemplates {
		tmpName := fmt.Sprintf("templates/pan/%s/view-%s.html", "mdui", vt)
		dataBuf, _ := fs.ReadFile(tmpName)
		data := string(dataBuf)
		if util.FileExist("./" + tmpName) {
			s, _ := ioutil.ReadFile("./" + tmpName)
			data = string(s)
		}
		tmpl.New(tmpName).Funcs(template.FuncMap{
			"unescaped":    unescaped,
			"contains":     strings.Contains,
			"iconclass":    iconclass,
			"FormateName":  FormateName,
			"TruncateName": TruncateName,
		}).Parse(data)
	}
	return tmpl
}

func addTemplatesFromFolder(folder string, tmpl *template.Template, fs embed.FS, templatesFileNames []string) {
	for _, vt := range templatesFileNames {
		tmpName := fmt.Sprintf("templates/pan/%s/%s.html", folder, vt)
		dataBuf, _ := fs.ReadFile(tmpName)
		data := string(dataBuf)
		if util.FileExist("./" + tmpName) {
			s, _ := ioutil.ReadFile("./" + tmpName)
			data = string(s)
		}
		tmpl.New(tmpName).Funcs(template.FuncMap{
			"unescaped":    unescaped,
			"contains":     strings.Contains,
			"iconclass":    iconclass,
			"isLast":       isLast,
			"FormateName":  FormateName,
			"TruncateName": TruncateName,
		}).Parse(data)
	}
}

type BootConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	LogLevel    string `json:"log_level"`
	DataPath    string `json:"data_path"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
	ConfigQuery string `json:"config_query"`
	DbType      string `json:"db_type"` //dao type:sqlite,mysql,postgres
	Dsn         string `json:"dsn"`     //dao dsn
}

func unescaped(x string) interface{} { return template.HTML(x) }

func isLast(index int, len int) bool {
	return index+1 == len
}

func iconclass(isFolder bool, fileType string) string {
	return util.GetIcon(isFolder, fileType)
}

func FormateName(filename string) string {
	filenameAll := path.Base(filename)
	fileSuffix := path.Ext(filename)
	filePrefix := filenameAll[0 : len(filenameAll)-len(fileSuffix)]
	return filePrefix
}

func TruncateName(filename string) string {
	nameRune := []rune(filename)
	if len(nameRune) > 20 {
		return string(nameRune[0:20]) + "..."
	}
	return filename
}
