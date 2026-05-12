package config

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"sfsEdgeStore/pathutil"
)

// 数据库场景常量（与 sfsDb/storage/options.go 保持一致）
/*
| 场景常量 | 内存占用 | 适合设备 | 说明 |
|----------|---------|---------|------|
| `ScenarioExtreme` | **6MB** | **≤128MB** | **极限生存模式，当前项目默认** |
| `ScenarioEmbedded` | 6MB | ≤64MB | 嵌入式场景，最轻量 |
| `ScenarioIoT` | 12MB | 64-256MB | 物联网场景，适合数据量大但内存受限 |
| `ScenarioEdge` | 48MB | 256-512MB | 边缘计算场景，适合有一定内存的设备 |
| `ScenarioGame` | 192MB | 不推荐 | 游戏场景，高吞吐低延迟 |
| `ScenarioDefault` | 192MB | 1GB+ | 默认场景，不限制内存占用 |
*/
const (
	ScenarioEmbedded = "embedded" // 嵌入式场景，最轻量
	ScenarioIoT      = "iot"      // 物联网场景
	ScenarioEdge     = "edge"     // 边缘场景
	ScenarioGame     = "game"     // 游戏场景
	ScenarioDefault  = "default"  // 默认场景
	ScenarioExtreme  = "extreme"  // 极限生存模式（128MB内存设备）
)

