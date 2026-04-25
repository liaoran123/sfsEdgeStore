# 配置管理模块 (Config) 技术文档

## 1. 概述

配置管理模块是sfsEdgeStore系统的核心组件之一，负责管理系统的所有配置信息。它提供了配置的加载、保存、更新和管理功能，支持从多个来源加载配置，包括EdgeX配置中心、本地配置文件和环境变量，并提供了配置更新的回调机制。

### 1.1 主要功能

- **多源配置加载**：支持从EdgeX配置中心、本地配置文件和环境变量加载配置
- **智能默认值**：提供合理的默认配置值
- **配置更新回调**：支持注册配置更新的回调函数
- **场景化配置**：根据不同的使用场景提供不同的数据库配置选项
- **许可证管理**：管理系统的许可证信息
- **配置持久化**：将配置保存到文件中

### 1.2 配置优先级

1. **环境变量**：优先级最高，会覆盖其他来源的配置
2. **EdgeX配置中心**：其次，从EdgeX配置中心加载配置
3. **本地配置文件**：再次，从本地配置文件加载配置
4. **默认配置**：最后，使用内置的默认配置

## 2. 数据结构

### 2.1 Config 结构体

```go
type Config struct {
    DBPath        string `json:"db_path" env:"EDGEX_DB_PATH"`
    MQTTBroker    string `json:"mqtt_broker" env:"EDGEX_MQTT_BROKER"`
    MQTTTopic     string `json:"mqtt_topic" env:"EDGEX_MQTT_TOPIC"`
    ClientID      string `json:"client_id" env:"EDGEX_CLIENT_ID"`
    HTTPPort      string `json:"http_port" env:"EDGEX_HTTP_PORT"`
    DevConfigPath string `json:"dev_config_path" env:"EDGEX_DEV_CONFIG_PATH"`
    AutoSubscribe bool   `json:"auto_subscribe" env:"EDGEX_AUTO_SUBSCRIBE"`
    // TLS 配置
    MQTTUseTLS     bool   `json:"mqtt_use_tls" env:"EDGEX_MQTT_USE_TLS"`
    MQTTCACert     string `json:"mqtt_ca_cert" env:"EDGEX_MQTT_CA_CERT"`
    MQTTClientCert string `json:"mqtt_client_cert" env:"EDGEX_MQTT_CLIENT_CERT"`
    MQTTClientKey  string `json:"mqtt_client_key" env:"EDGEX_MQTT_CLIENT_KEY"`
    // MQTT 连接配置
    ConnectionTimeout int `json:"connection_timeout" env:"EDGEX_MQTT_CONNECTION_TIMEOUT"`
    KeepAlive         int `json:"keep_alive" env:"EDGEX_MQTT_KEEP_ALIVE"`
    // MQTT 认证
    MQTTUsername string `json:"mqtt_username" env:"EDGEX_MQTT_USERNAME"`
    MQTTPassword string `json:"mqtt_password" env:"EDGEX_MQTT_PASSWORD"`
    HTTPUseTLS   bool   `json:"http_use_tls" env:"EDGEX_HTTP_USE_TLS"`
    HTTPCert     string `json:"http_cert" env:"EDGEX_HTTP_CERT"`
    HTTPKey      string `json:"http_key" env:"EDGEX_HTTP_KEY"`
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
    // 数据同步配置（企业版功能）
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
    // 设备异常监控配置
    DeviceOfflineThreshold int `json:"device_offline_threshold_seconds" env:"EDGEX_DEVICE_OFFLINE_THRESHOLD_SECONDS"`
    DataAnomalyThreshold   int `json:"data_anomaly_threshold_percent" env:"EDGEX_DATA_ANOMALY_THRESHOLD_PERCENT"`
    DataTrendMinPoints     int `json:"data_trend_min_points" env:"EDGEX_DATA_TREND_MIN_POINTS"`
    // 数据库场景配置
    DBScenario string `json:"db_scenario" env:"EDGEX_DB_SCENARIO"`
    // Prometheus 指标配置（可选，默认关闭）
    EnablePrometheus bool   `json:"enable_prometheus" env:"EDGEX_ENABLE_PROMETHEUS"`
    PrometheusPath   string `json:"prometheus_path" env:"EDGEX_PROMETHEUS_PATH"`
    // 模拟器配置
    EnableSimulator      bool `json:"enable_simulator" env:"EDGEX_ENABLE_SIMULATOR"`
    SimulatorIntervalMin int  `json:"simulator_interval_min" env:"EDGEX_SIMULATOR_INTERVAL_MIN"`
    SimulatorIntervalMax int  `json:"simulator_interval_max" env:"EDGEX_SIMULATOR_INTERVAL_MAX"`
    // 许可证配置
    LicenseType        string             `json:"license_type" env:"EDGEX_LICENSE_TYPE"` // "community" | "business" | "enterprise"
    LicenseKey         string             `json:"license_key" env:"EDGEX_LICENSE_KEY"`   // 许可证密钥
    // 行业模板配置
    IndustryTemplate   string             `json:"industry_template" env:"EDGEX_INDUSTRY_TEMPLATE"` // 行业模板
    // 设备配置
    Devices            []map[string]interface{} `json:"devices"`
    // 告警配置
    Alerts             []map[string]interface{} `json:"alerts"`
    // 基线配置
    Baseline           map[string]interface{}   `json:"baseline"`
    // 安全红线配置
    SafetyLimits       map[string]interface{}   `json:"safety_limits"`
    EnterpriseFeatures EnterpriseFeatures `json:"enterprise_features"`                   // 功能开关
    // 自定义订阅主题
    CustomTopics []string `json:"custom_topics"` // 自定义MQTT订阅主题
}
```

