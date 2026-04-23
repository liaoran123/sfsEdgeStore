package database

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
)

// RotateEncryptionKey 轮换加密密钥
func RotateEncryptionKey(newKey string) error {
	// 检查当前存储是否是加密的
	store := storage.GetDBManager().GetDB()
	if store == nil {
		return fmt.Errorf("database not initialized")
	}

	// 检查是否是加密存储
	encryptedStore, ok := store.(*storage.EncryptedStoreWrapper)
	if !ok {
		return fmt.Errorf("database is not encrypted")
	}

	// 准备新密钥
	masterKey := make([]byte, 32)
	copy(masterKey, []byte(newKey))
	for i := len(newKey); i < 32; i++ {
		masterKey[i] = 0
	}

	// 执行密钥轮换
	return encryptedStore.ReEncrypt(masterKey)
}

// GetEncryptionStatus 获取加密状态
func GetEncryptionStatus() (bool, string, error) {
	store := storage.GetDBManager().GetDB()
	if store == nil {
		return false, "", fmt.Errorf("database not initialized")
	}

	// 检查是否是加密存储
	encryptedStore, ok := store.(*storage.EncryptedStoreWrapper)
	if !ok {
		return false, "", nil
	}

	// 获取加密配置
	config := encryptedStore.GetEncryptionConfig()
	return true, config.Algorithm, nil
}
