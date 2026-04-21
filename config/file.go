package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// SimpleConfig 简化配置，只包含最核心的项
type SimpleConfig struct {
	MQTTBroker string `json:"mqtt_broker"`
	MQTTTopic  string `json:"mqtt_topic"`
	HTTPPort   string `json:"http_port"`
	LicenseType string `json:"license_type"`
}

// loadFromFile 从配置文件加载配置
// 优先尝试简单配置，再尝试完整配置
func loadFromFile(cfg *Config) error {
	// 1. 先尝试简单配置文件
	simpleFile := "config.json"
	if _, err := os.Stat(simpleFile); err == nil {
		data, err := os.ReadFile(simpleFile)
		if err != nil {
			return fmt.Errorf("failed to read simple config file: %v", err)
		}
		
		// 尝试解析为简单配置
		var simpleCfg SimpleConfig
		if err := json.Unmarshal(data, &simpleCfg); err == nil {
			// 简单配置成功，合并到完整配置
			mergeSimpleConfig(cfg, &simpleCfg)
			log.Println("Loaded simple config from file")
			return nil
		}
		
		// 否则尝试解析为完整配置
		if err := json.Unmarshal(data, cfg); err == nil {
			log.Println("Loaded full config from file")
			return nil
		}
		
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	return fmt.Errorf("config file not found")
}

// mergeSimpleConfig 将简单配置合并到完整配置
func mergeSimpleConfig(fullCfg *Config, simpleCfg *SimpleConfig) {
	if simpleCfg.MQTTBroker != "" {
		fullCfg.MQTTBroker = simpleCfg.MQTTBroker
	}
	if simpleCfg.MQTTTopic != "" {
		fullCfg.MQTTTopic = simpleCfg.MQTTTopic
	}
	if simpleCfg.HTTPPort != "" {
		fullCfg.HTTPPort = simpleCfg.HTTPPort
	}
	if simpleCfg.LicenseType != "" {
		fullCfg.LicenseType = simpleCfg.LicenseType
	}
}

// SaveToFile 保存配置到文件
func SaveToFile(cfg *Config) error {
	configFile := "config.json"

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	log.Println("Config saved to file")
	return nil
}

// ReloadFromFile 从文件重新加载配置
func ReloadFromFile() (*Config, error) {
	cfg := &Config{}
	if err := loadFromFile(cfg); err != nil {
		return nil, err
	}

	// 从环境变量重新加载（优先级最高）
	loadFromEnv(cfg)

	return cfg, nil
}