// Config 配置结构体 - 针对边缘计算资源受限设备优化
type Config struct {
	// 核心连接配置
	DBPath     string `json:"db_path"`
	MQTTBroker string `json:"mqtt_broker"`
	MQTTTopic  string `json:"mqtt_topic"`
	ClientID   string `json:"client_id"`
	HTTPPort   string `json:"http_port"`
	// MQTT 认证
	MQTTUsername string `json:"mqtt_username,omitempty"`
	MQTTPassword string `json:"mqtt_password,omitempty"`
	// TLS 配置
	MQTTUseTLS bool `json:"mqtt_use_tls"`
	HTTPUseTLS bool `json:"http_use_tls"`
	// 数据库加密配置
	DBUseEncryption bool   `json:"db_use_encryption"`
	DBEncryptionKey string `json:"db_encryption_key,omitempty"`
	// 分析引擎配置
	EnableAnalyzer        bool                       `json:"enable_analyzer"`
	AnalyzerMaxTimePerRun int                        `json:"analyzer_max_time_per_run"`
	AnalyzerThresholds    map[string]ThresholdConfig `json:"analyzer_thresholds,omitempty"`
	// 数据保留策略配置
	EnableRetentionPolicy bool `json:"enable_retention_policy"`
	RetentionDays         int  `json:"retention_days"`
	CleanupBatchSize      int  `json:"cleanup_batch_size"`
	// 告警通知配置
	EnableAlertNotifications bool   `json:"enable_alert_notifications"`
	AlertMQTTTopic           string `json:"alert_mqtt_topic,omitempty"`
	AlertMinSeverity         string `json:"alert_min_severity"`
	// 资源使用监控配置
	EnableResourceMonitoring bool    `json:"enable_resource_monitoring"`
	MaxMemoryMB              float64 `json:"max_memory_mb"`
	ResourceMonitorInterval  int     `json:"resource_monitor_interval_seconds"`
	// 数据库场景配置
	DBScenario string `json:"db_scenario"`
	// 自定义订阅主题
	CustomTopics []string `json:"custom_topics,omitempty"`
	// 数据同步配置
	EnableDataSync        bool   `json:"enable_data_sync"`
	DataSyncInterval      int    `json:"data_sync_interval_seconds"`
	DataSyncBatchSize     int    `json:"data_sync_batch_size"`
	DataSyncMaxRetryCount int    `json:"data_sync_max_retry"`
	DataSyncQueueDir      string `json:"data_sync_queue_dir,omitempty"`
	DataSyncUploadMode    string `json:"data_sync_upload_mode"`          // "http" 或 "mqtt"
	DataSyncBrokerURL     string `json:"data_sync_broker_url,omitempty"` // 云端 HTTP 端点或 MQTT Broker
	DataSyncMQTTTopic     string `json:"data_sync_mqtt_topic,omitempty"` // 云端 MQTT 话题
	DataSyncAuthToken     string `json:"data_sync_auth_token,omitempty"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Device string  `json:"device,omitempty"`
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

	if err := SaveToFile(newCfg); err != nil {
		log.Printf("Failed to save config to file: %v", err)
	}

	for _, handler := range handlers {
		handler(oldCfg, newCfg)
	}

	log.Println("Config updated successfully")
	return nil
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{
		DBPath:     "",
		MQTTBroker: "",
		MQTTTopic:  "",
		ClientID:   GenerateClientID(),
		HTTPPort:   "8081",
		// TLS 默认值
		MQTTUseTLS: false,
		HTTPUseTLS: false,
		// 数据库加密默认值
		DBUseEncryption: false,
		// 分析引擎默认值
		EnableAnalyzer:        false,
		AnalyzerMaxTimePerRun: 300,
		AnalyzerThresholds:    make(map[string]ThresholdConfig),
		// 数据保留策略默认值
		EnableRetentionPolicy: false,
		RetentionDays:         30,
		CleanupBatchSize:      500,
		// 告警通知默认值
		EnableAlertNotifications: false,
		AlertMQTTTopic:           "edgex/alerts",
		AlertMinSeverity:         "warning",
		// 资源使用监控默认值
		EnableResourceMonitoring: true,
		MaxMemoryMB:              32,
		ResourceMonitorInterval:  3,
		// 数据库场景默认值（极限生存模式，适合 128MB 以下内存设备）
		DBScenario: ScenarioExtreme,
		// 数据同步默认值
		EnableDataSync:        false,
		DataSyncInterval:      60,
		DataSyncBatchSize:     100,
		DataSyncMaxRetryCount: 3,
		DataSyncQueueDir:      "data/sync_queue",
		DataSyncUploadMode:    "http",
	}

	if err := loadFromConfigCenter(cfg); err != nil {
		log.Printf("Failed to load config from EdgeX config center: %v", err)
		log.Println("Falling back to local config file")

		if err := loadFromFile(cfg); err != nil {
			log.Printf("Failed to load config from file: %v", err)
			log.Println("Using default config")
		}
	}

	loadFromEnv(cfg)
	applySmartDefaults(cfg)

	GetConfigManager().SetConfig(cfg)

	return cfg, nil
}

// getSystemMemoryMB 获取系统总内存（MB）
func getSystemMemoryMB() uint64 {
	if runtime.GOOS == "windows" {
		return getWindowsSystemMemoryMB()
	}
	if runtime.GOOS == "linux" {
		return getLinuxSystemMemoryMB()
	}
	if runtime.GOOS == "darwin" {
		return getDarwinSystemMemoryMB()
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Sys / (1024 * 1024)
}

// getWindowsSystemMemoryMB 获取 Windows 系统总内存
func getWindowsSystemMemoryMB() uint64 {
	cmd := "wmic computersystem get totalphysicalmemory"
	output, err := runCommand(cmd)
	if err != nil {
		log.Printf("Failed to get system memory: %v", err)
		return 8192
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(strings.ToLower(line), "totalphysicalmemory") {
			continue
		}
		if val, err := strconv.ParseUint(line, 10, 64); err == nil {
			return val / (1024 * 1024)
		}
	}
	return 8192
}

// getLinuxSystemMemoryMB 获取 Linux 系统总内存
func getLinuxSystemMemoryMB() uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		log.Printf("Failed to read /proc/meminfo: %v", err)
		return 8192
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					return val / 1024
				}
			}
		}
	}
	return 8192
}

// getDarwinSystemMemoryMB 获取 macOS 系统总内存
func getDarwinSystemMemoryMB() uint64 {
	cmd := "sysctl hw.memsize"
	output, err := runCommand(cmd)
	if err != nil {
		log.Printf("Failed to get system memory: %v", err)
		return 8192
	}
	parts := strings.Fields(output)
	if len(parts) >= 2 {
		if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
			return val / (1024 * 1024)
		}
	}
	return 8192
}

// runCommand 执行命令并返回输出
func runCommand(cmd string) (string, error) {
	c := exec.Command("cmd", "/c", cmd)
	output, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// detectSystemMemory 检测系统内存并返回推荐配置
func detectSystemMemory() (totalMB uint64, recommendedScenario string, recommendedMaxMemoryMB float64) {
	totalMB = getSystemMemoryMB()

	switch {
	case totalMB <= 64:
		recommendedScenario = ScenarioEmbedded
		recommendedMaxMemoryMB = float64(totalMB) * 0.6
	case totalMB <= 128:
		recommendedScenario = ScenarioExtreme
		recommendedMaxMemoryMB = float64(totalMB) * 0.5
	case totalMB <= 256:
		recommendedScenario = ScenarioIoT
		recommendedMaxMemoryMB = float64(totalMB) * 0.4
	case totalMB <= 512:
		recommendedScenario = ScenarioEdge
		recommendedMaxMemoryMB = float64(totalMB) * 0.3
	case totalMB <= 1024:
		recommendedScenario = ScenarioDefault
		recommendedMaxMemoryMB = float64(totalMB) * 0.2
	default:
		recommendedScenario = ScenarioDefault
		recommendedMaxMemoryMB = float64(totalMB) * 0.15
	}

	return totalMB, recommendedScenario, recommendedMaxMemoryMB
}

// applySmartDefaults 应用智能默认值（自动检测系统内存）
func applySmartDefaults(cfg *Config) {
	if cfg.MQTTBroker == "" {
		cfg.MQTTBroker = "tcp://localhost:1883"
	}

	if cfg.MQTTTopic == "" {
		cfg.MQTTTopic = "edgex/events/#"
	}

	if cfg.DBPath == "" {
		if dbPath, err := pathutil.Join("data"); err == nil {
			cfg.DBPath = dbPath
		} else {
			cfg.DBPath = "data"
		}
	}

	totalMB, recommendedScenario, recommendedMaxMemoryMB := detectSystemMemory()
	log.Printf("Auto-detected system memory: %d MB", totalMB)

	if cfg.DBScenario == "" {
		cfg.DBScenario = recommendedScenario
		log.Printf("Auto-set DB scenario to: %s (based on memory detection)", cfg.DBScenario)
	}

	if cfg.MaxMemoryMB <= 0 {
		cfg.MaxMemoryMB = recommendedMaxMemoryMB
		log.Printf("Auto-set max memory limit to: %.2f MB", cfg.MaxMemoryMB)
	}

	log.Printf("Recommended configuration for %d MB system: scenario=%s, max_memory=%.2f MB",
		totalMB, recommendedScenario, recommendedMaxMemoryMB)

	ensureDataDir(cfg.DBPath)
}

// ensureDataDir 确保数据目录存在
func ensureDataDir(dbPath string) {
	dir := filepath.Dir(dbPath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create data directory: %v", err)
		}
	}
}

// GenerateClientID 生成客户端ID
func GenerateClientID() string {
	return "sfsdb-edgex-adapter-" + time.Now().Format("20060102150405")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
