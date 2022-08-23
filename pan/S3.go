package pan

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	RegisterPan("s3", &S3{})
}

type S3 struct {
}

var S3s = map[string]*session.Session{}

func (s S3) AuthLogin(account *module.Account) (string, error) {
	cfg := &aws.Config{
		Credentials:                    credentials.NewStaticCredentials(account.User, account.Password, ""),
		Endpoint:                       aws.String(account.ApiUrl),
		Region:                         aws.String(account.SiteId),
		DisableRestProtocolURICleaning: aws.Bool(true),
		S3ForcePathStyle:               aws.Bool(util.If(account.PathStyle == "Path", true, false).(bool)), //oss,minio,aws: false, oracle,cos: true
	}
	sess, err := session.NewSession(cfg)
	S3s[account.Id] = sess
	return "ok", err
}

func (s S3) IsLogin(account *module.Account) bool {
	return true
}

func (s S3) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	session := S3s[account.Id]
	c := s3.New(session)
	marker := ""
	for {
		input := &s3.ListObjectsInput{
			Bucket:    aws.String(account.RedirectUri),
			Marker:    aws.String(marker),
			Delimiter: aws.String("/"),
		}
		if fileId != "" {
			input.Prefix = aws.String(fileId)
		}
		listObjects, err := c.ListObjects(input)
		if err != nil {
			return fileNodes, err
		}
		for _, object := range listObjects.CommonPrefixes {
			fn := s.ToFolderNode(object)
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
		for _, object := range listObjects.Contents {
			if *object.Key == fileId {
				continue
			}
			fn := s.ToFileNode(object)
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
		if listObjects.IsTruncated == nil {
			return fileNodes, errors.New("IsTruncated nil")
		}
		if *listObjects.IsTruncated {
			marker = *listObjects.NextMarker
		} else {
			break
		}
	}
	return fileNodes, nil
}

func (s S3) File(account module.Account, fileId, path string) (module.FileNode, error) {
	fn := module.FileNode{}
	session := S3s[account.Id]
	c := s3.New(session)
	if fileId == "" {
		return module.FileNode{
			FileId:     "",
			FileName:   "",
			FileSize:   0,
			IsFolder:   true,
			Path:       "/",
			LastOpTime: time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	}
	input := &s3.HeadObjectInput{
		Bucket: aws.String(account.RedirectUri),
		Key:    aws.String(fileId),
	}
	object, err := c.HeadObject(input)
	if err != nil {
		return fn, err
	}
	if strings.Contains(*object.ContentType, "x-directory") {
		fn.FileId = fileId
		fn.FileName = util.Base(strings.Trim(fileId, "/"))
		fn.CreateTime = "-"
		fn.FileType = ""
		fn.IsFolder = true
		fn.FileSize = 0
		fn.SizeFmt = "-"
	} else {
		fn.FileId = fileId
		fn.FileName = util.GetFileName(fileId)
		fn.CreateTime = "-"
		fn.IsFolder = false
		fn.FileType = strings.TrimLeft(filepath.Ext(fn.FileName), ".")
		fn.ViewType = util.GetViewType(fn.FileType)
		fn.FileSize = *object.ContentLength
		fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	}
	fn.Id = uuid.NewV4().String()
	fn.AccountId = account.Id
	fn.IsDelete = 1
	lastOpTime := *object.LastModified
	fn.LastOpTime = lastOpTime.In(time.Local).Format("2006-01-02 15:04:05")
	fn.Path = path
	fn.ParentPath = util.GetParentPath(path)
	return fn, err
}

func (s S3) ToFolderNode(object *s3.CommonPrefix) module.FileNode {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = *object.Prefix
	fn.FileName = util.Base(strings.Trim(*object.Prefix, "/"))
	fn.CreateTime = "-"
	fn.LastOpTime = "-"
	fn.IsDelete = 1
	fn.FileType = ""
	fn.IsFolder = true
	fn.FileSize = 0
	fn.SizeFmt = "-"
	return fn
}

func (s S3) ToFileNode(object *s3.Object) module.FileNode {
	fn := module.FileNode{}
	fn.Id = uuid.NewV4().String()
	fn.FileId = *object.Key
	fn.FileName = util.GetFileName(*object.Key)
	fn.CreateTime = "-"
	lastOpTime := *object.LastModified
	fn.LastOpTime = lastOpTime.In(time.Local).Format("2006-01-02 15:04:05")
	fn.IsDelete = 1
	fn.IsFolder = false
	fn.FileType = strings.TrimLeft(filepath.Ext(fn.FileName), ".")
	fn.ViewType = util.GetViewType(fn.FileType)
	fn.FileSize = *object.Size
	fn.SizeFmt = util.FormatFileSize(fn.FileSize)
	return fn
}

func (s S3) UploadFiles(account module.Account, parentFileId string, files []*module.UploadInfo, overwrite bool) (bool, interface{}, error) {
	session := S3s[account.Id]
	for _, file := range files {
		t1 := time.Now()
		uploader := s3manager.NewUploader(session)
		input := &s3manager.UploadInput{
			Bucket: aws.String(account.RedirectUri),
			Key:    aws.String(parentFileId + file.FileName),
			Body:   bytes.NewReader(file.Content),
		}
		_, err := uploader.Upload(input)
		if err != nil {
			log.Errorf("Unable to upload file:%s, err:%v", parentFileId+file.FileName, err)
		}
		log.Debugf("file:%s，upload success,timespan:%s", file.FileName, util.ShortDur(time.Now().Sub(t1)))
	}
	return true, "all files uploaded", nil
}

func (s S3) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	session := S3s[account.Id]
	c := s3.New(session)
	bucket := account.RedirectUri
	source := bucket + "/" + fileId
	item := util.GetParentPath(fileId)[1:len(util.GetParentPath(fileId))] + "/" + name
	if s.isFolder(fileId) {
		str := fileId[0 : len(fileId)-1]
		path := util.GetParentPath(str)[1:len(util.GetParentPath(str))]
		if path == "" {
			item = name + "/"
		} else {
			item = path + "/" + name + "/"
		}
	}
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(account.RedirectUri),
		CopySource: aws.String(url.PathEscape(source)),
		Key:        aws.String(item),
	}
	output, err := c.CopyObject(input)
	if err != nil {
		log.Errorf("Unable to copy item from %s to %s, %v", source, item, err)
		return false, "File rename error", err
	}
	err = c.WaitUntilObjectExists(&s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(item)})
	if err != nil {
		log.Errorf("Error occurred while waiting for item %s to be copied to %s, %v", source, item, err)
	}
	//copy successfully, remove source file
	s.Remove(account, fileId)
	log.Debug("File rename: ", output.String())
	return true, output, nil
}

