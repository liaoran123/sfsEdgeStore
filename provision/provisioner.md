# 设备配置器 (Provisioner) 技术文档

## 1. 概述

设备配置器是sfsEdgeStore系统中负责自动配置和注册设备到EdgeX Foundry的核心组件。它通过读取设备配置文件，自动批量配置Modbus、MQTT等协议的设备，并将这些设备注册到EdgeX Core Metadata服务，实现设备的即插即用。

### 1.1 主要功能

- **批量设备配置**：支持从配置文件批量加载和配置多个设备
- **多协议支持**：支持Modbus-TCP、Modbus-RTU、MQTT等多种协议
- **设备生命周期管理**：支持设备的注册、删除、查询等操作
- **配置模板化**：支持基于模板的设备配置
- **连接验证**：提供EdgeX连接状态验证功能
- **设备Profile管理**：支持上传设备Profile到EdgeX

### 1.2 适用场景

- **工业自动化**：快速配置和部署大量工业传感器设备
- **物联网网关**：自动发现和配置连接到网关的设备
- **边缘计算**：简化边缘设备的管理和部署流程
- **系统集成**：与EdgeX Foundry生态系统无缝集成

## 2. 架构设计

### 2.1 核心组件

```
Provisioner
├── configManager  *devconfig.ConfigManager  # 配置管理器，用于加载设备配置
├── edgeXEndpoint  string                     # EdgeX Core Metadata API地址
└── httpClient    *http.Client               # HTTP客户端，30秒超时
```

### 2.2 工作流程

```
ProvisionAll()
    ↓
加载设备配置 (configManager.LoadDevices)
    ↓
遍历所有设备
    ↓
ProvisionDevice() → 单个设备配置
    ↓
├── 验证必填参数 (name, ip)
├── 生成Profile名称
├── 构建协议配置
├── 创建设备对象 (DeviceV2)
    ↓
registerDevice() → 注册到EdgeX
    ↓
记录配置结果
```

### 2.3 EdgeX集成架构

```
┌─────────────────┐     ┌──────────────────────┐     ┌─────────────────┐
│   Provisioner   │────▶│  EdgeX Core Metadata │────▶│   Device        │
│   (sfsEdgeStore)│     │  (Port 59881)        │     │   Registry      │
└─────────────────┘     └──────────────────────┘     └─────────────────┘
         │
         │ POST /api/v1/deviceprofile
         ▼
┌─────────────────┐
│  Device Profile │
│  (YAML)         │
└─────────────────┘
```

## 3. 数据结构

### 3.1 Provisioner 结构体

```go
type Provisioner struct {
    configManager *devconfig.ConfigManager  // 配置管理器
    edgeXEndpoint string                    // EdgeX API地址
    httpClient    *http.Client              // HTTP客户端（30秒超时）
}
```

### 3.2 设备相关结构体

#### DeviceProfile
```go
type DeviceProfile struct {
    Name         string   `json:"name"`          // 设备Profile名称
    Manufacturer string   `json:"manufacturer"`  // 制造商
    Model        string   `json:"model"`         // 型号
    Labels       []string `json:"labels"`       // 标签
    Description  string   `json:"description"`  // 描述
}
```

#### Device (V1)
```go
type Device struct {
    Name        string                       `json:"name"`         // 设备名称
    ServiceName string                       `json:"serviceName"`  // 服务名称
    ProfileName string                       `json:"profileName"`  // Profile名称
    Protocols   map[string]map[string]string `json:"protocols"`   // 协议配置
    AutoEvents  []AutoEvent                  `json:"autoEvents"`   // 自动事件
}
```

#### DeviceV2
```go
type DeviceV2 struct {
    Name        string                       `json:"name"`         // 设备名称
    ServiceName string                       `json:"serviceName"`  // 服务名称
    ProfileName string                       `json:"profileName"`  // Profile名称
    Protocols   map[string]map[string]string `json:"protocols"`   // 协议配置
    AutoEvents  []AutoEventV2                `json:"autoEvents"`   // 自动事件(V2)
}
```

### 3.3 自动事件结构体

#### AutoEvent (V1)
```go
type AutoEvent struct {
    Interval string `json:"interval"`  // 事件间隔 (如 "15s")
    OnChange bool   `json:"onChange"`  // 是否在值变化时触发
    Resource string `json:"resource"`  // 资源名称
}
```

