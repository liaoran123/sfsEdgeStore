package database

import (
	"fmt"
	"log"
	"sfsEdgeStore/config"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// CreateDB 创建数据库
func CreateDB(dbPath string, useEncryption bool, encryptionKey, algorithm string) error {
	// 设置数据库场景配置
	var dbScenario string
	cfgMgr := config.GetConfigManager()
	if cfgMgr != nil && cfgMgr.GetConfig() != nil {
		dbScenario = cfgMgr.GetConfig().DBScenario
		log.Printf("Using database scenario: %s", dbScenario)
	} else {
		dbScenario = storage.ScenarioEdge
		log.Printf("Using default database scenario: %s", dbScenario)
	}

	// 打开数据库，使用场景配置
	var err error
	if useEncryption == false && encryptionKey == "" {
		// 使用场景打开普通数据库
		_, err = storage.GetDBManager().OpenDBWithScenario(dbPath, dbScenario)
		return err
	}

	// 生成32字节的加密密钥
	masterKey := make([]byte, 32)
	copy(masterKey, []byte(encryptionKey))
	// 确保密钥长度为32字节
	for i := len(encryptionKey); i < 32; i++ {
		masterKey[i] = 0
	}
	// 创建加密配置
	encryptConfig := &storage.EncryptionConfig{
		Enabled:   true,
		Algorithm: algorithm,
		MasterKey: masterKey,
	}
	// 使用场景和加密打开数据库
	_, err = storage.GetDBManager().OpenDBWithScenario(dbPath, dbScenario, encryptConfig)

	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	return nil
}

// Table 全局表实例
var Table *engine.Table

// 创建表
func CreateReadingsTable() error {
	// 创建或获取表
	tableName := "edgex_readings"
	var createErr error
	Table, createErr = engine.TableNew(tableName)
	if createErr != nil {
		return fmt.Errorf("failed to create table: %v", createErr)
	}

	// 设置表字段
	fields := map[string]any{
		"id":         "",
		"deviceName": "",
		"deviceType": "",
		"reading":    "",
		"value":      0.0,
		"valueType":  "",
		"baseType":   "",
		"timestamp":  int64(0),
		"metadata":   "",
	}
	if err := Table.SetFields(fields); err != nil {
		return fmt.Errorf("failed to set table fields: %v", err)
	}

	// 创建组合主键索引 (deviceName + timestamp)
	primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
	if err != nil {
		return fmt.Errorf("failed to create primary key: %v", err)
	}
	primaryKey.AddFields("deviceName", "timestamp") // 创建deviceName和timestamp的组合主键
	if err := Table.CreateIndex(primaryKey); err != nil {
		// 忽略索引已存在的错误
		if err.Error() != "index already exists" {
			return fmt.Errorf("failed to create primary key index: %v", err)
		}
	}
	/*  查询业务该索引是必须的，但是当前项目并不需要。
		// 创建设备类型索引
		deviceTypeIndex, err := engine.DefaultPrimaryKeyNew("deviceType_idx")
		if err != nil {
			return fmt.Errorf("failed to create device type index: %v", err)
		}
		deviceTypeIndex.AddFields("deviceType") // 创建deviceType索引
		if err := Table.CreateIndex(deviceTypeIndex); err != nil {
			// 忽略索引已存在的错误
			if err.Error() != "index already exists" {
				return fmt.Errorf("failed to create device type index: %v", err)
			}
		}
	 	fieldTypeLen := map[string]uint8{
			"deviceName": 64, //组合主键使用了deviceName的不定长字符串类型，这里必须注册固定大小
		}
		Table.SetFieldTypeLen(&fieldTypeLen) //设置单个表字段长度
	*/
	return nil
}

// AuthTable 认证表实例
var AuthTable *engine.Table

func CreateAuthTable() error {
	var createErr error
	// 创建认证表
	authTableName := "edgex_auth"
	AuthTable, createErr = engine.TableNew(authTableName)
	if createErr != nil {
		return fmt.Errorf("failed to create auth table: %v", createErr)
	}

	// 设置认证表字段
	authFields := map[string]any{
		"id":         "",
		"key":        "",
		"hash":       "",
		"user_id":    "",
		"role":       "",
		"created_at": int64(0),
		"expires_at": int64(0),
		"active":     false,
	}
	if err := AuthTable.SetFields(authFields); err != nil {
		return fmt.Errorf("failed to set auth table fields: %v", err)
	}

	// 创建认证表主键索引
	authPrimaryKey, err := engine.DefaultPrimaryKeyNew("auth_pk")
	if err != nil {
		return fmt.Errorf("failed to create auth primary key: %v", err)
	}
	authPrimaryKey.AddFields("key") // 使用key作为主键
	if err := AuthTable.CreateIndex(authPrimaryKey); err != nil {
		// 忽略索引已存在的错误
		if err.Error() != "index already exists" {
			return fmt.Errorf("failed to create auth primary key index: %v", err)
		}
	}
	return nil
}
