package dao

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/pan"
	"github.com/libsgh/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/smallnest/weighted"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
)

var DB *gorm.DB
var InitConfigItems = []module.ConfigItem{
	{"site_name", "", "common"},
	{"account_choose", "default", "common"},
	{"admin_password", "PanIndex", "common"},
	{"s_column", "default", "common"},
	{"s_order", "asc", "common"},
	{"readme", "1", "common"},
	{"head", "1", "common"},
	{"favicon_url", "", "appearance"},
	{"footer", "", "appearance"},
	{"css", "", "appearance"},
	{"js", "", "appearance"},
	{"theme", "mdui", "appearance"},
	{"enable_preview", "1", "view"},
	{"image", "png,gif,jpg,bmp,jpeg,ico", "view"},
	{"video", "mp4,mkv,m3u8,ts,avi", "view"},
	{"audio", "mp3,wav,flac,ape", "view"},
	{"code", "txt,go,html,js,java,json,css,lua,sh,sql,py,cpp,xml,jsp,properties,yaml,ini", "view"},
	{"doc", "doc,docx,dotx,ppt,pptx,xls,xlsx", "view"},
	{"other", "*", "view"},
	{"enable_lrc", "0", "view"},
	{"lrc_path", "", "view"},
	{"enable_subtitle", "0", "view"},
	{"subtitle_path", "", "view"},
	{"enable_safety_link", "0", "safety"},
	{"only_referrer", "", "safety"},
	{"is_null_referrer", "0", "safety"},
	{"admin_path", "/admin", "common"},
	{"cdn", "1", "common"},
	{"dav_path", "/dav", "dav"},
	{"enable_dav", "0", "dav"},
	{"dav_mode", "0", "dav"},
	{"dav_down_mode", "1", "dav"},
	{"dav_user", "webdav", "dav"},
	{"dav_password", "1234", "dav"},
	{"proxy", "", "common"},
}

type Db interface {
	CreateDb(dsn string) //get dao connection
}

var DbMap = map[string]Db{}

func RegisterDb(driver string, db Db) {
	DbMap[driver] = db
}

func GetDb(driver string) (db Db, ok bool) {
	db, ok = DbMap[driver]
	return
}

func InitDb() {
	DB.AutoMigrate(&module.FileNode{})
	DB.AutoMigrate(&module.ShareInfo{})
	DB.AutoMigrate(&module.ConfigItem{})
	DB.AutoMigrate(&module.Account{})
	DB.AutoMigrate(&module.PwdFiles{})
	DB.AutoMigrate(&module.HideFiles{})
	DB.AutoMigrate(&module.Bypass{})
	DB.AutoMigrate(&module.BypassAccounts{})
	//init data
	var count int64
	err := DB.Model(module.ConfigItem{}).Count(&count).Error
	SaveConfigItems(InitConfigItems)
	if err != nil {
		panic(err)
	} else if count == 0 {
		//DB.Create(&InitConfigItems)
		rand.Seed(time.Now().UnixNano())
		ApiToken := strconv.Itoa(rand.Intn(10000000))
		configItem := module.ConfigItem{K: "api_token", V: ApiToken, G: "common"}
		DB.Create(configItem)
	}
}

func SaveConfigItems(items []module.ConfigItem) {
	for _, item := range items {
		var c module.ConfigItem
		err := DB.Where("k=?", item.K).First(&c).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			DB.Create(item)
		}
	}
}

//init global config
func InitGlobalConfig() {
	c := module.Config{}
	cis := []module.ConfigItem{}
	accounts := []module.Account{}
	DB.Raw("select * from config_item where 1=1").Find(&cis)
	configMap := make(map[string]interface{})
	for _, ci := range cis {
		configMap[ci.K] = ci.V
	}
	configJson, _ := jsoniter.MarshalToString(configMap)
	jsoniter.Unmarshal([]byte(configJson), &c)
	DB.Raw("select * from account order by seq asc").Find(&accounts)
	c.Accounts = accounts
	c.HideFiles = GetHideFilesMap()
	c.PwdFiles = GetPwdFilesMap()
	c.BypassList = GetBypassList()
	c.CdnFiles = util.GetCdnFilesMap(c.Cdn, module.VERSION)
	module.GloablConfig = c
	RoundRobin()
}

