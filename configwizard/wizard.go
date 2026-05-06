package configwizard

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"sfsEdgeStore/config"
)

// Wizard 配置向导结构体
type Wizard struct {
	config *config.Config
}

// NewWizard 创建新的配置向导
func NewWizard(cfg *config.Config) *Wizard {
	return &Wizard{
		config: cfg,
	}
}

// Run 运行配置向导
func (w *Wizard) Run() error {
	log.Println("Starting configuration wizard...")

	// 检查是否首次启动
	if w.isFirstRun() {
		log.Println("First run detected, starting setup wizard...")
		if err := w.firstRunWizard(); err != nil {
			return err
		}
	}

	// 检测EdgeX实例
	if err := w.detectEdgeX(); err != nil {
		log.Printf("EdgeX detection failed: %v", err)
		// 继续运行，不阻止启动
	}

	// 检测系统环境
	if err := w.detectEnvironment(); err != nil {
		log.Printf("Environment detection failed: %v", err)
		// 继续运行，不阻止启动
	}

	// 优化配置
	if err := w.optimizeConfig(); err != nil {
		log.Printf("Configuration optimization failed: %v", err)
		// 继续运行，不阻止启动
	}

	// 验证配置
	if err := w.validateConfig(); err != nil {
		log.Printf("Config validation failed: %v", err)
		// 继续运行，不阻止启动
	}

	log.Println("Configuration wizard completed successfully")
	return nil
}

// isFirstRun 检查是否首次运行
func (w *Wizard) isFirstRun() bool {
	// 检查配置文件是否存在
	configPath := "config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return true
	}

	// 检查数据目录是否存在
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return true
	}

	return false
}

// firstRunWizard 首次启动向导
func (w *Wizard) firstRunWizard() error {
	log.Println("=====================================")
	log.Println("        sfsEdgeStore 首次启动向导")
	log.Println("=====================================")

	// 创建数据目录
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}
	log.Println("✓ 创建数据目录成功")

	// 保存默认配置
	if err := config.SaveToFile(w.config); err != nil {
		return fmt.Errorf("failed to save default config: %v", err)
	}
	log.Println("✓ 保存默认配置成功")

	log.Println("=====================================")
	log.Println("首次启动向导完成！")
	log.Println("=====================================")

	return nil
}

// detectEdgeX 检测EdgeX实例
func (w *Wizard) detectEdgeX() error {
	log.Println("Detecting EdgeX instances...")

	// 尝试连接本地MQTT broker
	brokerAddr := "localhost:1883"
	conn, err := net.DialTimeout("tcp", brokerAddr, 2*time.Second)
	if err == nil {
		conn.Close()
		log.Printf("✓ EdgeX MQTT broker detected at %s", brokerAddr)
		// 如果配置中的broker为空，设置为本地地址
		if w.config.MQTTBroker == "" {
			w.config.MQTTBroker = "tcp://" + brokerAddr
			log.Println("✓ Updated MQTT broker configuration")
		}
		// 智能设置订阅主题
		if w.config.MQTTTopic == "" {
			w.config.MQTTTopic = "edgex/events/#"
			log.Println("✓ Updated MQTT topic configuration")
		}
	} else {
		log.Printf("EdgeX MQTT broker not found at %s: %v", brokerAddr, err)
		// 设置默认值
		if w.config.MQTTBroker == "" {
			w.config.MQTTBroker = "tcp://localhost:1883"
			log.Println("⚠ Using default MQTT broker: tcp://localhost:1883")
		}
		if w.config.MQTTTopic == "" {
			w.config.MQTTTopic = "edgex/events/#"
			log.Println("⚠ Using default MQTT topic: edgex/events/#")
		}
	}

	// 尝试连接其他常见端口
	commonPorts := []string{"1884", "1885", "8883"}
	for _, port := range commonPorts {
		brokerAddr := "localhost:" + port
		conn, err := net.DialTimeout("tcp", brokerAddr, 1*time.Second)
		if err == nil {
			conn.Close()
			log.Printf("✓ MQTT broker detected at %s", brokerAddr)
			// 如果当前配置的broker不可用，使用检测到的地址
			if w.config.MQTTBroker == "" || w.config.MQTTBroker == "tcp://localhost:1883" {
				w.config.MQTTBroker = "tcp://" + brokerAddr
				log.Println("✓ Updated MQTT broker configuration to detected address")
			}
			break
		}
	}

	return nil
}

