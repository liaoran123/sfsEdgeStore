package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/edgexfoundry/go-mod-configuration/v2/configuration"
	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"
)

// loadFromConfigCenter 从EdgeX配置中心加载配置
func loadFromConfigCenter(cfg *Config) error {
	// 尝试从环境变量获取配置中心信息
	configCenterAddress := os.Getenv("EDGEX_CONFIG_CENTER_ADDRESS")
	if configCenterAddress == "" {
		// 配置中心地址未设置，回退到本地配置
		return fmt.Errorf("EdgeX config center address not set")
	}

	// 从环境变量获取配置中心端口，默认为8500
	configCenterPort := 8500
	if portStr := os.Getenv("EDGEX_CONFIG_CENTER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			configCenterPort = port
		}
	}

	// 从环境变量获取配置中心类型，默认为consul
	configCenterType := os.Getenv("EDGEX_CONFIG_CENTER_TYPE")
	if configCenterType == "" {
		configCenterType = "consul"
	}

	// 从环境变量获取应用服务键，默认为sfsdb-edgex-adapter
	appServiceKey := os.Getenv("EDGEX_APP_SERVICE_KEY")
	if appServiceKey == "" {
		appServiceKey = "sfsdb-edgex-adapter"
	}

	// 使用go-mod-configuration库的配置服务
	configServiceConfig := types.ServiceConfig{
		Host:     configCenterAddress,
		Port:     configCenterPort,
		Type:     configCenterType,
		Protocol: "http",
		BasePath: appServiceKey,
	}

	// 创建配置客户端
	client, err := configuration.NewConfigurationClient(configServiceConfig)
	if err != nil {
		return fmt.Errorf("failed to create configuration client: %v", err)
	}

	// 尝试从配置中心获取配置
	configData, err := client.GetConfiguration(cfg)
	if err != nil {
		return fmt.Errorf("failed to get configuration from config center: %v", err)
	}

	// 解析配置数据
	if configMap, ok := configData.(*Config); ok {
		*cfg = *configMap
	}

	fmt.Println("Config loaded from EdgeX config center using go-mod-configuration")
	return nil
}