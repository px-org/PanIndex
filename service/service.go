package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/px-org/PanIndex/dao"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/pan/ftp"
	"github.com/px-org/PanIndex/pan/webdav"
	"github.com/px-org/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var UrlCache = gcache.New(100000).LRU().Build()
var FilesCache = gcache.New(100000).LRU().Build()
var FileCache = gcache.New(100000).LRU().Build()

func Index(ac module.Account, path, fullPath, sortColumn, sortOrder string, isView bool) ([]module.FileNode, bool, string, string) {
	fns := []module.FileNode{}
	isFile := false
	var err error
	if ac.CachePolicy == "nc" {
		fns, isFile, err = GetFilesFromApi(ac, path, fullPath, "default", "null")
		if err == nil {
			fns = FilterHideFiles(fns)
		} else {
			isFile = false
			fns = []module.FileNode{}
		}
	} else if ac.CachePolicy == "mc" {
		if FilesCache.Has(fullPath) {
			files, err := FilesCache.Get(fullPath)
			if err == nil {
				fcb := files.(FilesCacheBean)
				log.Infof("get file from cache:%s", fullPath)
				fns = fcb.FileNodes
				isFile = fcb.IsFile
			}
		} else {
			log.Infof("get file from api:%s", fullPath)
			fns, isFile, err = GetFilesFromApi(ac, path, fullPath, "default", "null")
			log.Infof("get file from api result:%d", len(fns))
			fns = FilterHideFiles(fns)
			cacheTime := time.Now().Format("2006-01-02 15:04:05")
			if err == nil {
				FilesCache.SetWithExpire(fullPath, FilesCacheBean{fns, isFile, cacheTime, util.GetExpireTime(cacheTime, time.Hour*time.Duration(ac.ExpireTimeSpan))}, time.Hour*time.Duration(ac.ExpireTimeSpan))
			}
		}
	} else if ac.CachePolicy == "dc" {
		fns, isFile = GetFilesFromDb(ac, fullPath, "default", "null")
		fns = FilterHideFiles(fns)
	}
	sortFns := make([]module.FileNode, len(fns))
	copy(sortFns, fns)
	util.SortFileNode(sortColumn, sortOrder, sortFns)
	var lastFile, nextFile = "", ""
	if isView && isFile {
		lastFile, nextFile = GetLastNextFile(ac, path, fullPath, "default", "null")
	}
	return sortFns, isFile, lastFile, nextFile
}

type FilesCacheBean struct {
	FileNodes  []module.FileNode
	IsFile     bool
	CacheTime  string
	ExpireTime string
}

type FileCacheBean struct {
	FileNode   module.FileNode
	CacheTime  string
	ExpireTime string
}

type DownUrlCacheBean struct {
	Url        string
	CacheTime  string
	ExpireTime string
}

func GetFilesFromApi(ac module.Account, path, fullPath, sortColumn, sortOrder string) ([]module.FileNode, bool, error) {
	var fns []module.FileNode
	p, _ := base.GetPan(ac.Mode)
	fileId := GetFileIdByPath(ac, path, fullPath)
	file, err := p.File(ac, fileId, fullPath)
	isFile := false
	if err != nil {
		log.Error(err)
	}
	if !file.IsFolder {
		//is file
		fns = append(fns, file)
		isFile = true
	} else {
		//is folder
		fns, err = p.Files(ac, fileId, fullPath, sortColumn, sortOrder)
		if err != nil {
			log.Error(err)
		}
	}
	return fns, isFile, err
}

func GetFilesFromDb(ac module.Account, path, sortColumn, sortOrder string) ([]module.FileNode, bool) {
	var fns []module.FileNode
	file, isFile := dao.FindFileByPath(ac, path)
	if isFile && file.FileId != "" && file.Id != "" {
		fns = append(fns, file)
	} else {
		isFile = false
		fns = dao.FindFileListByPath(ac, path, sortColumn, sortOrder)
	}
	return fns, isFile
}

