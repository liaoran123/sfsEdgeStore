package common

import "time"

// Alert 告警信息
type Alert struct {
	Type      string    `json:"type"`      // 告警类型
	Message   string    `json:"message"`   // 告警消息
	Severity  string    `json:"severity"`  // 告警级别
	Timestamp time.Time `json:"timestamp"` // 告警时间
	Resolved  bool      `json:"resolved"`  // 是否已解决
}
