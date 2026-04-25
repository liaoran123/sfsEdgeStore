# 数据保留管理器 (Retention Manager) 技术文档

## 1. 概述

数据保留管理器是sfsEdgeStore系统中的一个核心组件，负责管理数据库中历史数据的自动清理工作。它通过配置保留策略，自动删除超过指定时间范围的数据，以控制数据库存储空间的使用。

### 1.1 主要功能

- **自动数据清理**：根据配置的保留天数，自动清理过期的历史数据
- **可配置的保留策略**：支持自定义保留天数、清理间隔和批量大小
- **后台运行**：作为goroutine在后台持续运行，不影响主业务
- **分批处理**：采用分批删除策略，避免一次性删除大量数据对系统造成压力
- **实时配置更新**：从配置管理器获取最新配置，无需重启即可应用新配置

## 2. 架构设计

### 2.1 核心组件

```
RetentionManager
├── table: *engine.Table         # 数据库表引用
├── config: *config.Config       # 配置对象
├── stopChan: chan struct{}      # 停止信号通道
└── isRunning: bool              # 运行状态标志
```

### 2.2 工作流程

```
Start()
    ↓
cleanupLoop() [后台goroutine]
    ↓
定期触发CleanupOldData()
    ↓
cleanupBatch() [循环批量删除]
    ↓
直至所有过期数据被清理
```

## 3. 数据结构

### 3.1 RetentionManager 结构体

```go
type RetentionManager struct {
    table     *engine.Table    // 数据库表引用，用于执行数据删除操作
    config    *config.Config  // 配置对象，包含保留策略参数
    stopChan  chan struct{}   // 停止信号通道，用于优雅关闭
    isRunning bool            // 运行状态标志，防止重复启动
}
```

### 3.2 配置参数说明

| 参数 | 类型 | 说明 |
|------|------|------|
| EnableRetentionPolicy | bool | 是否启用数据保留策略 |
| RetentionDays | int | 数据保留天数（从当前时间往前计算） |
| CleanupInterval | int | 清理任务执行间隔（单位：小时） |
| CleanupBatchSize | int | 每次批量删除的记录数 |

## 4. 函数说明

### 4.1 构造函数

#### NewRetentionManager
```go
func NewRetentionManager(table *engine.Table, cfg *config.Config) *RetentionManager
```

**功能**：创建并初始化一个新的RetentionManager实例

**参数**：
- `table`: 数据库表引用
- `cfg`: 配置对象

**返回值**：初始化好的RetentionManager指针

---

### 4.2 生命周期管理

#### Start
```go
func (rm *RetentionManager) Start() error
```

**功能**：启动数据保留管理器

**处理逻辑**：
1. 检查是否启用保留策略，未启用则直接返回
2. 检查是否已在运行，避免重复启动
3. 启动后台清理循环goroutine

**返回值**：错误对象（通常为nil）

---

#### Stop
```go
func (rm *RetentionManager) Stop()
```

**功能**：停止数据保留管理器

**处理逻辑**：
1. 检查是否在运行
2. 发送停止信号到stopChan
3. 更新运行状态标志

---

### 4.3 核心业务逻辑

#### cleanupLoop
```go
func (rm *RetentionManager) cleanupLoop()
```

**功能**：后台清理循环，持续定期执行数据清理任务

**处理逻辑**：
1. 根据CleanupInterval创建定时器
2. 循环等待定时器触发或停止信号
3. 定时器触发时执行CleanupOldData
4. 记录清理结果和删除的记录数

**注意**：此函数运行在独立的goroutine中

---

#### CleanupOldData
```go
func (rm *RetentionManager) CleanupOldData() (int, error)
```

**功能**：执行数据清理任务

**处理逻辑**：
1. 计算截止时间戳（当前时间减去保留天数）
2. 循环执行批量删除，直至所有过期数据被清理
3. 累计删除的记录总数

**返回值**：
- 删除的记录总数
- 错误对象

---

#### cleanupBatch
```go
func (rm *RetentionManager) cleanupBatch(cutoffTimestamp int64, batchSize int) (int, error)
```

**功能**：批量删除指定时间戳之前的记录

**参数**：
- `cutoffTimestamp`: 截止时间戳（纳秒）
- `batchSize`: 批量大小

