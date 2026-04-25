# 资源监控器 (Resource Monitor) 技术文档

## 1. 概述

资源监控器是sfsEdgeStore系统中负责监控系统资源使用情况的核心组件。它实时监控进程的内存使用、CPU使用率、Goroutine数量等关键指标，并在资源使用超过预设阈值时触发告警和自我调整机制。

### 1.1 主要功能

- **实时资源监控**：持续监控内存、CPU、Goroutine等系统资源
- **可配置的监控间隔**：支持自定义监控检查间隔
- **资源限制检查**：当资源使用超过配置的限制时触发告警
- **自动内存释放**：当内存使用过高时自动尝试释放内存
- **自我调整机制**：当CPU使用过高时触发资源使用调整
- **线程安全**：使用互斥锁保护共享数据

## 2. 架构设计

### 2.1 核心组件

```
ResourceMonitor
├── config          *config.Config     # 配置对象
├── monitor         MonitorInterface    # 监控接口（用于记录错误）
├── isRunning       bool               # 运行状态标志
├── stopChan        chan struct{}      # 停止信号通道
├── mutex           sync.Mutex         # 互斥锁
├── lastUsage       ResourceUsage      # 上次收集的资源使用情况
└── alertSent       map[string]bool    # 告警发送记录（防止重复告警）

ResourceUsage
├── MemoryMB        float64           # 当前内存使用（MB）
├── CPUPercent      float64           # 当前CPU使用率（%）
├── Goroutines      int               # 当前Goroutine数量
├── Timestamp       int64             # 采集时间戳
├── MemoryLimitMB   float64           # 内存限制（MB）
└── CPULimitPercent float64           # CPU使用率限制（%）
```

### 2.2 工作流程

```
Start()
    ↓
monitorLoop() [后台goroutine]
    ↓
定期触发 checkResources()
    ↓
collectUsage() → 获取资源数据
    ↓
检查内存和CPU是否超限
    ↓
超限则触发 tryFreeMemory() 或 adjustResourceUsage()
    ↓
记录资源使用日志
```

## 3. 数据结构

### 3.1 ResourceUsage 结构体

```go
type ResourceUsage struct {
    MemoryMB        float64 `json:"memory_mb"`        // 当前内存使用（MB）
    CPUPercent      float64 `json:"cpu_percent"`       // 当前CPU使用率（%）
    Goroutines      int     `json:"goroutines"`        // 当前Goroutine数量
    Timestamp       int64   `json:"timestamp"`         // 采集时间戳
    MemoryLimitMB   float64 `json:"memory_limit_mb"`   // 内存限制（MB）
    CPULimitPercent float64 `json:"cpu_limit_percent"` // CPU使用率限制（%）
}
```

### 3.2 ResourceMonitor 结构体

```go
type ResourceMonitor struct {
    config    *config.Config     // 配置对象
    monitor   MonitorInterface   // 监控接口
    isRunning bool              // 运行状态标志
    stopChan  chan struct{}      // 停止信号通道
    mutex     sync.Mutex         // 互斥锁
    lastUsage ResourceUsage      // 上次收集的资源使用情况
    alertSent map[string]bool   // 告警发送记录
}
```

### 3.3 MonitorInterface 接口

```go
type MonitorInterface interface {
    RecordError(errorType, message string)
}
```

## 4. 函数说明

### 4.1 构造函数

#### NewResourceMonitor
```go
func NewResourceMonitor(cfg *config.Config, monitor MonitorInterface) *ResourceMonitor
```

**功能**：创建并初始化资源监控器实例

**参数**：
- `cfg`: 配置对象
- `monitor`: 监控接口（用于记录错误）

**返回值**：初始化好的ResourceMonitor指针

---

### 4.2 生命周期管理

#### Start
```go
func (rm *ResourceMonitor) Start() error
```

**功能**：启动资源监控

**处理逻辑**：
1. 检查是否启用资源监控，未启用则直接返回
2. 检查是否已在运行，避免重复启动
3. 启动后台监控循环goroutine

**返回值**：错误对象（通常为nil）

#### Stop
```go
func (rm *ResourceMonitor) Stop()
```

**功能**：停止资源监控

**处理逻辑**：
1. 检查是否在运行
2. 发送停止信号到stopChan
3. 更新运行状态标志

---

### 4.3 数据获取

#### GetCurrentUsage
```go
func (rm *ResourceMonitor) GetCurrentUsage() ResourceUsage
```

**功能**：获取当前资源使用情况

**返回值**：最近一次采集的ResourceUsage对象（线程安全）

---

### 4.4 核心监控逻辑

#### monitorLoop
```go
func (rm *ResourceMonitor) monitorLoop()
```

**功能**：后台监控循环，持续定期检查资源使用情况

**处理逻辑**：
1. 根据配置创建定时器（默认10秒间隔）
2. 循环等待定时器触发或停止信号
3. 定时器触发时执行checkResources

