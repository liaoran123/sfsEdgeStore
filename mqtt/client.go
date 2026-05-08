package mqtt

import (
	"log"

	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/config"
	"sfsEdgeStore/edgex"
	"sfsEdgeStore/monitor"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// standardTopics 标准主题列表
var standardTopics = []string{
	"edgex/events/#",
	"devices/+/data",
	"edgex/events/core/#",
}

// Client 统一客户端 - 整合 MQTTClient、BatchWriter、MessageProcessor
type Client struct {
	mqttClient       *MQTTClient
	batchWriter      *BatchWriter
	messageProcessor *MessageProcessor
}

// NewClient 创建统一客户端
func NewClient(cfg *config.Config, monitor *monitor.Monitor, analyzer *analyzer.Analyzer) (*Client, error) {

	// 1. 创建各组件
	batchWriter, err := NewBatchWriter(monitor, analyzer)
	if err != nil {
		return nil, err
	}
	messageProcessor := NewMessageProcessor()

	// 2. 创建 Client
	c := &Client{
		batchWriter:      batchWriter,
		messageProcessor: messageProcessor,
	}

	// 3. 创建 MQTTClient
	mqttClient, err := NewMQTTClient(cfg, c.handleMessage, monitor)
	if err != nil {
		return nil, err
	}
	c.mqttClient = mqttClient

	return c, nil
}

// BroadcastChan 返回广播通道，由 Server 直接监听
func (c *Client) BroadcastChan() <-chan *BroadcastMessage {
	return c.batchWriter.GetBroadcastChan()
}

// handleMessage 消息处理管道，接收所有 MQTT 消息，解析并处理事件
// client 参数没有用到。这是 mqtt.MessageHandler 类型签名的要求，必须保留但可以改为 _ 表示忽略
func (c *Client) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	// 1. 解析 MQTT 消息
	event, err := edgex.ProcessMessage(msg.Payload())
	if err != nil {
		log.Printf("Failed to process message: %v", err)
		return
	}
	// 2. 处理事件，收集记录数据
	records := c.messageProcessor.ProcessEvent(event)
	if len(records) > 0 {
		// 3. 写入记录数据。所有数据都在这里处理。
		c.batchWriter.Add(records)
	}
}

// Subscribe 订阅主题
func (c *Client) Subscribe(topics []string) error {
	return c.mqttClient.Subscribe(topics)
}

// Disconnect 断开连接
func (c *Client) Disconnect() {
	c.batchWriter.Stop()
	c.mqttClient.Disconnect()
	log.Println("MQTT Client shutdown complete")
}

// IsConnected 检查是否在线
func (c *Client) IsConnected() bool {
	return c.mqttClient.IsConnected()
}

// GetTopics 获取要订阅的主题列表
func GetTopics(cfg *config.Config) []string {
	topics := []string{}
	if cfg.MQTTTopic != "" { // 自定义主题，cfg.MQTTTopic保存的是用户自定义的主题列表
		topics = append(topics, cfg.MQTTTopic)
	}
	topics = append(topics, standardTopics...) // 始终订阅标准 EdgeX 主题
	return uniqueTopics(topics)
}

// uniqueTopics 去重主题
func uniqueTopics(topics []string) []string {
	topicMap := make(map[string]bool)
	for _, topic := range topics {
		topicMap[topic] = true
	}
	unique := []string{}
	for topic := range topicMap {
		unique = append(unique, topic)
	}
	return unique
}
