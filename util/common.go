package util

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/qingstor/go-mime"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	math_rand "math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Timespan time.Duration

var CacheCronMap = make(map[string]cron.EntryID)

var Cron *cron.Cron

func (t Timespan) Format(format string) string {
	z := time.Unix(0, 0).UTC()
	return z.Add(time.Duration(t)).Format(format)
}

func ShortDur(d time.Duration) string {
	z := time.Unix(0, 0).UTC()
	return z.Add(d).Format("15:04:05")
}

func FormatFileSize(fileSize int64) (size string) {
	if fileSize == 0 {
		return "-"
	} else if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2f B", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2f KB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f MB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f GB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2f TB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2f EB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

func UTCTimeFormat(timeStr string) string {
	t, _ := time.Parse(time.RFC3339, timeStr)
	timeUint := t.In(time.Local).Unix()
	return time.Unix(timeUint, 0).Format("2006-01-02 15:04:05")
}

func UTCTime(t time.Time) string {
	timeUint := t.In(time.Local).Unix()
	return time.Unix(timeUint, 0).Format("2006-01-02 15:04:05")
}

func GetViewType(fileType string) string {
	config := module.GloablConfig
	if fileType == "" {
		return "ns"
	}
	if strings.Contains(config.Image, fileType) {
		return "img"
	} else if strings.Contains(config.Audio, fileType) {
		return "audio"
	} else if strings.Contains(config.Video, fileType) {
		return "video"
	} else if strings.Contains(config.Code, fileType) {
		return "code"
	} else if strings.Contains(config.Doc, fileType) {
		return "office"
	} else if fileType == "pdf" {
		return "pdf"
	} else if fileType == "md" {
		return "md"
	} else if fileType == "epub" {
		return "epub"
	} else {
		return "ns"
	}
}

func GetExt(name string) string {
	ext := strings.ToLower(strings.TrimLeft(filepath.Ext(name), "."))
	return ext
}

type KV map[string]interface{}

func GetIcon(isFolder bool, fileType string) string {
	iconMap := KV{
		"mdui": KV{
			"folder":  "folder_open",       //文件夹
			"image":   "image",             //图片
			"audio":   "audio_file",        //音频
			"video":   "video_file",        //视频
			"apk":     "android",           //安卓apk
			"archive": "folder_zip",        //压缩包
			"file":    "insert_drive_file", //普通文件
			"exe":     "apps",              //windows可执行文件
			"code":    "code",              //代码
			"txt":     "text_snippet",      //文本
			"pdf":     "picture_as_pdf",
			"md":      "text_snippet",
		},
		"classic": KV{
			"folder":  "icon icon-dir",  //文件夹
			"image":   "icon icon-file", //图片
			"audio":   "icon icon-file", //音频
			"video":   "icon icon-file", //视频
			"apk":     "icon icon-file", //安卓apk
			"archive": "icon icon-file", //压缩包
			"file":    "icon icon-file", //普通文件
			"exe":     "icon icon-file", //windows可执行文件
			"code":    "icon icon-file", //代码
			"txt":     "icon icon-file", //文本
			"pdf":     "icon icon-file",
			"md":      "icon icon-file",
		},
		"bootstrap": KV{
			"folder":  "fas fa-folder",       //文件夹
			"image":   "far fa-file-image",   //图片
			"audio":   "fas fa-music",        //音频
			"video":   "fab fa-youtube",      //视频
			"apk":     "fab fa-android",      //安卓apk
			"archive": "fas fa-file-archive", //压缩包
			"file":    "fas fa-file",         //普通文件
			"exe":     "fab fa-microsoft",    //windows可执行文件
			"code":    "fas fa-code",         //代码
			"txt":     "fas fa-file-alt",     //文本
			"pdf":     "fas fa-file-alt",
			"md":      "fas fa-file-alt",
		},
	}
	config := module.GloablConfig
	fileKey := "file"
	if isFolder {
		fileKey = "folder"
	} else {
		if strings.Contains(config.Image, fileType) {
			fileKey = "image"
		} else if strings.Contains(config.Audio, fileType) {
			fileKey = "audio"
		} else if strings.Contains(config.Video, fileType) {
			fileKey = "video"
		} else if strings.Contains(config.Code, fileType) {
			fileKey = "txt"
		} else if strings.Contains(config.Doc, fileType) {
			fileKey = "txt"
		} else if strings.Contains("pdf", fileType) {
			fileKey = "pdf"
		} else if strings.Contains("md", fileType) {
			fileKey = "md"
		} else if strings.Contains("zip,gz,7z,rar", fileType) {
			fileKey = "archive"
		} else if fileType == "apk" {
			fileKey = "apk"
		} else if fileType == "exe" {
			fileKey = "exe"
		}
	}
	return iconMap[GetCurrentTheme(config.Theme)].(KV)[fileKey].(string)
}
func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	} else {
		n = n + len(start)
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}

