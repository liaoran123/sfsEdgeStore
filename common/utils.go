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

	// 尝试解析为整数
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// 尝试解析为无符号整数
	if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
		return uintVal
	}

	// 尝试解析为浮点数
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
