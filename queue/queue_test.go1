package queue

import (
	"os"
	"testing"
	"time"
)

// TestQueueCreation 测试队列的创建
func TestQueueCreation(t *testing.T) {
	// 创建临时目录作为队列目录
	queueDir := "./test_queue"
	defer os.RemoveAll(queueDir)

	// 创建队列
	q, err := NewQueue(queueDir)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	// 检查队列目录是否创建成功
	if _, err := os.Stat(queueDir); os.IsNotExist(err) {
		t.Fatalf("Queue directory was not created: %v", err)
	}

	// 测试队列大小
	size, err := q.Size()
	if err != nil {
		t.Fatalf("Failed to get queue size: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected queue size 0, got %d", size)
	}
}

// TestEnqueueAndDequeue 测试数据的入队和出队
func TestEnqueueAndDequeue(t *testing.T) {
	// 创建临时目录作为队列目录
	queueDir := "./test_queue"
	defer os.RemoveAll(queueDir)

	// 创建队列
	q, err := NewQueue(queueDir)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	// 测试数据
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// 入队
	if err := q.Enqueue(testData); err != nil {
		t.Fatalf("Failed to enqueue data: %v", err)
	}

	// 检查队列大小
	size, err := q.Size()
	if err != nil {
		t.Fatalf("Failed to get queue size: %v", err)
	}

	if size != 1 {
		t.Errorf("Expected queue size 1, got %d", size)
	}

	// 出队
	dequeuedData, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Failed to dequeue data: %v", err)
	}

	// 检查出队的数据是否正确
	dequeuedMap, ok := dequeuedData.(map[string]any)
	if !ok {
		t.Fatalf("Expected map[string]any, got %T", dequeuedData)
	}

	if dequeuedMap["key1"] != "value1" || dequeuedMap["key2"] != "value2" {
		t.Errorf("Dequeued data does not match enqueued data")
	}

	// 检查队列大小
	size, err = q.Size()
	if err != nil {
		t.Fatalf("Failed to get queue size: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected queue size 0 after dequeue, got %d", size)
	}
}

// TestProcessQueue 测试队列的处理功能
func TestProcessQueue(t *testing.T) {
	// 创建临时目录作为队列目录
	queueDir := "./test_queue"
	defer os.RemoveAll(queueDir)

	// 创建队列
	q, err := NewQueue(queueDir)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	// 测试数据
	testData := "test data"

	// 入队
	if err := q.Enqueue(testData); err != nil {
		t.Fatalf("Failed to enqueue data: %v", err)
	}

	// 处理结果
	var processedData string
	processComplete := false

	// 启动队列处理
	q.ProcessQueue(func(data interface{}) error {
		processedData = data.(string)
		processComplete = true
		return nil
	})

	// 等待处理完成
	time.Sleep(1 * time.Second)

	// 检查处理结果
	if !processComplete {
		t.Fatalf("Queue processing did not complete")
	}

	if processedData != testData {
		t.Errorf("Processed data does not match enqueued data: expected %s, got %s", testData, processedData)
	}

	// 检查队列大小
	size, err := q.Size()
	if err != nil {
		t.Fatalf("Failed to get queue size: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected queue size 0 after processing, got %d", size)
	}
}

// TestQueuePersistence 测试队列的持久化功能
func TestQueuePersistence(t *testing.T) {
	// 创建临时目录作为队列目录
	queueDir := "./test_queue"
	defer os.RemoveAll(queueDir)

	// 创建队列
	q1, err := NewQueue(queueDir)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	// 测试数据
	testData := "persistent data"

	// 入队
	if err := q1.Enqueue(testData); err != nil {
		t.Fatalf("Failed to enqueue data: %v", err)
	}

	// 检查队列大小
	size, err := q1.Size()
	if err != nil {
		t.Fatalf("Failed to get queue size: %v", err)
	}

	if size != 1 {
		t.Errorf("Expected queue size 1, got %d", size)
	}

	// 创建新的队列实例，使用相同的目录
	q2, err := NewQueue(queueDir)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	// 检查队列大小
	size, err = q2.Size()
	if err != nil {
		t.Fatalf("Failed to get queue size: %v", err)
	}

	if size != 1 {
		t.Errorf("Expected queue size 1 after re-creating queue, got %d", size)
	}

	// 出队
	dequeuedData, err := q2.Dequeue()
	if err != nil {
		t.Fatalf("Failed to dequeue data: %v", err)
	}

	// 检查出队的数据是否正确
	dequeuedStr, ok := dequeuedData.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", dequeuedData)
	}

	if dequeuedStr != testData {
		t.Errorf("Dequeued data does not match enqueued data: expected %s, got %s", testData, dequeuedStr)
	}
}
