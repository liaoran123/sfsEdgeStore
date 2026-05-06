package monitor

// Metrics 监控指标（返回给前端）
type Metrics struct {
	System      SystemMetrics      `json:"system"`
	Application ApplicationMetrics `json:"application"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsage float64 `json:"memory_usage_mb"`
	Goroutines  int     `json:"goroutines"`
	Uptime      int64   `json:"uptime_seconds"`
}

// ApplicationMetrics 应用指标
type ApplicationMetrics struct {
	MQTTMessagesReceived  int64 `json:"mqtt_messages_received"`
	MQTTMessagesProcessed int64 `json:"mqtt_messages_processed"`
	MQTTMessagesFiltered  int64 `json:"mqtt_messages_filtered"`
	TotalRecordsStored    int64 `json:"total_records_stored"`
}
