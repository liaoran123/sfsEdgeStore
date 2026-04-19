# Chapter 2: LevelDB/sfsDb Database Principles

## 2.1 LSM-Tree Architecture Explained

To understand why sfsEdgeStore achieves its exceptional performance, we need to understand the underlying storage engine: LSM-Tree (Log-Structured Merge-Tree). This chapter provides a deep dive into LSM-Tree architecture and explains why it excels at write-heavy workloads typical of IoT scenarios.

### What is LSM-Tree?

LSM-Tree is a data structure optimized for write-intensive workloads. Originally proposed by O'Neil et al. in 1996, it has become the foundation of many modern embedded databases including LevelDB, RocksDB, Cassandra, and sfsDb.

**Core Insight:**

> LSM-Tree converts random writes into sequential writes, dramatically improving write throughput on storage devices with sequential-write performance characteristics.

### LSM-Tree Components

A typical LSM-Tree implementation consists of multiple components:

```
┌─────────────────────────────────────────────────────────────┐
│                     LSM-Tree Architecture                     │
│                                                              │
│  ┌─────────────┐                                            │
│  │   Write     │ ← All writes go here first                 │
│  │   Buffer    │   (MemTable - In-memory)                   │
│  └──────┬──────┘                                            │
│         │ When buffer is full:                              │
│         ▼                                                    │
│  ┌─────────────┐                                            │
│  │   L0        │ ← First level on disk                     │
│  │   SSTable   │   (Immutable sorted files)                  │
│  └──────┬──────┘                                            │
│         │ Periodic compaction:                              │
│         ▼                                                    │
│  ┌─────────────┐                                            │
│  │   L1        │ ← Larger sorted runs                       │
│  └──────┬──────┘                                            │
│         │                                                   │
│         ▼                                                   │
│  ┌─────────────┐                                            │
│  │   L2        │ ← Even larger files                        │
│  └──────┬──────┘                                            │
│         │                                                   │
│         ▼                                                   │
│      ...    (Levels grow exponentially: L1=10MB, L2=100MB)   │
└─────────────────────────────────────────────────────────────┘
```

### The Write Path: From Memory to Disk

Let's trace what happens when you write data to a LevelDB/sfsDb:

**Step 1: Write to MemTable (In-Memory)**

```go
// All writes go to an in-memory buffer called MemTable
// This is a sorted skip-list or red-black tree
// Write operation: O(log N) where N is buffer size

// Example from sfsDb storage layer
func (db *DB) Put(key, value []byte) error {
    // 1. Write to WAL (Write-Ahead Log) for durability
    db.wal.Write(key, value)

    // 2. Insert into MemTable (sorted in memory)
    db.memTable.Put(key, value)

    // 3. Return immediately - no disk sync yet
    return nil
}
// Latency: ~0.1-0.5 ms (memory speed)
```

**Step 2: MemTable Flush (When Full)**

When the MemTable reaches its size limit (default 4MB in LevelDB):

```
MemTable (4MB) → L0 SSTable (4MB) on disk

┌─────────────────┐      ┌─────────────────┐
│    MemTable     │ ───→ │   L0 File       │
│   (sorted)      │ flush│  (immutable)    │
└─────────────────┘      └─────────────────┘
```

**Step 3: Compaction (Background Process)**

As L0 fills up, files are merged with L1, and so on:

```
Level 0 (4MB files)  ─┐
                       │ compact
Level 1 (40MB files) ─┴──→ Level 1 (fewer, larger files)

Each level is ~10x larger than the previous.
LevelDB max file size: 2MB per file
Level 0: 4 files max (8MB)
Level 1: 40MB total
Level 2: 400MB total
... and so on
```

### Why LSM-Tree Excels at Writes

**1. Sequential Write Optimization**

