package database

import (
	"os"
	"testing"
)

// TestEncryptionRotation 测试密钥轮换功能
func TestEncryptionRotation(t *testing.T) {
	// 清理测试目录
	testDBPath := "./test_encryption_rotation_db"
	defer func() {
		os.RemoveAll(testDBPath)
	}()

	// 初始密钥
	initialKey := "test_encryption_key_123456789012345678901234"
	// 新密钥
	newKey := "new_encryption_key_123456789012345678901234"

	// 初始化加密数据库
	err := Init(testDBPath, true, initialKey, "AES-256-GCM")
	if err != nil {
		t.Fatalf("Failed to initialize encrypted database: %v", err)
	}

	// 测试数据
	testData := map[string]any{
		"id":         "test_id",
		"deviceName": "TestDevice",
		"reading":    "temperature",
		"value":      25.5,
		"valueType":  "Float32",
		"baseType":   "Float",
		"timestamp":  int64(1677721600000000000),
		"metadata":   "{\"location\": \"room1\"}",
	}

	// 插入测试数据
	_, err = Table.Insert(&testData)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 验证数据可以正确读取 - 使用范围查询而不是索引查询
	startRange := map[string]any{
		"deviceName": "TestDevice",
		"timestamp":  0,
	}

	endRange := map[string]any{
		"deviceName": "TestDevice",
		"timestamp":  int64(1677721600000000000),
	}

	iter, err := Table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		t.Fatalf("Failed to search test data: %v", err)
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	defer records.Release()

	if len(records) == 0 {
		t.Fatal("No records found after insertion")
	}

	insertedRecord := records[0]
	if insertedRecord["value"] != testData["value"] {
		t.Fatalf("Inserted value does not match: got %v, want %v", insertedRecord["value"], testData["value"])
	}

	// 执行密钥轮换
	err = RotateEncryptionKey(newKey)
	if err != nil {
		t.Fatalf("Failed to rotate encryption key: %v", err)
	}

	// 验证数据在密钥轮换后仍然可以正确读取
	iter2, err := Table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		t.Fatalf("Failed to search test data after rotation: %v", err)
	}
	defer iter2.Release()

	records2 := iter2.GetRecords(true)
	defer records2.Release()

	if len(records2) == 0 {
		t.Fatal("No records found after key rotation")
	}

	rotatedRecord := records2[0]
	if rotatedRecord["value"] != testData["value"] {
		t.Fatalf("Rotated value does not match: got %v, want %v", rotatedRecord["value"], testData["value"])
	}

	// 验证加密状态
	enabled, algorithm, err := GetEncryptionStatus()
	if err != nil {
		t.Fatalf("Failed to get encryption status: %v", err)
	}

	if !enabled {
		t.Fatal("Encryption should be enabled")
	}

	if algorithm != "AES-256-GCM" {
		t.Fatalf("Expected algorithm AES-256-GCM, got %s", algorithm)
	}

	t.Log("Encryption rotation test passed successfully")
}

// TestEncryptionStatus 测试加密状态获取功能
func TestEncryptionStatus(t *testing.T) {
	// 清理测试目录
	testDBPath1 := "./test_encryption_status_db1"
	testDBPath2 := "./test_encryption_status_db2"
	defer func() {
		os.RemoveAll(testDBPath1)
		os.RemoveAll(testDBPath2)
	}()

	// 初始化非加密数据库
	err := Init(testDBPath1, false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize non-encrypted database: %v", err)
	}

	// 验证加密状态（应该未启用）
	enabled, algorithm, err := GetEncryptionStatus()
	if err != nil {
		t.Fatalf("Failed to get encryption status: %v", err)
	}

	if enabled {
		t.Fatal("Encryption should be disabled")
	}

	if algorithm != "" {
		t.Fatalf("Expected empty algorithm, got %s", algorithm)
	}

	// 初始化加密数据库（使用不同的路径）
	err = Init(testDBPath2, true, "test_key", "AES-256-GCM")
	if err != nil {
		t.Fatalf("Failed to initialize encrypted database: %v", err)
	}

	// 验证加密状态（应该启用）
	enabled, algorithm, err = GetEncryptionStatus()
	if err != nil {
		t.Fatalf("Failed to get encryption status: %v", err)
	}

	if !enabled {
		t.Fatal("Encryption should be enabled")
	}

	if algorithm != "AES-256-GCM" {
		t.Fatalf("Expected algorithm AES-256-GCM, got %s", algorithm)
	}

	t.Log("Encryption status test passed successfully")
}

// TestEncryptionRotationNonEncrypted 测试对非加密数据库执行密钥轮换
func TestEncryptionRotationNonEncrypted(t *testing.T) {
	// 清理测试目录
	defer func() {
		os.RemoveAll("./test_encryption_rotation_non_encrypted_db")
	}()

	// 初始化非加密数据库
	err := Init("./test_encryption_rotation_non_encrypted_db", false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize non-encrypted database: %v", err)
	}

	// 尝试执行密钥轮换（应该失败）
	err = RotateEncryptionKey("new_key")
	if err == nil {
		t.Fatal("Expected error when rotating key on non-encrypted database")
	}

	if err.Error() != "database is not encrypted" {
		t.Fatalf("Expected error 'database is not encrypted', got %s", err.Error())
	}

	t.Log("Non-encrypted rotation test passed successfully")
}
