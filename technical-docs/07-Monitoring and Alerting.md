# Monitoring and Alerting

## Overview

sfsEdgeStore provides a comprehensive monitoring system for tracking system runtime status, performance metrics, and sending alert notifications in case of anomalies.

## Monitoring Module

### Monitor Structure

```go
// monitor/monitor.go:20-30
type Monitor struct {
	monitorManager  *monitor.MonitorManager
	metrics         InternalMetrics
	startTime       time.Time
	alertThresholds AlertThresholds
	alerts          []common.Alert
	lastMetrics     InternalMetrics
	lastCollectTime time.Time
	mutex           sync.Mutex
	notifier        *alert.Notifier
}
```

**Main Fields:**
- `monitorManager`: sfsDb's monitoring manager
- `metrics`: Internal monitoring metrics (using atomic types)
- `alertThresholds`: Alert threshold configuration
- `alerts`: Alert list
- `notifier`: Alert notifier

### Monitoring Metrics Structure

```go
// monitor/monitor.go:33-76
type InternalMetrics struct {
	System      SystemMetrics
	Database    DatabaseMetrics
	Application InternalApplicationMetrics
}

type SystemMetrics struct {
	CPUUsage    float64
	MemoryUsage float64
	Goroutines  int
	Uptime      int64
}

type DatabaseMetrics struct {
	KeyStats   map[int]interface{}
	IndexStats map[int]interface{}
}

type InternalApplicationMetrics struct {
	MQTTMessagesReceived  atomic.Int64
	MQTTMessagesProcessed atomic.Int64
	HTTPRequests          atomic.Int64
	DatabaseOperations    atomic.Int64
	Errors                atomic.Int64
}
```

### Create Monitoring Manager

```go
// monitor/monitor.go:86-104
func NewMonitor() *Monitor {
	return &Monitor{
		monitorManager: monitor.NewMonitorManager(),
		metrics: InternalMetrics{
			System: SystemMetrics{
				Goroutines: runtime.NumGoroutine(),
			},
			Application: InternalApplicationMetrics{},
		},
		startTime: time.Now(),
		alertThresholds: AlertThresholds{
			HTTPRequestsPerMinute:       1000,
			ErrorsPerMinute:             10,
			DatabaseOperationsPerMinute: 5000,
		},
		alerts:          []common.Alert{},
		lastCollectTime: time.Now(),
	}
}
```

**Default Alert Thresholds:**
- HTTP requests: 1000 per minute
- Errors: 10 per minute
- Database operations: 5000 per minute

### Collect Monitoring Metrics

```go
// monitor/monitor.go:127-135
func (m *Monitor) CollectMetrics() Metrics {
	m.collectSystemMetrics()
	m.collectDatabaseMetrics()
	return m.toExportedMetrics()
}
```

### Collect System Metrics

```go
// monitor/monitor.go:138-146
func (m *Monitor) collectSystemMetrics() {
	m.metrics.System.Goroutines = runtime.NumGoroutine()
	m.metrics.System.Uptime = int64(time.Since(m.startTime).Seconds())

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.metrics.System.MemoryUsage = float64(memStats.Alloc) / 1024 / 1024
}
```

### Metrics Increment Functions

```go
// monitor/monitor.go:166-188
func (m *Monitor) IncrementMQTTMessagesReceived() {
	m.metrics.Application.MQTTMessagesReceived.Add(1)
}

func (m *Monitor) IncrementMQTTMessagesProcessed() {
	m.metrics.Application.MQTTMessagesProcessed.Add(1)
}

func (m *Monitor) IncrementHTTPRequests() {
	m.metrics.Application.HTTPRequests.Add(1)
}

func (m *Monitor) IncrementDatabaseOperations() {
	m.metrics.Application.DatabaseOperations.Add(1)
}

func (m *Monitor) IncrementErrors() {
	m.metrics.Application.Errors.Add(1)
}
```

### Record Error and Trigger Alert

```go
// monitor/monitor.go:191-214
func (m *Monitor) RecordError(errorType, message string) {
	m.IncrementErrors()

	alert := common.Alert{
		Type:      errorType,
		Message:   message,
		Severity:  "critical",
		Timestamp: time.Now(),
		Resolved:  false,
	}

	m.mutex.Lock()
	m.alerts = append(m.alerts, alert)
	m.mutex.Unlock()

	log.Printf("Critical error recorded: %s - %s", errorType, message)

	if m.notifier != nil {
		m.notifier.SendAlert(alert)
	}
}
```