//pwd files to map
func GetHideFilesMap() map[string]string {
	m := make(map[string]string)
	hidefiles := []module.HideFiles{}
	DB.Where("1 = 1").Find(&hidefiles)
	if len(hidefiles) > 0 {
		for _, hidefile := range hidefiles {
			m[hidefile.FilePath] = "1"
		}
	}
	return m
}

//hide files to map
func GetPwdFilesMap() map[string]string {
	m := make(map[string]string)
	pwdfiles := []module.PwdFiles{}
	DB.Where("1 = 1").Find(&pwdfiles)
	if len(pwdfiles) > 0 {
		for _, pwdfile := range pwdfiles {
			m[pwdfile.FilePath] = pwdfile.Password
		}
	}
	return m
}

//get pwd from full path
func GetPwdFromPath(path string) (module.PwdFiles, bool) {
	pwdfile := module.PwdFiles{}
	result := DB.Where("? like file_path", path+"%").First(&pwdfile)
	if result.Error != nil {
		return pwdfile, false
	}
	return pwdfile, true
}

//find account by name
func FindAccountByName(name string) module.Account {
	account := module.Account{}
	DB.Where("name = ?", name).First(&account)
	return account
}

//find first account by seq
func FindAccountBySeq(seq int) module.Account {
	account := module.Account{}
	DB.Where("seq = ?", seq).First(&account)
	return account
}

//find first file by path
func FindFileByPath(ac module.Account, path string) (module.FileNode, bool) {
	fn := module.FileNode{}
	ok := false
	sql := `select fn.* from file_node fn where fn.account_id = ? and fn.path = ?`
	DB.Raw(sql, ac.Id, path).First(&fn)
	if !fn.IsFolder {
		ok = true
	}
	if fn.FileId == "" {
		//check root file
		sql := `select fn.* from file_node fn where fn.account_id = ? and fn.parent_path = ?`
		DB.Raw(sql, ac.Id, path).First(&fn)
		if fn.FileId != "" {
			_, fileName := util.ParsePath(path)
			return module.FileNode{
				FileId:     ac.RootId,
				FileName:   fileName,
				FileSize:   0,
				IsFolder:   true,
				Path:       path,
				LastOpTime: time.Now().Format("2006-01-02 15:04:05"),
			}, true
		}

	}
	return fn, ok
}

//find first file  list by path
func FindFileListByPath(ac module.Account, path, sortColumn, sortOrder string) []module.FileNode {
	fns := []module.FileNode{}
	tx := DB.Where("is_delete=0 and hide =0 and account_id=? and parent_path=?", ac.Id, path)
	tx.Order("is_folder desc")
	/*if sortColumn != "default" && sortOrder != "" {
		tx = tx.Order(fmt.Sprintf("%s %s", sortColumn, sortOrder))
	} else {
		tx = tx.Order(fmt.Sprintf("last_op_time asc"))
	}*/
	tx.Find(&fns)
	return fns
}

//update config
func UpdateConfig(config map[string]string) {
	for key, value := range config {
		DB.Table("config_item").Where("k=?", key).Update("v", value)
	}
	InitGlobalConfig()
}

//get config
func GetConfig() module.Config {
	c := module.Config{}
	cis := []module.ConfigItem{}
	accounts := []module.Account{}
	DB.Raw("select * from config_item where 1=1").Find(&cis)
	configMap := make(map[string]interface{})
	for _, ci := range cis {
		configMap[ci.K] = ci.V
	}
	configJson, _ := jsoniter.MarshalToString(configMap)
	jsoniter.Unmarshal([]byte(configJson), &c)
	DB.Raw("select * from account order by `seq` asc").Find(&accounts)
	c.Accounts = accounts
	module.GloablConfig = c
	return c
}

//get account
func GetAccountById(id string) module.Account {
	account := module.Account{}
	DB.Where("id = ?", id).First(&account)
	return account
}

//delete accounts
func DeleteAccounts(ids []string) {
	for _, id := range ids {
		//delete account db file
		DB.Where("account_id = ?", id).Delete(module.FileNode{})
		//delete account
		var a module.Account
		var si module.ShareInfo
		a.Id = id
		si.AccountId = id
		DB.Model(module.Account{}).Where("1=1").Delete(a)
		//delete share info
		DB.Model(module.ShareInfo{}).Where("1=1").Delete(si)
		//delete login account
		pan.RemoveLoginAccount(id)
		// refresh global config
		InitGlobalConfig()
		//remove login account
		pan.RemoveLoginAccount(id)
	}
}

