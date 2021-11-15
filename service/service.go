package service

import (
	"PanIndex/Util"
	"PanIndex/config"
	"PanIndex/entity"
	"PanIndex/jobs"
	"PanIndex/model"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/nic"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var UrlCache = gcache.New(500).LRU().Build()

func GetFilesByPath(account entity.Account, path, pwd, sColumn, sOrder string, isView bool) map[string]interface{} {
	if path == "" {
		path = "/"
	}
	result := make(map[string]interface{})
	list := []entity.FileNode{}
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	result["HasReadme"] = false
	result["HasHead"] = false
	if account.Mode == "native" {
		//列出文件夹相对路径
		rootPath := account.RootId
		fullPath := filepath.Join(rootPath, path)
		if Util.FileExist(fullPath) {
			if Util.IsDirectory(fullPath) {
				//是目录
				// 读取该文件夹下所有文件
				fileInfos, err := ioutil.ReadDir(fullPath)
				//默认按照目录，时间倒序排列
				sort.Slice(fileInfos, func(i, j int) bool {
					d1 := 0
					if fileInfos[i].IsDir() {
						d1 = 1
					}
					d2 := 0
					if fileInfos[j].IsDir() {
						d2 = 1
					}
					if d1 > d2 {
						return true
					} else if d1 == d2 {
						if sColumn == "file_name" {
							c := strings.Compare(fileInfos[i].Name(), fileInfos[j].Name())
							if sOrder == "desc" {
								return c >= 0
							} else {
								return c <= 0
							}
						} else if sColumn == "file_size" {
							if sOrder == "desc" {
								return fileInfos[i].Size() >= fileInfos[j].Size()
							} else {
								return fileInfos[i].Size() <= fileInfos[j].Size()
							}
						} else if sColumn == "last_op_time" {
							if sOrder == "desc" {
								return fileInfos[i].ModTime().After(fileInfos[j].ModTime())
							} else {
								return fileInfos[i].ModTime().Before(fileInfos[j].ModTime())
							}
						} else {
							return fileInfos[i].ModTime().After(fileInfos[j].ModTime())
						}
					} else {
						return false
					}
				})
				if err != nil {
					panic(err.Error())
				} else {
					for _, fileInfo := range fileInfos {
						fileId := filepath.Join(fullPath, fileInfo.Name())
						// 当前文件是隐藏文件(以.开头)则不显示
						if Util.IsHiddenFile(fileInfo.Name()) {
							continue
						}
						if fileInfo.Name() == "README.md" {
							result["HasReadme"] = true
							result["ReadmeContent"] = Util.ReadStringByFile(fileId)
						}
						if fileInfo.Name() == "HEAD.md" {
							result["HasHead"] = true
							result["HeadContent"] = Util.ReadStringByFile(fileId)
						}
						//指定隐藏的文件或目录过滤
						/*if config.GloablConfig.HideFileId != "" {
							listSTring := strings.Split(config.GloablConfig.HideFileId, ",")
							sort.Strings(listSTring)
							i := sort.SearchStrings(listSTring, fileId)
							if i < len(listSTring) && listSTring[i] == fileId {
								continue
							}
						}*/
						hide := Util.CheckHide(fileId, config.GloablConfig.HideFileId)
						if hide {
							continue
						}
						fileType := Util.GetMimeType(fileInfo.Name())
						// 实例化FileNode
						file := entity.FileNode{
							FileId:     fileId,
							IsFolder:   fileInfo.IsDir(),
							FileName:   fileInfo.Name(),
							FileSize:   int64(fileInfo.Size()),
							SizeFmt:    Util.FormatFileSize(int64(fileInfo.Size())),
							FileType:   strings.TrimLeft(filepath.Ext(fileInfo.Name()), "."),
							Path:       Util.PathJoin(path, fileInfo.Name()),
							MediaType:  fileType,
							LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
							ParentId:   fullPath,
						}
						filepath.Join()
						// 添加到切片中等待json序列化
						list = append(list, file)
					}
				}
				result["isFile"] = false
			} else {
				fileInfo, err := os.Stat(fullPath)
				if err != nil {
					panic(err.Error())
				} else {
					if isView {
						dir := filepath.Dir(fullPath)
						p := filepath.Dir(path)
						fileInfos, _ := ioutil.ReadDir(dir)
						fileInfos = Util.FilterFiles(fileInfos, dir, sColumn, sOrder)
						last := Util.GetNextOrPrevious(fileInfos, fileInfo, -1)
						next := Util.GetNextOrPrevious(fileInfos, fileInfo, 1)
						if last != nil {
							result["LastFile"] = filepath.Join(p, last.Name())
						} else {
							result["LastFile"] = nil
						}
						if next != nil {
							result["NextFile"] = filepath.Join(p, next.Name())
						} else {
							result["NextFile"] = nil
						}
					}
					fileType := Util.GetMimeType(fileInfo.Name())
					file := entity.FileNode{
						FileId:     fullPath,
						IsFolder:   fileInfo.IsDir(),
						FileName:   fileInfo.Name(),
						FileSize:   int64(fileInfo.Size()),
						SizeFmt:    Util.FormatFileSize(int64(fileInfo.Size())),
						FileType:   strings.TrimLeft(filepath.Ext(fileInfo.Name()), "."),
						Path:       path,
						MediaType:  fileType,
						LastOpTime: time.Unix(fileInfo.ModTime().Unix(), 0).Format("2006-01-02 15:04:05"),
						ParentId:   filepath.Dir(fullPath),
					}
					hide := Util.CheckHide(fullPath, config.GloablConfig.HideFileId)
					if !hide {
						// 添加到切片中等待json序列化
						list = append(list, file)
					}
				}
				result["isFile"] = true
			}
			result["HasPwd"] = true
			hasPwd, pwdOk, msg := Util.CheckPwd(config.GloablConfig.PwdDirId, fullPath, pwd)
			result["PwdErrorMsg"] = msg
			if !hasPwd || (hasPwd && pwdOk) {
				result["HasPwd"] = false
			}
		}
	} else {
		order_sql := ""
		nl_column := "cache_time"
		if sColumn != "default" && sColumn != "null" {
			order_sql = fmt.Sprintf(" ORDER BY is_folder desc, %s %s", sColumn, sOrder)
			nl_column = sColumn
		} else {
			order_sql = fmt.Sprintf(" ORDER BY is_folder desc")
		}
		model.SqliteDb.Raw("select * from file_node where parent_path=? and `delete`=0 and hide = 0 and account_id=? "+order_sql, path, account.Id).Find(&list)
		result["isFile"] = false
		if len(list) == 0 {
			result["isFile"] = true
			model.SqliteDb.Raw("select * from file_node where path = ? and is_folder = 0 and `delete`=0 and hide = 0 and account_id=? limit 1", path, account.Id).Find(&list)
			if len(list) == 1 {
				var param interface{}
				if sColumn == "file_name" {
					param = list[0].FileName
				} else if sColumn == "file_size" {
					param = list[0].FileSize
				} else if sColumn == "last_op_time" {
					param = list[0].LastOpTime
				} else {
					param = list[0].CacheTime
				}
				if isView {
					next := entity.FileNode{}
					model.SqliteDb.Raw(fmt.Sprintf("select * from file_node where parent_path=? and account_id=? and is_folder=0 and hide = 0  and %s >? order by %s asc limit 1", nl_column, nl_column),
						list[0].ParentPath, account.Id, param).First(&next)
					result["NextFile"] = next.Path
					last := entity.FileNode{}
					model.SqliteDb.Raw(fmt.Sprintf("select * from file_node where parent_path=? and account_id=? and is_folder=0  and hide = 0 and %s < ? order by %s desc limit 1", nl_column, nl_column),
						list[0].ParentPath, account.Id, param).First(&last)
					result["LastFile"] = last.Path
				}
			}
		} else {
			mfs := []entity.FileNode{}
			model.SqliteDb.Raw("select * from file_node where parent_path=? and (file_name='README.md' or file_name='HEAD.md') and `delete`=0 and account_id=?", path, account.Id).Find(&mfs)
			for _, mf := range mfs {
				if !mf.IsFolder && mf.FileName == "README.md" {
					result["HasReadme"] = true
					dl := DownLock{}
					dl.FileId = mf.FileId
					dl.L = new(sync.Mutex)
					result["ReadmeContent"] = Util.ReadStringByUrl(account, dl.GetDownlaodUrl(account, mf), mf.FileId)
				}
				if !mf.IsFolder && mf.FileName == "HEAD.md" {
					result["HasHead"] = true
					dl := DownLock{}
					dl.FileId = mf.FileId
					dl.L = new(sync.Mutex)
					result["HeadContent"] = Util.ReadStringByUrl(account, dl.GetDownlaodUrl(account, mf), mf.FileId)
				}
			}
		}
		result["HasPwd"] = true
		fId := ""
		fileNode := entity.FileNode{}
		r := model.SqliteDb.Raw("select * from file_node where path = ? and `delete`=0 and account_id = ? and hide = 0", path, account.Id).First(&fileNode)
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			ac := entity.Account{}
			acr := model.SqliteDb.Raw("select * from account where id = ?", account.Id).First(&ac)
			if !errors.Is(acr.Error, gorm.ErrRecordNotFound) {
				fId = ac.RootId
			}
		} else {
			fId = fileNode.FileId
		}
		if fId != "" {
			hasPwd, pwdOk, msg := Util.CheckPwd(config.GloablConfig.PwdDirId, fId, pwd)
			result["PwdErrorMsg"] = msg
			if !hasPwd || (hasPwd && pwdOk) {
				result["HasPwd"] = false
			}
		} else {
			result["HasPwd"] = false
		}
	}
	result["List"] = list
	result["Path"] = path
	if path == "/" {
		result["HasParent"] = false
	} else {
		result["HasParent"] = true
	}
	result["ParentPath"] = PetParentPath(path)
	if account.Mode == "native" || account.Mode == "aliyundrive" {
		result["SurportFolderDown"] = true
	} else {
		result["SurportFolderDown"] = false
	}
	return result
}

