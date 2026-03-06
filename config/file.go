package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// loadFromFile 从配置文件加载配置
func loadFromFile(cfg *Config) error {
	configFile := "config.json"
	if _, err := os.Stat(configFile); err != nil {
		return fmt.Errorf("config file not found: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	log.Println("Loaded config from file")
	return nil
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
