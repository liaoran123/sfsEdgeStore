# MQTT 客户端实现

## 概述

MQTT 客户端是 sfsEdgeStore 与 EdgeX Foundry 通信的核心模块，负责订阅 EdgeX 事件消息、解析数据并存储到本地数据库。

## 核心结构

### Client 结构体

```go
// core/mqtt/client.go:27-37
type Client struct {
	client        mqtt.Client
	config        *config.Config
	dataQueue     *queue.Queue
	monitor       *monitor.Monitor
	analyzer      *analyzer.Analyzer
	batchMessages []map[string]interface{}
	batchSize     int
	batchInterval time.Duration
	lastBatchTime time.Time
}
```

**字段说明：**

- `client`: 底层 MQTT 客户端（来自 paho.mqtt.golang）
- `config`: 配置管理
- `dataQueue`: 数据队列，用于故障恢复
- `monitor`: 监控集成
- `analyzer`: 数据分析集成
- `batchMessages`: 批量消息缓冲区
- `batchSize/batchInterval`: 批量控制参数

## 创建客户端

### NewClient 函数

```go
// core/mqtt/client.go:40-136
func NewClient(cfg *config.Config, dataQueue *queue.Queue, monitor *monitor.Monitor, analyzer *analyzer.Analyzer) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker) // 连接到EdgeX的MQTT broker
	opts.SetClientID(cfg.ClientID)
	opts.SetCleanSession(true)                    // 启用清理会话
	opts.SetAutoReconnect(true)                   // 启用自动重连
	opts.SetMaxReconnectInterval(time.Minute * 5) // 最大重连间隔5分钟
	opts.SetResumeSubs(true)                      // 连接恢复后自动重新订阅

	// 设置遗嘱消息（不要在遗嘱主题中使用通配符 #）
	willTopic := "edgex/events/status"
	willMessage := map[string]interface{}{
		"status":    "offline",
		"clientId":  cfg.ClientID,
		"timestamp": time.Now().UnixNano(),
	}
	willPayload, _ := json.Marshal(willMessage)
	opts.SetWill(willTopic, string(willPayload), 0, false)

	// 添加 TLS 支持
	if cfg.MQTTUseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		// 加载 CA 证书
		if cfg.MQTTCACert != "" {
			caCert, err := os.ReadFile(cfg.MQTTCACert)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %v", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
		// 加载客户端证书和密钥
		if cfg.MQTTClientCert != "" && cfg.MQTTClientKey != "" {
			cert, err := tls.LoadX509KeyPair(cfg.MQTTClientCert, cfg.MQTTClientKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load client cert: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		opts.SetTLSConfig(tlsConfig)
	}

	client := &Client{
		config:        cfg,
		dataQueue:     dataQueue,
		monitor:       monitor,
		analyzer:      analyzer,
		batchMessages: make([]map[string]interface{}, 0),
		batchSize:     100,             // 默认批量大小
		batchInterval: 5 * time.Second, // 默认批量间隔
		lastBatchTime: time.Now(),
	}

	// 设置连接处理函数
	opts.SetOnConnectHandler(func(mqttClient mqtt.Client) {
		log.Println("MQTT broker connected")
		// 发布在线状态消息
		onlineTopic := "edgex/events/status"
		onlineMessage := map[string]interface{}{
			"status":    "online",
			"clientId":  cfg.ClientID,
			"timestamp": time.Now().UnixNano(),
		}
		onlinePayload, _ := json.Marshal(onlineMessage)
		token := mqttClient.Publish(onlineTopic, 1, false, onlinePayload)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Failed to publish online status: %v", token.Error())
		}
		// 重新订阅主题
		token = mqttClient.Subscribe(cfg.MQTTTopic, 1, client.messageHandler())
		token.Wait()
		if token.Error() != nil {
			log.Printf("Failed to resubscribe to topic %s: %v", cfg.MQTTTopic, token.Error())
		} else {
			log.Printf("Resubscribed to topic: %s", cfg.MQTTTopic)
		}
	})

	opts.SetConnectionLostHandler(func(mqttClient mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	mqttClient := mqtt.NewClient(opts)
	token := mqttClient.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	client.client = mqttClient

	log.Println("Connected to MQTT broker for agent")
	return client, nil
}
```

**关键特性：**

