package remote

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sfsEdgeStore/monitor"
)

// Script 脚本
type Script struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"` // 脚本内容
	Language    string    `json:"language"` // shell, python, etc.
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewScript 创建脚本
func NewScript(id, name, description, content, language string) *Script {
	now := time.Now()
	return &Script{
		ID:          id,
		Name:        name,
		Description: description,
		Content:     content,
		Language:    language,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ScriptExecutionResult 脚本执行结果
type ScriptExecutionResult struct {
	DeviceID   string    `json:"device_id"`
	Status     string    `json:"status"` // success, error
	Output     string    `json:"output"`
	ExecutedAt time.Time `json:"executed_at"`
}

// ScriptExecutor 脚本执行器
type ScriptExecutor struct {
	monitor *monitor.Monitor
}

// NewScriptExecutor 创建脚本执行器
func NewScriptExecutor(monitor *monitor.Monitor) *ScriptExecutor {
	return &ScriptExecutor{
		monitor: monitor,
	}
}

// Execute 执行脚本
func (se *ScriptExecutor) Execute(script *Script, device *RemoteDevice) (*ScriptExecutionResult, error) {
	result := &ScriptExecutionResult{
		DeviceID:   device.ID,
		ExecutedAt: time.Now(),
	}

	// 记录脚本执行开始
	se.monitor.RecordInfo("script", fmt.Sprintf("Executing script %s on device %s", script.Name, device.Name))

	// 根据脚本语言执行
	switch strings.ToLower(script.Language) {
	case "shell":
		output, err := se.executeShell(script.Content)
		if err != nil {
			result.Status = "error"
			result.Output = err.Error()
			se.monitor.RecordError("script", fmt.Sprintf("Script execution failed: %v", err))
			return result, err
		}
		result.Status = "success"
		result.Output = output
	case "python":
		output, err := se.executePython(script.Content)
		if err != nil {
			result.Status = "error"
			result.Output = err.Error()
			se.monitor.RecordError("script", fmt.Sprintf("Script execution failed: %v", err))
			return result, err
		}
		result.Status = "success"
		result.Output = output
	default:
		err := fmt.Errorf("unsupported script language: %s", script.Language)
		result.Status = "error"
		result.Output = err.Error()
		se.monitor.RecordError("script", err.Error())
		return result, err
	}

	// 记录脚本执行成功
	se.monitor.RecordInfo("script", fmt.Sprintf("Script %s executed successfully on device %s", script.Name, device.Name))

	return result, nil
}

// executeShell 执行Shell脚本
func (se *ScriptExecutor) executeShell(content string) (string, error) {
	cmd := exec.Command("cmd.exe", "/c", content)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// executePython 执行Python脚本
func (se *ScriptExecutor) executePython(content string) (string, error) {
	cmd := exec.Command("python", "-c", content)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}