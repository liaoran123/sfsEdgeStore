package database

import (
	"os"
	"testing"
	"time"

	"sfsEdgeStore/common"
)

func TestDatabaseInit(t *testing.T) {
	testDBPath := "./test_db_init"
	defer func() {
		os.RemoveAll(testDBPath)
	}()

	err := Init(testDBPath, false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	if Table == nil {
		t.Fatal("Table should not be nil")
	}

	if AuthTable == nil {
		t.Fatal("AuthTable should not be nil")
	}

	t.Log("Database init test passed successfully")
}

func TestInsertAndQuery(t *testing.T) {
	testDBPath := "./test_insert_query"
	defer func() {
		os.RemoveAll(testDBPath)
	}()

	err := Init(testDBPath, false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	formattedDeviceName := common.FormatDeviceName("Device001")
	testData := map[string]any{
		"id":         "test_id_1",
		"deviceName": formattedDeviceName,
		"reading":    "temperature",
		"value":      25.5,
		"valueType":  "Float32",
		"baseType":   "Float",
		"timestamp":  time.Now().UnixNano(),
		"metadata":   "{\"location\": \"room1\"}",
	}

	testRecords := []*map[string]any{&testData}
	err = BatchInsertWithRetry(Table, testRecords, 3, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	records, err := QueryRecords(Table, "Device001", "", "", false)
	if err != nil {
		t.Fatalf("Failed to query records: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("No records found")
	}

	insertedRecord := records[0]
	insertedValue, ok := insertedRecord["value"].(float64)
	if !ok {
		t.Fatalf("Value is not float64: %T", insertedRecord["value"])
	}
	testValue := testData["value"].(float64)
	if insertedValue != testValue {
		t.Fatalf("Value mismatch: got %v, want %v", insertedValue, testValue)
	}

	t.Log("Insert and query test passed successfully")
}

func TestAuthTableOperations(t *testing.T) {
	testDBPath := "./test_auth_table"
	defer func() {
		os.RemoveAll(testDBPath)
	}()

	err := Init(testDBPath, false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	authData := map[string]any{
		"id":         "auth_1",
		"key":        "api_key_123",
		"hash":       "hashed_secret",
		"user_id":    "user_001",
		"role":       "admin",
		"created_at": time.Now().UnixNano(),
		"expires_at": time.Now().Add(24 * time.Hour).UnixNano(),
		"active":     true,
	}

	authRecords := []*map[string]any{&authData}
	_, err = AuthTable.BatchInsertNoInc(authRecords)
	if err != nil {
		t.Fatalf("Failed to insert auth data: %v", err)
	}

	startRange := map[string]any{
		"key": "api_key_123",
	}

	endRange := map[string]any{
		"key": "api_key_123",
	}

	iter, err := AuthTable.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		t.Fatalf("Failed to search auth data: %v", err)
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	defer records.Release()

	if len(records) == 0 {
		t.Fatal("No auth records found")
	}

	insertedAuth := records[0]
	if insertedAuth["user_id"] != authData["user_id"] {
		t.Fatalf("User ID mismatch: got %v, want %v", insertedAuth["user_id"], authData["user_id"])
	}

	t.Log("Auth table operations test passed successfully")
}

func TestQueryRecordsOrder(t *testing.T) {
	testDBPath := "./test_query_order"
	// 先清理目录，确保没有残留数据
	os.RemoveAll(testDBPath)
	defer func() {
		os.RemoveAll(testDBPath)
	}()

	err := Init(testDBPath, false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	formattedDeviceName := common.FormatDeviceName("TestDevice")

	// 插入多条带有不同时间戳的记录
	var testRecords []*map[string]any
	for i := 0; i < 5; i++ {
		timestamp := time.Now().Add(time.Duration(i) * time.Second).UnixNano()
		data := map[string]any{
			"id":         "test_id_" + string(rune('0'+i)),
			"deviceName": formattedDeviceName,
			"reading":    "temperature",
			"value":      float64(20 + i),
			"valueType":  "Float32",
			"baseType":   "Float",
			"timestamp":  timestamp,
			"metadata":   "{\"test\": \"data\"}",
		}
		testRecords = append(testRecords, &data)
	}

	// 直接使用Table.Insert插入数据
	for _, record := range testRecords {
		_, err := Table.Insert(record)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 测试esc=false（默认顺序，按时间降序）
	recordsDesc, err := QueryRecords(Table, "TestDevice", "", "", false)
	if err != nil {
		t.Fatalf("Failed to query records with esc=false: %v", err)
	}
	defer recordsDesc.Release()

	if len(recordsDesc) != 5 {
		t.Fatalf("Expected 5 records, got %d", len(recordsDesc))
	}

	// 验证顺序是否为降序
	for i := 0; i < len(recordsDesc)-1; i++ {
		timestamp1 := recordsDesc[i]["timestamp"].(int64)
		timestamp2 := recordsDesc[i+1]["timestamp"].(int64)
		if timestamp1 < timestamp2 {
			t.Fatalf("Records should be in descending order, but timestamp %d < %d", timestamp1, timestamp2)
		}
	}

	// 测试esc=true（倒序，按时间升序）
	recordsAsc, err := QueryRecords(Table, "TestDevice", "", "", true)
	if err != nil {
		t.Fatalf("Failed to query records with esc=true: %v", err)
	}
	defer recordsAsc.Release()

	if len(recordsAsc) != 5 {
		t.Fatalf("Expected 5 records, got %d", len(recordsAsc))
	}

	// 验证顺序是否为升序
	for i := 0; i < len(recordsAsc)-1; i++ {
		timestamp1 := recordsAsc[i]["timestamp"].(int64)
		timestamp2 := recordsAsc[i+1]["timestamp"].(int64)
		if timestamp1 > timestamp2 {
			t.Fatalf("Records should be in ascending order, but timestamp %d > %d", timestamp1, timestamp2)
		}
	}

	t.Log("Query records order test passed successfully")
}

func TestExpiredDataDeletion(t *testing.T) {
	testDBPath := "./test_expired_deletion_" + time.Now().Format("20060102150405")
	// 先清理目录，确保没有残留数据
	os.RemoveAll(testDBPath)
	defer func() {
		os.RemoveAll(testDBPath)
	}()

	err := Init(testDBPath, false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	formattedDeviceName := common.FormatDeviceName("TestDevice")

	// 插入测试数据，包括过期和未过期的记录
	var testRecords []*map[string]any

	// 过期记录（10分钟前）
	expiredTimestamp := time.Now().Add(-10 * time.Minute).UnixNano()
	expiredData := map[string]any{
		"id":         "expired_id",
		"deviceName": formattedDeviceName,
		"reading":    "temperature",
		"value":      25.5,
		"valueType":  "Float32",
		"baseType":   "Float",
		"timestamp":  expiredTimestamp,
		"metadata":   "{\"test\": \"expired\"}",
	}
	testRecords = append(testRecords, &expiredData)

	// 未过期记录（当前时间）
	currentTimestamp := time.Now().UnixNano()
	currentData := map[string]any{
		"id":         "current_id",
		"deviceName": formattedDeviceName,
		"reading":    "temperature",
		"value":      26.5,
		"valueType":  "Float32",
		"baseType":   "Float",
		"timestamp":  currentTimestamp,
		"metadata":   "{\"test\": \"current\"}",
	}
	testRecords = append(testRecords, &currentData)

	// 插入数据
	for _, record := range testRecords {
		_, err := Table.Insert(record)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 验证初始数据量
	initialRecords, err := QueryRecords(Table, "", "", "", false)
	//initialRecords, err := QueryRecords(Table, "TestDevice", "", "", false)
	if err != nil {
		t.Fatalf("Failed to query initial records: %v", err)
	}
	defer initialRecords.Release()

	if len(initialRecords) != 2 {
		t.Fatalf("Expected 2 initial records, got %d", len(initialRecords))
	}

	// 模拟cleanupBatch函数的逻辑，删除过期数据
	cutoffTimestamp := time.Now().Add(-5 * time.Minute).UnixNano()
	startRange := make(map[string]any)
	endRange := make(map[string]any)
	startRange["deviceName"] = formattedDeviceName
	startRange["timestamp"] = nil
	endRange["deviceName"] = formattedDeviceName
	endRange["timestamp"] = cutoffTimestamp

	iter, err := Table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		t.Fatalf("Failed to search range: %v", err)
	}
	defer iter.Release()

	records := iter.GetRecords(true, 100)
	defer records.Release()

	// 验证找到的过期记录数量
	if len(records) != 1 {
		t.Fatalf("Expected 1 expired record, got %d", len(records))
	}

	// 删除过期记录
	err = iter.Delete()
	if err != nil {
		t.Fatalf("Failed to delete expired records: %v", err)
	}

	// 验证删除后的数据量
	remainingRecords, err := QueryRecords(Table, "TestDevice", "", "", false)
	if err != nil {
		t.Fatalf("Failed to query remaining records: %v", err)
	}
	defer remainingRecords.Release()

	if len(remainingRecords) != 1 {
		t.Fatalf("Expected 1 remaining record, got %d", len(remainingRecords))
	}

	// 验证剩余的是未过期记录
	remainingRecord := remainingRecords[0]
	if remainingRecord["id"] != "current_id" {
		t.Fatalf("Expected remaining record to be current_id, got %v", remainingRecord["id"])
	}

	t.Log("Expired data deletion test passed successfully")
}

func TestCleanupBatch(t *testing.T) {
	testDBPath := "./test_cleanup_batch_" + time.Now().Format("20060102150405")
	// 先清理目录，确保没有残留数据
	os.RemoveAll(testDBPath)
	defer func() {
		os.RemoveAll(testDBPath)
	}()

	err := Init(testDBPath, false, "", "")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	formattedDeviceName := common.FormatDeviceName("TestDevice")

	// 插入测试数据，包括多条过期记录
	var testRecords []*map[string]any

	// 过期记录1（10分钟前）
	expiredTimestamp1 := time.Now().Add(-10 * time.Minute).UnixNano()
	expiredData1 := map[string]any{
		"id":         "expired_id_1",
		"deviceName": formattedDeviceName,
		"reading":    "temperature",
		"value":      25.5,
		"valueType":  "Float32",
		"baseType":   "Float",
		"timestamp":  expiredTimestamp1,
		"metadata":   "{\"test\": \"expired1\"}",
	}
	testRecords = append(testRecords, &expiredData1)

	// 过期记录2（8分钟前）
	expiredTimestamp2 := time.Now().Add(-8 * time.Minute).UnixNano()
	expiredData2 := map[string]any{
		"id":         "expired_id_2",
		"deviceName": formattedDeviceName,
		"reading":    "temperature",
		"value":      26.0,
		"valueType":  "Float32",
		"baseType":   "Float",
		"timestamp":  expiredTimestamp2,
		"metadata":   "{\"test\": \"expired2\"}",
	}
	testRecords = append(testRecords, &expiredData2)

	// 未过期记录（当前时间）
	currentTimestamp := time.Now().UnixNano()
	currentData := map[string]any{
		"id":         "current_id",
		"deviceName": formattedDeviceName,
		"reading":    "temperature",
		"value":      27.0,
		"valueType":  "Float32",
		"baseType":   "Float",
		"timestamp":  currentTimestamp,
		"metadata":   "{\"test\": \"current\"}",
	}
	testRecords = append(testRecords, &currentData)

	// 插入数据
	for _, record := range testRecords {
		_, err := Table.Insert(record)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 验证初始数据量
	startRange := make(map[string]any)
	endRange := make(map[string]any)
	startRange["deviceName"] = formattedDeviceName
	startRange["timestamp"] = nil
	endRange["deviceName"] = formattedDeviceName
	endRange["timestamp"] = time.Now().Add(1 * time.Minute).UnixNano() // 确保包含所有测试数据

	iter, err := Table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		t.Fatalf("Failed to search range: %v", err)
	}
	defer iter.Release()

	initialRecords := iter.GetRecords(true, 100)
	defer initialRecords.Release()

	if len(initialRecords) != 3 {
		t.Fatalf("Expected 3 initial records, got %d", len(initialRecords))
	}

	// 模拟cleanupBatch函数的逻辑，删除过期数据
	cutoffTimestamp := time.Now().Add(-5 * time.Minute).UnixNano()
	batchSize := 2

	// 第一次清理批次
	startRange = make(map[string]any)
	endRange = make(map[string]any)
	startRange["deviceName"] = formattedDeviceName
	startRange["timestamp"] = nil
	endRange["deviceName"] = formattedDeviceName
	endRange["timestamp"] = cutoffTimestamp

	iter, err = Table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		t.Fatalf("Failed to search range: %v", err)
	}
	defer iter.Release()

	records := iter.GetRecords(true, batchSize)
	defer records.Release()

	// 验证找到的过期记录数量
	if len(records) != 2 {
		t.Fatalf("Expected 2 expired records, got %d", len(records))
	}

	// 删除过期记录
	err = iter.Delete()
	if err != nil {
		t.Fatalf("Failed to delete expired records: %v", err)
	}

	// 验证删除后的数据量
	remainingStartRange := make(map[string]any)
	remainingEndRange := make(map[string]any)
	remainingStartRange["deviceName"] = formattedDeviceName
	remainingStartRange["timestamp"] = nil
	remainingEndRange["deviceName"] = formattedDeviceName
	remainingEndRange["timestamp"] = time.Now().Add(1 * time.Minute).UnixNano()

	remainingIter, err := Table.SearchRange(nil, &remainingStartRange, &remainingEndRange)
	if err != nil {
		t.Fatalf("Failed to search range: %v", err)
	}
	defer remainingIter.Release()

	remainingRecords := remainingIter.GetRecords(true, 100)
	defer remainingRecords.Release()

	if len(remainingRecords) != 1 {
		t.Fatalf("Expected 1 remaining record, got %d", len(remainingRecords))
	}

	// 验证剩余的是未过期记录
	remainingRecord := remainingRecords[0]
	if remainingRecord["id"] != "current_id" {
		t.Fatalf("Expected remaining record to be current_id, got %v", remainingRecord["id"])
	}

	// 第二次清理批次（应该没有更多过期记录）
	startRange = make(map[string]any)
	endRange = make(map[string]any)
	startRange["deviceName"] = formattedDeviceName
	startRange["timestamp"] = nil
	endRange["deviceName"] = formattedDeviceName
	endRange["timestamp"] = cutoffTimestamp

	iter, err = Table.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		t.Fatalf("Failed to search range: %v", err)
	}
	defer iter.Release()

	records = iter.GetRecords(true, batchSize)
	defer records.Release()

	// 验证没有更多过期记录
	if len(records) != 0 {
		t.Fatalf("Expected 0 expired records, got %d", len(records))
	}

	t.Log("Cleanup batch test passed successfully")
}
