package service

import (
	"os"
	"runtime"
	"time"

	"github.com/px-org/PanIndex/dao"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/util"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var processStartTime = time.Now()

type DashboardOverview struct {
	TodayDownloads  int64 `json:"today_downloads"`
	TotalDownloads  int64 `json:"total_downloads"`
	AccountCount    int   `json:"account_count"`
	AccountOk       int   `json:"account_ok"`
	CacheFiles      int   `json:"cache_files"`
	CacheFile       int   `json:"cache_file"`
	CacheUrl        int   `json:"cache_url"`
	CacheDbNodes    int64 `json:"cache_db_nodes"`
	ShareCount      int64 `json:"share_count"`
	PwdCount        int   `json:"pwd_count"`
	HideCount       int   `json:"hide_count"`
	Syncing         bool  `json:"syncing"`
	StatsEnabled    bool  `json:"stats_enabled"`
	Version         string `json:"version"`
	GoVersion       string `json:"go_version"`
	BuildTime       string `json:"build_time"`
	GitCommit       string `json:"git_commit"`
	UptimeSeconds   int64  `json:"uptime_seconds"`
}

type DashboardSystem struct {
	CPUPercent     float64 `json:"cpu_percent"`
	CPUCores       int     `json:"cpu_cores"`
	MemTotal       uint64  `json:"mem_total"`
	MemUsed        uint64  `json:"mem_used"`
	MemPercent     float64 `json:"mem_percent"`
	DiskPath       string  `json:"disk_path"`
	DiskTotal      uint64  `json:"disk_total"`
	DiskUsed       uint64  `json:"disk_used"`
	DiskPercent    float64 `json:"disk_percent"`
	Goroutines     int     `json:"goroutines"`
	HeapAlloc      uint64  `json:"heap_alloc"`
	HeapSys        uint64  `json:"heap_sys"`
	GCNum          uint32  `json:"gc_num"`
	GCPauseTotalMs float64 `json:"gc_pause_total_ms"`
	LastGC         string  `json:"last_gc"`
	CronJobs       int     `json:"cron_jobs"`
	SyncStatus     int     `json:"sync_status"`
}

type DashboardData struct {
	Overview DashboardOverview          `json:"overview"`
	Trend    []module.DownloadDaily     `json:"trend"`
	TopFiles []module.DownloadStatistics `json:"top_files"`
	System   DashboardSystem            `json:"system"`
}

func RecordDownload(ac module.Account, fileNode module.FileNode) {
	dao.SyncDownloadInfo(ac, fileNode)
}

func GetDashboardData() DashboardData {
	cfg := module.GloablConfig
	accountOk := 0
	for _, ac := range cfg.Accounts {
		if ac.CookieStatus == 2 || ac.Mode == "native" || ac.Mode == "ftp" || ac.Mode == "webdav" || ac.Mode == "s3" {
			accountOk++
		}
	}

	overview := DashboardOverview{
		TodayDownloads: dao.GetTodayDownloadCount(),
		TotalDownloads: dao.GetTotalDownloadCount(),
		AccountCount:   len(cfg.Accounts),
		AccountOk:      accountOk,
		CacheFiles:     len(FilesCache.Keys(false)),
		CacheFile:      len(FileCache.Keys(false)),
		CacheUrl:       len(UrlCache.Keys(false)),
		CacheDbNodes:   dao.CountFileNodes(),
		ShareCount:     dao.CountShareInfos(),
		PwdCount:       len(cfg.PwdFiles),
		HideCount:      len(cfg.HideFiles),
		Syncing:        dao.SYNC_STATUS == 1,
		StatsEnabled:   cfg.EnableDownloadStatistics == "1",
		Version:        module.VERSION,
		GoVersion:      module.GO_VERSION,
		BuildTime:      module.BUILD_TIME,
		GitCommit:      module.GIT_COMMIT_SHA,
		UptimeSeconds:  int64(time.Since(processStartTime).Seconds()),
	}

	return DashboardData{
		Overview: overview,
		Trend:    dao.GetDownloadTrend(14),
		TopFiles: dao.GetTopDownloadFiles(20),
		System:   collectSystemStats(),
	}
}

func collectSystemStats() DashboardSystem {
	sys := DashboardSystem{
		Goroutines: runtime.NumGoroutine(),
		SyncStatus: dao.SYNC_STATUS,
		CronJobs:   len(util.CacheCronMap),
	}
	if util.Cron != nil {
		sys.CronJobs = len(util.Cron.Entries())
	}

	if percents, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(percents) > 0 {
		sys.CPUPercent = percents[0]
	}
	if cores, err := cpu.Counts(true); err == nil {
		sys.CPUCores = cores
	}

	if vm, err := mem.VirtualMemory(); err == nil {
		sys.MemTotal = vm.Total
		sys.MemUsed = vm.Used
		sys.MemPercent = vm.UsedPercent
	}

	diskPath, _ := os.Getwd()
	sys.DiskPath = diskPath
	if du, err := disk.Usage(diskPath); err == nil {
		sys.DiskTotal = du.Total
		sys.DiskUsed = du.Used
		sys.DiskPercent = du.UsedPercent
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	sys.HeapAlloc = ms.HeapAlloc
	sys.HeapSys = ms.HeapSys
	sys.GCNum = ms.NumGC
	sys.GCPauseTotalMs = float64(ms.PauseTotalNs) / 1e6
	if ms.LastGC > 0 {
		sys.LastGC = time.Unix(0, int64(ms.LastGC)).Format("2006-01-02 15:04:05")
	} else {
		sys.LastGC = "-"
	}
	return sys
}
