package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/*
数据库写入错误时，将数据加入队列，等待后续重试。 可能不会用到。只是预留。
错误处理与重试（中优先级）
问题：批量写入失败仅记录日志，但没有重试或放到死信队列；若数据库暂不可用会丢失数据。
建议：实现重试策略、或把写入失败的记录写入本地队列并周期性重试；或者引入确认机制与持久化队列（如磁盘队列）以保证可靠性。
Queue 结构体实现了临时持久化功能，具体来说：

### 临时持久化的实现原理
1. 基于文件系统 ：队列使用文件系统作为存储介质，每个数据项对应一个 JSON 文件
2. 自动序列化/反序列化 ：数据被自动序列化为 JSON 格式并写入文件，读取时自动反序列化
3. 持久化保证 ：即使系统重启，队列中的数据也会保存在磁盘上，不会丢失
4. 顺序处理 ：数据按照先进先出的顺序处理，确保处理顺序的正确性
### 临时持久化的特点
1. 轻量级 ：不需要复杂的数据库配置，直接使用文件系统
2. 可靠性 ：数据写入磁盘后才返回成功，确保数据不会丢失
3. 简单易用 ：提供了简洁的 API，如 Enqueue 、 Dequeue 和 ProcessQueue
4. 自动重试 ：配合 ProcessQueue 函数，可以自动处理队列中的数据并在失败时重试
*/
// Queue 本地磁盘队列
type Queue struct {
	queueDir string
	mutex    sync.Mutex
}

// NewQueue 创建一个新的磁盘队列
func NewQueue(queueDir string) (*Queue, error) {
	// 确保队列目录存在
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %v", err)
	}

	return &Queue{
		queueDir: queueDir,
	}, nil
}

// Enqueue 将数据加入队列
func (q *Queue) Enqueue(data interface{}) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 生成唯一文件名
	filename := fmt.Sprintf("%d.json", time.Now().UnixNano())
	filepath := filepath.Join(q.queueDir, filename)

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write to queue: %v", err)
	}

	return nil
}

// Dequeue 从队列中取出数据
func (q *Queue) Dequeue() (interface{}, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 读取队列目录中的文件
	files, err := os.ReadDir(q.queueDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read queue directory: %v", err)
	}

	// 找到第一个文件
	var targetFile os.DirEntry
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return nil, nil // 队列为空
	}

	// 读取文件内容
	filepath := filepath.Join(q.queueDir, targetFile.Name())
	jsonData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read queue file: %v", err)
	}

	// 反序列化数据
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %v", err)
	}

	// 删除文件
	if err := os.Remove(filepath); err != nil {
		return nil, fmt.Errorf("failed to remove queue file: %v", err)
	}

	return data, nil
}

// Size 获取队列大小
func (q *Queue) Size() (int, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 读取队列目录中的文件
	files, err := os.ReadDir(q.queueDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read queue directory: %v", err)
	}

	// 统计 json 文件数量
	count := 0
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			count++
		}
	}

	return count, nil
}

/*
- 后台 goroutine 定期检查队列，尝试处理队列中的数据
- 当数据库恢复可用时，队列中的数据会被成功写入
*/
// ProcessQueue 处理队列中的数据
func (q *Queue) ProcessQueue(processFunc func(interface{}) error) {
	go func() {
		for {
			// 检查队列大小
			size, err := q.Size()
			if err != nil {
				log.Printf("Failed to get queue size: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// 如果队列为空，等待一段时间
			if size == 0 {
				time.Sleep(5 * time.Second)
				continue
			}

			// 处理队列中的数据
			data, err := q.Dequeue()
			if err != nil {
				log.Printf("Failed to dequeue data: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if data == nil {
				time.Sleep(5 * time.Second)
				continue
			}

			// 调用处理函数
			if err := processFunc(data); err != nil {
				log.Printf("Failed to process queue data: %v", err)
				// 将数据重新加入队列
				if err := q.Enqueue(data); err != nil {
					log.Printf("Failed to re-enqueue data: %v", err)
				}
				time.Sleep(5 * time.Second)
			}
		}
	}()
}
