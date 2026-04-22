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
	lastValues        map[string]interface{}
	lastTimes         map[string]time.Time
	allowedDevices    []string // 允许的设备列表
	allowedResources  []string // 允许的资源名称列表
	excludedDevices   []string // 排除的设备列表
	excludedResources []string // 排除的资源名称列表
}

func NewFilterManager() *FilterManager {
	return &FilterManager{
		deviceConfig:      make(map[string]*devconfig.Device),
		lastValues:        make(map[string]interface{}),
		lastTimes:         make(map[string]time.Time),
		allowedDevices:    []string{},
		allowedResources:  []string{},
		excludedDevices:   []string{},
		excludedResources: []string{},
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

func (fm *FilterManager) updateLastValue(key string, value interface{}) {
	fm.lastValues[key] = value
}

func (fm *FilterManager) updateLastTime(key string) {
	fm.lastTimes[key] = time.Now()
}

func (fm *FilterManager) SetAllowedDevices(devices []string) {
	fm.allowedDevices = devices
	log.Printf("Set allowed devices: %v", devices)
}

func (fm *FilterManager) SetAllowedResources(resources []string) {
	fm.allowedResources = resources
	log.Printf("Set allowed resources: %v", resources)
}

func (fm *FilterManager) SetExcludedDevices(devices []string) {
	fm.excludedDevices = devices
	log.Printf("Set excluded devices: %v", devices)
}

func (fm *FilterManager) SetExcludedResources(resources []string) {
	fm.excludedResources = resources
	log.Printf("Set excluded resources: %v", resources)
}

func (fm *FilterManager) ShouldStore(deviceName, readingName string, value interface{}) bool {
	// 1. 检查设备是否在排除列表中
	for _, excludedDevice := range fm.excludedDevices {
		if excludedDevice == deviceName {
			return false
		}
	}

	// 2. 检查资源是否在排除列表中
	for _, excludedResource := range fm.excludedResources {
		if excludedResource == readingName {
			return false
		}
	}

	// 3. 检查设备是否在允许列表中（如果允许列表不为空）
	if len(fm.allowedDevices) > 0 {
		found := false
		for _, allowedDevice := range fm.allowedDevices {
			if allowedDevice == deviceName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 4. 检查资源是否在允许列表中（如果允许列表不为空）
	if len(fm.allowedResources) > 0 {
		found := false
		for _, allowedResource := range fm.allowedResources {
			if allowedResource == readingName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 5. 查找设备配置
	device, ok := fm.deviceConfig[deviceName]
	if !ok {
		// 没有配置，默认存储
		return true
	}

	// 生成唯一键（设备名+读数名）
	key := fmt.Sprintf("%s:%s", deviceName, readingName)

	// 6. 检查最小存储间隔
	if device.MinInterval != "" {
		if !fm.checkMinInterval(key, device.MinInterval) {
			return false
		}
	}

	// 7. 检查OnChange过滤
	if device.OnChange {
		if !fm.checkOnChange(key, value) {
			return false
		}
	}

	// 8. 检查阈值过滤
	if device.ValueOperator != "" && device.ValueThreshold != 0 {
		if !fm.checkValueThreshold(value, device.ValueOperator, device.ValueThreshold) {
			return false
		}
	}

	// 所有过滤条件通过，存储数据
	fm.updateLastValue(key, value)
	fm.updateLastTime(key)
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
