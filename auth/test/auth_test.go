package test

import (
	"testing"
	"time"

	"sfsdb-edgex-adapter-enterprise/auth"
)

// TestAPIKeyGeneration 测试API Key生成
func TestAPIKeyGeneration(t *testing.T) {
	// 生成API Key
	apiKey, err := auth.GenerateAPIKey("test-user", "admin", 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// 验证API Key不为空
	if apiKey.Key == "" {
		t.Error("Generated API key is empty")
	}

	// 验证API Key有效
	if !apiKey.IsValid() {
		t.Error("Generated API key should be valid")
	}

	// 验证角色设置正确
	if apiKey.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", apiKey.Role)
	}

	// 验证用户ID设置正确
	if apiKey.UserID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got '%s'", apiKey.UserID)
	}
}

// TestRBACPermissions 测试RBAC权限检查
func TestRBACPermissions(t *testing.T) {
	testCases := []struct {
		role       string
		permission string
		expected   bool
	}{
		// Admin should have all permissions
		{"admin", "read", true},
		{"admin", "write", true},
		{"admin", "admin", true},
		{"admin", "backup", true},
		{"admin", "restore", true},

		// User should have read, write, and backup permissions
		{"user", "read", true},
		{"user", "write", true},
		{"user", "admin", false},
		{"user", "backup", true},
		{"user", "restore", false},

		// Readonly should only have read permission
		{"readonly", "read", true},
		{"readonly", "write", false},
		{"readonly", "admin", false},
		{"readonly", "backup", false},
		{"readonly", "restore", false},

		// Invalid role should have no permissions
		{"invalid", "read", false},
	}

	for _, tc := range testCases {
		result := auth.HasPermission(auth.Role(tc.role), auth.Permission(tc.permission))
		if result != tc.expected {
			t.Errorf("Role '%s' should have permission '%s': expected %v, got %v", tc.role, tc.permission, tc.expected, result)
		}
	}
}

// TestRoleValidation 测试角色验证
func TestRoleValidation(t *testing.T) {
	validRoles := []string{"admin", "user", "readonly"}
	invalidRoles := []string{"invalid", "superuser", "guest"}

	for _, role := range validRoles {
		if !auth.ValidateRole(role) {
			t.Errorf("Role '%s' should be valid", role)
		}
	}

	for _, role := range invalidRoles {
		if auth.ValidateRole(role) {
			t.Errorf("Role '%s' should be invalid", role)
		}
	}
}

// TestAPIKeyExpiration 测试API Key过期
func TestAPIKeyExpiration(t *testing.T) {
	// 生成一个1秒后过期的API Key
	apiKey, err := auth.GenerateAPIKey("test-user", "admin", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// 验证API Key初始状态为有效
	if !apiKey.IsValid() {
		t.Error("API key should be valid initially")
	}

	// 等待2秒，让API Key过期
	time.Sleep(2 * time.Second)

	// 验证API Key现在已过期
	if apiKey.IsValid() {
		t.Error("API key should be expired after 2 seconds")
	}
}
