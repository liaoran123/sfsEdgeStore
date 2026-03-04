package config

import (
	"log"
	"os"
	"time"
)

/*
要与实际 EdgeX 系统通信，需要配置：
- MQTT broker 地址（ EDGEX_MQTT_BROKER 环境变量）
- 订阅主题（ EDGEX_MQTT_TOPIC 环境变量）
- 客户端 ID（ EDGEX_CLIENT_ID 环境变量）
*/
// Config 配置结构体
type Config struct {
	DBPath     string `json:"db_path" env:"EDGEX_DB_PATH"`
	MQTTBroker string `json:"mqtt_broker" env:"EDGEX_MQTT_BROKER"`
	MQTTTopic  string `json:"mqtt_topic" env:"EDGEX_MQTT_TOPIC"`
	ClientID   string `json:"client_id" env:"EDGEX_CLIENT_ID"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 1. 设置默认配置
	cfg := &Config{
		DBPath:     "./edgex_data",
		MQTTBroker: "tcp://localhost:1883",
		MQTTTopic:  "edgex/events/core/#",
		ClientID:   generateClientID(),
	}

	// 2. 尝试从EdgeX配置中心加载
	if err := loadFromConfigCenter(cfg); err != nil {
		log.Printf("Failed to load config from EdgeX config center: %v", err)
		log.Println("Falling back to local config file")

		// 3. 从配置文件加载
		if err := loadFromFile(cfg); err != nil {
			log.Printf("Failed to load config from file: %v", err)
			log.Println("Using default config")
		}
	}

	// 4. 从环境变量加载（优先级最高）
	loadFromEnv(cfg)

	return cfg, nil
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return "sfsdb-edgex-adapter-" + time.Now().Format("20060102150405")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
