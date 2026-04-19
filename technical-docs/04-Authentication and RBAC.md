# Authentication and RBAC

## Overview

sfsEdgeStore provides API Key-based authentication mechanism and Role-Based Access Control (RBAC) system.

## API Key Structure

### APIKey Struct

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

## Generate API Key

### GenerateAPIKey Function

```go
// auth/api_key.go:22-45
func GenerateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}

	key := hex.EncodeToString(keyBytes)

	hash := key

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

### Check API Key Validity

```go
// auth/api_key.go:48-50
func (k *APIKey) IsValid() bool {
	return k.Active && time.Now().Before(k.ExpiresAt)
}
```

## RBAC Roles and Permissions

### Role Definitions

```go
// auth/rbac.go:6-11
const (
	RoleAdmin     Role = "admin"
	RoleUser      Role = "user"
	RoleReadOnly  Role = "readonly"
)
```

### Permission Definitions

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

### Role Permission Mapping

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

### Check Permission

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

### Get Role Permissions

```go
// auth/rbac.go:61-66
func GetRolePermissions(role Role) []Permission {
	if permissions, ok := RolePermissions[role]; ok {
		return permissions
	}
	return []Permission{}
}
```

### Validate Role

```go
// auth/rbac.go:69-72
func ValidateRole(role string) bool {
	_, ok := RolePermissions[Role(role)]
	return ok
}
```

## AuthManager

### AuthManager Structure

```go
// auth/auth.go:11-13
type AuthManager struct {
}
```

### Create AuthManager

```go
// auth/auth.go:16-18
func NewAuthManager() *AuthManager {
	return &AuthManager{}
}
```

### Add API Key

```go
// auth/auth.go:21-37
func (am *AuthManager) AddAPIKey(apiKey *APIKey) error {
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

	_, err := database.AuthTable.Insert(&record)
	return err
}
```

### Get API Key by Key

```go
// auth/auth.go:40-76
func (am *AuthManager) GetAPIKeyByKey(key string) (*APIKey, error) {
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

### Revoke API Key

```go
// auth/auth.go:79-87
func (am *AuthManager) RevokeAPIKey(key string) error {
	updateRecord := map[string]any{
		"key":    key,
		"active": false,
	}

	return database.AuthTable.Update(&updateRecord)
}
```

### List All API Keys

```go
// auth/auth.go:90-126
func (am *AuthManager) ListAPIKeys() ([]*APIKey, error) {
	fields := map[string]any{
		"key": nil,
	}

	iter, err := database.AuthTable.Search(&fields)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	defer records.Release()

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

### Clean Expired Keys

```go
// auth/auth.go:129-154
func (am *AuthManager) CleanExpiredKeys() (int, error) {
	apiKeys, err := am.ListAPIKeys()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	cleanupCount := 0

	for _, apiKey := range apiKeys {
		if now.After(apiKey.ExpiresAt) {
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

### Start Cleanup Task

```go
// auth/auth.go:157-174
func (am *AuthManager) StartCleanupTask(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C
			count, err := am.CleanExpiredKeys()
			if err != nil {
				continue
			}
			if count > 0 {
			}
		}
	}()
}
```

### Create API Key

```go
// auth/auth.go:177-196
func (am *AuthManager) CreateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error) {
	if !ValidateRole(role) {
		return nil, errors.New("invalid role")
	}

	apiKey, err := GenerateAPIKey(userID, role, expiresIn)
	if err != nil {
		return nil, err
	}

	err = am.AddAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}
```

## Testing

### Authentication Testing

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

### Running Tests

```bash
go test ./auth/test -v
```

## API Interface

### GenerateAPIKey Generate API Key

```go
func GenerateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error)
```

### HasPermission Check Permission

```go
func HasPermission(role Role, permission Permission) bool
```

### ValidateRole Validate Role

```go
func ValidateRole(role string) bool
```

### NewAuthManager Create Authentication Manager

```go
func NewAuthManager() *AuthManager
```

### AddAPIKey Add API Key

```go
func (am *AuthManager) AddAPIKey(apiKey *APIKey) error
```

### GetAPIKeyByKey Get API Key by Key

```go
func (am *AuthManager) GetAPIKeyByKey(key string) (*APIKey, error)
```

### RevokeAPIKey Revoke API Key

```go
func (am *AuthManager) RevokeAPIKey(key string) error
```

### ListAPIKeys List All API Keys

```go
func (am *AuthManager) ListAPIKeys() ([]*APIKey, error)
```

### CreateAPIKey Create API Key

```go
func (am *AuthManager) CreateAPIKey(userID, role string, expiresIn time.Duration) (*APIKey, error)
```