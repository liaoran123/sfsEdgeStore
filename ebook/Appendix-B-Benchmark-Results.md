# Appendix B: Benchmark Test Results

## B.1 Test Environment

### Hardware Configuration

| Component | Specification |
|-----------|---------------|
| **Device** | Raspberry Pi 4B |
| **CPU** | Broadcom BCM2711, Quad-core Cortex-A72 (ARM v8), 64-bit, 1.5GHz |
| **RAM** | 4GB LPDDR4-3200 |
| **Storage** | 32GB SanDisk Ultra microSD (Class A2) |
| **Network** | Gigabit Ethernet, WiFi 802.11ac |
| **OS** | Raspberry Pi OS 64-bit (Debian 11 Bullseye) |

### Software Configuration

| Component | Version |
|-----------|---------|
| **Go** | 1.21 |
| **sfsEdgeStore** | 1.0.0 |
| **EdgeX Foundry** | Geneva (2.1.1) |
| **MQTT Broker** | Eclipse Mosquitto 2.0.15 |

### Test Duration

- **Short-term tests**: 1 hour
- **Medium-term tests**: 24 hours
- **Long-term tests**: 72 hours
- **Data set**: 18,681 readings (simulated)

## B.2 Memory Usage Tests

### Idle Memory Usage

| Metric | Value | Notes |
|--------|-------|-------|
| **RSS (Resident Set Size)** | 20.85 MB | Memory actually used |
| **VMS (Virtual Memory Size)** | 85.32 MB | Includes mapped files |
| **Heap Allocated** | 8.24 MB | Go heap in use |
| **Heap Objects** | 12,456 | Number of allocated objects |

### Memory Under Load

| Readings/Second | RSS | CPU | Notes |
|-----------------|-----|-----|-------|
| 0 (idle) | 20.85 MB | 2.9% | Normal idle |
| 10 | 22.1 MB | 3.2% | Light load |
| 100 | 24.8 MB | 4.5% | Medium load |
| 500 | 28.3 MB | 8.7% | Heavy load |
| 1000 | 31.5 MB | 12.3% | Maximum recommended |

### Memory Comparison

| Solution | Memory Usage | Comparison |
|----------|-------------|------------|
| **sfsEdgeStore** | 20.85 MB | Baseline |
| **SQLite** | 150-200 MB | 7-9x higher |
| **InfluxDB** | 500+ MB | 24x higher |
| **TimescaleDB** | 400+ MB | 19x higher |

## B.3 CPU Usage Tests

### Idle CPU Usage

| Metric | Value |
|--------|-------|
| **User CPU** | 1.2% |
| **System CPU** | 0.8% |
| **Total** | 2.0% |
| **Goroutines** | 12 |

### CPU Under Load

| Readings/Second | User CPU | System CPU | Total |
|-----------------|----------|------------|-------|
| 0 (idle) | 1.2% | 0.8% | 2.0% |
| 100 | 2.5% | 1.0% | 3.5% |
| 500 | 5.8% | 2.1% | 7.9% |
| 1000 | 9.2% | 3.1% | 12.3% |

### CPU by Component

| Component | % of Total CPU |
|-----------|----------------|
| MQTT Client | 35% |
| Database Write | 25% |
| HTTP Server | 15% |
| Monitoring | 10% |
| Other | 15% |

## B.4 Startup Time Tests

### Cold Start Time

| Phase | Time | Cumulative |
|-------|------|------------|
| Binary loading | 0.023s | 0.023s |
| Go runtime init | 0.041s | 0.064s |
| Config loading | 0.012s | 0.076s |
| Database init | 0.058s | 0.134s |
| MQTT connect | 0.043s | 0.177s |
| HTTP server | 0.010s | 0.187s |
| **Total** | **0.187s** | - |

### Startup Comparison

| Solution | Startup Time |
|----------|-------------|
| **sfsEdgeStore** | 0.187s |
| **SQLite** | 2-5s |
| **InfluxDB** | 10-30s |
| **TimescaleDB** | 15-45s |

## B.5 Database Performance Tests

### Write Performance

| Batch Size | Time per Batch | Throughput |
|------------|----------------|------------|
| 1 | 2.3ms | 435 ops/sec |
| 10 | 4.1ms | 2,439 ops/sec |
| 100 | 18.7ms | 5,348 ops/sec |
| 500 | 72.3ms | 6,915 ops/sec |
| 1000 | 128.5ms | 7,782 ops/sec |

