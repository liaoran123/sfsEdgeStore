package remote

import (
	"fmt"
	"strings"
	"time"
)

// RemoteDevice 远程设备
type RemoteDevice struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	IP        string                 `json:"ip"`
	Status    string                 `json:"status"` // online, offline, error
	Tags      map[string]string      `json:"tags"`   // 设备标签
	Config    map[string]interface{} `json:"config"` // 设备配置
	LastSeen  time.Time              `json:"last_seen"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewRemoteDevice 创建远程设备
func NewRemoteDevice(id, name, ip string) *RemoteDevice {
	return &RemoteDevice{
		ID:        id,
		Name:      name,
		IP:        ip,
		Status:    "offline",
		Tags:      make(map[string]string),
		Config:    make(map[string]interface{}),
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
	}
}

// AddTag 添加标签
func (d *RemoteDevice) AddTag(key, value string) {
	d.Tags[key] = value
}

// RemoveTag 移除标签
func (d *RemoteDevice) RemoveTag(key string) {
	delete(d.Tags, key)
}

// HasTag 检查是否有标签
func (d *RemoteDevice) HasTag(key, value string) bool {
	if v, ok := d.Tags[key]; ok {
		return v == value
	}
	return false
}

// MatchTags 匹配标签条件
func (d *RemoteDevice) MatchTags(conditions map[string]string) bool {
	for key, value := range conditions {
		if !d.HasTag(key, value) {
			return false
		}
	}
	return true
}

// MatchTagExpression 匹配标签表达式
// 支持 AND 逻辑，格式: "region:shanghai AND type:pump"
func (d *RemoteDevice) MatchTagExpression(expression string) bool {
	conditions := make(map[string]string)
	parts := strings.Split(expression, "AND")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tagParts := strings.Split(part, ":")
		if len(tagParts) == 2 {
			key := strings.TrimSpace(tagParts[0])
			value := strings.TrimSpace(tagParts[1])
			conditions[key] = value
		}
	}
	return d.MatchTags(conditions)
}

// UpdateStatus 更新设备状态
func (d *RemoteDevice) UpdateStatus(status string) {
	d.Status = status
	d.LastSeen = time.Now()
}

// UpdateConfig 更新设备配置
func (d *RemoteDevice) UpdateConfig(key string, value interface{}) {
	d.Config[key] = value
}

// GetConfig 获取设备配置
func (d *RemoteDevice) GetConfig(key string) (interface{}, bool) {
	value, ok := d.Config[key]
	return value, ok
}

// IsOnline 检查设备是否在线
func (d *RemoteDevice) IsOnline() bool {
	return d.Status == "online"
}

// IsOffline 检查设备是否离线
func (d *RemoteDevice) IsOffline() bool {
	return d.Status == "offline"
}

// IsError 检查设备是否有错误
func (d *RemoteDevice) IsError() bool {
	return d.Status == "error"
}

// Age 获取设备运行时间
func (d *RemoteDevice) Age() time.Duration {
	return time.Since(d.CreatedAt)
}

// LastSeenAgo 获取最后一次在线时间
func (d *RemoteDevice) LastSeenAgo() time.Duration {
	return time.Since(d.LastSeen)
}

// String 设备字符串表示
func (d *RemoteDevice) String() string {
	return fmt.Sprintf("%s (%s) - %s", d.Name, d.IP, d.Status)
}