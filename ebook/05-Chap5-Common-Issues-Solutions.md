# Chapter 5: Common Issues and Solutions

## 5.1 Memory Overflow Troubleshooting

Memory issues are the most common problem in edge deployments. This section provides systematic approaches to diagnose and resolve them.

### Symptom: sfsEdgeStore Crashes with OOM (Out of Memory)

**Diagnosis Steps:**

```bash
# 1. Check system logs for OOM killer messages
dmesg | grep -i "out of memory"
# Or on systemd systems:
journalctl | grep -i "killed"

# 2. Check memory usage before crash
free -h
# Total: 2Gi, Used: 1.9Gi, Free: 100Mi ← Problem!

# 3. Check sfsEdgeStore memory limit (if using systemd)
cat /proc/$(pidof sfsedgestore)/status | grep VmRSS
```

**Common Causes and Solutions:**

| Cause | Diagnosis | Solution |
|-------|-----------|----------|
| **sfsDb cache too large** | Check `OpenFilesCacheCapacity` | Reduce in config, use `edge` scenario |
| **Too many open files** | Check `ulimit -n` | Increase file descriptor limit |
| **Memory leak** | Check RSS growth over time | Update to latest version |
| **System RAM too low** | `free -h` | Upgrade hardware or reduce other services |

**Configuration Fix:**

```json
{
  "DBScenario": "edge",
  "DBCustomOptions": {
    "OpenFilesCacheCapacity": 64,
    "BlockCacheCapacity": 8388608
  }
}
```

**System Tuning:**

```bash
# Add to /etc/security/limits.conf
sfsedgestore soft nofile 65536
sfsedgestore hard nofile 65536

# Add to /etc/sysctl.conf
vm.swappiness = 10
```

### Symptom: Memory Usage Grows Over Time

**Diagnosis:**

```bash
# Monitor memory usage over time
while true; do
    date
    ps aux | grep sfsedgestore | grep -v grep
    sleep 60
done >> memory_log.txt

# Analyze growth pattern
# If RSS grows linearly → memory leak
# If RSS plateaus → normal cache behavior
```

**Common Causes:**

1. **LevelDB compaction backlog**
   - Symptom: Memory grows, then suddenly drops
   - Solution: Normal behavior, ensure sufficient disk I/O

2. **Goroutine leak**
   - Symptom: Goroutine count keeps increasing
   - Solution: Check for goroutine leaks in MQTT reconnection logic

3. **Database cache growing unbounded**
   - Symptom: Memory grows without bound
   - Solution: Set `BlockCacheCapacity` explicitly

**Fix Example:**

```json
{
  "DBScenario": "edge",
  "DBCustomOptions": {
    "BlockCacheCapacity": 16777216
  }
}
```

### Symptom: High Memory on Raspberry Pi

**Diagnosis:**

```bash
# Check available memory
free -h

# Check which processes use memory
ps aux --sort=-%mem | head -10

# Typical Pi 4 (4GB) memory allocation:
# - OS: ~500MB
# - EdgeX: ~800MB
# - Your app: ~300MB
# - sfsEdgeStore: ~50MB
# - Buffers: ~200MB
```

**Optimized Configuration for 1GB Raspberry Pi:**

```json
{
  "DBScenario": "edge",
  "DBCustomOptions": {
    "OpenFilesCacheCapacity": 32,
    "BlockCacheCapacity": 8388608,
    "WriteBuffer": 2097152
  }
}
```

## 5.2 Network Interruption Recovery

Network instability is a fact of life at the edge. sfsEdgeStore is designed to handle interruptions gracefully.

### Symptom: Data Loss During Network Outage

**Problem:** When MQTT connection drops, messages from EdgeX are lost.

**Solution:** sfsEdgeStore implements automatic reconnection and local queueing:

```go
// From mqtt/client.go - Connection handling
opts.SetAutoReconnect(true)           // Enable auto-reconnect
opts.SetMaxReconnectInterval(5 * time.Minute)  // Exponential backoff
opts.SetCleanSession(false)           // Don't lose subscriptions
```

**Queue for Recovery:**

```go
// When database write fails, data goes to local queue
if err := database.BatchInsertWithRetry(...) != nil {
    // Enqueue for later processing
    dataQueue.Enqueue(records)
}
```

**Recovery Process:**

```
Network Outage Timeline:

T0: Network goes down
T1: MQTT connection lost
T2: sfsEdgeStore detects failure, enables local queue
T3: Network restored
T4: MQTT reconnected
T5: Queue processor retries failed writes
T6: All data recovered
```

**Configuration for Better Recovery:**

```json
{
  "MQTTConnectionOptions": {
    "AutoReconnect": true,
    "MaxReconnectInterval": 300,
    "ConnectTimeout": 30
  },
  "QueueSettings": {
    "Enabled": true,
    "RetryInterval": "5s",
    "MaxRetries": 10
  }
}
```

### Symptom: MQTT Reconnection Fails Repeatedly

**Diagnosis:**

```bash
# Check MQTT broker status
systemctl status mosquitto
# Or
docker ps | grep mosquitto

# Check network connectivity
ping mqtt.broker.local

# Check broker logs
journalctl -u mosquitto -n 50
```

**Common Causes:**

| Cause | Diagnosis | Solution |
|-------|-----------|----------|
| **Broker down** | `systemctl status mosquitto` | Restart broker |
| **Wrong credentials** | Check broker auth logs | Update credentials |
| **Firewall blocking** | `telnet broker 1883` | Open firewall port |
| **Network partition** | Multiple gateways affected | Check network infrastructure |

### Symptom: Database Lock Errors

**Error Message:** `"Error: database is locked"`

**Cause:** Multiple processes accessing the same database, or previous crash leaving lock file.

**Solution:**

```bash
# 1. Check for running sfsEdgeStore instances
ps aux | grep sfsedgestore
pgrep -a sfsedgestore

# 2. If multiple instances, kill extras
pkill -f sfsedgestore
# Or
kill -9 $(pgrep sfsedgestore)

# 3. Remove stale lock file
rm -f /var/lib/sfsedgestore/data/LOCK

# 4. Restart
systemctl restart sfsedgestore
```

**Prevention:**

```json
{
  "DBPath": "/var/lib/sfsedgestore/data",
  "DBExclusive": true
}
```

## 5.3 Performance Optimization Techniques

### Symptom: Slow Query Response

**Diagnosis:**

```bash
# Enable query timing
curl -w "Time: %{time_total}s\n" \
  "http://localhost:8081/api/readings?deviceName=Device001&limit=100"

# Check database size
du -sh /var/lib/sfsedgestore/data

# Check for fragmentation
# (LevelDB doesn't fragment like SQLite, but compaction backlog can affect performance)
```

**Optimization Techniques:**

**1. Ensure Proper Index Usage:**

```go
// Verify composite index is used
// Key: deviceName + timestamp
// Query should specify deviceName for index prefix match
```

**2. Limit Result Set:**

```bash
# Add limit parameter
curl "http://localhost:8081/api/readings?deviceName=Device001&limit=1000"
```

**3. Use Time Range for Faster Queries:**

```bash
# Narrow time range = fewer records to scan
curl "http://localhost:8081/api/readings?deviceName=Device001&startTime=2024-01-15T00:00:00Z&endTime=2024-01-15T23:59:59Z"
```

### Symptom: High CPU Usage

**Diagnosis:**

```bash
# Check CPU usage over time
top -b -n 5 -d 1 | grep sfsedgestore

# Check for compaction storms
# (LevelDB uses CPU during compaction)
ls -la /var/lib/sfsedgestore/data/*.sst | wc -l
```

**Solutions:**

**1. Reduce Compaction Frequency:**