Traditional B-tree storage:
```yaml
# Random 4KB write on HDD: ~10ms
# Random 4KB write on SSD: ~0.1ms

Problem: Each write might need to:
1. Read the page containing the key
2. Modify the page
3. Write the page back
4. Update index structures
Total: Multiple I/O operations per write
```

LSM-Tree storage:
```yaml
# Sequential write: ~50MB/s on HDD, ~500MB/s on SSD

Solution:
1. Append to MemTable (memory)
2. Batch flush to L0 (sequential)
3. Background compaction (sequential)

Total: Near-optimal sequential write performance!
```

**2. Write Amplification is Acceptable**

LSM-Tree does more total writes due to compaction, but each write is sequential:

```
Traditional B-tree:
  1 logical write = 4-10 physical I/O operations
  Mixed read/write: unpredictable performance

LSM-Tree:
  1 logical write = 1 sequential write to MemTable
  Compaction: 10 sequential reads + 10 sequential writes = 1 new file
  Result: Predictable, high throughput
```

**3. Bloom Filter for Fast Lookups**

Without Bloom filters, finding a key in LevelDB would require checking multiple levels:

```
Without Bloom filter:
  GET key → Check MemTable → Check L0 → Check L1 → ... → NotFound

With Bloom filter:
  GET key → Check Bloom filter (O(1)) → "definitely not" or "maybe in L1"
  Bloom filter: False positive rate ~1%, no false negatives
```

Implementation from LevelDB:

```go
// Each SSTable has a Bloom filter
// Key is hashed multiple times, bits set in filter
type FilterPolicy struct {
    bits []byte  // Bitmap
    k    int     // Number of hash functions (typically 10)
}

func (f *FilterPolicy) Contains(key []byte) bool {
    h := hash(key)
    for i := 0; i < f.k; i++ {
        idx := (h + uint64(i)*f.b) % uint64(len(f.bits)*8)
        if f.bits[idx/8]&(1<<(idx%8)) == 0 {
            return false  // Definitely not in table
        }
    }
    return true   // Probably in table
}
```

### LSM-Tree Trade-offs

LSM-Tree isn't perfect. Understanding its trade-offs helps with configuration:

| Aspect | LSM-Tree Advantage | LSM-Tree Disadvantage |
|--------|-------------------|----------------------|
| **Write Speed** | Sequential, batched | More total writes (amplification) |
| **Read Speed** | Bloom filters help | May need multiple levels |
| **Memory** | Configurable buffer | Bloom filters need memory |
| **Storage** | Excellent compression | Multiple versions stored |
| **Compaction** | Background, non-blocking | Occasional CPU spikes |

**The Compaction Trade-off:**

```
Write-heavy workload (IoT scenario):
  Logical writes: 1,000,000
  Physical writes: 1,200,000 (20% amplification)
  But all sequential → Fast!

Read-heavy workload (analytics):
  May need to check multiple levels
  Bloom filter saves most lookups
  Still fast, but not optimal
```

## 2.2 Why LevelDB Excels at Write Operations

LevelDB's design choices make it particularly suitable for IoT/edge data patterns. Let's analyze specific design decisions and their impact.

### Write Operation Internals

When you call LevelDB's Write():

```go
// From LevelDB source (simplified)
func (db *DB) Write(batch *WriteBatch) error {
    // 1. Serialize write to internal format
    data := batch.Serialize()

    // 2. Append to WAL ( durability )
    db.wal.Append(data)

    // 3. Insert into MemTable ( memory )
    db.memTable.Insert(batch)

    // 4. If MemTable full → Trigger flush
    if db.memTable.Size() >= db.options.MemTableSize {
        go db.CompactMemTable()  // Background flush
    }

    return nil
}
// Total time: ~0.1-0.5ms (memory speed)
```

**Key Insight: The write returns immediately after memory operations. Disk I/O happens asynchronously via background compaction.**

### Batch Write Optimization

LevelDB's batch operations are highly optimized:

```go
// Batch writes are atomic and sequential
batch := NewWriteBatch()
batch.Put("device1/temp", []byte("25.5"))
batch.Put("device1/humidity", []byte("60"))
batch.Put("device2/temp", []byte("26.0"))
batch.Put("device2/pressure", []byte("1013"))

db.Write(batch)  // All 4 puts are sequential on disk

// Performance comparison:
Single puts:  4 × 0.3ms = 1.2ms
Batch put:    1 × 0.5ms = 0.5ms
// Batch is 2.4x faster!
```

### Time-Series Write Patterns

IoT data has a specific pattern: sequential timestamps from multiple devices. LevelDB handles this exceptionally well:

```
Traditional storage (B-tree):
  Write #1: device1@T1 → Page 1
  Write #2: device2@T1 → Page 50 (random!)
  Write #3: device1@T2 → Page 1 (already full, split needed)

LevelDB:
  Write #1: device1@T1 → MemTable (append)
  Write #2: device2@T1 → MemTable (append, same region)
  Write #3: device1@T2 → MemTable (append)

Result: Sequential I/O, optimal for SSD and HDD alike!
```

### sfsDb: LevelDB Optimized for Edge

sfsDb builds on LevelDB's foundation with specific optimizations for edge/IoT workloads:

**1. Simplified Configuration**

LevelDB has 50+ tuning parameters. sfsDb provides preset scenarios:

```go
// sfsDb scenario configuration
type Scenario string

const (
    ScenarioEdge      Scenario = "edge"       // IoT edge gateways
    ScenarioStandard   Scenario = "standard"  // General workloads
    ScenarioAnalytics Scenario = "analytics"  // Read-heavy
)

// Scenario options - pre-tuned for common use cases
type ScenarioOptions struct {
    WriteBuffer       int  // MemTable size
    OpenFilesCapacity int  // Files to keep in memory
    BlockSize         int  // SSTable block size
    Compression       bool // Enable compression
}

func GetScenarioOptions(scenario Scenario) ScenarioOptions {
    switch scenario {
    case ScenarioEdge:
        return ScenarioOptions{
            WriteBuffer:       4 * 1024 * 1024,  // 4MB
            OpenFilesCapacity: 100,
            BlockSize:         4096,              // 4KB
            Compression:       true,
        }
    case ScenarioStandard:
        return ScenarioOptions{
            WriteBuffer:       4 * 1024 * 1024,
            OpenFilesCapacity: 1000,
            BlockSize:         4096,
            Compression:       true,
        }
    case ScenarioAnalytics:
        return ScenarioOptions{
            WriteBuffer:       8 * 1024 * 1024,
            OpenFilesCapacity: 5000,
            BlockSize:         8192,
            Compression:       false,  // Faster reads
        }
    }
}
```

**2. Time-Series Index Optimization**

sfsDb optimizes for the common IoT query pattern: device + time range:

```go
// Composite key design for efficient time-series queries
// Key format: deviceName (padded 64 chars) + timestamp (nanoseconds)

// For device="Device001", timestamps:
// Key: "Device001                                                       /1704067200000000000"
// Key: "Device001                                                       /1704067200001000000"
// Key: "Device001                                                       /1704067200002000000"
// ...

// Range query for Device001 between T1 and T2:
// Seek to "Device001" + T1
// Scan forward until key > "Device001" + T2
// Result: Only touches relevant data, no full table scan
```

**3. Automatic Key Padding**

```go
// sfsDb utility function for device name padding
// Ensures consistent 64-character device names for index optimization
func FormatDeviceName(deviceName string) string {
    const maxLength = 64
    if len(deviceName) >= maxLength {
        return deviceName[:maxLength]
    }
    // Pad with spaces for consistent index entries
    return deviceName + strings.Repeat(" ", maxLength-len(deviceName))
}

// Result: All keys have same length
// Benefit: Fixed-size comparison, faster range scans
```

## 2.3 Index Design for Time-Series Data