func (s S3) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	session := S3s[account.Id]
	c := s3.New(session)
	output, err := c.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(account.RedirectUri), Key: aws.String(fileId)})
	if err != nil {
		log.Errorf("Unable to delete object %s from bucket %s, %v", fileId, account.RedirectUri, err)
		return false, "File remove error", err
	}
	err = c.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(account.RedirectUri),
		Key:    aws.String(fileId),
	})
	if err != nil {
		log.Errorf("Error occurred while waiting for object %s to be deleted, %v", fileId, err)
		return false, "File remove error", err
	}
	log.Debug("File remove success: ", output.String())
	return true, output, err
}

func (s S3) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	session := S3s[account.Id]
	c := s3.New(session)
	if parentFileId == "/" {
		parentFileId = ""
	}
	output, err := c.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(account.RedirectUri),
		Key:         aws.String(parentFileId + name + "/"),
		ContentType: aws.String("application/x-directory; charset=UTF-8"),
	})
	log.Debug("Dir create: ", output.String())
	if err == nil {
		return true, "Dir create success", nil
	}
	return false, "Dir create error", err
}

func (s S3) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	result, _, err := s.Copy(account, fileId, targetFileId, overwrite)
	if result {
		//remove original file
		s.Remove(account, fileId)
		return true, "File move success", err
	}
	return false, "File move error", err
}

func (s S3) isFolder(fileId string) bool {
	if fileId == "" || strings.HasSuffix(fileId, "/") {
		return true
	}
	return false
}

func (s S3) GetNewFileName(fileName string, isFolder bool) string {
	if isFolder {
		return fileName + "_" + util.GetRandomStr(4) + "/"
	}
	fileNameAll := path.Base(fileName)
	fileSuffix := path.Ext(fileName)
	filePrefix := fileNameAll[0 : len(fileNameAll)-len(fileSuffix)]
	return filePrefix + "_" + util.GetRandomStr(4) + fileSuffix
}

func (s S3) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	if strings.HasPrefix(fileId, targetFileId) && overwrite {
		return false, "File copy failed, file conflict", nil
	}
	_, item, err := s.SingleCopy(account, fileId, targetFileId, overwrite)
	if err == nil {
		if s.isFolder(fileId) {
			//copy children files
			s.loopCopyFiles(account, fileId, item, overwrite)
		}
	}
	return true, "File copy success", nil
}

func (s S3) SingleCopy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, string, error) {
	session := S3s[account.Id]
	c := s3.New(session)
	bucket := account.RedirectUri
	source := bucket + "/" + fileId
	sourceFileName := path.Base(fileId)
	item := targetFileId + sourceFileName
	if s.isFolder(fileId) {
		//folder
		item = targetFileId + sourceFileName + "/"
	}
	if overwrite {
		s.Remove(account, item)
	} else {
		_, err := s.File(account, item, "")
		if err == nil {
			item = targetFileId + s.GetNewFileName(sourceFileName, s.isFolder(fileId))
		}
	}
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(account.RedirectUri),
		CopySource: aws.String(url.PathEscape(source)),
		Key:        aws.String(item),
	}
	if s.isFolder(fileId) {
		input.ContentType = aws.String("application/x-directory; charset=UTF-8")
	}
	output, err := c.CopyObject(input)
	if err != nil {
		log.Errorf("Unable to copy item from %s to %s, %v", source, item, err)
		return false, "File copy error", err
	}
	err = c.WaitUntilObjectExists(&s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(item)})
	if err != nil {
		log.Errorf("Error occurred while waiting for item %s to be copied to %s, %v", source, item, err)
	}
	log.Debug("File copy: ", output.String())
	return true, item, nil
}

func (s S3) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	if fileId == "" {
		return "", nil
	}
	session := S3s[account.Id]
	c := s3.New(session)
	req, _ := c.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(account.RedirectUri),
		Key:    aws.String(fileId),
	})
	urlStr, err := req.Presign(24 * time.Hour)
	if err != nil {
		return "", err
	}
	return urlStr, nil
}

func (s S3) GetSpaceSzie(account module.Account) (int64, int64) {
	return 0, 0
}

func (s S3) loopCopyFiles(account module.Account, fileId, targetFileId string, overwrite bool) {
	fns, _ := s.Files(account, fileId, "/", "", "")
	for _, fn := range fns {
		if fn.IsFolder {
			s.loopCopyFiles(account, fn.FileId, targetFileId+fn.FileName+"/", overwrite)
		}
		s.SingleCopy(account, fn.FileId, targetFileId, overwrite)
	}
}
