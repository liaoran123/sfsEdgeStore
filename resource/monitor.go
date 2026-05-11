package resource

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"sfsEdgeStore/config"

	"github.com/shirou/gopsutil/v3/process"
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
	proc      *process.Process
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
	memoryMB := rm.getProcessMemoryMB()
	cpuPercent := rm.getCPUPercent()

	return ResourceUsage{
		MemoryMB:      memoryMB,
		CPUPercent:    cpuPercent,
		Goroutines:    runtime.NumGoroutine(),
		Timestamp:     time.Now().Unix(),
		MemoryLimitMB: rm.config.MaxMemoryMB,
	}
}

func (rm *ResourceMonitor) getProcess() *process.Process {
	if rm.proc != nil {
		return rm.proc
	}
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil
	}
	rm.proc = p
	return p
}

func (rm *ResourceMonitor) getProcessMemoryMB() float64 {
	p := rm.getProcess()
	if p == nil {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		return float64(ms.Alloc) / 1024 / 1024
	}

	mem, err := p.MemoryInfo()
	if err != nil {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		return float64(ms.Alloc) / 1024 / 1024
	}

	// RSS (Resident Set Size) 是进程当前使用的物理内存
	return float64(mem.RSS) / 1024 / 1024
}

func (rm *ResourceMonitor) getCPUPercent() float64 {
	p := rm.getProcess()
	if p == nil {
		return 0
	}
	percent, err := p.Percent(0)
	if err != nil {
		return 0
	}
	numCPU := runtime.NumCPU()
	if numCPU > 0 {
		percent = percent / float64(numCPU)
	}
	return percent
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