func SearchFilesByKey(key, sColumn, sOrder string) map[string]interface{} {
	result := make(map[string]interface{})
	acountIndex := make(map[string]interface{})
	if len(config.GloablConfig.Accounts) == 1 {
		acountIndex[config.GloablConfig.Accounts[0].Id] = ""
	} else {
		for i, account := range config.GloablConfig.Accounts {
			acountIndex[account.Id] = fmt.Sprintf("/d_%d", i)
		}
	}

	list := []entity.SearchNode{}
	accouts := []entity.Account{}
	defer func() {
		if p := recover(); p != nil {
			log.Errorln(p)
		}
	}()
	sql := `
		SELECT
			a.*
		FROM
			file_node a
			LEFT JOIN account b ON b.id = a.account_id
		WHERE
			a.file_name LIKE ?
			AND a.` + "`delete`" + `= 0 
			AND (a.hide = 0 or a.hide is null)
			AND a.has_pwd = 0
			AND b.mode != 'native'
			`
	model.SqliteDb.Raw(sql, "%"+key+"%").Find(&list)
	model.SqliteDb.Raw("select * from account where mode = ?", "native").Find(&accouts)
	if len(accouts) > 0 {
		for _, account := range accouts {
			nfs := Util.FileSearch(account.RootId, "", key)
			//model.SqliteDb.Raw(sql, account.Id).Find(&dx)
			for _, fs := range nfs {
				sn := entity.SearchNode{fs, account.Id}
				list = append(list, sn)
			}
		}
	}
	//排序
	sort.Slice(list, func(i, j int) bool {
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
			if sColumn == "file_name" {
				c := strings.Compare(list[i].FileName, list[j].FileName)
				if sOrder == "desc" {
					return c >= 0
				} else {
					return c <= 0
				}
			} else if sColumn == "file_size" {
				if sOrder == "desc" {
					return list[i].FileSize >= list[j].FileSize
				} else {
					return list[i].FileSize <= list[j].FileSize
				}
			} else if sColumn == "last_op_time" {
				if sOrder == "desc" {
					return li.After(lj)
				} else {
					return li.Before(lj)
				}
			} else {
				return li.After(lj)
			}
		} else {
			return false
		}
	})
	result["List"] = list
	result["Path"] = "/"
	result["HasParent"] = false
	result["ParentPath"] = PetParentPath("/")
	result["SurportFolderDown"] = false
	result["AcountIndex"] = acountIndex
	return result
}

