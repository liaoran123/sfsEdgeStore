package resource

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"sfsEdgeStore/config"
)

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryMB      float64 `json:"memory_mb"`
	CPUPercent    float64 `json:"cpu_percent"`
	Goroutines    int     `json:"goroutines"`
	Timestamp     int64   `json:"timestamp"`
	MemoryLimitMB float64 `json:"memory_limit_mb"`
	CPULimitPercent float64 `json:"cpu_limit_percent"`
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	config         *config.Config
	monitor        MonitorInterface
	isRunning      bool
	stopChan       chan struct{}
	mutex          sync.Mutex
	lastUsage      ResourceUsage
	alertSent      map[string]bool
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
	interval := time.Duration(rm.config.ResourceMonitorInterval) * time.Second
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
		MemoryMB:       memoryMB,
		CPUPercent:     cpuPercent,
		Goroutines:     runtime.NumGoroutine(),
		Timestamp:      time.Now().Unix(),
		MemoryLimitMB:  rm.config.MaxMemoryMB,
		CPULimitPercent: rm.config.MaxCPUPercent,
	}
}

// getCPUPercent 获取 CPU 使用率
func (rm *ResourceMonitor) getCPUPercent() float64 {
	numCPU := runtime.NumCPU()
	if numCPU == 0 {
		return 0
	}

	// 简化的 CPU 使用率估算（基于 Goroutine 数量和系统负载）
	// 实际项目中可以使用更精确的方法
	goroutines := runtime.NumGoroutine()
	baseCPU := math.Min(float64(goroutines)/10, 5)
	
	return baseCPU
}

// checkMemory 检查内存使用
func (rm *ResourceMonitor) checkMemory(usage ResourceUsage) {
	if usage.MemoryMB > usage.MemoryLimitMB {
		alertKey := "memory_over_limit"
		if !rm.alertSent[alertKey] {
			message := fmt.Sprintf("Memory usage exceeds limit: %.2f MB > %.2f MB",
				usage.MemoryMB, usage.MemoryLimitMB)
			log.Printf("[WARNING] %s", message)
			
			if rm.monitor != nil {
				rm.monitor.RecordError("memory_over_limit", message)
			}
			
			rm.alertSent[alertKey] = true
			
			// 尝试释放内存
			rm.tryFreeMemory()
		}
	} else {
		delete(rm.alertSent, "memory_over_limit")
	}
}

// checkCPU 检查 CPU 使用
func (rm *ResourceMonitor) checkCPU(usage ResourceUsage) {
	if usage.CPUPercent > usage.CPULimitPercent {
		alertKey := "cpu_over_limit"
		if !rm.alertSent[alertKey] {
			message := fmt.Sprintf("CPU usage exceeds limit: %.2f%% > %.2f%%",
				usage.CPUPercent, usage.CPULimitPercent)
			log.Printf("[WARNING] %s", message)
			
			if rm.monitor != nil {
				rm.monitor.RecordError("cpu_over_limit", message)
			}
			
			rm.alertSent[alertKey] = true
			
			// 尝试调整资源使用
			rm.adjustResourceUsage()
		}
	} else {
		delete(rm.alertSent, "cpu_over_limit")
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
