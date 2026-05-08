package mqtt

import (
	"log"
	"strconv"
	"sync"
	"time"

	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/database"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/pool"
)

type BatchWriter struct {
	writePool *pool.Pool         // 写入任务池
	monitor   *monitor.Monitor   // 监控器
	analyzer  *analyzer.Analyzer // 分析器

	mu             sync.Mutex             // 互斥锁，保护 pendingRecords 访问安全
	pendingRecords []*map[string]any      // 待写入记录缓冲区
	lastBatchTime  time.Time              // 上次写入时间
	stopChan       chan struct{}          // 用于停止 flushLoop() 定时写入协程。
	broadcastChan  chan *BroadcastMessage // 广播通道：设备数据 + 告警 + 其他信息。所有信息的通道。
}

func NewBatchWriter(monitor *monitor.Monitor, analyzer *analyzer.Analyzer) (*BatchWriter, error) {
	w := &BatchWriter{
		writePool:      pool.NewPoolForIO(),
		monitor:        monitor,
		analyzer:       analyzer,
		pendingRecords: make([]*map[string]any, 0, batchSize),
		lastBatchTime:  time.Now(),
		stopChan:       make(chan struct{}),
		broadcastChan:  make(chan *BroadcastMessage, 64),
	}
	go w.flushLoop()
	return w, nil
}

// GetBroadcastChan 返回告警广播通道
func (w *BatchWriter) GetBroadcastChan() <-chan *BroadcastMessage {
	return w.broadcastChan
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
		case <-w.stopChan: // ← 收到关闭信号，退出协程。程序退出时需要通知它停止，否则会变成 goroutine 泄漏
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
	// 交换缓冲区，避免在写入过程中被修改
	w.pendingRecords = make([]*map[string]any, 0, cap(records))
	w.lastBatchTime = time.Now()

	w.writePool.Submit(func() {
		w.doWrite(records)
	})
}

// 写入数据，所有数据在这里处理。
/*
doWrite() 分发到两个出口：
         ├── 1. database.BatchInsertNoInc()  → 写入数据库
         └── 2. analyzeData()                → 数据分析，告警推入 BroadcastChan
*/
func (w *BatchWriter) doWrite(records []*map[string]any) {
	if len(records) == 0 {
		return
	}

	// 写入数据库 Insert 批量插入记录
	insertedCount, err := database.Insert(database.Table, records)
	if err != nil {
		w.monitor.RecordError("db_write_failed", err.Error())
		return
	}
	// 检查是否所有记录都成功插入
	if insertedCount != len(records) {
		w.monitor.RecordError("partial_write", "inserted "+strconv.Itoa(insertedCount)+" of "+strconv.Itoa(len(records)))
	}
	w.monitor.IncrementTotalRecordsStored(int64(insertedCount))    // 总存储记录数 +n
	w.monitor.IncrementMQTTMessagesProcessed(int64(insertedCount)) // MQTT 处理消息数 +n

	// 广播实时数据到 Web 端
	NewBroadcastMessage(w.broadcastChan, "device_data", map[string]any{"records": records})

	// 数据分析，仅当记录数小于等于 50 时才分析
	if w.analyzer != nil && w.analyzer.IsEnabled() && len(records) <= 50 {
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