func Search(searchKey string) []module.FileNode {
	var fns []module.FileNode
	//only support db cache mode
	byPassAccounts := []string{}
	dao.DB.Model(&module.BypassAccounts{}).
		Select("max(account_id)").
		Group("bypass_id").
		Find(&byPassAccounts)
	keys := strings.Split(searchKey, ":")
	if len(byPassAccounts) > 0 {
		sql := `select
				fn.*
			from
				file_node fn
			left join account a on
				fn.account_id = a.id
			where
				fn.file_name like ? and a.id in ?`
		dao.DB.Raw(sql, "%"+searchKey+"%", byPassAccounts).Find(&fns)
	} else {
		sql := `select
				fn.*
			from
				file_node fn
			left join account a on
				fn.account_id = a.id
			where
				fn.file_name like ? and fn.hide = 0`
		if len(keys) > 1 {
			searchKey = keys[1]
			sql += " and a.name=?"
			dao.DB.Raw(sql, "%"+searchKey+"%", keys[0]).Find(&fns)
		} else {
			dao.DB.Raw(sql, "%"+searchKey+"%").Find(&fns)
		}
	}
	return fns
}

func FilterHideFiles(files []module.FileNode) []module.FileNode {
	var fns []module.FileNode
	hideMap := dao.GetHideFilesMap()
	for _, file := range files {
		_, ok := hideMap[file.Path]
		if !ok {
			fns = append(fns, file)
		}
	}
	return fns
}

func FilterFilesByType(files []module.FileNode, viewType string) []module.FileNode {
	var fns []module.FileNode
	for _, file := range files {
		if viewType == "" || file.ViewType == viewType {
			fns = append(fns, file)
		}
	}
	return fns
}

func HasParent(path string) (bool, string) {
	hasParent := false
	parentPath := ""
	if path != "/" {
		hasParent = true
	}
	parentPath = util.GetParentPath(path)
	return hasParent, parentPath
}

func GetFileIdByPath(ac module.Account, path, fullPath string) string {
	fileId := ac.RootId
	if path == "/" || path == "" {
		return fileId
	}
	if ac.CachePolicy == "dc" {
		fileId = dao.GetFileIdByPath(ac.Id, fullPath)
		return fileId
	} else if ac.CachePolicy == "mc" {
		parentPath := util.GetParentPath(fullPath)
		if FilesCache.Has(parentPath) {
			files, err := FilesCache.Get(parentPath)
			if err == nil {
				fcb := files.(FilesCacheBean)
				if len(fcb.FileNodes) > 0 {
					for _, fn := range fcb.FileNodes {
						if fn.Path == fullPath {
							return fn.FileId
						}
					}
				}
			}
		}
	}
	paths := util.GetPrePath(path)
	for _, pathMap := range paths {
		fId, ok := LoopGetFileId(ac, fileId, pathMap["PathUrl"], path)
		fileId = fId
		if ok {
			break
		}
	}
	return fileId
}

func LoopGetFileId(ac module.Account, fileId, path, filePath string) (string, bool) {
	fileName := util.GetFileName(path)
	p, _ := base.GetPan(ac.Mode)
	fns, _ := p.Files(ac, fileId, util.GetParentPath(path), "", "")
	fId, fnPath := GetCurrentId(fileName, fns)
	if fId != "" {
		if fnPath == filePath {
			return fId, true
		}
		return fId, false
	} else {
		return "", false
	}
}

func GetFileIdFromApi(ac module.Account, path string) string {
	fileId := ac.RootId
	paths := util.GetPrePath(path)
	for _, pathMap := range paths {
		fId, ok := LoopGetFileId(ac, fileId, pathMap["PathUrl"], path)
		fileId = fId
		if ok {
			break
		}
	}
	return fileId
}

func GetCurrentId(pathName string, fns []module.FileNode) (string, string) {
	for _, fn := range fns {
		if fn.FileName == pathName {
			return fn.FileId, fn.Path
		}
	}
	return "", ""
}

type DownLock struct {
	FileId string
	L      *sync.Mutex
}

var dls = sync.Map{}

func GetDownloadUrl(ac module.Account, fileId string) string {
	if fileId == "" {
		return ""
	}
	var dl = DownLock{}
	if _, ok := dls.Load(fileId); ok {
		ss, _ := dls.Load(fileId)
		dl = ss.(DownLock)
	} else {
		dl.FileId = fileId
		dl.L = new(sync.Mutex)
		dls.LoadOrStore(fileId, dl)
	}
	downUrl := dl.GetDownlaodUrl(ac, fileId)
	return downUrl
}

