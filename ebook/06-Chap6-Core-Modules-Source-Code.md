# Chapter 6: Core Module Source Code Analysis

## 6.1 Project Structure Overview

Before diving into specific modules, let's understand the overall project structure and how components interact.

### Project Directory Structure

```bash
sfsEdgeStore/
├── main.go              # Application entry point
├── agent/               # Management agent
├── alert/               # Alert notification system
├── analyzer/            # Data analysis engine
├── auth/                # Authentication and authorization
├── common/              # Shared utilities
├── config/              # Configuration management
├── database/            # Database encapsulation (sfsDb)
├── edgex/              # EdgeX message processing
├── logger/              # Logging
├── monitor/             # Monitoring metrics
├── mqtt/                # MQTT client
├── queue/               # Data queue (fault recovery)
├── resource/            # Resource monitoring
├── retention/           # Data retention
├── server/              # HTTP server
├── simulator/           # Data simulator
├── sync/                # Cloud synchronization
└── time/                # Time series processing
```

### Component Interaction

```
┌─────────────────────────────────────────────────────────────────┐
│                         main.go                                  │
│                    (Application Entry)                          │
└───────────────────────────┬─────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│     mqtt      │  │   database    │  │     queue     │
│   (Input)     │  │   (Storage)   │  │  (Recovery)   │
└───────┬───────┘  └───────────────┘  └───────┬───────┘
        │                                      │
        │ Data                                 │ Retry
        │                                       │
        ▼                                       │
┌───────────────┐                                │
│     edgex     │                                │
│  (Processing) │                                │
└───────┬───────┘                                │
        │                                        │
        │ Records                                │
        │                                        │
        ▼                                        │
┌───────────────┐                                │
│    monitor    │◄───────────────────────────────┘
│  (Metrics)    │          Recovery
└───────────────┘
        │
        │ Alerts
        ▼
┌───────────────┐
│     alert     │
│ (Notification)│
└───────────────┘
```

## 6.2 MQTT Client Implementation

The MQTT client is the entry point for data ingestion. This section analyzes the implementation from `mqtt/client.go`.

### Client Structure

```go
// From mqtt/client.go:27-37
// This is the core MQTT client structure

type Client struct {
    client        mqtt.Client           // Paho MQTT client
    config        *config.Config        // Configuration
    dataQueue     *queue.Queue          // Fault recovery queue
    monitor       *monitor.Monitor      // Metrics collector
    analyzer      *analyzer.Analyzer   // Data analyzer
    batchMessages []map[string]interface{}  // Message buffer
    batchSize     int                  // Batch size threshold
    batchInterval time.Duration        // Batch timeout
    lastBatchTime time.Time            // Last batch timestamp
}
```

**Field Design Rationale:**

| Field | Purpose | Why It Matters |
|-------|---------|----------------|
| `client` | MQTT connection | Core protocol handler |
| `config` | Settings access | Reconnection, TLS config |
| `dataQueue` | Fault tolerance | Survives database outages |
| `monitor` | Observability | Track message flow |
| `analyzer` | Real-time analytics | Threshold alerts |
| `batchMessages` | Performance | Reduce database writes |

### Creating the Client