func GetPrePath(path string) []map[string]string {
	//path := "/a/b/c/d"
	prePaths := []map[string]string{}
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

func GetParentPath(p string) string {
	if p == "/" {
		return ""
	} else {
		s := ""
		ss := strings.Split(p, "/")
		for i := 0; i < len(ss)-1; i++ {
			if ss[i] != "" {
				s += "/" + ss[i]
			}
		}
		if s == "" {
			s = "/"
		}
		return s
	}
}

func Md5(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

const (
	VAL   = 0x3FFFFFFF
	INDEX = 0x0000003D
)

var (
	alphabet = []byte("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

/** implementation of short url algorithm **/
func Transform(longURL string) ([4]string, error) {
	md5Str := Md5(longURL)
	//var hexVal int64
	var tempVal int64
	var result [4]string
	var tempUri []byte
	for i := 0; i < 4; i++ {
		tempSubStr := md5Str[i*8 : (i+1)*8]
		hexVal, err := strconv.ParseInt(tempSubStr, 16, 64)
		if err != nil {
			return result, nil
		}
		tempVal = int64(VAL) & hexVal
		var index int64
		tempUri = []byte{}
		for i := 0; i < 6; i++ {
			index = INDEX & tempVal
			tempUri = append(tempUri, alphabet[index])
			tempVal = tempVal >> 5
		}
		result[i] = string(tempUri)
	}
	return result, nil
}

func SortFileNodeNew(sortColumn, sortOrder string, list []module.FileNode) {
	if sortColumn != "default" && sortOrder != "null" {
		sort.SliceStable(list, func(i, j int) bool {
			li, _ := time.Parse("2006-01-02 15:04:05", list[i].LastOpTime)
			lj, _ := time.Parse("2006-01-02 15:04:05", list[j].LastOpTime)
			d1 := 0
			if list[i].IsFolder {
				d1 = 1
			}
			d2 := 0
			if list[j].IsFolder {
				d2 = 1
			}
			if d1 > d2 {
				return true
			} else if d1 == d2 {
				if sortColumn == "file_name" {
					c := strings.Compare(list[i].FileName, list[j].FileName)
					if sortOrder == "desc" {
						return c >= 0
					} else {
						return c <= 0
					}
				} else if sortColumn == "file_size" {
					if sortOrder == "desc" {
						return list[i].FileSize >= list[j].FileSize
					} else {
						return list[i].FileSize <= list[j].FileSize
					}
				} else if sortColumn == "last_op_time" {
					if sortOrder == "desc" {
						return li.After(lj)
					} else {
						return li.Before(lj)
					}
				} else {
					return lj.After(li)
				}
			} else {
				return false
			}
		})
	}
}
func SortFileNode(sortColumn, sortOrder string, list []module.FileNode) {
	if sortColumn == "default" {
		sortColumn = "file_name"
	}
	if sortOrder == "null" {
		sortColumn = "asc"
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].IsFolder != list[j].IsFolder {
			return list[i].IsFolder
		}
		if sortColumn == "file_name" {
			c := strings.Compare(list[i].FileName, list[j].FileName)
			if c != 0 {
				if sortOrder == "desc" {
					return c > 0
				} else {
					return c < 0
				}
			}
		} else if sortColumn == "file_size" {
			if list[i].FileSize != list[j].FileSize {
				if sortOrder == "desc" {
					return list[i].FileSize > list[j].FileSize
				} else {
					return list[i].FileSize < list[j].FileSize
				}
			}
		} else if sortColumn == "last_op_time" {
			li, _ := time.Parse("2006-01-02 15:04:05", list[i].LastOpTime)
			lj, _ := time.Parse("2006-01-02 15:04:05", list[j].LastOpTime)
			if sortOrder == "desc" {
				return li.After(lj)
			} else {
				return li.Before(lj)
			}
		}
		return false
	})
}

//TODO
/*func GetPwdFromPath(path string) (module.PwdFiles, bool) {
	pwdDir := module.PwdFiles{}
	pwdFile := module.PwdFiles{}
	pwdMaps := module.GloablConfig.PwdFiles
	for k, v := range pwdMaps {
		if strings.HasPrefix(path, k) {
			pwdDir.FilePath = k
			pwdDir.Password = v
		}
		if k == path {
			pwdFile.FilePath = k
			pwdFile.Password = v
		}
	}
	if pwdFile.FilePath != "" {
		return pwdFile, true
	}
	if pwdDir.FilePath != "" {
		return pwdDir, true
	}
	return pwdDir, false
}*/

func GetCdnFilesMap(cdn, version string) map[string]string {
	if version == "" {
		version = "main"
	}
	prefix := module.GloablConfig.PathPrefix
	jp := "https://fastly.jsdelivr.net/gh/px-org/PanIndex@" + version
	//jp := "https://fastly.jsdelivr.net/gh/libsgh/PanIndex@" + version
	m := map[string]string{}
	cdnMap := KV{
		"0": KV{
			"mdui@css":                   prefix + "/static/lib/mdui@1.0.2/css/mdui.min.css",
			"mdui@js":                    prefix + "/static/lib/mdui@1.0.2/js/mdui.min.js",
			"viewer@css":                 prefix + "/static/lib/viewerjs@1.10.1/dist/viewer.min.css",
			"viewer@js":                  prefix + "/static/lib/viewerjs@1.10.1/dist/viewer.min.js",
			"jquery@js":                  prefix + "/static/lib/jquery@3.5.1/jquery.min.js",
			"cookie@js":                  prefix + "/static/lib/js-cookie@3.0.1/dist/js.cookie.min.js",
			"md5@js":                     prefix + "/static/lib/md5/md5.min.js",
			"marked@js":                  prefix + "/static/lib/marked/marked.min.js",
			"clipboard@js":               prefix + "/static/lib/clipboard@2.0.8/clipboard.min.js",
			"mdui@index@js":              prefix + "/static/js/mdui.index.js",
			"mdui@index@css":             prefix + "/static/css/index.css",
			"sortablejs@js":              prefix + "/static/lib/sortablejs@1.14.0/Sortable.min.js",
			"admin@js":                   prefix + "/static/js/admin.js",
			"fontawesome@css":            prefix + "/static/lib/fontawesome@5.15.4/css/all.min.css",
			"APlayer@css":                prefix + "/static/lib/aplayer@1.10.1/dist/APlayer.min.css",
			"APlayer@js":                 prefix + "/static/lib/aplayer@1.10.1/dist/APlayer.min.js",
			"sweetalert2@css":            prefix + "/static/lib/sweetalert2@11.3.0/dist/sweetalert2.min.css",
			"sweetalert2@js":             prefix + "/static/lib/sweetalert2@11.3.0/dist/sweetalert2.min.js",
			"hls@js":                     prefix + "/static/lib/hls.js@1.1.2/dist/hls.min.js",
			"flv@js":                     prefix + "/static/lib/flv.js@1.6.2/dist/flv.min.js",
			"artplayer@js":               prefix + "/static/lib/artplayer/artplayer.js",
			"artplayer-danmuku@js":       prefix + "/static/lib/artplayer/artplayer-plugin-danmuku.js",
			"video@mdui@js":              prefix + "/static/js/mdui.video.js",
			"video@simple@js":            prefix + "/static/js/simple.video.js",
			"simple@index@js":            prefix + "/static/js/simple.index.js",
			"highlightjs@atom@dark@css":  prefix + "/static/lib/highlightjs/cdn-release@11.4.0/build/styles/atom-one-dark.min.css",
			"highlightjs@atom@light@css": prefix + "/static/lib/highlightjs/cdn-release@11.4.0/build/styles/atom-one-light.min.css",
			"highlight@js":               prefix + "/static/lib/highlightjs/cdn-release@11.4.0/build/highlight.min.js",
			"jszip@js":                   prefix + "/static/lib/jszip@3.1.5/jszip.min.js",
			"epub@js":                    prefix + "/static/lib/epubjs@0.3.88/epub.min.js",
			"pdfh5@css":                  prefix + "/static/lib/pdfh5@1.4.0/css/pdfh5.css",
			"pdf@js":                     prefix + "/static/lib/pdfh5@1.4.0/js/pdf.js",
			"pdf@worker@js":              prefix + "/static/lib/pdfh5@1.4.0/js/pdf.worker.js",
			"pdfh5@js":                   prefix + "/static/lib/pdfh5@1.4.0/js/pdfh5.js",
			"natural@compare@js":         prefix + "/static/lib/natural-compare-lite@1.4.0/index.min.js",
			"bootstrap@css":              prefix + "/static/lib/bootstrap@4.6.1/css/bootstrap.min.css",
			"bootstrap@js":               prefix + "/static/lib/bootstrap@4.6.1/js/bootstrap.min.js",
			"Material+Icons@css":         prefix + "/static/css/Material+Icons.css",
		},
		"1": KV{
			"mdui@css":                   "//cdn.staticfile.org/mdui/1.0.2/css/mdui.min.css",
			"mdui@js":                    "//cdn.staticfile.org/mdui/1.0.2/js/mdui.min.js",
			"viewer@css":                 "//cdn.staticfile.org/viewerjs/1.10.1/viewer.min.css",
			"viewer@js":                  "//cdn.staticfile.org/viewerjs/1.10.1/viewer.min.js",
			"jquery@js":                  "//cdn.staticfile.org/jquery/3.5.1/jquery.min.js",
			"cookie@js":                  "//cdn.staticfile.org/js-cookie/latest/js.cookie.min.js",
			"md5@js":                     "//cdn.staticfile.org/blueimp-md5/1.0.1/js/md5.min.js",
			"marked@js":                  "//cdn.staticfile.org/marked/4.0.2/marked.min.js",
			"clipboard@js":               "//cdn.staticfile.org/clipboard.js/2.0.8/clipboard.min.js",
			"mdui@index@js":              prefix + "/static/js/mdui.index.js",
			"mdui@index@css":             prefix + "/static/css/index.css",
			"sortablejs@js":              "//cdn.staticfile.org/Sortable/1.14.0/Sortable.min.js",
			"admin@js":                   prefix + "/static/js/admin.js",
			"fontawesome@css":            "//cdn.staticfile.org/font-awesome/5.15.4/css/all.min.css",
			"APlayer@css":                "//lf6-cdn-tos.bytecdntp.com/cdn/expire-1-M/aplayer/1.10.1/APlayer.min.css",
			"APlayer@js":                 "//lf26-cdn-tos.bytecdntp.com/cdn/expire-1-M/aplayer/1.10.1/APlayer.min.js",
			"sweetalert2@css":            prefix + "/static/lib/sweetalert2@11.3.0/dist/sweetalert2.min.css",
			"sweetalert2@js":             prefix + "/static/lib/sweetalert2@11.3.0/dist/sweetalert2.min.js",
			"hls@js":                     "//cdn.staticfile.org/hls.js/1.1.2/hls.min.js",
			"flv@js":                     "//cdn.staticfile.org/flv.js/1.6.2/flv.min.js",
			"artplayer@js":               "//fastly.jsdelivr.net/npm/artplayer@4.5.4/dist/artplayer.js",
			"artplayer-danmuku@js":       "//fastly.jsdelivr.net/npm/artplayer-plugin-danmuku@4.4.8/dist/artplayer-plugin-danmuku.js",
			"video@mdui@js":              prefix + "/static/js/mdui.video.js",
			"video@simple@js":            prefix + "/static/js/simple.video.js",
			"simple@index@js":            prefix + "/static/js/simple.index.js",
			"highlightjs@atom@dark@css":  "//cdn.bootcdn.net/ajax/libs/highlight.js/11.4.0/styles/atom-one-dark.min.css",
			"highlightjs@atom@light@css": "//cdn.bootcdn.net/ajax/libs/highlight.js/11.4.0/styles/atom-one-light.min.css",
			"highlight@js":               "//cdn.bootcdn.net/ajax/libs/highlight.js/11.4.0/highlight.min.js",
			"jszip@js":                   "//lf9-cdn-tos.bytecdntp.com/cdn/expire-1-M/jszip/3.1.5/jszip.js",
			"epub@js":                    "//fastly.jsdelivr.net/npm/epubjs@0.3.88/dist/epub.js",
			"pdfh5@css":                  "//fastly.jsdelivr.net/npm/pdfh5@1.4.2/css/pdfh5.css",
			"pdf@js":                     "//fastly.jsdelivr.net/npm/pdfh5@1.4.2/js/pdf.js",
			"pdf@worker@js":              "//fastly.jsdelivr.net/npm/pdfh5@1.4.2/js/pdf.worker.js",
			"pdfh5@js":                   "//fastly.jsdelivr.net/npm/pdfh5@1.4.2/js/pdfh5.js",
			"natural@compare@js":         "//fastly.jsdelivr.net/npm/natural-compare-lite@1.4.0/index.js",
			"bootstrap@css":              "//cdn.staticfile.org/bootstrap/4.6.1/css/bootstrap.min.css",
			"bootstrap@js":               "//cdn.staticfile.org/bootstrap/4.6.1/js/bootstrap.min.js",
			"Material+Icons@css":         "//fonts.loli.net/icon?family=Material+Icons",
		},
		"2": KV{
			"mdui@css":                   jp + "/static/lib/mdui@1.0.2/css/mdui.min.css",
			"mdui@js":                    jp + "/static/lib/mdui@1.0.2/js/mdui.min.js",
			"viewer@css":                 jp + "/static/lib/viewerjs@1.10.1/dist/viewer.min.css",
			"viewer@js":                  jp + "/static/lib/viewerjs@1.10.1/dist/viewer.min.js",
			"jquery@js":                  jp + "/static/lib/jquery@3.5.1/jquery.min.js",
			"cookie@js":                  jp + "/static/lib/js-cookie@3.0.1/dist/js.cookie.min.js",
			"md5@js":                     jp + "/static/lib/md5/md5.min.js",
			"marked@js":                  jp + "/static/lib/marked/marked.min.js",
			"clipboard@js":               jp + "/static/lib/clipboard@2.0.8/clipboard.min.js",
			"mdui@index@js":              jp + "/static/js/mdui.index.js",
			"mdui@index@css":             jp + "/static/css/index.css",
			"sortablejs@js":              jp + "/static/lib/sortablejs@1.14.0/Sortable.min.js",
			"admin@js":                   jp + "/static/js/admin.js",
			"fontawesome@css":            jp + "/static/lib/fontawesome@5.15.4/css/all.min.css",
			"APlayer@css":                jp + "/static/lib/aplayer@1.10.1/dist/APlayer.min.css",
			"APlayer@js":                 jp + "/static/lib/aplayer@1.10.1/dist/APlayer.min.js",
			"sweetalert2@css":            jp + "/static/lib/sweetalert2@11.3.0/dist/sweetalert2.min.css",
			"sweetalert2@js":             jp + "/static/lib/sweetalert2@11.3.0/dist/sweetalert2.min.js",
			"hls@js":                     jp + "/static/lib/hls.js@1.1.2/dist/hls.min.js",
			"flv@js":                     jp + "/static/lib/flv.js@1.6.2/dist/flv.min.js",
			"artplayer@js":               jp + "/static/lib/artplayer/artplayer.js",
			"artplayer-danmuku@js":       jp + "/static/lib/artplayer/artplayer-plugin-danmuku.js",
			"video@mdui@js":              jp + "/static/js/mdui.video.js",
			"video@simple@js":            jp + "/static/js/simple.video.js",
			"simple@index@js":            jp + "/static/js/simple.index.js",
			"highlightjs@atom@dark@css":  jp + "/static/lib/highlightjs/cdn-release@11.4.0/build/styles/atom-one-dark.min.css",
			"highlightjs@atom@light@css": jp + "/static/lib/highlightjs/cdn-release@11.4.0/build/styles/atom-one-light.min.css",
			"highlight@js":               jp + "/static/lib/highlightjs/cdn-release@11.4.0/build/highlight.min.js",
			"jszip@js":                   jp + "/static/lib/jszip@3.1.5/jszip.min.js",
			"epub@js":                    jp + "/static/lib/epubjs@0.3.88/epub.min.js",
			"pdfh5@css":                  jp + "/static/lib/pdfh5@1.4.0/css/pdfh5.css",
			"pdf@js":                     jp + "/static/lib/pdfh5@1.4.0/js/pdf.js",
			"pdf@worker@js":              jp + "/static/lib/pdfh5@1.4.0/js/pdf.worker.js",
			"pdfh5@js":                   jp + "/static/lib/pdfh5@1.4.0/js/pdfh5.js",
			"natural@compare@js":         jp + "/static/lib/natural-compare-lite@1.4.0/index.min.js",
			"bootstrap@css":              jp + "/static/lib/bootstrap@4.6.1/css/bootstrap.min.css",
			"bootstrap@js":               jp + "/static/lib/bootstrap@4.6.1/js/bootstrap.min.js",
			"Material+Icons@css":         "//fonts.loli.net/icon?family=Material+Icons",
		},
	}
	cdnKV := cdnMap["0"].(KV)
	if cdn != "" {
		cdnKV = cdnMap[cdn].(KV)
	}
	for k, v := range cdnKV {
		m[k] = v.(string)
	}
	return m
}

func Base64Decode(str string) (string, bool) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", true
	}
	return string(data), false
}

