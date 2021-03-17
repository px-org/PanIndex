package Util

import (
	"bytes"
	"fmt"
	"github.com/bluele/gcache"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var GC = gcache.New(10).LRU().Build()

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

func GetPrePath(path string) []map[string]string {
	//path := "/a/b/c/d"
	prePaths := []map[string]string{}
	//result := make(map[string]interface{})
	paths := strings.Split(path, "/")
	for i, n := range paths {
		item := make(map[string]string)
		var buffer bytes.Buffer
		for j := 0; j <= i; j++ {
			if paths[j] == "" {
				buffer.WriteString(paths[j])
			} else {
				buffer.WriteString("/")
				buffer.WriteString(paths[j])
			}
		}
		if buffer.String() != "" {
			item["PathName"] = n
			item["PathUrl"] = buffer.String()
			prePaths = append(prePaths, item)
		}
	}
	return prePaths
}
func ReadStringByFile(filePth string) string {
	f, err := os.Open(filePth)
	if err != nil {
		log.Errorln(err)
		return ""
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorln(err)
		return ""
	}
	return fmt.Sprintf("%s", b)
}
func ReadStringByUrl(url, fileId string) string {
	content := ""
	//为了提高效率，从缓存查询
	value, _ := GC.Get(fileId)
	if value != nil {
		log.Debugf("从缓存中读取README.md内容{%s}", fileId)
		return value.(string)
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Errorln(err)
		return content
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln(err)
		return content
	}
	content = fmt.Sprintf("%s", data)
	GC.Set(fileId, content)
	return content
}