```go
// From mqtt/client.go:40-136
func NewClient(cfg *config.Config, dataQueue *queue.Queue,
               monitor *monitor.Monitor, analyzer *analyzer.Analyzer) (*Client, error) {

    // 1. Create MQTT client options
    opts := mqtt.NewClientOptions()

    // 2. Configure broker connection
    opts.AddBroker(cfg.MQTTBroker)              // Broker URL
    opts.SetClientID(cfg.ClientID)              // Unique ID
    opts.SetCleanSession(false)                 // Persist subscription
    opts.SetAutoReconnect(true)                 // Auto recover
    opts.SetMaxReconnectInterval(5 * time.Minute) // Backoff

    // 3. Set up last will (for visibility)
    willTopic := cfg.MQTTTopic + "/status"
    willMessage := map[string]interface{}{
        "status":    "offline",
        "clientId":  cfg.ClientID,
        "timestamp": time.Now().UnixNano(),
    }
    willPayload, _ := json.Marshal(willMessage)
    opts.SetWill(willTopic, string(willPayload), 1, false)

    // 4. Configure TLS if enabled
    if cfg.MQTTUseTLS {
        tlsConfig := &tls.Config{
            InsecureSkipVerify: false,  // Always verify!
        }
        // ... load certificates ...
        opts.SetTLSConfig(tlsConfig)
    }

    // 5. Create client instance
    client := &Client{
        config:        cfg,
        dataQueue:    dataQueue,
        monitor:      monitor,
        analyzer:      analyzer,
        batchMessages: make([]map[string]interface{}, 0),
        batchSize:     100,           // Batch threshold
        batchInterval: 5 * time.Second, // Batch timeout
        lastBatchTime: time.Now(),
    }

    // 6. Set up connection handlers
    opts.SetOnConnectHandler(func(mqttClient mqtt.Client) {
        log.Println("MQTT broker connected")
        // Publish online status
        onlineTopic := cfg.MQTTTopic + "/status"
        // ... publish online message ...

        // Re-subscribe to data topic
        token := mqttClient.Subscribe(cfg.MQTTTopic, 1, client.messageHandler())
        token.Wait()
    })

    // 7. Connect
    mqttClient := mqtt.NewClient(opts)
    token := mqttClient.Connect()
    token.Wait()
    if token.Error() != nil {
        return nil, token.Error()
    }

    return client, nil
}
```

**Key Design Decisions:**

1. **CleanSession=false**: Maintains subscription across reconnections
2. **AutoReconnect=true**: Automatically recovers from network issues
3. **MaxReconnectInterval=5min**: Prevents connection storm
4. **Will message**: Publishes offline status for monitoring

### Message Handler

The message handler is called for each received MQTT message:

```go
// From mqtt/client.go:238-396
func (c *Client) messageHandler() mqtt.MessageHandler {
    return func(client mqtt.Client, msg mqtt.Message) {

        // 1. Track received message
        if c.monitor != nil {
            c.monitor.IncrementMQTTMessagesReceived()
        }

        log.Printf("Received message on topic: %s", msg.Topic())

        // 2. Process in goroutine (non-blocking)
        go func() {
            // 3. Parse EdgeX message format
            event, err := edgex.ProcessMessage(msg.Payload())
            if err != nil {
                log.Printf("Failed to process message: %v", err)
                return
            }
            if event == nil {
                return  // Ignored message type
            }

            // 4. Convert to storage records
            records := make([]*map[string]any, 0, len(event.Readings))

            for _, reading := range event.Readings {
                // Get reusable map from pool (performance optimization)
                data := objPool.GetMap()

                // Parse reading value
                metadataStr := ""
                if reading.Metadata != nil {
                    metadataStr = string(reading.Metadata)
                }
                value := common.ParseValue(reading.Value)

                // Fill record
                data["id"] = reading.ID
                data["deviceName"] = event.DeviceName
                data["reading"] = reading.ResourceName
                data["value"] = value
                data["valueType"] = reading.ValueType
                data["baseType"] = reading.BaseType
                data["timestamp"] = reading.Origin
                data["metadata"] = metadataStr

                records = append(records, &data)
            }

            // 5. Store to database
            if len(records) > 0 {
                if c.monitor != nil {
                    c.monitor.IncrementDatabaseOperations()
                }

                err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)

                if err != nil {
                    // Database write failed - queue for retry
                    log.Printf("Failed to store data: %v", err)

                    // Categorize error
                    errorMsg := err.Error()
                    if strings.Contains(errorMsg, "no space left") {
                        c.monitor.RecordError("storage_error", errorMsg)
                    }

                    // Enqueue for later retry
                    if err := c.dataQueue.Enqueue(records); err != nil {
                        log.Printf("Failed to enqueue data: %v", err)
                    }

                } else {
                    // Success
                    log.Printf("Batch stored %d readings", len(records))
                    c.monitor.IncrementMQTTMessagesProcessed()

                    // Run analytics
                    if c.analyzer != nil && c.analyzer.IsEnabled() {
                        // ... analysis code ...
                    }
                }

                // Return maps to pool (cleanup)
                for _, data := range records {
                    objPool.PutMap(*data)
                }
            }
        }()
    }
}
```

**Processing Flow:**

```
MQTT Message Received
        │
        ▼
┌───────────────────┐
│ Parse EdgeX JSON  │ ← edgex.ProcessMessage()
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ Convert Readings  │ ← Map to storage format
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ Batch Insert      │ ← With retry logic
└─────────┬─────────┘
          │
    ┌─────┴─────┐
    │           │
Success     Failure
    │           │
    ▼           ▼
┌───────────┐ ┌───────────────┐
│ Analytics │ │ Enqueue for   │
│ (if on)   │ │ Retry         │
└───────────┘ └───────────────┘
```

