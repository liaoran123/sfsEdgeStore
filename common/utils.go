package common

import (
	"encoding/base64"
	"strconv"
	"strings"

	"sfsEdgeStore/config"
)

// ParseValue 根据 value 的内容自动判断类型并进行相应的转换
func ParseValue(value string) any {
	// 尝试解析为布尔值
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// 尝试解析为浮点数（统一存储为 float64 类型，避免类型不匹配）
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// 尝试解析为 base64 编码的二进制数据
	if after, ok := strings.CutPrefix(value, "base64:"); ok {
		base64Data := after
		if binaryData, err := base64.StdEncoding.DecodeString(base64Data); err == nil {
			return binaryData
		}
	}

	// 默认为字符串
	return value
}

// FormatDeviceName 格式化设备名称，确保长度为配置的最大长度
// 如果长度超过最大长度，则截断；如果不足最大长度，则用空格补全
func FormatDeviceName(deviceName string) string {
	maxLen := 64
	if cfg := config.GetConfigManager().GetConfig(); cfg != nil && cfg.DeviceNameMaxLength > 0 {
		maxLen = cfg.DeviceNameMaxLength
	}

	if len(deviceName) > maxLen {
		return deviceName[:maxLen]
	}

	if len(deviceName) < maxLen {
		return deviceName + strings.Repeat(" ", maxLen-len(deviceName))
	}

	return deviceName
}
