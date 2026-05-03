package mqtt

import (
	"encoding/json"
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

// 写入数据
/*
doWrite() 分发到三个出口：
         ├── 1. database.BatchInsertNoInc()  → 写入数据库
         ├── 2. broadcastData()                  → 推送到 WebSocket
         └── 3. analyzeData()                    → 数据分析
*/
func (w *BatchWriter) doWrite(records []*map[string]any) {
	if len(records) == 0 {
		return
	}

	w.monitor.IncrementDatabaseOperations()                       // 数据库操作次数 +1
	w.monitor.IncrementTotalRecordsStored(int64(len(records)))    // 总存储记录数 +n
	w.monitor.IncrementDataStoredBytes(int64(len(records) * 100)) // 存储数据字节数 +n*100
	// 批量写入 LevelDB数据库
	_, err := database.Table.BatchInsertNoInc(records)
	if err != nil {
		w.monitor.RecordError("storage_error", err.Error())
		log.Printf("Database write failed: %v", err)
		return
	}

	w.monitor.IncrementMQTTMessagesProcessed() // MQTT 处理消息数 +1
	/*
		每 50 次 flush 调用 debug.FreeOSMemory() 会触发全局 GC STW（Stop The World），影响实时性。
			count := flushCount.Add(1)
			if count%50 == 0 { // 内存优化（每 50 次 flush）
				debug.FreeOSMemory() // 释放未使用的内存回操作系统，防止边缘设备内存泄漏
			}
	*/
	if w.broadcaster != nil {
		// 广播数据给所有 WebSocket 连接的 Web 客户端
		w.broadcastData("device_data", map[string]any{
			"deviceName": (*records[0])["deviceName"],
			"records":    records,
		})
	}
	// 分析数据
	if w.analyzer.IsEnabled() && len(records) <= 50 { // 仅分析 50 条记录
		w.analyzeData(records, (*records[0])["deviceName"].(string))
	}
}

// 接口是将数据广播给所有 WebSocket 连接的 Web 客户端
func (w *BatchWriter) broadcastData(dataType string, data any) {
	if w.broadcaster == nil {
		return
	}

	broadcastData := map[string]any{
		"type":      dataType,
		"data":      data,
		"timestamp": time.Now().UnixNano(),
	}

	jsonData, err := json.Marshal(broadcastData)
	if err != nil {
		log.Printf("Failed to marshal broadcast data: %v", err)
		return
	}

	w.broadcaster.Broadcast(jsonData)
}

func (w *BatchWriter) analyzeData(records []*map[string]any, deviceName string) {
	if !w.analyzer.IsEnabled() {
		return
	}

	readingDataMap := make(map[string][]map[string]any, len(records))
	for _, record := range records {
		readingName, ok := (*record)["reading"].(string)
		if !ok {
			continue
		}
		readingDataMap[readingName] = append(readingDataMap[readingName], *record)
	}

	for readingName, analysisData := range readingDataMap {
		_, alerts := w.analyzer.Analyze(analysisData, deviceName, readingName)

		if len(alerts) > 0 {
			for _, alert := range alerts {
				w.monitor.RecordError(alert.AlertType, alert.Message)
			}

			w.broadcastData("alerts", map[string]any{
				"deviceName": deviceName,
				"alerts":     alerts,
			})
		}
	}
}

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