// return ParentPath, fileName
func ParsePath(path string) (string, string) {
	filePath, fileName := filepath.Split(path)
	if filePath != "/" && filePath[len(filePath)-1:] == "/" {
		filePath = filePath[0 : len(filePath)-1]
	}
	return filePath, fileName
}

// clear Suffix
func ClearSuffix(filePath string) string {
	if filePath != "/" && filePath[len(filePath)-1:] == "/" {
		filePath = filePath[0 : len(filePath)-1]
	}
	return filePath
}

// return fileName
func GetFileName(path string) string {
	paths := strings.Split(path, "/")
	return paths[len(paths)-1]
}

func GetTransferDomain(transferConfig, domain string) string {
	domainGroups := strings.Split(transferConfig, ",")
	for _, dg := range domainGroups {
		domains := strings.Split(dg, "|")
		if domains[0] == domain {
			return domains[1]
		}
	}
	return domain
}

func ChunkBytes(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func RsaEncode(origData []byte, j_rsakey string) string {
	publicKey := []byte("-----BEGIN PUBLIC KEY-----\n" + j_rsakey + "\n-----END PUBLIC KEY-----")
	block, _ := pem.Decode(publicKey)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	pub := pubInterface.(*rsa.PublicKey)
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		log.Errorf("err: %s", err.Error())
	}
	return b64tohex(base64.StdEncoding.EncodeToString(b))
}

