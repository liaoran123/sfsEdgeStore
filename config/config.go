package config

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb/opt"
)

// 数据库场景常量
const (
	ScenarioEmbedded = "embedded"
	ScenarioIoT      = "iot"
	ScenarioEdge     = "edge"
	ScenarioGame     = "game"
	ScenarioDefault  = "default"
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
	DBUseEncryption       bool   `json:"db_use_encryption" env:"EDGEX_DB_USE_ENCRYPTION"`
	DBEncryptionKey       string `json:"db_encryption_key" env:"EDGEX_DB_ENCRYPTION_KEY"`
	DBEncryptionAlgorithm string `json:"db_encryption_algorithm" env:"EDGEX_DB_ENCRYPTION_ALGORITHM"`
	// 分析引擎配置
	EnableAnalyzer        bool                       `json:"enable_analyzer" env:"EDGEX_ENABLE_ANALYZER"`
	AnalyzerMaxMemory     int                        `json:"analyzer_max_memory" env:"EDGEX_ANALYZER_MAX_MEMORY"`
	AnalyzerMaxTimePerRun int                        `json:"analyzer_max_time_per_run" env:"EDGEX_ANALYZER_MAX_TIME_PER_RUN"`
	AnalyzerThresholds    map[string]ThresholdConfig `json:"analyzer_thresholds"`
	// 数据保留策略配置
	EnableRetentionPolicy bool `json:"enable_retention_policy" env:"EDGEX_ENABLE_RETENTION_POLICY"`
	RetentionDays         int  `json:"retention_days" env:"EDGEX_RETENTION_DAYS"`
	CleanupInterval       int  `json:"cleanup_interval_hours" env:"EDGEX_CLEANUP_INTERVAL_HOURS"`
	CleanupBatchSize      int  `json:"cleanup_batch_size" env:"EDGEX_CLEANUP_BATCH_SIZE"`
	// 告警通知配置
	EnableAlertNotifications  bool     `json:"enable_alert_notifications" env:"EDGEX_ENABLE_ALERT_NOTIFICATIONS"`
	AlertNotificationChannels []string `json:"alert_notification_channels" env:"EDGEX_ALERT_NOTIFICATION_CHANNELS"`
	AlertMQTTTopic            string   `json:"alert_mqtt_topic" env:"EDGEX_ALERT_MQTT_TOPIC"`
	AlertWebhookURL           string   `json:"alert_webhook_url" env:"EDGEX_ALERT_WEBHOOK_URL"`
	AlertMinSeverity          string   `json:"alert_min_severity" env:"EDGEX_ALERT_MIN_SEVERITY"`
	// 数据同步配置
	EnableDataSync        bool   `json:"enable_data_sync" env:"EDGEX_ENABLE_DATA_SYNC"`
	DataSyncMQTTTopic     string `json:"data_sync_mqtt_topic" env:"EDGEX_DATA_SYNC_MQTT_TOPIC"`
	DataSyncQueueDir      string `json:"data_sync_queue_dir" env:"EDGEX_DATA_SYNC_QUEUE_DIR"`
	DataSyncBatchSize     int    `json:"data_sync_batch_size" env:"EDGEX_DATA_SYNC_BATCH_SIZE"`
	DataSyncInterval      int    `json:"data_sync_interval_seconds" env:"EDGEX_DATA_SYNC_INTERVAL_SECONDS"`
	DataSyncMaxRetryCount int    `json:"data_sync_max_retry_count" env:"EDGEX_DATA_SYNC_MAX_RETRY_COUNT"`
	// 资源使用监控配置
	EnableResourceMonitoring bool    `json:"enable_resource_monitoring" env:"EDGEX_ENABLE_RESOURCE_MONITORING"`
	MaxMemoryMB              float64 `json:"max_memory_mb" env:"EDGEX_MAX_MEMORY_MB"`
	MaxCPUPercent            float64 `json:"max_cpu_percent" env:"EDGEX_MAX_CPU_PERCENT"`
	ResourceMonitorInterval  int     `json:"resource_monitor_interval_seconds" env:"EDGEX_RESOURCE_MONITOR_INTERVAL_SECONDS"`
	// 数据库场景配置
	DBScenario string `json:"db_scenario" env:"EDGEX_DB_SCENARIO"`
	// Prometheus 指标配置（可选，默认关闭）
	EnablePrometheus bool   `json:"enable_prometheus" env:"EDGEX_ENABLE_PROMETHEUS"`
	PrometheusPath   string `json:"prometheus_path" env:"EDGEX_PROMETHEUS_PATH"`
	// 模拟器配置
	EnableSimulator       bool          `json:"enable_simulator" env:"EDGEX_ENABLE_SIMULATOR"`
	SimulatorIntervalMin  int           `json:"simulator_interval_min" env:"EDGEX_SIMULATOR_INTERVAL_MIN"`
	SimulatorIntervalMax  int           `json:"simulator_interval_max" env:"EDGEX_SIMULATOR_INTERVAL_MAX"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// ConfigUpdateHandler 配置更新回调函数类型
type ConfigUpdateHandler func(oldCfg, newCfg *Config)

// ConfigManager 配置管理器
type ConfigManager struct {
	currentConfig  *Config
	mutex          sync.RWMutex
	updateHandlers []ConfigUpdateHandler
}

var configManager *ConfigManager
var initOnce sync.Once

// GetConfigManager 获取配置管理器单例
func GetConfigManager() *ConfigManager {
	initOnce.Do(func() {
		configManager = &ConfigManager{
			updateHandlers: make([]ConfigUpdateHandler, 0),
		}
	})
	return configManager
}

// SetConfig 设置当前配置
func (cm *ConfigManager) SetConfig(cfg *Config) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.currentConfig = cfg
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.currentConfig
}

// RegisterUpdateHandler 注册配置更新回调
func (cm *ConfigManager) RegisterUpdateHandler(handler ConfigUpdateHandler) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.updateHandlers = append(cm.updateHandlers, handler)
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(newCfg *Config) error {
	cm.mutex.Lock()
	oldCfg := cm.currentConfig
	cm.currentConfig = newCfg
	handlers := make([]ConfigUpdateHandler, len(cm.updateHandlers))
	copy(handlers, cm.updateHandlers)
	cm.mutex.Unlock()

	// 保存配置到文件
	if err := SaveToFile(newCfg); err != nil {
		log.Printf("Failed to save config to file: %v", err)
	}

	// 通知所有更新处理器
	for _, handler := range handlers {
		handler(oldCfg, newCfg)
	}

	log.Println("Config updated successfully")
	return nil
}

// GetScenarioOptions 根据场景获取数据库配置选项
func (cm *ConfigManager) GetScenarioOptions() *opt.Options {
	cfg := cm.GetConfig()
	switch cfg.DBScenario {
	case ScenarioEmbedded:
		return &opt.Options{
			WriteBuffer:            2 * 1024 * 1024,
			OpenFilesCacheCapacity: 5,
			BlockCacheCapacity:     4 * 1024 * 1024,
			Compression:            opt.DefaultCompression,
		}
	case ScenarioIoT:
		return &opt.Options{
			WriteBuffer:            4 * 1024 * 1024,
			OpenFilesCacheCapacity: 10,
			BlockCacheCapacity:     8 * 1024 * 1024,
			Compression:            opt.DefaultCompression,
		}
	case ScenarioEdge:
		return &opt.Options{
			WriteBuffer:            16 * 1024 * 1024,
			OpenFilesCacheCapacity: 50,
			BlockCacheCapacity:     32 * 1024 * 1024,
			Compression:            opt.DefaultCompression,
		}
	case ScenarioGame:
		return &opt.Options{
			WriteBuffer:            64 * 1024 * 1024,
			OpenFilesCacheCapacity: 200,
			BlockCacheCapacity:     128 * 1024 * 1024,
			Compression:            opt.NoCompression,
		}
	default:
		return &opt.Options{
			WriteBuffer:            64 * 1024 * 1024,
			OpenFilesCacheCapacity: 200,
			BlockCacheCapacity:     128 * 1024 * 1024,
			Compression:            opt.DefaultCompression,
		}
	}
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
		DBUseEncryption:       false,
		DBEncryptionAlgorithm: "AES-256-GCM",
		// 分析引擎默认值
		EnableAnalyzer:        false,
		AnalyzerMaxMemory:     10485760, // 10MB
		AnalyzerMaxTimePerRun: 500,      // 500ms
		AnalyzerThresholds:    make(map[string]ThresholdConfig),
		// 数据保留策略默认值
		EnableRetentionPolicy: false,
		RetentionDays:         30,   // 默认保留30天数据
		CleanupInterval:       24,   // 默认每24小时清理一次
		CleanupBatchSize:      1000, // 默认每批清理1000条记录
		// 告警通知默认值
		EnableAlertNotifications:  false,
		AlertNotificationChannels: []string{},
		AlertMQTTTopic:            "edgex/alerts",
		AlertWebhookURL:           "",
		AlertMinSeverity:          "warning",
		// 数据同步默认值
		EnableDataSync:        false,
		DataSyncMQTTTopic:     "edgex/data/sync",
		DataSyncQueueDir:      "./data_sync_queue",
		DataSyncBatchSize:     100,
		DataSyncInterval:      30,
		DataSyncMaxRetryCount: 5,
		// 资源使用监控默认值
		EnableResourceMonitoring: true,
		MaxMemoryMB:              50, // 50MB 内存限制
		MaxCPUPercent:            5,  // 5% CPU 限制
		ResourceMonitorInterval:  10, // 每10秒检查一次
		// 数据库场景默认值
		DBScenario: ScenarioEdge, // 默认使用边缘场景
		// Prometheus 指标默认值（默认关闭，避免性能影响）
		EnablePrometheus: false,
		PrometheusPath:   "/metrics",
		// 模拟器默认配置（默认关闭）
		EnableSimulator:      false,
		SimulatorIntervalMin: 2,
		SimulatorIntervalMax: 5,
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

	// 设置到配置管理器
	GetConfigManager().SetConfig(cfg)

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
