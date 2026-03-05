package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// APIKey API Key结构体
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

// GenerateAPIKey 生成新的API Key
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

// IsValid 检查API Key是否有效
func (k *APIKey) IsValid() bool {
	return k.Active && time.Now().Before(k.ExpiresAt)
}

// generateID 生成唯一ID
func generateID() string {
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(idBytes)
}