```json
{
  "DBScenario": "edge",
  "DBCustomOptions": {
    "WriteBuffer": 8388608,
    "MaxGrandParentOverlap": 10485760
  }
}
```

**2. Enable Compression (trades CPU for I/O):**

```json
{
  "DBCompression": true
}
```

**3. Limit HTTP Requests (if HTTP is CPU bottleneck):**

```json
{
  "HTTPRateLimit": 1000
}
```

## 5.4 Data Corruption Recovery

### Symptom: Database Won't Open

**Diagnosis:**

```bash
# Check for corruption
/var/lib/sfsedgestore/data/
ls -la
# If MANIFEST or CURRENT files missing → corruption

# Check for disk errors
dmesg | grep sda
smartctl -a /dev/sda
```

**Recovery Steps:**

```bash
# 1. Backup corrupted database
cp -r /var/lib/sfsedgestore/data /var/lib/sfsedgestore/data.corrupted

# 2. Try repair (LevelDB can repair some corruption)
./sfsedgestore --repair /var/lib/sfsedgestore/data

# 3. If repair fails, restore from backup
systemctl stop sfsedgestore
rm -rf /var/lib/sfsedgestore/data
cp -r /var/backups/sfsedgestore/last/ /var/lib/sfsedgestore/data
chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore/data
systemctl start sfsedgestore
```

### Symptom: Missing Data After Restart

**Diagnosis:**

```bash
# Check if data was queued during shutdown
ls -la /var/lib/sfsedgestore/data_queue/
# If files present, they weren't processed

# Check logs for shutdown sequence
journalctl -u sfsedgestore | grep "Shutting down"
```

**Prevention:**

```bash
# Always use graceful shutdown
sudo systemctl stop sfsedgestore
# Wait for completion (max 30 seconds)

# Or use API shutdown
curl -X POST http://localhost:8081/api/admin/shutdown
```

## 5.5 EdgeX Integration Issues

### Symptom: sfsEdgeStore Not Receiving EdgeX Events

**Diagnosis:**

```bash
# 1. Verify EdgeX is publishing
mosquitto_sub -t "edgex/events" -v
# Should see JSON messages

# 2. Verify sfsEdgeStore is subscribed
# Check logs
grep -i "subscribed" /var/log/sfsedgestore/sfsedgestore.log

# 3. Check topic matching
# EdgeX publishes to: edgex/events
# sfsEdgeStore subscribes to: edgex/events
# Must EXACTLY match!
```

**Common Issues:**

| Issue | Diagnosis | Solution |
|-------|-----------|----------|
| **Topic mismatch** | `grep topic` | Ensure exact match including leading/trailing slashes |
| **QoS mismatch** | `grep QoS` | sfsEdgeStore accepts QoS 0, 1, 2 |
| **Format mismatch** | Check EdgeX docs | sfsEdgeStore expects EdgeX event format |
| **Broker not reachable** | `ping broker` | Check network/firewall |

### Symptom: Invalid JSON from EdgeX

**Error:** `"Failed to parse message: invalid character..."`

**Diagnosis:**

```bash
# Capture raw message
mosquitto_sub -t "edgex/events" | head -1 | jq .

# Check for non-UTF8 characters
hexdump -C message.bin | grep -v "^[0-9a-f]*  "

# Common issues:
# - Binary metadata
# - Wrong encoding
# - Truncated message
```

**Solution:**

```bash
# Option 1: Filter in EdgeX
# Add to app-service-config.yaml:
DeviceNames: "Device1,Device2"

# Option 2: Update sfsEdgeStore to handle format
# Check for updates
git pull
go build
```

## 5.6 Startup and Shutdown Issues

### Symptom: sfsEdgeStore Won't Start

**Diagnosis:**