func (dl *DownLock) GetDownlaodUrl(account module.Account, fileId string) string {
	var downloadUrl = ""
	if UrlCache.Has(account.Id + fileId) {
		downUrlCache, err := UrlCache.Get(account.Id + fileId)
		if err == nil {
			downloadUrl = downUrlCache.(DownUrlCacheBean).Url
			log.Debugf("get download url from cache:%s", downloadUrl)
		}
	} else {
		p, _ := base.GetPan(account.Mode)
		url, err := p.GetDownloadUrl(account, fileId)
		if err != nil {
			log.Error(err)
		}
		downloadUrl = url
		cacheTime := time.Now().Format("2006-01-02 15:04:05")
		if downloadUrl != "" {
			if account.Mode == "aliyundrive" {
				UrlCache.SetWithExpire(account.Id+fileId, DownUrlCacheBean{downloadUrl, cacheTime, util.GetExpireTime(cacheTime, time.Minute*230)}, time.Minute*230)
			} else {
				UrlCache.SetWithExpire(account.Id+fileId, DownUrlCacheBean{downloadUrl, cacheTime, util.GetExpireTime(cacheTime, time.Minute*10)}, time.Minute*10)
			}
			log.Debugf("get download url from api:" + downloadUrl)
		}
	}
	return downloadUrl
}

// clear file memory cache
func ClearFileCache(p string) {
	keys := FilesCache.Keys(false)
	for _, key := range keys {
		k := key.(string)
		if strings.HasPrefix(k, p) {
			FilesCache.Remove(k)
		}
	}
}

// upload file
func Upload(accountId, p string, c *gin.Context) string {
	_, fullPath, path, _ := util.ParseFullPath(p, "")
	form, _ := c.MultipartForm()
	files := form.File["uploadFile"]
	account := module.Account{}
	result := dao.DB.Raw("select * from account where id=?", accountId).Take(&account)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "指定的账号不存在"
	}
	pan, _ := base.GetPan(account.Mode)
	fileId := GetFileIdByPath(account, path, fullPath)
	if fileId == "" {
		return "指定目录不存在"
	}
	fileInfos := []*module.UploadInfo{}
	for _, file := range files {
		fileContent, _ := file.Open()
		byteContent, _ := ioutil.ReadAll(fileContent)
		fileInfos = append(fileInfos, &module.UploadInfo{
			FileName:    file.Filename,
			FileSize:    file.Size,
			ContentType: file.Header.Get("Content-Type"),
			Content:     byteContent,
		})
	}
	ok, r, err := pan.UploadFiles(account, fileId, fileInfos, true)
	if ok && err == nil {
		log.Debug(r)
		return "上传成功"
	} else {
		log.Debug(r)
		return "上传失败"
	}
}

// sync file cache
func Async(accountId, path string) string {
	account := module.Account{}
	result := dao.DB.Raw("select * from account where id=?", accountId).Take(&account)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "指定的账号不存在"
	}
	account.SyncDir = path
	account.SyncChild = 0
	if account.CachePolicy == "dc" {
		dao.SyncFilesCache(account)
	} else {
		ClearFileCache(path)
	}
	return "刷新成功"
}

// short url & qrcode
func ShortInfo(path, prefix string) (string, string, string) {
	si := module.ShareInfo{}
	dao.DB.Raw("select * from share_info where file_path = ?", path).First(&si)
	shortUrl := ""
	if path == "" {
		return "", "", "无效的id"
	}
	shortCode := ""
	if si.ShortCode != "" {
		shortCode = si.ShortCode
	} else {
		shortCodes, err := util.Transform(path)
		if err != nil {
			log.Errorln(err)
			return "", "", "短链生成失败"
		}
		for i, code := range shortCodes {
			si = module.ShareInfo{}
			dao.DB.Raw("select * from share_info where short_code = ?", code).First(&si)
			if si.ShortCode == "" {
				shortCode = shortCodes[i]
				break
			}
		}
		dao.DB.Create(module.ShareInfo{
			path, shortCode, []module.PwdFiles{},
		})
	}
	go dao.InitGlobalConfig()
	shortUrl = prefix + shortCode
	png, err := qrcode.Encode(shortUrl, qrcode.Medium, 256)
	if err != nil {
		panic(err)
	}
	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	return shortUrl, dataURI, "短链生成成功"
}

// get file data
func GetFileData(account module.Account, downUrl, r string) ([]byte, string, int) {
	client := httpClient(r)
	req, _ := http.NewRequest("GET", downUrl, nil)
	req.Header.Add("Range", r)
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	mtype := mimetype.Detect(data)
	return data, mtype.String(), resp.StatusCode
}

func httpClient(r string) *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	return &client
}

