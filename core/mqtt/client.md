# client.go 技术文档

## 1. 概述

`client.go` 是 sfsEdgeStore 系统的 MQTT 客户端核心模块，负责与 EdgeX Foundry 的 MQTT Broker 进行通信。该模块实现了消息订阅、接收、批量处理、设备管理和数据转发等功能。

### 1.1 主要功能

- 连接和管理 MQTT Broker 连接
- 订阅 EdgeX 事件主题
- 接收和处理 IoT 设备数据
- Worker Pool 模式限制并发处理
- 批量消息发布（带 Gzip 压缩）
- 设备数量限制（许可证管理）
- TLS/SSL 安全通信
- 动态订阅管理

## 2. 架构设计

### 2.1 组件关系

```
Client
├── mqtt.Client           # Paho MQTT 客户端实例
├── config.Config         # 应用配置
├── queue.Queue           # 数据队列（失败重试）
├── monitor.Monitor       # 系统监控
├── analyzer.Analyzer     # 数据分析（异常检测）
├── filter.FilterManager  # 数据过滤器
├── broadcast.Broadcaster # WebSocket 广播
├── messageQueue          # Worker 消息队列（chan）
└── registeredDevices     # 已注册设备集合
```

### 2.2 消息处理流程

```
MQTT Broker ──→ messageHandler() ──→ messageQueue (chan, 2000)
                                          │
                                          ▼
                                    Worker Pool (2 个 Goroutine)
                                          │
                                          ▼
                                  processMessageAsync()
                                          │
                        ┌─────────────────┼─────────────────┐
                        ▼                 ▼                 ▼
                  edgex.Parse()     isDeviceAllowed()   storeData()
                        │                 │                 │
                        ▼                 ▼                 ▼
                  EdgeXEvent          设备注册/限制    数据库写入 + 广播 + 分析
```

## 3. 数据结构

### 3.1 Client 结构体

```go
type Client struct {
    client            mqtt.Client              // Paho MQTT 客户端
    config            *config.Config           // 应用配置
    dataQueue         *queue.Queue             // 数据队列
    monitor           *monitor.Monitor         // 监控器
    analyzer          *analyzer.Analyzer       // 分析器
    filterManager     *filter.FilterManager    // 过滤器管理器
    batchMessages     []map[string]interface{} // 批量消息缓冲
    batchSize         int                      // 批量大小阈值（100）
    batchInterval     time.Duration            // 批量时间阈值（5秒）
    lastBatchTime     time.Time                // 上次批量处理时间
    registeredDevices map[string]bool          // 已注册设备集合
    broadcaster       broadcast.Broadcaster    // WebSocket 广播器
    messageQueue      chan mqtt.Message        // Worker 消息队列
}
```

### 3.2 常量定义

```go
const (
    workerQueueSize = 2000  // 消息队列缓冲区大小
    workerCount     = 2     // 工作协程数量
)
```

### 3.3 EdgeX 主题定义

```go
var standardTopics = []string{
    "edgex/events/#",      // EdgeX v2/v3 标准主题
    "devices/+/data",      // 设备数据主题
    "edgex/events/core/#", // EdgeX v1 核心主题
}
```

## 4. 核心函数

### 4.1 NewClient - 创建 MQTT 客户端

```go
func NewClient(cfg *config.Config, dataQueue *queue.Queue, 
    monitor *monitor.Monitor, analyzer *analyzer.Analyzer, 
    broadcaster broadcast.Broadcaster) (*Client, error)
```

**功能**：
1. 初始化 MQTT 客户端选项
2. 配置连接参数（Broker 地址、ClientID、重连策略）
3. 设置遗嘱消息（LWT - Last Will and Testament）
4. 配置用户名/密码认证
5. 初始化 Client 结构体字段
6. 启动 Worker Pool（2 个 Goroutine）
7. 配置 TLS（如果启用）
8. 执行安全检查
9. 设置连接成功/丢失回调
10. 连接到 MQTT Broker

**连接选项配置**：
- `CleanSession`: true（不保留会话状态）
- `AutoReconnect`: true（自动重连）
- `MaxReconnectInterval`: 5 分钟
- `ResumeSubs`: true（重连后自动重新订阅）

**遗嘱消息**：
- 主题：`edgex/events/status`
- 内容：`{"status": "offline", "clientId": "...", "timestamp": ...}`

### 4.2 messageHandler - 消息处理器

