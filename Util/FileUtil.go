package Util

import (
	"mime"
	"os"
	"path/filepath"
	"strings"
)

func FileExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func IsDirectory(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func IsHiddenFile(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	return strings.HasPrefix(name, ".")
}

func GetMimeType(fileInfo os.FileInfo) int {
	mime := strings.Split(mime.TypeByExtension(filepath.Ext(fileInfo.Name())), "/")[0]
	if mime == "image" {
		return 1
	} else if mime == "audio" {
		return 2
	} else if mime == "video" {
		return 3
	} else if mime == "text" {
		return 4
	} else {
		return 0
	}
}
