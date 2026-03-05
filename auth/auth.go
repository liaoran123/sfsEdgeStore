package auth

import (
	"errors"
	"time"

	"sfsdb-edgex-adapter-enterprise/database"
)

// AuthManager 认证管理器
type AuthManager struct {
	// 使用sfsDb存储API Key
}

// NewAuthManager 创建新的认证管理器
func NewAuthManager() *AuthManager {
	return &AuthManager{}
}

// AddAPIKey 添加API Key
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

// GetAPIKeyByKey 根据密钥获取API Key
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

// RevokeAPIKey 撤销API Key
func (am *AuthManager) RevokeAPIKey(key string) error {
	// 更新数据库中的记录
	updateRecord := map[string]any{
		"key":    key,
		"active": false,
	}

	return database.AuthTable.Update(&updateRecord)
}

// ListAPIKeys 列出所有API Key
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

// CleanExpiredKeys 清理过期的API Key
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

// StartCleanupTask 启动定期清理任务
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

// CreateAPIKey 创建新的API Key
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