#### AutoEventV2 (V2)
```go
type AutoEventV2 struct {
    Interval string `json:"interval"`  // 事件间隔 (如 "15s")
    OnChange bool   `json:"onChange"`  // 是否在值变化时触发
    Resource string `json:"resource"`  // 资源名称
}
```

## 4. 常量定义

### 4.1 EdgeX API端点

```go
const (
    EdgeXCoreMetadataAPI = "http://localhost:59881/api/v1/device"   // 设备管理API
    EdgeXProfileAPI      = "http://localhost:59880/api/v1/deviceprofile"  // 设备Profile API
)
```

### 4.2 默认值

| 常量/默认值 | 值 | 说明 |
|-------------|-----|------|
| 默认协议 | "modbus" | 当未指定协议时使用 |
| 默认间隔 | "15s" | 当未指定采集间隔时使用 |
| Modbus端口 | "502" | Modbus-TCP默认端口 |
| HTTP超时 | 30秒 | HTTP请求超时时间 |

## 5. 函数说明

### 5.1 构造函数

#### NewProvisioner
```go
func NewProvisioner(cm *devconfig.ConfigManager, edgeXEndpoint string) *Provisioner
```

**功能**：创建并初始化设备配置器实例

**参数**：
- `cm`: 配置管理器实例
- `edgeXEndpoint`: EdgeX API地址（可为空，使用默认值）

**返回值**：初始化好的Provisioner指针

---

### 5.2 设备配置

#### ProvisionDevice
```go
func (p *Provisioner) ProvisionDevice(name, ip, protocol, templateName, interval string, onChange bool, minInterval string, subscriptionTopic string) error
```

**功能**：配置并注册单个设备到EdgeX

**参数**：
- `name`: 设备名称（必填）
- `ip`: 设备IP地址（必填）
- `protocol`: 协议类型（可选，默认"modbus"）
- `templateName`: 模板名称（可选）
- `interval`: 采集间隔（可选，默认"15s"）
- `onChange`: 是否在值变化时触发（可选）
- `minInterval`: 最小采集间隔（可选，会覆盖interval参数）
- `subscriptionTopic`: 订阅主题（可选，仅记录日志）

**处理逻辑**：
1. 验证必填参数（name、ip）
2. 设置默认值（protocol="modbus", interval="15s"）
3. 生成Profile名称
4. 构建协议配置（基于protocol类型）
5. 创建DeviceV2对象
6. 调用registerDevice注册到EdgeX

**返回值**：错误对象（成功时为nil）

#### ProvisionAll
```go
func (p *Provisioner) ProvisionAll() error
```

**功能**：批量配置所有设备

**处理逻辑**：
1. 从配置管理器加载设备列表
2. 遍历所有设备
3. 调用ProvisionDevice配置每个设备
4. 统计成功/失败数量
5. 输出详细的配置结果

**返回值**：错误对象（如果有设备配置失败）

---

### 5.3 设备管理

#### RemoveDevice
```go
func (p *Provisioner) RemoveDevice(name string) error
```

**功能**：从EdgeX删除指定设备

**参数**：
- `name`: 设备名称（必填）

**处理逻辑**：
1. 构建删除请求URL（/name/{name}）
2. 发送DELETE请求
3. 处理响应（204 No Content 或 404 Not Found都视为成功）

**返回值**：错误对象

#### ListDevices
```go
func (p *Provisioner) ListDevices() error
```

**功能**：列出EdgeX中所有已注册的设备

**处理逻辑**：
1. 发送GET请求到EdgeX API
2. 解析返回的JSON数组
3. 格式化并打印设备信息

**返回值**：错误对象

---

### 5.4 连接验证

#### ValidateConnection
```go
func (p *Provisioner) ValidateConnection() error
```

**功能**：验证与EdgeX Core Metadata的连接

**处理逻辑**：
1. 访问EdgeX健康检查端点（/api/v1/health）
2. 检查HTTP响应状态码
3. 输出连接状态

**返回值**：错误对象（连接失败时）

---

### 5.5 Profile管理

#### UploadProfile
```go
func (p *Provisioner) UploadProfile(profilePath string) error
```

**功能**：上传设备Profile到EdgeX

**参数**：
- `profilePath`: Profile文件路径（YAML格式）

**处理逻辑**：
1. 读取Profile文件内容
2. 创建POST请求（Content-Type: application/yaml）
3. 发送到EdgeX Profile API
4. 检查响应状态（201 Created 或 409 Conflict都视为成功）

