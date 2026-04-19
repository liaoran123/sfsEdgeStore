# Chapter 7: Performance Optimization Implementation

## 7.1 Performance Optimization Philosophy

Performance optimization in sfsEdgeStore follows a clear philosophy:

**"Optimize for the common case, don't break the edge case."**

This means:
1. **Measure before optimizing** - Use benchmarks to identify bottlenecks
2. **Focus on the hot path** - Optimize the code that runs most frequently
3. **Trade-offs are intentional** - Every optimization has a cost

### The Three Pillars of Optimization

```
┌─────────────────────────────────────────────────────────────────┐
│                    Performance Optimization                       │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
│  │   Memory    │  │    CPU      │  │     I/O     │           │
│  │  Efficiency │  │  Efficiency  │  │  Throughput │           │
│  └─────────────┘  └─────────────┘  └─────────────┘           │
│         │                │                │                     │
│         ▼                ▼                ▼                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
│  │ Object Pool │  │Batch Process│  │ Sequential  │           │
│  │             │  │             │  │    Write    │           │
│  └─────────────┘  └─────────────┘  └─────────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

## 7.2 Object Pool Implementation Details

The object pool is the first line of defense against memory bloat and GC pressure.

### Why Object Pool?

In Go, memory allocation and garbage collection have costs:

```go
// Without pool - allocations per message
for _, reading := range readings {
    data := make(map[string]any)  // 1 allocation
    data["key"] = value
    // Use data...
    // GC must collect when data is unreachable
}
// 1000 readings = 1000 allocations + 1000 collections

// With pool - reuse existing memory
for _, reading := range readings {
    data := objPool.GetMap()      // 0 allocations (usually)
    data["key"] = value
    // Use data...
    objPool.PutMap(data)          // 0 collections (reuse)
}
// 1000 readings = ~10-50 allocations (pool sized)
```

### sync.Pool Mechanics

```go
// From mqtt/mapPool.go - Full implementation

var objPool = NewObjectPool()  // Global singleton

type objectPool struct {
    mapPool sync.Pool  // Go's pool implementation
}

func NewObjectPool() *objectPool {
    return &objectPool{
        mapPool: sync.Pool{
            New: func() interface{} {
                // Called when pool is empty
                // Should return a "new" object
                return make(map[string]any)
            },
        },
    }
}

func (p *objectPool) GetMap() map[string]any {
    // Get from pool (may allocate if empty)
    m := p.mapPool.Get().(map[string]any)

    // CRITICAL: Clear before use!
    // Pool may contain dirty data from previous use
    for k := range m {
        delete(m, k)
    }
    return m
}

func (p *objectPool) PutMap(m map[string]any) {
    // CRITICAL: Clear before return!
    // Prevents data leakage between uses
    for k := range m {
        delete(m, k)
    }
    p.mapPool.Put(m)
}
```

### Memory Savings Analysis

```go
// Test: Process 10,000 messages with pool vs without

// WITHOUT pool:
// Allocations: 10,000 maps × ~200 bytes = 2 MB
// GC pressure: 10,000 collections
// Time: ~50ms overhead

// WITH pool:
// Allocations: ~20 maps (pool sized at ~20)
// GC pressure: ~20 collections
// Time: ~2ms overhead (25x improvement)

// BUT: Pool must be cleared properly!
// Failure to clear = data leakage (security issue!)
```

### Benchmark: Object Pool

```go
// From mqtt/mqtt_test.go - Run with: go test -bench=BenchmarkObjectPool

func BenchmarkObjectPool(b *testing.B) {
    // Without pool
    for i := 0; i < b.N; i++ {
        data := make(map[string]any)
        data["key1"] = "value1"
        data["key2"] = 42
        _ = data
    }
}

func BenchmarkObjectPoolWithPool(b *testing.B) {
    // With pool
    for i := 0; i < b.N; i++ {
        data := objPool.GetMap()
        data["key1"] = "value1"
        data["key2"] = 42
        objPool.PutMap(data)
    }
}

// Typical results:
// BenchmarkObjectPool-8           5000000    243 ns/op
// BenchmarkObjectPoolWithPool-8  20000000     89 ns/op
// Improvement: ~2.7x faster
```

## 7.3 Batch Processing Strategies

Batch processing reduces per-item overhead by grouping operations.

### Why Batch?

```
Single insert: 1000 items × 1ms = 1000ms
Batch insert:   1000 items ÷ 100 batch × 2ms = 20ms
Improvement:   50x faster!
```

### Batch Insert Implementation

```go
// From database/database.go - Batch insert with retry

func BatchInsertWithRetry(tbl *engine.Table, records []*map[string]any,
                          maxRetries int, retryInterval time.Duration) error {

    // Retry loop
    for i := 0; i < maxRetries; i++ {
        // Attempt batch insert
        _, err := tbl.BatchInsertNoInc(records)

        if err == nil {
            return nil  // Success!
        }

        // Log failure
        log.Printf("Batch insert failed (attempt %d/%d): %v",
                   i+1, maxRetries, err)

        // Wait before retry (exponential backoff optional)
        if i < maxRetries-1 {
            time.Sleep(retryInterval)
        }
    }

    return fmt.Errorf("failed after %d attempts", maxRetries)
}
```

**Why Not Batch Forever?**

```
Batch size trade-off:
- Small batches: Low latency, high overhead
- Large batches: High latency, low overhead

