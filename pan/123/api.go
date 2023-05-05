package _123

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var Sessions = map[string]LoginResp{}

var RootId = "0"

func init() {
	base.RegisterPan("123", &Pan123{})
}

type Pan123 struct{}

func (p Pan123) AuthLogin(account *module.Account) (string, error) {
	var resp LoginResp
	if account.RefreshToken != "" {
		isLogin := p.IsLogin(account)
		if isLogin {
			return Sessions[account.Id].Data.Token, nil
		}
	}
	reqBody := base.KV{"passport": account.User, "password": account.Password, "remember": true}
	//phone
	if strings.Contains(account.User, "@") {
		//mail
		reqBody = base.KV{"mail": account.User, "password": account.Password, "type": 2}
	}
	_, err := base.Client.R().
		SetResult(&resp).
		SetBody(reqBody).
		SetHeaders(map[string]string{
			"Origin":       "https://www.123pan.com",
			"Content-Type": "application/json;charset=UTF-8",
			"platform":     "web",
			"App-Version":  "3",
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
		}).
		Post("https://www.123pan.com/b/api/user/sign_in")
	//refresh_token_expire_time 1 month
	if err != nil || resp.Code != 200 {
		log.Errorln(err)
		return "", err
	}
	Sessions[account.Id] = resp
	return resp.Data.Token, nil
}

func (p Pan123) IsLogin(account *module.Account) bool {
	if _, ok := Sessions[account.Id]; !ok {
		resp := LoginResp{}
		resp.Data.Token = account.RefreshToken
		Sessions[account.Id] = resp
	}
	_, err := p.request(account, "https://www.123pan.com/a/api/user/info", http.MethodGet, nil, nil)
	if err == nil {
		return true
	}
	return false
}

func (p Pan123) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	var resp FilesResp
	fileNodes := make([]module.FileNode, 0)
	limit := 10
	page := 1
	for {
		_, err := p.request(&account, "https://www.123pan.com/b/api/file/list/new", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(map[string]string{
				"driveId":        "0",
				"limit":          fmt.Sprintf("%d", limit),
				"next":           "0",
				"orderBy":        "fileId",
				"orderDirection": "desc",
				"parentFileId":   fileId,
				"trashed":        "false",
				"SearchData":     "",
				"Page":           fmt.Sprintf("%d", page),
			})
		}, &resp)
		if err != nil || resp.Code != 0 {
			log.Errorln(err)
			return fileNodes, err
		}
		for _, f := range resp.Data.InfoList {
			fn, _ := p.ToFileNode(f)
			if path == "/" {
				fn.Path = path + fn.FileName
			} else {
				fn.Path = path + "/" + fn.FileName
			}
			fn.AccountId = account.Id
			fn.ParentId = fileId
			fn.ParentPath = path
			fileNodes = append(fileNodes, fn)
		}
		if resp.Data.Next == "-1" {
			break
		} else {
			page++
		}
	}
	return fileNodes, nil
}

func (p Pan123) File(account module.Account, fileId, path string) (module.FileNode, error) {
	var resp FileResp
	fn := module.FileNode{}
	if fileId == RootId {
		return module.FileNode{
			FileId:     "0",
			FileName:   "root",
			FileSize:   0,
			IsFolder:   true,
			Path:       "/",
			LastOpTime: time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	}
	_, err := p.request(&account, "https://www.123pan.com/b/api/file/info", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.KV{
			"fileIdList": []base.KV{
				{
					"fileId": fileId,
				},
			},
		})
	}, &resp)
	if err != nil || resp.Code != 0 {
		log.Errorln(err)
		return fn, err
	}
	fn, _ = p.ToFileNode(resp.Data.InfoList[0])
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	fn.AccountId = account.Id
	fn.ExtraData = map[string]interface{}{
		"etag":      resp.Data.InfoList[0].Etag,
		"s3keyFlag": resp.Data.InfoList[0].S3KeyFlag,
		"type":      resp.Data.InfoList[0].Type,
	}
	return fn, err
}

