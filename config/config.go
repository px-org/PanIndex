package config

import (
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"log"
	"os"
)

var Config189 Cloud189Config

func LoadCloud189Config(path string) {
	//配置读取优先级,自定义路径->当前路径->环境变量
	if path == "" {
		path = "config.json"
	}
	b, err := PathExists(path)
	if err != nil {
		log.Fatal("PathExists(%s),err(%v)\n", path, err)
	}
	config := os.Getenv("CONFIG")
	port := os.Getenv("PORT")
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
		log.Fatal("配置文件读取失败：", err)
	}
	if port != "" {
		Config189.Port = port
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