## 6.3 Object Pool Implementation

The object pool reduces GC pressure by reusing map objects:

```go
// From mqtt/mapPool.go
// This is a sync.Pool based object pool for map reuse

package mqtt

var objPool = NewObjectPool()  // Global pool instance

type objectPool struct {
    mapPool sync.Pool  // Standard library pool
}

func NewObjectPool() *objectPool {
    return &objectPool{
        mapPool: sync.Pool{
            New: func() interface{} {
                // Factory function - creates new map when pool is empty
                return make(map[string]any)
            },
        },
    }
}

// GetMap retrieves a map from the pool
// IMPORTANT: Always returns an EMPTY map
func (p *objectPool) GetMap() map[string]any {
    m := p.mapPool.Get().(map[string]any)

    // Clear any residual data
    // This is crucial for safety!
    for k := range m {
        delete(m, k)
    }
    return m
}

// PutMap returns a map to the pool
// IMPORTANT: Must clear before returning!
func (p *objectPool) PutMap(m map[string]any) {
    // Clear all entries before pooling
    for k := range m {
        delete(m, k)
    }
    p.mapPool.Put(m)
}
```

**Why This Matters:**

```go
// WITHOUT object pool:
for _, reading := range event.Readings {
    data := make(map[string]any)  // Allocation!
    // ... fill data ...
}
// GC pressure: 1000 readings = 1000 allocations per message
// GC pause: Frequent, unpredictable pauses

// WITH object pool:
for _, reading := range event.Readings {
    data := objPool.GetMap()      // Reuse existing
    // ... fill data ...
    objPool.PutMap(data)          // Return for reuse
}
// GC pressure: Minimal allocations
// Memory usage: Bounded by pool size
```

**Test Case:**

```go
// From mqtt/mqtt_test.go
func TestObjectPool(t *testing.T) {
    // Get a map
    m1 := objPool.GetMap()
    if m1 == nil {
        t.Fatal("Expected non-nil map")
    }
    if len(m1) != 0 {
        t.Errorf("Expected empty map, got %d entries", len(m1))
    }

    // Use it
    m1["key1"] = "value1"
    m1["key2"] = 42

    // Return to pool
    objPool.PutMap(m1)

    // Get again - should be empty
    m2 := objPool.GetMap()
    if len(m2) != 0 {
        t.Errorf("Expected empty map after PutMap, got %d entries", len(m2))
    }

    t.Log("Object pool test passed")
}

func TestObjectPoolMultiple(t *testing.T) {
    // Test concurrent access
    maps := make([]map[string]any, 10)

    for i := 0; i < 10; i++ {
        maps[i] = objPool.GetMap()
        maps[i]["index"] = i
    }

    // Return all
    for i := 0; i < 10; i++ {
        objPool.PutMap(maps[i])
    }

    // Verify they were properly cleared
    for i := 0; i < 10; i++ {
        m := objPool.GetMap()
        if len(m) != 0 {
            t.Errorf("Expected empty map at index %d", i)
        }
        objPool.PutMap(m)
    }

    t.Log("Multiple object pool test passed")
}
```

## 6.4 Database Encapsulation

The database module wraps sfsDb with additional features:

### Database Initialization

