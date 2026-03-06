package retention

import (
	"log"
	"time"

	"sfsEdgeStore/config"

	"github.com/liaoran123/sfsDb/engine"
)

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
	log.Printf("Retention manager started with retention days: %d, cleanup interval: %d hours", rm.config.RetentionDays, rm.config.CleanupInterval)
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
	interval := time.Duration(rm.config.CleanupInterval) * time.Hour
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
	return map[string]any{
		"enabled":            rm.config.EnableRetentionPolicy,
		"retention_days":     rm.config.RetentionDays,
		"cleanup_interval":   rm.config.CleanupInterval,
		"cleanup_batch_size": rm.config.CleanupBatchSize,
		"is_running":         rm.isRunning,
	}
}