### Check Alerts

```go
// monitor/monitor.go:224-303
func (m *Monitor) CheckAlerts() []common.Alert {
	var newAlerts []common.Alert

	timeDiff := time.Since(m.lastCollectTime).Minutes()
	if timeDiff == 0 {
		timeDiff = 1
	}

	currentHTTP := m.metrics.Application.HTTPRequests.Load()
	currentErrors := m.metrics.Application.Errors.Load()
	currentDBOps := m.metrics.Application.DatabaseOperations.Load()

	lastHTTP := m.lastMetrics.Application.HTTPRequests.Load()
	lastErrors := m.lastMetrics.Application.Errors.Load()
	lastDBOps := m.lastMetrics.Application.DatabaseOperations.Load()

	httpRequestsPerMinute := (currentHTTP - lastHTTP) / int64(timeDiff)
	errorsPerMinute := (currentErrors - lastErrors) / int64(timeDiff)
	dbOperationsPerMinute := (currentDBOps - lastDBOps) / int64(timeDiff)

	if httpRequestsPerMinute > m.alertThresholds.HTTPRequestsPerMinute {
		newAlerts = append(newAlerts, common.Alert{
			Type:      "http_requests",
			Message:   fmt.Sprintf("HTTP requests rate too high: %d per minute", httpRequestsPerMinute),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	if errorsPerMinute > m.alertThresholds.ErrorsPerMinute {
		newAlerts = append(newAlerts, common.Alert{
			Type:      "errors",
			Message:   fmt.Sprintf("Error rate too high: %d per minute", errorsPerMinute),
			Severity:  "critical",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	if dbOperationsPerMinute > m.alertThresholds.DatabaseOperationsPerMinute {
		newAlerts = append(newAlerts, common.Alert{
			Type:      "database_operations",
			Message:   fmt.Sprintf("Database operations rate too high: %d per minute", dbOperationsPerMinute),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	m.mutex.Lock()
	m.alerts = append(m.alerts, newAlerts...)
	
	m.lastMetrics.Application.MQTTMessagesReceived.Store(m.metrics.Application.MQTTMessagesReceived.Load())
	m.lastMetrics.Application.MQTTMessagesProcessed.Store(m.metrics.Application.MQTTMessagesProcessed.Load())
	m.lastMetrics.Application.HTTPRequests.Store(m.metrics.Application.HTTPRequests.Load())
	m.lastMetrics.Application.DatabaseOperations.Store(m.metrics.Application.DatabaseOperations.Load())
	m.lastMetrics.Application.Errors.Store(m.metrics.Application.Errors.Load())
	
	m.lastCollectTime = time.Now()
	m.mutex.Unlock()

	if m.notifier != nil {
		for _, alert := range newAlerts {
			m.notifier.SendAlert(alert)
		}
	}

	return newAlerts
}
```

## Alert Notification Module

### Notifier Structure

```go
// alert/notifier.go:17-23
type Notifier struct {
	config      *config.Config
	mqttClient  mqtt.Client
	alertChan   chan common.Alert
	stopChan    chan struct{}
	isRunning   bool
}
```

### Create Notifier

```go
// alert/notifier.go:33-39
func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{
		config:    cfg,
		alertChan: make(chan common.Alert, 100),
		stopChan:  make(chan struct{}),
	}
}
```

### Start Notifier

```go
// alert/notifier.go:41-69
func (n *Notifier) Start() error {
	if !n.config.EnableAlertNotifications {
		log.Println("Alert notifications are disabled")
		return nil
	}

	if n.isRunning {
		log.Println("Alert notifier is already running")
		return nil
	}

	if len(n.config.AlertNotificationChannels) == 0 {
		log.Println("No alert notification channels configured")
		return nil
	}

	for _, channel := range n.config.AlertNotificationChannels {
		if channel == "mqtt" {
			if err := n.initMQTTClient(); err != nil {
				log.Printf("Failed to initialize MQTT client: %v", err)
			}
		}
	}

	n.isRunning = true
	go n.notificationLoop()
	log.Printf("Alert notifier started with channels: %v", n.config.AlertNotificationChannels)
	return nil
}
```

### Send Alert

```go
// alert/notifier.go:83-97
func (n *Notifier) SendAlert(alert common.Alert) {
	if !n.config.EnableAlertNotifications || !n.isRunning {
		return
	}

	if !n.shouldSendAlert(alert) {
		return
	}

	select {
	case n.alertChan <- alert:
	default:
		log.Println("Alert channel is full, dropping alert")
	}
}
```

