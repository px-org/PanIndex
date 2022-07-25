package pan

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/libsgh/PanIndex/module"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestS3(t *testing.T) {
	t.Run("S3", func(t *testing.T) {
		account := &module.Account{}
		account.Id = uuid.NewV4().String()
		account.User = "AKIATV6EP2HEGFOV5ZZ7"
		account.Password = "68xiOj3kOzFmIwU6B9ZWAmWR/fgItEg0sA9teTex"
		account.RedirectUri = "panindexbucket"
		account.ApiUrl = "s3.amazonaws.com"
		account.SiteId = "us-east-1"
		account.Mode = "s3"
		account.RootId = ""
		p, _ := GetPan(account.Mode)
		result, err := p.AuthLogin(account)
		log.Info(result, err)
		fileNodes, err := p.Files(*account, account.RootId, "/", "", "")
		files, _ := jsoniter.MarshalToString(fileNodes)
		log.Info(files, err)
		fmt.Println(files)
	})
}
