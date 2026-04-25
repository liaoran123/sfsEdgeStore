package remote

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"sfsEdgeStore/monitor"
)

// Rule 规则
type Rule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Condition   string    `json:"condition"` // 条件表达式
	Action      string    `json:"action"`    // 执行的操作
	DeviceIDs   []string  `json:"device_ids"` // 适用的设备
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewRule 创建规则
func NewRule(id, name, description, condition, action string, deviceIDs []string) *Rule {
	now := time.Now()
	return &Rule{
		ID:          id,
		Name:        name,
		Description: description,
		Condition:   condition,
		Action:      action,
		DeviceIDs:   deviceIDs,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// RuleEngine 规则引擎
type RuleEngine struct {
	rules   map[string]*Rule
	devices map[string]*RemoteDevice
	monitor *monitor.Monitor
	running bool
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(monitor *monitor.Monitor) *RuleEngine {
	return &RuleEngine{
		rules:   make(map[string]*Rule),
		devices: make(map[string]*RemoteDevice),
		monitor: monitor,
		running: false,
	}
}

// AddRule 添加规则
func (re *RuleEngine) AddRule(rule *Rule) {
	re.rules[rule.ID] = rule
	log.Printf("Added rule: %s", rule.Name)
}

// RemoveRule 移除规则
func (re *RuleEngine) RemoveRule(id string) {
	delete(re.rules, id)
	log.Printf("Removed rule: %s", id)
}

// AddDevice 添加设备
func (re *RuleEngine) AddDevice(device *RemoteDevice) {
	re.devices[device.ID] = device
}

// Start 启动规则引擎
func (re *RuleEngine) Start() {
	if re.running {
		return
	}

	re.running = true
	go re.run()
	log.Println("Rule engine started")
}

// Stop 停止规则引擎
func (re *RuleEngine) Stop() {
	re.running = false
	log.Println("Rule engine stopped")
}

// run 运行规则引擎
func (re *RuleEngine) run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for re.running {
		select {
		case <-ticker.C:
			re.checkRules()
		}
	}
}

// checkRules 检查规则
func (re *RuleEngine) checkRules() {
	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}

		// 检查规则是否适用于任何设备
		for _, deviceID := range rule.DeviceIDs {
			device := re.devices[deviceID]
			if device == nil {
				continue
			}

			// 评估条件
			if re.evaluateCondition(rule.Condition, device) {
				re.executeAction(rule, device)
			}
		}
	}
}

// evaluateCondition 评估条件
func (re *RuleEngine) evaluateCondition(condition string, device *RemoteDevice) bool {
	// 解析条件表达式
	parts := strings.Split(condition, " ")
	if len(parts) < 3 {
		return false
	}

	metric := parts[0]
	operator := parts[1]
	value := parts[2]

	switch metric {
	case "device_status":
		return re.evaluateDeviceStatus(device, operator, value)
	case "last_seen":
		return re.evaluateLastSeen(device, operator, value)
	case "cpu_usage":
		return re.evaluateCPUUsage(operator, value)
	default:
		return false
	}
}

// evaluateDeviceStatus 评估设备状态
func (re *RuleEngine) evaluateDeviceStatus(device *RemoteDevice, operator, value string) bool {
	switch operator {
	case "=":
		return device.Status == value
	case "!=":
		return device.Status != value
	default:
		return false
	}
}

// evaluateLastSeen 评估最后在线时间
func (re *RuleEngine) evaluateLastSeen(device *RemoteDevice, operator, value string) bool {
	hours, err := time.ParseDuration(value + "h")
	if err != nil {
		return false
	}

	lastSeenAgo := time.Since(device.LastSeen)
	switch operator {
	case ">":
		return lastSeenAgo > hours
	case "<":
		return lastSeenAgo < hours
	default:
		return false
	}
}

// evaluateCPUUsage 评估CPU使用率
func (re *RuleEngine) evaluateCPUUsage(operator, value string) bool {
	// 这里可以实现CPU使用率的评估
	// 简化实现，实际应该获取真实的CPU使用率
	return false
}

// executeAction 执行操作
func (re *RuleEngine) executeAction(rule *Rule, device *RemoteDevice) {
	log.Printf("Executing action for rule %s on device %s: %s", rule.Name, device.Name, rule.Action)
	re.monitor.RecordInfo("rule", fmt.Sprintf("Rule %s triggered action: %s on device %s", rule.Name, rule.Action, device.Name))

	// 解析并执行操作
	parts := strings.Split(rule.Action, " ")
	if len(parts) == 0 {
		return
	}

	actionType := parts[0]
	switch actionType {
	case "restart":
		re.executeRestartAction(device)
	case "send_alert":
		re.executeSendAlertAction(rule, device)
	case "run_script":
		if len(parts) > 1 {
			script := strings.Join(parts[1:], " ")
			re.executeRunScriptAction(script, device)
		}
	default:
		log.Printf("Unknown action type: %s", actionType)
	}
}

// executeRestartAction 执行重启操作
func (re *RuleEngine) executeRestartAction(device *RemoteDevice) {
	log.Printf("Restarting device: %s", device.Name)
	// 这里可以实现设备重启逻辑
}

// executeSendAlertAction 执行发送告警操作
func (re *RuleEngine) executeSendAlertAction(rule *Rule, device *RemoteDevice) {
	log.Printf("Sending alert for device: %s", device.Name)
	// 这里可以实现发送告警逻辑
}

// executeRunScriptAction 执行脚本操作
func (re *RuleEngine) executeRunScriptAction(script string, device *RemoteDevice) {
	log.Printf("Running script on device %s: %s", device.Name, script)
	
	// 执行脚本
	cmd := exec.Command("cmd.exe", "/c", script)
	if err := cmd.Run(); err != nil {
		log.Printf("Script execution failed: %v", err)
		re.monitor.RecordError("rule", fmt.Sprintf("Script execution failed on device %s: %v", device.Name, err))
	} else {
		log.Printf("Script executed successfully on device: %s", device.Name)
	}
}

// UpdateRule 更新规则
func (re *RuleEngine) UpdateRule(rule *Rule) {
	re.rules[rule.ID] = rule
	rule.UpdatedAt = time.Now()
	log.Printf("Updated rule: %s", rule.Name)
}

// ListRules 列出所有规则
func (re *RuleEngine) ListRules() []*Rule {
	rules := make([]*Rule, 0, len(re.rules))
	for _, rule := range re.rules {
		rules = append(rules, rule)
	}
	return rules
}