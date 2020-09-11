package config

import (
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var Config189 Cloud189Config

func LoadCloud189Config(path string) {
	//配置文件读取优先级,自定义路径->当前路径->环境变量
	//配置优先级，环境变量最高
	if path == "" {
		path = "config.json"
	}
	b, err := PathExists(path)
	if err != nil {
		log.Fatal("PathExists(%s),err(%v)\n", path, err)
	}
	config := os.Getenv("CONFIG")
	port := os.Getenv("PORT")
	user := os.Getenv("CLOUD_USER")
	pwd := os.Getenv("CLOUD_PASSWORD")
	ri := os.Getenv("ROOT_ID")
	pdi := os.Getenv("PWD_DIR_ID")
	hfi := os.Getenv("HIDE_FILE_ID")
	hau := os.Getenv("HEROKU_APP_URL")
	if b {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fd, err := ioutil.ReadAll(file)
		config = string(fd)
	}
	err = jsoniter.Unmarshal([]byte(config), &Config189)
	if err != nil {
		log.Println("配置文件读取失败，从环境变量读取配置")
	}
	if port != "" {
		portInt, _ := strconv.Atoi(port)
		Config189.Port = portInt
	}
	if user != "" {
		Config189.User = user
	}
	if pwd != "" {
		Config189.Password = pwd
	}
	if ri != "" {
		Config189.RootId = ri
	}

	if pdi != "" {
		s := []PwdDirId{}
		pdiArr := strings.Split(pdi, ";")
		for _, a := range pdiArr {
			pwdDirId := PwdDirId{strings.Split(a, ":")[0], strings.Split(a, ":")[1]}
			s = append(s, pwdDirId)
		}
		Config189.PwdDirId = s
		//	Config189.Password = pwd
	}
	if hfi != "" {
		Config189.HideFileId = hfi
	}
	if hau != "" {
		Config189.HerokuAppUrl = hau
	}
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

type Cloud189Config struct {
	Port         int        `json:"port"`
	User         string     `json:"user"`
	Password     string     `json:"password"`
	RootId       string     `json:"root_id"`
	PwdDirId     []PwdDirId `json:"pwd_dir_id"`
	HideFileId   string     `json:"hide_file_id"`
	HerokuAppUrl string     `json:"heroku_app_url"`
}
type PwdDirId struct {
	Id  string `json:"id"`
	Pwd string `json:"pwd"`
}
