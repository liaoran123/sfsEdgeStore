# 认证授权与 RBAC  xxxxxx  不是项目需要实现的功能，抛弃。

## 概述

sfsEdgeStore 提供了基于 API Key 的认证机制和基于角色的访问控制（RBAC）系统。

## API Key 结构

### APIKey 结构体

```go
// auth/api_key.go:10-19
type APIKey struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Hash      string    `json:"hash"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
}
```

## 生成 API Key

### GenerateAPIKey 函数

```go
// auth/api_key.go:21-45
func GenerateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error) {
	// 生成随机密钥
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}

	key := hex.EncodeToString(keyBytes)
	
	// 计算哈希值（实际应用中应该使用bcrypt等安全哈希）
	hash := key // 简化处理，实际应用中应该使用bcrypt

	now := time.Now()
	return &APIKey{
		ID:        generateID(),
		Key:       key,
		Hash:      hash,
		UserID:    userID,
		Role:      role,
		CreatedAt: now,
		ExpiresAt: now.Add(expiresIn),
		Active:    true,
	}, nil
}
```

### 检查 API Key 有效性

```go
// auth/api_key.go:47-50
func (k *APIKey) IsValid() bool {
	return k.Active && time.Now().Before(k.ExpiresAt)
}
```

## RBAC 角色与权限

### 角色定义

```go
// auth/rbac.go:7-11
const (
	RoleAdmin     Role = "admin"
	RoleUser      Role = "user"
	RoleReadOnly  Role = "readonly"
)
```

### 权限定义

```go
// auth/rbac.go:17-23
const (
	PermissionRead      Permission = "read"
	PermissionWrite     Permission = "write"
	PermissionAdmin     Permission = "admin"
	PermissionBackup    Permission = "backup"
	PermissionRestore   Permission = "restore"
)
```

### 角色权限映射

```go
// auth/rbac.go:26-42
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionRead,
		PermissionWrite,
		PermissionAdmin,
		PermissionBackup,
		PermissionRestore,
	},
	RoleUser: {
		PermissionRead,
		PermissionWrite,
		PermissionBackup,
	},
	RoleReadOnly: {
		PermissionRead,
	},
}
```

### 检查权限

```go
// auth/rbac.go:45-58
func HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}
```

### 获取角色权限

```go
// auth/rbac.go:60-66
func GetRolePermissions(role Role) []Permission {
	if permissions, ok := RolePermissions[role]; ok {
		return permissions
	}
	return []Permission{}
}
```

### 验证角色

```go
// auth/rbac.go:68-72
func ValidateRole(role string) bool {
	_, ok := RolePermissions[Role(role)]
	return ok
}
```

## AuthManager

### AuthManager 结构

```go
// auth/auth.go:11-13
type AuthManager struct {
	// 使用sfsDb存储API Key
}
```

### 创建 AuthManager

```go
// auth/auth.go:16-18
func NewAuthManager() *AuthManager {
	return &AuthManager{}
}
```

### 添加 API Key

```go
// auth/auth.go:20-37
func (am *AuthManager) AddAPIKey(apiKey *APIKey) error {
	// 转换为map存储到数据库
	record := map[string]any{
		"id":         apiKey.ID,
		"key":        apiKey.Key,
		"hash":       apiKey.Hash,
		"user_id":    apiKey.UserID,
		"role":       apiKey.Role,
		"created_at": apiKey.CreatedAt.UnixNano(),
		"expires_at": apiKey.ExpiresAt.UnixNano(),
		"active":     apiKey.Active,
	}

	// 插入到数据库
	_, err := database.AuthTable.Insert(&record)
	return err
}
```

### 根据 Key 获取 API Key

```go
// auth/auth.go:39-76
func (am *AuthManager) GetAPIKeyByKey(key string) (*APIKey, error) {
	// 从数据库查询
	fields := map[string]any{
		"key": key,
	}

	iter, err := database.AuthTable.Search(&fields)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	defer records.Release()

	if len(records) == 0 {
		return nil, errors.New("API key not found")
	}

	// 转换为APIKey结构
	record := records[0]
	createdAt := time.Unix(0, record["created_at"].(int64))
	expiresAt := time.Unix(0, record["expires_at"].(int64))

	apiKey := &APIKey{
		ID:        record["id"].(string),
		Key:       record["key"].(string),
		Hash:      record["hash"].(string),
		UserID:    record["user_id"].(string),
		Role:      record["role"].(string),
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		Active:    record["active"].(bool),
	}

	return apiKey, nil
}
```

### 撤销 API Key

```go
// auth/auth.go:78-87
func (am *AuthManager) RevokeAPIKey(key string) error {
	// 更新数据库中的记录
	updateRecord := map[string]any{
		"key":    key,
		"active": false,
	}

	return database.AuthTable.Update(&updateRecord)
}
```

### 列出所有 API Key

```go
// auth/auth.go:89-126
func (am *AuthManager) ListAPIKeys() ([]*APIKey, error) {
	// 查询所有记录
	fields := map[string]any{
		"key": nil, // 查询所有记录
	}

	iter, err := database.AuthTable.Search(&fields)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	defer records.Release()

	// 转换为APIKey数组
	apiKeys := make([]*APIKey, 0, len(records))
	for _, record := range records {
		createdAt := time.Unix(0, record["created_at"].(int64))
		expiresAt := time.Unix(0, record["expires_at"].(int64))

		apiKey := &APIKey{
			ID:        record["id"].(string),
			Key:       record["key"].(string),
			Hash:      record["hash"].(string),
			UserID:    record["user_id"].(string),
			Role:      record["role"].(string),
			CreatedAt: createdAt,
			ExpiresAt: expiresAt,
			Active:    record["active"].(bool),
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}
```

### 清理过期 Key

```go
// auth/auth.go:128-154
func (am *AuthManager) CleanExpiredKeys() (int, error) {
	// 获取所有记录
	apiKeys, err := am.ListAPIKeys()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	cleanupCount := 0

	for _, apiKey := range apiKeys {
		if now.After(apiKey.ExpiresAt) {
			// 删除过期的API Key
			deleteFields := map[string]any{
				"key": apiKey.Key,
			}
			err := database.AuthTable.Delete(&deleteFields)
			if err != nil {
				return cleanupCount, err
			}
			cleanupCount++
		}
	}

	return cleanupCount, nil
}
```

### 启动清理任务

```go
// auth/auth.go:156-174
func (am *AuthManager) StartCleanupTask(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C
			count, err := am.CleanExpiredKeys()
			if err != nil {
				// 记录错误但继续运行
				continue
			}
			if count > 0 {
				// 可以添加日志记录
			}
		}
	}()
}
```

### 创建 API Key

```go
// auth/auth.go:176-196
func (am *AuthManager) CreateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error) {
	// 验证角色
	if !ValidateRole(role) {
		return nil, errors.New("invalid role")
	}

	// 生成API Key
	apiKey, err := GenerateAPIKey(userID, role, expiresIn)
	if err != nil {
		return nil, err
	}

	// 添加到存储
	err = am.AddAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}
