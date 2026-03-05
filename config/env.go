package config

import (
	"log"
	"os"
	"reflect"
)

// loadFromEnv 从环境变量加载配置
func loadFromEnv(cfg *Config) {
	// 使用反射自动从环境变量加载配置
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag != "" {
			if value := os.Getenv(envTag); value != "" {
				// 根据字段类型进行不同的处理
				switch v.Field(i).Kind() {
				case reflect.Bool:
					// 处理布尔类型
					switch value {
					case "true", "1":
						v.Field(i).SetBool(true)
					case "false", "0":
						v.Field(i).SetBool(false)
					}
				case reflect.String:
					// 处理字符串类型
					v.Field(i).SetString(value)
				}
				log.Printf("Loaded %s from environment variable: %s", field.Name, value)
			}
		}
	}
}