type DownLock struct {
	FileId string
	L      *sync.Mutex
}

func (dl *DownLock) GetDownlaodUrl(account entity.Account, fileNode entity.FileNode) string {
	var downloadUrl = ""
	var err error
	dl.L.Lock()
	defer func() {
		if err == nil {
			dl.L.Unlock()
		}
	}()
	if UrlCache.Has(fileNode.FileId) {
		cachUrl, err := UrlCache.Get(fileNode.FileId)
		if err == nil {
			downloadUrl = cachUrl.(string)
			log.Debugf("从缓存中获取下载地址:%s", downloadUrl)
		}
	} else {
		if account.Mode == "cloud189" {
			downloadUrl = Util.GetDownlaodUrlNew(account.Id, fileNode.FileId)
		} else if account.Mode == "teambition" {
			if Util.TeambitionSessions[account.Id].IsPorject {
				downloadUrl = Util.GetTeambitionProDownUrl("www", account.Id, fileNode.FileId)
			} else {
				return Util.GetTeambitionDownUrl(account.Id, fileNode.FileId)
			}
		} else if account.Mode == "teambition-us" {
			if Util.TeambitionSessions[account.Id].IsPorject {
				downloadUrl = Util.GetTeambitionProDownUrl("us", account.Id, fileNode.FileId)
			} else {
				//国际版暂时没有个人文件
			}
		} else if account.Mode == "aliyundrive" {
			downloadUrl = Util.AliGetDownloadUrl(account.Id, fileNode.FileId)
		} else if account.Mode == "onedrive" {
			downloadUrl = Util.GetOneDriveDownloadUrl(account.Id, fileNode.FileId)
		} else if account.Mode == "onedrive-cn" {
			downloadUrl = Util.GetOneDriveDownloadUrl(account.Id, fileNode.FileId)
		} else if account.Mode == "native" {
		} else if account.Mode == "webdav" {
		} else if account.Mode == "ftp" {
		} else if account.Mode == "yun139" {
			downloadUrl = Util.GetYun139DownUrl(account.Id, fileNode.FileId)
		} else if account.Mode == "googledrive" {
			downloadUrl = Util.GDGetDownUrl(account.Id, fileNode.FileId)
		}
		if downloadUrl != "" {
			//阿里云盘15分钟
			//天翼云盘15分钟
			//onedrive > 15分钟
			if account.Mode == "aliyundrive" {
				UrlCache.SetWithExpire(fileNode.FileId, downloadUrl, time.Minute*230)
			} else {
				UrlCache.SetWithExpire(fileNode.FileId, downloadUrl, time.Minute*14)
			}
			log.Debugf("调用api获取下载地址:" + downloadUrl)
		}
	}
	return downloadUrl
}