func AccountsToNodes(host string) []module.FileNode {
	fns := []module.FileNode{}
	ids := map[string]string{}
	for _, bypass := range module.GloablConfig.BypassList {
		fn := module.FileNode{
			FileId:     fmt.Sprintf("/%s", bypass.Name),
			IsFolder:   true,
			FileName:   bypass.Name,
			FileSize:   0,
			SizeFmt:    "-",
			FileType:   "",
			Path:       fmt.Sprintf("/%s", bypass.Name),
			ViewType:   "",
			LastOpTime: "",
			ParentId:   "",
		}
		fns = append(fns, fn)
		for _, ac := range bypass.Accounts {
			ids[ac.Id] = ac.Id
		}
	}
	for _, account := range module.GloablConfig.Accounts {
		_, exists := ids[account.Id]
		if !exists {
			fn := module.FileNode{
				FileId:     fmt.Sprintf("/%s", account.Name),
				IsFolder:   true,
				FileName:   account.Name,
				FileSize:   int64(account.FilesCount),
				SizeFmt:    "-",
				FileType:   "",
				Path:       fmt.Sprintf("/%s", account.Name),
				ViewType:   "",
				LastOpTime: account.LastOpTime,
				ParentId:   "",
			}
			if host != "" && account.Host != "" {
				if host == account.Host {
					fns = append(fns, fn)
				}
			} else {
				fns = append(fns, fn)
			}
		}
	}
	return fns
}

func GetRedirectUri(shorCode string) (string, string) {
	redirectUri := "/"
	v := ""
	si := module.ShareInfo{}
	result := dao.DB.Raw("select * from share_info where short_code=?", shorCode).First(&si)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if module.GloablConfig.EnablePreview == "0" || module.GloablConfig.ShortAction == "1" {
				redirectUri = si.FilePath
			} else {
				v = "v"
				redirectUri = si.FilePath
			}
		}
	}
	return redirectUri, v
}

// path: filePath
func GetLastNextFile(ac module.Account, path, fullPath, sortColumn, sortOrder string) (string, string) {
	var fns []module.FileNode
	var err error
	parentPath := util.GetParentPath(path)
	fileFullPath := fullPath
	fullPath = util.GetParentPath(fullPath)
	//get from cache first
	if FilesCache.Has(fullPath) {
		files, err := FilesCache.Get(fullPath)
		if err == nil {
			fcb := files.(FilesCacheBean)
			log.Debugf("get file from cache:%s", fullPath)
			fns = fcb.FileNodes
		}
	} else {
		if ac.CachePolicy == "dc" {
			fns, _ = GetFilesFromDb(ac, fullPath, "default", "null")
		} else {
			fns, _, err = GetFilesFromApi(ac, parentPath, fullPath, sortColumn, sortOrder)
			fns = FilterHideFiles(fns)
			if err == nil {
				cacheTime := time.Now().Format("2006-01-02 15:04:05")
				FilesCache.SetWithExpire(fullPath, FilesCacheBean{fns, false, cacheTime, util.GetExpireTime(cacheTime, time.Hour*time.Duration(ac.ExpireTimeSpan))}, time.Hour*time.Duration(ac.ExpireTimeSpan))
			}
		}
	}
	util.SortFileNode(sortColumn, sortOrder, fns)
	var lastFile, nextFile = "", ""
	for i, fn := range fns {
		if fn.Path == fileFullPath && i > 0 {
			if !fns[i-1].IsFolder {
				lastFile = fns[i-1].Path
			}
		}
		if fn.Path == fileFullPath && i < len(fns)-1 {
			if !fns[i+1].IsFolder {
				nextFile = fns[i+1].Path
			}
		}
	}
	return lastFile, nextFile
}

func GetFiles(ac module.Account, path, fullPath, sortColumn, sortOrder, viewType string) []module.FileNode {
	var fns []module.FileNode
	var err error
	//get from cache first
	if FilesCache.Has(fullPath) {
		files, err := FilesCache.Get(fullPath)
		if err == nil {
			fcb := files.(FilesCacheBean)
			log.Debugf("get file from cache:%s", fullPath)
			fns = fcb.FileNodes
		}
	} else {
		if ac.CachePolicy == "dc" {
			fns, _ = GetFilesFromDb(ac, fullPath, "default", "null")
		} else {
			fns, _, err = GetFilesFromApi(ac, path, fullPath, sortColumn, sortOrder)
			fns = FilterHideFiles(fns)
			if err == nil {
				cacheTime := time.Now().Format("2006-01-02 15:04:05")
				FilesCache.SetWithExpire(fullPath, FilesCacheBean{fns, false, cacheTime, util.GetExpireTime(cacheTime, time.Hour*time.Duration(ac.ExpireTimeSpan))}, time.Hour*time.Duration(ac.ExpireTimeSpan))
			}
		}
	}
	fns = FilterFilesByType(fns, viewType)
	util.SortFileNode(sortColumn, sortOrder, fns)
	return fns
}

