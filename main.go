// sfsDb 与 EdgeX MQTT 适配器示例（改进版）
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sfsdb-edgex-adapter-enterprise/common"
	"sfsdb-edgex-adapter-enterprise/config"
	"sfsdb-edgex-adapter-enterprise/database"
	"sfsdb-edgex-adapter-enterprise/queue"
	"sfsdb-edgex-adapter-enterprise/server"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// EdgeX 消息结构（符合 MessageEnvelope 格式）
type EdgeXMessage struct {
	CorrelationID string          `json:"correlationId,omitempty"`
	MessageType   string          `json:"messageType,omitempty"`
	Origin        int64           `json:"origin,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

// EdgeX 事件结构
type EdgeXEvent struct {
	ID          string         `json:"id"`
	DeviceName  string         `json:"deviceName"`
	Readings    []EdgeXReading `json:"readings"`
	Origin      int64          `json:"origin"`
	ProfileName string         `json:"profileName,omitempty"`
	SourceName  string         `json:"sourceName,omitempty"`
}

// EdgeX 读数结构
type EdgeXReading struct {
	ID           string          `json:"id"`
	ResourceName string          `json:"resourceName"`
	Value        string          `json:"value"`
	ValueType    string          `json:"valueType,omitempty"`
	Origin       int64           `json:"origin"`
	ProfileName  string          `json:"profileName,omitempty"`
	DeviceName   string          `json:"deviceName,omitempty"`
	BaseType     string          `json:"baseType,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

var appConfig *config.Config
var dataQueue *queue.Queue

func main() {
	// 加载配置
	var err error
	appConfig, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接 sfsDb
	if err := database.Init(appConfig.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 函数初始化 MQTT 客户端，连接到 EdgeX 的 MQTT broker
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

	// 初始化数据队列
	dataQueue, err = queue.NewQueue("./data_queue")
	if err != nil {
		log.Fatalf("Failed to initialize data queue: %v", err)
	}

	// 启动队列处理 goroutine，处理可能存在添加失败的数据
	/*
	   启动队列处理 goroutine，用于处理数据队列中的数据
	   1. 从队列中取出数据
	   2. 转换为 []*map[string]any 类型
	   3. 使用重试机制批量插入到数据库中（默认重试3次，每次间隔2秒）
	   4. 如果插入失败，将数据重新加入队列，等待后续重试
	*/
	dataQueue.ProcessQueue(func(data interface{}) error {
		records, ok := data.([]*map[string]any)
		if !ok {
			return fmt.Errorf("invalid data type in queue")
		}

		// 使用重试机制插入数据
		return database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
	})

	// 启动 HTTP 服务器
	serverInstance := server.NewServer(database.Table, appConfig)
	if err := serverInstance.Start(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)                      // 创建一个信号通道，用于接收中断信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 注册信号通知，当收到SIGINT或SIGTERM时，将信号发送到quit通道
	<-quit                                               // 阻塞直到收到中断信号。
	log.Println("Shutting down adapter...")
	// 给服务器 5 秒的时间来完成正在处理的请求
	time.Sleep(5 * time.Second)

	log.Println("Adapter exited")
}

// 初始化 MQTT 客户端
// 异步处理 ：MQTT 客户端在后台运行，当收到消息时会自动调用 messageHandler() 函数
func initMQTT() (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(appConfig.MQTTBroker)
	opts.SetClientID(appConfig.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Second * 30)
	opts.SetDefaultPublishHandler(messageHandler())

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	log.Printf("Connected to MQTT broker: %s", appConfig.MQTTBroker)
	return client, nil
}

// 订阅 EdgeX 消息
func subscribeToEdgeX(client mqtt.Client) error {
	token := client.Subscribe(appConfig.MQTTTopic, 1, nil)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %v", appConfig.MQTTTopic, token.Error())
	}

	log.Printf("Subscribed to topic: %s", appConfig.MQTTTopic)
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

		// 检查 MessageType 是否为 "event"
		if edgexMsg.MessageType != "event" && edgexMsg.MessageType != "Event" {
			log.Printf("Ignoring message with type: %s", edgexMsg.MessageType)
			return
		}

		// 解析 payload 中的事件
		var event EdgeXEvent
		if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
			log.Printf("Failed to parse event: %v", err)
			return
		}

		// 收集所有读数，准备批量插入
		var records []*map[string]any

		// 处理每个读数
		for _, reading := range event.Readings {
			// 准备数据
			metadataStr := ""
			if reading.Metadata != nil {
				metadataStr = string(reading.Metadata)
			}

			// 解析值的类型
			value := common.ParseValue(reading.Value)

			data := map[string]any{
				"id":         reading.ID,
				"deviceName": event.DeviceName,
				"reading":    reading.ResourceName,
				"value":      value,
				"valueType":  reading.ValueType,
				"baseType":   reading.BaseType,
				"timestamp":  reading.Origin, // 纳秒级时间戳，类型为 int64
				"metadata":   metadataStr,
			}

			records = append(records, &data)
		}
		/*
		   ### 具体流程
		   1. 收到 EdgeX 消息后，解析并准备数据
		   2. 尝试使用 BatchInsertWithRetry 函数将数据批量插入到数据库中（默认重试3次）
		   3. 如果所有重试都失败，将数据加入队列
		   4. 后台 goroutine 定期检查队列，尝试处理队列中的数据
		   5. 当数据库恢复可用时，队列中的数据会被成功写入
		*/
		// 批量存储到 sfsDb
		if len(records) > 0 {
			// 使用重试机制插入数据
			err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
			if err != nil {
				log.Printf("Failed to batch store data after retries: %v", err)
				// 将数据加入队列，以便后续处理
				if err := dataQueue.Enqueue(records); err != nil {
					log.Printf("Failed to enqueue data: %v", err)
				} else {
					log.Printf("Enqueued %d readings for later processing", len(records))
				}
			} else {
				log.Printf("Batch stored %d readings from %s", len(records), event.DeviceName)
			}
		}
	}
}