### 2.2 EnterpriseFeatures 结构体

```go
type EnterpriseFeatures struct {
    EnableCloudSync         bool `json:"enable_cloud_sync"`         // 云端数据同步
    EnableRemoteConfig      bool `json:"enable_remote_config"`      // 远程配置管理
    EnableMultiTenant       bool `json:"enable_multi_tenant"`       // 多租户支持
    EnableAdvancedAnalytics bool `json:"enable_advanced_analytics"` // 高级数据分析
    EnableBigScreenMode     bool `json:"enable_big_screen_mode"`    // 本地大屏模式（解锁多图表排版）
    MaxDevices              int  `json:"max_devices"`               // 最大设备数限制（0表示无限制）
}
```

### 2.3 ThresholdConfig 结构体

```go
type ThresholdConfig struct {
    Min float64 `json:"min"`
    Max float64 `json:"max"`
}
```

### 2.4 ConfigManager 结构体

```go
type ConfigManager struct {
    currentConfig  *Config
    mutex          sync.RWMutex
    updateHandlers []ConfigUpdateHandler
}
```

### 2.5 License 结构体

```go
type License struct {
    LicenseType string `json:"license_type"` // "community" | "business" | "enterprise"
    IssuedDate  string `json:"issued_date"`  // 发证日期，ISO8601格式
    ExpiryDate  string `json:"expiry_date"`  // 到期日期，ISO8601格式
    MaxDevices  int    `json:"max_devices"`  // 最大设备数
    LicenseKey  string `json:"license_key"`  // 许可证密钥
}
```

## 3. 常量定义

### 3.1 数据库场景常量

```go
const (
    ScenarioEmbedded = "embedded" // 嵌入式场景
    ScenarioIoT      = "iot"      // IoT场景
    ScenarioEdge     = "edge"     // 边缘计算场景
    ScenarioGame     = "game"     // 游戏场景
    ScenarioDefault  = "default"  // 默认场景
)
```

### 3.2 许可证常量

```go
// 宽限期（天） - 到期后 30 天内依然有温馨提示
const GracePeriodDays = 30

// 许可证文件路径
const LicenseFilePath = "license.json"
```

