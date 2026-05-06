package retention

import (
	"log"
	"time"

	"sfsEdgeStore/config"

	"github.com/liaoran123/sfsDb/engine"
)

// 管理数据的保留策略
type RetentionManager struct {
	table     *engine.Table
	config    *config.Config
	stopChan  chan struct{}
	isRunning bool
}

func NewRetentionManager(table *engine.Table, cfg *config.Config) *RetentionManager {
	return &RetentionManager{
		table:    table,
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

func (rm *RetentionManager) Start() error {
	if !rm.config.EnableRetentionPolicy {
		log.Println("Retention policy is disabled")
		return nil
	}

	if rm.isRunning {
		log.Println("Retention manager is already running")
		return nil
	}

	rm.isRunning = true
	go rm.cleanupLoop()
	log.Printf("Retention manager started with retention days: %d", rm.config.RetentionDays)
	return nil
}

func (rm *RetentionManager) Stop() {
	if !rm.isRunning {
		return
	}
	close(rm.stopChan)
	rm.isRunning = false
	log.Println("Retention manager stopped")
}

func (rm *RetentionManager) cleanupLoop() {
	interval := 24 * time.Hour // 默认 24 小时检查一次
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Starting scheduled data cleanup")
			deleted, err := rm.CleanupOldData()
			if err != nil {
				log.Printf("Data cleanup failed: %v", err)
			} else {
				log.Printf("Data cleanup completed, deleted %d records", deleted)
			}
		case <-rm.stopChan:
			return
		}
	}
}

func (rm *RetentionManager) CleanupOldData() (int, error) {
	if rm.table == nil {
		return 0, nil
	}

	retentionDuration := time.Duration(rm.config.RetentionDays) * 24 * time.Hour
	cutoffTime := time.Now().Add(-retentionDuration)
	cutoffTimestamp := cutoffTime.UnixNano()

	log.Printf("Cleaning up data older than %v (timestamp: %d)", cutoffTime, cutoffTimestamp)

	totalDeleted := 0
	batchSize := rm.config.CleanupBatchSize

	for {
		deleted, err := rm.cleanupBatch(cutoffTimestamp, batchSize)
		if err != nil {
			return totalDeleted, err
		}
		if deleted == 0 {
			break
		}
		totalDeleted += deleted
		log.Printf("Cleaned up %d records in this batch", deleted)
	}

	return totalDeleted, nil
}

func (rm *RetentionManager) cleanupBatch(cutoffTimestamp int64, batchSize int) (int, error) {
	startRange := make(map[string]any)
	endRange := make(map[string]any)
	startRange["timestamp"] = nil
	endRange["timestamp"] = cutoffTimestamp

	iter, err := rm.table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		return 0, err
	}
	defer iter.Release()

	records := iter.GetRecords(true, batchSize)
	defer records.Release()
	if len(records) == 0 {
		return 0, nil
	}

	err = iter.Delete()
	if err != nil {
		return 0, err
	}

	return batchSize, nil
}

func (rm *RetentionManager) GetRetentionStatus() map[string]any {
	// 从配置管理器获取最新的配置
	latestConfig := config.GetConfigManager().GetConfig()
	if latestConfig == nil {
		latestConfig = rm.config
	}

	return map[string]any{
		"enabled":            latestConfig.EnableRetentionPolicy,
		"retention_days":     latestConfig.RetentionDays,
		"cleanup_batch_size": latestConfig.CleanupBatchSize,
		"is_running":         rm.isRunning,
	}
}
