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

	"sfsEdgeStore/agent"
	"sfsEdgeStore/alert"
	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/auth"
	"sfsEdgeStore/cli"
	"sfsEdgeStore/cloudsync/sync"
	"sfsEdgeStore/config"
	"sfsEdgeStore/configwizard"
	"sfsEdgeStore/core/database"
	"sfsEdgeStore/core/mqtt"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/pathutil"
	"sfsEdgeStore/queue"
	"sfsEdgeStore/resource"
	"sfsEdgeStore/retention"
	"sfsEdgeStore/server"
)

func init() {
	// GC 优化：平衡内存和 CPU
	debug.SetGCPercent(100) // 默认值，避免过度 GC 消耗 CPU
}

type Components struct {
	Monitor         *monitor.Monitor
	AlertNotifier   *alert.Notifier
	Analyzer        *analyzer.Analyzer
	Agent           *agent.Agent
	Retention       *retention.RetentionManager
	ResourceMonitor *resource.ResourceMonitor
	SyncManager     *sync.SyncManager
	Server          *server.Server
	MQTTClient      *mqtt.Client
	DataQueue       *queue.Queue
}

func main() {
	// 1. 解析命令行参数
	args := cli.Parse()
	if args.Help {
		args.ShowHelp()
		return
	}

	// 2. 初始化配置
	appConfig, appLicense, err := initConfig(args)
	if err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 3. 初始化核心组件
	components, err := initComponents(appConfig)
	if err != nil {
		log.Fatalf("组件初始化失败: %v", err)
	}

	// 4. 启动服务
	startServices(components, appConfig, appLicense)

	// 5. 等待中断信号并优雅关闭
	waitForShutdown(components)
}

func initConfig(args *cli.Args) (*config.Config, *config.License, error) {
	// 加载配置
	appConfig, err := config.Load()
	if err != nil {
		log.Printf("没有找到 config.json，使用默认配置")
	}

	// 运行配置向导
	wizard := configwizard.NewWizard(appConfig)
	if err := wizard.Run(); err != nil {
		log.Printf("配置向导失败: %v", err)
	}

	// 加载许可证
	appLicense, err := config.LoadLicense()
	if err != nil {
		log.Printf("加载许可证失败: %v，使用默认社区版", err)
		appLicense = &config.License{
			LicenseType: "community",
			MaxDevices:  5,
		}
	}

	// 更新配置中的许可证信息
	appConfig.LicenseType = appLicense.LicenseType
	appConfig.EnterpriseFeatures.MaxDevices = appLicense.GetMaxDevices()

	// 命令行参数覆盖配置
	if args.MQTTBroker != "" {
		appConfig.MQTTBroker = args.MQTTBroker
	}
	if args.MQTTTopic != "" {
		appConfig.MQTTTopic = args.MQTTTopic
	}
	if args.HTTPPort != "" {
		appConfig.HTTPPort = args.HTTPPort
	}

	return appConfig, appLicense, nil
}

