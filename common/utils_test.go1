package common

import (
	"strings"
	"testing"

	"sfsEdgeStore/config"
)

func TestFormatDeviceName(t *testing.T) {
	// 测试用例1: 默认长度64
	longName := "ThisIsAVeryLongDeviceNameThatExceedsSixtyFourCharactersLimitByQuiteALot"
	formattedLongName := FormatDeviceName(longName)
	if len(formattedLongName) != 64 {
		t.Errorf("Expected length 64, got %d", len(formattedLongName))
	}
	if formattedLongName != longName[:64] {
		t.Errorf("Expected truncated name, got %s", formattedLongName)
	}

	// 测试用例2: 长度不足64的设备名称
	shortName := "ShortDevice"
	formattedShortName := FormatDeviceName(shortName)
	if len(formattedShortName) != 64 {
		t.Errorf("Expected length 64, got %d", len(formattedShortName))
	}
	expectedShortName := shortName + strings.Repeat(" ", 64-len(shortName))
	if formattedShortName != expectedShortName {
		t.Errorf("Expected padded name, got %s", formattedShortName)
	}

	// 测试用例3: 长度正好64的设备名称
	exactName := "1234567890123456789012345678901234567890123456789012345678901234"
	formattedExactName := FormatDeviceName(exactName)
	if len(formattedExactName) != 64 {
		t.Errorf("Expected length 64, got %d", len(formattedExactName))
	}
	if formattedExactName != exactName {
		t.Errorf("Expected unchanged name, got %s", formattedExactName)
	}

	// 测试用例4: 空字符串
	emptyName := ""
	formattedEmptyName := FormatDeviceName(emptyName)
	if len(formattedEmptyName) != 64 {
		t.Errorf("Expected length 64, got %d", len(formattedEmptyName))
	}
}

func TestFormatDeviceNameWithConfig(t *testing.T) {
	// 设置配置为32
	cfg := &config.Config{
		DeviceNameMaxLength: 32,
	}
	config.GetConfigManager().SetConfig(cfg)

	// 测试用例1: 自定义长度32
	longName := "ThisIsAVeryLongDeviceNameThatExceedsThirtyTwoCharactersLimit"
	formattedLongName := FormatDeviceName(longName)
	if len(formattedLongName) != 32 {
		t.Errorf("Expected length 32, got %d", len(formattedLongName))
	}
	if formattedLongName != longName[:32] {
		t.Errorf("Expected truncated name, got %s", formattedLongName)
	}

	// 测试用例2: 长度不足32的设备名称
	shortName := "ShortDevice"
	formattedShortName := FormatDeviceName(shortName)
	if len(formattedShortName) != 32 {
		t.Errorf("Expected length 32, got %d", len(formattedShortName))
	}
	expectedShortName := shortName + strings.Repeat(" ", 32-len(shortName))
	if formattedShortName != expectedShortName {
		t.Errorf("Expected padded name, got %s", formattedShortName)
	}

	// 恢复默认配置
	config.GetConfigManager().SetConfig(nil)
}

func TestFormatDeviceNameWithCustomLength128(t *testing.T) {
	// 设置配置为128
	cfg := &config.Config{
		DeviceNameMaxLength: 128,
	}
	config.GetConfigManager().SetConfig(cfg)

	// 测试自定义长度128
	mediumName := "MediumDeviceName"
	formattedMediumName := FormatDeviceName(mediumName)
	if len(formattedMediumName) != 128 {
		t.Errorf("Expected length 128, got %d", len(formattedMediumName))
	}

	// 恢复默认配置
	config.GetConfigManager().SetConfig(nil)
}
