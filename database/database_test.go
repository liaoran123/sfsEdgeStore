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