var b64map = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var BI_RM = "0123456789abcdefghijklmnopqrstuvwxyz"

func int2char(a int) string {
	return strings.Split(BI_RM, "")[a]
}

func b64tohex(a string) string {
	d := ""
	e := 0
	c := 0
	for i := 0; i < len(a); i++ {
		m := strings.Split(a, "")[i]
		if m != "=" {
			v := strings.Index(b64map, m)
			if 0 == e {
				e = 1
				d += int2char(v >> 2)
				c = 3 & v
			} else if 1 == e {
				e = 2
				d += int2char(c<<2 | v>>4)
				c = 15 & v
			} else if 2 == e {
				e = 3
				d += int2char(c)
				d += int2char(v >> 2)
				c = 3 & v
			} else {
				e = 0
				d += int2char(c<<2 | v>>4)
				d += int2char(15 & v)
			}
		}
	}
	if e == 1 {
		d += int2char(c << 2)
	}
	return d
}

// 获取随机数
func Random() string {
	return fmt.Sprintf("0.%17v", math_rand.New(math_rand.NewSource(time.Now().UnixNano())).Int63n(100000000000000000))
}

func GetRandomStr(length int) string {
	baseStr := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	r := math_rand.New(math_rand.NewSource(time.Now().UnixNano() + math_rand.Int63()))
	bytes := make([]byte, length, length)
	l := len(baseStr)
	for i := 0; i < length; i++ {
		bytes[i] = baseStr[r.Intn(l)]
	}
	return string(bytes)
}

