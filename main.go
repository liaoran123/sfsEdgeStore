// sfsDb 与 EdgeX MQTT 适配器示例（改进版）
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sfsdb-edgex-adapter/config"
	"sfsdb-edgex-adapter/database"
	"sfsdb-edgex-adapter/mqtt"
	"sfsdb-edgex-adapter/queue"
	"sfsdb-edgex-adapter/server"
)

var appConfig *config.Config
var dataQueue *queue.Queue

func main() {
	// 加载配置
	var err error
	appConfig, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接 sfsDb
	if err = database.Init(appConfig.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化数据队列，数据库添加失败的数据会被缓存到队列中重新尝试添加（如在断电后重启）
	dataQueue, err = queue.NewQueue("./data_queue")
	if err != nil {
		log.Fatalf("Failed to initialize data queue: %v", err)
	}

	// 创建MQTT客户端，通过 mqtt.NewClient 连接到配置的MQTT broker
	mqttClient, err := mqtt.NewClient(appConfig, dataQueue)
	if err != nil {
		log.Fatalf("Failed to initialize MQTT: %v", err)
	}
	defer mqttClient.Disconnect()

	// 订阅EdgeX的MQTT主题，订阅 edgex/events/core/# 主题，接收所有核心服务的事件
	if err := mqttClient.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to EdgeX messages: %v", err)
	}

	log.Println("sfsDb EdgeX adapter started successfully")

	// 启动队列处理 goroutine，处理可能存在添加失败等异常数据
	/*
	   启动队列处理 goroutine，用于处理数据队列中的数据
	   1. 从队列中取出数据
	   2. 转换为 []*map[string]any 类型
	   3. 使用重试机制批量插入到数据库中（默认重试3次，每次间隔2秒）
	   4. 如果插入失败，将数据重新加入队列，等待后续重试
	*/
	dataQueue.ProcessQueue(func(data interface{}) error {
		records, ok := data.([]*map[string]any)
		if !ok {
			return fmt.Errorf("invalid data type in queue")
		}
		// 使用重试机制插入数据
		return database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
	})

	// 启动 HTTP 服务器，提供查询接口
	serverInstance := server.NewServer(database.Table, appConfig)
	if err := serverInstance.Start(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)                      // 创建一个信号通道，用于接收中断信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 注册信号通知，当收到SIGINT或SIGTERM时，将信号发送到quit通道
	<-quit                                               // 阻塞直到收到中断信号。
	log.Println("Shutting down adapter...")
	// 给服务器 5 秒的时间来完成正在处理的请求
	time.Sleep(5 * time.Second)

	log.Println("Adapter exited")
}
