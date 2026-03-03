// sfsDb 与 EdgeX MQTT 适配器示例（改进版）
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sfsdb-edgex-adapter/backup"
	"sfsdb-edgex-adapter/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// EdgeX 消息结构（符合 MessageEnvelope 格式）
type EdgeXMessage struct {
	CorrelationID string          `json:"correlationId,omitempty"`
	MessageType   string          `json:"messageType,omitempty"`
	Origin        int64           `json:"origin,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

// EdgeX 事件结构
type EdgeXEvent struct {
	ID          string         `json:"id"`
	DeviceName  string         `json:"deviceName"`
	Readings    []EdgeXReading `json:"readings"`
	Origin      int64          `json:"origin"`
	ProfileName string         `json:"profileName,omitempty"`
	SourceName  string         `json:"sourceName,omitempty"`
}

// EdgeX 读数结构
type EdgeXReading struct {
	ID           string          `json:"id"`
	ResourceName string          `json:"resourceName"`
	Value        string          `json:"value"`
	Origin       int64           `json:"origin"`
	ProfileName  string          `json:"profileName,omitempty"`
	DeviceName   string          `json:"deviceName,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

var table *engine.Table
var appConfig *config.Config

func main() {
	// 加载配置
	var err error
	appConfig, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接 sfsDb
	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化 MQTT 客户端
	client, err := initMQTT()
	if err != nil {
		log.Fatalf("Failed to initialize MQTT: %v", err)
	}
	defer client.Disconnect(250)

	// 订阅 EdgeX 消息
	if err := subscribeToEdgeX(client); err != nil {
		log.Fatalf("Failed to subscribe to EdgeX messages: %v", err)
	}

	log.Println("sfsDb EdgeX adapter started successfully")

	// 启动 HTTP 服务器，提供健康检查接口
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// 添加数据查询API
	http.HandleFunc("/api/readings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 获取查询参数
		deviceName := r.URL.Query().Get("deviceName")
		startTime := r.URL.Query().Get("startTime")
		endTime := r.URL.Query().Get("endTime")

		// 查询数据
		readings, err := queryReadings(table, deviceName, startTime, endTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":    len(readings),
			"readings": readings,
		})
	})

	// 添加数据备份API
	http.HandleFunc("/api/backup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		// 获取备份路径参数
		backupPath := r.URL.Query().Get("path")
		if backupPath == "" {
			backupPath = "./backups"
		}

		// 创建备份管理器
		backupManager := backup.NewBackupManager(storage.KVDb)

		// 执行备份
		backupFile, err := backupManager.Backup(backupPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status":     "success",
			"backupFile": backupFile,
		})
	})

	// 添加数据恢复API
	http.HandleFunc("/api/restore", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		// 获取备份文件路径
		backupFile := r.URL.Query().Get("file")
		if backupFile == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Backup file path is required"})
			return
		}

		// 创建备份管理器
		backupManager := backup.NewBackupManager(storage.KVDb)

		// 执行恢复
		if err := backupManager.Restore(backupFile); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Database restored successfully",
		})
	})

	// 添加测试端点，用于模拟 EdgeX 消息
	http.HandleFunc("/api/test-edgex", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		// 模拟 EdgeX 消息
		edgexMsg := EdgeXMessage{
			CorrelationID: "test-correlation-id",
			MessageType:   "event",
			Origin:        time.Now().UnixNano(),
			Payload: json.RawMessage(`{
				"id": "test-event-id",
				"deviceName": "TestDevice-001",
				"readings": [
					{
						"id": "reading-1",
						"resourceName": "temperature",
						"value": "25.5",
						"origin": 1677721600000000000,
						"deviceName": "TestDevice-001"
					},
					{
						"id": "reading-2",
						"resourceName": "humidity",
						"value": "45",
						"origin": 1677721600000000000,
						"deviceName": "TestDevice-001"
					},
					{
						"id": "reading-3",
						"resourceName": "pressure",
						"value": "1013.25",
						"origin": 1677721600000000000,
						"deviceName": "TestDevice-001"
					}
				],
				"origin": 1677721600000000000
			}`),
		}

		// 解析 payload 中的事件
		var event EdgeXEvent
		if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// 收集所有读数，准备批量插入
		var records []*map[string]any

		// 处理每个读数
		for _, reading := range event.Readings {
			// 准备数据
			metadataStr := ""
			if reading.Metadata != nil {
				metadataStr = string(reading.Metadata)
			}

			// 将字符串值转换为浮点数
			value := 0.0
			fmt.Sscanf(reading.Value, "%f", &value)

			data := map[string]any{
				"id":         reading.ID,
				"deviceName": event.DeviceName,
				"reading":    reading.ResourceName,
				"value":      value,
				"timestamp":  int(reading.Origin / 1000000000), // 转换为秒，类型为 int
				"metadata":   metadataStr,
			}

			records = append(records, &data)
		}

		// 批量存储到 sfsDb
		if len(records) > 0 {
			_, err := table.BatchInsertNoInc(records)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			} else {
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "success",
					"message": fmt.Sprintf("Batch stored %d readings from %s", len(records), event.DeviceName),
				})
			}
		} else {
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "No readings to store",
			})
		}
	})

	// 在后台启动 HTTP 服务器
	go func() {
		log.Println("Starting HTTP server for health checks on port 8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down adapter...")

	// 给服务器 5 秒的时间来完成正在处理的请求
	time.Sleep(5 * time.Second)

	log.Println("Adapter exited")
}

