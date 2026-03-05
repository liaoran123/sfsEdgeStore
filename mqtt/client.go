package mqtt

import (
	"fmt"
	"log"
	"time"

	"sfsdb-edgex-adapter/common"
	"sfsdb-edgex-adapter/config"
	"sfsdb-edgex-adapter/database"
	"sfsdb-edgex-adapter/edgex"
	"sfsdb-edgex-adapter/queue"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client MQTT客户端结构体
type Client struct {
	client    mqtt.Client
	config    *config.Config
	dataQueue *queue.Queue
}

// NewClient 创建新的MQTT客户端
func NewClient(cfg *config.Config, dataQueue *queue.Queue) (*Client, error) {
	client, err := initMQTT(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    client,
		config:    cfg,
		dataQueue: dataQueue,
	}, nil
}

// initMQTT 初始化MQTT客户端
func initMQTT(cfg *config.Config) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker) // 连接到EdgeX的MQTT broker
	opts.SetClientID(cfg.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Second * 30)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	log.Printf("Connected to MQTT broker: %s", cfg.MQTTBroker)
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
	c.client.Disconnect(250)
}

// Publish 发布消息到MQTT主题
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := c.client.Publish(topic, qos, retained, payload)
	token.Wait()
	return token.Error()
}

// messageHandler 适配器处理收到的EdgeX消息
func (c *Client) messageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message on topic: %s", msg.Topic())

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
				"deviceName": event.DeviceName, // 设备名称已经在ProcessMessage中格式化
				"reading":    reading.ResourceName,
				"value":      value,
				"valueType":  reading.ValueType,
				"baseType":   reading.BaseType,
				"timestamp":  reading.Origin, // 纳秒级时间戳，类型为 int64
				"metadata":   metadataStr,
			}

			records = append(records, &data)
		}

		// 批量存储到 sfsDb
		if len(records) > 0 {
			// 使用重试机制插入数据
			err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
			if err != nil {
				log.Printf("Failed to batch store data after retries: %v", err)
				// 将数据加入队列，以便后续处理
				if err := c.dataQueue.Enqueue(records); err != nil {
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