// detectEnvironment 检测系统环境
func (w *Wizard) detectEnvironment() error {
	log.Println("Detecting system environment...")

	// 检测系统资源
	sysInfo := getSystemInfo()
	log.Printf("System info: CPU cores=%d, Memory=%d MB", sysInfo.CPUCores, sysInfo.MemoryMB)

	// 根据系统资源自动调整配置
	if sysInfo.MemoryMB < 512 {
		// 低内存环境
		log.Println("⚠ Low memory environment detected, optimizing for resource usage")
		w.config.DBScenario = "embedded"
		w.config.EnableAnalyzer = false
		w.config.EnableResourceMonitoring = true
	} else if sysInfo.MemoryMB < 2048 {
		// 中等内存环境
		log.Println("Detected medium memory environment")
		w.config.DBScenario = "iot"
		w.config.EnableAnalyzer = false
	} else {
		// 高内存环境
		log.Println("Detected high memory environment")
		w.config.DBScenario = "edge"
		w.config.EnableAnalyzer = true
	}

	// 检测网络环境
	if isNetworkAvailable() {
		log.Println("✓ Network connectivity detected")
	} else {
		log.Println("⚠ No network connectivity detected")
	}

	return nil
}

// optimizeConfig 优化配置
func (w *Wizard) optimizeConfig() error {
	log.Println("Optimizing configuration...")

	if w.config.DBPath == "" {
		w.config.DBPath = "./data"
		log.Println("✓ Set default database path: ./data")
	}

	if w.config.HTTPPort == "" {
		w.config.HTTPPort = "8081"
		log.Println("✓ Set default HTTP port: 8081")
	}

	w.config.MQTTUseTLS = false
	w.config.HTTPUseTLS = false
	w.config.DBUseEncryption = false
	log.Println("✓ Optimized security settings for ease of use")

	w.config.EnableRetentionPolicy = false
	w.config.EnableAlertNotifications = false
	w.config.EnableResourceMonitoring = true
	log.Println("✓ Optimized feature settings")

	dbDir := filepath.Dir(w.config.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	log.Println("✓ Configuration optimization completed")
	return nil
}

// SystemInfo 系统信息
 type SystemInfo struct {
	CPUCores  int
	MemoryMB  int
	OS        string
	Arch      string
}

// getSystemInfo 获取系统信息
func getSystemInfo() SystemInfo {
	info := SystemInfo{
		CPUCores: runtime.NumCPU(),
		MemoryMB: int(memoryStats() / (1024 * 1024)),
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}
	return info
}

// memoryStats 获取内存使用情况
func memoryStats() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.TotalAlloc
}

// isNetworkAvailable 检测网络连接
func isNetworkAvailable() bool {
	_, err := net.DialTimeout("tcp", "8.8.8.8:53", 3*time.Second)
	return err == nil
}

// validateConfig 验证配置
func (w *Wizard) validateConfig() error {
	log.Println("Validating configuration...")

	// 验证MQTT broker配置
	if w.config.MQTTBroker == "" {
		log.Println("⚠ MQTT broker not configured, using default: tcp://localhost:1883")
		w.config.MQTTBroker = "tcp://localhost:1883"
	}

	// 验证MQTT主题
	if w.config.MQTTTopic == "" {
		log.Println("⚠ MQTT topic not configured, using default: edgex/events/#")
		w.config.MQTTTopic = "edgex/events/#"
	}

	// 验证客户端ID
	if w.config.ClientID == "" {
		w.config.ClientID = config.GenerateClientID()
		log.Println("⚠ Client ID not configured, generated automatically")
	}

	// 验证数据目录
	if w.config.DBPath == "" {
		w.config.DBPath = "./data"
		log.Println("⚠ Database path not configured, using default: ./data")
	}

	// 确保数据目录存在
	dbDir := filepath.Dir(w.config.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// 验证HTTP端口
	if w.config.HTTPPort == "" {
		w.config.HTTPPort = "8081"
		log.Println("⚠ HTTP port not configured, using default: 8081")
	}

	// 验证数据库场景
	if w.config.DBScenario == "" {
		w.config.DBScenario = "extreme"
		log.Println("⚠ Database scenario not configured, using default: extreme")
	}

	// 验证资源监控设置
	if !w.config.EnableResourceMonitoring {
		w.config.EnableResourceMonitoring = true
		log.Println("⚠ Resource monitoring not enabled, enabling by default")
	}

	// 验证分析引擎设置
	if w.config.EnableAnalyzer {
		if w.config.AnalyzerMaxTimePerRun == 0 {
			w.config.AnalyzerMaxTimePerRun = 500 // 500ms
		}
	}

	// 验证数据保留策略设置
	if w.config.EnableRetentionPolicy {
		if w.config.RetentionDays == 0 {
			w.config.RetentionDays = 30 // 30天
		}
		if w.config.CleanupBatchSize == 0 {
			w.config.CleanupBatchSize = 500 // 每批500条
		}
	}

	log.Println("✓ Configuration validation completed")
	return nil
}
