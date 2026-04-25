package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// DeadLetterQueue 死信队列
type DeadLetterQueue struct {
	path  string
	items []DeadLetterItem
	mutex sync.RWMutex
}

// DeadLetterItem 死信队列项
type DeadLetterItem struct {
	Data  map[string]interface{} `json:"data"`
	Error string                 `json:"error"`
}

// NewDeadLetterQueue 创建死信队列
func NewDeadLetterQueue(path string) *DeadLetterQueue {
	queue := &DeadLetterQueue{
		path: path,
	}

	// 确保目录存在
	os.MkdirAll(path, 0755)

	// 加载现有数据
	queue.load()

	return queue
}

// Add 添加项到死信队列
func (q *DeadLetterQueue) Add(data map[string]interface{}, err string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	item := DeadLetterItem{
		Data:  data,
		Error: err,
	}

	q.items = append(q.items, item)
	q.save()
}

// Count 获取队列长度
func (q *DeadLetterQueue) Count() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.items)
}

// Get 获取所有项
func (q *DeadLetterQueue) Get() []DeadLetterItem {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.items
}

// Clear 清空队列
func (q *DeadLetterQueue) Clear() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.items = []DeadLetterItem{}
	q.save()
}

// save 保存队列
func (q *DeadLetterQueue) save() {
	filePath := filepath.Join(q.path, "dead_letter.json")
	data, err := json.Marshal(q.items)
	if err != nil {
		return
	}

	os.WriteFile(filePath, data, 0644)
}

// load 加载队列
func (q *DeadLetterQueue) load() {
	filePath := filepath.Join(q.path, "dead_letter.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	json.Unmarshal(data, &q.items)
}