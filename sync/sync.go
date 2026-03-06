package sync

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"sfsEdgeStore/config"
	"sfsEdgeStore/database"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SyncRecord 同步记录
type SyncRecord struct {
	ID         string    `json:"id"`
	Data       any       `json:"data"`
	Timestamp  time.Time `json:"timestamp"`
	RetryCount int       `json:"retry_count"`
	LastError  string    `json:"last_error,omitempty"`
}

// SyncStatus 同步状态
type SyncStatus struct {
	IsRunning      bool   `json:"is_running"`
	QueueSize      int    `json:"queue_size"`
	LastSyncTime   int64  `json:"last_sync_time"`
	TotalSynced    int64  `json:"total_synced"`
	TotalFailed    int64  `json:"total_failed"`
	CurrentBatch   int    `json:"current_batch"`
}

// SyncManager 数据同步管理器
type SyncManager struct {
	config       *config.Config
	mqttClient   mqtt.Client
	syncQueue    *syncQueue
	stopChan     chan struct{}
	isRunning    bool
	mutex        sync.Mutex
	totalSynced  int64
	totalFailed  int64
	lastSyncTime int64
}

// syncQueue 同步队列（基于磁盘）
type syncQueue struct {
	queueDir string
	mutex    sync.Mutex
}

// NewSyncManager 创建数据同步管理器
func NewSyncManager(cfg *config.Config) (*SyncManager, error) {
	sq, err := newSyncQueue(cfg.DataSyncQueueDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync queue: %v", err)
	}

	return &SyncManager{
		config:    cfg,
		syncQueue: sq,
		stopChan:  make(chan struct{}),
	}, nil
}

// newSyncQueue 创建同步队列
func newSyncQueue(queueDir string) (*syncQueue, error) {
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %v", err)
	}

	return &syncQueue{
		queueDir: queueDir,
	}, nil
}

// Start 启动同步管理器
func (sm *SyncManager) Start() error {
	if !sm.config.EnableDataSync {
		log.Println("Data sync is disabled")
		return nil
	}

	if sm.isRunning {
		log.Println("Sync manager is already running")
		return nil
	}

	if err := sm.initMQTTClient(); err != nil {
		return fmt.Errorf("failed to initialize MQTT client: %v", err)
	}

	sm.isRunning = true
	go sm.syncLoop()
	log.Println("Data sync manager started")
	return nil
}

// Stop 停止同步管理器
func (sm *SyncManager) Stop() {
	if !sm.isRunning {
		return
	}
	close(sm.stopChan)
	if sm.mqttClient != nil && sm.mqttClient.IsConnected() {
		sm.mqttClient.Disconnect(250)
	}
	sm.isRunning = false
	log.Println("Data sync manager stopped")
}

// EnqueueData 将数据加入同步队列
func (sm *SyncManager) EnqueueData(data any) error {
	if !sm.config.EnableDataSync {
		return nil
	}

	record := &SyncRecord{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Data:       data,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	return sm.syncQueue.enqueue(record)
}

// GetStatus 获取同步状态
func (sm *SyncManager) GetStatus() SyncStatus {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	queueSize, _ := sm.syncQueue.size()

	return SyncStatus{
		IsRunning:    sm.isRunning,
		QueueSize:    queueSize,
		LastSyncTime: sm.lastSyncTime,
		TotalSynced:  sm.totalSynced,
		TotalFailed:  sm.totalFailed,
	}
}

// initMQTTClient 初始化 MQTT 客户端
func (sm *SyncManager) initMQTTClient() error {
	if sm.config.MQTTBroker == "" {
		return fmt.Errorf("MQTT broker not configured")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(sm.config.MQTTBroker)
	opts.SetClientID(sm.config.ClientID + "-sync-manager")
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Minute)

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Sync manager connected to MQTT broker")
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Sync manager MQTT connection lost: %v", err)
	})

	sm.mqttClient = mqtt.NewClient(opts)
	token := sm.mqttClient.Connect()
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	return nil
}

// syncLoop 同步循环
func (sm *SyncManager) syncLoop() {
	interval := time.Duration(sm.config.DataSyncInterval) * time.Second

	for {
		select {
		case <-sm.stopChan:
			return
		case <-time.After(interval):
			sm.processBatch()
		}
	}
}

