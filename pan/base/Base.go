package base

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/px-org/PanIndex/module"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	ErrNotImplement = errors.New("not implement")
	LoginCaptcha    = errors.New("captcha")
	FlowLimit       = errors.New("flow limit")
)

type Pan interface {
	// AuthLogin 获取授权（cookie、token获取）
	AuthLogin(account *module.Account) (string, error)
	// IsLogin 登录状态是否有效
	IsLogin(account *module.Account) bool
	// Files 文件列表
	Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error)
	// File 获取文件信息
	File(account module.Account, fileId, path string) (module.FileNode, error)
	// UploadFiles 上传文件 (多)
	UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error)
	// Rename 重命名
	Rename(account module.Account, fileId, name string) (bool, interface{}, error)
	// Remove 删除
	Remove(account module.Account, fileId string) (bool, interface{}, error)
	// Mkdir 创建文件夹
	Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error)
	// Move 移动文件(夹)
	Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error)
	// Copy 复制文件(夹)
	Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error)
	// GetDownloadUrl 获取文件下载直链
	GetDownloadUrl(account module.Account, fileId string) (string, error)
	// GetSpaceSzie 获取网盘空间大小
	GetSpaceSzie(account module.Account) (int64, int64)
}

var PanMap = map[string]Pan{}
var Client = resty.New()

type KV map[string]interface{}

func RegisterPan(mode string, pan Pan) {
	PanMap[mode] = pan
}

func GetPan(mode string) (pan Pan, ok bool) {
	pan, ok = PanMap[mode]
	return
}

type Callback func(req *resty.Request)

func SimpleTest() {
	account := &module.Account{}
	account.Id = uuid.NewV4().String()
	account.User = os.Getenv("ACCOUNT_USER")
	account.Password = os.Getenv("ACCOUNT_PASSWORD")
	account.Mode = os.Getenv("MODE")
	account.RootId = os.Getenv("ROOT_ID")
	account.SiteId = os.Getenv("SITE_ID")
	account.RefreshToken = os.Getenv("REFRESH_TOKEN")
	p, _ := GetPan(account.Mode)
	result, err := p.AuthLogin(account)
	//p.IsLogin(account)
	log.Info(result, err)
	fs, _ := p.Files(*account, account.RootId, "/", "", "")
	log.Info(fs)
	f, _ := p.File(*account, fs[0].FileId, fs[0].Path)
	log.Info(f)
	log.Info(p.GetDownloadUrl(*account, f.FileId))
}
