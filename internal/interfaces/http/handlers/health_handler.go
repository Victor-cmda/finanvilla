package handlers

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type HealthStatus struct {
	Status      string    `json:"status"`
	Time        time.Time `json:"timestamp"`
	Environment string    `json:"environment"`

	System struct {
		CPUUsage    float64 `json:"cpu_usage"`
		MemoryUsage struct {
			Total     uint64  `json:"total"`
			Used      uint64  `json:"used"`
			Free      uint64  `json:"free"`
			UsagePerc float64 `json:"usage_percentage"`
		} `json:"memory_usage"`
		DiskUsage struct {
			Total     uint64  `json:"total"`
			Used      uint64  `json:"used"`
			Free      uint64  `json:"free"`
			UsagePerc float64 `json:"usage_percentage"`
		} `json:"disk_usage"`
	} `json:"system"`

	Application struct {
		Version    string `json:"version"`
		GoVersion  string `json:"go_version"`
		Goroutines int    `json:"goroutines"`
		StartTime  string `json:"start_time"`
		UpTime     string `json:"uptime"`
	} `json:"application"`
}

type HealthHandler struct {
	startTime   time.Time
	environment string
	appVersion  string
}

func NewHealthHandler(environment, appVersion string) *HealthHandler {
	return &HealthHandler{
		startTime:   time.Now(),
		environment: environment,
		appVersion:  appVersion,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	health := HealthStatus{
		Status:      "healthy",
		Time:        time.Now(),
		Environment: h.environment,
	}

	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		health.System.CPUUsage = cpuPercent[0]
	}

	if memInfo, err := mem.VirtualMemory(); err == nil {
		health.System.MemoryUsage.Total = memInfo.Total
		health.System.MemoryUsage.Used = memInfo.Used
		health.System.MemoryUsage.Free = memInfo.Free
		health.System.MemoryUsage.UsagePerc = memInfo.UsedPercent
	}

	if diskInfo, err := disk.Usage("/"); err == nil {
		health.System.DiskUsage.Total = diskInfo.Total
		health.System.DiskUsage.Used = diskInfo.Used
		health.System.DiskUsage.Free = diskInfo.Free
		health.System.DiskUsage.UsagePerc = diskInfo.UsedPercent
	}

	health.Application.Version = h.appVersion
	health.Application.GoVersion = runtime.Version()
	health.Application.Goroutines = runtime.NumGoroutine()
	health.Application.StartTime = h.startTime.Format(time.RFC3339)
	health.Application.UpTime = time.Since(h.startTime).String()

	statusCode := 200
	if health.Status != "healthy" {
		statusCode = 503
	}

	c.JSON(statusCode, health)
}