func GetDownlaodMultiFiles(account entity.Account, fileId, ua string) string {
	downUrl := ""
	if account.Mode == "cloud189" {
		Util.GetDownlaodMultiFiles(account.Id, fileId)
	} else if account.Mode == "aliyundrive" {
		fileNode := entity.FileNode{}
		model.SqliteDb.Raw("select * from file_node where account_id = ? and file_id = ? limit 1", account.Id, fileId).Find(&fileNode)
		//fmt.Println(Util.AliGetDownloadUrl(account.Id, fileId))
		downUrl = Util.AliFolderDownload(account.Id, fileId, fileNode.FileName, ua)
	}
	return downUrl
}

func PetParentPath(p string) string {
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

//获取查询游标start
func GetPageStart(pageNo, pageSize int) int {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 0
	}
	return (pageNo - 1) * pageSize
}

func GetConfig() entity.Config {
	c := entity.Config{}
	cis := []entity.ConfigItem{}
	accounts := []entity.Account{}
	model.SqliteDb.Raw("select * from config_item where 1=1").Find(&cis)
	configMap := make(map[string]interface{})
	for _, ci := range cis {
		configMap[ci.K] = ci.V
	}
	configJson, _ := jsoniter.MarshalToString(configMap)
	jsoniter.Unmarshal([]byte(configJson), &c)
	model.SqliteDb.Raw("select * from account order by `seq` asc").Find(&accounts)
	c.Accounts = accounts
	config.GloablConfig = c
	return c
}

func TransformPwdDirs(pwdDirId string) []entity.PwdDirs {
	pwdDirs := []entity.PwdDirs{}
	if pwdDirId != "" {
		arr := strings.Split(pwdDirId, ",")
		for _, pwdDir := range arr {
			s := strings.Split(pwdDir, ":")
			pwdDirs = append(pwdDirs, entity.PwdDirs{
				s[0], s[1],
			})
		}
	}
	return pwdDirs
}

func TransformHideFiles(hideFileId string) []string {
	hideFiles := []string{}
	if hideFileId != "" {
		arr := strings.Split(hideFileId, ",")
		for _, hideFile := range arr {
			hideFiles = append(hideFiles, hideFile)
		}
	}
	return hideFiles
}

