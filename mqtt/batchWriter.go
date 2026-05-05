package mqtt

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/broadcast"
	"sfsEdgeStore/database"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/pool"
)

var flushCount atomic.Int64

// BatchWriter 批量写入层 - 只负责批量缓冲、并发写入
// LevelDB 已保证持久化，无需额外的 dataQueue
type BatchWriter struct {
	writePool   *pool.Pool
	monitor     *monitor.Monitor
	broadcaster broadcast.Broadcaster
	analyzer    *analyzer.Analyzer

	mu             sync.Mutex
	pendingRecords []*map[string]any
	lastBatchTime  time.Time
	stopChan       chan struct{}
}

func NewBatchWriter(monitor *monitor.Monitor, broadcaster broadcast.Broadcaster, analyzer *analyzer.Analyzer) (*BatchWriter, error) {
	if monitor == nil {
		return nil, fmt.Errorf("monitor cannot be nil")
	}
	if analyzer == nil {
		return nil, fmt.Errorf("analyzer cannot be nil")
	}

	w := &BatchWriter{
		writePool:      pool.NewPoolForIO(), // 并发写入任务池，固定协程数、复用、自带背压控制
		monitor:        monitor,
		broadcaster:    broadcaster,
		analyzer:       analyzer,
		pendingRecords: make([]*map[string]any, 0, batchSize),
		lastBatchTime:  time.Now(),
		stopChan:       make(chan struct{}),
	}
	go w.flushLoop()
	return w, nil
}

// 定时器写入循环
// 当缓冲区大小超过 batchSize 时触发立即写入，定时写入由 flushLoop 协程负责。2个并发写入任务
func (w *BatchWriter) flushLoop() {
	ticker := time.NewTicker(batchTime * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.mu.Lock()
			if len(w.pendingRecords) > 0 {
				w.flush()
			}
			w.mu.Unlock()
		case <-w.stopChan:
			return
		}
	}
}

// Add 将记录添加到待写入缓冲区
// 当缓冲区大小超过 batchSize 时触发立即写入，定时写入由 flushLoop 协程负责
func (w *BatchWriter) Add(records []*map[string]any) {
	w.mu.Lock()
	w.pendingRecords = append(w.pendingRecords, records...)
	if len(w.pendingRecords) >= batchSize {
		w.flush()
	}
	w.mu.Unlock()
}

// 写入缓冲区数据
func (w *BatchWriter) flush() {
	records := w.pendingRecords
	w.pendingRecords = make([]*map[string]any, 0, cap(records))
	w.lastBatchTime = time.Now()

	w.writePool.Submit(func() {
		w.doWrite(records)
	})
}

// 写入数据，所有数据在这里处理。
/*
doWrite() 分发到三个出口：
         ├── 1. database.BatchInsertNoInc()  → 写入数据库
         ├── 2. BroadcastData()                  → 推送到 WebSocket
         └── 3. analyzeData()                    → 数据分析
*/
func (w *BatchWriter) doWrite(records []*map[string]any) {
	if len(records) == 0 {
		return
	}

	// 写入数据库 Insert 批量插入记录
	insertedCount, err := database.Insert(database.Table, records)
	if err != nil {
		log.Printf("Database write failed: %v", err)
		return
	}
	// 检查是否所有记录都成功插入
	if insertedCount != len(records) {
		log.Printf("Database write partial: inserted %d records, expected %d", insertedCount, len(records))
	}
	w.monitor.IncrementTotalRecordsStored(int64(insertedCount))    // 总存储记录数 +n
	w.monitor.IncrementMQTTMessagesProcessed(int64(insertedCount)) // MQTT 处理消息数 +n

	if w.broadcaster != nil {
		// 广播数据到所有 WebSocket 连接的 Web 客户端
		w.BroadcastData("device_data", map[string]any{
			"records": records,
		})
	}
	// 数据分析，仅当记录数小于等于 50 时才分析
	if w.analyzer != nil && w.analyzer.IsEnabled() && len(records) <= 50 {
		// 数据分析后，广播告警到所有 WebSocket 连接的 Web 客户端
		w.analyzeData(records)
	}
}

// Stop 停止批量写入层，等待所有待写入记录处理完成
func (w *BatchWriter) Stop() {
	close(w.stopChan)
	w.mu.Lock()
	if len(w.pendingRecords) > 0 {
		w.flush()
	}
	w.mu.Unlock()
	w.writePool.Stop()
	log.Println("BatchWriter stopped")
}