func GetCacheData(pathEsc string) []module.Cache {
	cache := []module.Cache{}
	fns := []module.FileNode{}
	if pathEsc != "" {
		dao.DB.Raw("select * from file_node where is_delete = 0 and path like ?", "%"+pathEsc+"%").Find(&fns)
	} else {
		dao.DB.Raw("select * from file_node where is_delete = 0 limit 100").Find(&fns)
	}
	for _, fn := range fns {
		cache = append(cache, module.Cache{fn.Path, fn.CreateTime, "", "DB", fn})
	}
	filesCache := FilesCache.GetALL(false)
	for filePath, data := range filesCache {
		fc := data.(FilesCacheBean)
		if pathEsc != "" {
			if strings.Contains(filePath.(string), pathEsc) {
				cache = append(cache, module.Cache{filePath.(string), fc.CacheTime, fc.ExpireTime, "M-Files", fc.FileNodes})
			}
		} else {
			cache = append(cache, module.Cache{filePath.(string), fc.CacheTime, fc.ExpireTime, "M-Files", fc.FileNodes})
		}
	}
	fileCache := FileCache.GetALL(false)
	for filePath, data := range fileCache {
		fc := data.(FileCacheBean)
		if pathEsc != "" {
			if strings.Contains(filePath.(string), pathEsc) {
				cache = append(cache, module.Cache{filePath.(string), fc.CacheTime, fc.ExpireTime, "M-File", fc.FileNode})
			}
		} else {
			cache = append(cache, module.Cache{filePath.(string), fc.CacheTime, fc.ExpireTime, "M-File", fc.FileNode})
		}
	}
	urlCache := UrlCache.GetALL(false)
	for path, data := range urlCache {
		downUrlCache := data.(DownUrlCacheBean)
		if pathEsc != "" {
			if strings.Contains(path.(string), pathEsc) {
				cache = append(cache, module.Cache{path.(string), downUrlCache.CacheTime, downUrlCache.ExpireTime, "M-Download", downUrlCache.Url})
			}
		} else {
			cache = append(cache, module.Cache{path.(string), downUrlCache.CacheTime, downUrlCache.ExpireTime, "M-Download", downUrlCache.Url})
		}
	}
	return cache
}

func GetCacheByPath(path string) []module.Cache {
	cache := []module.Cache{}
	fn := module.FileNode{}
	dao.DB.Raw("select * from file_node where is_delete = 0 and path=? limit 100", path).First(&fn)
	if fn.Path != "" {
		cache = append(cache, module.Cache{fn.Path, fn.CreateTime, "", "DB", fn})
	}
	filesCache, err := FilesCache.Get(path)
	if err == nil {
		fc := filesCache.(FilesCacheBean)
		cache = append(cache, module.Cache{path, fc.CacheTime, fc.ExpireTime, "M-Files", fc.FileNodes})
	}
	fileCache, err := FileCache.Get(path)
	if err == nil {
		fc := fileCache.(FileCacheBean)
		cache = append(cache, module.Cache{path, fc.CacheTime, fc.ExpireTime, "M-File", fc.FileNode})
	}
	urlCache, err := UrlCache.Get(path)
	if err == nil {
		uc := urlCache.(DownUrlCacheBean)
		cache = append(cache, module.Cache{path, uc.CacheTime, uc.ExpireTime, "M-Download", uc.Url})
	}
	return cache
}

func CacheClear(path string, isLoopChildren string) {
	if isLoopChildren == "0" {
		keys := FilesCache.Keys(false)
		for _, key := range keys {
			if key.(string) == path || strings.HasPrefix(key.(string)+"/", path) {
				FilesCache.Remove(key)
			}
		}
	} else {
		FilesCache.Remove(path)
	}
	FileCache.Remove(path)
	UrlCache.Remove(path)
	accounts, cachePath := dao.FindAccountsByPath(path)
	for _, account := range accounts {
		account.SyncDir = cachePath
		if isLoopChildren == "0" {
			account.SyncChild = 0
		} else {
			account.SyncChild = 1
		}
		go dao.SyncFilesCache(account)
	}
}

func GetBypassByAccountId(accountId string) module.Bypass {
	return dao.SelectBypassByAccountId(accountId)
}

