package pan

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/libsgh/PanIndex/module"
)

var (
	ErrNotImplement = errors.New("not implement")
	LoginCaptcha    = errors.New("captcha")
	FlowLimit       = errors.New("flow limit")
)

type Pan interface {
	//获取授权（cookie、token获取）
	AuthLogin(account *module.Account) (string, error)
	//登录状态是否有效
	IsLogin(account *module.Account) bool
	//文件列表
	Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error)
	//获取文件信息
	File(account module.Account, fileId, path string) (module.FileNode, error)
	//上传文件 (多)
	UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error)
	//重命名
	Rename(account module.Account, fileId, name string) (bool, interface{}, error)
	//删除
	Remove(account module.Account, fileId string) (bool, interface{}, error)
	//创建文件夹
	Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error)
	//移动文件(夹)
	Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error)
	//复制文件(夹)
	Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error)
	//获取文件下载直链
	GetDownloadUrl(account module.Account, fileId string) (string, error)
	//获取文件文件path
	//GetPaths(account module.Account, fileId string) ([]module.FileNode, error)
	//获取网盘空间大小
	GetSpaceSzie(account module.Account) (int64, int64)
}

var PanMap = map[string]Pan{}
var client = resty.New()

type KV map[string]interface{}

func RegisterPan(mode string, pan Pan) {
	PanMap[mode] = pan
}

func GetPan(mode string) (pan Pan, ok bool) {
	pan, ok = PanMap[mode]
	return
}

func RemoveLoginAccount(id string) (ok bool) {
	delete(Alis, id)
	delete(OneDrives, id)
	delete(TeambitionSessions, id)
	delete(CLoud189s, id)
	delete(GoogleDrives, id)
	delete(S3s, id)
	delete(Pikpaks, id)
	return true
}