### Check Alert Severity

```go
// alert/notifier.go:99-110
func (n *Notifier) shouldSendAlert(alert common.Alert) bool {
	severityOrder := map[string]int{
		"info":     0,
		"warning":  1,
		"critical": 2,
	}

	minSeverity := severityOrder[n.config.AlertMinSeverity]
	alertSeverity := severityOrder[alert.Severity]

	return alertSeverity >= minSeverity
}
```

### Notification Loop

```go
// alert/notifier.go:112-121
func (n *Notifier) notificationLoop() {
	for {
		select {
		case alert := <-n.alertChan:
			n.sendToAllChannels(alert)
		case <-n.stopChan:
			return
		}
	}
}
```

### Send to All Channels

```go
// alert/notifier.go:123-142
func (n *Notifier) sendToAllChannels(alert common.Alert) {
	notification := AlertNotification{
		Type:      alert.Type,
		Message:   alert.Message,
		Severity:  alert.Severity,
		Timestamp: alert.Timestamp,
		Source:    n.config.ClientID,
	}

	for _, channel := range n.config.AlertNotificationChannels {
		switch channel {
		case "mqtt":
			n.sendToMQTT(notification)
		case "webhook":
			n.sendToWebhook(notification)
		case "log":
			n.sendToLog(notification)
		}
	}
}
```

### Send to MQTT

```go
// alert/notifier.go:173-195
func (n *Notifier) sendToMQTT(notification AlertNotification) {
	if n.mqttClient == nil || !n.mqttClient.IsConnected() {
		log.Println("MQTT client not connected, cannot send alert")
		return
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal alert for MQTT: %v", err)
		return
	}

	topic := n.config.AlertMQTTTopic
	if topic == "" {
		topic = "edgex/alerts"
	}

	token := n.mqttClient.Publish(topic, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to send alert to MQTT: %v", token.Error())
	}
}
```

### Send to Webhook

```go
// alert/notifier.go:197-218
func (n *Notifier) sendToWebhook(notification AlertNotification) {
	if n.config.AlertWebhookURL == "" {
		return
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal alert for webhook: %v", err)
		return
	}

	resp, err := http.Post(n.config.AlertWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to send alert to webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Webhook returned non-success status: %d", resp.StatusCode)
	}
}
```

## HTTP Handlers

### Register Handlers

```go
// monitor/monitor.go:313-317
func (m *Monitor) RegisterHandlers() {
	http.HandleFunc("/metrics", m.handleMetrics)
	http.HandleFunc("/health", m.handleHealth)
	http.HandleFunc("/alerts", m.handleAlerts)
}
```

### Get Metrics

```go
// monitor/monitor.go:320-329
func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.CollectMetrics()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("Error encoding metrics: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
```

### Health Check

```go
// monitor/monitor.go:332-348
func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := m.CollectMetrics()

	healthStatus := map[string]interface{}{
		"status":          "healthy",
		"uptime_seconds":  metrics.System.Uptime,
		"goroutines":      metrics.System.Goroutines,
		"memory_usage_mb": metrics.System.MemoryUsage,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
		log.Printf("Error encoding health status: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
```

## Configuration Example

### Alert Configuration

```json
{
  "EnableAlertNotifications": true,
  "AlertNotificationChannels": ["mqtt", "webhook", "log"],
  "AlertMinSeverity": "warning",
  "AlertMQTTTopic": "edgex/alerts",
  "AlertWebhookURL": "https://example.com/webhook/alerts"
}
```

## API Interface

### Monitor API

```go
func NewMonitor() *Monitor
func (m *Monitor) SetNotifier(notifier *alert.Notifier)
func (m *Monitor) CollectMetrics() Metrics
func (m *Monitor) IncrementMQTTMessagesReceived()
func (m *Monitor) IncrementMQTTMessagesProcessed()
func (m *Monitor) IncrementHTTPRequests()
func (m *Monitor) IncrementDatabaseOperations()
func (m *Monitor) IncrementErrors()
func (m *Monitor) RecordError(errorType, message string)
func (m *Monitor) CheckAlerts() []common.Alert
func (m *Monitor) GetAlerts() []common.Alert
func (m *Monitor) RegisterHandlers()
```

### Notifier API

```go
func NewNotifier(cfg *config.Config) *Notifier
func (n *Notifier) Start() error
func (n *Notifier) Stop()
func (n *Notifier) SendAlert(alert common.Alert)
func (n *Notifier) GetNotifierStatus() map[string]any
```