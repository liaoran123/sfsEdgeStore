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
	HTTPPort   string `json:"http_port" env:"EDGEX_HTTP_PORT"`
	// TLS 配置
	MQTTUseTLS     bool   `json:"mqtt_use_tls" env:"EDGEX_MQTT_USE_TLS"`
	MQTTCACert     string `json:"mqtt_ca_cert" env:"EDGEX_MQTT_CA_CERT"`
	MQTTClientCert string `json:"mqtt_client_cert" env:"EDGEX_MQTT_CLIENT_CERT"`
	MQTTClientKey  string `json:"mqtt_client_key" env:"EDGEX_MQTT_CLIENT_KEY"`
	HTTPUseTLS     bool   `json:"http_use_tls" env:"EDGEX_HTTP_USE_TLS"`
	HTTPCert       string `json:"http_cert" env:"EDGEX_HTTP_CERT"`
	HTTPKey        string `json:"http_key" env:"EDGEX_HTTP_KEY"`
	// 数据库加密配置
	DBUseEncryption bool   `json:"db_use_encryption" env:"EDGEX_DB_USE_ENCRYPTION"`
	DBEncryptionKey string `json:"db_encryption_key" env:"EDGEX_DB_ENCRYPTION_KEY"`
	DBEncryptionAlgorithm string `json:"db_encryption_algorithm" env:"EDGEX_DB_ENCRYPTION_ALGORITHM"`
	// 分析引擎配置
	EnableAnalyzer     bool              `json:"enable_analyzer" env:"EDGEX_ENABLE_ANALYZER"`
	AnalyzerMaxMemory  int               `json:"analyzer_max_memory" env:"EDGEX_ANALYZER_MAX_MEMORY"`
	AnalyzerMaxTimePerRun int            `json:"analyzer_max_time_per_run" env:"EDGEX_ANALYZER_MAX_TIME_PER_RUN"`
	AnalyzerThresholds  map[string]ThresholdConfig `json:"analyzer_thresholds"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 1. 设置默认配置
	cfg := &Config{
		DBPath:     "./edgex_data",
		MQTTBroker: "tcp://localhost:1883",
		MQTTTopic:  "edgex/events/core/#",
		ClientID:   generateClientID(),
		HTTPPort:   "8081", // 默认HTTP端口
		// TLS 默认值
		MQTTUseTLS: false,
		HTTPUseTLS: false,
		// 数据库加密默认值
		DBUseEncryption: false,
		DBEncryptionAlgorithm: "AES-256-GCM",
		// 分析引擎默认值
		EnableAnalyzer:     false,
		AnalyzerMaxMemory:  10485760, // 10MB
		AnalyzerMaxTimePerRun: 500,    // 500ms
		AnalyzerThresholds:  make(map[string]ThresholdConfig),
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