```go
func (c *Client) messageHandler() mqtt.MessageHandler
```

**功能**：
- 返回一个闭包函数作为 MQTT 消息回调
- 记录消息接收统计
- 将消息放入 Worker 队列（非阻塞）
- 队列满时记录错误并丢弃消息

### 4.3 messageWorker - 工作协程

```go
func (c *Client) messageWorker(workerID int)
```

**功能**：
- 从 `messageQueue` 通道接收消息
- 调用 `processMessageAsync()` 异步处理每条消息
- 持续运行直到通道关闭

### 4.4 subscribeTopics - 订阅主题

```go
func (c *Client) subscribeTopics(mqttClient mqtt.Client) error
```

**功能**：
- 获取要订阅的主题列表（`getTopicsToSubscribe()`）
- 遍历主题并逐个订阅
- QoS 级别：1（至少一次投递）
- 记录订阅结果

### 4.5 getTopicsToSubscribe - 获取订阅主题

```go
func (c *Client) getTopicsToSubscribe() []string
```

**逻辑**：
1. 如果配置了 `MQTTTopic`，添加到列表
2. 如果启用 `AutoSubscribe`：
   - 检测 EdgeX 版本（`detectEdgeXVersion()`）
   - 根据版本获取对应主题
   - 版本未知时使用标准主题
3. 对主题列表去重后返回

### 4.6 detectEdgeXVersion - 检测 EdgeX 版本

```go
func (c *Client) detectEdgeXVersion() (string, error)
```

**版本常量**：
- `EdgeXVersionUnknown` = "unknown"
- `EdgeXVersionV1` = "v1"
- `EdgeXVersionV2` = "v2"
- `EdgeXVersionLatest` = "latest"

**检测策略**：
1. 尝试订阅 V2 主题，成功则返回 "v2"
2. 尝试订阅 V1 主题，成功则返回 "v1"
3. 无法确定时返回 "latest"

### 4.7 createTLSConfig - 创建 TLS 配置

```go
func (c *Client) createTLSConfig(cfg *config.Config) (*tls.Config, error)
```

**支持模式**：
- **单向 TLS**：仅验证服务器证书
  - 可配置自定义 CA 证书（`MQTTCACert`）
  - 或使用系统默认证书池
- **双向 TLS（mTLS）**：客户端也提供证书
  - 需要 `MQTTClientCert` 和 `MQTTClientKey`

### 4.8 performSecurityCheck - 安全检查

```go
func (c *Client) performSecurityCheck(cfg *config.Config)
```

**检查项**：
- 是否启用用户名/密码认证
- 是否启用 TLS 加密
- 是否配置 CA 证书
- 是否启用双向 TLS
- 密码强度（至少 8 个字符）

### 4.9 PublishBatch - 批量发布

```go
func (c *Client) PublishBatch(topic string, qos byte, messages []map[string]any) error
```

**流程**：
1. 将消息序列化为 JSON
2. 使用 Gzip 压缩 JSON 数据
3. 发布压缩后的消息

### 4.10 processBatchMessages - 处理批量消息

```go
func (c *Client) processBatchMessages()
```

**触发条件**：
- 消息数量达到 `batchSize`（100 条）
- 或时间间隔达到 `batchInterval`（5 秒）

**处理逻辑**：
1. 发布批量消息到 `{MQTTTopic}/batch`
2. 发布失败时消息入队重试
3. 发布成功时更新监控计数
4. 清空批量缓冲区

### 4.11 isDeviceAllowed - 设备许可检查

```go
func (c *Client) isDeviceAllowed(deviceName string) bool
```

**设备数量限制**：

| 许可证类型 | 最大设备数 |
|-----------|-----------|
| 社区版     | 5 台      |
| 商业版     | 50 台     |
| 企业版     | 无限制    |

**逻辑**：
1. 企业版用户始终允许
2. 已注册设备始终允许
3. 检查是否达到设备数量上限
4. 未达上限时注册新设备并返回 true

### 4.12 动态订阅管理

```go
func (c *Client) AddSubscription(topic string) error
func (c *Client) RemoveSubscription(topic string) error
func (c *Client) GetSubscriptions() []string
```

## 5. Worker Pool 模式

### 5.1 设计目的

- 限制并发处理的 Goroutine 数量
- 防止高并发场景下内存泄漏
- 提供可控的消息处理吞吐量

### 5.2 参数说明

