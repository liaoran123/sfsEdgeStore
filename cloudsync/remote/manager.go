package remote

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"sfsEdgeStore/config"
	"sfsEdgeStore/monitor"
)

// RemoteManager 远程管理管理器
type RemoteManager struct {
	config       *config.Config
	monitor      *monitor.Monitor
	devices      map[string]*RemoteDevice
	scripts      map[string]*Script
	rules        map[string]*Rule
	scriptExecutor *ScriptExecutor
	ruleEngine   *RuleEngine
	mutex        sync.RWMutex
}

// NewRemoteManager 创建远程管理管理器
func NewRemoteManager(cfg *config.Config, monitor *monitor.Monitor) *RemoteManager {
	manager := &RemoteManager{
		config:       cfg,
		monitor:      monitor,
		devices:      make(map[string]*RemoteDevice),
		scripts:      make(map[string]*Script),
		rules:        make(map[string]*Rule),
		scriptExecutor: NewScriptExecutor(monitor),
		ruleEngine:   NewRuleEngine(monitor),
	}

	// 启动规则引擎
	go manager.ruleEngine.Start()

	return manager
}

// AddDevice 添加远程设备
func (rm *RemoteManager) AddDevice(device *RemoteDevice) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.devices[device.ID] = device
	log.Printf("Added remote device: %s", device.Name)
}

// GetDevice 获取远程设备
func (rm *RemoteManager) GetDevice(id string) *RemoteDevice {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.devices[id]
}

// ListDevices 列出所有设备
func (rm *RemoteManager) ListDevices() []*RemoteDevice {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	devices := make([]*RemoteDevice, 0, len(rm.devices))
	for _, device := range rm.devices {
		devices = append(devices, device)
	}
	return devices
}

// AddScript 添加脚本
func (rm *RemoteManager) AddScript(script *Script) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.scripts[script.ID] = script
	log.Printf("Added script: %s", script.Name)
}

// GetScript 获取脚本
func (rm *RemoteManager) GetScript(id string) *Script {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.scripts[id]
}

// ListScripts 列出所有脚本
func (rm *RemoteManager) ListScripts() []*Script {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	scripts := make([]*Script, 0, len(rm.scripts))
	for _, script := range rm.scripts {
		scripts = append(scripts, script)
	}
	return scripts
}

// AddRule 添加规则
func (rm *RemoteManager) AddRule(rule *Rule) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.rules[rule.ID] = rule
	rm.ruleEngine.AddRule(rule)
	log.Printf("Added rule: %s", rule.Name)
}

// GetRule 获取规则
func (rm *RemoteManager) GetRule(id string) *Rule {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.rules[id]
}

// ListRules 列出所有规则
func (rm *RemoteManager) ListRules() []*Rule {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	rules := make([]*Rule, 0, len(rm.rules))
	for _, rule := range rm.rules {
		rules = append(rules, rule)
	}
	return rules
}

// ExecuteScript 执行脚本
func (rm *RemoteManager) ExecuteScript(scriptID string, deviceIDs []string) ([]*ScriptExecutionResult, error) {
	script := rm.GetScript(scriptID)
	if script == nil {
		return nil, fmt.Errorf("script not found: %s", scriptID)
	}

	results := make([]*ScriptExecutionResult, 0, len(deviceIDs))
	for _, deviceID := range deviceIDs {
		device := rm.GetDevice(deviceID)
		if device == nil {
			results = append(results, &ScriptExecutionResult{
				DeviceID:   deviceID,
				Status:     "error",
				Output:     "device not found",
				ExecutedAt: time.Now(),
			})
			continue
		}

		result, err := rm.scriptExecutor.Execute(script, device)
		if err != nil {
			results = append(results, &ScriptExecutionResult{
				DeviceID:   deviceID,
				Status:     "error",
				Output:     err.Error(),
				ExecutedAt: time.Now(),
			})
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// UpdateDeviceConfig 更新设备配置
func (rm *RemoteManager) UpdateDeviceConfig(deviceIDs []string, configUpdates map[string]interface{}) error {
	for _, deviceID := range deviceIDs {
		device := rm.GetDevice(deviceID)
		if device == nil {
			continue
		}

		// 这里可以实现配置更新逻辑
		// 例如通过MQTT发送配置更新消息
		device.Config = configUpdates
		log.Printf("Updated config for device: %s", device.Name)
	}

	return nil
}

// ProcessCommand 处理远程命令
func (rm *RemoteManager) ProcessCommand(command []byte) (interface{}, error) {
	var cmd struct {
		Type      string      `json:"type"`
		DeviceIDs []string    `json:"device_ids"`
		ScriptID  string      `json:"script_id"`
		Config    interface{} `json:"config"`
	}

	if err := json.Unmarshal(command, &cmd); err != nil {
		return nil, err
	}

	switch cmd.Type {
	case "execute_script":
		results, err := rm.ExecuteScript(cmd.ScriptID, cmd.DeviceIDs)
		return results, err
	case "update_config":
		err := rm.UpdateDeviceConfig(cmd.DeviceIDs, cmd.Config.(map[string]interface{}))
		return map[string]string{"status": "success"}, err
	default:
		return nil, fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

// Close 关闭远程管理器
func (rm *RemoteManager) Close() {
	rm.ruleEngine.Stop()
}