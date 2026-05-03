package mqtt

import (
	"encoding/json"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/*
MQTT Broker 连接地址 | tcp://localhost:1883 | 连接用的 URL |
MQTT 主题 | edgex/events/status | 发布/订阅用的主题 |
*/

// 发布上线，离线状态消息到MQTT broker
const statusTopic = "edgex/events/status"

// createStatusPayload 创建状态消息的 JSON payload
func createStatusPayload(clientID string, status string) ([]byte, error) {
	message := map[string]any{
		"status":    status,
		"clientId":  clientID,
		"timestamp": time.Now().UnixNano(),
	}
	return json.Marshal(message)
}

// publishStatusMessage 发布状态消息到 MQTT broker
func publishStatusMessage(client mqtt.Client, clientID string, status string) error {
	// 构建状态消息的 JSON payload
	payload, err := createStatusPayload(clientID, status)
	if err != nil {
		return err
	}

	// 发布状态消息到 MQTT broker
	token := client.Publish(statusTopic, 1, false, payload)
	token.Wait()
	return token.Error()
}
