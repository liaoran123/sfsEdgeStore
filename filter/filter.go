package filter

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"sfsEdgeStore/devconfig"
)

type FilterManager struct {
	deviceConfig      map[string]*devconfig.Device
	lastValues        map[string]any
	lastTimes         map[string]time.Time
	allowedDevices    map[string]struct{} // 允许的设备集合（O(1)查找）// struct{} 是零大小类型（zero-size type）大小 = 0 字节
	allowedResources  map[string]struct{} // 允许的资源集合
	excludedDevices   map[string]struct{} // 排除的设备集合
	excludedResources map[string]struct{} // 排除的资源集合
}

func NewFilterManager() *FilterManager {
	return &FilterManager{
		deviceConfig:      make(map[string]*devconfig.Device),
		lastValues:        make(map[string]any),
		lastTimes:         make(map[string]time.Time),
		allowedDevices:    make(map[string]struct{}),
		allowedResources:  make(map[string]struct{}),
		excludedDevices:   make(map[string]struct{}),
		excludedResources: make(map[string]struct{}),
	}
}

func (fm *FilterManager) LoadDeviceConfig(devices []devconfig.Device) {
	for i := range devices {
		fm.deviceConfig[devices[i].Name] = &devices[i]
	}
	log.Printf("Loaded filter configuration for %d devices", len(devices))
}

func (fm *FilterManager) checkMinInterval(key, minInterval string) bool {
	lastTime, exists := fm.lastTimes[key]
	if !exists {
		return true
	}

	duration, err := time.ParseDuration(minInterval)
	if err != nil {
		log.Printf("Invalid minInterval format: %s, using default 1s", minInterval)
		duration = time.Second
	}

	return time.Since(lastTime) >= duration
}

func (fm *FilterManager) checkOnChange(key string, value interface{}) bool {
	lastValue, exists := fm.lastValues[key]
	if !exists {
		return true
	}

	// 比较值是否变化
	return !fm.valuesEqual(lastValue, value)
}

func (fm *FilterManager) checkValueThreshold(value interface{}, operator string, threshold float64) bool {
	// 将值转换为float64进行比较
	floatValue, err := fm.toFloat64(value)
	if err != nil {
		log.Printf("Failed to convert value to float64: %v", err)
		return true
	}

	switch operator {
	case ">":
		return floatValue > threshold
	case "<":
		return floatValue < threshold
	case ">=":
		return floatValue >= threshold
	case "<=":
		return floatValue <= threshold
	case "==":
		return floatValue == threshold
	case "!=":
		return floatValue != threshold
	default:
		log.Printf("Invalid operator: %s", operator)
		return true
	}
}

func (fm *FilterManager) valuesEqual(a, b interface{}) bool {
	// 简单的相等性比较
	sa := fmt.Sprintf("%v", a)
	sb := fmt.Sprintf("%v", b)
	return sa == sb
}

func (fm *FilterManager) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

func (fm *FilterManager) SetAllowedDevices(devices []string) {
	fm.allowedDevices = make(map[string]struct{}, len(devices))
	for _, d := range devices {
		fm.allowedDevices[d] = struct{}{}
	}
	log.Printf("Set allowed devices: %d", len(devices))
}

func (fm *FilterManager) SetAllowedResources(resources []string) {
	fm.allowedResources = make(map[string]struct{}, len(resources))
	for _, r := range resources {
		fm.allowedResources[r] = struct{}{}
	}
	log.Printf("Set allowed resources: %d", len(resources))
}

func (fm *FilterManager) SetExcludedDevices(devices []string) {
	fm.excludedDevices = make(map[string]struct{}, len(devices))
	for _, d := range devices {
		fm.excludedDevices[d] = struct{}{}
	}
	log.Printf("Set excluded devices: %d", len(devices))
}

func (fm *FilterManager) SetExcludedResources(resources []string) {
	fm.excludedResources = make(map[string]struct{}, len(resources))
	for _, r := range resources {
		fm.excludedResources[r] = struct{}{}
	}
	log.Printf("Set excluded resources: %d", len(resources))
}

func (fm *FilterManager) ShouldStore(deviceName, readingName string, value interface{}) bool {
	// 1. 检查排除列表
	if _, excluded := fm.excludedDevices[deviceName]; excluded {
		return false
	}
	if _, excluded := fm.excludedResources[readingName]; excluded {
		return false
	}

	// 2. 检查允许列表（非空时才检查）
	if len(fm.allowedDevices) > 0 {
		if _, allowed := fm.allowedDevices[deviceName]; !allowed {
			return false
		}
	}
	if len(fm.allowedResources) > 0 {
		if _, allowed := fm.allowedResources[readingName]; !allowed {
			return false
		}
	}

	// 3. 查找设备配置
	device, ok := fm.deviceConfig[deviceName]
	if !ok {
		return true // 无配置，默认存储
	}

	// 4. 检查过滤规则
	key := deviceName + ":" + readingName
	if device.MinInterval != "" && !fm.checkMinInterval(key, device.MinInterval) {
		return false
	}
	if device.OnChange && !fm.checkOnChange(key, value) {
		return false
	}
	if device.ValueOperator != "" && device.ValueThreshold != 0 && !fm.checkValueThreshold(value, device.ValueOperator, device.ValueThreshold) {
		return false
	}

	// 5. 更新最后记录的值和时间
	fm.lastValues[key] = value
	fm.lastTimes[key] = time.Now()
	return true
}

func (fm *FilterManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"devices_configured": len(fm.deviceConfig),
		"last_values_stored": len(fm.lastValues),
		"allowed_devices":    len(fm.allowedDevices),
		"allowed_resources":  len(fm.allowedResources),
		"excluded_devices":   len(fm.excludedDevices),
		"excluded_resources": len(fm.excludedResources),
	}
}
