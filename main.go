// sfsDb 与 EdgeX MQTT 适配器示例（改进版）
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"sfsEdgeStore/alert"
	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/auth"
	"sfsEdgeStore/config"
	"sfsEdgeStore/configwizard"
	"sfsEdgeStore/database"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/mqtt"
	"sfsEdgeStore/resource"
	"sfsEdgeStore/retention"
	"sfsEdgeStore/server"
)

/*
	func init() {
		debug.SetGCPercent(100)
	}
*/
type Components struct {
	Monitor         *monitor.Monitor
	AlertNotifier   *alert.Notifier
	Analyzer        *analyzer.Analyzer
	Retention       *retention.RetentionManager
	ResourceMonitor *resource.ResourceMonitor
	Server          *server.Server
	MQTTClient      *mqtt.Client
}

func main() {
	appConfig, err := initConfig()
	if err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	limitMB := int64(appConfig.MaxMemoryMB)
	if limitMB <= 0 {
		limitMB = 32
	}
	goLimitMB := limitMB * 4
	if goLimitMB < 128 {
		goLimitMB = 128
	}
	debug.SetMemoryLimit(goLimitMB * 1024 * 1024)
	log.Printf("Go runtime memory limit set to %d MB (RSS alert threshold: %d MB)", goLimitMB, limitMB)

	components, err := initComponents(appConfig)
	if err != nil {
		log.Fatalf("组件初始化失败: %v", err)
	}

	startServices(components, appConfig)
	waitForShutdown(components)
}

func initConfig() (*config.Config, error) {
	appConfig, err := config.Load()
	if err != nil {
		log.Printf("没有找到 config.json，使用默认配置")
	}

	wizard := configwizard.NewWizard(appConfig)
	if err := wizard.Run(); err != nil {
		log.Printf("配置向导失败: %v", err)
	}

	return appConfig, nil
}

func initComponents(appConfig *config.Config) (*Components, error) {
	monitorInstance := monitor.NewMonitor()

	alertNotifier := alert.NewNotifier(appConfig)
	if err := alertNotifier.Start(); err != nil {
		log.Printf("告警通知器启动失败: %v", err)
	}

	analyzerInstance := analyzer.NewAnalyzer(appConfig)
	if appConfig.EnableAnalyzer {
		log.Println("分析引擎已启用")
	} else {
		log.Println("分析引擎已禁用")
	}

	if err := database.Init(appConfig.DBPath, appConfig.DBUseEncryption, appConfig.DBEncryptionKey, ""); err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %v", err)
	}

	authManager := auth.NewAuthManager()
	authManager.StartCleanupTask(24 * time.Hour)

	retentionManager := retention.NewRetentionManager(database.Table, appConfig)
	if err := retentionManager.Start(); err != nil {
		log.Printf("数据保留策略管理器启动失败: %v", err)
	}

	resourceMonitor := resource.NewResourceMonitor(appConfig, monitorInstance)
	if err := resourceMonitor.Start(); err != nil {
		log.Printf("资源监控器启动失败: %v", err)
	}

	serverInstance := server.NewServer(database.Table, appConfig, monitorInstance, retentionManager, alertNotifier, resourceMonitor)
	serverInstance.SeedDeviceStatusCache()

	mqttClient, err := mqtt.NewClient(appConfig, monitorInstance, analyzerInstance)
	if err != nil {
		log.Printf("MQTT客户端初始化失败: %v", err)
	} else {
		// 启动广播监听通道
		serverInstance.StartBroadcast(mqttClient.BroadcastChan())
		mqttClient.SetOnRecordStored(serverInstance.OnRecordStored)
		if err := mqttClient.Subscribe(mqtt.GetTopics(appConfig)); err != nil {
			log.Printf("MQTT订阅失败: %v", err)
		}
	}

	return &Components{
		Monitor:         monitorInstance,
		AlertNotifier:   alertNotifier,
		Analyzer:        analyzerInstance,
		Retention:       retentionManager,
		ResourceMonitor: resourceMonitor,
		Server:          serverInstance,
		MQTTClient:      mqttClient,
	}, nil
}

func printWelcome() {
	fmt.Print(`
╔═══════════════════════════════════════════════════════════════╗
║                    sfsEdgeStore 启动成功                       ║
╠═══════════════════════════════════════════════════════════════╣
║  轻量级工业物联网边缘数据存储适配器                             ║
║  内存占用: <50MB | 极轻量 | 高可靠                              ║
╠═══════════════════════════════════════════════════════════════╣
║  Web 监控界面: http://localhost:8081                          ║
║  停止方法: Ctrl+C                                              ║
╚═══════════════════════════════════════════════════════════════╝
`)
}

func startServices(c *Components, appConfig *config.Config) {
	printWelcome()

	log.Printf("MQTT Broker: %s", appConfig.MQTTBroker)
	log.Printf("MQTT Topic: %s", appConfig.MQTTTopic)
	log.Printf("HTTP端口: %s", appConfig.HTTPPort)
	log.Println("sfsEdgeStore started successfully")

	if err := c.Server.Start(); err != nil {
		log.Fatalf("HTTP服务器启动失败: %v", err)
	}
}

func waitForShutdown(c *Components) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	if c.Retention != nil {
		c.Retention.Stop()
	}

	if c.AlertNotifier != nil {
		c.AlertNotifier.Stop()
	}

	if c.ResourceMonitor != nil {
		c.ResourceMonitor.Stop()
	}

	if c.MQTTClient != nil {
		c.MQTTClient.Disconnect()
	}

	time.Sleep(5 * time.Second)

	log.Println("服务已退出")
}
