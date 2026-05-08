package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sfsEdgeStore/config"
	"sfsEdgeStore/configwizard"
)

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   currentConfig,
	})
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var newConfig config.Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	configManager := config.GetConfigManager()
	if err := configManager.UpdateConfig(&newConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Config updated successfully",
	})
}

func (s *Server) handleReloadConfig(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	newConfig, err := config.ReloadFromFile()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	configManager := config.GetConfigManager()
	if err := configManager.UpdateConfig(newConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Config reloaded successfully",
	})
}

func (s *Server) handleOneClickConfig(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	wizard := configwizard.NewWizard(s.Config)
	if err := wizard.Run(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	configManager := config.GetConfigManager()

	smartConfig := &config.Config{
		DBPath:     "./data",
		MQTTBroker: "tcp://localhost:1883",
		MQTTTopic:  "edgex/events/#",
		ClientID:   config.GenerateClientID(),
		HTTPPort:   "8081",
		MQTTUseTLS:               false,
		HTTPUseTLS:               false,
		DBUseEncryption:          false,
		EnableAnalyzer:           false,
		EnableRetentionPolicy:    false,
		EnableAlertNotifications: false,
		EnableResourceMonitoring: true,
		DBScenario:               config.ScenarioExtreme,
	}

	if err := configManager.UpdateConfig(smartConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "One-click configuration completed successfully",
		"config":  smartConfig,
	})
}

func (s *Server) handleMQTTConfigUpdate(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var mqttConfig struct {
		MqttBroker   string `json:"mqtt_broker"`
		MqttTopic    string `json:"mqtt_topic"`
		ClientID     string `json:"client_id"`
		MqttUseTls   bool   `json:"mqtt_use_tls"`
		MqttUsername string `json:"mqtt_username"`
		MqttPassword string `json:"mqtt_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&mqttConfig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()
	if currentConfig == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Config not initialized"})
		return
	}

	currentConfig.MQTTBroker = mqttConfig.MqttBroker
	currentConfig.MQTTTopic = mqttConfig.MqttTopic
	currentConfig.ClientID = mqttConfig.ClientID
	currentConfig.MQTTUseTLS = mqttConfig.MqttUseTls
	if mqttConfig.MqttUsername != "" {
		currentConfig.MQTTUsername = mqttConfig.MqttUsername
	}
	if mqttConfig.MqttPassword != "" {
		currentConfig.MQTTPassword = mqttConfig.MqttPassword
	}

	if err := configManager.UpdateConfig(currentConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "MQTT config updated successfully",
	})
}

func (s *Server) handleSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	status := map[string]interface{}{
		"broker":         currentConfig.MQTTBroker,
		"topic":          currentConfig.MQTTTopic,
		"client_id":      currentConfig.ClientID,
		"use_tls":        currentConfig.MQTTUseTLS,
		"username":       currentConfig.MQTTUsername,
		"auto_subscribe": true,
		"edgex_version":  "latest",
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   status,
	})
}

func (s *Server) handleSubscriptionTest(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		Topic string `json:"topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Topic == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Topic is required"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Topic test for %s would be performed here", req.Topic),
	})
}

func (s *Server) handleSubscriptionThemes(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.handleGetSubscriptionThemes(w, r)
	case http.MethodPost:
		s.handleAddSubscriptionTheme(w, r)
	case http.MethodDelete:
		s.handleDeleteSubscriptionTheme(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func (s *Server) handleGetSubscriptionThemes(w http.ResponseWriter, r *http.Request) {
	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	themes := []string{}
	if currentConfig.MQTTTopic != "" {
		themes = append(themes, currentConfig.MQTTTopic)
	}

	customThemes := []map[string]interface{}{}
	for _, topic := range currentConfig.CustomTopics {
		customThemes = append(customThemes, map[string]interface{}{
			"topic":    topic,
			"added_at": time.Now().Format("2006-01-02 15:04:05"),
			"active":   true,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"themes":        themes,
		"custom_topics": customThemes,
	})
}

func (s *Server) handleAddSubscriptionTheme(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic string `json:"topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Topic == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Topic is required"})
		return
	}

	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	for _, topic := range currentConfig.CustomTopics {
		if topic == req.Topic {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "主题已存在"})
			return
		}
	}

	currentConfig.CustomTopics = append(currentConfig.CustomTopics, req.Topic)

	if err := config.SaveToFile(currentConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "保存配置失败"})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "主题添加成功",
		"topic":   req.Topic,
	})
}

func (s *Server) handleDeleteSubscriptionTheme(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic string `json:"topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Topic == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Topic is required"})
		return
	}

	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	newCustomTopics := []string{}
	found := false
	for _, topic := range currentConfig.CustomTopics {
		if topic != req.Topic {
			newCustomTopics = append(newCustomTopics, topic)
		} else {
			found = true
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "主题不存在"})
		return
	}

	currentConfig.CustomTopics = newCustomTopics

	if err := config.SaveToFile(currentConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "保存配置失败"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "主题删除成功",
		"topic":   req.Topic,
	})
}
