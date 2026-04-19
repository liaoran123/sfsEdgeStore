# Performance Optimization and Best Practices

## Overview

This chapter introduces sfsEdgeStore's performance optimization strategies, time series processing, performance testing, and best practices for production environments.

## Time Granularity Processing

### Time Granularity Types

```go
// time/granularity.go:9-22
type TimeGranularity string

const (
	TimeGranularityMillisecond TimeGranularity = "millisecond"
	TimeGranularityMicrosecond TimeGranularity = "microsecond"
	TimeGranularitySecond      TimeGranularity = "second"
	TimeGranularityMinute      TimeGranularity = "minute"
	TimeGranularityHour        TimeGranularity = "hour"
	TimeGranularityDay         TimeGranularity = "day"
	TimeGranularityWeek        TimeGranularity = "week"
	TimeGranularityMonth       TimeGranularity = "month"
	TimeGranularityQuarter     TimeGranularity = "quarter"
	TimeGranularityYear        TimeGranularity = "year"
)
```

### Time Formatting

```go
// time/granularity.go:50-77
func FormatTimeByGranularity(t time.Time, granularity TimeGranularity) string {
	switch granularity {
	case TimeGranularityMillisecond:
		return t.Format("2006-01-02 15:04:05.000")
	case TimeGranularityMicrosecond:
		return t.Format("2006-01-02 15:04:05.000000")
	case TimeGranularitySecond:
		return t.Format("2006-01-02 15:04:05")
	case TimeGranularityMinute:
		return t.Format("2006-01-02 15:04:00")
	case TimeGranularityHour:
		return t.Format("2006-01-02 15:00:00")
	case TimeGranularityDay:
		return t.Format("2006-01-02")
	case TimeGranularityWeek:
		_, week := t.ISOWeek()
		return t.Format("2006-") + fmt.Sprintf("W%02d", week)
	case TimeGranularityMonth:
		return t.Format("2006-01")
	case TimeGranularityQuarter:
		quarter := GetQuarter(t)
		return t.Format("2006-") + fmt.Sprintf("Q%d", quarter)
	case TimeGranularityYear:
		return t.Format("2006")
	default:
		return t.Format("2006-01-02 15:04:05")
	}
}
```

### Adjust Time to Granularity Boundary

```go
// time/time_query.go:97-129
func AdjustTimeToGranularity(t time.Time, granularity TimeGranularity) time.Time {
	switch granularity {
	case TimeGranularityMillisecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000*1000000, t.Location())
	case TimeGranularityMicrosecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000*1000, t.Location())
	case TimeGranularitySecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	case TimeGranularityMinute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	case TimeGranularityHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case TimeGranularityDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case TimeGranularityWeek:
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 0, 0, 0, 0, t.Location())
	case TimeGranularityMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case TimeGranularityQuarter:
		quarter := (int(t.Month())-1)/3 + 1
		return time.Date(t.Year(), time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, t.Location())
	case TimeGranularityYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}
```

## Time Range Query

### Query Options Structure

```go
// time/time_query.go:10-16
type TimeRangeQueryOptions struct {
	FieldName       string
	StartTime       time.Time
	EndTime         time.Time
	TimeGranularity TimeGranularity
	Inclusive       bool
}
```

### Create Query Options

```go
// time/time_query.go:19-27
func NewTimeRangeQueryOptions(fieldName string, startTime, endTime time.Time, granularity TimeGranularity) *TimeRangeQueryOptions {
	return &TimeRangeQueryOptions{
		FieldName:       fieldName,
		StartTime:       startTime,
		EndTime:         endTime,
		TimeGranularity: granularity,
		Inclusive:       true,
	}
}
```

### Execute Time Range Query

```go
// time/time_query.go:30-41
func SearchTimeRange(table *engine.Table, options *TimeRangeQueryOptions) (*engine.TableIter, error) {
	iter, err := table.SearchRange(nil, &map[string]any{options.FieldName: options.StartTime}, &map[string]any{options.FieldName: options.EndTime})
	if err != nil {
		return nil, err
	}

	return iter, nil
}
```

### Query by Time Granularity

```go
// time/time_query.go:50-59
func SearchTimeRangeWithGranularity(table *engine.Table, fieldName string, startTime, endTime time.Time, granularity TimeGranularity) (*engine.TableIter, error) {
	adjustedStart, adjustedEnd := AdjustTimeRangeByGranularity(startTime, endTime, granularity)
	options := NewTimeRangeQueryOptions(fieldName, adjustedStart, adjustedEnd, granularity)
	return SearchTimeRange(table, options)
}
```

## Performance Testing

### Basic Testing

```go
package database

import "testing"

func TestInit(t *testing.T) {
    err := Init("./testdb", false, "", "")
    if err != nil {
        t.Fatalf("Failed to init database: %v", err)
    }
}

func TestQueryRecords(t *testing.T) {
    records, err := QueryRecords(Table, "TestDevice", "", "")
    if err != nil {
        t.Fatalf("Failed to query records: %v", err)
    }
    t.Logf("Got %d records", len(records))
}
```