func UpdateCache(account module.Account, cachePath string) string {
	msg := "缓存清理成功"
	if account.CachePolicy == "nc" {
		msg = "当前网盘无需刷新操作！"
	} else if account.CachePolicy == "mc" {
		ClearFileCache(cachePath)
	} else {
		if dao.SYNC_STATUS == 1 {
			msg = "缓存任务正在执行，请稍后重试！"
		} else {
			if account.Status == -1 {
				msg = "目录缓存中，请勿重复操作！"
			} else {
				account.SyncDir = cachePath
				account.SyncChild = 0
				dao.DB.Table("account").Where("id=?", account.Id).UpdateColumn("status", -1)
				if dao.DB_TYPE == "sqlite" {
					dao.SYNC_STATUS = 1
				}
				go dao.SyncFilesCache(account)
				go dao.InitGlobalConfig()
				msg = "正在缓存目录，请稍后刷新页面查看缓存结果！"
			}
		}
	}
	return msg
}

func BatchUpdateCache(ids []string) string {
	msg := "缓存清理成功"
	if dao.SYNC_STATUS == 1 {
		msg = "缓存任务正在执行，请稍后重试！"
	} else {
		go func() {
			accounts := dao.SelectAccountsById(ids)
			for _, account := range accounts {

				bypass := GetBypassByAccountId(account.Id)
				cachePath := "/" + account.Name
				if bypass.Name != "" {
					cachePath = "/" + bypass.Name
				}
				if account.CachePolicy == "nc" {
				} else if account.CachePolicy == "mc" {
					ClearFileCache(cachePath)
				} else {
					account.SyncDir = cachePath
					account.SyncChild = 0
					dao.DB.Table("account").Where("id=?", account.Id).UpdateColumn("status", -1)
					if dao.DB_TYPE == "sqlite" {
						dao.InitGlobalConfig()
						dao.SyncFilesCache(account)
					} else {
						go dao.InitGlobalConfig()
						go dao.SyncFilesCache(account)
					}

				}
			}
		}()
		msg = "正在缓存目录，请稍后刷新页面查看缓存结果！"
	}
	return msg
}

func GetAccounts() []module.Account {
	accounts := []module.Account{}
	bacIds := []string{}
	bypasses := dao.GetBypassList()
	if len(bypasses) > 0 {
		for _, bypass := range bypasses {
			ac := module.Account{}
			ac.Name = bypass.Name
			ac.Mode = "native"
			accounts = append(accounts, ac)
			if len(bypass.Accounts) > 0 {
				for _, bac := range bypass.Accounts {
					bacIds = append(bacIds, bac.Id)
				}
			}
		}
	}
	if len(module.GloablConfig.Accounts) > 0 {
		for _, account := range module.GloablConfig.Accounts {
			f := false
			if len(bacIds) > 0 {
				for _, id := range bacIds {
					if id == account.Id {
						f = true
					}
				}
			}
			if !f {
				accounts = append(accounts, account)
			}
		}
	}
	return accounts
}

func Files(ac module.Account, path, fullPath string) []module.FileNode {
	fns := []module.FileNode{}
	isFile := false
	var err error
	if ac.CachePolicy == "nc" {
		fns, isFile, _ = GetFilesFromApi(ac, path, fullPath, "default", "null")
		fns = FilterHideFiles(fns)
	} else if ac.CachePolicy == "mc" {
		if FilesCache.Has(fullPath) {
			files, err := FilesCache.Get(fullPath)
			if err == nil {
				fcb := files.(FilesCacheBean)
				log.Debugf("get file from cache:%s", fullPath)
				fns = fcb.FileNodes
				isFile = fcb.IsFile
			}
		} else {
			fns, isFile, err = GetFilesFromApi(ac, path, fullPath, "default", "null")
			if err == nil {
				fns = FilterHideFiles(fns)
				cacheTime := time.Now().Format("2006-01-02 15:04:05")
				FilesCache.SetWithExpire(fullPath, FilesCacheBean{fns, isFile, cacheTime, util.GetExpireTime(cacheTime, time.Hour*time.Duration(ac.ExpireTimeSpan))}, time.Hour*time.Duration(ac.ExpireTimeSpan))
			}
		}
	} else if ac.CachePolicy == "dc" {
		fns, isFile = GetFilesFromDb(ac, fullPath, "default", "null")
		fns = FilterHideFiles(fns)
	}

	return fns
}

func DeleteFile(ac module.Account, path, fullPath string) error {
	//1. delete file from disk
	disk, _ := base.GetPan(ac.Mode)
	fileId := GetFileIdByPath(ac, path, fullPath)
	_, _, err := disk.Remove(ac, fileId)
	//2. delete file from cache
	if err == nil {
		FileCache.Remove(fullPath)
		FilesCache.Remove(fullPath)
		FilesCache.Remove(util.GetParentPath(fullPath))
		dao.DeleteFileNodes(ac.Id, fileId)
	}
	return err
}

