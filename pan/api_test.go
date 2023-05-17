package pan

import (
	_ "github.com/px-org/PanIndex/pan/123"
	_ "github.com/px-org/PanIndex/pan/ali"
	_ "github.com/px-org/PanIndex/pan/alishare"
	"github.com/px-org/PanIndex/pan/base"
	_ "github.com/px-org/PanIndex/pan/cloud189"
	_ "github.com/px-org/PanIndex/pan/ftp"
	_ "github.com/px-org/PanIndex/pan/googledrive"
	_ "github.com/px-org/PanIndex/pan/native"
	_ "github.com/px-org/PanIndex/pan/onedrive"
	_ "github.com/px-org/PanIndex/pan/pikpak"
	_ "github.com/px-org/PanIndex/pan/s3"
	_ "github.com/px-org/PanIndex/pan/teambition"
	_ "github.com/px-org/PanIndex/pan/webdav"
	_ "github.com/px-org/PanIndex/pan/yun139"
	"os"
	"testing"
)

func TestApi(t *testing.T) {
	t.Run(os.Getenv("MODE"), func(t *testing.T) {
		base.SimpleTest()
	})
}