**返回值**：错误对象

---

### 5.6 内部辅助函数

#### generateProfileName
```go
func (p *Provisioner) generateProfileName(templateName string) string
```

**功能**：从模板名称生成Profile名称

**处理逻辑**：
1. 如果模板名称包含路径，提取最后一部分
2. 去除.yaml后缀
3. 添加"-profile"后缀

**返回值**：生成的Profile名称

**示例**：
- 输入: "templates/modbus-sensor.yaml"
- 输出: "modbus-sensor-profile"

#### buildProtocolConfig
```go
func (p *Provisioner) buildProtocolConfig(protocol, ip string) map[string]map[string]string
```

**功能**：根据协议类型构建协议配置

**支持的协议**：
- `modbus` / `modbus-tcp`: Modbus-TCP配置
- `modbus-rtu`: Modbus-RTU配置
- `mqtt`: MQTT配置
- 其他: 通用配置

**返回值**：协议配置map

#### registerDevice
```go
func (p *Provisioner) registerDevice(device DeviceV2) error
```

**功能**：将设备注册到EdgeX（V2 API）

**处理逻辑**：
1. 将DeviceV2对象序列化为JSON
2. 创建POST请求（Content-Type: application/json）
3. 发送到EdgeX Device API
4. 检查响应状态（201 Created 或 409 Conflict都视为成功）

**返回值**：错误对象

#### GetHTTPClient
```go
func (p *Provisioner) GetHTTPClient() *http.Client
```

**功能**：获取HTTP客户端实例

**返回值**：HTTP客户端指针

## 6. 协议配置详解

### 6.1 Modbus-TCP

```go
{
    "modbus": {
        "Address": "192.168.1.100",  // 设备IP地址
        "Port":    "502",            // Modbus端口
        "UnitID":  "1"               // Modbus单元ID
    }
}
```

### 6.2 Modbus-RTU

```go
{
    "modbus-rtu": {
        "Address":  "192.168.1.100",  // 设备IP地址
        "BaudRate": "9600",           // 波特率
        "DataBits": "8",              // 数据位
        "StopBits": "1",              // 停止位
        "Parity":   "none",           // 校验位
        "UnitID":   "1"               // Modbus单元ID
    }
}
```

### 6.3 MQTT

```go
{
    "mqtt": {
        "BrokerURL": "192.168.1.100",      // MQTT Broker地址
        "ClientID":  "device-mqtt",        // 客户端ID
        "Topic":     "devices/192.168.1.100/events"  // 设备主题
    }
}
```

## 7. 使用示例

### 7.1 基本使用

```go
// 创建设备配置管理器
cm := devconfig.NewConfigManager("devices.csv")

// 创建设备配置器
provisioner := provision.NewProvisioner(cm, "")

// 验证EdgeX连接
if err := provisioner.ValidateConnection(); err != nil {
    log.Fatalf("EdgeX connection failed: %v", err)
}

// 批量配置所有设备
if err := provisioner.ProvisionAll(); err != nil {
    log.Printf("Some devices failed to provision: %v", err)
}
```

### 7.2 单个设备配置

```go
// 创建设备配置器
provisioner := provision.NewProvisioner(nil, "")

// 配置单个Modbus设备
err := provisioner.ProvisionDevice(
    name:             "Temperature-Sensor-01",
    ip:               "192.168.1.100",
    protocol:         "modbus",
    templateName:     "templates/modbus-temp.yaml",
    interval:         "10s",
    onChange:         false,
    minInterval:      "",
    subscriptionTopic: "devices/temp-01/data",
)

if err != nil {
    log.Printf("Failed to provision device: %v", err)
}
```

### 7.3 设备管理

```go
provisioner := provision.NewProvisioner(nil, "")

// 列出所有设备
provisioner.ListDevices()

// 删除指定设备
err := provisioner.RemoveDevice("Temperature-Sensor-01")
if err != nil {
    log.Printf("Failed to remove device: %v", err)
}
```

### 7.4 上传Device Profile

```go
provisioner := provision.NewProvisioner(nil, "")

// 上传设备Profile
err := provisioner.UploadProfile("/path/to/profile.yaml")
if err != nil {
    log.Printf("Failed to upload profile: %v", err)
}
```

## 8. 错误处理

### 8.1 常见错误及处理方式