//Loop add files
func LoopCreateFiles(account module.Account, fileId, path string, hide, hasPwd int) {
	pan, _ := pan.GetPan(account.Mode)
	fileNodes, err := pan.Files(account, fileId, path, "default", "null")
	if err != nil {
		log.Warningf("%s get files error", account.Mode)
		log.Errorln(err)
	}
	for _, fn := range fileNodes {
		util.FileNodeAuth(&fn, hide, hasPwd)
		if fn.IsFolder && account.SyncChild == 0 {
			LoopCreateFiles(account, fn.FileId, fn.Path, fn.Hide, fn.HasPwd)
		}
	}
	if len(fileNodes) > 0 {
		DB.Create(&fileNodes)
	}
}

//sync account status
func SyncAccountStatus(account module.Account) {
	DB.Table("account").Where("id=?", account.Id).Update("cookie_status", -1)
	pan, _ := pan.GetPan(account.Mode)
	auth, err := pan.AuthLogin(&account)
	if err == nil && auth != "" {
		log.Debugf("[%s] %s auth login success", account.Mode, account.Name)
		DB.Table("account").Where("id=?", account.Id).Update("cookie_status", 2)
		DB.Table("account").Where("id=?", account.Id).Update("refresh_token", auth)
	} else {
		log.Errorln(err)
		log.Errorf("[%s] %s auth login fail, api return : %s", account.Mode, account.Name, auth)
		DB.Table("account").Where("id=?", account.Id).Update("cookie_status", 4)
	}
}

//sync files cache
func SyncFilesCache(account module.Account) {
	t1 := time.Now()
	dbFile := module.FileNode{}
	result := DB.Raw("select * from file_node where path=? and is_delete=0 and account_id=?", account.SyncDir, account.Id).Take(&dbFile)
	isRoot := false
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		account.RootId = dbFile.FileId
		isRoot = true
	}
	//cache new files
	LoopCreateFiles(account, account.RootId, account.SyncDir, 0, 0)
	//handle old files && update account status
	var fileNodeCount int64
	DB.Model(&module.FileNode{}).Where("account_id=? and is_delete=1", account.Id).Count(&fileNodeCount)
	status := 3
	if int(fileNodeCount) > 0 {
		status = 2
		if isRoot {
			//删除旧数据
			DB.Where("account_id=? and is_delete=0", account.Id).Delete(module.FileNode{})
			//暴露新数据
			DB.Table("file_node").Where("account_id=?", account.Id).Update("is_delete", 0)
		} else {
			RefreshFileNodes(account.Id, account.RootId)
		}
		log.Infoln("[DB cache][" + account.Name + "]refresh >> success")
	}
	t2 := time.Now()
	d := t2.Sub(t1)
	now := time.Now().UTC().Add(8 * time.Hour)
	DB.Table("account").Where("id=?", account.Id).Updates(map[string]interface{}{
		"status": status, "files_count": int(fileNodeCount), "last_op_time": now.Format("2006-01-02 15:04:05"),
		"time_span": util.ShortDur(d),
	})
	InitGlobalConfig()
}

func RefreshFileNodes(accountId, fileId string) {
	tmpList := []module.FileNode{}
	list := []module.FileNode{}
	DB.Raw("select * from file_node where parent_id=? and is_delete=0 and account_id=?", fileId, accountId).Find(&tmpList)
	GetAllNodes(&tmpList, &list)
	for _, fn := range list {
		DB.Where("id=?", fn.Id).Delete(module.FileNode{})
	}
	DB.Table("file_node").Where("account_id=?", accountId).Update("is_delete", 0)
}

func DeleteFileNodes(accountId, fileId string) {
	tmpList := []module.FileNode{}
	list := []module.FileNode{}
	DB.Raw("select * from file_node where parent_id=? and is_delete=0 and account_id=?", fileId, accountId).Find(&tmpList)
	GetAllNodes(&tmpList, &list)
	for _, fn := range list {
		DB.Where("id=?", fn.Id).Delete(module.FileNode{})
	}
	DB.Where("file_id=? and account_id=?", fileId, accountId).Delete(module.FileNode{})
}

func GetAllNodes(tmpList, list *[]module.FileNode) {
	for _, fn := range *tmpList {
		tmpList = &[]module.FileNode{}
		DB.Raw("select * from file_node where parent_id=? and is_delete=0", fn.FileId).Find(&tmpList)
		*list = append(*list, fn)
		if len(*tmpList) != 0 {
			GetAllNodes(tmpList, list)
		}
	}
}