### Subtests

```go
func TestBatchInsert(t *testing.T) {
    tests := []struct {
        name      string
        batchSize int
    }{
        {"small", 10},
        {"medium", 100},
        {"large", 1000},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            records := generateTestRecords(tt.batchSize)
            err := BatchInsertWithRetry(Table, records, 3, time.Second)
            if err != nil {
                t.Fatalf("Batch insert failed: %v", err)
            }
        })
    }
}
```

### Benchmark Testing

```go
func BenchmarkQueryRecords(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = QueryRecords(Table, "TestDevice", "", "")
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

### Running Benchmark Tests

```bash
# Run benchmark tests
go test -bench=. -benchmem

# Run specific benchmark test
go test -bench=BenchmarkQueryRecords -benchmem

# Run for longer time
go test -bench=. -benchtime=10s -benchmem
```

### Performance Analysis Tools

#### pprof

```go
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Main program logic
}
```

```bash
# Collect CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Collect memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Collect goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

#### trace Tool

```go
import "runtime/trace"

func main() {
    f, err := os.Create("trace.out")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    trace.Start(f)
    defer trace.Stop()
    
    // Main program logic
}
```

```bash
# Run and generate trace
go test -trace=trace.out

# View trace
go tool trace trace.out
```

## Load Testing

```go
package main

import (
    "net/http"
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    
    requestCount := 1000
    concurrency := 10
    
    semaphore := make(chan struct{}, concurrency)
    startTime := time.Now()
    
    for i := 0; i < requestCount; i++ {
        wg.Add(1)
        semaphore <- struct{}{}
        
        go func() {
            defer wg.Done()
            defer func() { <-semaphore }()
            
            resp, err := client.Get("http://localhost:8081/api/readings?deviceName=TestDevice")
            if err != nil {
                log.Printf("Request failed: %v", err)
                return
            }
            defer resp.Body.Close()
        }()
    }
    
    wg.Wait()
    duration := time.Since(startTime)
    
    log.Printf("Requests: %d", requestCount)
    log.Printf("Duration: %v", duration)
    log.Printf("RPS: %.2f", float64(requestCount)/duration.Seconds())
}
```

## Performance Optimization Strategies

### 1. Object Pool Optimization

```go
// mqtt/mapPool.go
type MapPool struct {
	pool sync.Pool
}

func NewMapPool() *MapPool {
	return &MapPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]any, 16)
			},
		},
	}
}

func (p *MapPool) Get() map[string]any {
	return p.pool.Get().(map[string]any)
}

func (p *MapPool) Put(m map[string]any) {
	for k := range m {
		delete(m, k)
	}
	p.pool.Put(m)
}
```

**Advantages:**
- Reduce GC pressure
- Avoid frequent memory allocation
- Improve performance

### 2. Batch Processing

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

**Best Practices:**
- Batch size: 100-1000 records
- Retry count: 3-5 times
- Retry interval: 1-2 seconds

### 3. Concurrency Safety

```go
// Use atomic package for lock-free counting
type InternalApplicationMetrics struct {
	MQTTMessagesReceived  atomic.Int64
	MQTTMessagesProcessed atomic.Int64
	HTTPRequests          atomic.Int64
	DatabaseOperations    atomic.Int64
	Errors                atomic.Int64
}
```

**Advantages:**
- Lock-free design, higher performance
- Avoid deadlock risks
- Suitable for high concurrency scenarios

### 4. Index Optimization

```go
// Use composite primary key to optimize queries
primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
primaryKey.AddFields("deviceName", "timestamp")
if err := Table.CreateIndex(primaryKey); err != nil {
	// Handle error
}
```

**Best Practices:**
- Put high-cardinality fields first
- Put time fields last
- Avoid too many indexes

## Configuration Optimization

### Database Configuration

```json
{
  "DBScenario": "edge",
  "DBPath": "./data",
  "DBUseEncryption": false,
  "DBCompression": true
}
```

**Scenario Configurations:**
- `edge`: Edge scenario, optimize for write performance
- `standard`: Standard scenario, balance read and write
- `analytics`: Analytics scenario, optimize for read performance

### MQTT Configuration

```json
{
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTClientID": "sfsedgestore-001",
  "MQTTQoS": 1,
  "MQTTKeepAlive": 30
}
```

**Best Practices:**
- QoS 1 or 2 to ensure message delivery
- KeepAlive 30-60 seconds
- Use unique ClientID

## Production Environment Deployment

### System Requirements

**Minimum Configuration:**
- CPU: 2 cores
- Memory: 2 GB
- Disk: 10 GB (SSD recommended)
- OS: Linux/Windows/macOS

**Recommended Configuration:**
- CPU: 4+ cores
- Memory: 8+ GB
- Disk: 100+ GB (SSD)
- OS: Linux (Ubuntu 20.04+)

### System Parameter Tuning

```bash
# /etc/sysctl.conf
# Increase file descriptor limit
fs.file-max = 100000

# Network tuning
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_tw_reuse = 1

# Memory tuning
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
```

