package monitor

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	System      SystemMetrics      `json:"system"`
	Application ApplicationMetrics `json:"application"`
}

type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	Goroutines  int     `json:"goroutines"`
	Uptime      int64   `json:"uptime_seconds"`
}

type ApplicationMetrics struct {
	MQTTMessagesReceived  int64 `json:"mqtt_messages_received"`
	MQTTMessagesProcessed int64 `json:"mqtt_messages_processed"`
	MQTTMessagesFiltered  int64 `json:"mqtt_messages_filtered"`
	TotalRecordsStored    int64 `json:"total_records_stored"`
	HTTPRequests          int64 `json:"http_requests"`
	DatabaseOperations    int64 `json:"database_operations"`
	Errors                int64 `json:"errors"`
	MQTTConnected         bool  `json:"mqtt_connected"`
	DataReceivedBytes     int64 `json:"data_received_bytes"`
	DataStoredBytes       int64 `json:"data_stored_bytes"`
}

type metricsData struct {
	System      SystemMetrics
	Application atomicApplicationMetrics
}

type atomicApplicationMetrics struct {
	MQTTMessagesReceived  atomic.Int64
	MQTTMessagesProcessed atomic.Int64
	MQTTMessagesFiltered  atomic.Int64
	TotalRecordsStored    atomic.Int64
	HTTPRequests          atomic.Int64
	DatabaseOperations    atomic.Int64
	Errors                atomic.Int64
	MQTTConnectionStatus  atomic.Bool
	DataReceivedBytes     atomic.Int64
	DataStoredBytes       atomic.Int64
}

type AlertThresholds struct {
	HTTPRequestsPerMinute       int64 `json:"http_requests_per_minute"`
	ErrorsPerMinute             int64 `json:"errors_per_minute"`
	DatabaseOperationsPerMinute int64 `json:"database_operations_per_minute"`
}

type AlertGroup struct {
	Type        string
	Message     string
	DeviceCount int
	Devices     []string
	LastTime    time.Time
	Severity    string
	Count       int
}

type DeviceStatus struct {
	LastActive     time.Time
	IsOnline       bool
	LastAlertTime  time.Time
	DataCount      int64
	LastDataValue  float64
	LastDataTime   time.Time
	DataHistory    []float64
	LastDataChange float64
}