// processBatch 处理一批数据
func (sm *SyncManager) processBatch() {
	records, err := sm.syncQueue.dequeueBatch(sm.config.DataSyncBatchSize)
	if err != nil {
		log.Printf("Failed to dequeue sync records: %v", err)
		return
	}

	if len(records) == 0 {
		return
	}

	if err := sm.sendToMQTT(records); err != nil {
		log.Printf("Failed to send sync batch: %v", err)
		sm.handleFailure(records, err.Error())
	} else {
		log.Printf("Successfully synced %d records", len(records))
		sm.mutex.Lock()
		sm.totalSynced += int64(len(records))
		sm.lastSyncTime = time.Now().Unix()
		sm.mutex.Unlock()
	}
}

// sendToMQTT 发送数据到 MQTT
func (sm *SyncManager) sendToMQTT(records []*SyncRecord) error {
	if sm.mqttClient == nil || !sm.mqttClient.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	payload, err := sm.compressRecords(records)
	if err != nil {
		return fmt.Errorf("failed to compress records: %v", err)
	}

	token := sm.mqttClient.Publish(sm.config.DataSyncMQTTTopic, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	return nil
}

// compressRecords 压缩同步记录
func (sm *SyncManager) compressRecords(records []*SyncRecord) ([]byte, error) {
	jsonData, err := json.Marshal(records)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gzw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// handleFailure 处理失败的记录
func (sm *SyncManager) handleFailure(records []*SyncRecord, errorMsg string) {
	sm.mutex.Lock()
	sm.totalFailed += int64(len(records))
	sm.mutex.Unlock()

	for _, record := range records {
		record.RetryCount++
		record.LastError = errorMsg

		if record.RetryCount > sm.config.DataSyncMaxRetryCount {
			log.Printf("Record %s exceeded max retries, dropping", record.ID)
			continue
		}

		if err := sm.syncQueue.enqueue(record); err != nil {
			log.Printf("Failed to re-enqueue record %s: %v", record.ID, err)
		}
	}
}

// SyncFromDatabase 从数据库同步数据到中心系统
func (sm *SyncManager) SyncFromDatabase(startTimestamp int64, endTimestamp int64) error {
	if !sm.config.EnableDataSync {
		return fmt.Errorf("data sync is disabled")
	}

	if database.Table == nil {
		return fmt.Errorf("database table not initialized")
	}

	startRange := map[string]any{}
	endRange := map[string]any{}
	if startTimestamp > 0 {
		startRange["timestamp"] = startTimestamp
	}
	if endTimestamp > 0 {
		endRange["timestamp"] = endTimestamp
	}

	iter, err := database.Table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		return fmt.Errorf("failed to search database: %v", err)
	}
	defer iter.Release()

	batchSize := sm.config.DataSyncBatchSize
	for {
		records := iter.GetRecords(true, batchSize)
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			if err := sm.EnqueueData(record); err != nil {
				log.Printf("Failed to enqueue record for sync: %v", err)
			}
		}

		records.Release()
	}

	log.Println("Database sync completed")
	return nil
}

// enqueue 入队操作
func (sq *syncQueue) enqueue(record *SyncRecord) error {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	filename := fmt.Sprintf("%s.json", record.ID)
	filepath := filepath.Join(sq.queueDir, filename)

	jsonData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %v", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write record: %v", err)
	}

	return nil
}

// dequeueBatch 批量出队
func (sq *syncQueue) dequeueBatch(batchSize int) ([]*SyncRecord, error) {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	files, err := os.ReadDir(sq.queueDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read queue directory: %v", err)
	}

	var records []*SyncRecord
	count := 0

	for _, file := range files {
		if count >= batchSize {
			break
		}

		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filepath := filepath.Join(sq.queueDir, file.Name())
		jsonData, err := os.ReadFile(filepath)
		if err != nil {
			log.Printf("Failed to read record file %s: %v", file.Name(), err)
			continue
		}

		var record SyncRecord
		if err := json.Unmarshal(jsonData, &record); err != nil {
			log.Printf("Failed to unmarshal record %s: %v", file.Name(), err)
			os.Remove(filepath)
			continue
		}

		records = append(records, &record)
		os.Remove(filepath)
		count++
	}

	return records, nil
}

// size 获取队列大小
func (sq *syncQueue) size() (int, error) {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	files, err := os.ReadDir(sq.queueDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read queue directory: %v", err)
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			count++
		}
	}

	return count, nil
}
