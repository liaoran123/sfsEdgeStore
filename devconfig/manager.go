package devconfig

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Device struct {
	Name              string
	IP                string
	Protocol          string
	Template          string
	Interval          string
	SubscriptionTopic string
	// 过滤相关字段
	OnChange       bool
	ValueThreshold float64
	ValueOperator  string
	MinInterval    string
}

type DeviceV2 struct {
	Name     string   `yaml:"name"`
	Protocol string   `yaml:"protocol"`
	Address  string   `yaml:"address"`
	UnitId   int      `yaml:"unitId,omitempty"`
	Topic    string   `yaml:"topic,omitempty"`
	Template string   `yaml:"template"`
	Interval string   `yaml:"interval,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
	// 订阅相关字段
	SubscriptionTopic string `yaml:"subscriptionTopic,omitempty"` // EdgeX订阅主题
	// 过滤相关字段
	OnChange       bool    `yaml:"onChange,omitempty"`       // 仅当数值变化时存储
	ValueThreshold float64 `yaml:"valueThreshold,omitempty"` // 数值阈值
	ValueOperator  string  `yaml:"valueOperator,omitempty"`  // 比较操作符: >, <, >=, <=, ==, !=
	MinInterval    string  `yaml:"minInterval,omitempty"`    // 最小存储间隔
}

type DevicesConfig struct {
	Devices []DeviceV2 `yaml:"devices"`
}

type Template struct {
	Name            string
	Protocol        string
	DeviceType      string
	DeviceResources []DeviceResource
}

type DeviceResource struct {
	Name         string
	Description  string
	Address      string
	RegisterType string
	DataType     string
	ReadWrite    string
}

type ConfigManager struct {
	configDir   string
	templateDir string
	devicesFile string
	yamlFile    string
}

func NewConfigManager(configDir string) (*ConfigManager, error) {
	cm := &ConfigManager{
		configDir:   configDir,
		templateDir: filepath.Join(configDir, "templates"),
		devicesFile: filepath.Join(configDir, "devices.csv"),
		yamlFile:    filepath.Join(configDir, "devices.yaml"),
	}

	if err := os.MkdirAll(cm.templateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	return cm, nil
}

func (cm *ConfigManager) GetTemplateDir() string {
	return cm.templateDir
}

func (cm *ConfigManager) GetDevicesFile() string {
	return cm.devicesFile
}

func (cm *ConfigManager) GetYamlFile() string {
	return cm.yamlFile
}

func (cm *ConfigManager) ListTemplates() ([]string, error) {
	var templates []string

	entries, err := os.ReadDir(cm.templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}

	return templates, nil
}

func (cm *ConfigManager) ListTemplateFiles(protocol string) ([]string, error) {
	protocolDir := filepath.Join(cm.templateDir, protocol)

	if _, err := os.Stat(protocolDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("protocol directory not found: %s", protocol)
	}

	var files []string

	entries, err := os.ReadDir(protocolDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (cm *ConfigManager) LoadDevices() ([]Device, error) {
	if _, err := os.Stat(cm.yamlFile); err == nil {
		return cm.loadDevicesFromYaml()
	}

	return cm.loadDevicesFromCsv()
}

func (cm *ConfigManager) loadDevicesFromYaml() ([]Device, error) {
	data, err := os.ReadFile(cm.yamlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	var config DevicesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	var devices []Device
	for _, d := range config.Devices {
		interval := d.Interval
		if interval == "" {
			interval = "15s"
		}

		device := Device{
			Name:              d.Name,
			IP:                d.Address,
			Protocol:          d.Protocol,
			Template:          d.Template,
			Interval:          interval,
			SubscriptionTopic: d.SubscriptionTopic,
			// 过滤相关字段
			OnChange:       d.OnChange,
			ValueThreshold: d.ValueThreshold,
			ValueOperator:  d.ValueOperator,
			MinInterval:    d.MinInterval,
		}

		if device.Name == "" || device.IP == "" {
			continue
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func (cm *ConfigManager) loadDevicesFromCsv() ([]Device, error) {
	file, err := os.Open(cm.devicesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Device{}, nil
		}
		return nil, fmt.Errorf("failed to open devices file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	if len(records) < 2 {
		return []Device{}, nil
	}

	var devices []Device
	for i, record := range records[1:] {
		if len(record) < 5 {
			return nil, fmt.Errorf("invalid CSV format at line %d: expected 5 fields, got %d", i+2, len(record))
		}

		device := Device{
			Name:     strings.TrimSpace(record[0]),
			IP:       strings.TrimSpace(record[1]),
			Protocol: strings.TrimSpace(record[2]),
			Template: strings.TrimSpace(record[3]),
			Interval: strings.TrimSpace(record[4]),
		}

		if device.Name == "" || device.IP == "" {
			return nil, fmt.Errorf("invalid device at line %d: name and IP are required", i+2)
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func (cm *ConfigManager) SaveDevices(devices []Device) error {
	if _, err := os.Stat(cm.yamlFile); err == nil {
		return cm.saveDevicesToYaml(devices)
	}
	return cm.saveDevicesToCsv(devices)
}

func (cm *ConfigManager) saveDevicesToYaml(devices []Device) error {
	var config DevicesConfig
	for _, d := range devices {
		protocol := d.Protocol
		if protocol == "modbus-tcp" {
			protocol = "modbus-tcp"
		}

		deviceV2 := DeviceV2{
			Name:              d.Name,
			Protocol:          d.Protocol,
			Address:           d.IP,
			Template:          d.Template,
			Interval:          d.Interval,
			SubscriptionTopic: d.SubscriptionTopic,
			// 过滤相关字段
			OnChange:       d.OnChange,
			ValueThreshold: d.ValueThreshold,
			ValueOperator:  d.ValueOperator,
			MinInterval:    d.MinInterval,
		}

		config.Devices = append(config.Devices, deviceV2)
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	if err := os.WriteFile(cm.yamlFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %v", err)
	}

	return nil
}

func (cm *ConfigManager) saveDevicesToCsv(devices []Device) error {
	file, err := os.Create(cm.devicesFile)
	if err != nil {
		return fmt.Errorf("failed to create devices file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"name", "ip", "protocol", "template", "interval"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	for _, device := range devices {
		record := []string{
			device.Name,
			device.IP,
			device.Protocol,
			device.Template,
			device.Interval,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write device: %v", err)
		}
	}

	return nil
}

func (cm *ConfigManager) AddDevice(device Device) error {
	devices, err := cm.LoadDevices()
	if err != nil {
		return err
	}

	for _, d := range devices {
		if d.Name == device.Name {
			return fmt.Errorf("device already exists: %s", device.Name)
		}
	}

	devices = append(devices, device)
	return cm.SaveDevices(devices)
}

func (cm *ConfigManager) RemoveDevice(name string) error {
	devices, err := cm.LoadDevices()
	if err != nil {
		return err
	}

	found := false
	var updated []Device
	for _, d := range devices {
		if d.Name == name {
			found = true
		} else {
			updated = append(updated, d)
		}
	}

	if !found {
		return fmt.Errorf("device not found: %s", name)
	}

	return cm.SaveDevices(updated)
}

func (cm *ConfigManager) ValidateConfig() ([]string, error) {
	var errors []string

	if _, err := os.Stat(cm.configDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("config directory not found: %s", cm.configDir))
		return errors, nil
	}

	if _, err := os.Stat(cm.templateDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("template directory not found: %s", cm.templateDir))
	} else {
		templates, err := cm.ListTemplates()
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to list templates: %v", err))
		} else if len(templates) == 0 {
			errors = append(errors, "no templates found")
		}
	}

	devices, err := cm.LoadDevices()
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to load devices: %v", err))
	} else {
		for _, device := range devices {
			templatePath := cm.GetTemplatePath(device.Template)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("template not found for device %s: %s", device.Name, device.Template))
			}
		}
	}

	return errors, nil
}

func (cm *ConfigManager) GetTemplatePath(templateName string) string {
	path := filepath.Join(cm.templateDir, templateName)
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		path = path + ".yaml"
	}
	return path
}

func (cm *ConfigManager) TemplateExists(templateName string) bool {
	path := cm.GetTemplatePath(templateName)
	_, err := os.Stat(path)
	return err == nil
}

func (cm *ConfigManager) GetConfigMode() string {
	if _, err := os.Stat(cm.yamlFile); err == nil {
		return "yaml"
	}
	if _, err := os.Stat(cm.devicesFile); err == nil {
		return "csv"
	}
	return "none"
}
