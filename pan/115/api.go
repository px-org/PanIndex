package _115

import (
	"bytes"
	"fmt"
	driver115 "github.com/SheltonZhu/115driver/pkg/driver"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

var Sessions = map[string]*driver115.Pan115Client{}

var UA = driver115.UADefalut

var RootId = "0"

func init() {
	base.RegisterPan("115", &Pan115{})
}

type Pan115 struct {
}

func (p Pan115) AuthLogin(account *module.Account) (string, error) {
	var err error
	client := driver115.Defalut()
	cr := &driver115.Credential{}
	if err = cr.FromCookie(account.Password); err != nil {
		return "", errors.New("invalid cookies")
	}
	client = client.ImportCredential(cr)
	Sessions[account.Id] = client
	return account.Password, client.LoginCheck()
}

func (p Pan115) IsLogin(account *module.Account) bool {
	if err := GetClient(account).LoginCheck(); err != nil {
		return true
	}
	return false
}

func (p Pan115) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	client := GetClient(&account)
	fileNodes := make([]module.FileNode, 0)
	limit := 3
	offset := 0
	for {
		files, err := client.ListPage(fileId, int64(offset), int64(limit))
		if err != nil {
			log.Errorf("[115]List file error: %s", err)
			return fileNodes, err
		}
		for _, f := range *files {
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
		if len(*files) == 0 {
			break
		} else {
			offset += limit
		}
	}
	return fileNodes, nil
}

func (p Pan115) File(account module.Account, fileId, path string) (module.FileNode, error) {
	client := GetClient(&account)
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
	file, err := client.GetFile(fileId)
	if err != nil {
		log.Errorf("[115]File error: %s", err)
		return fn, err
	}
	fn, _ = p.ToFileNode(*file)
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	fn.AccountId = account.Id
	fn.ExtraData = map[string]interface{}{
		"pickCode": file.PickCode,
	}
	return fn, err
}

func (p Pan115) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	client := GetClient(&account)
	for _, file := range files {
		var uploadResp UploadResp
		log.Debugf("Upload started：%s，Size：%d", file.FileName, file.FileSize)
		t1 := time.Now()
		resp, err := client.NewRequest().SetFormData(map[string]string{
			"filename": file.FileName,
			"filesize": fmt.Sprintf("%d", file.FileSize),
			"target":   "U_1_" + parentFileId,
		}).Post("https://uplb.115.com/3.0/sampleinitupload.php")
		err = jsoniter.Unmarshal(resp.Body(), &uploadResp)
		if err != nil {
			log.Errorf("[115]UploadFiles ,error:%s", err)
			return false, uploadResp, err
		}
		resp, err = client.NewRequest().SetFormData(map[string]string{
			"name":                  file.FileName,
			"key":                   uploadResp.Object,
			"policy":                uploadResp.Policy,
			"OSSAccessKeyId":        uploadResp.Accessid,
			"success_action_status": "200",
			"callback":              uploadResp.Callback,
			"signature":             uploadResp.Signature,
		}).SetFileReader("file", file.FileName, bytes.NewReader(file.Content)).Post(uploadResp.Host)
		log.Debugf("file：%s，upload success，timespan：%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "all file upload success", nil
}

func (p Pan115) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	client := GetClient(&account)
	err := client.Rename(fileId, name)
	if err != nil {
		log.Errorf("[115]Rename file, error:%s", err)
		return false, nil, err
	}
	return true, nil, err
}

func (p Pan115) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	client := GetClient(&account)
	err := client.Delete(fileId)
	if err != nil {
		log.Errorf("[115]Remove file, error:%s", err)
		return false, nil, err
	}
	return true, nil, err
}

func (p Pan115) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	client := GetClient(&account)
	categoryID, err := client.Mkdir(parentFileId, name)
	if err != nil {
		log.Errorf("[115]Mkdir file, error:%s", err)
		return false, categoryID, err
	}
	return true, categoryID, err
}

func (p Pan115) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	client := GetClient(&account)
	err := client.Move(targetFileId, fileId)
	if err != nil {
		log.Errorf("[115]Move file, error:%s", err)
		return false, nil, err
	}
	return true, nil, err
}

func (p Pan115) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	client := GetClient(&account)
	err := client.Copy(targetFileId, fileId)
	if err != nil {
		log.Errorf("[115]Copy file, error:%s", err)
		return false, nil, err
	}
	return true, nil, err
}

func (p Pan115) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	fn, err := p.File(account, fileId, "")
	client := GetClient(&account)
	downloadInfo, err := client.Download(fn.ExtraData["pickCode"].(string))
	if err != nil {
		log.Errorf("[115]Get download url, error:%s", err)
		return "", err
	}
	return downloadInfo.Url.Url, err
}

func (p Pan115) GetSpaceSzie(account module.Account) (int64, int64) {
	client := GetClient(&account)
	resp, err := client.NewRequest().Get("https://webapi.115.com/files/index_info")
	if err != nil {
		log.Errorf("[115]get index_info ,error:%s", err)
		return 0, 0
	}
	totalSize := jsoniter.Get(resp.Body(), "data").Get("space_info").Get("all_total").Get("size").ToInt64()
	usedSize := jsoniter.Get(resp.Body(), "data").Get("space_info").Get("all_use").Get("size").ToInt64()
	return totalSize, usedSize
}

func (p Pan115) ToFileNode(item driver115.File) (module.FileNode, error) {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = item.GetID()
	fn.FileName = item.GetName()
	fn.CreateTime = util.UTCTime(item.CreateTime)
	fn.LastOpTime = util.UTCTime(item.UpdateTime)
	fn.ParentId = fmt.Sprintf("%d", item.ParentID)
	fn.IsDelete = 1
	if !item.IsDir() {
		fn.IsFolder = false
		fn.FileType = util.GetExt(fn.FileName)
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = int64(item.Size)
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	} else {
		fn.IsFolder = true
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	}
	return fn, nil
}
