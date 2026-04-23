package database

import (
	"fmt"
	"log"
	"time"

	"github.com/liaoran123/sfsDb/engine"
)

// BatchInsertWithRetry 批量插入数据并带有重试机制，以防数据库暂时不可用。
func BatchInsertWithRetry(tbl *engine.Table, records []*map[string]any, maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		_, err := tbl.BatchInsertNoInc(records)
		if err == nil {
			return nil
		}

		log.Printf("Failed to batch insert data (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("failed to batch insert data after %d attempts", maxRetries)
}