func SaveConfig(config map[string]interface{}) {
	if config["accounts"] == nil {
		//基本配置
		for key, value := range config {
			model.SqliteDb.Table("config_item").Where("k=?", key).Update("v", Util.Strval(value))
		}
		if config["hide_file_id"] != nil {
			hideFiles := config["hide_file_id"].(string)
			if hideFiles != "" {
				model.SqliteDb.Table("file_node").Where("1 = 1").Update("hide", 0)
				for _, hf := range strings.Split(hideFiles, ",") {
					model.SqliteDb.Table("file_node").Where("file_id = ?", hf).Update("hide", 1)
					go hideChildrenFiles(hf)
				}
			}
		}
		if config["pwd_dir_id"] != nil {
			pwdFiles := config["pwd_dir_id"].(string)
			if pwdFiles != "" {
				model.SqliteDb.Table("file_node").Where("1 = 1").Update("has_pwd", 0)
				for _, opf := range strings.Split(pwdFiles, ",") {
					pf := strings.Split(opf, ":")[0]
					model.SqliteDb.Table("file_node").
						Where("file_id = ?", pf).
						Updates(map[string]interface{}{"has_pwd": 1})
					go pwdChildrenFiles(pf)
				}
			}
		}
	} else {
		//账号信息
		for _, account := range config["accounts"].([]interface{}) {
			mode := account.(map[string]interface{})["Mode"]
			ID := ""
			if account.(map[string]interface{})["id"] != nil && account.(map[string]interface{})["id"] != "" {
				old := entity.Account{}
				model.SqliteDb.Table("account").Where("id = ?", account.(map[string]interface{})["id"]).First(&old)
				if account.(map[string]interface{})["password"] == "" {
					delete(account.(map[string]interface{}), "password")
				}
				//更新网盘账号
				model.SqliteDb.Table("account").Where("id = ?", account.(map[string]interface{})["id"]).Updates(account.(map[string]interface{}))
				if mode != old.Mode {
					delete(Util.CLoud189Sessions, old.Id)
					delete(Util.TeambitionSessions, old.Id)
					if mode == "cloud189" {
						Util.CLoud189Sessions[old.Id] = entity.Cloud189{}
					} else if mode == "teambition" {
						Util.TeambitionSessions[old.Id] = entity.Teambition{nic.Session{}, "", "", "", "", "", false}
					} else if mode == "teambition-us" {
						Util.TeambitionSessions[old.Id] = entity.Teambition{nic.Session{}, "", "", "", "", "", false}
					} else if mode == "onedrive" || mode == "onedrive-cn" {
						Util.OneDrives[old.Id] = entity.OneDriveAuthInfo{}
					} else if mode == "aliyundrive" {
						Util.Alis[old.Id] = entity.TokenResp{}
					} else if mode == "yun139" {
						Util.Yun139Credentials[old.Id] = entity.Yun139{}
					}
				}
				ID = old.Id
			} else {
				//添加网盘账号
				id := uuid.NewV4().String()
				ID = id
				account.(map[string]interface{})["id"] = id
				account.(map[string]interface{})["status"] = 1
				account.(map[string]interface{})["cookie_status"] = 1
				account.(map[string]interface{})["files_count"] = 0
				var seq int
				model.SqliteDb.Table("account").Raw("select seq from account where 1=1 order by seq desc").First(&seq)
				account.(map[string]interface{})["seq"] = seq + 1
				model.SqliteDb.Table("account").Create(account.(map[string]interface{}))
				if mode == "cloud189" {
					Util.CLoud189Sessions[id] = entity.Cloud189{}
				} else if mode == "teambition" {
					Util.TeambitionSessions[id] = entity.Teambition{nic.Session{}, "", "", "", "", "", false}
				} else if mode == "teambition-us" {
					Util.TeambitionSessions[id] = entity.Teambition{nic.Session{}, "", "", "", "", "", false}
				} else if mode == "onedrive" || mode == "onedrive-cn" {
					Util.OneDrives[id] = entity.OneDriveAuthInfo{}
				} else if mode == "aliyundrive" {
					Util.Alis[id] = entity.TokenResp{}
				} else if mode == "yun139" {
					Util.Yun139Credentials[id] = entity.Yun139{}
				} else if mode == "googledrive" {
					Util.GoogleDrives[id] = entity.GoogleDriveAuthInfo{}
				}
			}
			ac := entity.Account{}
			model.SqliteDb.Table("account").Where("id=?", ID).Take(&ac)
			ac.SyncDir = "/"
			ac.SyncChild = 0
			go jobs.SyncInit(ac)
			jobs.CacheCron.Stop()
			jobs.AutoCacheRun()
		}
	}
	go GetConfig()
	//其他（打码狗）
}
func DeleteAccount(ids []string) {
	for _, id := range ids {
		//删除账号对应节点数据
		model.SqliteDb.Where("account_id = ?", id).Delete(entity.FileNode{})
		//删除账号数据
		var a entity.Account
		var si entity.ShareInfo
		a.Id = id
		si.AccountId = id
		model.SqliteDb.Model(entity.Account{}).Delete(a)
		model.SqliteDb.Model(entity.ShareInfo{}).Delete(si)
		go GetConfig()
		delete(Util.CLoud189Sessions, id)
		delete(Util.TeambitionSessions, id)
		delete(Util.TeambitionSessions, id)
	}
}
func SortAccounts(ids []string) {
	for i, id := range ids {
		i++
		model.SqliteDb.Model(entity.Account{}).Where("id=?", id).Update("seq", i)
	}
	go GetConfig()
}
func GetAccount(id string) entity.Account {
	account := entity.Account{}
	model.SqliteDb.Where("id = ?", id).First(&account)
	return account
}
func SetDefaultAccount(id string) {
	accountMap := make(map[string]interface{})
	accountMap["default"] = 0
	model.SqliteDb.Model(entity.Account{}).Where("1=1").Updates(accountMap)
	accountMap["default"] = 1
	model.SqliteDb.Table("account").Where("id=?", id).Updates(accountMap)
	go GetConfig()
}
func EnvToConfig() {
	config := os.Getenv("PAN_INDEX_CONFIG")
	if config != "" {
		//从环境变量写入数据库
		c := make(map[string]interface{})
		jsoniter.UnmarshalFromString(config, &c)
		if os.Getenv("PORT") != "" {
			c["port"] = os.Getenv("PORT")
		}
		c["damagou"] = nil
		delete(c, "damagou")
		model.SqliteDb.Where("1 = 1").Delete(&entity.Account{})
		//model.SqliteDb.Where("1 = 1").Delete(&entity.FileNode{})
		for _, account := range c["accounts"].([]interface{}) {
			//添加网盘账号
			account.(map[string]interface{})["status"] = 1
			account.(map[string]interface{})["files_count"] = 0
			model.SqliteDb.Table("account").Create(account.(map[string]interface{}))
		}
		delete(c, "accounts")
		SaveConfig(c)
	}
}
func Upload(accountId, path string, c *gin.Context) string {
	form, _ := c.MultipartForm()
	files := form.File["uploadFile"]
	dbFile := entity.FileNode{}
	account := entity.Account{}
	result := model.SqliteDb.Raw("select * from account where id=?", accountId).Take(&account)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "指定的账号不存在"
	}
	if account.Mode == "native" {
		p := filepath.FromSlash(account.RootId + path)
		if !Util.FileExist(p) {
			return "指定的目录不存在"
		}
		//服务器本地模式
		for _, file := range files {
			log.Debugf("开始上传文件：%s，大小：%d", file.Filename, file.Size)
			t1 := time.Now()
			p = filepath.FromSlash(account.RootId + path + "/" + file.Filename)
			c.SaveUploadedFile(file, p)
			log.Debugf("文件：%s，上传成功，耗时：%s", file.Filename, Util.ShortDur(time.Now().Sub(t1)))
		}
		return "上传成功"
	} else {
		if path == "/" {
			result = model.SqliteDb.Raw("select * from file_node where parent_path=? and `delete`=0 and account_id=? limit 1", path, accountId).Take(&dbFile)
		} else {
			result = model.SqliteDb.Raw("select * from file_node where path=? and `delete`=0 and account_id=?", path, accountId).Take(&dbFile)
		}
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "指定的目录不存在"
		} else {
			fileId := dbFile.FileId
			if path == "/" {
				fileId = dbFile.ParentId
			}
			if account.Mode == "teambition" && !Util.TeambitionSessions[accountId].IsPorject {
				//teambition 个人文件上传
				Util.TeambitionUpload(accountId, fileId, files)
			} else if account.Mode == "teambition" && Util.TeambitionSessions[accountId].IsPorject {
				//teambition 项目文件上传
				Util.TeambitionProUpload("", accountId, fileId, files)
			} else if account.Mode == "teambition-us" && Util.TeambitionSessions[accountId].IsPorject {
				//teambition-us 项目文件上传
				Util.TeambitionProUpload("us", accountId, fileId, files)
			} else if account.Mode == "cloud189" {
				//天翼云盘文件上传
				Util.Cloud189UploadFiles(accountId, fileId, files)
			} else if account.Mode == "aliyundrive" {
				//阿里云盘文件上传
				Util.AliUpload(accountId, fileId, files)
			} else if account.Mode == "onedrive" {
				//微软云盘文件上传
				Util.OneDriveUpload(accountId, fileId, files)
			} else if account.Mode == "onedrive-cn" {
				//世纪互联文件上传
				Util.OneDriveUpload(accountId, fileId, files)
			} else if account.Mode == "ftp" {
				//FTP文件上传
				Util.FtpUpload(account, fileId, files)
			} else if account.Mode == "webdav" {
				//WebDav文件上传
				Util.WebDavUpload(account, fileId, files)
			} else if account.Mode == "yun139" {
				//和彩云上传文件
				Util.Yun139Upload(account.Id, fileId, files)
			} else if account.Mode == "googledrive" {
				//谷歌云上传文件
				Util.GDUpload(account.Id, fileId, files)
			}
			return "上传成功"
		}
	}
}

