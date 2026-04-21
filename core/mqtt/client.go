package mqtt

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/common"
	"sfsEdgeStore/config"
	"sfsEdgeStore/core/database"
	"sfsEdgeStore/core/edgex"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/queue"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client MQTT客户端结构体
type Client struct {
	client            mqtt.Client
	config            *config.Config
	dataQueue         *queue.Queue
	monitor           *monitor.Monitor
	analyzer          *analyzer.Analyzer
	batchMessages     []map[string]interface{}
	batchSize         int
	batchInterval     time.Duration
	lastBatchTime     time.Time
	registeredDevices map[string]bool // 已注册设备集合
}

// NewClient 创建新的MQTT客户端
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

	// 添加用户名/密码认证
	if cfg.MQTTUsername != "" {
		opts.SetUsername(cfg.MQTTUsername)
		if cfg.MQTTPassword != "" {
			opts.SetPassword(cfg.MQTTPassword)
		}
	}

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
		config:            cfg,
		dataQueue:         dataQueue,
		monitor:           monitor,
		analyzer:          analyzer,
		batchMessages:     make([]map[string]interface{}, 0),
		batchSize:         100,             // 默认批量大小
		batchInterval:     5 * time.Second, // 默认批量间隔
		lastBatchTime:     time.Now(),
		registeredDevices: make(map[string]bool),
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

// Subscribe 订阅EdgeX消息
func (c *Client) Subscribe() error {
	token := c.client.Subscribe(c.config.MQTTTopic, 1, c.messageHandler())
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %v", c.config.MQTTTopic, token.Error())
	}

	log.Printf("Subscribed to topic: %s", c.config.MQTTTopic)
	return nil
}

// Disconnect 断开MQTT连接
func (c *Client) Disconnect() {
	// 发送剩余的批量消息
	if len(c.batchMessages) > 0 {
		c.processBatchMessages()
	}
	c.client.Disconnect(250)
}

// Publish 发布消息到MQTT主题
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := c.client.Publish(topic, qos, retained, payload)
	token.Wait()
	return token.Error()
}

// PublishBatch 批量发布消息到MQTT主题
func (c *Client) PublishBatch(topic string, qos byte, messages []map[string]interface{}) error {
	// 压缩消息
	compressedPayload, err := c.compressMessages(messages)
	if err != nil {
		return fmt.Errorf("failed to compress messages: %v", err)
	}

	// 发布压缩后的消息
	return c.Publish(topic, qos, false, compressedPayload)
}

// compressMessages 压缩消息
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

// processBatchMessages 处理批量消息
func (c *Client) processBatchMessages() {
	if len(c.batchMessages) == 0 {
		return
	}

	// 发布批量消息
	topic := c.config.MQTTTopic + "/batch"
	err := c.PublishBatch(topic, 1, c.batchMessages)
	if err != nil {
		log.Printf("Failed to publish batch messages: %v", err)
		// 将消息加入队列，以便后续处理
		if err := c.dataQueue.Enqueue(c.batchMessages); err != nil {
			log.Printf("Failed to enqueue batch messages: %v", err)
		}
	} else {
		log.Printf("Published batch of %d messages", len(c.batchMessages))
		// 增加MQTT消息处理计数
		if c.monitor != nil {
			c.monitor.IncrementMQTTMessagesProcessed()
		}
	}

	// 清空批量消息
	c.batchMessages = make([]map[string]interface{}, 0)
	c.lastBatchTime = time.Now()
}

// AddToBatch 添加消息到批量队列
func (c *Client) AddToBatch(message map[string]interface{}) {
	c.batchMessages = append(c.batchMessages, message)

	// 检查是否达到批量大小或时间间隔
	if len(c.batchMessages) >= c.batchSize || time.Since(c.lastBatchTime) >= c.batchInterval {
		c.processBatchMessages()
	}
}

// isDeviceAllowed 检查设备是否被允许发送数据
// 社区版: 5台设备, 商业版: 50台设备, 企业版: 无限制
func (c *Client) isDeviceAllowed(deviceName string) bool {
	// 企业版用户无设备数量限制
	if c.config.LicenseType == "enterprise" {
		return true
	}

	// 已注册的设备始终允许
	if c.registeredDevices[deviceName] {
		return true
	}

	// 获取设备数限制
	maxDevices := 5 // 社区版默认5台
	if c.config.LicenseType == "business" {
		maxDevices = 50 // 商业版50台
	}
	// 从配置中读取（如果有配置）
	if c.config.EnterpriseFeatures.MaxDevices > 0 {
		maxDevices = c.config.EnterpriseFeatures.MaxDevices
	}

	// 检查当前设备数量是否已达上限
	if len(c.registeredDevices) >= maxDevices {
		return false
	}

	// 注册新设备
	c.registeredDevices[deviceName] = true
	log.Printf("New device registered: %s (total: %d/%d, license: %s)", deviceName, len(c.registeredDevices), maxDevices, c.config.LicenseType)
	return true
}

// GetRegisteredDeviceCount 返回已注册设备数量
func (c *Client) GetRegisteredDeviceCount() int {
	return len(c.registeredDevices)
}

// messageHandler 适配器处理收到的EdgeX消息
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

			// 检查设备数量限制（仅对免费用户生效）
			if !c.isDeviceAllowed(event.DeviceName) {
				log.Printf("Device limit reached, rejecting data from device: %s", event.DeviceName)
				if c.monitor != nil {
					c.monitor.RecordError("device_limit_reached", fmt.Sprintf("Device %s rejected due to limit", event.DeviceName))
				}
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
