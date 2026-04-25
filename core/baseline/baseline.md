# 基线管理模块 (Baseline Manager) 技术文档

## 1. 概述

基线管理模块是sfsEdgeStore系统中负责计算和管理设备读数基线的核心组件。它通过分析历史数据，计算设备读数的统计特征（如平均值、标准差、最小值、最大值），并基于这些统计特征生成动态阈值，用于检测异常值。

### 1.1 主要功能

- **基线计算**：基于历史数据计算设备读数的统计特征
- **动态阈值**：根据基线和标准差倍数生成动态阈值
- **异常检测**：检查当前值是否超出动态阈值范围
- **基线缓存**：缓存已计算的基线，避免重复计算
- **基线管理**：支持获取、更新、列出基线

### 1.2 应用场景

- **设备异常检测**：检测设备读数的异常值
- **预测性维护**：基于基线变化预测设备故障
- **数据质量监控**：监控数据的稳定性和一致性
- **告警阈值动态调整**：根据历史数据自动调整告警阈值

## 2. 数据结构

### 2.1 Baseline 结构体

```go
type Baseline struct {
    DeviceName     string    // 设备名称
    ReadingName    string    // 读数名称
    Average        float64   // 平均值
    StdDev         float64   // 标准差
    MinValue       float64   // 最小值
    MaxValue       float64   // 最大值
    SampleCount    int       // 样本数量
    LastUpdated    time.Time // 最后更新时间
    LearningPeriod int       // 学习期（天）
    Enabled        bool      // 是否启用
}
```

**字段说明**：

| 字段             | 类型        | 说明                  |
| -------------- | --------- | ------------------- |
| DeviceName     | string    | 设备名称                |
| ReadingName    | string    | 读数名称（如"GetValue"）   |
| Average        | float64   | 读数的平均值              |
| StdDev         | float64   | 读数的标准差              |
| MinValue       | float64   | 读数的最小值              |
| MaxValue       | float64   | 读数的最大值              |
| SampleCount    | int       | 用于计算的样本数量           |
| LastUpdated    | time.Time | 基线最后更新时间            |
| LearningPeriod | int       | 学习期（天），用于控制数据收集时间范围 |
| Enabled        | bool      | 基线是否启用              |

### 2.2 Manager 结构体

```go
type Manager struct {
    baselines      map[string]Baseline  // 基线缓存，键为 "device:reading"
    learningPeriod int                  // 学习期（天）
    enabled        bool                 // 是否启用基线管理
}
```

**字段说明**：

| 字段             | 类型                   | 说明                                 |
| -------------- | -------------------- | ---------------------------------- |
| baselines      | map\[string]Baseline | 基线缓存，键格式为 "deviceName:readingName" |
| learningPeriod | int                  | 学习期（天），用于控制数据收集时间范围                |
| enabled        | bool                 | 是否启用基线管理                           |

## 3. 函数说明

### 3.1 构造函数

#### NewManager

```go
func NewManager(learningPeriod int, enabled bool) *Manager
```

**功能**：创建基线管理器实例

**参数**：

- `learningPeriod`: 学习期（天）
- `enabled`: 是否启用基线管理

**返回值**：初始化好的Manager指针

***

### 3.2 核心功能

#### CalculateBaseline

```go
func (m *Manager) CalculateBaseline(deviceName, readingName string) (Baseline, error)
```

**功能**：计算设备读数的基线

**参数**：

- `deviceName`: 设备名称
- `readingName`: 读数名称

**处理逻辑**：

1. 检查是否启用基线管理
2. 生成基线键（"deviceName:readingName"）
3. 检查缓存中是否已有基线且未过期（24小时内）
4. 从数据库查询历史数据
5. 调用calculateStatistics计算统计数据
6. 缓存并返回基线

**返回值**：

- `Baseline`: 计算的基线
- `error`: 错误对象（成功时为nil）

#### calculateStatistics

```go
func (m *Manager) calculateStatistics(records record.Records, deviceName, readingName string) (Baseline, error)
```

**功能**：计算统计数据

**参数**：

- `records`: 从数据库查询的记录
- `deviceName`: 设备名称
- `readingName`: 读数名称

**处理逻辑**：

1. 检查记录是否为空
2. 计算总和、最小值、最大值
3. 计算平均值
4. 计算标准差
5. 构建并返回Baseline对象

**返回值**：

- `Baseline`: 计算的基线
- `error`: 错误对象（成功时为nil）

***

### 3.3 阈值和异常检测

#### GetDynamicThreshold

```go
func (m *Manager) GetDynamicThreshold(deviceName, readingName string, stdMultiplier float64) (float64, float64, error)
```

**功能**：获取动态阈值