func (p Pan123) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	for _, file := range files {
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		t1 := time.Now()
		etag := fmt.Sprintf("%x", md5.Sum(file.Content))
		duplicate := 1
		if overwrite {
			duplicate = 2
		}
		var uploadRequestResp UploadRequestResp
		_, err := p.request(&account, "https://www.123pan.com/a/api/file/upload_request", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.KV{
				"driveId":      0,
				"duplicate":    duplicate,
				"etag":         etag,
				"fileName":     file.FileName,
				"parentFileId": parentFileId,
				"size":         file.FileSize,
				"type":         0,
			})
		}, &uploadRequestResp)
		if err != nil || uploadRequestResp.Code != 0 {
			log.Errorln(err)
			return false, uploadRequestResp, err
		}
		if uploadRequestResp.Data.Reuse {
			//file reuse success
			log.Debugf("file：%s，reuse success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
		} else {
			fileParts := ReadBlock(5242880, file)
			for _, fp := range fileParts {
				var repareResp RepareResp
				_, err = p.request(&account, "https://www.123pan.com/a/api/file/s3_repare_upload_parts_batch", http.MethodPost, func(req *resty.Request) {
					req.SetBody(base.KV{
						"bucket":          uploadRequestResp.Data.Bucket,
						"key":             uploadRequestResp.Data.Key,
						"partNumberEnd":   fp.partNumberEnd,
						"partNumberStart": fp.partNumberStart,
						"uploadId":        uploadRequestResp.Data.UploadID,
					})
				}, &repareResp)
				if err != nil || repareResp.Code != 0 {
					return false, repareResp, err
				}
				uploadUrl := repareResp.Data.PresignedUrls[fp.partNumberStart]
				r, _ := http.NewRequest(http.MethodPut, uploadUrl, bytes.NewReader(fp.Content))
				r.Header.Add("Content-Length", fmt.Sprintf("%d", len(fp.Content)))
				r.Header.Add("Origin", "https://www.123pan.com")
				r.Header.Add("Host", "file.123pan.com")
				res, _ := http.DefaultClient.Do(r)
				defer res.Body.Close()
			}
			var listUploadResp ListUploadResp
			_, err = p.request(&account, "https://www.123pan.com/a/api/file/s3_list_upload_parts", http.MethodPost, func(req *resty.Request) {
				req.SetBody(base.KV{
					"bucket":   uploadRequestResp.Data.Bucket,
					"key":      uploadRequestResp.Data.Key,
					"uploadId": uploadRequestResp.Data.UploadID,
				})
			}, &listUploadResp)
			if err != nil || listUploadResp.Code != 0 {
				return false, listUploadResp, err
			}
			if len(listUploadResp.Data.Parts) == len(fileParts) {
				//all parts upload success
				//upload complete
				var commonResp CommonResp
				_, err = p.request(&account, "https://www.123pan.com/a/api/file/s3_complete_multipart_upload", http.MethodPost, func(req *resty.Request) {
					req.SetBody(base.KV{
						"bucket":   uploadRequestResp.Data.Bucket,
						"key":      uploadRequestResp.Data.Key,
						"uploadId": uploadRequestResp.Data.UploadID,
					})
				}, &commonResp)
				if err == nil && commonResp.Code == 0 {
					_, err = p.request(&account, "https://www.123pan.com/a/api/file/upload_complete", http.MethodPost, func(req *resty.Request) {
						req.SetBody(base.KV{
							"fileId": uploadRequestResp.Data.FileID,
						})
					}, &commonResp)
					log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
				} else {
					return false, commonResp, err
				}
			}
		}
	}
	return true, "all file upload success", nil
}

func (p Pan123) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	var commonResp CommonResp
	_, err := p.request(&account, "https://www.123pan.com/a/api/file/rename", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.KV{
			"driveId":   0,
			"fileId":    fileId,
			"fileName":  name,
			"duplicate": 1,
		})
	}, &commonResp)
	if err != nil || commonResp.Code != 0 {
		log.Errorln(err)
		return false, commonResp, err
	}
	return true, commonResp, err
}

