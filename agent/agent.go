package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"sfsEdgeStore/config"
	"sfsEdgeStore/monitor"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Agent 极简管理Agent
type Agent struct {
	client         mqtt.Client
	config         *config.Config
	monitor        *monitor.Monitor
	deviceID       string
	heartbeatTopic string
	commandTopic   string
	responseTopic  string
}

// Command 指令结构
type Command struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp string          `json:"timestamp"`
}

// Response 响应结构
type Response struct {
	RequestID string          `json:"request_id"`
	Status    string          `json:"status"`
	Result    json.RawMessage `json:"result,omitempty"`
	Timestamp string          `json:"timestamp"`
}

// Heartbeat 心跳消息结构
type Heartbeat struct {
	DeviceID  string `json:"device_id"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// NewAgent 创建新的Agent实例
func NewAgent(cfg *config.Config, monitor *monitor.Monitor) (*Agent, error) {
	deviceID := cfg.ClientID
	if deviceID == "" {
		deviceID = "edgex-adapter-" + time.Now().Format("20060102150405")
	}

	heartbeatTopic := "edgex/agents/" + deviceID + "/heartbeat"
	commandTopic := "edgex/agents/" + deviceID + "/commands"
	responseTopic := "edgex/agents/" + deviceID + "/responses"

	// 初始化MQTT客户端
	client, err := initMQTT(cfg, commandTopic, heartbeatTopic, deviceID)
	if err != nil {
		return nil, err
	}

	return &Agent{
		client:         client,
		config:         cfg,
		monitor:        monitor,
		deviceID:       deviceID,
		heartbeatTopic: heartbeatTopic,
		commandTopic:   commandTopic,
		responseTopic:  responseTopic,
	}, nil
}

// initMQTT 初始化MQTT客户端
func initMQTT(cfg *config.Config, commandTopic, heartbeatTopic, deviceID string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.ClientID)
	opts.SetCleanSession(false)                   // 启用持久会话
	opts.SetAutoReconnect(true)                   // 启用自动重连
	opts.SetMaxReconnectInterval(time.Minute * 5) // 最大重连间隔5分钟

	// 设置连接处理函数
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("MQTT broker connected")
		// 重新订阅指令主题
		token := client.Subscribe(commandTopic, 1, nil) // 临时设置为nil，实际处理会在Start方法中设置
		token.Wait()
		if token.Error() != nil {
			log.Printf("Failed to resubscribe to command topic: %v", token.Error())
		} else {
			log.Printf("Resubscribed to command topic: %s", commandTopic)
			// 发送上线心跳
			heartbeat := Heartbeat{
				DeviceID:  deviceID,
				Status:    "online",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			payload, err := json.Marshal(heartbeat)
			if err != nil {
				log.Printf("Failed to marshal online heartbeat: %v", err)
				return
			}
			token := client.Publish(heartbeatTopic, 1, false, payload)
			token.Wait()
			if token.Error() != nil {
				log.Printf("Failed to send online heartbeat: %v", token.Error())
			} else {
				log.Printf("Sent online heartbeat for device: %s", deviceID)
			}
		}
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}

	log.Println("Connected to MQTT broker for agent")
	return client, nil
}

// Start 启动Agent
func (a *Agent) Start() error {
	// 启动心跳协程
	go a.sendHeartbeat()

	// 订阅指令主题
	token := a.client.Subscribe(a.commandTopic, 1, a.handleCommand)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to subscribe to command topic: %v", token.Error())
		return fmt.Errorf("failed to subscribe to command topic: %v", token.Error())
	}

	log.Printf("Agent started with device ID: %s", a.deviceID)
	log.Printf("Subscribed to command topic: %s", a.commandTopic)
	return nil
}

// sendHeartbeat 发送心跳
func (a *Agent) sendHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		heartbeat := Heartbeat{
			DeviceID:  a.deviceID,
			Status:    "online",
			Timestamp: time.Now().Format(time.RFC3339),
		}

		payload, err := json.Marshal(heartbeat)
		if err != nil {
			log.Printf("Failed to marshal heartbeat: %v", err)
			continue
		}

		token := a.client.Publish(a.heartbeatTopic, 1, false, payload)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Failed to send heartbeat: %v", err)
		}
	}
}

// sendOnlineHeartbeat 发送上线心跳
func (a *Agent) sendOnlineHeartbeat() {
	heartbeat := Heartbeat{
		DeviceID:  a.deviceID,
		Status:    "online",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(heartbeat)
	if err != nil {
		log.Printf("Failed to marshal online heartbeat: %v", err)
		return
	}

	token := a.client.Publish(a.heartbeatTopic, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to send online heartbeat: %v", token.Error())
	} else {
		log.Printf("Sent online heartbeat for device: %s", a.deviceID)
	}
}

// handleCommand 处理指令
func (a *Agent) handleCommand(client mqtt.Client, msg mqtt.Message) {
	var cmd Command
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Printf("Failed to parse command: %v", err)
		a.sendErrorResponse("", "Invalid command format")
		return
	}

	log.Printf("Received command: %s", cmd.Type)

	switch cmd.Type {
	case "REBOOT":
		a.handleReboot(cmd.RequestID)
	case "UPDATE_CONFIG":
		a.handleUpdateConfig(cmd.RequestID, cmd.Payload)
	case "GET_STATUS":
		a.handleGetStatus(cmd.RequestID)
	default:
		log.Printf("Unknown command: %s", cmd.Type)
		a.sendErrorResponse(cmd.RequestID, "Unknown command")
	}
}

// handleReboot 处理重启指令
func (a *Agent) handleReboot(requestID string) {
	log.Println("Executing reboot command")

	// 发送响应
	response := Response{
		RequestID: requestID,
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	token := a.client.Publish(a.responseTopic, 1, false, payload)
	token.Wait()

	// 重启服务
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Rebooting service...")
		// 这里可以根据实际部署环境实现重启逻辑
		// 例如，对于systemd服务，可以使用systemctl restart
		cmd := exec.Command("sh", "-c", "kill -HUP $(pgrep -f sfsdb-edgex-adapter)")
		cmd.Run()
	}()
}

// handleUpdateConfig 处理更新配置指令
func (a *Agent) handleUpdateConfig(requestID string, payload json.RawMessage) {
	log.Println("Executing update config command")

	var newConfig config.Config
	if err := json.Unmarshal(payload, &newConfig); err != nil {
		log.Printf("Failed to parse config payload: %v", err)
		a.sendErrorResponse(requestID, "Invalid config format")
		return
	}

	configManager := config.GetConfigManager()
	if err := configManager.UpdateConfig(&newConfig); err != nil {
		log.Printf("Failed to update config: %v", err)
		a.sendErrorResponse(requestID, fmt.Sprintf("Failed to update config: %v", err))
		return
	}

	response := Response{
		RequestID: requestID,
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respPayload, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	token := a.client.Publish(a.responseTopic, 1, false, respPayload)
	token.Wait()
}

// handleGetStatus 处理获取状态指令
func (a *Agent) handleGetStatus(requestID string) {
	log.Println("Executing get status command")

	// 收集状态信息
	status := map[string]interface{}{
		"device_id": a.deviceID,
		"status":    "online",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// 如果有监控实例，添加监控指标
	if a.monitor != nil {
		metrics := a.monitor.CollectMetrics()
		status["metrics"] = metrics
	}

	statusPayload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal status: %v", err)
		a.sendErrorResponse(requestID, "Failed to collect status")
		return
	}

	response := Response{
		RequestID: requestID,
		Status:    "success",
		Result:    statusPayload,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respPayload, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	token := a.client.Publish(a.responseTopic, 1, false, respPayload)
	token.Wait()
}

// sendErrorResponse 发送错误响应
func (a *Agent) sendErrorResponse(requestID, message string) {
	response := Response{
		RequestID: requestID,
		Status:    "error",
		Result:    json.RawMessage(`{"message": "` + message + `"}`),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal error response: %v", err)
		return
	}

	token := a.client.Publish(a.responseTopic, 1, false, payload)
	token.Wait()
}

// Stop 停止Agent
func (a *Agent) Stop() {
	if a.client != nil {
		a.client.Disconnect(250)
	}
	log.Println("Agent stopped")
}