```go
// From database/database.go:23-136
func Init(dbPath string, useEncryption bool, encryptionKey, algorithm string) error {

    // 1. Create database directory
    if err := os.MkdirAll(dbPath, 0755); err != nil {
        return fmt.Errorf("failed to create database directory: %v", err)
    }

    // 2. Load scenario configuration
    var dbScenario string
    cfgMgr := config.GetConfigManager()
    if cfgMgr != nil && cfgMgr.GetConfig() != nil {
        dbScenario = cfgMgr.GetConfig().DBScenario
    } else {
        dbScenario = storage.ScenarioEdge  // Default to edge scenario
    }

    scenarioOptions := storage.GetConfigManager().GetScenarioOptions(dbScenario)

    // 3. Configure storage engine
    storageConfig := storage.Config{
        WriteBuffer:            scenarioOptions.WriteBuffer,
        OpenFilesCacheCapacity: scenarioOptions.OpenFilesCacheCapacity,
        BlockCacheCapacity:     scenarioOptions.BlockCacheCapacity,
        Compression:            scenarioOptions.Compression,
    }
    storage.SetConfig(storageConfig)

    // 4. Open database (with or without encryption)
    var err error
    if useEncryption {
        if encryptionKey == "" {
            return fmt.Errorf("encryption enabled but no key provided")
        }
        // Prepare 32-byte key for AES-256
        masterKey := make([]byte, 32)
        copy(masterKey, []byte(encryptionKey))
        for i := len(encryptionKey); i < 32; i++ {
            masterKey[i] = 0  // Pad with zeros
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

    // 5. Create main readings table
    tableName := "edgex_readings"
    Table, err = engine.TableNew(tableName)
    if err != nil {
        return fmt.Errorf("failed to create table: %v", err)
    }

    // 6. Define table schema
    fields := map[string]any{
        "id":         "",        // Reading ID
        "deviceName": "",        // Device identifier (padded to 64 chars)
        "reading":    "",        // Resource/reading name
        "value":      0.0,      // Numeric value
        "valueType":  "",        // Data type
        "baseType":   "",        // Base type
        "timestamp":  int64(0), // Nanosecond timestamp
        "metadata":   "",        // Optional metadata JSON
    }
    if err := Table.SetFields(fields); err != nil {
        return fmt.Errorf("failed to set table fields: %v", err)
    }

    // 7. Create composite primary key
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

    // 8. Create auth table (for API keys)
    authTableName := "edgex_auth"
    AuthTable, err = engine.TableNew(authTableName)
    // ... similar initialization ...

    log.Println("Database initialized successfully")
    return nil
}
```

**Schema Design:**

```go
// Fields define the data model
fields := map[string]any{
    "id":         "",        // Unique per reading
    "deviceName": "",        // 64-char padded device ID
    "reading":    "",        // Resource name (e.g., "temperature")
    "value":      0.0,      // Value is always stored as float64
    "valueType":  "",        // Original type (Float32, Int64, etc.)
    "baseType":   "",        // Base type (Float, Int, etc.)
    "timestamp":  int64(0), // Unix nanoseconds
    "metadata":   "",        // JSON metadata
}

// Primary key: deviceName + timestamp
// This enables:
// 1. Fast device lookup (prefix scan)
// 2. Fast time range within device (range scan)
// 3. Sorted results (natural ordering)
```

### Batch Insert with Retry

```go
// From database/database.go:208-222
// Batch insert with automatic retry on failure

func BatchInsertWithRetry(tbl *engine.Table, records []*map[string]any,
                          maxRetries int, retryInterval time.Duration) error {

    for i := 0; i < maxRetries; i++ {
        // Attempt batch insert
        _, err := tbl.BatchInsertNoInc(records)

        if err == nil {
            return nil  // Success
        }

        // Log failure
        log.Printf("Failed to batch insert data (attempt %d/%d): %v",
                   i+1, maxRetries, err)

        // Wait before retry (except on last attempt)
        if i < maxRetries-1 {
            time.Sleep(retryInterval)
        }
    }

    return fmt.Errorf("failed to batch insert data after %d attempts", maxRetries)
}
```

**Retry Logic:**

```go
// Why retry?
// 1. Transient lock conflicts (other process accessing DB)
// 2. Temporary disk full
// 3. System under heavy load

// Retry strategy:
// - 3 attempts
// - 2 second interval between attempts
// - Total time: ~4 seconds max
```

### Query Records

```go
// From database/database.go:225-288
// Query with device name and optional time range

func QueryRecords(tbl *engine.Table, deviceName, startTime, endTime string) (record.Records, error) {

    // Format device name (pad to 64 chars for index)
    formattedDeviceName := common.FormatDeviceName(deviceName)

    // Parse time range to nanoseconds
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

    // Build range query bounds
    startRange := make(map[string]any)
    endRange := make(map[string]any)

    startRange["deviceName"] = formattedDeviceName
    endRange["deviceName"] = formattedDeviceName

    if startTimestamp != nil {
        startRange["timestamp"] = *startTimestamp
    } else {
        startRange["timestamp"] = nil  // No lower bound
    }

    if endTimestamp != nil {
        endRange["timestamp"] = *endTimestamp
    } else {
        endRange["timestamp"] = nil  // No upper bound
    }

    // Execute range query
    iter, err := tbl.SearchRange(nil, &startRange, &endRange)
    if err != nil {
        return nil, fmt.Errorf("failed to search readings: %v", err)
    }
    defer iter.Release()

    // Get results
    records := iter.GetRecords(true)
    return records, nil
}
```