func GetFileFromApi(ac module.Account, path, fullPath string) (module.FileNode, error) {
	p, _ := base.GetPan(ac.Mode)
	fileId := GetFileIdByPath(ac, path, fullPath)
	if fileId != "" || (fileId == "" && path == "/") {
		file, err := p.File(ac, fileId, fullPath)
		return file, err
	}
	return module.FileNode{}, nil
}

func File(ac module.Account, path, fullPath string) (module.FileNode, error) {
	if ac.CachePolicy == "nc" {
		return GetFileFromApi(ac, path, fullPath)
	} else if ac.CachePolicy == "mc" {
		if FileCache.Has(fullPath) {
			files, err := FileCache.Get(fullPath)
			if err == nil {
				fcb := files.(FileCacheBean)
				log.Debugf("get file from cache:%s", fullPath)
				return fcb.FileNode, nil
			}
		} else {
			fn, err := GetFileFromApi(ac, path, fullPath)
			cacheTime := time.Now().Format("2006-01-02 15:04:05")
			FileCache.SetWithExpire(fullPath, FileCacheBean{fn, cacheTime, util.GetExpireTime(cacheTime, time.Hour*time.Duration(ac.ExpireTimeSpan))}, time.Hour*time.Duration(ac.ExpireTimeSpan))
			return fn, err
		}
	} else if ac.CachePolicy == "dc" {
		fn, _ := dao.FindFileByPath(ac, fullPath)
		return fn, nil
	}
	return module.FileNode{}, nil
}

// webdav upload callback
func UploadCall(account module.Account, file module.FileNode, overwrite bool) {
	if account.CachePolicy == "mc" {
		parentPath := util.GetParentPath(file.Path)
		FilesCache.Remove(parentPath)
		FileCache.Remove(file.Path)
	} else if account.CachePolicy == "dc" {
		file.Id = uuid.NewV4().String()
		file.IsDelete = 0
		fn := module.FileNode{}
		exist := false
		result := dao.DB.Table("file_node").
			Where("path=?", file.Path).Take(&fn)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			exist = true
		}
		if overwrite && exist {
			dao.DB.Table("file_node").
				Where("path=?", file.Path).
				Update("last_op_time", file.LastOpTime).
				Update("file_size", file.FileSize).
				Update("size_fmt", util.FormatFileSize(file.FileSize))
		} else {
			dao.DB.Create(&file)
		}
	}
}

// webdav mkdir callback
func MkdirCall(account module.Account, file module.FileNode) {
	UploadCall(account, file, false)
}

// webdav move callback
func MoveCall(account module.Account, fileId, srcFullPath, dstFullPath string) {
	UrlCache.Remove(account.Id + fileId)
	if account.CachePolicy == "mc" {
		parentPath := util.GetParentPath(srcFullPath)
		dstParentPath := util.GetParentPath(dstFullPath)
		FilesCache.Remove(parentPath)
		FilesCache.Remove(dstParentPath)
		FileCache.Remove(dstFullPath)
		FileCache.Remove(srcFullPath)
	} else if account.CachePolicy == "dc" {
		fn := module.FileNode{}
		dao.DB.Where("path=?", srcFullPath).First(&fn)
		p, _ := base.GetPan(account.Mode)
		newFn, _ := p.File(account, fileId, dstFullPath)
		fn.LastOpTime = newFn.LastOpTime
		fn.Path = newFn.Path
		fn.ParentPath = newFn.ParentPath
		fn.FileId = newFn.FileId
		fn.ParentId = newFn.ParentId
		fn.FileName = newFn.FileName
		dao.DB.Model(&module.FileNode{}).Where("id=?", fn.Id).Updates(fn)
		if fn.IsFolder {
			//TODO if file_id is path may not work
			account.SyncDir = newFn.Path
			account.SyncChild = 0
			dao.SyncFilesCache(account)
		}
	}
}