//sort accounts
func SortAccounts(ids []string) {
	for i, id := range ids {
		i++
		DB.Model(module.Account{}).Where("id=?", id).Update("seq", i)
	}
	InitGlobalConfig()
}

//get file id by path
func GetFileIdByPath(accountId, path string) string {
	fn := module.FileNode{}
	if path == "/" {
		DB.Raw("select * from file_node where parent_path=? and is_delete=0 and account_id=? limit 1", path, accountId).Take(&fn)
		return fn.ParentId
	} else {
		DB.Raw("select * from file_node where path=? and is_delete=0 and account_id=?", path, accountId).Take(&fn)
		return fn.FileId
	}
}

//save hide file
func SaveHideFile(filePath string) {
	hideFile := module.HideFiles{FilePath: filePath}
	err := DB.Where("file_path=?", filePath).First(&module.HideFiles{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		DB.Create(&hideFile)
	}
	InitGlobalConfig()
}

//delete hide file
func DeleteHideFiles(filePaths []string) {
	for _, filePath := range filePaths {
		c := DB.Where("file_path=?", filePath).Delete(&module.HideFiles{}).RowsAffected
		log.Debugf("delete hide file [%s], result [%d]", filePath, c)
	}
	InitGlobalConfig()
}

//save pwd file
func SavePwdFile(pwdFile module.PwdFiles) {
	err := DB.Where("file_path=?", pwdFile.FilePath).First(&module.PwdFiles{}).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		DB.Create(&pwdFile)
	} else {
		DB.Table("pwd_files").Where("file_path=?", pwdFile.FilePath).Update("password", pwdFile.Password)
	}
	InitGlobalConfig()
}

//delete hide file
func DeletePwdFiles(filePaths []string) {
	for _, filePath := range filePaths {
		c := DB.Where("file_path=?", filePath).Delete(&module.PwdFiles{}).RowsAffected
		log.Debugf("delete pwd file [%s], result [%d]", filePath, c)
	}
	InitGlobalConfig()
}

func GetBypassList() []module.Bypass {
	list := []module.Bypass{}
	DB.Find(&list)
	if len(list) > 0 {
		for i := 0; i < len(list); i++ {
			accounts := []module.Account{}
			DB.Raw(`select
				a.*
					from
				account a
				left join bypass_accounts ba on
				ba.account_id = a.id
				where
				ba.bypass_id =?`, list[i].Id).Find(&accounts)
			list[i].Accounts = accounts
		}
	}
	return list
}

func SaveBypass(bypass module.Bypass) string {
	if bypass.Id != "" {
		err := DB.Where("name=? and id!=?", bypass.Name, bypass.Id).First(&module.Bypass{}).Error
		if err == nil {
			return "保存失败，分流名称已存在！"
		}
		//check account bind
		for _, account := range bypass.Accounts {
			acs := []module.BypassAccounts{}
			DB.Where("account_id = ? and and bypass_id!=?", account.Id, bypass.Id).Find(&acs)
			if len(acs) > 0 {
				return "保存失败，网盘已被其他分流绑定！"
			}
		}
		DB.Where("id=?", bypass.Id).Updates(&bypass)
	} else {
		err := DB.Where("name=?", bypass.Name).First(&module.Bypass{}).Error
		if err == nil {
			return "保存失败，分流名称已存在！"
		}
		//check account bind
		for _, account := range bypass.Accounts {
			acs := []module.BypassAccounts{}
			DB.Where("account_id = ?", account.Id).Find(&acs)
			if len(acs) > 0 {
				return "保存失败，网盘已被其他分流绑定！"
			}
		}
		bypass.Id = uuid.NewV4().String()
		DB.Create(&bypass)
	}
	DB.Where("bypass_id = ?", bypass.Id).Delete(&module.BypassAccounts{})
	for _, account := range bypass.Accounts {
		ba := module.BypassAccounts{bypass.Id, account.Id}
		DB.Create(&ba)
	}
	InitGlobalConfig()
	return "保存成功！"
}

func DeleteBypass(ids []string) {
	for _, id := range ids {
		c := DB.Where("id=?", id).Delete(&module.Bypass{}).RowsAffected
		DB.Where("bypass_id=?", id).Delete(&module.BypassAccounts{})
		log.Debugf("delete bypass [%s], result [%d]", id, c)
	}
	InitGlobalConfig()
}

