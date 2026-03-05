package common

import (
	"encoding/base64"
	"strconv"
	"strings"
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
	if strings.HasPrefix(value, "base64:") {
		base64Data := strings.TrimPrefix(value, "base64:")
		if binaryData, err := base64.StdEncoding.DecodeString(base64Data); err == nil {
			return binaryData
		}
	}

	// 默认为字符串
	return value
}

// FormatDeviceName 格式化设备名称，确保长度为 64 字符
// 如果长度超过 64，则截断；如果不足 64，则用空格补全
func FormatDeviceName(deviceName string) string {
	const maxLength = 64

	// 截断过长的设备名称
	if len(deviceName) > maxLength {
		return deviceName[:maxLength]
	}

	// 补全不足 64 字符的设备名称
	if len(deviceName) < maxLength {
		return deviceName + strings.Repeat(" ", maxLength-len(deviceName))
	}

	return deviceName
}