func (p Pan123) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	var trashResp TrashResp
	_, err := p.request(&account, "https://www.123pan.com/a/api/file/trash", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.KV{
			"driveId": 0,
			"fileTrashInfoList": []base.KV{
				{"fileId": fileId},
			},
			"operation": true,
		})
	}, &trashResp)
	if err != nil || trashResp.Code != 0 {
		log.Errorln(err)
		return false, trashResp, err
	}
	return true, trashResp, err
}

func (p Pan123) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	var commonResp CommonResp
	_, err := p.request(&account, "https://www.123pan.com/a/api/file/upload_request", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.KV{
			"driveId":      0,
			"NotReuse":     true,
			"duplicate":    1,
			"etag":         "",
			"fileName":     name,
			"parentFileId": parentFileId,
			"size":         0,
			"type":         1,
		})
	}, &commonResp)
	if err != nil || commonResp.Code != 0 {
		log.Errorln(err)
		return false, commonResp, err
	}
	return true, commonResp, err
}

func (p Pan123) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	var commonResp CommonResp
	_, err := p.request(&account, "https://www.123pan.com/a/api/file/mod_pid", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.KV{
			"parentFileId": targetFileId,
			"fileIdList": []base.KV{
				{"FileId": fileId},
			},
		})
	}, &commonResp)
	if err != nil || commonResp.Code != 0 {
		log.Errorln(err)
		return false, commonResp, err
	}
	return true, commonResp, err
}

// api not support
func (p Pan123) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	return false, nil, nil
}

func (p Pan123) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	fn, err := p.File(account, fileId, "")
	if err == nil {
		var downloadResp DownloadResp
		_, err = p.request(&account, "https://www.123pan.com/a/api/file/download_info", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.KV{
				"driveId":   0,
				"etag":      fn.ExtraData["etag"].(string),
				"fileId":    fileId,
				"fileName":  fn.FileName,
				"s3keyFlag": fn.ExtraData["s3keyFlag"].(string),
				"size":      fn.FileSize,
				"type":      fn.ExtraData["type"].(int),
			})
		}, &downloadResp)
		if err != nil || downloadResp.Code != 0 {
			log.Errorln(err)
			return "", err
		}
		downloadUrl := downloadResp.Data.DownloadURL
		u, err := url.Parse(downloadUrl)
		if err != nil {
			return "", err
		}
		nu := u.Query().Get("params")
		if nu != "" {
			du, _ := base64.StdEncoding.DecodeString(nu)
			u, err = url.Parse(string(du))
			if err != nil {
				return "", err
			}
		}
		var downloadUrlResp DownloadUrlResp
		_, err = p.request(&account, u.String(), http.MethodGet, nil, &downloadUrlResp)
		if downloadUrlResp.Code == 0 {
			return downloadUrlResp.Data.RedirectURL, err
		}
	}
	return "", err
}

func (p Pan123) GetSpaceSzie(account module.Account) (int64, int64) {
	var resp UserInfoResp
	_, err := p.request(&account, "https://www.123pan.com/a/api/user/info", http.MethodGet, nil, resp)
	if err != nil {
		return 0, 0
	}
	if err != nil {
		log.Errorln(err)
		return 0, 0
	}
	totalSize := resp.Data.SpacePermanent
	usedSize := resp.Data.SpaceUsed
	return totalSize, usedSize
}

func (p Pan123) ToFileNode(item Item) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = fmt.Sprintf("%d", item.FileId)
	fn.FileName = item.FileName
	fn.CreateTime = util.UTCTimeFormat(item.CreateAt)
	fn.LastOpTime = util.UTCTimeFormat(item.UpdateAt)
	fn.ParentId = fmt.Sprintf("%d", item.ParentFileId)
	fn.IsDelete = 1
	if item.Type == 0 {
		fn.IsFolder = false
		fn.FileType = util.GetExt(fn.FileName)
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = int64(item.Size)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
		fn.Thumbnail = item.DownloadUrl
		fn.DownloadUrl = item.DownloadUrl
	} else {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	}
	return fn, nil
}