| 参数 | 值 | 说明 |
|------|-----|------|
| `workerCount` | 2 | 工作协程数量 |
| `workerQueueSize` | 2000 | 消息队列缓冲区大小 |

### 5.3 启动时机

在 `NewClient()` 中启动：

```go
for i := 0; i < workerCount; i++ {
    go client.messageWorker(i)
}
```

### 5.4 消息流转

```
messageHandler()                    // 接收消息
    │
    ▼
select { messageQueue <- msg }     // 非阻塞入队
    │
    ▼
messageWorker()                    // Worker 取消息
    │
    ▼
processMessageAsync()              // 处理消息
```

## 6. 连接生命周期

### 6.1 连接建立

```
NewClient()
    │
    ├─ 初始化客户端选项
    ├─ 配置 TLS（可选）
    ├─ 启动 Worker Pool
    ├─ 设置回调函数
    │     ├─ OnConnect: 发布在线状态 + 订阅主题
    │     └─ OnConnectionLost: 更新监控状态
    └─ Connect()
```

### 6.2 连接成功回调

```go
opts.SetOnConnectHandler(func(mqttClient mqtt.Client) {
    // 1. 更新监控状态为在线
    // 2. 发布在线状态消息
    // 3. 重新订阅所有主题
})
```

### 6.3 连接丢失回调

```go
opts.SetConnectionLostHandler(func(mqttClient mqtt.Client, err error) {
    // 1. 记录错误日志
    // 2. 更新监控状态为离线
})
```

### 6.4 断开连接

```go
func (c *Client) Disconnect() {
    // 1. 发送剩余的批量消息
    // 2. 断开 MQTT 连接（250ms 超时）
}
```

## 7. 依赖关系

| 依赖包 | 用途 |
|--------|------|
| `github.com/eclipse/paho.mqtt.golang` | MQTT 客户端库 |
| `sfsEdgeStore/config` | 配置管理 |
| `sfsEdgeStore/queue` | 数据队列 |
| `sfsEdgeStore/monitor` | 系统监控 |
| `sfsEdgeStore/analyzer` | 数据分析 |
| `sfsEdgeStore/filter` | 数据过滤 |
| `sfsEdgeStore/broadcast` | WebSocket 广播 |
| `sfsEdgeStore/common` | 通用工具函数 |
| `sfsEdgeStore/core/database` | 数据库操作 |
| `sfsEdgeStore/core/edgex` | EdgeX 消息解析 |

## 8. 错误处理

### 8.1 队列满处理

当 `messageQueue` 满（2000 条）时：
- 使用 `select...default` 非阻塞丢弃消息
- 记录错误到监控系统

### 8.2 发布失败处理

批量消息发布失败时：
- 消息重新入队等待重试
- 记录错误日志

### 8.3 连接失败处理

- MQTT 客户端启用自动重连
- 重连后自动重新订阅主题
- 监控状态自动更新

## 9. 性能特性

### 9.1 批量处理

- 批量大小：100 条
- 批量间隔：5 秒
- 满足任一条件即触发批量发布

### 9.2 数据压缩

- 使用 Gzip 压缩批量消息
- 减少网络传输带宽

### 9.3 并发控制

- Worker Pool 限制为 2 个并发处理
- 适合资源受限的边缘设备

## 10. 安全特性

| 特性 | 状态 |
|------|------|
| 用户名/密码认证 | 支持 |
| TLS 加密传输 | 支持 |
| 自定义 CA 证书 | 支持 |
| 双向 TLS (mTLS) | 支持 |
| 密码强度检查 | 支持（≥8 字符） |
| 遗嘱消息（LWT） | 已启用 |

## 11. 配置项

| 配置字段 | 类型 | 说明 |
|---------|------|------|
| `MQTTBroker` | string | MQTT Broker 地址 |
| `MQTTTopic` | string | 订阅的主题 |
| `ClientID` | string | 客户端 ID |
| `MQTTUsername` | string | 用户名 |
| `MQTTPassword` | string | 密码 |
| `MQTTUseTLS` | bool | 是否启用 TLS |
| `MQTTCACert` | string | CA 证书路径 |
| `MQTTClientCert` | string | 客户端证书路径 |
| `MQTTClientKey` | string | 客户端私钥路径 |
| `AutoSubscribe` | bool | 是否自动订阅标准主题 |
| `LicenseType` | string | 许可证类型（community/business/enterprise） |
| `EnterpriseFeatures.MaxDevices` | int | 最大设备数限制 |