- 清理会话：`SetCleanSession(true)`，确保每次连接都是新会话
- 自动重连：`SetAutoReconnect(true)`，网络中断后自动重新连接
- 恢复订阅：`SetResumeSubs(true)`，连接恢复后自动重新订阅主题
- 遗嘱消息：异常断开时通知
- TLS 支持：双向认证

## 消息处理

### 消息处理函数

```go
// core/mqtt/client.go:238-396
func (c *Client) messageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		// 增加MQTT消息接收计数
		if c.monitor != nil {
			c.monitor.IncrementMQTTMessagesReceived()
		}

		log.Printf("Received message on topic: %s", msg.Topic())

		// 使用goroutine异步处理消息，避免阻塞MQTT消息接收
		go func() {
			// 使用edgex包处理消息
			event, err := edgex.ProcessMessage(msg.Payload())
			if err != nil {
				log.Printf("Failed to process message: %v", err)
				return
			}

			// 如果消息类型不是event，event会为nil
			if event == nil {
				return
			}

			// 预分配切片容量，避免动态扩容
			records := make([]*map[string]any, 0, len(event.Readings))

			// 处理每个读数
			for _, reading := range event.Readings {
				// 从对象池获取map，减少内存分配
				data := objPool.GetMap()

				// 准备数据
				metadataStr := ""
				if reading.Metadata != nil {
					metadataStr = string(reading.Metadata)
				}

				// 解析值的类型
				value := common.ParseValue(reading.Value)

				data["id"] = reading.ID
				data["deviceName"] = event.DeviceName // 设备名称已经在ProcessMessage中格式化
				data["reading"] = reading.ResourceName
				data["value"] = value
				data["valueType"] = reading.ValueType
				data["baseType"] = reading.BaseType
				data["timestamp"] = reading.Origin // 纳秒级时间戳，类型为 int64
				data["metadata"] = metadataStr

				records = append(records, &data)
			}

			// 批量存储到 sfsDb
			if len(records) > 0 {
				// 增加数据库操作计数
				if c.monitor != nil {
					c.monitor.IncrementDatabaseOperations()
				}

				// 使用重试机制插入数据
				err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
				if err != nil {
					log.Printf("Failed to batch store data after retries: %v", err)

					// 分析错误类型，针对边缘设备常见故障进行处理
					errorMsg := err.Error()

					// 边缘设备常见故障类型判断
					if strings.Contains(errorMsg, "no space left") ||
						strings.Contains(errorMsg, "disk full") ||
						strings.Contains(errorMsg, "file system") ||
						strings.Contains(errorMsg, "I/O error") {
						// 磁盘空间不足或文件系统错误，属于致命错误，重试无效
						log.Printf("Fatal storage error detected: %v", err)

						// 触发监控告警
						if c.monitor != nil {
							c.monitor.RecordError("storage_error", errorMsg)
						}
					} else if strings.Contains(errorMsg, "lock") ||
						strings.Contains(errorMsg, "busy") {
						// 锁竞争或资源忙，短暂重试可能有效
						log.Printf("Resource contention error detected: %v", err)
						if c.monitor != nil {
							c.monitor.RecordError("resource_contention", errorMsg)
						}
					} else {
						// 其他错误
						log.Printf("Other database error: %v", err)
						if c.monitor != nil {
							c.monitor.RecordError("database_error", errorMsg)
						}
					}

					// 将数据加入队列，以便后续处理
					if err := c.dataQueue.Enqueue(records); err != nil {
						log.Printf("Failed to enqueue data: %v", err)
					} else {
						log.Printf("Enqueued %d readings for later processing", len(records))
					}

					// 归还map对象到池中
					for _, data := range records {
						objPool.PutMap(*data)
					}
				} else {
					log.Printf("Batch stored %d readings from %s", len(records), event.DeviceName)
					// 增加MQTT消息处理计数
					if c.monitor != nil {
						c.monitor.IncrementMQTTMessagesProcessed()
					}

					// 分析数据
					if c.analyzer != nil && c.analyzer.IsEnabled() {
						// 按reading分组分析数据
						readingDataMap := make(map[string][]map[string]interface{})
						for _, record := range records {
							// 从记录中获取reading信息
							readingName, ok := (*record)["reading"].(string)
							if !ok {
								continue
							}
							readingDataMap[readingName] = append(readingDataMap[readingName], *record)
						}

						// 对每个reading进行分析
						for readingName, analysisData := range readingDataMap {
							// 分析数据
							results, alerts := c.analyzer.Analyze(analysisData, event.DeviceName, readingName)

							// 处理分析结果
							if len(results) > 0 {
								log.Printf("Analysis completed for %s: %d results", readingName, len(results))
								// 这里可以将分析结果存储或发送到其他系统
							}

							// 处理告警
							if len(alerts) > 0 {
								log.Printf("Detected %d alerts for %s", len(alerts), readingName)
								// 这里可以将告警发送到监控系统或其他通知渠道
								for _, alert := range alerts {
									log.Printf("Alert: %s - %s - %s", alert.Severity, alert.Message, alert.Reading)
									// 触发监控告警
									if c.monitor != nil {
										c.monitor.RecordError(alert.AlertType, alert.Message)
									}
								}
							}
						}
					}

					// 归还map对象到池中
					for _, data := range records {
						objPool.PutMap(*data)
					}
				}
			}
		}()
	}
}
```

