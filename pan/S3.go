package pan

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/libsgh/PanIndex/module"
	"github.com/libsgh/PanIndex/util"
	uuid "github.com/satori/go.uuid"
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
		Credentials:      credentials.NewStaticCredentials(account.User, account.Password, ""),
		Endpoint:         aws.String(account.ApiUrl),
		Region:           aws.String(account.SiteId),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	}
	sess, err := session.NewSession(cfg)
	S3s[account.Id] = sess
	return "ok", err
}

func (s S3) IsLogin(account *module.Account) bool {
	//TODO implement me
	panic("implement me")
}

func (s S3) Files(account module.Account, fileId, path, sortColumn, sortOrder string) ([]module.FileNode, error) {
	fileNodes := make([]module.FileNode, 0)
	session := S3s[account.Id]
	c := s3.New(session)
	marker := ""
	for {
		input := &s3.ListObjectsInput{
			Bucket:    aws.String(account.RedirectUri),
			Marker:    &marker,
			Prefix:    &fileId,
			Delimiter: aws.String("/"),
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
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (s S3) Rename(account module.Account, fileId, name string) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s S3) Remove(account module.Account, fileId string) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s S3) Mkdir(account module.Account, parentFileId, name string) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s S3) Move(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s S3) Copy(account module.Account, fileId, targetFileId string, overwrite bool) (bool, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s S3) GetDownloadUrl(account module.Account, fileId string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s S3) GetSpaceSzie(account module.Account) (int64, int64) {
	//TODO implement me
	panic("implement me")
}
