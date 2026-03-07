package database

import (
	"fmt"
	"strings"

	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// New 创建一个新的 GORM PostgreSQL 数据库连接
func New(c config.Database) (*gorm.DB, error) {
	// 设置日志级别
	var logLevel logger.LogLevel
	switch strings.ToLower(c.LogLevel) {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: logger.Default.LogMode(logLevel),
	}

	// 构建 DSN
	if c.DataSource == "" {
		return nil, fmt.Errorf("database data source is required")
	}

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(c.DataSource), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(c.Pool.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.Pool.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.Pool.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(c.Pool.ConnMaxIdleTime)

	return db, nil
}

// CloseDB 关闭数据库连接
func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
