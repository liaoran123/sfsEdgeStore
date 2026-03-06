package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sfsEdgeStore/common"
	"sfsEdgeStore/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Notifier struct {
	config      *config.Config
	mqttClient  mqtt.Client
	alertChan   chan common.Alert
	stopChan    chan struct{}
	isRunning   bool
}

type AlertNotification struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{
		config:    cfg,
		alertChan: make(chan common.Alert, 100),
		stopChan:  make(chan struct{}),
	}
}

func (n *Notifier) Start() error {
	if !n.config.EnableAlertNotifications {
		log.Println("Alert notifications are disabled")
		return nil
	}

	if n.isRunning {
		log.Println("Alert notifier is already running")
		return nil
	}

	if len(n.config.AlertNotificationChannels) == 0 {
		log.Println("No alert notification channels configured")
		return nil
	}

	for _, channel := range n.config.AlertNotificationChannels {
		if channel == "mqtt" {
			if err := n.initMQTTClient(); err != nil {
				log.Printf("Failed to initialize MQTT client: %v", err)
			}
		}
	}

	n.isRunning = true
	go n.notificationLoop()
	log.Printf("Alert notifier started with channels: %v", n.config.AlertNotificationChannels)
	return nil
}

func (n *Notifier) Stop() {
	if !n.isRunning {
		return
	}
	close(n.stopChan)
	if n.mqttClient != nil && n.mqttClient.IsConnected() {
		n.mqttClient.Disconnect(250)
	}
	n.isRunning = false
	log.Println("Alert notifier stopped")
}

func (n *Notifier) SendAlert(alert common.Alert) {
	if !n.config.EnableAlertNotifications || !n.isRunning {
		return
	}

	if !n.shouldSendAlert(alert) {
		return
	}

	select {
	case n.alertChan <- alert:
	default:
		log.Println("Alert channel is full, dropping alert")
	}
}

func (n *Notifier) shouldSendAlert(alert common.Alert) bool {
	severityOrder := map[string]int{
		"info":     0,
		"warning":  1,
		"critical": 2,
	}

	minSeverity := severityOrder[n.config.AlertMinSeverity]
	alertSeverity := severityOrder[alert.Severity]

	return alertSeverity >= minSeverity
}

func (n *Notifier) notificationLoop() {
	for {
		select {
		case alert := <-n.alertChan:
			n.sendToAllChannels(alert)
		case <-n.stopChan:
			return
		}
	}
}

func (n *Notifier) sendToAllChannels(alert common.Alert) {
	notification := AlertNotification{
		Type:      alert.Type,
		Message:   alert.Message,
		Severity:  alert.Severity,
		Timestamp: alert.Timestamp,
		Source:    n.config.ClientID,
	}

	for _, channel := range n.config.AlertNotificationChannels {
		switch channel {
		case "mqtt":
			n.sendToMQTT(notification)
		case "webhook":
			n.sendToWebhook(notification)
		case "log":
			n.sendToLog(notification)
		}
	}
}

func (n *Notifier) initMQTTClient() error {
	if n.config.MQTTBroker == "" {
		return fmt.Errorf("MQTT broker not configured")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(n.config.MQTTBroker)
	opts.SetClientID(n.config.ClientID + "-alert-notifier")
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Alert notifier connected to MQTT broker")
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Alert notifier MQTT connection lost: %v", err)
	})

	n.mqttClient = mqtt.NewClient(opts)
	token := n.mqttClient.Connect()
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (n *Notifier) sendToMQTT(notification AlertNotification) {
	if n.mqttClient == nil || !n.mqttClient.IsConnected() {
		log.Println("MQTT client not connected, cannot send alert")
		return
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal alert for MQTT: %v", err)
		return
	}

	topic := n.config.AlertMQTTTopic
	if topic == "" {
		topic = "edgex/alerts"
	}

	token := n.mqttClient.Publish(topic, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to send alert to MQTT: %v", token.Error())
	}
}

func (n *Notifier) sendToWebhook(notification AlertNotification) {
	if n.config.AlertWebhookURL == "" {
		return
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal alert for webhook: %v", err)
		return
	}

	resp, err := http.Post(n.config.AlertWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to send alert to webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Webhook returned non-success status: %d", resp.StatusCode)
	}
}

func (n *Notifier) sendToLog(notification AlertNotification) {
	log.Printf("[ALERT] [%s] [%s] %s (from: %s, at: %v)",
		notification.Severity,
		notification.Type,
		notification.Message,
		notification.Source,
		notification.Timestamp)
}

func (n *Notifier) GetNotifierStatus() map[string]any {
	return map[string]any{
		"enabled":            n.config.EnableAlertNotifications,
		"channels":           n.config.AlertNotificationChannels,
		"min_severity":       n.config.AlertMinSeverity,
		"mqtt_topic":         n.config.AlertMQTTTopic,
		"webhook_url_configured": n.config.AlertWebhookURL != "",
		"is_running":         n.isRunning,
	}
}
