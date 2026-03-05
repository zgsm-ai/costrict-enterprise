package internal

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zgsm-ai/client-manager/models"
)

// Global database instance
var DB *gorm.DB

/**
 * InitDB initializes the database connection
 * @returns {gorm.DB, error} Database connection and error if any
 * @description
 * - Creates SQLite database connection
 * - Auto-migrates database models
 * - Sets database connection pool settings
 * - Configures logging
 * @throws
 * - Database connection errors
 * - Migration errors
 */
func InitDB() (*gorm.DB, error) {
	// Get DSN from configuration
	dsn := "./data/client-manager.db" // Default DSN, should be from config

	// Configure GORM logger
	newLogger := logger.New(
		logrus.New(),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Warn,
			Colorful:      false,
		},
	)

	// Connect to database
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate models
	err = autoMigrate(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Store global instance
	DB = db

	return db, nil
}

/**
 * autoMigrate performs database migration for all models
 * @param {gorm.DB} db - Database connection
 * @returns {error} Error if migration fails
 * @description
 * - Migrates all defined models
 * - Creates tables if they don't exist
 * - Updates table structures if needed
 * @throws
 * - Migration errors
 */
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Log{},
	)
}

/**
 * GetDB returns the global database instance
 * @returns {gorm.DB} Database connection
 * @description
 * - Provides access to the global database instance
 * - Returns nil if database is not initialized
 */
func GetDB() *gorm.DB {
	return DB
}

/**
 * CloseDB closes the database connection
 * @description
 * - Closes the database connection
 * - Should be called on application shutdown
 * @throws
 * - Database close errors
 */
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