func initComponents(appConfig *config.Config) (*Components, error) {
	// 初始化监控
	monitorInstance := monitor.NewMonitor(appConfig)

	// 启动监控清理任务（定期清理过期设备和告警）
	monitorInstance.StartCleanupRoutine()

	// 初始化告警通知器
	alertNotifier := alert.NewNotifier(appConfig)
	monitorInstance.SetNotifier(alertNotifier)
	if err := alertNotifier.Start(); err != nil {
		log.Printf("告警通知器启动失败: %v", err)
	}

	// 看门狗已禁用（功能与资源监控器重复，且阈值500MB永远不会触发）
	// watchdogInstance := watchdog.NewWatchdog(monitorInstance)
	// watchdogInstance.Start()

	// 初始化分析引擎
	analyzerInstance := analyzer.NewAnalyzer(appConfig)
	if appConfig.EnableAnalyzer {
		log.Println("分析引擎已启用")
	} else {
		log.Println("分析引擎已禁用")
	}

	// 连接数据库
	if err := database.Init(appConfig.DBPath, appConfig.DBUseEncryption, appConfig.DBEncryptionKey, appConfig.DBEncryptionAlgorithm); err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %v", err)
	}

	// 启动认证清理任务（24小时清理一次，极低频率）
	authManager := auth.NewAuthManager()
	authManager.StartCleanupTask(24 * time.Hour)

	//// 初始化数据队列
	dataQueuePath, _ := pathutil.Join("data_queue")
	dataQueue, err := queue.NewQueue(dataQueuePath)
	if err != nil {
		return nil, fmt.Errorf("数据队列初始化失败: %v", err)
	}

	// 初始化Agent（已禁用，减少 CPU 占用）
	// agentInstance, err := agent.NewAgent(appConfig, monitorInstance)
	// if err != nil {
	// 	log.Printf("Agent初始化失败: %v", err)
	// }
	var agentInstance *agent.Agent

	// 初始化数据保留策略管理器（低频：每5分钟检查一次）
	retentionManager := retention.NewRetentionManager(database.Table, appConfig)
	if err := retentionManager.Start(); err != nil {
		log.Printf("数据保留策略管理器启动失败: %v", err)
	}

	// 初始化资源监控器（低频：每30秒检查一次）
	resourceMonitor := resource.NewResourceMonitor(appConfig, monitorInstance)
	if err := resourceMonitor.Start(); err != nil {
		log.Printf("资源监控器启动失败: %v", err)
	}

	// 初始化HTTP服务器
	serverInstance := server.NewServer(database.Table, appConfig, monitorInstance, retentionManager, alertNotifier, resourceMonitor)

	// 初始化同步管理器（企业版功能）
	var syncManager *sync.SyncManager
	if appConfig.LicenseType == "enterprise" && appConfig.EnableDataSync {
		syncManager = sync.NewSyncManager(appConfig, monitorInstance)
		if err := syncManager.Start(); err != nil {
			log.Printf("同步管理器启动失败: %v", err)
		} else {
			log.Println("同步管理器已启动（企业版功能）")
		}
	}

	// 初始化MQTT客户端
	var mqttClient *mqtt.Client
	mqttClient, err = mqtt.NewClient(appConfig, dataQueue, monitorInstance, analyzerInstance, serverInstance)
	if err != nil {
		log.Printf("MQTT客户端初始化失败: %v", err)
		// 即使MQTT初始化失败，也继续启动HTTP服务器
	} else {
		if err := mqttClient.Subscribe(); err != nil {
			log.Printf("MQTT订阅失败: %v", err)
			// 即使MQTT订阅失败，也继续启动HTTP服务器
		}
	}

	return &Components{
		Monitor:         monitorInstance,
		AlertNotifier:   alertNotifier,
		Analyzer:        analyzerInstance,
		Agent:           agentInstance,
		Retention:       retentionManager,
		ResourceMonitor: resourceMonitor,
		SyncManager:     syncManager,
		Server:          serverInstance,
		MQTTClient:      mqttClient,
		DataQueue:       dataQueue,
	}, nil
}

func startServices(c *Components, appConfig *config.Config, appLicense *config.License) {
	// 打印欢迎信息
	cli.PrintWelcome()

	// 显示许可证信息
	config.PrintLicenseInfo(appLicense)

	// 显示当前配置摘要
	log.Printf("MQTT Broker: %s", appConfig.MQTTBroker)
	log.Printf("MQTT Topic: %s", appConfig.MQTTTopic)
	log.Printf("HTTP端口: %s", appConfig.HTTPPort)
	log.Println("sfsDb EdgeX adapter started successfully")

	// 启动队列处理
	c.DataQueue.ProcessQueue(func(data interface{}) error {
		records, ok := data.([]*map[string]any)
		if !ok {
			return fmt.Errorf("队列数据类型错误")
		}
		return database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
	})

	// 启动Agent
	if c.Agent != nil {
		if err := c.Agent.Start(); err != nil {
			log.Printf("Agent启动失败: %v", err)
		}
	}

	// 启动HTTP服务器
	if err := c.Server.Start(); err != nil {
		log.Fatalf("HTTP服务器启动失败: %v", err)
	}
}

func waitForShutdown(c *Components) {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	// 停止Agent
	if c.Agent != nil {
		c.Agent.Stop()
	}

	// 停止数据保留策略管理器
	if c.Retention != nil {
		c.Retention.Stop()
	}

	// 停止告警通知器
	if c.AlertNotifier != nil {
		c.AlertNotifier.Stop()
	}

	// 停止资源监控器
	if c.ResourceMonitor != nil {
		c.ResourceMonitor.Stop()
	}

	// 停止同步管理器（企业版功能）
	if c.SyncManager != nil {
		c.SyncManager.Stop()
	}

	// 看门狗已禁用
	// if c.Watchdog != nil {
	// 	c.Watchdog.Stop()
	// }

	// 断开MQTT连接
	if c.MQTTClient != nil {
		c.MQTTClient.Disconnect()
	}

	// 给服务器5秒时间完成正在处理的请求
	time.Sleep(5 * time.Second)

	log.Println("服务已退出")
}
