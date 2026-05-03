package template

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sfsEdgeStore/config"

	"gopkg.in/yaml.v2"
)

// Template 模板结构
type Template struct {
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	Devices      []map[string]interface{} `json:"devices"`
	Alerts       []map[string]interface{} `json:"alerts"`
	Baseline     map[string]interface{}   `json:"baseline"`
	SafetyLimits map[string]interface{}   `json:"safety_limits"`
}

// Manager 模板管理器
type Manager struct {
	templates map[string]Template
}

// NewManager 创建模板管理器
func NewManager() *Manager {
	return &Manager{
		templates: make(map[string]Template),
	}
}

// LoadTemplates 加载所有模板
func (m *Manager) LoadTemplates() error {
	templateDir := "./core/template/templates"

	// 遍历所有行业模板目录
	industryDirs, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return err
	}

	for _, industryDir := range industryDirs {
		if !industryDir.IsDir() {
			continue
		}

		industry := industryDir.Name()
		configPath := filepath.Join(templateDir, industry, "config.yaml")

		// 检查配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Printf("Template config not found for industry: %s", industry)
			continue
		}

		// 读取配置文件
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			log.Printf("Failed to read template config: %v", err)
			continue
		}

		// 解析配置
		var template Template
		if err := yaml.Unmarshal(data, &template); err != nil {
			log.Printf("Failed to parse template config: %v", err)
			continue
		}

		// 保存模板
		m.templates[industry] = template
		log.Printf("Loaded template for industry: %s", industry)
	}

	return nil
}

// GetTemplate 获取指定行业的模板
func (m *Manager) GetTemplate(industry string) (Template, bool) {
	template, exists := m.templates[industry]
	return template, exists
}

// ApplyTemplate 应用模板到配置
func (m *Manager) ApplyTemplate(industry string, cfg *config.Config) error {
	template, exists := m.GetTemplate(industry)
	if !exists {
		return nil
	}

	// 应用设备配置
	if template.Devices != nil {
		cfg.Devices = template.Devices
	}

	// 应用告警配置
	if template.Alerts != nil {
		cfg.Alerts = template.Alerts
	}

	// 应用基线配置
	if template.Baseline != nil {
		cfg.Baseline = template.Baseline
	}

	// 应用安全红线
	if template.SafetyLimits != nil {
		cfg.SafetyLimits = template.SafetyLimits
	}

	return nil
}

// ListIndustries 列出所有可用的行业模板
func (m *Manager) ListIndustries() []string {
	industries := make([]string, 0, len(m.templates))
	for industry := range m.templates {
		industries = append(industries, industry)
	}
	return industries
}
