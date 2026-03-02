# sfsDb 与 EdgeX 消息总线无缝对接方案

与 EdgeX 消息总线（Message Bus）无缝对接，意味着你的应用或服务能够高效、可靠地融入 EdgeX Foundry 的生态系统，与其他微服务（如设备服务、核心服务、应用服务）进行标准化通信。这种"无缝"特性通常指遵循其消息格式（如 MessageEnvelope）、使用其支持的传输协议（如 Redis Pub/Sub、MQTT）并理解其内部数据模型（如 Event、Reading）。

实现这种对接可以带来以下几方面的核心益处：

🚀 **提升系统性能与解耦**
- **服务解耦**：通过消息总线，你的服务不需要直接调用其他服务的 REST API，从而降低了服务间的耦合度。发布者和订阅者互不知晓对方的存在，只需关注特定的主题（Topic），这使得系统更易于维护和扩展。
- **降低延迟与负载**：在 EdgeX 的较新版本中，越来越多的服务间通信（如设备服务到核心数据、核心命令到设备服务）从 REST 调用转向直接使用消息总线。无缝对接意味着你的服务也能享受这种高效的通信方式，减少 HTTP 握手开销，提升整体吞吐量。

🔄 **实现数据一致性与标准化**
- **统一的消息格式**：无缝对接要求使用 MessageEnvelope 封装消息，其中包含负载内容类型（如 JSON 或 CBOR）、相关 ID 等元数据。这保证了整个 EdgeX 系统内消息格式的一致性，便于日志追踪、调试和跨语言处理。
- **原生数据模型支持**：你的服务可以直接生产和消费 EdgeX 原生的 Event 和 Reading 对象，无需进行繁琐的数据结构转换，确保了数据在边缘侧的完整性和语义一致性。

🛠️ **简化开发与集成**
- **利用现有生态**：无缝对接后，你的应用服务可以像原生的 EdgeX 服务一样，轻松订阅来自各种设备服务的数据流，或者发布命令请求。这大大简化了与不同厂商设备（通过设备服务抽象）交互的开发工作量。
- **规则引擎集成**：EdgeX 的 eKuiper 规则引擎与消息总线紧密集成。无缝对接总线意味着你可以直接利用 eKuiper 对实时数据流进行过滤、分析和转发，而无需自己编写复杂的流处理逻辑。

🌐 **增强可靠性与灵活性**
- **支持多种传输后端**：EdgeX 消息总线抽象了底层实现（如 Redis Pub/Sub 默认、MQTT 3.1 等）。无缝对接意味着你的服务具有灵活性，可以在不同的部署环境中切换底层消息中间件，而无需修改业务逻辑代码。
- **系统事件感知**：除了设备数据，消息总线还传输系统指标和事件。无缝对接让你的服务能够监听这些系统级消息，从而实现更智能的运维和自适应行为（例如，在系统负载过高时降低数据采集频率）。

📌 **总结**
简而言之，与 EdgeX 消息总线无缝对接，是构建高性能、高内聚、低耦合边缘应用的关键。它让你的解决方案能够"原生"地运行在 EdgeX 框架内，充分享受其作为边缘计算框架带来的标准化、解耦和生态优势。

要实现 sfsDb 与 EdgeX 消息总线（MQTT/ZeroMQ）的无缝对接，需要设计一个适配器组件，实现数据的即插即用。以下是详细的实现方案：


## 1. 适配器架构设计

### 核心组件
- **消息订阅模块**：订阅 EdgeX 消息总线上的设备数据主题
- **数据转换模块**：将 EdgeX 消息格式转换为 sfsDb 存储格式
- **数据存储模块**：将转换后的数据写入 sfsDb
- **状态管理模块**：处理连接状态、错误重试等

### 架构图
```
┌─────────────┐     ┌────────────────┐     ┌─────────────┐
│ EdgeX设备   │────>│ EdgeX消息总线  │────>│ sfsDb适配器  │────>│ sfsDb存储  │
│ (传感器等)  │     │ (MQTT/ZeroMQ) │     │             │     │           │
└─────────────┘     └────────────────┘     └─────────────┘     └─────────────┘
```


## 2. 技术实现方案

### 2.1 消息订阅实现
- **MQTT 订阅**：使用 Go 的 `paho.mqtt.golang` 库连接 EdgeX MQTT 总线
- **ZeroMQ 订阅**：使用 `github.com/pebbe/zmq4` 库连接 EdgeX ZeroMQ 总线
- **主题配置**：支持配置需要订阅的 EdgeX 数据主题（如 `edgex/events/core/#`）

### 2.2 数据转换逻辑
- **消息解析**：解析 EdgeX 标准消息格式（JSON）
- **数据映射**：将 EdgeX 消息字段映射到 sfsDb 表结构
- **批量处理**：支持批量转换和存储，提高性能

