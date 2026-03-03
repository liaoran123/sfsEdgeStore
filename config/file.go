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