**参数**：

- `deviceName`: 设备名称
- `readingName`: 读数名称
- `stdMultiplier`: 标准差倍数

**处理逻辑**：

1. 调用CalculateBaseline获取基线
2. 计算上下阈值：
   - 上阈值 = 平均值 + (标准差 × 倍数)
   - 下阈值 = 平均值 - (标准差 × 倍数)

**返回值**：

- `float64`: 下阈值
- `float64`: 上阈值
- `error`: 错误对象（成功时为nil）

#### CheckAnomaly

```go
func (m *Manager) CheckAnomaly(deviceName, readingName string, value float64, stdMultiplier float64) (bool, error)
```

**功能**：检查是否异常

**参数**：

- `deviceName`: 设备名称
- `readingName`: 读数名称
- `value`: 当前值
- `stdMultiplier`: 标准差倍数

**处理逻辑**：

1. 调用GetDynamicThreshold获取阈值
2. 检查当前值是否超出阈值范围

**返回值**：

- `bool`: 是否异常（超出阈值为true）
- `error`: 错误对象（成功时为nil）

***

### 3.4 基线管理

#### GetBaseline

```go
func (m *Manager) GetBaseline(deviceName, readingName string) (Baseline, bool)
```

**功能**：获取指定设备和读数的基线

**参数**：

- `deviceName`: 设备名称
- `readingName`: 读数名称

**返回值**：

- `Baseline`: 基线对象
- `bool`: 是否存在

#### UpdateBaseline

```go
func (m *Manager) UpdateBaseline(baseline Baseline)
```

**功能**：更新基线

**参数**：

- `baseline`: 基线对象

#### ListBaselines

```go
func (m *Manager) ListBaselines() map[string]Baseline
```

**功能**：列出所有基线

**返回值**：

- `map[string]Baseline`: 所有基线的映射

## 4. 算法说明

### 4.1 统计计算

#### 平均值计算

```go
average := sum / float64(len(records))
```

#### 标准差计算

```go
// 计算方差
var variance float64
for _, record := range records {
    if value, ok := record["value"].(float64); ok {
        variance += math.Pow(value-average, 2)
    }
}

// 计算标准差
stdDev := math.Sqrt(variance / float64(len(records)))
```

### 4.2 动态阈值计算

```go
upperThreshold := baseline.Average + (stdMultiplier * baseline.StdDev)
lowerThreshold := baseline.Average - (stdMultiplier * baseline.StdDev)
```

### 4.3 异常检测

```go
return value < lowerThreshold || value > upperThreshold
```

## 5. 使用示例

### 5.1 基本使用

```go
// 创建基线管理器（30天学习期，启用）
baselineManager := baseline.NewManager(30, true)

// 计算设备读数的基线
baseline, err := baselineManager.CalculateBaseline("Temperature-Sensor-01", "GetValue")
if err != nil {
    log.Printf("Failed to calculate baseline: %v", err)
    return
}

// 打印基线信息
fmt.Printf("Baseline for Temperature-Sensor-01:GetValue:\n")
fmt.Printf("  Average: %.2f\n", baseline.Average)
fmt.Printf("  StdDev: %.2f\n", baseline.StdDev)
fmt.Printf("  Min: %.2f\n", baseline.MinValue)
fmt.Printf("  Max: %.2f\n", baseline.MaxValue)
fmt.Printf("  Sample Count: %d\n", baseline.SampleCount)
fmt.Printf("  Last Updated: %s\n", baseline.LastUpdated)
```

### 5.2 动态阈值和异常检测

```go
// 获取动态阈值（使用2倍标准差）
lower, upper, err := baselineManager.GetDynamicThreshold("Temperature-Sensor-01", "GetValue", 2.0)
if err != nil {
    log.Printf("Failed to get dynamic threshold: %v", err)
    return
}

fmt.Printf("Dynamic thresholds for Temperature-Sensor-01:GetValue:\n")
fmt.Printf("  Lower: %.2f\n", lower)
fmt.Printf("  Upper: %.2f\n", upper)

// 检查当前值是否异常
currentValue := 35.0 // 假设当前温度值
isAnomaly, err := baselineManager.CheckAnomaly("Temperature-Sensor-01", "GetValue", currentValue, 2.0)
if err != nil {
    log.Printf("Failed to check anomaly: %v", err)
    return
}

if isAnomaly {
    fmt.Printf("Current value %.2f is ANOMALOUS!\n", currentValue)
} else {
    fmt.Printf("Current value %.2f is within normal range.\n", currentValue)
}
```

### 5.3 基线管理

