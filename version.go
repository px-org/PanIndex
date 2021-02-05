package main

import (
	"log"
	"strings"
)

var (
	VERSION        string
	GO_VERSION     string
	BUILD_TIME     string
	GIT_COMMIT_SHA string
)

func PrintVersion() {
	GO_VERSION = strings.ReplaceAll(GO_VERSION, "go version ", "")
	log.Printf("Git Commit Hash: %s \n", GIT_COMMIT_SHA)
	log.Printf("Version: %s \n", VERSION)
	log.Printf("Go Version: %s \n", GO_VERSION)
	log.Printf("Build TimeStamp: %s \n", BUILD_TIME)

}
