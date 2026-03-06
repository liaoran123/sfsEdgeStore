package main

import (
	"os"
	"testing"

	"sfsEdgeStore/config"
)

// TestConfigFromEnvironment 测试从环境变量加载配置
func TestConfigFromEnvironment(t *testing.T) {
	// 设置环境变量
	testDBPath := "./test_edgex_data"
	testMQTTBroker := "tcp://test-broker:1883"
	testMQTTTopic := "test/topic"
	testClientID := "test-client"

	// 保存原始环境变量
	originalDBPath := os.Getenv("EDGEX_DB_PATH")
	originalMQTTBroker := os.Getenv("EDGEX_MQTT_BROKER")
	originalMQTTTopic := os.Getenv("EDGEX_MQTT_TOPIC")
	originalClientID := os.Getenv("EDGEX_CLIENT_ID")

	// 设置测试环境变量
	os.Setenv("EDGEX_DB_PATH", testDBPath)
	os.Setenv("EDGEX_MQTT_BROKER", testMQTTBroker)
	os.Setenv("EDGEX_MQTT_TOPIC", testMQTTTopic)
	os.Setenv("EDGEX_CLIENT_ID", testClientID)

	// 加载配置
	var err error
	appConfig, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置
	if appConfig.DBPath != testDBPath {
		t.Errorf("Expected DBPath %s, got %s", testDBPath, appConfig.DBPath)
	}

	if appConfig.MQTTBroker != testMQTTBroker {
		t.Errorf("Expected MQTTBroker %s, got %s", testMQTTBroker, appConfig.MQTTBroker)
	}

	if appConfig.MQTTTopic != testMQTTTopic {
		t.Errorf("Expected MQTTTopic %s, got %s", testMQTTTopic, appConfig.MQTTTopic)
	}

	if appConfig.ClientID != testClientID {
		t.Errorf("Expected ClientID %s, got %s", testClientID, appConfig.ClientID)
	}

	// 恢复原始环境变量
	os.Setenv("EDGEX_DB_PATH", originalDBPath)
	os.Setenv("EDGEX_MQTT_BROKER", originalMQTTBroker)
	os.Setenv("EDGEX_MQTT_TOPIC", originalMQTTTopic)
	os.Setenv("EDGEX_CLIENT_ID", originalClientID)

	t.Log("Config from environment test passed")
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	// 清除环境变量
	os.Unsetenv("EDGEX_DB_PATH")
	os.Unsetenv("EDGEX_MQTT_BROKER")
	os.Unsetenv("EDGEX_MQTT_TOPIC")
	os.Unsetenv("EDGEX_CLIENT_ID")

	// 加载配置
	var err error
	appConfig, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证默认配置
	if appConfig.DBPath != "./edgex_data" {
		t.Errorf("Expected default DBPath ./edgex_data, got %s", appConfig.DBPath)
	}

	if appConfig.MQTTBroker != "tcp://localhost:1883" {
		t.Errorf("Expected default MQTTBroker tcp://localhost:1883, got %s", appConfig.MQTTBroker)
	}

	if appConfig.MQTTTopic != "edgex/events/core/#" {
		t.Errorf("Expected default MQTTTopic edgex/events/core/#, got %s", appConfig.MQTTTopic)
	}

	// ClientID 是动态生成的，只需要验证不为空
	if appConfig.ClientID == "" {
		t.Error("Expected non-empty ClientID, got empty")
	}

	t.Log("Default config test passed")
}
