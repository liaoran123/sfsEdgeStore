package common

import (
	"strings"
	"testing"
)

func TestFormatDeviceName(t *testing.T) {
	// 测试用例1: 长度超过64的设备名称
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
	// 计算需要的空格数
	expectedShortName := shortName + strings.Repeat(" ", 64-len(shortName))
	if formattedShortName != expectedShortName {
		t.Errorf("Expected padded name, got %s", formattedShortName)
	}

	// 测试用例3: 长度正好64的设备名称
	exactName := "1234567890123456789012345678901234567890123456789012345678901234" // 正好64个字符
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