**注意**：此函数运行在独立的goroutine中

---

#### checkResources
```go
func (rm *ResourceMonitor) checkResources()
```

**功能**：检查资源使用情况的主函数

**处理逻辑**：
1. 调用collectUsage收集当前资源使用数据
2. 更新lastUsage（线程安全）
3. 调用checkMemory检查内存使用
4. 调用checkCPU检查CPU使用
5. 记录资源使用日志

---

### 4.5 资源数据采集

#### collectUsage
```go
func (rm *ResourceMonitor) collectUsage() ResourceUsage
```

**功能**：收集当前资源使用数据

**返回值**：包含以下字段的ResourceUsage对象：
- MemoryMB: 进程实际物理内存
- CPUPercent: CPU使用率
- Goroutines: 当前Goroutine数量
- Timestamp: 当前时间戳
- MemoryLimitMB: 配置的内存限制
- CPULimitPercent: 配置的CPU限制

#### getProcessMemoryMB
```go
func (rm *ResourceMonitor) getProcessMemoryMB() float64
```

**功能**：获取进程的实际物理内存（Working Set）

**处理逻辑**：
1. 获取当前进程PID
2. 使用gopsutil库获取进程内存信息
3. 返回RSS（Resident Set Size），单位MB
4. 如果获取失败，回退到使用runtime.MemStats.Alloc

**返回值**：进程物理内存使用量（MB）

#### getCPUPercent
```go
func (rm *ResourceMonitor) getCPUPercent() float64
```

**功能**：获取CPU使用率

**处理逻辑**：
1. 调用cpu_percent获取原始CPU使用率
2. 限制CPU使用率在0-100%范围内

**返回值**：CPU使用率百分比

#### cpu_percent
```go
func cpu_percent() (float64, error)
```

**功能**：获取进程CPU使用率（非阻塞方式）

**处理逻辑**：
1. 第一次调用时进行阻塞采样
2. 短暂等待确保采样准确性
3. 返回进程CPU使用率

**返回值**：
- CPU使用率百分比
- 错误对象

---

### 4.6 资源检查和告警

#### checkMemory
```go
func (rm *ResourceMonitor) checkMemory(usage ResourceUsage)
```

**功能**：检查内存使用是否超过限制

**处理逻辑**：
1. 比较MemoryMB和MemoryLimitMB
2. 如果超过限制：
   - 记录警告日志
   - 调用tryFreeMemory尝试释放内存

#### checkCPU
```go
func (rm *ResourceMonitor) checkCPU(usage ResourceUsage)
```

**功能**：检查CPU使用是否超过限制

**处理逻辑**：
1. 比较CPUPercent和CPULimitPercent
2. 如果超过限制：
   - 记录警告日志
   - 调用adjustResourceUsage尝试调整资源使用

---

### 4.7 资源调整

#### tryFreeMemory
```go
func (rm *ResourceMonitor) tryFreeMemory()
```

**功能**：尝试释放内存

**处理逻辑**：
1. 触发Go运行时垃圾回收（GC）
2. 再次触发GC以释放更多内存
3. 调用debug.FreeOSMemory将内存释放回操作系统
4. 记录内存清理完成日志

#### adjustResourceUsage
```go
func (rm *ResourceMonitor) adjustResourceUsage()
```

**功能**：调整资源使用

**处理逻辑**：
1. 记录调整开始日志
2. 这里可以实现资源使用调整逻辑
   - 减少批量处理大小
   - 降低并发数
   - 调整同步间隔
3. 记录调整完成日志

**注意**：当前版本为占位实现，可根据实际需求扩展

## 5. 配置说明

### 5.1 相关配置项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| EnableResourceMonitoring | bool | false | 是否启用资源监控 |
| ResourceMonitorInterval | int | 10 | 监控检查间隔（秒） |
| MaxMemoryMB | float64 | 0 | 内存限制（MB），0表示不限制 |
| MaxCPUPercent | float64 | 0 | CPU使用率限制（%），0表示不限制 |

### 5.2 配置示例

```json
{
    "enable_resource_monitoring": true,
    "resource_monitor_interval": 10,
    "max_memory_mb": 512.0,
    "max_cpu_percent": 80.0
}
```

### 5.3 资源限制建议

| 场景 | MaxMemoryMB | MaxCPUPercent | ResourceMonitorInterval |
|------|-------------|---------------|------------------------|
| 开发环境 | 0 (不限制) | 0 (不限制) | 30 |
| 小规模部署 | 256 | 70 | 15 |
| 大规模部署 | 512 | 80 | 10 |
| 高性能要求 | 1024 | 90 | 5 |

## 6. 使用示例

### 6.1 初始化和启动

