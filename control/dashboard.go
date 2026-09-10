package control

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/service"
)

// GetDashboard returns admin homepage metrics.
func GetDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    "success",
		"data":   service.GetDashboardData(),
	})
}

func recordDownloadStats(ac module.Account, fileNode module.FileNode) {
	if module.GloablConfig.EnableDownloadStatistics != "1" {
		return
	}
	go service.RecordDownload(ac, fileNode)
}