func Async(accountId, path string) string {
	account := entity.Account{}
	result := model.SqliteDb.Raw("select * from account where id=?", accountId).Take(&account)
	dbFile := entity.FileNode{}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "指定的账号不存在"
	}
	if account.Mode == "native" {
		return "无需刷新"
	} else {
		if path == "/" {
			result = model.SqliteDb.Raw("select * from file_node where parent_path=? and `delete`=0 and account_id=? limit 1", path, accountId).Take(&dbFile)
		} else {
			result = model.SqliteDb.Raw("select * from file_node where path=? and `delete`=0 and account_id=?", path, accountId).Take(&dbFile)
		}
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "指定的目录不存在"
		} else {
			fileId := dbFile.FileId
			if path == "/" {
				fileId = dbFile.ParentId
			}
			if account.Mode == "teambition" && !Util.TeambitionSessions[accountId].IsPorject {
				//teambition 个人文件
				Util.TeambitionGetFiles(account.Id, fileId, fileId, path, 0, 0, true)
			} else if account.Mode == "teambition" && Util.TeambitionSessions[accountId].IsPorject {
				//teambition 项目文件
				Util.TeambitionGetProjectFiles("www", account.Id, fileId, path, 0, 0, true)
			} else if account.Mode == "teambition-us" && Util.TeambitionSessions[accountId].IsPorject {
				//teambition-us 项目文件
				Util.TeambitionGetProjectFiles("us", account.Id, fileId, path, 0, 0, true)
			} else if account.Mode == "cloud189" {
				Util.Cloud189GetFiles(account.Id, fileId, fileId, path, 0, 0, true)
			} else if account.Mode == "aliyundrive" {
				Util.AliGetFiles(account.Id, fileId, fileId, path, 0, 0, true)
			} else if account.Mode == "onedrive" {
				Util.OndriveGetFiles("", account.Id, fileId, path, 0, 0, true)
			} else if account.Mode == "onedrive-cn" {
				Util.OndriveGetFiles("", account.Id, fileId, path, 0, 0, true)
			} else if account.Mode == "ftp" {
				Util.FtpGetFiles(account, fileId, path, 0, 0, true)
			} else if account.Mode == "webdav" {
				Util.WebDavGetFiles(account, fileId, path, 0, 0, true)
			} else if account.Mode == "yun139" {
				Util.Yun139GetFiles(account.Id, fileId, path, 0, 0, true)
			} else if account.Mode == "googledrive" {
				Util.GDGetFiles(account.Id, fileId, path, 0, 0, true)
			}
			jobs.RefreshFileNodes(account.Id, fileId)
			return "刷新成功"
		}
	}
}
func GetViewTemplate(fn entity.FileNode) string {
	t := ""
	if config.GloablConfig.EnablePreview == "0" {
		return t
	}
	if strings.Contains(config.GloablConfig.Image, fn.FileType) {
		t = "img"
	} else if strings.Contains(config.GloablConfig.Audio, fn.FileType) {
		t = "audio"
	} else if strings.Contains(config.GloablConfig.Video, fn.FileType) {
		t = "video"
	} else if fn.FileType == "pdf" {
		t = "pdf"
	} else if fn.FileType == "md" {
		t = "md"
	} else {
		if config.GloablConfig.Other == "*" {
			t = "ns"
		} else if strings.Contains(config.GloablConfig.Other, fn.FileType) {
			t = "ns"
		}
	}
	return t
}
func AccountsToNodes(accounts []entity.Account, pwd string) map[string]interface{} {
	result := make(map[string]interface{})
	result["HasReadme"] = true
	fns := []entity.FileNode{}
	for i, account := range accounts {
		fn := entity.FileNode{
			FileId:     fmt.Sprintf("/d_%d", i),
			IsFolder:   true,
			FileName:   account.Name,
			FileSize:   int64(account.FilesCount),
			SizeFmt:    "-",
			FileType:   "",
			Path:       fmt.Sprintf("/d_%d", i),
			MediaType:  0,
			LastOpTime: account.LastOpTime,
			ParentId:   "",
		}
		fns = append(fns, fn)
	}
	result["isFile"] = false
	result["HasPwd"] = true
	fId := "all"
	hasPwd, pwdOk, msg := Util.CheckPwd(config.GloablConfig.PwdDirId, fId, pwd)
	result["PwdErrorMsg"] = msg
	if !hasPwd || (hasPwd && pwdOk) {
		result["HasPwd"] = false
	}
	result["List"] = fns
	result["Path"] = "/"
	result["HasParent"] = false
	result["ParentPath"] = PetParentPath("/")
	result["SurportFolderDown"] = false
	return result
}