```go
// 获取配置
cfg := config.GetConfig()

// 创建监控接口实现（通常是monitor包）
monitor := monitor.NewMonitor()

// 创建资源监控器
resourceMonitor := resource.NewResourceMonitor(cfg, monitor)

// 启动资源监控
if err := resourceMonitor.Start(); err != nil {
    log.Printf("Failed to start resource monitor: %v", err)
}
```

### 6.2 获取当前资源使用

```go
// 获取当前资源使用情况
usage := resourceMonitor.GetCurrentUsage()

fmt.Printf("Current Resource Usage:\n")
fmt.Printf("  Memory: %.2f MB (Limit: %.2f MB)\n", usage.MemoryMB, usage.MemoryLimitMB)
fmt.Printf("  CPU: %.2f%% (Limit: %.2f%%)\n", usage.CPUPercent, usage.CPULimitPercent)
fmt.Printf("  Goroutines: %d\n", usage.Goroutines)
fmt.Printf("  Timestamp: %d\n", usage.Timestamp)
```

### 6.3 停止资源监控

```go
// 优雅停止资源监控
resourceMonitor.Stop()
```

## 7. 性能优化

### 7.1 监控间隔优化

根据实际需求调整监控间隔：

```go
// 高频率监控（适用于高负载场景）
cfg.ResourceMonitorInterval = 5

// 低频率监控（适用于低负载场景）
cfg.ResourceMonitorInterval = 30
```

### 7.2 内存采集优化

- 使用gopsutil库获取进程实际物理内存（RSS）
- 当gopsutil获取失败时，回退到runtime.MemStats.Alloc
- 使用Working Set而非虚拟内存，与任务管理器显示一致

### 7.3 CPU采集优化

- 第一次调用时进行阻塞采样以获取基准值
- 短暂等待（100ms）确保采样准确性
- 限制CPU使用率在0-100%范围内，避免异常值

## 8. 错误处理

### 8.1 错误场景及处理

| 错误场景 | 处理方式 |
|---------|---------|
| 获取进程内存失败 | 回退到使用runtime.MemStats.Alloc |
| 获取CPU使用率失败 | 返回0，继续监控 |
| 获取进程信息失败 | 记录错误日志，继续运行 |

### 8.2 日志记录

系统会在以下关键节点记录日志：

- 资源监控启动/停止
- 资源使用情况（定期）
- 内存警告和清理操作
- CPU警告和调整操作
- 错误信息

## 9. 线程安全

### 9.1 互斥锁保护

ResourceMonitor使用互斥锁保护共享数据：

```go
func (rm *ResourceMonitor) GetCurrentUsage() ResourceUsage {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()
    return rm.lastUsage
}
```

### 9.2 并发安全

- 监控循环运行在独立的goroutine中
- 通过通道（stopChan）进行goroutine间通信
- 使用互斥锁保护lastUsage的读写

## 10. 扩展性

### 10.1 自定义资源调整逻辑

可以在adjustResourceUsage中实现自定义的资源调整逻辑：

```go
func (rm *ResourceMonitor) adjustResourceUsage() {
    // 减少批量处理大小
    batchSize := config.GetConfig().BatchSize
    if batchSize > 100 {
        config.GetConfig().BatchSize = batchSize / 2
        log.Printf("Reduced batch size to %d", config.GetConfig().BatchSize)
    }

    // 降低并发数
    workerCount := config.GetConfig().WorkerCount
    if workerCount > 1 {
        config.GetConfig().WorkerCount = workerCount / 2
        log.Printf("Reduced worker count to %d", config.GetConfig().WorkerCount)
    }
}
```

### 10.2 添加新的资源监控项

可以扩展ResourceUsage结构体添加新的监控项：

```go
type ResourceUsage struct {
    // 现有字段...
    DiskUsageMB    float64 `json:"disk_usage_mb"`
    NetworkBytesIn uint64  `json:"network_bytes_in"`
    NetworkBytesOut uint64 `json:"network_bytes_out"`
}
```

## 11. 注意事项

1. **资源限制设置**：建议根据实际硬件配置合理设置资源限制
2. **监控间隔选择**：监控间隔过短会增加系统开销，过长可能错过资源峰值
3. **内存释放效果**：tryFreeMemory释放内存的效果取决于Go运行时状态
4. **CPU采样准确性**：第一次CPU采样需要短暂阻塞，后续采样为非阻塞
5. **自我调整限制**：adjustResourceUsage仅为占位实现，实际资源调整需要业务代码配合

## 12. 总结

资源监控器是sfsEdgeStore系统运维管理的核心组件，通过实时监控和自动调整机制，确保系统资源使用在可控范围内。其设计遵循以下原则：

- **实时性**：持续监控资源使用，及时发现异常
- **可靠性**：多重保护机制，防止资源耗尽
- **可扩展性**：模块化设计，便于添加新的监控项和调整策略
- **易用性**：简单的API设计，便于集成和使用
- **线程安全**：完善的并发控制，保证数据一致性