```bash
# /etc/security/limits.conf
sfsedgestore soft nofile 65536
sfsedgestore hard nofile 65536
sfsedgestore soft nproc 4096
sfsedgestore hard nproc 4096
```

### Containerized Deployment

#### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o sfsedgestore .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/sfsedgestore .
COPY --from=builder /app/config.json .

EXPOSE 8081

CMD ["./sfsedgestore"]
```

#### Docker Compose

```yaml
version: '3.8'

services:
  sfsedgestore:
    build: .
    ports:
      - "8081:8081"
    volumes:
      - data:/var/lib/sfsedgestore/data
      - logs:/var/log/sfsedgestore
    environment:
      - MQTT_BROKER=tcp://mosquitto:1883
    depends_on:
      - mosquitto
    restart: unless-stopped

  mosquitto:
    image: eclipse-mosquitto:2.0
    ports:
      - "1883:1883"
    volumes:
      - mosquitto-config:/mosquitto/config
      - mosquitto-data:/mosquitto/data
    restart: unless-stopped

volumes:
  data:
  logs:
  mosquitto-config:
  mosquitto-data:
```

### Kubernetes Deployment

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sfsedgestore
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sfsedgestore
  template:
    metadata:
      labels:
        app: sfsedgestore
    spec:
      containers:
      - name: sfsedgestore
        image: sfsedgestore:latest
        ports:
        - containerPort: 8081
        env:
        - name: MQTT_BROKER
          value: "tcp://mqtt-broker:1883"
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2000m"
            memory: "2Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: data
          mountPath: /var/lib/sfsedgestore/data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: sfsedgestore-data
```

### Monitoring and Alerting

#### Prometheus Metrics

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    messagesReceived = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "sfsedgestore_messages_received_total",
            Help: "Total number of messages received",
        },
    )
    httpRequests = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "sfsedgestore_http_requests_total",
            Help: "Total number of HTTP requests",
        },
    )
    databaseLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "sfsedgestore_database_latency_seconds",
            Help:    "Database operation latency",
            Buckets: prometheus.DefBuckets,
        },
    )
)

func init() {
    prometheus.MustRegister(messagesReceived)
    prometheus.MustRegister(httpRequests)
    prometheus.MustRegister(databaseLatency)
}
```

#### Backup and Recovery

```bash
#!/bin/bash
BACKUP_DIR="/var/backups/sfsedgestore"
DATA_DIR="/var/lib/sfsedgestore/data"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/sfsedgestore_$DATE.tar.gz"

mkdir -p $BACKUP_DIR

tar -czf $BACKUP_FILE -C $DATA_DIR .

find $BACKUP_DIR -name "sfsedgestore_*.tar.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE"
```

## Monitoring Metrics

### Key Monitoring Metrics

| Metric | Description | Threshold |
|--------|-------------|-----------|
| MQTT Messages Received | Number of MQTT messages received per minute | - |
| MQTT Messages Processed | Number of MQTT messages processed per minute | - |
| HTTP Requests | Number of HTTP requests per minute | 1000 |
| Database Operations | Number of database operations per minute | 5000 |
| Errors | Number of errors per minute | 10 |
| Goroutines | Current number of goroutines | - |
| Memory Usage | Current memory usage | - |

## Production Environment Best Practices

### 1. Data Backup

```bash
# Regular backup
curl -X POST -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/backup?path=./backups/$(date +%Y%m%d)"
```

### 2. Health Check

```bash
# Use monitoring script
while true; do
  curl -f http://localhost:8081/health || echo "Service unhealthy"
  sleep 60
done
```

### 3. Log Management

```json
{
  "LogLevel": "info",
  "LogFile": "./logs/sfsedgestore.log",
  "LogMaxSize": 100,
  "LogMaxBackups": 10
}
```

### 4. Resource Limits

```bash
# Use systemd to limit resources
[Service]
MemoryLimit=512M
CPUQuota=50%
```

## API Interface

### Time Processing API

```go
func FormatTimeByGranularity(t time.Time, granularity TimeGranularity) string
func AdjustTimeToGranularity(t time.Time, granularity TimeGranularity) time.Time
func AdjustTimeRangeByGranularity(startTime, endTime time.Time, granularity TimeGranularity) (time.Time, time.Time)
func NewTimeRangeQueryOptions(fieldName string, startTime, endTime time.Time, granularity TimeGranularity) *TimeRangeQueryOptions
func SearchTimeRange(table *engine.Table, options *TimeRangeQueryOptions) (*engine.TableIter, error)
func SearchTimeRangeWithGranularity(table *engine.Table, fieldName string, startTime, endTime time.Time, granularity TimeGranularity) (*engine.TableIter, error)
```

### Packaging API

```go
func NewPackageConfig(name string, version Version) *PackageConfig
func (pc *PackageConfig) Build() error
func GetBuildInfo() BuildInfo
```

## Running Tests

```bash
# Run time package tests
go test ./time -v

# Run all tests
go test ./... -v

# Run benchmark tests
go test -bench=. -benchmem

# Build for all platforms
./packaging/build.sh
```