var dls = sync.Map{}

func GetFileData(account entity.Account, path string) ([]byte, string) {
	f := entity.FileNode{}
	if account.Mode == "native" {
		rootPath := account.RootId
		fullPath := filepath.Join(rootPath, path)
		f, err := os.Open(fullPath)
		if err != nil {
			log.Debug(err)
			return nil, "image/png"
		}
		fileInfo, err := os.Stat(fullPath)
		mt := Util.GetMimeType(fileInfo.Name())
		if mt == 4 {
			return Util.TransformText(f)
		} else {
			b, _ := ioutil.ReadAll(f)
			contentType := http.DetectContentType(b)
			return b, contentType
		}

	} else if account.Mode == "ftp" {
		fileId := filepath.Join(account.RootId, path)
		data := Util.FtpReadFileToBytes(account, fileId)
		if data == nil {
			return nil, "image/png"
		} else {
			mt := Util.GetMimeType(path)
			if mt == 4 {
				return Util.TransformTextFromBytes(data)
			} else {
				contentType := http.DetectContentType(data)
				return data, contentType
			}

		}
	} else if account.Mode == "webdav" {
		fileId := filepath.Join(account.RootId, path)
		data := Util.WebDavReadFileToBytes(account, fileId)
		if data == nil {
			return nil, "image/png"
		} else {
			mt := Util.GetMimeType(path)
			if mt == 4 {
				return Util.TransformTextFromBytes(data)
			} else {
				contentType := http.DetectContentType(data)
				return data, contentType
			}

		}
	} else {
		result := model.SqliteDb.Raw("select * from file_node where path = ? and is_folder = 0 and `delete`=0 and ((hide = 0) or (hide=1 and file_name='README.md')) and account_id=? limit 1", path, account.Id).First(&f)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, "image/png"
		}
		var dl = DownLock{}
		if _, ok := dls.Load(f.FileId); ok {
			ss, _ := dls.Load(f.FileId)
			dl = ss.(DownLock)
		} else {
			dl.FileId = f.FileId
			dl.L = new(sync.Mutex)
			dls.LoadOrStore(f.FileId, dl)
		}
		if f.FileName == "README.md" {
			readmeContent := Util.ReadStringByUrl(account, dl.GetDownlaodUrl(account, f), f.FileId)
			contentType := http.DetectContentType([]byte(readmeContent))
			return []byte(readmeContent), contentType
		}
		dUrl := dl.GetDownlaodUrl(account, f)
		resp, err := httpClient().Get(dUrl)
		if err != nil {
			log.Errorln(err)
		}
		defer resp.Body.Close()
		mt := Util.GetMimeType(f.FileName)
		if mt == 4 {
			return Util.TransformByte(resp.Body)
		} else {
			data, _ := ioutil.ReadAll(resp.Body)
			contentType := http.DetectContentType(data)
			return data, contentType
		}
	}
}

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func GetFiles(accountId, parentPath, sColumn, sOrder, mediaType string) []entity.FileNode {
	account := GetAccount(accountId)
	mt, _ := strconv.Atoi(mediaType)
	list := []entity.FileNode{}
	if account.Mode == "native" {
		//本地模式
		list = Util.FileQuery(account.RootId, parentPath, mt)
	} else {
		//其他网盘模式
		sql := `
		SELECT
			a.*,('/d_'||(select count(*) from account c where c.rowid < b.rowid )) as dx,b.id as account_id
		FROM
			file_node a
			LEFT JOIN account b ON b.id = a.account_id
		WHERE
			a.parent_path = ?
			AND a.` + "`delete`" + `= 0 
			AND a.hide = 0
			AND a.is_folder = 0
			AND a.account_id = ?
			`
		if mt != -1 {
			sql += " AND a.media_type = ?"
		}
		model.SqliteDb.Raw(sql, parentPath, accountId, mt).Find(&list)
	}
	//字段排序
	sort.Slice(list, func(i, j int) bool {
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
			if sColumn == "file_name" {
				c := strings.Compare(list[i].FileName, list[j].FileName)
				if sOrder == "desc" {
					return c >= 0
				} else {
					return c <= 0
				}
			} else if sColumn == "file_size" {
				if sOrder == "desc" {
					return list[i].FileSize >= list[j].FileSize
				} else {
					return list[i].FileSize <= list[j].FileSize
				}
			} else if sColumn == "last_op_time" {
				if sOrder == "desc" {
					return li.After(lj)
				} else {
					return li.Before(lj)
				}
			} else {
				return li.After(lj)
			}
		} else {
			return false
		}
	})
	return list
}
func ShortInfo(accountId, path, prefix string) (string, string, string) {
	si := entity.ShareInfo{}
	model.SqliteDb.Raw("select * from share_info where account_id = ? and file_path=?", accountId, path).First(&si)
	shortUrl := ""
	if accountId == "" || path == "" {
		return "", "", "无效的id"
	}
	shortCode := ""
	if si.ShortCode != "" {
		shortCode = si.ShortCode
	} else {
		shortCodes, err := Util.Transform(accountId + path)
		if err != nil {
			log.Errorln(err)
			return "", "", "短链生成失败"
		}
		shortCode = shortCodes[0]
		model.SqliteDb.Create(entity.ShareInfo{
			accountId, path, shortCode,
		})
	}
	shortUrl = prefix + shortCode
	png, err := qrcode.Encode(shortUrl, qrcode.Medium, 256)
	if err != nil {
		panic(err)
	}
	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(png))
	return shortUrl, dataURI, "短链生成成功"
}
func GetRedirectUri(shorCode string) string {
	redirectUri := "/"
	si := entity.ShareInfo{}
	result := model.SqliteDb.Raw("select * from share_info where short_code=?", shorCode).First(&si)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		ac := entity.Account{}
		result = model.SqliteDb.Raw("select * from account where id=?", si.AccountId).First(&ac)
		drive := "/d_0"
		for i, account := range config.GloablConfig.Accounts {
			if account.Id == ac.Id {
				drive = fmt.Sprintf("/d_%d", i)
			}
		}
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			redirectUri = drive + si.FilePath + "?v"
		}
	}
	return redirectUri
}

