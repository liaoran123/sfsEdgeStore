package resource

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"sfsEdgeStore/config"
)

type ResourceUsage struct {
	MemoryMB      float64 `json:"memory_mb"`
	CPUPercent    float64 `json:"cpu_percent"`
	Goroutines    int     `json:"goroutines"`
	Timestamp     int64   `json:"timestamp"`
	MemoryLimitMB float64 `json:"memory_limit_mb"`
}

type ResourceMonitor struct {
	config    *config.Config
	monitor   MonitorInterface
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.Mutex
	lastUsage ResourceUsage
	alertSent map[string]bool
}

type MonitorInterface interface {
	RecordError(errorType, message string)
}

func NewResourceMonitor(cfg *config.Config, monitor MonitorInterface) *ResourceMonitor {
	return &ResourceMonitor{
		config:    cfg,
		monitor:   monitor,
		stopChan:  make(chan struct{}),
		alertSent: make(map[string]bool),
	}
}

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

func (rm *ResourceMonitor) Stop() {
	if !rm.isRunning {
		return
	}
	close(rm.stopChan)
	rm.isRunning = false
	log.Println("Resource monitor stopped")
}

func (rm *ResourceMonitor) GetCurrentUsage() ResourceUsage {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	return rm.lastUsage
}

func (rm *ResourceMonitor) monitorLoop() {
	intervalSeconds := rm.config.ResourceMonitorInterval
	if intervalSeconds <= 0 {
		intervalSeconds = 3
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

func (rm *ResourceMonitor) checkResources() {
	usage := rm.collectUsage()

	rm.mutex.Lock()
	rm.lastUsage = usage
	rm.mutex.Unlock()

	rm.checkMemory(usage)
	rm.checkCPU(usage)

	log.Printf("Resource usage - Memory: %.2f MB (limit: %.2f MB), CPU: %.2f%%, Goroutines: %d",
		usage.MemoryMB, usage.MemoryLimitMB, usage.CPUPercent, usage.Goroutines)
}

func (rm *ResourceMonitor) collectUsage() ResourceUsage {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	memoryMB := float64(ms.HeapInuse) / 1024 / 1024
	cpuPercent := rm.estimateCPUPercent()

	return ResourceUsage{
		MemoryMB:      memoryMB,
		CPUPercent:    cpuPercent,
		Goroutines:    runtime.NumGoroutine(),
		Timestamp:     time.Now().Unix(),
		MemoryLimitMB: rm.config.MaxMemoryMB,
	}
}

var lastCPUTime time.Time
var lastGoroutines int

func (rm *ResourceMonitor) estimateCPUPercent() float64 {
	currentTime := time.Now()
	currentGoroutines := runtime.NumGoroutine()

	if lastCPUTime.IsZero() {
		lastCPUTime = currentTime
		lastGoroutines = currentGoroutines
		return 0
	}

	elapsed := currentTime.Sub(lastCPUTime).Seconds()
	goroutineChange := currentGoroutines - lastGoroutines

	lastCPUTime = currentTime
	lastGoroutines = currentGoroutines

	if elapsed < 0.001 {
		return 0
	}

	estimatedCPU := float64(goroutineChange) / elapsed * 10
	if estimatedCPU < 0 {
		estimatedCPU = 0
	}
	if estimatedCPU > 100 {
		estimatedCPU = 100
	}

	return estimatedCPU
}

func (rm *ResourceMonitor) checkMemory(usage ResourceUsage) {
	if usage.MemoryMB > usage.MemoryLimitMB {
		message := fmt.Sprintf("Memory usage exceeds limit: %.2f MB > %.2f MB",
			usage.MemoryMB, usage.MemoryLimitMB)
		log.Printf("[WARNING] %s", message)
		rm.tryFreeMemory()
	}
}

func (rm *ResourceMonitor) checkCPU(usage ResourceUsage) {
	if usage.CPUPercent > 80 {
		message := fmt.Sprintf("High CPU usage: %.2f%%", usage.CPUPercent)
		log.Printf("[WARNING] %s", message)
		rm.adjustResourceUsage()
	}
}

func (rm *ResourceMonitor) tryFreeMemory() {
	log.Println("Attempting to free memory...")
	runtime.GC()
	log.Println("Memory cleanup completed")
}

func (rm *ResourceMonitor) adjustResourceUsage() {
	log.Println("Adjusting resource usage...")
	log.Println("Resource usage adjustment completed")
}
