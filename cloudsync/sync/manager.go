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
	"sfsEdgeStore/monitor"
)

// SyncManager 同步管理器
type SyncManager struct {
	config       *config.Config
	monitor      *monitor.Monitor
	uploader     Uploader
	deadLetter   *DeadLetterQueue
	lastSyncID   string
	mutex        sync.RWMutex
	running      bool
	tokenBucket  *TokenBucket
}

// NewSyncManager 创建同步管理器
func NewSyncManager(cfg *config.Config, monitor *monitor.Monitor) *SyncManager {
	manager := &SyncManager{
		config:       cfg,
		monitor:      monitor,
		deadLetter:   NewDeadLetterQueue(filepath.Join(cfg.DataSyncQueueDir, "dead_letter")),
		tokenBucket:  NewTokenBucket(50, 10), // 每秒50个令牌，初始10个
	}

	// 加载上次同步ID
	manager.loadLastSyncID()

	// 初始化上传器
	uploader, err := NewUploader(cfg)
	if err != nil {
		log.Printf("Failed to create uploader: %v", err)
		// 默认使用HTTP上传器
		uploader = NewHTTPUploader(cfg)
	}
	manager.uploader = uploader

	return manager
}

// Start 启动同步管理器
func (sm *SyncManager) Start() error {
	if sm.running {
		return nil
	}

	sm.running = true
	go sm.run()
	log.Println("Sync manager started")
	return nil
}

// Stop 停止同步管理器
func (sm *SyncManager) Stop() {
	sm.running = false
	log.Println("Sync manager stopped")
}

// run 运行同步管理器
func (sm *SyncManager) run() {
	ticker := time.NewTicker(time.Duration(sm.config.DataSyncInterval) * time.Second)
	defer ticker.Stop()

	for sm.running {
		select {
		case <-ticker.C:
			sm.syncData()
		}
	}
}

// syncData 同步数据
func (sm *SyncManager) syncData() {
	// 获取需要同步的数据
	records, err := sm.getPendingRecords()
	if err != nil {
		log.Printf("Failed to get pending records: %v", err)
		return
	}

	if len(records) == 0 {
		return
	}

	// 批量处理
	batches := sm.batchRecords(records, sm.config.DataSyncBatchSize)
	for _, batch := range batches {
		sm.processBatch(batch)
	}
}

// getPendingRecords 获取待同步的数据
func (sm *SyncManager) getPendingRecords() ([]map[string]interface{}, error) {
	// 根据上次同步ID查询待同步的数据
	// 这样可以实现真正的断点续传
	records, err := database.QueryRecords(database.Table, "", "", sm.lastSyncID, false)
	if err != nil {
		return nil, err
	}
	defer records.Release()

	result := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		result = append(result, record)
	}

	log.Printf("Found %d pending records to sync", len(result))
	return result, nil
}

// batchRecords 批量处理记录
func (sm *SyncManager) batchRecords(records []map[string]interface{}, batchSize int) [][]map[string]interface{} {
	var batches [][]map[string]interface{}
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batches = append(batches, records[i:end])
	}
	return batches
}

// processBatch 处理批次
func (sm *SyncManager) processBatch(batch []map[string]interface{}) {
	// 检查令牌桶
	if !sm.tokenBucket.TryTake(len(batch)) {
		log.Println("Rate limit exceeded, skipping batch")
		return
	}

	// 压缩数据
	compressedData, err := sm.compressData(batch)
	if err != nil {
		log.Printf("Failed to compress data: %v", err)
		return
	}

	// 上传数据
	err = sm.uploadWithRetry(compressedData)
	if err != nil {
		log.Printf("Failed to upload data: %v", err)
		// 处理失败的数据
		sm.handleFailedBatch(batch, err)
		return
	}

	// 更新同步状态
	sm.updateSyncStatus(batch)
	log.Printf("Synced %d records", len(batch))
}

// compressData 压缩数据
func (sm *SyncManager) compressData(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
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

// uploadWithRetry 带重试的上传
func (sm *SyncManager) uploadWithRetry(data []byte) error {
	maxRetries := sm.config.DataSyncMaxRetryCount
	for i := 0; i < maxRetries; i++ {
		err := sm.uploader.Upload(data)
		if err == nil {
			return nil
		}

		// 指数退避
		waitTime := time.Duration(1<<i) * time.Second
		log.Printf("Upload failed, retrying in %v...", waitTime)
		time.Sleep(waitTime)
	}

	return fmt.Errorf("failed to upload after %d retries", maxRetries)
}

// handleFailedBatch 处理失败的批次
func (sm *SyncManager) handleFailedBatch(batch []map[string]interface{}, err error) {
	for _, record := range batch {
		sm.deadLetter.Add(record, err.Error())
	}
	log.Printf("Added %d records to dead letter queue", len(batch))
}

// updateSyncStatus 更新同步状态
func (sm *SyncManager) updateSyncStatus(batch []map[string]interface{}) {
	// 这里简化处理，实际应该更新数据库中的状态
	if len(batch) > 0 {
		lastRecord := batch[len(batch)-1]
		if id, ok := lastRecord["id"].(string); ok {
			sm.lastSyncID = id
			sm.saveLastSyncID()
		}
	}
}

// loadLastSyncID 加载上次同步ID
func (sm *SyncManager) loadLastSyncID() {
	filePath := filepath.Join(sm.config.DataSyncQueueDir, "last_sync_id.txt")
	data, err := os.ReadFile(filePath)
	if err == nil {
		sm.lastSyncID = string(data)
	}
}

// saveLastSyncID 保存上次同步ID
func (sm *SyncManager) saveLastSyncID() {
	filePath := filepath.Join(sm.config.DataSyncQueueDir, "last_sync_id.txt")
	err := os.WriteFile(filePath, []byte(sm.lastSyncID), 0644)
	if err != nil {
		log.Printf("Failed to save last sync ID: %v", err)
	}
}

// GetSyncStatus 获取同步状态
func (sm *SyncManager) GetSyncStatus() map[string]interface{} {
	return map[string]interface{}{
		"last_sync_id": sm.lastSyncID,
		"dead_letter_count": sm.deadLetter.Count(),
		"running": sm.running,
	}
}