func hideChildrenFiles(fileId string) {
	files := []entity.FileNode{}
	model.SqliteDb.Table("file_node").Where("parent_id = ?", fileId).Find(&files)
	model.SqliteDb.Table("file_node").Where("parent_id = ?", fileId).Update("hide", 1)
	for _, file := range files {
		if file.IsFolder {
			hideChildrenFiles(file.FileId)
		}
	}
}

func pwdChildrenFiles(fileId string) {
	files := []entity.FileNode{}
	model.SqliteDb.Table("file_node").Where("parent_id = ?", fileId).Find(&files)
	model.SqliteDb.Table("file_node").
		Where("parent_id = ?", fileId).
		Updates(map[string]interface{}{"has_pwd": 1})
	for _, file := range files {
		if file.IsFolder {
			pwdChildrenFiles(file.FileId)
		}
	}
}

func AutoAddPassword(oldDirPwd string) string {
	oldDirPwds := strings.Split(oldDirPwd, ",")
	newDirPwds := []string{}
	for _, odp := range oldDirPwds {
		fileId := strings.Split(odp, ":")[0]
		pwd := strings.Split(odp, ":")[1]
		fileIds := []string{}
		GetAllFileIds(fileId, &fileIds)
		newDirPwds = append(newDirPwds, fileId+":"+pwd)
		if len(fileIds) > 0 {
			for _, npd := range fileIds {
				flag := false
				for _, odp2 := range oldDirPwds {
					if odp2 == npd {
						flag = true
					}
				}
				if !flag {
					newDirPwds = append(newDirPwds, npd+":"+pwd)
				}
			}
		}
	}
	return strings.Join(newDirPwds, ",")
}

func AutoRemovePassword(cookieDirPwd, fileId string) string {
	cookieDirPwds := strings.Split(cookieDirPwd, ",")
	newDirPwds := []string{}
	fileIds := []string{}
	fileIds = append(fileIds, fileId)
	GetAllFileIds(fileId, &fileIds)
	for _, odp := range cookieDirPwds {
		fId := strings.Split(odp, ":")[0]
		flag := true
		for _, f := range fileIds {
			if f == fId {
				flag = false
				break
			}
		}
		if flag {
			newDirPwds = append(newDirPwds, fId)
		}
	}
	return strings.Join(newDirPwds, ",")
}

func GetAllFileIds(fileId string, fileIds *[]string) {
	files := []entity.FileNode{}
	model.SqliteDb.Table("file_node").Where("parent_id = ?", fileId).Find(&files)
	for _, file := range files {
		*fileIds = append(*fileIds, file.FileId)
		if file.IsFolder {
			GetAllFileIds(file.FileId, fileIds)
		}
	}
}
