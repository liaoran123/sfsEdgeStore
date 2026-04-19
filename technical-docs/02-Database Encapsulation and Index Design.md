# Database Encapsulation and Index Design

## Overview

sfsEdgeStore uses sfsDb (based on LevelDB) as local data storage, providing efficient edge data persistence capabilities.

## Core Structure

### Global Table Instances

```go
// database/database.go:16-20
var Table *engine.Table
var AuthTable *engine.Table
```

## Initializing Database

### Init Function

```go
// database/database.go:23-162
func Init(dbPath string, useEncryption bool, encryptionKey, algorithm string) error {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	var dbScenario string
	cfgMgr := config.GetConfigManager()
	if cfgMgr != nil && cfgMgr.GetConfig() != nil {
		dbScenario = cfgMgr.GetConfig().DBScenario
		log.Printf("Using database scenario: %s", dbScenario)
	} else {
		dbScenario = storage.ScenarioEdge
		log.Printf("Using default database scenario: %s", dbScenario)
	}

	scenarioOptions := storage.GetConfigManager().GetScenarioOptions(dbScenario)

	storageConfig := storage.Config{
		WriteBuffer:            scenarioOptions.WriteBuffer,
		OpenFilesCacheCapacity: scenarioOptions.OpenFilesCacheCapacity,
		BlockCacheCapacity:     scenarioOptions.BlockCacheCapacity,
		Compression:            scenarioOptions.Compression,
	}
	storage.SetConfig(storageConfig)

	var err error
	if useEncryption {
		if encryptionKey == "" {
			return fmt.Errorf("encryption enabled but no encryption key provided")
		}
		masterKey := make([]byte, 32)
		copy(masterKey, []byte(encryptionKey))
		for i := len(encryptionKey); i < 32; i++ {
			masterKey[i] = 0
		}
		encryptConfig := &storage.EncryptionConfig{
			Enabled:   true,
			Algorithm: algorithm,
			MasterKey: masterKey,
		}
		_, err = storage.GetDBManager().OpenDBWithEncryption(dbPath, encryptConfig)
	} else {
		_, err = storage.GetDBManager().OpenDB(dbPath)
	}
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	tableName := "edgex_readings"
	var createErr error
	Table, createErr = engine.TableNew(tableName)
	if createErr != nil {
		return fmt.Errorf("failed to create table: %v", createErr)
	}

	fields := map[string]any{
		"id":         "",
		"deviceName": "",
		"reading":    "",
		"value":      0.0,
		"valueType":  "",
		"baseType":   "",
		"timestamp":  int64(0),
		"metadata":   "",
	}
	if err := Table.SetFields(fields); err != nil {
		return fmt.Errorf("failed to set table fields: %v", err)
	}

	primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
	if err != nil {
		return fmt.Errorf("failed to create primary key: %v", err)
	}
	primaryKey.AddFields("deviceName", "timestamp")
	if err := Table.CreateIndex(primaryKey); err != nil {
		if err.Error() != "index already exists" {
			return fmt.Errorf("failed to create primary key index: %v", err)
		}
	}

	authTableName := "edgex_auth"
	AuthTable, createErr = engine.TableNew(authTableName)
	if createErr != nil {
		return fmt.Errorf("failed to create auth table: %v", createErr)
	}

	authFields := map[string]any{
		"id":         "",
		"key":        "",
		"hash":       "",
		"user_id":    "",
		"role":       "",
		"created_at": int64(0),
		"expires_at": int64(0),
		"active":     false,
	}
	if err := AuthTable.SetFields(authFields); err != nil {
		return fmt.Errorf("failed to set auth table fields: %v", err)
	}

	authPrimaryKey, err := engine.DefaultPrimaryKeyNew("auth_pk")
	if err != nil {
		return fmt.Errorf("failed to create auth primary key: %v", err)
	}
	authPrimaryKey.AddFields("key")
	if err := AuthTable.CreateIndex(authPrimaryKey); err != nil {
		if err.Error() != "index already exists" {
			return fmt.Errorf("failed to create auth primary key index: %v", err)
		}
	}

	log.Println("Database initialized successfully")
	return nil
}
```

**Key Features:**
- Scenario configuration: Optimized configuration for edge scenarios
- Encryption support: Optional database encryption
- Composite primary key: `(deviceName + timestamp)` for optimized query performance

## Batch Insertion

### BatchInsertWithRetry Function