## 4. 函数说明

### 4.1 配置管理

#### Load
```go
func Load() (*Config, error)
```

**功能**：加载配置

**处理逻辑**：
1. 设置默认配置
2. 尝试从EdgeX配置中心加载
3. 从配置文件加载（如果EdgeX加载失败）
4. 从环境变量加载（优先级最高）
5. 应用智能默认值
6. 设置到配置管理器

**返回值**：
- `*Config`: 加载的配置
- `error`: 错误对象（成功时为nil）

#### GetConfigManager
```go
func GetConfigManager() *ConfigManager
```

**功能**：获取配置管理器单例

**返回值**：配置管理器实例

#### SetConfig
```go
func (cm *ConfigManager) SetConfig(cfg *Config)
```

**功能**：设置当前配置

**参数**：
- `cfg`: 配置对象

#### GetConfig
```go
func (cm *ConfigManager) GetConfig() *Config
```

**功能**：获取当前配置

**返回值**：当前配置对象

#### RegisterUpdateHandler
```go
func (cm *ConfigManager) RegisterUpdateHandler(handler ConfigUpdateHandler)
```

**功能**：注册配置更新回调

**参数**：
- `handler`: 配置更新回调函数

#### UpdateConfig
```go
func (cm *ConfigManager) UpdateConfig(newCfg *Config) error
```

**功能**：更新配置

**参数**：
- `newCfg`: 新配置对象

**处理逻辑**：
1. 保存旧配置
2. 设置新配置
3. 保存配置到文件
4. 通知所有更新处理器

**返回值**：
- `error`: 错误对象（成功时为nil）

#### GetScenarioOptions
```go
func (cm *ConfigManager) GetScenarioOptions() *opt.Options
```

**功能**：根据场景获取数据库配置选项

**返回值**：
- `*opt.Options`: 数据库配置选项

### 4.2 辅助函数

#### applySmartDefaults
```go
func applySmartDefaults(cfg *Config)
```

**功能**：应用智能默认值

**参数**：
- `cfg`: 配置对象

**处理逻辑**：
1. 设置MQTT Broker默认值
2. 设置MQTT Topic默认值
3. 设置DB Path默认值
4. 确保数据目录存在

#### ensureDataDir
```go
func ensureDataDir(dbPath string)
```

**功能**：确保数据目录存在

**参数**：
- `dbPath`: 数据库路径

#### GenerateClientID
```go
func GenerateClientID() string
```

**功能**：生成客户端ID

**返回值**：
- `string`: 生成的客户端ID

### 4.3 许可证管理

#### LoadLicense
```go
func LoadLicense() (*License, error)
```

**功能**：加载许可证

**处理逻辑**：
1. 尝试读取独立的license.json
2. 如果没有找到，返回默认的社区版许可证

**返回值**：
- `*License`: 许可证对象
- `error`: 错误对象（成功时为nil）

#### SaveLicense
```go
func SaveLicense(lic *License) error
```

**功能**：保存许可证

**参数**：
- `lic`: 许可证对象

**返回值**：
- `error`: 错误对象（成功时为nil）

#### GetStatus
```go
func (l *License) GetStatus() LicenseStatus
```

**功能**：获取许可证状态

**返回值**：
- `LicenseStatus`: 许可证状态

#### GetStatusText
```go
func (l *License) GetStatusText() string
```

**功能**：获取状态文本

**返回值**：
- `string`: 状态文本

#### GetRemainingDays
```go
func (l *License) GetRemainingDays() int
```

**功能**：获取剩余天数

**返回值**：
- `int`: 剩余天数

#### ShouldShowRenewalNotice
```go
func (l *License) ShouldShowRenewalNotice() bool
```

**功能**：是否显示续费提示

**返回值**：
- `bool`: 是否显示续费提示

#### PrintLicenseInfo
```go
func PrintLicenseInfo(l *License)
```

**功能**：打印许可证信息

**参数**：
- `l`: 许可证对象