| 错误场景 | 处理方式 |
|---------|---------|
| 设备名称为空 | 返回错误：device name is required |
| 设备IP为空 | 返回错误：device IP is required |
| EdgeX连接失败 | 返回错误：cannot connect to EdgeX Core Metadata |
| 设备注册失败 | 返回错误：failed to register device |
| Profile上传失败 | 返回错误：failed to upload profile |
| HTTP请求超时 | 返回错误：context deadline exceeded |

### 8.2 HTTP状态码处理

| 状态码 | 说明 | 处理方式 |
|--------|------|---------|
| 200 OK | 请求成功 | 继续处理 |
| 201 Created | 资源创建成功 | 继续处理 |
| 204 No Content | 删除成功 | 继续处理 |
| 404 Not Found | 资源不存在 | 删除操作视为成功 |
| 409 Conflict | 资源冲突（已存在） | 上传/注册视为成功 |
| 500 Internal Server Error | 服务器错误 | 返回错误 |

## 9. 配置要求

### 9.1 设备配置文件格式 (devices.csv)

```csv
Name,IP,Protocol,Template,Interval,OnChange,MinInterval,SubscriptionTopic
Temperature-Sensor-01,192.168.1.100,modbus,templates/modbus-temp.yaml,15s,false,,devices/temp-01/data
Humidity-Sensor-01,192.168.1.101,modbus,templates/modbus-humid.yaml,15s,false,,devices/humid-01/data
Power-Meter-01,192.168.1.102,mqtt,,30s,true,,devices/power-01/data
```

### 9.2 必需字段

| 字段 | 说明 | 必需 |
|------|------|------|
| Name | 设备名称 | 是 |
| IP | 设备IP地址 | 是 |
| Protocol | 协议类型 | 否（默认modbus） |
| Template | 配置模板 | 否 |
| Interval | 采集间隔 | 否（默认15s） |
| OnChange | 值变化时触发 | 否（默认false） |
| MinInterval | 最小采集间隔 | 否 |
| SubscriptionTopic | MQTT订阅主题 | 否 |

## 10. 性能优化

### 10.1 批量配置优化

- 使用goroutine并发配置设备（需要控制并发数）
- 配置前先验证EdgeX连接
- 使用批量操作减少API调用次数

### 10.2 HTTP客户端优化

```go
httpClient: &http.Client{
    Timeout: 30 * time.Second,  // 合理的超时设置
    Transport: &http.Transport{
        MaxIdleConns:        100,      // 最大空闲连接数
        IdleConnTimeout:     90 * time.Second,  // 空闲连接超时
    },
}
```

### 10.3 并发控制

```go
// 使用信号量控制并发数
sem := make(chan struct{}, 10)  // 最多10个并发

for _, device := range devices {
    sem <- struct{}{}
    go func(d Device) {
        defer func() { <-sem }()
        p.ProvisionDevice(...)
    }(device)
}
```

## 11. 安全性

### 11.1 输入验证

```go
// 设备名称验证
if name == "" {
    return fmt.Errorf("device name is required")
}

// IP地址格式验证（可以使用正则表达式）
ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
if !ipRegex.MatchString(ip) {
    return fmt.Errorf("invalid IP address format")
}
```

### 11.2 网络安全

- EdgeX API地址使用localhost（生产环境应配置HTTPS）
- HTTP客户端设置合理的超时时间
- 敏感信息（如密码）不应在日志中输出

## 12. 注意事项

1. **EdgeX依赖**：设备配置器依赖EdgeX Core Metadata服务运行
2. **设备名称唯一性**：EdgeX要求设备名称全局唯一
3. **协议配置**：不同协议需要不同的配置参数
4. **Profile管理**：设备注册前需要确保相关Profile已上传
5. **错误恢复**：部分设备配置失败不应中断其他设备的配置流程
6. **幂等性**：重复注册同一设备会返回409 Conflict，但不会报错

## 13. 总结

设备配置器是sfsEdgeStore与EdgeX Foundry集成的核心组件，通过自动化的设备配置和注册，大大简化了物联网设备的管理和部署。其设计遵循以下原则：

- **自动化**：支持批量自动配置，减少人工干预
- **标准化**：遵循EdgeX API规范，支持标准协议
- **可靠性**：完善的错误处理，保证配置过程的稳定性
- **灵活性**：支持多种协议和配置模板
- **易用性**：简单的API设计，便于集成和使用