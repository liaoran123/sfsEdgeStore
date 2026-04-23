package device

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// DeviceProfile 设备配置文件
type DeviceProfile struct {
	Name         string            `yaml:"name"`
	Manufacturer string            `yaml:"manufacturer"`
	Model        string            `yaml:"model"`
	Resources    []DeviceResource  `yaml:"deviceResources"`
	ResourceMap  map[string]DeviceResource `yaml:"-"` // 用于快速查找
}

// DeviceResource 设备资源
type DeviceResource struct {
	Name        string                `yaml:"name"`
	Properties  ResourceProperties    `yaml:"properties"`
	Attributes  map[string]string     `yaml:"attributes"`
}

// ResourceProperties 资源属性
type ResourceProperties struct {
	ValueType     string  `yaml:"valueType"`
	Minimum       float64 `yaml:"minimum"`
	Maximum       float64 `yaml:"maximum"`
	DefaultValue  float64 `yaml:"defaultValue"`
}

// ProfileManager 设备配置管理器
type ProfileManager struct {
	profiles map[string]DeviceProfile
}

var profileManager *ProfileManager

// GetProfileManager 获取设备配置管理器单例
func GetProfileManager() *ProfileManager {
	if profileManager == nil {
		profileManager = &ProfileManager{
			profiles: make(map[string]DeviceProfile),
		}
	}
	return profileManager
}

// LoadProfiles 加载设备配置文件
func (pm *ProfileManager) LoadProfiles(dir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Device profile directory %s does not exist, using default profiles", dir)
		return nil
	}

	// 读取目录中的所有YAML文件
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read device profile directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".yml" || filepath.Ext(file.Name()) == ".yaml" {
			filePath := filepath.Join(dir, file.Name())
			if err := pm.loadProfile(filePath); err != nil {
				log.Printf("Failed to load profile %s: %v", file.Name(), err)
			}
		}
	}

	log.Printf("Loaded %d device profiles", len(pm.profiles))
	return nil
}

// loadProfile 加载单个设备配置文件
func (pm *ProfileManager) loadProfile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read profile file: %v", err)
	}

	var profile DeviceProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("failed to unmarshal profile: %v", err)
	}

	// 构建资源映射
	profile.ResourceMap = make(map[string]DeviceResource)
	for _, resource := range profile.Resources {
		profile.ResourceMap[resource.Name] = resource
	}

	pm.profiles[profile.Name] = profile
	log.Printf("Loaded device profile: %s", profile.Name)
	return nil
}

// GetProfile 获取设备配置
func (pm *ProfileManager) GetProfile(name string) (DeviceProfile, bool) {
	profile, exists := pm.profiles[name]
	return profile, exists
}

// GetAllProfiles 获取所有设备配置
func (pm *ProfileManager) GetAllProfiles() map[string]DeviceProfile {
	return pm.profiles
}

// GetResourceThreshold 获取资源阈值
func (pm *ProfileManager) GetResourceThreshold(profileName, resourceName string) (float64, bool) {
	profile, exists := pm.GetProfile(profileName)
	if !exists {
		return 0, false
	}

	resource, exists := profile.ResourceMap[resourceName]
	if !exists {
		return 0, false
	}

	return resource.Properties.Maximum, true
}

// FindProfileByDeviceName 根据设备名称查找对应的设备配置
func (pm *ProfileManager) FindProfileByDeviceName(deviceName string) (string, bool) {
	deviceNameLower := strings.ToLower(deviceName)
	deviceNameNormalized := strings.ReplaceAll(strings.ReplaceAll(deviceNameLower, "-", ""), "_", "")

	for name := range pm.profiles {
		profileNameLower := strings.ToLower(name)
		profileNameNormalized := strings.ReplaceAll(strings.ReplaceAll(profileNameLower, "-", ""), "_", "")

		if strings.Contains(deviceNameNormalized, profileNameNormalized) ||
			strings.Contains(profileNameNormalized, deviceNameNormalized) {
			return name, true
		}
	}

	return "", false
}