**Query Optimization:**

```go
// This query uses the composite index efficiently:
// Index: (deviceName, timestamp)

// For query: "All readings from Device001 in January 2024"
// 1. Seek to "Device001" + 2024-01-01 00:00:00
// 2. Scan forward until > "Device001" + 2024-01-31 23:59:59
// 3. Return all matching records

// Time complexity: O(log n + k) where k = matching records
// NOT O(n) full table scan!
```

### Database Tests

```go
// From database/database_test.go
package database

import (
    "os"
    "testing"
    "time"
)

func TestDatabaseInit(t *testing.T) {
    testDBPath := "./test_db_init"
    defer os.RemoveAll(testDBPath)

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
    defer os.RemoveAll(testDBPath)

    err := Init(testDBPath, false, "", "")
    if err != nil {
        t.Fatalf("Failed to initialize database: %v", err)
    }

    // Insert test data
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

    // Query data
    records, err := QueryRecords(Table, "Device001", "", "")
    if err != nil {
        t.Fatalf("Failed to query records: %v", err)
    }

    if len(records) == 0 {
        t.Fatal("No records found")
    }

    // Verify data
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
    // Similar structure for auth table testing
    // ... see full test file for details ...
}
```

## 6.5 Data Queue Implementation

The queue provides fault tolerance for database failures:

### Queue Structure

```go
// From queue/queue.go:32-35
type Queue struct {
    queueDir string  // Directory for queue files
    mutex    sync.Mutex  // Thread safety
}
```

### Creating Queue

```go
// From queue/queue.go:38-47
func NewQueue(queueDir string) (*Queue, error) {
    if err := os.MkdirAll(queueDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create queue directory: %v", err)
    }

    return &Queue{
        queueDir: queueDir,
    }, nil
}
```

### Enqueue Operation

```go
// From queue/queue.go:50-70
func (q *Queue) Enqueue(data interface{}) error {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    // Create unique filename using timestamp
    filename := fmt.Sprintf("%d.json", time.Now().UnixNano())
    filepath := filepath.Join(q.queueDir, filename)

    // Serialize data to JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal data: %v", err)
    }

    // Write to file (durable)
    if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
        return fmt.Errorf("failed to write to queue: %v", err)
    }

    return nil
}
```

**Why File-Based Queue?**

```go
// Memory-based queue: Fast but lost on crash
// File-based queue: Survives crashes, slower but reliable

// Design choice: Reliability over speed
// For IoT, reliability is more important than speed
// (Network interruption is usually longer than local I/O)
```

### Dequeue Operation

```go
// From queue/queue.go:73-115
func (q *Queue) Dequeue() (interface{}, error) {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    // Read directory
    files, err := os.ReadDir(q.queueDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read queue directory: %v", err)
    }

    // Find oldest JSON file
    var targetFile os.DirEntry
    for _, file := range files {
        if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
            targetFile = file
            break  // First file = oldest (sorted by name)
        }
    }

    if targetFile == nil {
        return nil, nil  // Queue empty
    }

    // Read file
    filepath := filepath.Join(q.queueDir, targetFile.Name())
    jsonData, err := os.ReadFile(filepath)
    if err != nil {
        return nil, fmt.Errorf("failed to read queue file: %v", err)
    }

    // Deserialize
    var data interface{}
    if err := json.Unmarshal(jsonData, &data); err != nil {
        return nil, fmt.Errorf("failed to unmarshal data: %v", err)
    }

    // Delete file
    if err := os.Remove(filepath); err != nil {
        return nil, fmt.Errorf("failed to remove queue file: %v", err)
    }

    return data, nil
}
```

### Queue Processing

