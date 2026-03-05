package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sfsEdgeStore/config"
	"sfsEdgeStore/database"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/server"
)

func mainEncryption() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 启用加密
	cfg.DBUseEncryption = true
	cfg.DBEncryptionKey = "test_encryption_key_123456789012345678901234"
	cfg.DBEncryptionAlgorithm = "AES-256-GCM"

	// 初始化数据库
	if err = database.Init(cfg.DBPath, cfg.DBUseEncryption, cfg.DBEncryptionKey, cfg.DBEncryptionAlgorithm); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化监控
	monitorInstance := monitor.NewMonitor()

	// 创建服务器实例
	serverInstance := server.NewServer(database.Table, cfg, monitorInstance)

	// 启动服务器
	if err := serverInstance.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试1: 获取加密状态
	fmt.Println("Test 1: Get encryption status")
	testGetEncryptionStatus()

	// 测试2: 轮换加密密钥
	fmt.Println("\nTest 2: Rotate encryption key")
	testRotateEncryptionKey()

	// 测试3: 再次获取加密状态
	fmt.Println("\nTest 3: Get encryption status after rotation")
	testGetEncryptionStatus()

	fmt.Println("\nAll tests completed")
}

func testGetEncryptionStatus() {
	// 注意：这里需要使用实际的API Key
	apiKey := "your_admin_api_key"

	// 创建请求
	req, err := http.NewRequest("GET", "http://localhost:8081/api/encryption/status", nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	// 添加认证头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return
	}

	fmt.Printf("Encryption status: %+v\n", response)
}

func testRotateEncryptionKey() {
	// 注意：这里需要使用实际的API Key
	apiKey := "your_admin_api_key"

	// 准备请求数据
	reqData := map[string]string{
		"new_key": "new_encryption_key_123456789012345678901234",
	}
	data, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("Failed to marshal request data: %v", err)
		return
	}

	// 创建请求
	req, err := http.NewRequest("POST", "http://localhost:8081/api/encryption/rotate-key", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	// 添加认证头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return
	}

	fmt.Printf("Key rotation result: %+v\n", response)
}