Proper index design is crucial for query performance. In this section, we'll cover how to design indexes that optimize for common IoT query patterns.

### Understanding LevelDB Index Structure

LevelDB uses a hierarchical index:

```
┌─────────────────────────────────────────────────────────┐
│                    Index Hierarchy                        │
│                                                           │
│  Level 0: ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐               │
│           │ L0-0│ │ L0-1│ │ L0-2│ │ L0-3│ (max 4 files)│
│           └─────┘ └─────┘ └─────┘ └─────┘               │
│                │         │         │                     │
│                └─────────┴─────────┘                     │
│                         │                                 │
│                         ▼ Compaction                     │
│  Level 1: ┌─────────────────────┐                       │
│           │     L1-0 (10MB)     │                       │
│           └─────────────────────┘                       │
│                         │                               │
│                         ▼ Compaction                     │
│  Level 2: ┌─────────────────────────┐                   │
│           │     L2-0 (100MB)        │                   │
│           └─────────────────────────┘                   │
│                         │                               │
│                        ...                              │
└─────────────────────────────────────────────────────────┘

Each level is 10x larger than the previous.
L1 max: 10 files × 10MB = 100MB total
L2 max: 100 files × 10MB = 1GB total
...
```

### Index Design Principles for Time-Series

**Principle 1: Primary Key Should Support Your Most Common Query**

Most IoT queries follow this pattern:
```sql
SELECT * FROM readings
WHERE deviceName = 'XYZ'
  AND timestamp BETWEEN T1 AND T2
ORDER BY timestamp DESC
```

Therefore, your key design should be:
```go
// ❌ Wrong: Timestamp first (can't range scan by device)
key = timestamp + deviceName

// ✅ Correct: Device first, timestamp second
key = deviceName (padded) + timestamp

// This enables:
// 1. Efficient lookup by device (prefix search)
// 2. Efficient time range within device (range scan)
// 3. Sorted results by time (natural ordering)
```

**Principle 2: Fixed-Width Fields Improve Performance**

```go
// Variable width (bad for scanning):
deviceName = "Device001"      // 9 bytes
deviceName = "Sensor-12-AB"   // 13 bytes
// Comparison: must compare character by character

// Fixed width (good for scanning):
deviceName = "Device001                           "  // 64 bytes
deviceName = "Sensor-12-AB                        "  // 64 bytes
// Comparison: memcmp() of 64 bytes, faster!
```

**Principle 3: Use Appropriate Data Types**

```go
// ❌ String timestamps (slow comparison)
key = deviceName + ":" + "2024-01-01T00:00:00Z"

// ✅ Numeric timestamps (fast comparison)
key = deviceName + ":" + "1704067200000000000"  // Unix nanoseconds
// Numeric comparison is single CPU instruction!

// Example from sfsEdgeStore:
data["timestamp"] = reading.Origin  // int64, nanoseconds
```

### sfsEdgeStore Index Implementation

Here's how sfsEdgeStore implements these principles:

```go
// database/database.go:78-101
// Creating the primary key index

fields := map[string]any{
    "id":         "",        // Unique reading ID
    "deviceName": "",        // Device identifier
    "reading":    "",        // Reading name (e.g., "temperature")
    "value":      0.0,      // Reading value
    "valueType":  "",        // Data type (Float32, Int64, etc.)
    "baseType":   "",        // Base type
    "timestamp":  int64(0),  // Nanosecond timestamp
    "metadata":   "",        // Optional metadata JSON
}

// Create composite primary key: deviceName + timestamp
primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
if err != nil {
    return fmt.Errorf("failed to create primary key: %v", err)
}

// The order matters! deviceName first for prefix queries
primaryKey.AddFields("deviceName", "timestamp")

// Create the index
if err := Table.CreateIndex(primaryKey); err != nil {
    if err.Error() != "index already exists" {
        return fmt.Errorf("failed to create primary key index: %v", err)
    }
}
```

### Query Optimization Examples

