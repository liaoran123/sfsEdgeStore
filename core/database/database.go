package database

import (
	"fmt"
	"log"
)

// Init 初始化数据库
func Init(dbPath string, useEncryption bool, encryptionKey, algorithm string) error {
	err := CreateDB(dbPath, useEncryption, encryptionKey, algorithm)
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}
	err = CreateReadingsTable()
	if err != nil {
		return fmt.Errorf("failed to create readings table: %v", err)
	}
	err = CreateAuthTable()
	if err != nil {
		return fmt.Errorf("failed to create auth table: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}