// 初始化数据库
func initDatabase() error {
	// 确保数据库目录存在
	if err := os.MkdirAll(appConfig.DBPath, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// 打开数据库
	_, err := storage.GetDBManager().OpenDB(appConfig.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// 创建或获取表
	tableName := "edgex_readings"
	var createErr error
	table, createErr = engine.TableNew(tableName)
	if createErr != nil {
		return fmt.Errorf("failed to create table: %v", createErr)
	}

	// 设置表字段
	fields := map[string]any{
		"id":         "",
		"deviceName": "",
		"reading":    "",
		"value":      0.0,
		"timestamp":  0,
		"metadata":   "",
	}
	if err := table.SetFields(fields); err != nil {
		return fmt.Errorf("failed to set table fields: %v", err)
	}

	// 创建组合主键索引 (deviceName + timestamp)
	primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
	if err != nil {
		return fmt.Errorf("failed to create primary key: %v", err)
	}
	primaryKey.AddFields("deviceName", "timestamp") // 创建deviceName和timestamp的组合主键
	if err := table.CreateIndex(primaryKey); err != nil {
		// 忽略索引已存在的错误
		if err.Error() != "index already exists" {
			return fmt.Errorf("failed to create primary key index: %v", err)
		}
	}
	/*
		//由于组合主键中包括deviceName不定长类型，所以不能单独创建其他索引
		//如果业务确实需要使用单独索引，如按时间查询，则需要将组合主键更改为组合索引。时间戳为主键。

		//现在的实现需要按设备和时间戳进行查询的场景性能最高，资源最少。权衡之下，是最优解。

						// 创建设备名称索引
						deviceIndex, err := engine.DefaultNormalIndexNew("device_index")
						if err != nil {
							return fmt.Errorf("failed to create device index: %v", err)
						}
						deviceIndex.AddFields("deviceName")
						if err := table.CreateIndex(deviceIndex); err != nil {
							// 忽略索引已存在的错误
							if err.Error() != "index already exists" {
								return fmt.Errorf("failed to create device index: %v", err)
							}
						}

					// 创建时间戳索引
					timeIndex, err := engine.DefaultNormalIndexNew("time_index")
					if err != nil {
						return fmt.Errorf("failed to create time index: %v", err)
					}
					timeIndex.AddFields("timestamp")
					if err := table.CreateIndex(timeIndex); err != nil {
						// 忽略索引已存在的错误
						if err.Error() != "index already exists" {
							return fmt.Errorf("failed to create time index: %v", err)
						}
					}
	*/
	log.Println("Database initialized successfully")
	return nil
}

// 初始化 MQTT 客户端
func initMQTT() (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(appConfig.MQTTBroker)
	opts.SetClientID(appConfig.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Second * 30)
	opts.SetDefaultPublishHandler(messageHandler())

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	log.Printf("Connected to MQTT broker: %s", appConfig.MQTTBroker)
	return client, nil
}

// 订阅 EdgeX 消息
func subscribeToEdgeX(client mqtt.Client) error {
	token := client.Subscribe(appConfig.MQTTTopic, 1, nil)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %v", appConfig.MQTTTopic, token.Error())
	}

	log.Printf("Subscribed to topic: %s", appConfig.MQTTTopic)
	return nil
}

// 消息处理函数
func messageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message on topic: %s", msg.Topic())

		var edgexMsg EdgeXMessage
		if err := json.Unmarshal(msg.Payload(), &edgexMsg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			return
		}

		// 解析 payload 中的事件
		var event EdgeXEvent
		if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
			log.Printf("Failed to parse event: %v", err)
			return
		}

		// 收集所有读数，准备批量插入
		var records []*map[string]any

		// 处理每个读数
		for _, reading := range event.Readings {
			// 准备数据
			metadataStr := ""
			if reading.Metadata != nil {
				metadataStr = string(reading.Metadata)
			}

			// 将字符串值转换为浮点数
			value := 0.0
			fmt.Sscanf(reading.Value, "%f", &value)

			data := map[string]any{
				"id":         reading.ID,
				"deviceName": event.DeviceName,
				"reading":    reading.ResourceName,
				"value":      value,
				"timestamp":  int(reading.Origin / 1000000000), // 转换为秒，类型为 int
				"metadata":   metadataStr,
			}

			records = append(records, &data)
		}

		// 批量存储到 sfsDb
		if len(records) > 0 {
			_, err := table.BatchInsertNoInc(records)
			if err != nil {
				log.Printf("Failed to batch store data: %v", err)
			} else {
				log.Printf("Batch stored %d readings from %s", len(records), event.DeviceName)
			}
		}
	}
}

