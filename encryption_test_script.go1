package main

import (
	"fmt"
	"log"
	"sfsEdgeStore/config"
	"sfsEdgeStore/database"
)

func mainencryption() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 打印当前配置
	fmt.Println("Current configuration:")
	fmt.Printf("DBPath: %s\n", cfg.DBPath)
	fmt.Printf("DBUseEncryption: %v\n", cfg.DBUseEncryption)
	fmt.Printf("DBEncryptionAlgorithm: %s\n", cfg.DBEncryptionAlgorithm)
	fmt.Printf("DBEncryptionKey length: %d\n", len(cfg.DBEncryptionKey))

	// 初始化数据库
	if err = database.Init(cfg.DBPath, cfg.DBUseEncryption, cfg.DBEncryptionKey, cfg.DBEncryptionAlgorithm); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("Database initialized successfully")
	fmt.Println("Encryption test completed")
}