```

## 测试

### 认证测试

```go
// auth/test/auth_test.go
package test

import (
	"testing"
	"time"

	"sfsEdgeStore/auth"
)

func TestAPIKeyGeneration(t *testing.T) {
	apiKey, err := auth.GenerateAPIKey("test-user", "admin", 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if apiKey.Key == "" {
		t.Error("Generated API key is empty")
	}

	if !apiKey.IsValid() {
		t.Error("Generated API key should be valid")
	}

	if apiKey.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", apiKey.Role)
	}

	if apiKey.UserID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got '%s'", apiKey.UserID)
	}
}

func TestRBACPermissions(t *testing.T) {
	testCases := []struct {
		role       string
		permission string
		expected   bool
	}{
		{"admin", "read", true},
		{"admin", "write", true},
		{"admin", "admin", true},
		{"admin", "backup", true},
		{"admin", "restore", true},

		{"user", "read", true},
		{"user", "write", true},
		{"user", "admin", false},
		{"user", "backup", true},
		{"user", "restore", false},

		{"readonly", "read", true},
		{"readonly", "write", false},
		{"readonly", "admin", false},
		{"readonly", "backup", false},
		{"readonly", "restore", false},

		{"invalid", "read", false},
	}

	for _, tc := range testCases {
		result := auth.HasPermission(auth.Role(tc.role), auth.Permission(tc.permission))
		if result != tc.expected {
			t.Errorf("Role '%s' should have permission '%s': expected %v, got %v", tc.role, tc.permission, tc.expected, result)
		}
	}
}

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

func TestAPIKeyExpiration(t *testing.T) {
	apiKey, err := auth.GenerateAPIKey("test-user", "admin", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if !apiKey.IsValid() {
		t.Error("API key should be valid initially")
	}

	time.Sleep(2 * time.Second)

	if apiKey.IsValid() {
		t.Error("API key should be expired after 2 seconds")
	}
}
```

### 运行测试

```bash
go test ./auth/test -v
```

## API 接口

### GenerateAPIKey 生成 API Key

```go
func GenerateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error)
```

### HasPermission 检查权限

```go
func HasPermission(role Role, permission Permission) bool
```

### ValidateRole 验证角色

```go
func ValidateRole(role string) bool
```

### NewAuthManager 创建认证管理器

```go
func NewAuthManager() *AuthManager
```

### AddAPIKey 添加 API Key

```go
func (am *AuthManager) AddAPIKey(apiKey *APIKey) error
```

### GetAPIKeyByKey 根据 Key 获取 API Key

```go
func (am *AuthManager) GetAPIKeyByKey(key string) (*APIKey, error)
```

### RevokeAPIKey 撤销 API Key

```go
func (am *AuthManager) RevokeAPIKey(key string) error
```

### ListAPIKeys 列出所有 API Key

```go
func (am *AuthManager) ListAPIKeys() ([]*APIKey, error)
```

### CreateAPIKey 创建 API Key

```go
func (am *AuthManager) CreateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error)
```