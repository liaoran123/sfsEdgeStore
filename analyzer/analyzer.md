# 轻量级分析引擎 (Analyzer) 技术文档

## 1. 概述

轻量级分析引擎是sfsEdgeStore系统中的一个核心组件，负责对设备数据进行实时分析和异常检测。它提供了滑动窗口分析、阈值异常检测和趋势异常检测等功能，帮助用户发现数据中的异常模式和趋势。

### 1.1 主要功能

- **滑动窗口分析**：对数据进行时间窗口的聚合分析
- **阈值异常检测**：基于配置的阈值检测异常值
- **趋势异常检测**：检测数据的异常趋势变化
- **内存管理**：使用内存池减少内存分配
- **执行时间限制**：限制分析执行时间，避免影响系统性能

### 1.2 设计目标

- **轻量高效**：最小化资源占用，适合在边缘设备上运行
- **实时分析**：对数据进行实时分析，及时发现异常
- **可配置**：支持通过配置调整分析参数
- **可扩展**：易于添加新的分析算法

## 2. 数据结构

### 2.1 Analyzer 结构体

```go
type Analyzer struct {
    config        *config.Config
    memoryPool    *sync.Pool
    isEnabled     bool
    maxMemory     int
    maxTimePerRun time.Duration
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| config | *config.Config | 配置对象 |
| memoryPool | *sync.Pool | 内存池，用于减少内存分配 |
| isEnabled | bool | 分析引擎是否启用 |
| maxMemory | int | 最大内存使用限制 |
| maxTimePerRun | time.Duration | 每次分析的最大执行时间 |

### 2.2 AnalysisResult 结构体

```go
type AnalysisResult struct {
    DeviceName   string    `json:"device_name"`
    Reading      string    `json:"reading"`
    AnalysisType string    `json:"analysis_type"`
    Value        float64   `json:"value"`
    WindowStart  time.Time `json:"window_start"`
    WindowEnd    time.Time `json:"window_end"`
    Timestamp    time.Time `json:"timestamp"`
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| DeviceName | string | 设备名称 |
| Reading | string | 读数名称 |
| AnalysisType | string | 分析类型 |
| Value | float64 | 分析结果值 |
| WindowStart | time.Time | 分析窗口开始时间 |
| WindowEnd | time.Time | 分析窗口结束时间 |
| Timestamp | time.Time | 分析时间戳 |

### 2.3 Alert 结构体

```go
type Alert struct {
    DeviceName string    `json:"device_name"`
    Reading    string    `json:"reading"`
    AlertType  string    `json:"alert_type"`
    Message    string    `json:"message"`
    Value      float64   `json:"value"`
    Threshold  float64   `json:"threshold"`
    Timestamp  time.Time `json:"timestamp"`
    Severity   string    `json:"severity"`
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| DeviceName | string | 设备名称 |
| Reading | string | 读数名称 |
| AlertType | string | 告警类型 |
| Message | string | 告警消息 |
| Value | float64 | 触发告警的值 |
| Threshold | float64 | 阈值 |
| Timestamp | time.Time | 告警时间戳 |
| Severity | string | 告警严重程度 |

## 3. 函数说明

### 3.1 构造函数

#### NewAnalyzer
```go
func NewAnalyzer(cfg *config.Config) *Analyzer
```

**功能**：创建新的分析引擎

**参数**：
- `cfg`: 配置对象

**返回值**：初始化好的Analyzer指针

### 3.2 核心功能

#### Analyze
```go
func (a *Analyzer) Analyze(data []map[string]interface{}, deviceName, reading string) ([]AnalysisResult, []Alert)
```

**功能**：分析数据

**参数**：
- `data`: 要分析的数据
- `deviceName`: 设备名称
- `reading`: 读数名称

**处理逻辑**：
1. 检查分析引擎是否启用
2. 限制分析时间
3. 执行滑动窗口分析
4. 执行异常检测

**返回值**：
- `[]AnalysisResult`: 分析结果
- `[]Alert`: 检测到的告警

#### analyzeSlidingWindow
```go
func (a *Analyzer) analyzeSlidingWindow(data []map[string]interface{}, deviceName, reading string) []AnalysisResult
```

**功能**：滑动窗口分析

**参数**：
- `data`: 要分析的数据
- `deviceName`: 设备名称
- `reading`: 读数名称

**处理逻辑**：
1. 创建5分钟滑动窗口，步长1分钟
2. 转换数据格式
3. 计算滑动窗口的平均值
4. 构建分析结果

**返回值**：
- `[]AnalysisResult`: 滑动窗口分析结果

#### detectAnomalies
```go
func (a *Analyzer) detectAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert
```

**功能**：异常检测

**参数**：
- `data`: 要分析的数据
- `deviceName`: 设备名称
- `reading`: 读数名称

**处理逻辑**：
1. 执行阈值异常检测
2. 执行趋势异常检测

**返回值**：
- `[]Alert`: 检测到的告警

#### detectThresholdAnomalies
```go
func (a *Analyzer) detectThresholdAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert
```

**功能**：阈值异常检测

**参数**：
- `data`: 要分析的数据
- `deviceName`: 设备名称
- `reading`: 读数名称

**处理逻辑**：
1. 获取阈值配置
2. 检查每个数据点是否超出阈值
3. 生成告警

**返回值**：
- `[]Alert`: 阈值异常告警

#### detectTrendAnomalies
```go
func (a *Analyzer) detectTrendAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert
```

**功能**：趋势异常检测

**参数**：
- `data`: 要分析的数据
- `deviceName`: 设备名称
- `reading`: 读数名称

**处理逻辑**：
1. 提取值和时间戳
2. 检查最近三个值的变化
3. 检测连续快速增长或下降
4. 生成告警

**返回值**：
- `[]Alert`: 趋势异常告警

### 3.3 控制功能

#### Enable
```go
func (a *Analyzer) Enable()
```

**功能**：启用分析引擎

#### Disable
```go
func (a *Analyzer) Disable()
```

**功能**：禁用分析引擎

#### IsEnabled
```go
func (a *Analyzer) IsEnabled() bool
```

**功能**：检查分析引擎是否启用

**返回值**：
- `bool`: 分析引擎是否启用

## 4. 分析算法

### 4.1 滑动窗口分析

**算法描述**：
- 创建5分钟的滑动窗口，步长为1分钟
- 计算每个窗口内数据的平均值
- 生成分析结果

**应用场景**：
- 平滑数据波动
- 识别数据趋势
- 减少数据噪声

### 4.2 阈值异常检测

**算法描述**：
- 从配置中获取阈值（最小值和最大值）
- 检查每个数据点是否超出阈值范围
- 超出阈值时生成告警

**应用场景**：
- 检测传感器读数的异常值
- 监控设备运行状态
- 及时发现设备故障

### 4.3 趋势异常检测

**算法描述**：
- 计算连续数据点之间的变化率
- 检测连续快速增长或下降的趋势
- 发现异常趋势时生成告警

**应用场景**：
- 检测设备性能的突然变化
- 预测设备故障
- 识别异常的业务模式

## 5. 使用示例

### 5.1 基本使用

```go
// 创建分析引擎
cfg, _ := config.Load()
analyzer := analyzer.NewAnalyzer(cfg)

// 启用分析引擎
analyzer.Enable()

// 准备数据
data := []map[string]interface{}{
    {"value": 25.5, "timestamp": time.Now().Add(-2 * time.Minute)},
    {"value": 26.0, "timestamp": time.Now().Add(-1 * time.Minute)},
    {"value": 26.5, "timestamp": time.Now()},
}

// 分析数据
results, alerts := analyzer.Analyze(data, "Temperature-Sensor-01", "GetValue")

// 处理分析结果
fmt.Printf("Analysis results: %d\n", len(results))
for _, result := range results {
    fmt.Printf("%s: %.2f (%.2f - %.2f)\n", 
        result.AnalysisType, 
        result.Value, 
        result.WindowStart.Unix(), 
        result.WindowEnd.Unix())
}

// 处理告警
fmt.Printf("Alerts: %d\n", len(alerts))
for _, alert := range alerts {
    fmt.Printf("Alert: %s - %s (%.2f)\n", 
        alert.AlertType, 
        alert.Message, 
        alert.Value)
}
```

### 5.2 阈值配置

```go
// 在配置中设置阈值
cfg := config.Load()
cfg.EnableAnalyzer = true
cfg.AnalyzerThresholds = map[string]config.ThresholdConfig{
    "GetValue": {Min: 0, Max: 30},  // 温度阈值：0-30
    "Humidity": {Min: 20, Max: 80},  // 湿度阈值：20-80
}

// 创建分析引擎
analyzer := analyzer.NewAnalyzer(cfg)
```

### 5.3 异常检测

```go
// 检测异常数据
anomalyData := []map[string]interface{}{
    {"value": 25.5, "timestamp": time.Now().Add(-2 * time.Minute)},
    {"value": 26.0, "timestamp": time.Now().Add(-1 * time.Minute)},
    {"value": 35.0, "timestamp": time.Now()},  // 超出阈值的异常值
}

results, alerts := analyzer.Analyze(anomalyData, "Temperature-Sensor-01", "GetValue")

// 处理告警
if len(alerts) > 0 {
    fmt.Println("Anomalies detected!")
    for _, alert := range alerts {
        fmt.Printf("%s: %s - Severity: %s\n", 
            alert.DeviceName, 
            alert.Message, 
            alert.Severity)
    }
}
```

## 6. 性能优化

### 6.1 内存管理

- **内存池**：使用sync.Pool减少内存分配
- **数据复用**：复用分析结果对象
- **内存限制**：设置最大内存使用限制

### 6.2 执行时间控制

- **时间限制**：限制每次分析的执行时间
- **超时检测**：检测分析是否超时
- **资源保护**：避免分析占用过多系统资源

### 6.3 算法优化

- **滑动窗口**：使用高效的滑动窗口算法
- **批量处理**：批量处理数据，减少函数调用开销
- **早期返回**：对于数据不足的情况，提前返回

## 7. 错误处理

### 7.1 常见错误

| 错误场景 | 原因 | 处理方式 |
|---------|------|---------|
| 数据不足 | 数据点少于3个 | 跳过趋势分析，只执行基本分析 |
| 类型转换失败 | 数据类型不是float64 | 跳过该数据点 |
| 时间戳缺失 | 数据中没有timestamp字段 | 使用当前时间 |
| 执行超时 | 分析时间超过限制 | 记录警告日志，继续执行 |

### 7.2 错误处理策略

- **优雅降级**：当出现错误时，尝试使用替代方案
- **日志记录**：记录错误信息，便于问题排查
- **继续执行**：即使部分分析失败，也继续执行其他分析

## 8. 配置选项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| EnableAnalyzer | bool | false | 是否启用分析引擎 |
| AnalyzerMaxMemory | int | 10485760 (10MB) | 分析引擎最大内存使用 |
| AnalyzerMaxTimePerRun | int | 500 (ms) | 每次分析的最大执行时间 |
| AnalyzerThresholds | map[string]ThresholdConfig | {} | 各读数的阈值配置 |

## 9. 注意事项

1. **数据质量**：分析结果的质量依赖于输入数据的质量
2. **性能影响**：分析操作会占用系统资源，建议在适当的时机执行
3. **阈值配置**：合理设置阈值，避免过多的误告警
4. **内存使用**：对于大量数据的分析，注意内存使用情况
5. **执行时间**：对于实时系统，注意分析的执行时间

## 10. 总结

轻量级分析引擎为sfsEdgeStore系统提供了强大的数据分析和异常检测能力，通过滑动窗口分析、阈值检测和趋势检测等算法，帮助用户及时发现设备数据中的异常模式。其设计遵循以下原则：

- **轻量高效**：最小化资源占用，适合边缘设备
- **实时分析**：对数据进行实时分析，及时发现异常
- **可配置**：通过配置调整分析参数
- **可扩展**：易于添加新的分析算法
- **可靠稳定**：完善的错误处理和资源管理

通过分析引擎，系统可以更智能地处理设备数据，提高设备监控的效率和准确性，为用户提供更好的设备管理体验。