## 5. 配置文件格式

### 5.1 配置文件示例

```json
{
  "db_path": "./data/sfs.db",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/#",
  "http_port": "8081",
  "dev_config_path": "./devconfig",
  "auto_subscribe": true,
  "mqtt_use_tls": false,
  "http_use_tls": false,
  "db_use_encryption": false,
  "enable_analyzer": false,
  "enable_retention_policy": false,
  "retention_days": 30,
  "cleanup_interval_hours": 24,
  "cleanup_batch_size": 1000,
  "enable_alert_notifications": false,
  "alert_mqtt_topic": "edgex/alerts",
  "enable_data_sync": false,
  "data_sync_mqtt_topic": "edgex/data/sync",
  "enable_resource_monitoring": false,
  "max_memory_mb": 50,
  "max_cpu_percent": 5,
  "resource_monitor_interval_seconds": 10,
  "device_offline_threshold_seconds": 300,
  "data_anomaly_threshold_percent": 50,
  "data_trend_min_points": 5,
  "db_scenario": "edge",
  "enable_prometheus": false,
  "prometheus_path": "/metrics",
  "enable_simulator": false,
  "simulator_interval_min": 5,
  "simulator_interval_max": 10,
  "license_type": "community",
  "industry_template": "",
  "enterprise_features": {
    "enable_cloud_sync": false,
    "enable_remote_config": false,
    "enable_multi_tenant": false,
    "enable_advanced_analytics": false,
    "enable_big_screen_mode": false,
    "max_devices": 5
  },
  "custom_topics": []
}
```

## 6. 环境变量

### 6.1 常用环境变量

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| EDGEX_DB_PATH | 数据库路径 | ./data/sfs.db |
| EDGEX_MQTT_BROKER | MQTT broker地址 | tcp://localhost:1883 |
| EDGEX_MQTT_TOPIC | MQTT订阅主题 | edgex/events/# |
| EDGEX_HTTP_PORT | HTTP端口 | 8081 |
| EDGEX_DEV_CONFIG_PATH | 设备配置路径 | ./devconfig |
| EDGEX_AUTO_SUBSCRIBE | 是否自动订阅 | true |
| EDGEX_MQTT_USE_TLS | MQTT是否使用TLS | false |
| EDGEX_HTTP_USE_TLS | HTTP是否使用TLS | false |
| EDGEX_DB_USE_ENCRYPTION | 数据库是否使用加密 | false |
| EDGEX_ENABLE_ANALYZER | 是否启用分析引擎 | false |
| EDGEX_ENABLE_RETENTION_POLICY | 是否启用数据保留策略 | false |
| EDGEX_RETENTION_DAYS | 数据保留天数 | 30 |
| EDGEX_CLEANUP_INTERVAL_HOURS | 清理间隔（小时） | 24 |
| EDGEX_CLEANUP_BATCH_SIZE | 每批清理记录数 | 1000 |
| EDGEX_ENABLE_ALERT_NOTIFICATIONS | 是否启用告警通知 | false |
| EDGEX_ENABLE_DATA_SYNC | 是否启用数据同步 | false |
| EDGEX_ENABLE_RESOURCE_MONITORING | 是否启用资源监控 | false |
| EDGEX_MAX_MEMORY_MB | 内存限制（MB） | 50 |
| EDGEX_MAX_CPU_PERCENT | CPU限制（%） | 5 |
| EDGEX_RESOURCE_MONITOR_INTERVAL_SECONDS | 资源监控间隔（秒） | 10 |
| EDGEX_DB_SCENARIO | 数据库场景 | edge |
| EDGEX_ENABLE_SIMULATOR | 是否启用模拟器 | false |
| EDGEX_LICENSE_TYPE | 许可证类型 | community |

## 7. 使用示例

### 7.1 加载配置

```go
// 加载配置
cfg, err := config.Load()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 使用配置
fmt.Printf("MQTT Broker: %s\n", cfg.MQTTBroker)
fmt.Printf("HTTP Port: %s\n", cfg.HTTPPort)
fmt.Printf("DB Path: %s\n", cfg.DBPath)
```