**处理逻辑**：
1. 构建查询范围（timestamp从nil到cutoffTimestamp）
2. 执行范围查询获取待删除记录
3. 执行批量删除操作
4. 返回实际删除的记录数

**返回值**：
- 删除的记录数
- 错误对象

---

### 4.4 配置查询

#### GetRetentionStatus
```go
func (rm *RetentionManager) GetRetentionStatus() map[string]any
```

**功能**：获取当前数据保留状态

**处理逻辑**：
1. 从配置管理器获取最新配置
2. 如果获取失败，使用本地配置作为后备
3. 构建包含所有保留策略参数的状态map

**返回值**：包含以下字段的map：
- `enabled`: 是否启用
- `retention_days`: 保留天数
- `cleanup_interval`: 清理间隔（小时）
- `cleanup_batch_size`: 批量大小
- `is_running`: 是否正在运行

## 5. 配置示例

### 5.1 配置文件结构

```json
{
  "enable_retention_policy": true,
  "retention_days": 30,
  "cleanup_interval": 24,
  "cleanup_batch_size": 1000
}
```

### 5.2 配置说明

| 配置项 | 值 | 说明 |
|--------|-----|------|
| enable_retention_policy | true | 启用数据保留策略 |
| retention_days | 30 | 保留最近30天的数据 |
| cleanup_interval | 24 | 每24小时执行一次清理 |
| cleanup_batch_size | 1000 | 每次批量删除1000条记录 |

## 6. 使用示例

### 6.1 初始化和启动

```go
// 创建数据库表引用
table := engine.OpenTable("data")

// 获取配置
cfg := config.GetConfig()

// 创建保留管理器
rm := retention.NewRetentionManager(table, cfg)

// 启动保留管理器
if err := rm.Start(); err != nil {
    log.Printf("Failed to start retention manager: %v", err)
}
```

### 6.2 查询保留状态

```go
// 获取当前保留状态
status := rm.GetRetentionStatus()

fmt.Printf("Retention Status:\n")
fmt.Printf("  Enabled: %v\n", status["enabled"])
fmt.Printf("  Retention Days: %d\n", status["retention_days"])
fmt.Printf("  Cleanup Interval: %d hours\n", status["cleanup_interval"])
fmt.Printf("  Is Running: %v\n", status["is_running"])
```

### 6.3 停止保留管理器

```go
// 优雅停止保留管理器
rm.Stop()
```

## 7. 性能优化

### 7.1 分批处理策略

系统采用分批删除策略，避免一次性删除大量数据对系统造成压力：

1. **内存优化**：每次只加载batchSize条记录到内存
2. **锁竞争优化**：减少长时间锁定数据库
3. **日志优化**：每批删除后记录进度，便于监控

### 7.2 配置建议

| 场景 | retention_days | cleanup_interval | cleanup_batch_size |
|------|---------------|-----------------|-------------------|
| 开发环境 | 7 | 1 | 100 |
| 小规模部署 | 30 | 24 | 500 |
| 大规模部署 | 90 | 12 | 1000 |

## 8. 注意事项

1. **数据安全**：删除操作不可逆，请确保配置正确的保留天数
2. **资源占用**：清理任务会占用一定的数据库I/O资源
3. **启动检查**：重复调用Start()不会重启系统，只会在日志中记录警告
4. **优雅关闭**：Stop()会等待当前批次删除完成后再退出
5. **配置更新**：GetRetentionStatus()会实时获取最新配置，但实际清理参数在启动时确定

## 9. 错误处理

### 9.1 常见错误及处理方式

| 错误场景 | 处理方式 |
|---------|---------|
| 数据库连接失败 | 返回错误，清理任务中断 |
| 删除权限不足 | 返回错误，记录日志 |
| 配置获取失败 | 使用本地配置作为后备 |

### 9.2 日志记录

系统会在以下关键节点记录日志：

- 保留管理器启动/停止
- 清理任务开始/完成
- 每批删除的记录数
- 错误信息

## 10. 总结

数据保留管理器是sfsEdgeStore系统存储管理的核心组件，通过自动化的数据清理机制，确保数据库存储空间得到合理控制。其设计遵循以下原则：

- **可靠性**：分批处理确保系统稳定运行
- **可配置性**：灵活的参数满足不同场景需求
- **易用性**：简单的API设计便于集成和使用
- **可观测性**：完善的日志记录便于问题排查