Optimal: Balance between latency and throughput

sfsEdgeStore defaults:
- batchSize: 100 messages
- batchInterval: 5 seconds
- Result: < 5 second data lag, good throughput
```

### Monitoring Batch Performance

```go
// From monitor/monitor.go - Metrics collection

type InternalApplicationMetrics struct {
    MQTTMessagesReceived  atomic.Int64  // Total received
    MQTTMessagesProcessed atomic.Int64  // Total processed
    HTTPRequests          atomic.Int64  // HTTP API calls
    DatabaseOperations    atomic.Int64  // DB operations
    Errors                atomic.Int64  // Errors
}

// These are lock-free counters (atomic)
// No mutex needed = no contention = high performance
```

## 7.4 Concurrency Safety Design

Go's concurrency primitives must be used carefully to avoid performance degradation.

### Lock-Free Atomic Operations

```go
// From monitor/monitor.go - Atomic counters

type InternalApplicationMetrics struct {
    // atomic.Int64 is lock-free
    // Uses CPU instructions (LOCK XADD on x86)
    // No mutex needed = no blocking

    MQTTMessagesReceived  atomic.Int64
    MQTTMessagesProcessed atomic.Int64
    HTTPRequests          atomic.Int64
    DatabaseOperations    atomic.Int64
    Errors                atomic.Int64
}

// Usage: Simple increment, no locks needed
func (m *Monitor) IncrementMQTTMessagesReceived() {
    m.metrics.Application.MQTTMessagesReceived.Add(1)
}

// Compare to mutex approach (BAD):
type BadMetrics struct {
    mutex   sync.Mutex
    counter int64
}

func (m *BadMetrics) Increment() {
    m.mutex.Lock()
    m.counter++
    m.mutex.Unlock()  // Blocking! Other goroutines wait
}
```

### Mutex Usage Guidelines

```go
// When to use mutex vs atomic:

// USE ATOMIC for:
// - Simple counters (increment/decrement)
// - Boolean flags
// - No compound operations

// USE MUTEX for:
// - Complex data structures
// - Compound operations (check-then-act)
// - Read-write patterns (RWMutex)

// sfsEdgeStore examples:

// 1. Atomic counter (good)
type Monitor struct {
    errors atomic.Int64  // Lock-free
}

// 2. Mutex for complex state (necessary)
type Queue struct {
    mutex    sync.Mutex  // Protects queue state
    queueDir string
}

// 3. RWMutex for read-heavy patterns (optional optimization)
type Cache struct {
    rwmutex sync.RWMutex  // Allows concurrent readers
    data    map[string]string
}
```

### Goroutine Management

```go
// From main.go - Graceful shutdown

// 1. Start components
dataQueue, err = queue.NewQueue("./data_queue")
// ...

// 2. Handle shutdown signals
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit  // Block until signal

log.Println("Shutting down adapter...")

// 3. Stop components gracefully
if agentInstance != nil {
    agentInstance.Stop()
}
if retentionManager != nil {
    retentionManager.Stop()
}
// ... etc

// 4. Cleanup
time.Sleep(5 * time.Second)  // Allow cleanup
log.Println("Adapter exited")
```

## 7.5 Memory Optimization

### Memory Layout of sfsEdgeStore

```
Memory Usage Breakdown:

Total: ~20-25 MB

├── Go Runtime: ~8 MB
│   ├── Goroutine stacks: ~2 MB
│   └── Go scheduler: ~1 MB
│
├── sfsDb: ~10 MB
│   ├── MemTable (WriteBuffer): ~4 MB
│   ├── Block Cache: ~4 MB
│   └── Open Files Cache: ~2 MB
│
├── MQTT Client: ~2 MB
│   ├── Client buffers: ~1 MB
│   └── Message handlers: ~1 MB
│
└── HTTP Server: ~1 MB
    └── Request handlers: ~1 MB
```

### Controlling Memory Usage

```json
// Configuration to limit memory

{
  "DBScenario": "edge",
  "DBCustomOptions": {
    // Limit MemTable size
    "WriteBufferSize": 2097152,  // 2MB instead of 4MB

    // Limit file cache
    "OpenFilesCacheCapacity": 50,  // Fewer files in memory

    // Limit block cache
    "BlockCacheCapacity": 4194304  // 4MB instead of 8MB
  }
}
```

### Memory Benchmark

```go
// From tests - Memory usage measurement