### 2.3 数据存储策略
- **表结构设计**：根据 EdgeX 数据模型设计 sfsDb 表结构
- **索引优化**：为常用查询字段创建索引
- **过期策略**：支持数据过期和自动清理

### 2.4 可靠性保障
- **连接管理**：自动重连机制
- **消息缓存**：本地缓存未处理的消息
- **错误处理**：完善的错误日志和告警机制


## 3. 实现步骤

### 步骤 1：环境准备
- 安装 EdgeX Foundry（推荐使用 Docker 部署）
- 确保 sfsDb 已安装并运行
- 配置 EdgeX 消息总线（MQTT 或 ZeroMQ）

### 步骤 2：开发适配器
- 创建 Go 项目，引入必要依赖
- 实现消息订阅模块
- 实现数据转换模块
- 实现数据存储模块
- 实现状态管理和错误处理

### 步骤 3：配置与部署
- 配置 EdgeX 消息总线连接参数
- 配置 sfsDb 连接参数
- 配置数据映射规则
- 部署适配器服务

### 步骤 4：测试与验证
- 模拟 EdgeX 设备数据
- 验证数据是否正确存储到 sfsDb
- 测试异常情况（网络中断、消息丢失等）


## 4. 代码示例

```go
// sfsDb 与 EdgeX MQTT 适配器示例（改进版）
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

// 配置结构体
type Config struct {
    DBPath     string `json:"db_path"`
    MQTTBroker string `json:"mqtt_broker"`
    MQTTTopic  string `json:"mqtt_topic"`
    ClientID   string `json:"client_id"`
}

// EdgeX 消息结构
type EdgeXMessage struct {
    ID          string          `json:"id"`
    DeviceName  string          `json:"deviceName"`
    Reading     string          `json:"reading"`
    Value       float64         `json:"value"`
    Timestamp   time.Time       `json:"timestamp"`
    Metadata    json.RawMessage `json:"metadata,omitempty"`
}

var table *engine.Table
var config Config

func main() {
    // 加载配置
    if err := loadConfig(); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 连接 sfsDb
    if err := initDatabase(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // 初始化 MQTT 客户端
    client, err := initMQTT()
    if err != nil {
        log.Fatalf("Failed to initialize MQTT: %v", err)
    }
    defer client.Disconnect(250)

    // 订阅 EdgeX 消息
    if err := subscribeToEdgeX(client); err != nil {
        log.Fatalf("Failed to subscribe to EdgeX messages: %v", err)
    }

    log.Println("sfsDb EdgeX adapter started successfully")

    // 保持运行
    select {}
}

// 加载配置
func loadConfig() error {
    // 默认配置
    config = Config{
        DBPath:     "./edgex_data",
        MQTTBroker: "tcp://localhost:1883",
        MQTTTopic:  "edgex/events/core/#",
        ClientID:   fmt.Sprintf("sfsdb-edgex-adapter-%d", time.Now().Unix()),
    }

    // 尝试从配置文件加载
    configFile := "config.json"
    if _, err := os.Stat(configFile); err == nil {
        data, err := os.ReadFile(configFile)
        if err != nil {
            return fmt.Errorf("failed to read config file: %v", err)
        }
        if err := json.Unmarshal(data, &config); err != nil {
            return fmt.Errorf("failed to parse config file: %v", err)
        }
        log.Println("Loaded config from file")
    } else {
        log.Println("Using default config")
    }

    return nil
}

// 初始化数据库
func initDatabase() error {
    // 确保数据库目录存在
    if err := os.MkdirAll(config.DBPath, 0755); err != nil {
        return fmt.Errorf("failed to create database directory: %v", err)
    }

    // 打开数据库
    _, err := storage.GetDBManager().OpenDB(config.DBPath)
    if err != nil {
        return fmt.Errorf("failed to open database: %v", err)
    }

    // 创建或获取表
    tableName := "edgex_readings"
    var createErr error
    table, createErr = engine.TableNew(tableName)
    if createErr != nil {
        return fmt.Errorf("failed to create table: %v", createErr)
    }

    // 设置表字段
    fields := map[string]any{
        "id":         "",
        "deviceName": "",
        "reading":    "",
        "value":      0.0,
        "timestamp":  0,
        "metadata":   "",
    }
    if err := table.SetFields(fields); err != nil {
        return fmt.Errorf("failed to set table fields: %v", err)
    }

    // 创建主键索引
    primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
    if err != nil {
        return fmt.Errorf("failed to create primary key: %v", err)
    }
    primaryKey.AddFields("id")
    if err := table.CreateIndex(primaryKey); err != nil {
        // 忽略索引已存在的错误
        if err.Error() != "index already exists" {
            return fmt.Errorf("failed to create primary key index: %v", err)
        }
    }

    // 创建设备名称索引
    deviceIndex, err := engine.DefaultNormalIndexNew("device_index")
    if err != nil {
        return fmt.Errorf("failed to create device index: %v", err)
    }
    deviceIndex.AddFields("deviceName")
    if err := table.CreateIndex(deviceIndex); err != nil {
        // 忽略索引已存在的错误
        if err.Error() != "index already exists" {
            return fmt.Errorf("failed to create device index: %v", err)
        }
    }

    // 创建时间戳索引
    timeIndex, err := engine.DefaultNormalIndexNew("time_index")
    if err != nil {
        return fmt.Errorf("failed to create time index: %v", err)
    }
    timeIndex.AddFields("timestamp")
    if err := table.CreateIndex(timeIndex); err != nil {
        // 忽略索引已存在的错误
        if err.Error() != "index already exists" {
            return fmt.Errorf("failed to create time index: %v", err)
        }
    }

    log.Println("Database initialized successfully")
    return nil
}

// 初始化 MQTT 客户端
func initMQTT() (mqtt.Client, error) {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(config.MQTTBroker)
    opts.SetClientID(config.ClientID)
    opts.SetCleanSession(true)
    opts.SetAutoReconnect(true)
    opts.SetMaxReconnectInterval(time.Second * 30)
    opts.SetDefaultPublishHandler(messageHandler())

    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
    }

    log.Printf("Connected to MQTT broker: %s", config.MQTTBroker)
    return client, nil
}

// 订阅 EdgeX 消息
func subscribeToEdgeX(client mqtt.Client) error {
    token := client.Subscribe(config.MQTTTopic, 1, nil)
    if token.Wait() && token.Error() != nil {
        return fmt.Errorf("failed to subscribe to topic %s: %v", config.MQTTTopic, token.Error())
    }

    log.Printf("Subscribed to topic: %s", config.MQTTTopic)
    return nil
}

// 验证 EdgeX 消息
func validateMessage(msg *EdgeXMessage) error {
    if msg.ID == "" {
        return fmt.Errorf("missing message ID")
    }
    if msg.DeviceName == "" {
        return fmt.Errorf("missing device name")
    }
    if msg.Reading == "" {
        return fmt.Errorf("missing reading type")
    }
    if msg.Timestamp.IsZero() {
        msg.Timestamp = time.Now()
    }
    return nil
}

// 消息处理函数
func messageHandler() mqtt.MessageHandler {
    return func(client mqtt.Client, msg mqtt.Message) {
        log.Printf("Received message on topic: %s", msg.Topic())

        var edgexMsg EdgeXMessage
        if err := json.Unmarshal(msg.Payload(), &edgexMsg); err != nil {
            log.Printf("Failed to parse message: %v", err)
            return
        }

        // 验证消息
        if err := validateMessage(&edgexMsg); err != nil {
            log.Printf("Invalid message: %v", err)
            return
        }

        // 准备数据
        metadataStr := ""
        if edgexMsg.Metadata != nil {
            metadataStr = string(edgexMsg.Metadata)
        }

        data := map[string]any{
            "id":         edgexMsg.ID,
            "deviceName": edgexMsg.DeviceName,
            "reading":    edgexMsg.Reading,
            "value":      edgexMsg.Value,
            "timestamp":  edgexMsg.Timestamp.Unix(),
            "metadata":   metadataStr,
        }

        // 存储到 sfsDb
        _, err := table.Insert(&data)
        if err != nil {
            log.Printf("Failed to store data: %v", err)
        } else {
            log.Printf("Stored reading from %s: %s = %f", 
                edgexMsg.DeviceName, edgexMsg.Reading, edgexMsg.Value)
        }
    }
}
```


## 5. 优势与价值

### 技术优势
- **轻量高效**：sfsDb 的低资源占用与 EdgeX 边缘计算理念高度契合
- **即插即用**：适配器实现后，无需修改 EdgeX 核心组件
- **可靠性高**：支持离线存储和网络恢复后的数据同步
- **易于扩展**：可根据业务需求添加数据处理逻辑

### 业务价值
- **数据本地化**：边缘数据本地存储，减少云端依赖
- **实时分析**：支持边缘侧实时数据查询和分析
- **成本降低**：减少网络传输和云存储成本
- **合规性**：敏感数据本地处理，满足数据隐私要求


## 6. 集成建议

1. **作为 EdgeX 服务部署**：将适配器打包为 Docker 容器，与 EdgeX 其他服务一起部署
2. **配置管理**：使用 EdgeX 的配置服务管理适配器配置
3. **监控集成**：将适配器状态纳入 EdgeX 监控系统
4. **文档完善**：提供详细的部署和使用文档

通过这种无缝对接方案，sfsDb 可以成为 EdgeX 生态系统中的重要数据存储组件，为边缘计算应用提供高效、可靠的数据持久化解决方案。