**Query 1: Get all readings from one device (last hour)**

```go
// ❌ Without proper index: Full table scan
records, _ := table.Query("SELECT * FROM readings")  // O(n)

// ✅ With composite index: Prefix scan
// Seek to deviceName + startTime
startKey := map[string]any{
    "deviceName": "Device001                               ",  // Padded!
    "timestamp":  now.Add(-1*time.Hour).UnixNano(),
}
endKey := map[string]any{
    "deviceName": "Device001                               ",
    "timestamp":  now.UnixNano(),
}

// LevelDB will:
// 1. Use Bloom filter to confirm data exists
// 2. Seek to first matching key
// 3. Scan sequentially until end key
// Result: O(records returned) instead of O(total records)
```

**Query 2: Get latest reading from each device**

```go
// Reverse scan with prefix
// Start from highest timestamp, go backwards
// Stop when deviceName changes

// Implementation:
iter, _ := table.SearchRange(nil, nil)  // Full range
iter.Prefix("Device001")                  // Device prefix only
iter.Reverse()                           // Descending order
iter.Limit(1)                            // Top 1 result
```

### Storage Layout and Performance

The way data is organized on disk affects I/O performance:

```
sfsEdgeStore Storage Layout:

/data/
├── LOG              # LevelDB write-ahead log
├── LOG.old          # Previous log (after rotation)
├── CURRENT          # Points to current MANIFEST
├── MANIFEST-000001  # Describes all files and levels
├── 000003.log       # MemTable dump (L0 SSTable)
├── 000004.log       # Another L0 SSTable
├── 000005.sst       # L1 SSTable (after compaction)
├── 000006.sst       # L1 SSTable
└── ...

Total size for 18,681 readings: ~250 KB
(Average: ~14 bytes per reading with compression)
```

**Why So Small?**

1. **Key-value separation**: Keys are repeated in SSTables, but sfsDb uses key prefix compression
2. **Compression**: Enabled by default in edge scenario (Snappy compression, 60-70% ratio)
3. **Efficient timestamps**: int64 (8 bytes) vs ISO string (30 bytes)
4. **Fixed padding**: 64 chars = 64 bytes max, but common names are much shorter

## 2.4 Chapter Summary

This chapter provided a deep understanding of how LSM-Tree based storage engines work and why they excel at edge IoT workloads.

**Key Takeaways:**

1. **LSM-Tree converts random writes to sequential**: By buffering writes in memory and batch-flushing to disk, LSM-Tree achieves near-optimal sequential write performance.

2. **Bloom filters enable fast lookups**: Without Bloom filters, reads would need to check every level. With Bloom filters, most lookups are resolved in O(1).

3. **Compaction is the trade-off**: LSM-Tree does more total writes (amplification) but achieves better write throughput and predictable performance.

4. **Key design is crucial**: For time-series data, use `deviceName + timestamp` as the key to enable efficient prefix and range queries.

5. **sfsDb optimizes LevelDB for edge**: Simplified configuration, preset scenarios, and IoT-specific optimizations make sfsDb the ideal storage for sfsEdgeStore.

**What's Next:**

In Chapter 3, we'll put this theory into practice. You'll learn how to deploy sfsEdgeStore in your environment, from a simple 5-minute installation to production-grade configurations.

**Hands-On Preview:**

- Install sfsEdgeStore in 5 minutes
- Configure for your specific hardware
- Connect to EdgeX Foundry
- Verify data flow end-to-end

---

➡️ Next: [Chapter 3: sfsEdgeStore Quick Start](./03-Chap3-sfsEdgeStore-Quick-Start.md)

---

**Deep Dive Questions:**

1. What happens when the MemTable is full?
2. How does compaction affect write throughput during heavy loads?
3. Why does fixed-width key padding improve range scan performance?
4. How do Bloom filters achieve O(1) lookup with minimal memory?

*Understanding these details will help you diagnose performance issues in production.*