func FtpDownload(ac module.Account, downUrl string, fileNode module.FileNode, c *gin.Context) {
	p, _ := base.GetPan(ac.Mode)
	statusCode := http.StatusOK
	if c.GetHeader("Range") != "" {
		statusCode = http.StatusPartialContent
	}
	r, err := p.(*ftp.FTP).ReadFileReader(ac, downUrl, 0)
	defer r.Close()
	if err == nil {
		fileName := url.QueryEscape(fileNode.FileName)
		extraHeaders := map[string]string{
			"Content-Disposition": `attachment; filename="` + fileName + `"`,
			"Accept-Ranges":       "bytes",
			"Content-Range":       fmt.Sprintf("bytes %d-%d/%d", 0, fileNode.FileSize-1, fileNode.FileSize),
		}
		c.DataFromReader(statusCode, fileNode.FileSize,
			util.GetMimeTypeByExt(fileNode.FileType), r, extraHeaders)
	}
}

func WebDavDownload(ac module.Account, downUrl string, fileNode module.FileNode, c *gin.Context) {
	p, _ := base.GetPan(ac.Mode)
	statusCode := http.StatusOK
	if c.GetHeader("Range") != "" {
		statusCode = http.StatusPartialContent
	}
	r, err := p.(*webdav.WebDav).ReadFileReader(ac, downUrl, 0, fileNode.FileSize)
	defer r.Close()
	if err == nil {
		fileName := url.QueryEscape(fileNode.FileName)
		extraHeaders := map[string]string{
			"Content-Disposition": `attachment; filename="` + fileName + `"`,
			"Accept-Ranges":       "bytes",
			"Content-Range":       fmt.Sprintf("bytes %d-%d/%d", 0, fileNode.FileSize-1, fileNode.FileSize),
		}
		c.DataFromReader(statusCode, fileNode.FileSize,
			util.GetMimeTypeByExt(fileNode.FileType), r, extraHeaders)
	}
}

func UploadConfig(config module.Config) string {
	//common config
	configItem := util.ConfigToItem(config)
	for k, v := range configItem {
		val, ok := v.(string)
		if ok {
			dao.DB.Table("config_item").Where("k=?", k).Update("v", val)
		}
	}
	//accounts
	if len(config.Accounts) > 0 {
		for _, account := range config.Accounts {
			ac := module.Account{}
			err := dao.DB.Table("account").Where("id=?", account.Id).First(&ac).Error
			if err == gorm.ErrRecordNotFound {
				//insert
				dao.DB.Table("account").Create(util.AccountToMap(account))
			} else {
				//update
				dao.DB.Table("account").Where("id=?", account.Id).Updates(util.AccountToMap(account))
			}
		}
	}
	//bypass
	if len(config.BypassList) > 0 {
		for _, bypass := range config.BypassList {
			dao.SaveBypass(bypass)
		}
	}
	//short_info
	if len(config.ShareInfoList) > 0 {
		for _, shareInfo := range config.ShareInfoList {
			dao.SaveShareInfo(shareInfo)
		}
	}
	//pwd
	if len(config.PwdFiles) > 0 {
		for _, pd := range config.PwdFiles {
			dao.SavePwdFile(pd)
		}
	}
	//hide
	if len(config.HideFiles) > 0 {
		for hf, _ := range config.HideFiles {
			dao.SaveHideFile(hf)
		}
	}
	return "配置导入成功，如有密码修改，请重新登录"
}

func UploadPwdFile(content string) {
	content = strings.TrimSpace(content)
	if content != "" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			columns := strings.Split(line, "\t")
			pwdFile := module.PwdFiles{}
			if len(columns) > 0 && columns[0] != "" {
				pwdFile.FilePath = strings.TrimSpace(columns[0])
				if len(columns) > 1 {
					pwdFile.Password = strings.TrimRight(columns[1], "\r")
				} else {
					pwdFile.Password = util.RandomPassword(8)
					pwdFile.ExpireAt = 0
				}
				if len(columns) > 2 && columns[2] != "" {
					ex, _ := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimRight(columns[2], "\r"), time.Local)
					pwdFile.ExpireAt = ex.UTC().Unix()
				}
				if len(columns) > 3 && columns[3] != "" {
					pwdFile.Info = strings.TrimRight(columns[3], "\r")
				}
				dao.SavePwdFile(pwdFile)
				dao.InitGlobalConfig()
			}
		}
	}
}

func GenShareInfo(urlPrefix, pwdId string) string {
	//1. pwd
	pwdFile := module.PwdFiles{}
	dao.DB.Where("id=?", pwdId).First(&pwdFile)
	fileName := util.GetFileName(pwdFile.FilePath)
	//2. share
	shortUrl, _, _ := ShortInfo(pwdFile.FilePath, urlPrefix)
	msg := fmt.Sprintf("「%s」%s 密码: %s", fileName, shortUrl, pwdFile.Password)
	return msg
}