func Yun139Sign(timestamp, key, json string) string {
	//去除多余空格
	json = strings.TrimSpace(json)
	json = EncodeURIComponent(json)
	c := strings.Split(json, "")
	sort.Strings(c)
	json = strings.Join(c, "")
	s1 := Md5(base64.StdEncoding.EncodeToString([]byte(json)))
	s2 := Md5(timestamp + ":" + key)
	return strings.ToUpper(Md5(s1 + s2))
}

func EncodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	r = strings.Replace(r, "%21", "!", -1)
	r = strings.Replace(r, "%27", "'", -1)
	r = strings.Replace(r, "%28", "(", -1)
	r = strings.Replace(r, "%29", ")", -1)
	r = strings.Replace(r, "%2A", "*", -1)
	return r
}

func GetClient(timeout int) *http.Client {
	if module.GloablConfig.Proxy != "" {
		proxy, _ := url.Parse(module.GloablConfig.Proxy)
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Second * time.Duration(timeout),
		}
		return client
	} else {
		return &http.Client{
			Timeout: time.Second * time.Duration(timeout),
		}
	}
}

func GetOffsetByRange(rangeStr string) uint64 {
	rangeStr = strings.Split(rangeStr, "=")[1]
	rangeStr = strings.Split(rangeStr, ",")[0]
	rangeStr = strings.Split(rangeStr, "-")[0]
	offset, err := strconv.ParseInt(rangeStr, 10, 64)
	if err != nil {
		return uint64(offset)
	}
	return 0
}