```bash
# 1. Check configuration syntax
./sfsedgestore --validate-config /etc/sfsedgestore/config.json

# 2. Check port availability
ss -tlnp | grep 8081

# 3. Check data directory permissions
ls -la /var/lib/sfsedgestore/

# 4. Check logs
journalctl -u sfsedgestore -n 50 --no-pager
```

**Common Solutions:**

| Error | Cause | Solution |
|-------|-------|----------|
| `Port already in use` | Another process on 8081 | Change port or kill other process |
| `Permission denied` | Bad directory permissions | `chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore` |
| `Config parse error` | Invalid JSON | Validate JSON syntax |
| `Database lock` | Another instance running | `pkill sfsedgestore` then restart |

### Symptom: Startup Takes Too Long

**Diagnosis:**

```bash
# Measure startup time
time ./sfsedgestore
# Should be < 1 second
```

**Common Causes:**

1. **Large database recovery**
   - Symptom: First start after crash/power loss
   - Solution: Normal, wait for recovery to complete

2. **DNS resolution timeout**
   - Symptom: Waiting for MQTT broker hostname
   - Solution: Use IP address instead of hostname

3. **Disk I/O slow**
   - Symptom: All operations slow
   - Solution: Check disk health, move to SSD

## 5.7 Troubleshooting Tools

### Health Check Endpoint

```bash
# Basic health
curl http://localhost:8081/health

# Detailed metrics
curl http://localhost:8081/metrics

# Response example:
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "mqtt_connected": true,
  "database_size_mb": 2.5,
  "memory_usage_mb": 20.85
}
```

### Log Analysis

```bash
# Follow logs in real-time
journalctl -u sfsedgestore -f

# Search for errors
journalctl -u sfsedgestore | grep -i error

# Search for specific patterns
journalctl -u sfsedgestore | grep -E "(MQTT|database|error)"
```

### Performance Monitoring

```bash
# CPU and memory
top -p $(pgrep sfsedgestore)

# I/O stats
iostat -x 1 5

# Network stats
netstat -an | grep 1883
```

## 5.8 Getting Help

If you can't resolve the issue:

1. **Collect diagnostic information:**
```bash
# Create diagnostic bundle
mkdir -p diagnostics
cp /etc/sfsedgestore/config.json diagnostics/
cp -r /var/lib/sfsedgestore/data/*.log diagnostics/ 2>/dev/null
journalctl -u sfsedgestore > diagnostics/logs.txt
ps aux > diagnostics/process.txt
free -h > diagnostics/memory.txt
df -h > diagnostics/disk.txt
tar -czf diagnostics.tar.gz diagnostics/
```

2. **Check documentation:**
   - GitHub Issues: https://github.com/liaoran123/sfsEdgeStore/issues
   - This book (Chapter 5-8)

3. **Contact support:**
   - Technical support details in Chapter 8

## 5.9 Chapter Summary

This chapter covered common issues and their solutions.

**Key Takeaways:**

1. **Memory issues** are usually solved by proper `edge` scenario configuration and system limits.

2. **Network interruptions** are handled gracefully by sfsEdgeStore's auto-reconnect and queue features.

3. **Performance issues** often require query optimization and proper indexing.

4. **Data corruption** can be recovered from backups or repaired by LevelDB.

5. **Always collect diagnostics** before asking for help.

**Common Issues Quick Reference:**

| Issue | Quick Fix |
|-------|-----------|
| OOM | Reduce `OpenFilesCacheCapacity`, use `edge` scenario |
| Data loss | Enable queue, check logs |
| Slow queries | Add time range, limit results |
| Won't start | Validate config, check port, check permissions |
| EdgeX not working | Verify topic match, broker connectivity |

**What's Next:**

➡️ Next: [Chapter 6: Core Module Source Code Analysis](./06-Chap6-Core-Modules-Source-Code.md)

---

**Remember:**

> "The most important troubleshooting skill is knowing how to read logs."

Spend time getting familiar with:
- Log locations
- Log levels
- Common error patterns

This will save hours when issues arise.