## 批量消息处理

### 批量发布功能

```go
// core/mqtt/client.go:167-177
func (c *Client) PublishBatch(topic string, qos byte, messages []map[string]interface{}) error {
	// 压缩消息
	compressedPayload, err := c.compressMessages(messages)
	if err != nil {
		return fmt.Errorf("failed to compress messages: %v", err)
	}

	// 发布压缩后的消息
	return c.Publish(topic, qos, false, compressedPayload)
}
```

### 消息压缩

```go
// core/mqtt/client.go:180-198
func (c *Client) compressMessages(messages []map[string]interface{}) ([]byte, error) {
	// 将消息序列化为JSON
	jsonData, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	// 压缩JSON数据
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gzw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
```

## 对象池

### MapPool 实现

```go
// core/mqtt/mapPool.go
package mqtt

import "sync"

var objPool = NewObjectPool()

type objectPool struct {
	mapPool sync.Pool
}

func NewObjectPool() *objectPool {
	return &objectPool{
		mapPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]any)
			},
		},
	}
}

func (p *objectPool) GetMap() map[string]any {
	m := p.mapPool.Get().(map[string]any)
	for k := range m {
		delete(m, k)
	}
	return m
}

func (p *objectPool) PutMap(m map[string]any) {
	for k := range m {
		delete(m, k)
	}
	p.mapPool.Put(m)
}
```

## 测试

### 对象池测试

```go
// core/mqtt/mqtt_test.go
package mqtt

import (
	"testing"
)

func TestObjectPool(t *testing.T) {
	m1 := objPool.GetMap()
	if m1 == nil {
		t.Fatal("Expected non-nil map")
	}
	if len(m1) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(m1))
	}

	m1["key1"] = "value1"
	m1["key2"] = 42

	objPool.PutMap(m1)

	m2 := objPool.GetMap()
	if m2 == nil {
		t.Fatal("Expected non-nil map")
	}
	if len(m2) != 0 {
		t.Errorf("Expected empty map after PutMap, got %d entries", len(m2))
	}

	t.Log("Object pool test passed")
}

func TestObjectPoolMultiple(t *testing.T) {
	maps := make([]map[string]any, 10)

	for i := 0; i < 10; i++ {
		maps[i] = objPool.GetMap()
		if maps[i] == nil {
			t.Fatalf("Expected non-nil map at index %d", i)
		}
		maps[i]["index"] = i
	}

	for i := 0; i < 10; i++ {
		objPool.PutMap(maps[i])
	}

	for i := 0; i < 10; i++ {
		m := objPool.GetMap()
		if len(m) != 0 {
			t.Errorf("Expected empty map at index %d, got %d entries", i, len(m))
		}
		objPool.PutMap(m)
	}

	t.Log("Multiple object pool test passed")
}
```

### 运行测试

```bash
go test ./core/mqtt -v
```

## API 接口

### Subscribe 订阅主题

```go
func (c *Client) Subscribe() error
```

### Disconnect 断开连接

```go
func (c *Client) Disconnect()
```

### Publish 发布消息

```go
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error
```

### PublishBatch 批量发布

```go
func (c *Client) PublishBatch(topic string, qos byte, messages []map[string]interface{}) error
```

### AddToBatch 添加消息到批量队列

```go
func (c *Client) AddToBatch(message map[string]interface{})
```