```go
// 获取基线
baseline, exists := baselineManager.GetBaseline("Temperature-Sensor-01", "GetValue")
if exists {
    fmt.Printf("Baseline exists: %.2f (avg)\n", baseline.Average)
} else {
    fmt.Println("Baseline does not exist")
}

// 列出所有基线
baselines := baselineManager.ListBaselines()
fmt.Printf("Total baselines: %d\n", len(baselines))
for key, bl := range baselines {
    fmt.Printf("  %s: %.2f\n", key, bl.Average)
}
```

## 6. 性能优化

### 6.1 基线缓存

- 使用map缓存已计算的基线，避免重复计算
- 24小时内的基线视为有效，不需要重新计算

### 6.2 数据查询优化

- 直接使用数据库查询功能，避免额外的数据处理
- 利用数据库的索引加速查询

### 6.3 计算优化

- 单次遍历计算总和、最小值、最大值，减少遍历次数
- 标准差计算使用单次遍历，避免多次遍历数据

## 7. 错误处理

### 7.1 常见错误

| 错误场景    | 原因       | 处理方式         |
| ------- | -------- | ------------ |
| 数据库查询失败 | 数据库连接问题  | 返回错误         |
| 无数据     | 设备无历史数据  | 返回空基线（平均值为0） |
| 基线计算失败  | 数据类型转换失败 | 返回错误         |

### 7.2 错误返回

```go
// 数据库查询失败
records, err := database.QueryRecords(...)
if err != nil {
    return Baseline{}, err
}

// 无数据处理
if len(records) == 0 {
    return Baseline{
        DeviceName:     deviceName,
        ReadingName:    readingName,
        Average:        0,
        StdDev:         0,
        MinValue:       0,
        MaxValue:       0,
        SampleCount:    0,
        LastUpdated:    time.Now(),
        LearningPeriod: m.learningPeriod,
        Enabled:        m.enabled,
    }, nil
}
```

## 8. 与其他模块的集成

### 8.1 告警模块

```go
// 在告警模块中使用基线检测
func checkAlert(deviceName, readingName string, value float64) {
    // 检查是否异常
    isAnomaly, err := baselineManager.CheckAnomaly(deviceName, readingName, value, 2.0)
    if err != nil {
        log.Printf("Failed to check anomaly: %v", err)
        return
    }

    if isAnomaly {
        // 触发告警
        alertManager.TriggerAlert(deviceName, readingName, value, "Anomaly detected")
    }
}
```

### 8.2 数据处理模块

```go
// 在数据处理模块中更新基线
func processReading(deviceName, readingName string, value float64) {
    // 存储数据
    database.StoreReading(deviceName, readingName, value)

    // 检查是否需要更新基线
    baseline, exists := baselineManager.GetBaseline(deviceName, readingName)
    if !exists || time.Since(baseline.LastUpdated) > 24*time.Hour {
        // 重新计算基线
        _, err := baselineManager.CalculateBaseline(deviceName, readingName)
        if err != nil {
            log.Printf("Failed to update baseline: %v", err)
        }
    }
}
```

## 9. 配置建议

### 9.1 学习期设置

| 场景   | 学习期（天） | 说明              |
| ---- | ------ | --------------- |
| 稳定环境 | 30     | 环境稳定，数据波动小      |
| 动态环境 | 7-14   | 环境变化快，需要更频繁更新基线 |
| 新设备  | 1-3    | 新设备，需要快速建立初始基线  |

### 9.2 标准差倍数设置

| 场景   | 倍数  | 说明      |
| ---- | --- | ------- |
| 严格检测 | 1.5 | 检测轻微异常  |
| 标准检测 | 2.0 | 检测明显异常  |
| 宽松检测 | 3.0 | 只检测严重异常 |

## 10. 注意事项

1. **数据质量**：基线计算依赖于历史数据的质量，确保数据采集的准确性
2. **计算开销**：首次计算基线可能需要较长时间，尤其是历史数据较多时
3. **基线更新**：基线默认24小时更新一次，可根据实际需求调整
4. **异常处理**：无历史数据时会返回空基线，需要处理这种情况
5. **性能考虑**：对于大量设备，建议分批计算基线，避免系统负载过高

## 11. 总结

基线管理模块为sfsEdgeStore系统提供了智能的异常检测能力，通过统计分析历史数据，自动生成动态阈值，实现了设备数据的异常检测。其设计遵循以下原则：

- **智能性**：基于统计分析自动生成基线和阈值
- **灵活性**：支持不同的学习期和标准差倍数设置
- **效率**：缓存机制避免重复计算
- **可靠性**：完善的错误处理机制
- **可扩展性**：模块化设计便于与其他模块集成

通过基线管理模块，系统可以更准确地检测设备异常，提高系统的可靠性和预测性维护能力。