```go
// database/database.go:208-222
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
```

## Query Records

### QueryRecords Function

```go
// database/database.go:225-288
func QueryRecords(tbl *engine.Table, deviceName, startTime, endTime string) (record.Records, error) {
	formattedDeviceName := common.FormatDeviceName(deviceName)

	log.Println("Querying readings with filters:")
	log.Printf("  deviceName: %s", deviceName)
	log.Printf("  formattedDeviceName: %s", formattedDeviceName)
	log.Printf("  startTime: %s", startTime)
	log.Printf("  endTime: %s", endTime)

	var startTimestamp, endTimestamp *int64

	if startTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			ts := start.UnixNano()
			startTimestamp = &ts
		}
	}

	if endTime != "" {
		end, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			ts := end.UnixNano()
			endTimestamp = &ts
		}
	}

	startRange := make(map[string]any)
	endRange := make(map[string]any)

	startRange["deviceName"] = formattedDeviceName
	endRange["deviceName"] = formattedDeviceName

	if startTimestamp != nil {
		startRange["timestamp"] = *startTimestamp
	} else {
		startRange["timestamp"] = nil
	}

	if endTimestamp != nil {
		endRange["timestamp"] = *endTimestamp
	} else {
		endRange["timestamp"] = nil
	}

	iter, err := tbl.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		return nil, fmt.Errorf("failed to search readings: %v", err)
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	return records, nil
}
```

## Data Export/Import

### Export Functions

```go
// database/database.go:291-313
func ExportTableToCSV(tbl *engine.Table, filePath string) error {
	return tbl.ExportToCSV(filePath)
}

func ImportTableFromCSV(tbl *engine.Table, filePath string, batchSize int) error {
	return tbl.ImportFromCSV(filePath, batchSize)
}

func ExportTableToJSON(tbl *engine.Table, filePath string) error {
	return tbl.ExportToJSON(filePath)
}

func ImportTableFromJSON(tbl *engine.Table, filePath string, batchSize int) error {
	return tbl.ImportFromJSON(filePath, batchSize)
}

func ExportTableToSQL(tbl *engine.Table, filePath string) error {
	return tbl.ExportToSQL(filePath)
}
```

## Encryption Management

### Key Rotation

```go
// database/database.go:165-187
func RotateEncryptionKey(newKey string) error {
	store := storage.GetDBManager().GetDB()
	if store == nil {
		return fmt.Errorf("database not initialized")
	}

	encryptedStore, ok := store.(*storage.EncryptedStoreWrapper)
	if !ok {
		return fmt.Errorf("database is not encrypted")
	}

	masterKey := make([]byte, 32)
	copy(masterKey, []byte(newKey))
	for i := len(newKey); i < 32; i++ {
		masterKey[i] = 0
	}

	return encryptedStore.ReEncrypt(masterKey)
}
```

### Get Encryption Status

```go
// database/database.go:190-205
func GetEncryptionStatus() (bool, string, error) {
	store := storage.GetDBManager().GetDB()
	if store == nil {
		return false, "", fmt.Errorf("database not initialized")
	}

	encryptedStore, ok := store.(*storage.EncryptedStoreWrapper)
	if !ok {
		return false, "", nil
	}

	config := encryptedStore.GetEncryptionConfig()
	return true, config.Algorithm, nil
}
```

## Testing

### Database Testing

```go
// database/database_test.go
package database

import (
	"os"
	"testing"
	"time"
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

	testData := map[string]any{
		"id":         "test_id_1",
		"deviceName": "Device001",
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

	records, err := QueryRecords(Table, "Device001", "", "")
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
```

### Running Tests

```bash
go test ./database -v
```

## API Interface

### Init Initialize Database

```go
func Init(dbPath string, useEncryption bool, encryptionKey, algorithm string) error
```

### BatchInsertWithRetry Batch Insert with Retry

```go
func BatchInsertWithRetry(tbl *engine.Table, records []*map[string]any, maxRetries int, retryInterval time.Duration) error
```

### QueryRecords Query Records

```go
func QueryRecords(tbl *engine.Table, deviceName, startTime, endTime string) (record.Records, error)
```

### RotateEncryptionKey Rotate Encryption Key

```go
func RotateEncryptionKey(newKey string) error
```

### GetEncryptionStatus Get Encryption Status

```go
func GetEncryptionStatus() (bool, string, error)
```