func GetMimeTypeByExt(ext string) string {
	mimeType := mime.DetectFileExt(ext)
	return mimeType
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func Md5Params(params map[string]string) string {
	keys := []string{}
	for k, v := range params {
		keys = append(keys, k+"="+v)
	}
	sort.Strings(keys)
	signStr := strings.Join(keys, "&")
	h := md5.New()
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}
func GetCurrentTheme(theme string) string {
	if strings.HasPrefix(theme, "mdui") {
		return "mdui"
	}
	return theme
}

func GetExpireTime(timeStr string, d time.Duration) string {
	t, err := time.Parse(timeStr, "2006-01-02 15:04:05")
	if err != nil {
		return ""
	}
	t1 := t.Add(d)
	return t1.Format("2006-01-02 15:04:05")
}

func ConfigToItem(obj interface{}) map[string]interface{} {
	json, _ := jsoniter.MarshalToString(obj)
	m := make(map[string]interface{})
	err := jsoniter.Unmarshal([]byte(json), &m)
	if err != nil {
		log.Error(err)
	}
	return m
}

func AccountToMap(account module.Account) map[string]interface{} {
	return map[string]interface{}{
		"id":               account.Id,
		"name":             account.Name,
		"mode":             account.Mode,
		"user":             account.User,
		"password":         account.Password,
		"refresh_token":    account.RefreshToken,
		"access_token":     account.RefreshToken,
		"redirect_uri":     account.RedirectUri,
		"api_url":          account.ApiUrl,
		"root_id":          account.RootId,
		"seq":              account.Seq,
		"files_count":      account.FilesCount,
		"status":           account.Status,
		"cookie_status":    account.CookieStatus,
		"time_span":        account.TimeSpan,
		"last_op_time":     account.LastOpTime,
		"sync_dir":         account.SyncDir,
		"sync_child":       account.SyncChild,
		"sync_cron":        account.SyncCron,
		"down_transfer":    account.DownTransfer,
		"cache_policy":     account.CachePolicy,
		"expire_time_span": account.ExpireTimeSpan,
		"host":             account.Host,
	}
}

func In(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	if index < len(str_array) && str_array[index] == target {
		return true
	}
	return false
}

func ParseFullPath(path, host string) (module.Account, string, string, string) {
	if strings.HasPrefix(path, "/d_") {
		//old path
		path = OldParseFullPath(path)
	}
	if path == "" {
		path = "/"
	}
	if strings.HasPrefix(path, module.GloablConfig.PathPrefix) {
		path = strings.TrimPrefix(path, module.GloablConfig.PathPrefix)
	}
	if path == "/" && module.GloablConfig.AccountChoose == "default" && len(module.GloablConfig.BypassList) > 0 {
		path = "/" + module.GloablConfig.BypassList[0].Name
	} else {
		if path == "/" && module.GloablConfig.AccountChoose == "default" && len(module.GloablConfig.Accounts) > 0 {
			path = "/" + module.GloablConfig.Accounts[0].Name
		}
	}
	fullPath := path
	if path != "/" && path[len(path)-1:] == "/" {
		path = path[0 : len(path)-1]
	}
	paths := strings.Split(path, "/")
	accountName := paths[1]
	account, bypassName := GetCurrentAccount(accountName, host)
	path = strings.Join(paths[2:], "/")
	path = "/" + path
	if fullPath != "/" && fullPath[len(fullPath)-1:] == "/" {
		fullPath = fullPath[0 : len(fullPath)-1]
	}
	return account, fullPath, path, bypassName
}

func OldParseFullPath(path string) string {
	iStr := GetBetweenStr(path, "_", "/")
	index, _ := strconv.Atoi(iStr)
	account := module.GloablConfig.Accounts[index]
	return strings.ReplaceAll(path, "/d_"+iStr, "/"+account.Name)
}

func GetCurrentAccount(name, host string) (module.Account, string) {
	//get account from bypass
	var bypass module.Bypass
	if len(module.GloablConfig.BypassList) > 0 {
		if name == "" {
			bypass = module.GloablConfig.BypassList[0]
		} else {
			for _, item := range module.GloablConfig.BypassList {
				if item.Name == name {
					bypass = item
					break
				}
			}
		}
		if bypass.Name != "" {
			//get round robin account
			bypassAccount := bypass.Rw.Next().(module.Account)
			log.Debugf("access bypass account: %s", bypassAccount.Name)
			return bypassAccount, bypass.Name
		}
	}
	//get account from accounts
	var account module.Account
	if name == "" {
		if len(module.GloablConfig.Accounts) > 0 {
			account = module.GloablConfig.Accounts[0]
		} else {
			account = module.Account{}
		}
	} else {
		for _, ac := range module.GloablConfig.Accounts {
			if ac.Name == name {
				if ac.Host != "" && host != "" {
					if ac.Host == host {
						account = ac
						break
					}
				} else {
					account = ac
					break
				}
			}
		}
	}
	return account, bypass.Name
}

func RandomPassword(length int) string {
	baseStr := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()[]{}+-*/_=."
	r := math_rand.New(math_rand.NewSource(time.Now().UnixNano() + math_rand.Int63()))
	bytes := make([]byte, length, length)
	l := len(baseStr)
	for i := 0; i < length; i++ {
		bytes[i] = baseStr[r.Intn(l)]
	}
	return string(bytes)
}

func Base(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func ExeFilePath(def string) string {
	if def != "" {
		return def
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func Group(nodes []module.FileNode, subGroupLength int64) [][]module.FileNode {
	max := int64(len(nodes))
	var segmens = make([][]module.FileNode, 0)
	quantity := max / subGroupLength
	remainder := max % subGroupLength
	i := int64(0)
	for i = int64(0); i < quantity; i++ {
		segmens = append(segmens, nodes[i*subGroupLength:(i+1)*subGroupLength])
	}
	if quantity == 0 || remainder != 0 {
		segmens = append(segmens, nodes[i*subGroupLength:i*subGroupLength+remainder])
	}
	return segmens
}