func RoundRobin() {
	if len(module.GloablConfig.BypassList) > 0 {
		for i := 0; i < len(module.GloablConfig.BypassList); i++ {
			rrw := weighted.NewRandW()
			for _, account := range module.GloablConfig.BypassList[i].Accounts {
				rrw.Add(account, 1)
			}
			module.GloablConfig.BypassList[i].Rw = rrw
		}
	}
}

func FindAccountsByPath(path string) ([]module.Account, string) {
	accounts := []module.Account{}
	fn := module.FileNode{}
	DB.Where("path=?", path).First(&fn)
	if fn.Path != "" {
		if !fn.IsFolder {
			path = fn.ParentPath
		}
	}
	DB.Raw(`select a.* from file_node fn left join account a on a.id = fn.account_id where fn.path = ? group by a.id`, path).
		Find(&accounts)
	if len(accounts) == 0 {
		DB.Raw(`select a.* from file_node fn left join account a on a.id = fn.account_id where fn.parent_path = ? group by a.id`, path).
			Find(&accounts)
	}
	return accounts, path
}

func UpdateCacheConfig(account module.Account) {
	DB.Model(&[]module.Account{}).
		Select("CachePolicy", "SyncDir", "SyncChild", "ExpireTimeSpan", "SyncCron").
		Where("id=?", account.Id).
		Updates(&account)
	if account.CachePolicy == "dc" {
		ac := GetAccountById(account.Id)
		go SyncFilesCache(ac)
		go SaveCacheCron(ac)
	} else {
		DB.Where("account_id = ?", account.Id).Delete(module.FileNode{})
	}
	InitGlobalConfig()
}

func SaveCacheCron(ac module.Account) {
	c, ok := util.CacheCronMap[ac.Id]
	if ok {
		if ac.CachePolicy != "dc" || ac.SyncCron == "" || ac.SyncDir == "" {
			util.Cron.Remove(c)
			delete(util.CacheCronMap, ac.Id)
		} else {
			util.Cron.Remove(c)
			entryId, err := util.Cron.AddFunc(ac.SyncCron, func() {
				SyncFilesCache(ac)
			})
			if err == nil {
				util.CacheCronMap[ac.Id] = entryId
			}
		}
	} else {
		if ac.CachePolicy == "dc" && ac.SyncCron != "" && ac.SyncDir != "" {
			util.Cron.Remove(c)
			entryId, err := util.Cron.AddFunc(ac.SyncCron, func() {
				SyncFilesCache(ac)
			})
			if err == nil {
				util.CacheCronMap[ac.Id] = entryId
			}
		}
	}

}

func SaveAccount(account module.Account) string {
	//check name exists
	if AccountNameExist(account.Id, account.Name) {
		return "保存失败，网盘（或分流）名称已存在！"
	}
	if account.Id == "" {
		account.Id = uuid.NewV4().String()
		account.CachePolicy = "nc"
		account.SyncDir = "/" + account.Name
		account.SyncChild = 0
		account.ExpireTimeSpan = 1
		account.SyncCron = ""
		account.LastOpTime = time.Now().Format("2006-01-02 15:04:05")
		var seq int
		DB.Table("account").Raw("select seq from account where 1=1 order by seq desc").First(&seq)
		account.Seq = seq + 1
		DB.Create(&account)
	} else {
		account.LastOpTime = time.Now().Format("2006-01-02 15:04:05")
		DB.Model(&[]module.Account{}).
			Select("Id", "Name", "Mode", "User", "Password", "RefreshToken", "AccessToken",
				"RedirectUri", "ApiUrl", "RootId", "LastOpTime", "DownTransfer", "TransferUrl").
			Where("id=?", account.Id).
			Updates(&account)
	}
	SyncAccountStatus(account)
	InitGlobalConfig()
	return "保存成功！"
}

func AccountNameExist(id, name string) bool {
	if id != "" {
		err := DB.Where("name=? and id!=?", name, id).First(&module.Account{}).Error
		if err == nil {
			return true
		}
	} else {
		err := DB.Where("name=?", name).First(&module.Account{}).Error
		if err == nil {
			return true
		}
	}
	err := DB.Where("name=?", name).First(&module.Bypass{}).Error
	if err == nil {
		return true
	}
	return false
}
