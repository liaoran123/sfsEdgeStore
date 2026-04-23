package resource

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"sfsEdgeStore/config"

	"github.com/shirou/gopsutil/v4/process"
)

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryMB        float64 `json:"memory_mb"`
	CPUPercent      float64 `json:"cpu_percent"`
	Goroutines      int     `json:"goroutines"`
	Timestamp       int64   `json:"timestamp"`
	MemoryLimitMB   float64 `json:"memory_limit_mb"`
	CPULimitPercent float64 `json:"cpu_limit_percent"`
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	config    *config.Config
	monitor   MonitorInterface
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.Mutex
	lastUsage ResourceUsage
	alertSent map[string]bool
}

// MonitorInterface 监控接口
type MonitorInterface interface {
	RecordError(errorType, message string)
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(cfg *config.Config, monitor MonitorInterface) *ResourceMonitor {
	return &ResourceMonitor{
		config:    cfg,
		monitor:   monitor,
		stopChan:  make(chan struct{}),
		alertSent: make(map[string]bool),
	}
}

// Start 启动资源监控
func (rm *ResourceMonitor) Start() error {
	if !rm.config.EnableResourceMonitoring {
		log.Println("Resource monitoring is disabled")
		return nil
	}

	if rm.isRunning {
		log.Println("Resource monitor is already running")
		return nil
	}

	rm.isRunning = true
	go rm.monitorLoop()
	log.Println("Resource monitor started")
	return nil
}

// Stop 停止资源监控
func (rm *ResourceMonitor) Stop() {
	if !rm.isRunning {
		return
	}
	close(rm.stopChan)
	rm.isRunning = false
	log.Println("Resource monitor stopped")
}

// GetCurrentUsage 获取当前资源使用情况
func (rm *ResourceMonitor) GetCurrentUsage() ResourceUsage {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	return rm.lastUsage
}

// monitorLoop 监控循环
func (rm *ResourceMonitor) monitorLoop() {
	intervalSeconds := rm.config.ResourceMonitorInterval
	if intervalSeconds <= 0 {
		intervalSeconds = 10 // 默认10秒
	}
	interval := time.Duration(intervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.checkResources()
		}
	}
}

// checkResources 检查资源使用情况
func (rm *ResourceMonitor) checkResources() {
	usage := rm.collectUsage()

	rm.mutex.Lock()
	rm.lastUsage = usage
	rm.mutex.Unlock()

	rm.checkMemory(usage)
	rm.checkCPU(usage)

	log.Printf("Resource usage - Memory: %.2f MB (limit: %.2f MB), CPU: %.2f%% (limit: %.2f%%), Goroutines: %d",
		usage.MemoryMB, usage.MemoryLimitMB,
		usage.CPUPercent, usage.CPULimitPercent,
		usage.Goroutines)
}

// collectUsage 收集资源使用数据
func (rm *ResourceMonitor) collectUsage() ResourceUsage {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memoryMB := float64(memStats.Alloc) / 1024 / 1024
	cpuPercent := rm.getCPUPercent()

	return ResourceUsage{
		MemoryMB:        memoryMB,
		CPUPercent:      cpuPercent,
		Goroutines:      runtime.NumGoroutine(),
		Timestamp:       time.Now().Unix(),
		MemoryLimitMB:   rm.config.MaxMemoryMB,
		CPULimitPercent: rm.config.MaxCPUPercent,
	}
}

// 全局变量，用于存储上一次的CPU使用率
var lastCPUPercent float64
var cpuInitialized bool

// getCPUPercent 获取 CPU 使用率
func (rm *ResourceMonitor) getCPUPercent() float64 {
	percent, err := cpu_percent()
	if err != nil {
		return 0
	}
	
	// 限制CPU使用率在合理范围内
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}
	
	return percent
}

// cpu_percent 获取 CPU 使用率（非阻塞方式）
func cpu_percent() (float64, error) {
	// 获取当前进程
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		return 0, err
	}

	// 第一次调用时阻塞采样
	if !cpuInitialized {
		_, err = proc.CPUPercent()
		if err != nil {
			return 0, err
		}
		cpuInitialized = true
		time.Sleep(100 * time.Millisecond) // 短暂等待确保采样
	}

	// 获取进程 CPU 使用率
	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		return 0, err
	}

	// 存储上次值
	lastCPUPercent = cpuPercent
	return cpuPercent, nil
}

// checkMemory 检查内存使用
func (rm *ResourceMonitor) checkMemory(usage ResourceUsage) {
	if usage.MemoryMB > usage.MemoryLimitMB {
		message := fmt.Sprintf("Memory usage exceeds limit: %.2f MB > %.2f MB",
			usage.MemoryMB, usage.MemoryLimitMB)
		log.Printf("[WARNING] %s", message)

		// 尝试释放内存
		rm.tryFreeMemory()
	}
}

// checkCPU 检查 CPU 使用
func (rm *ResourceMonitor) checkCPU(usage ResourceUsage) {
	if usage.CPUPercent > usage.CPULimitPercent {
		message := fmt.Sprintf("CPU usage exceeds limit: %.2f%% > %.2f%%",
			usage.CPUPercent, usage.CPULimitPercent)
		log.Printf("[WARNING] %s", message)

		// 尝试调整资源使用
		rm.adjustResourceUsage()
	}
}

// tryFreeMemory 尝试释放内存
func (rm *ResourceMonitor) tryFreeMemory() {
	log.Println("Attempting to free memory...")

	// 触发 GC
	runtime.GC()

	// 再次触发 GC 以释放更多内存
	runtime.GC()

	// 释放到操作系统
	debug.FreeOSMemory()

	log.Println("Memory cleanup completed")
}

// adjustResourceUsage 调整资源使用
func (rm *ResourceMonitor) adjustResourceUsage() {
	log.Println("Adjusting resource usage...")

	// 这里可以实现资源使用调整逻辑
	// 例如：
	// - 减少批量处理大小
	// - 降低并发数
	// - 调整同步间隔

	log.Println("Resource usage adjustment completed")
}
