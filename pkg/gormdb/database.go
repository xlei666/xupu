// Package gormdb GORM数据库连接（用于用户认证）
package gormdb

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/xlei/xupu/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
	once sync.Once
)

// Get 获取GORM数据库实例
func Get() *gorm.DB {
	if db == nil {
		var err error
		db, err = initDB()
		if err != nil {
			log.Fatalf("Failed to initialize GORM database: %v", err)
		}
	}
	return db
}

// initDB 初始化数据库连接
func initDB() (*gorm.DB, error) {
	// 从环境变量获取数据库配置
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "xupu")

	// 构建连接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, port, user, password, dbname)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&models.User{},
		&models.AuthToken{},
		&models.Chapter{}, // 章节表
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("GORM database connected successfully")
	return db, nil
}

// getEnv 获取环境变量，支持默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