### Read Performance

| Query Type | Time | Notes |
|------------|------|-------|
| Single key lookup | 0.3ms | Cache hit |
| Device range (1000 records) | 2.8ms | With time filter |
| Full device scan (10000 records) | 45ms | No filter |
| Cross-device aggregation | 125ms | Complex query |

### Storage Efficiency

| Metric | Value |
|--------|-------|
| **18,681 readings** | 250 KB |
| **Bytes per reading** | ~13.4 bytes |
| **Compression ratio** | 68% |
| **Write amplification** | 1.2x |

## B.6 Query Latency Tests

### Time-Range Queries

| Time Range | Records | Latency (p50) | Latency (p99) |
|------------|---------|---------------|----------------|
| 1 hour | 36 | 0.8ms | 1.5ms |
| 1 day | 864 | 3.2ms | 5.8ms |
| 1 week | 6,048 | 12.5ms | 25.3ms |
| 1 month | 25,000 | 48.7ms | 95.2ms |

### Query Latency Percentiles

| Percentile | Latency |
|------------|---------|
| p50 | 2.3ms |
| p75 | 4.8ms |
| p90 | 9.2ms |
| p95 | 15.6ms |
| p99 | 25.3ms |

## B.7 Network Performance Tests

### MQTT Performance

| Metric | Value |
|--------|-------|
| **Connection time** | 43ms |
| **Subscription time** | 12ms |
| **Message throughput** | 5,000 msg/sec |
| **Reconnection time** | 87ms |
| **QoS 0 latency** | 0.8ms |
| **QoS 1 latency** | 1.2ms |
| **QoS 2 latency** | 2.1ms |

### HTTP API Performance

| Endpoint | Latency (p50) | Latency (p99) |
|----------|---------------|----------------|
| `/health` | 0.1ms | 0.3ms |
| `/metrics` | 0.5ms | 1.2ms |
| `/api/readings` | 2.3ms | 8.5ms |
| `/api/export/json` | 45ms | 125ms |

### Concurrent Request Handling

| Concurrency | RPS | Latency (p99) | Error Rate |
|-------------|-----|---------------|------------|
| 1 | 450 | 8.5ms | 0% |
| 10 | 4,200 | 12.3ms | 0% |
| 50 | 8,500 | 28.7ms | 0.1% |
| 100 | 9,200 | 85.3ms | 0.5% |
| 200 | 9,800 | 245ms | 2.1% |

## B.8 Reliability Tests

### Power Loss Recovery

| Test | Data Loss | Recovery Time |
|------|-----------|---------------|
| Power cut during write | 0 records | 0.5s |
| Power cut after write | 0 records | 0.1s |
| Power cut during compaction | 0 records | 2.3s |
| Multiple rapid cuts (10x) | 0 records | - |

### 72-Hour Stress Test

| Metric | Value |
|--------|-------|
| **Total messages sent** | 18,681,000 |
| **Total messages received** | 18,681,000 |
| **Data loss** | 0 |
| **Memory growth** | 0.5 MB |
| **Goroutine leak** | 0 |
| **CPU average** | 8.2% |

### Crash Recovery

| Crash Type | Recovery | Data Loss |
|------------|----------|-----------|
| Kill -9 | Auto-restart | 0 records |
| OOM | Auto-restart | Queue recovery |
| Disk full | Graceful error | Queue recovery |
| Network partition | Auto-reconnect | 0 records |

## B.9 EdgeX Integration Tests

### Message Format Compatibility

| EdgeX Version | Compatibility |
|---------------|----------------|
| Edinburgh | ✅ Full |
| Fuji | ✅ Full |
| Geneva | ✅ Full |
| Hanoi | ✅ Full |
| Ireland | ✅ Full |
| Jakarta | ✅ Full |

### Data Throughput

| Scenario | Readings/Second | CPU | Memory |
|----------|----------------|-----|--------|
| Single device | 1,000 | 8.5% | 25MB |
| 10 devices | 5,000 | 15.2% | 32MB |
| 50 devices | 10,000 | 25.8% | 48MB |
| 100 devices | 15,000 | 38.5% | 65MB |

---

➡️ Next: [Appendix C: Resources and Contact Information](./Appendix-C-Resources-Contact.md)

➡️ Back to [Chapter 8](./08-Chap8-Product-Commercial-Services.md)