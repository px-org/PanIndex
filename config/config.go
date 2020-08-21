package config

import (
	jsoniter "github.com/json-iterator/go"
	"log"
	"os"
)

var Config189 Cloud189Config

func LoadCloud189Config() {
	config := os.Getenv("CONFIG")
	err := jsoniter.Unmarshal([]byte(config), &Config189)
	if err != nil {
		log.Fatal("err：", err)
	}
}

type Cloud189Config struct {
	User      string `json:"user"`
	Password  string `json:"password"`
	RootId    string `json:"root_id"`
	PwdDirId  string `json:"pwd_dir_id"`
	HideDirId string `json:"hide_dir_id"`
}