### 7.2 配置更新

```go
// 获取配置管理器
cm := config.GetConfigManager()

// 注册配置更新回调
cm.RegisterUpdateHandler(func(oldCfg, newCfg *config.Config) {
    log.Printf("Config updated: MQTT Broker changed from %s to %s", oldCfg.MQTTBroker, newCfg.MQTTBroker)
})

// 创建新配置
newCfg := *cm.GetConfig()
newCfg.MQTTBroker = "tcp://broker.example.com:1883"

// 更新配置
if err := cm.UpdateConfig(&newCfg); err != nil {
    log.Printf("Failed to update config: %v", err)
}
```

### 7.3 许可证管理

```go
// 加载许可证
lic, err := config.LoadLicense()
if err != nil {
    log.Printf("Failed to load license: %v", err)
}

// 打印许可证信息
config.PrintLicenseInfo(lic)

// 检查许可证状态
status := lic.GetStatus()
fmt.Printf("License status: %s\n", lic.GetStatusText())

// 检查是否需要显示续费提示
if lic.ShouldShowRenewalNotice() {
    fmt.Println("Renewal notice should be shown")
}
```

## 8. 场景化配置

### 8.1 数据库场景配置

| 场景 | WriteBuffer | OpenFilesCacheCapacity | BlockCacheCapacity | Compression |
|------|------------|------------------------|--------------------|-------------|
| embedded | 2MB | 5 | 4MB | Default |
| iot | 4MB | 10 | 8MB | Default |
| edge | 16MB | 50 | 32MB | Default |
| game | 64MB | 200 | 128MB | NoCompression |
| default | 64MB | 200 | 128MB | Default |

## 9. 性能优化

### 9.1 配置加载优化

- 优先级机制：环境变量 > EdgeX配置中心 > 本地文件 > 默认配置
- 智能默认值：根据环境自动设置合理的默认值
- 数据目录确保：自动创建所需的数据目录

### 9.2 配置更新优化

- 线程安全：使用互斥锁保护配置的读写
- 回调机制：配置更新时通知所有注册的处理器
- 持久化：自动保存配置到文件

### 9.3 许可证管理优化

- 缓存机制：许可证信息加载后缓存
- 宽限期：过期后30天内仍有温馨提示
- 自动补全：自动补全默认值

## 10. 错误处理

### 10.1 常见错误

| 错误场景 | 原因 | 处理方式 |
|---------|------|---------|
| 配置文件读取失败 | 文件不存在或格式错误 | 使用默认配置 |
| EdgeX配置中心连接失败 | 网络问题或服务未运行 | 回退到本地配置文件 |
| 数据目录创建失败 | 权限不足 | 记录错误日志，继续运行 |
| 许可证文件读取失败 | 文件不存在或格式错误 | 使用默认社区版许可证 |

## 11. 注意事项

1. **配置优先级**：环境变量优先级最高，会覆盖其他配置
2. **数据目录**：确保数据目录有写入权限
3. **许可证**：社区版有5个设备限制，商业版和企业版有不同的限制
4. **场景配置**：选择合适的数据库场景以获得最佳性能
5. **安全配置**：生产环境建议启用TLS和数据库加密

## 12. 总结

配置管理模块为sfsEdgeStore系统提供了灵活、强大的配置管理能力，支持多源配置加载、智能默认值、配置更新回调等功能。通过场景化配置和许可证管理，满足了不同用户的需求。其设计遵循以下原则：

- **灵活性**：支持多种配置来源和优先级
- **可靠性**：完善的错误处理和回退机制
- **易用性**：智能默认值和自动配置
- **安全性**：支持TLS和数据库加密
- **可扩展性**：配置更新回调机制

通过配置管理模块，用户可以轻松配置和管理sfsEdgeStore系统，适应不同的使用场景和需求。