func TestMemoryUsage(t *testing.T) {
    // Baseline
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)

    // Run workload
    for i := 0; i < 10000; i++ {
        data := objPool.GetMap()
        data["key"] = "value"
        objPool.PutMap(data)
    }

    // After
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)

    alloc := m2.Alloc - m1.Alloc
    t.Logf("Memory allocated: %d bytes (%.2f MB)",
           alloc, float64(alloc)/1024/1024)
}
```

## 7.6 Time Series Optimization

### Time Granularity Processing

Time-series data benefits from aggregated queries:

```go
// From time/granularity.go - Time formatting

type TimeGranularity string

const (
    TimeGranularityMillisecond TimeGranularity = "millisecond"
    TimeGranularitySecond      TimeGranularity = "second"
    TimeGranularityMinute      TimeGranularity = "minute"
    TimeGranularityHour        TimeGranularity = "hour"
    TimeGranularityDay         TimeGranularity = "day"
)

// Format time by granularity
func FormatTimeByGranularity(t time.Time, granularity TimeGranularity) string {
    switch granularity {
    case TimeGranularitySecond:
        return t.Format("2006-01-02 15:04:05")
    case TimeGranularityMinute:
        return t.Format("2006-01-02 15:04:00")
    case TimeGranularityHour:
        return t.Format("2006-01-02 15:00:00")
    case TimeGranularityDay:
        return t.Format("2006-01-02")
    // ...
    }
}
```

### Adjusting Time to Granularity Boundary

```go
// From time/time_query.go - Align time to boundaries

func AdjustTimeToGranularity(t time.Time, granularity TimeGranularity) time.Time {
    switch granularity {
    case TimeGranularitySecond:
        return time.Date(t.Year(), t.Month(), t.Day(),
                        t.Hour(), t.Minute(), t.Second(), 0, t.Location())

    case TimeGranularityMinute:
        return time.Date(t.Year(), t.Month(), t.Day(),
                        t.Hour(), t.Minute(), 0, 0, t.Location())

    case TimeGranularityHour:
        return time.Date(t.Year(), t.Month(), t.Day(),
                        t.Hour(), 0, 0, 0, t.Location())

    case TimeGranularityDay:
        return time.Date(t.Year(), t.Month(), t.Day(),
                        0, 0, 0, 0, t.Location())
    }
}
```

## 7.7 Profiling and Benchmarking

### Using pprof

```go
// Add to main.go for CPU/memory profiling

import _ "net/http/pprof"

func main() {
    // Start pprof server
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // ... rest of main ...
}
```

```bash
# Collect CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Collect memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# View in browser
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile
```

### Benchmark Tests

```go
// From database/database_test.go

func BenchmarkQueryRecords(b *testing.B) {
    // Setup: Insert test data first
    // ...

    // Reset timer to exclude setup
    b.ResetTimer()

    // Benchmark
    for i := 0; i < b.N; i++ {
        records, _ := QueryRecords(Table, "Device001", "", "")
        _ = records
    }
}

func BenchmarkBatchInsert(b *testing.B) {
    records := generateTestRecords(100)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = BatchInsertWithRetry(Table, records, 3, time.Second)
    }
}
```

```bash
# Run benchmarks
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkQueryRecords -benchmem

# Run with longer duration
go test -bench=. -benchtime=10s -benchmem
```

## 7.8 Production Performance Checklist

Before deploying to production:

```markdown
## Performance Checklist

### Memory
□ Memory usage < 50% of available RAM
□ No memory leaks (RSS stable over 24h)
□ Object pool properly sized
□ GC pause acceptable (< 100ms)

### CPU
□ CPU usage < 50% under normal load
□ No CPU spikes (except during compaction)
□ Batch processing enabled
□ Rate limiting configured if needed

### I/O
□ Disk I/O < 50% utilization
□ No I/O wait issues
□ Compression enabled (unless CPU-bound)
□ Compaction not causing I/O spikes

### Network
□ MQTT connection stable
□ Reconnection logic working
□ No message loss during reconnects
□ TLS overhead acceptable

### Query Performance
□ Query latency < 100ms for typical queries
□ Index being used (check logs)
□ Time range filters applied
□ Result limits set
```

## 7.9 Chapter Summary

This chapter covered the performance optimization techniques used in sfsEdgeStore.

**Key Takeaways:**

1. **Object Pool**: Reduces allocations and GC pressure by reusing objects

2. **Batch Processing**: Groups operations to reduce per-item overhead

3. **Atomic Operations**: Lock-free counters for high-performance metrics

4. **Memory Control**: Configurable limits prevent unbounded growth

5. **Benchmarking**: Measure before optimizing, use pprof to identify bottlenecks

**Optimization Priority:**

```
1. Object Pool        - Immediate impact, low risk
2. Batch Processing   - Major throughput improvement
3. Atomic Counters    - Eliminate contention
4. Memory Limits      - Prevent OOM crashes
5. Profiling         - Find remaining bottlenecks
```

**What's Next:**

➡️ Next: [Chapter 8: Product Introduction and Commercial Services](./08-Chap8-Product-Commercial-Services.md)

---

**Remember:**

> "The best optimization is the one you don't need to do."

Start with a well-designed architecture (like sfsEdgeStore), use it as intended, and only optimize when you have actual performance problems measured with real benchmarks.