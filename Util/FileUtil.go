package Util

import (
	"PanIndex/config"
	"PanIndex/entity"
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/bluele/gcache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
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

func GetMimeType(fileName string) int {
	mime := strings.Split(mime.TypeByExtension(filepath.Ext(fileName)), "/")[0]
	ext := filepath.Ext(fileName)
	if mime == "image" {
		return 1
	} else if mime == "audio" {
		return 2
	} else if mime == "video" {
		return 3
	} else if mime == "text" {
		if strings.Contains(".pdf", ext) {
			return 0
		}
		return 4
	} else {
		if strings.Contains(".yml,.properties,.conf.js,.txt,.py,.go,.css,.lua,.sh,.sql,.html,.json,.java,.jsp", ext) {
			return 4
		} else if strings.Contains(".mp4,.m4v,.mkv,.webm,.mov,.avi,.wmv,.mpg,.flv,.3gp,.m3u8,.ts", ext) {
			return 3
		} else if strings.Contains(".jpg,.png,.gif,.webp,.cr2,.tif,.bmp,.heif,.jxr,.psd,.ico,.dwg", ext) {
			return 1
		} else if strings.Contains(".mid,.mp3,.m4a,.ogg,.flac,.wav,.amr,.aac", ext) {
			return 2
		}
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
func ReadStringByUrl(account entity.Account, url, fileId string) string {
	content := ""
	//为了提高效率，从缓存查询
	value, _ := GC.Get(fileId)
	if value != nil {
		log.Debugf("从缓存中读取文本内容{%s}", fileId)
		return value.(string)
	}
	data := []byte("")
	if account.Mode == "webdav" {
		data = WebDavReadFileToBytes(account, fileId)
	} else if account.Mode == "ftp" {
		data = FtpReadFileToBytes(account, fileId)
	} else {
		resp, err := http.Get(url)
		if err != nil {
			log.Errorln(err)
			return content
		}
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorln(err)
			return content
		}
	}
	content = fmt.Sprintf("%s", data)
	GC.Set(fileId, content)
	return content
}
func FileSearch(rootPath, path, key string) []entity.FileNode {
	if path == "" {
		path = "/"
	}
	list := []entity.FileNode{}
	//列出文件夹相对路径
	fullPath := filepath.Join(rootPath, path)
	if FileExist(fullPath) {
		//是目录
		// 读取该文件夹下所有文件
		fileInfos, err := ioutil.ReadDir(fullPath)
		if err != nil {
			panic(err.Error())
		} else {
			for _, fileInfo := range fileInfos {
				fileId := filepath.Join(fullPath, fileInfo.Name())
				// 按照文件名过滤
				if !fileInfo.IsDir() && !strings.Contains(fileInfo.Name(), key) {
					continue
				}
				// 当前文件是隐藏文件(以.开头)则不显示
				if IsHiddenFile(fileInfo.Name()) {
					continue
				}
				//指定隐藏的文件或目录过滤
				if config.GloablConfig.HideFileId != "" {
					listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fileId)
					if i < len(listSTring) && listSTring[i] == fileId {
						continue
					}
				}
				if config.GloablConfig.PwdDirId != "" {
					listSTring := strings.Split(config.GloablConfig.PwdDirId, ",")
					hide := false
					for _, v := range listSTring {
						f1 := strings.Split(v, ":")[0]
						if f1 == fileId {
							hide = true
							break
						}
					}
					if hide {
						continue
					}
				}
				fileType := GetMimeType(fileInfo.Name())
				// 实例化FileNode
				file := entity.FileNode{
					FileId:     fileId,
					IsFolder:   fileInfo.IsDir(),
					FileName:   fileInfo.Name(),
					FileSize:   int64(fileInfo.Size()),
					SizeFmt:    FormatFileSize(int64(fileInfo.Size())),
					FileType:   strings.TrimLeft(filepath.Ext(fileInfo.Name()), "."),
					Path:       PathJoin(path, fileInfo.Name()),
					MediaType:  fileType,
					LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
				}
				if fileInfo.IsDir() {
					childList := FileSearch(rootPath, file.Path, key)
					if len(childList) == 0 && !strings.Contains(fileInfo.Name(), key) {
						continue
					}
					for _, fn := range childList {
						list = append(list, fn)
					}

				}
				// 添加到切片中等待json序列化
				if fileInfo.IsDir() && !strings.Contains(fileInfo.Name(), key) {
					continue
				}
				list = append(list, file)

			}
		}
	}
	return list
}
func FileQuery(rootPath, path string, mt int) []entity.FileNode {
	if path == "" {
		path = "/"
	}
	list := []entity.FileNode{}
	//列出文件夹相对路径
	fullPath := filepath.Join(rootPath, path)
	if FileExist(fullPath) {
		//是目录
		// 读取该文件夹下所有文件
		fileInfos, err := ioutil.ReadDir(fullPath)
		if err != nil {
			panic(err.Error())
		} else {
			for _, fileInfo := range fileInfos {
				fileId := filepath.Join(fullPath, fileInfo.Name())
				// 当前文件是隐藏文件(以.开头)则不显示
				if IsHiddenFile(fileInfo.Name()) {
					continue
				}
				//指定隐藏的文件或目录过滤
				if config.GloablConfig.HideFileId != "" {
					listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
					sort.Strings(listSTring)
					i := sort.SearchStrings(listSTring, fileId)
					if i < len(listSTring) && listSTring[i] == fileId {
						continue
					}
				}
				fileType := GetMimeType(fileInfo.Name())
				// 实例化FileNode
				file := entity.FileNode{
					FileId:     fileId,
					IsFolder:   fileInfo.IsDir(),
					FileName:   fileInfo.Name(),
					FileSize:   int64(fileInfo.Size()),
					SizeFmt:    FormatFileSize(int64(fileInfo.Size())),
					FileType:   strings.TrimLeft(filepath.Ext(fileInfo.Name()), "."),
					Path:       PathJoin(path, fileInfo.Name()),
					MediaType:  fileType,
					LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
				}
				if !fileInfo.IsDir() {
					if mt != -1 && file.MediaType == mt {
						list = append(list, file)
					} else if mt == -1 {
						list = append(list, file)
					} else {
						continue
					}
				}

			}
		}
	}
	return list
}
func Zip(dst, src string) (err error) {
	// 创建准备写入的文件
	fw, err := os.Create(dst)
	defer fw.Close()
	if err != nil {
		return err
	}

	// 通过 fw 来创建 zip.Write
	zw := zip.NewWriter(fw)
	defer func() {
		// 检测一下是否成功关闭
		if err := zw.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// 下面来将文件写入 zw ，因为有可能会有很多个目录及文件，所以递归处理
	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return
		}

		// 替换文件信息中的文件名
		fh.Name = strings.TrimPrefix(path, filepath.Dir(src)+string(filepath.Separator))
		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}
		// 写入文件信息，并返回一个 Write 结构
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}

		// 打开要压缩的文件
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return
		}

		// 将打开的文件 Copy 到 w
		n, err := io.Copy(w, fr)
		if err != nil {
			return
		}
		// 输出压缩的内容
		log.Debugf("成功压缩文件： %s, 共写入了 %d 个字符的数据\n", path, n)
		return nil
	})
}
func PathJoin(path, fileName string) string {
	if path == "/" {
		return fmt.Sprintf("%s%s", path, fileName)
	} else {
		return fmt.Sprintf("%s/%s", path, fileName)
	}
}
func TransformText(f *os.File) ([]byte, string) {
	content, _ := ioutil.ReadAll(f)
	contentType := http.DetectContentType(content)
	if !utf8.Valid(content) {
		rr := bytes.NewReader(content)
		r := transform.NewReader(rr, simplifiedchinese.GBK.NewDecoder())
		b, _ := ioutil.ReadAll(r)
		return b, contentType
	}
	return content, contentType
}
func TransformTextFromBytes(content []byte) ([]byte, string) {
	contentType := http.DetectContentType(content)
	if !utf8.Valid(content) {
		rr := bytes.NewReader(content)
		r := transform.NewReader(rr, simplifiedchinese.GBK.NewDecoder())
		b, _ := ioutil.ReadAll(r)
		return b, contentType
	}
	return content, contentType
}
func TransformByte(reader io.ReadCloser) ([]byte, string) {
	content := StreamToByte(reader)
	contentType := http.DetectContentType(content)
	if !utf8.Valid(content) {
		rr := bytes.NewReader(content)
		r := transform.NewReader(rr, simplifiedchinese.GBK.NewDecoder())
		b, _ := ioutil.ReadAll(r)
		return b, contentType
	}
	return content, contentType
}

func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}