```go
// From queue/queue.go:144-185
// Background worker that processes queued items

func (q *Queue) ProcessQueue(processFunc func(interface{}) error) {
    go func() {
        for {
            // Check queue size
            size, err := q.Size()
            if err != nil {
                log.Printf("Failed to get queue size: %v", err)
                time.Sleep(5 * time.Second)
                continue
            }

            if size == 0 {
                time.Sleep(5 * time.Second)
                continue
            }

            // Dequeue item
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

            // Process with provided function
            if err := processFunc(data); err != nil {
                log.Printf("Failed to process queue data: %v", err)

                // Re-enqueue on failure (for retry)
                if err := q.Enqueue(data); err != nil {
                    log.Printf("Failed to re-enqueue data: %v", err)
                }

                time.Sleep(5 * time.Second)
            }
        }
    }()
}
```

### Queue Tests

```go
// From queue/queue_test.go
func TestEnqueueAndDequeue(t *testing.T) {
    queueDir := "./test_queue"
    defer os.RemoveAll(queueDir)

    q, err := NewQueue(queueDir)
    if err != nil {
        t.Fatalf("Failed to create queue: %v", err)
    }

    // Enqueue test data
    testData := map[string]string{
        "key1": "value1",
        "key2": "value2",
    }

    if err := q.Enqueue(testData); err != nil {
        t.Fatalf("Failed to enqueue data: %v", err)
    }

    // Verify size
    size, err := q.Size()
    if err != nil {
        t.Fatalf("Failed to get queue size: %v", err)
    }
    if size != 1 {
        t.Errorf("Expected queue size 1, got %d", size)
    }

    // Dequeue and verify
    dequeuedData, err := q.Dequeue()
    if err != nil {
        t.Fatalf("Failed to dequeue data: %v", err)
    }

    dequeuedMap, ok := dequeuedData.(map[string]any)
    if !ok {
        t.Fatalf("Expected map[string]any, got %T", dequeuedData)
    }

    if dequeuedMap["key1"] != "value1" || dequeuedMap["key2"] != "value2" {
        t.Errorf("Dequeued data does not match enqueued data")
    }

    // Verify empty
    size, err = q.Size()
    if err != nil {
        t.Fatalf("Failed to get queue size: %v", err)
    }
    if size != 0 {
        t.Errorf("Expected queue size 0 after dequeue, got %d", size)
    }
}

func TestQueuePersistence(t *testing.T) {
    // Test that queue survives process restart
    queueDir := "./test_queue"
    defer os.RemoveAll(queueDir)

    // First instance: Enqueue data
    q1, err := NewQueue(queueDir)
    if err != nil {
        t.Fatalf("Failed to create queue: %v", err)
    }

    testData := "persistent data"
    if err := q1.Enqueue(testData); err != nil {
        t.Fatalf("Failed to enqueue data: %v", err)
    }

    // Second instance: Should see queued data
    q2, err := NewQueue(queueDir)
    if err != nil {
        t.Fatalf("Failed to create queue: %v", err)
    }

    size, err := q2.Size()
    if err != nil {
        t.Fatalf("Failed to get queue size: %v", err)
    }
    if size != 1 {
        t.Errorf("Expected queue size 1 after re-creating queue, got %d", size)
    }

    // Dequeue with second instance
    dequeuedData, err := q2.Dequeue()
    if err != nil {
        t.Fatalf("Failed to dequeue data: %v", err)
    }

    dequeuedStr, ok := dequeuedData.(string)
    if !ok {
        t.Fatalf("Expected string, got %T", dequeuedData)
    }

    if dequeuedStr != testData {
        t.Errorf("Dequeued data does not match enqueued data")
    }
}
```

## 6.6 Chapter Summary

This chapter covered the core module implementations.

**Key Takeaways:**

1. **MQTT Client**:
   - Handles EdgeX message subscription
   - Automatic reconnection and session persistence
   - Batch processing for performance

2. **Object Pool**:
   - Reduces GC pressure with map reuse
   - Critical for high-throughput scenarios
   - Always clear maps before returning to pool

3. **Database Module**:
   - Wraps sfsDb with initialization
   - Composite index on (deviceName, timestamp)
   - Batch insert with retry logic

4. **Data Queue**:
   - File-based for crash survival
   - Background processing worker
   - Automatic retry on failure

**What's Next:**

➡️ Next: [Chapter 7: Performance Optimization Implementation](./07-Chap7-Performance-Optimization.md)

---

**Code Reading Tips:**

1. **Start with structures**: Understand data structures before logic
2. **Trace the flow**: Follow data from input to output
3. **Note error handling**: How does code handle failures?
4. **Check tests**: Tests reveal intended behavior