// 查询读数数据
func queryReadings(tbl *engine.Table, deviceName, startTime, endTime string) ([]map[string]any, error) {
	var readings []map[string]any

	log.Println("Querying readings with filters:")
	log.Printf("  deviceName: %s", deviceName)
	log.Printf("  startTime: %s", startTime)
	log.Printf("  endTime: %s", endTime)

	// 构建时间范围查询
	var startTimestamp, endTimestamp *int64

	// 解析开始时间
	if startTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			ts := start.Unix()
			startTimestamp = &ts
		}
	}

	// 解析结束时间
	if endTime != "" {
		end, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			ts := end.Unix()
			endTimestamp = &ts
		}
	}

	// 构建查询范围
	startRange := make(map[string]any)
	endRange := make(map[string]any)

	// 利用组合主键 (deviceName + timestamp) 进行更高效的查询
	// 设置设备名称
	startRange["deviceName"] = deviceName
	endRange["deviceName"] = deviceName

	// 设置时间范围
	if startTimestamp != nil {
		startRange["timestamp"] = *startTimestamp
	} else {
		startRange["timestamp"] = nil // 从最小值开始
	}

	if endTimestamp != nil {
		endRange["timestamp"] = *endTimestamp
	} else {
		endRange["timestamp"] = nil // 到最大值结束
	}

	// 执行范围查询
	iter, err := tbl.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		return nil, fmt.Errorf("failed to search readings: %v", err)
	}
	defer iter.Release()

	// 获取记录
	records := iter.GetRecords(true)
	defer records.Release()

	// 转换记录为map切片
	for _, record := range records {
		readings = append(readings, record)
	}

	log.Printf("Found %d readings", len(readings))
	return readings, nil
}
