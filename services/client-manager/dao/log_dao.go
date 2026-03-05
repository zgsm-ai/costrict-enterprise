package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zgsm-ai/client-manager/models"
)

/**
 * LogDAO handles data access operations for log data
 * @description
 * - Provides CRUD operations for log data using GORM
 * - Supports client and user based log filtering
 * - Implements database operations for performance optimization
 */
type LogDAO struct {
	db  *gorm.DB
	log *logrus.Logger
}

/**
 * NewLogDAO creates a new LogDAO instance
 * @param {*gorm.DB} db - Database connection
 * @param {logrus.Logger} log - Logger instance
 * @returns {*LogDAO} New LogDAO instance
 */
func NewLogDAO(db *gorm.DB, log *logrus.Logger) *LogDAO {
	return &LogDAO{
		db:  db,
		log: log,
	}
}

/**
 * Upsert creates or updates a log record
 * @param {context.Context} ctx - Context for request cancellation
 * @param {*models.Log} log - Log data to upsert
 * @returns {error} Error if any
 * @description
 * - Creates new log record if not exists
 * - Updates existing record if found
 * - Uses ClientID and FileName as unique identifier
 * - Logs upsert operation
 * @throws
 * - Database operation errors
 */
func (dao *LogDAO) Upsert(ctx context.Context, log *models.Log) error {
	if dao.db == nil {
		return fmt.Errorf("Database is not initialized")
	}

	// Set timestamps
	now := time.Now()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}
	log.UpdatedAt = now

	// Check if log record exists
	var existingLog models.Log
	err := dao.db.Where("client_id = ? AND file_name = ?", log.ClientID, log.FileName).First(&existingLog).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record
		err = dao.db.Create(log).Error
		if err != nil {
			dao.log.WithError(err).Error("Failed to create log")
			return err
		}
	} else if err != nil {
		// Database error
		dao.log.WithError(err).Error("Failed to check existing log")
		return err
	} else {
		// Update existing record
		log.ID = existingLog.ID
		err = dao.db.Save(log).Error
		if err != nil {
			dao.log.WithError(err).Error("Failed to update log")
			return err
		}
	}

	dao.log.WithFields(logrus.Fields{
		"client_id": log.ClientID,
		"file_name": log.FileName,
		"user_id":   log.UserID,
	}).Debug("Successfully upserted log")

	return nil
}

/**
 * ListLogs retrieves logs with filtering and pagination
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} clientID - Client identifier filter (optional)
 * @param {string} userID - User identifier filter (optional)
 * @param {string} fileName - File name filter (optional)
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {[]models.Log, int64, error} List of logs, total count, and error
 * @description
 * - Retrieves log records with optional filtering
 * - Supports pagination for large datasets
 * - Returns total count for frontend pagination
 * - Combines multiple filters with AND logic
 * @throws
 * - Database query errors
 */
func (dao *LogDAO) ListLogs(ctx context.Context, clientID, userID, fileName string, page, pageSize int) ([]models.Log, int64, error) {
	if dao.db == nil {
		return nil, 0, fmt.Errorf("Database is not initialized")
	}

	// Build database query
	query := dao.db.Model(&models.Log{})

	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if fileName != "" {
		query = query.Where("file_name = ?", fileName)
	}

	// Get total count
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		dao.log.WithError(err).Error("Failed to count logs")
		return nil, 0, err
	}

	// Calculate pagination
	offset := (page - 1) * pageSize

	// Execute query with pagination and ordering
	var logs []models.Log
	err = query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		dao.log.WithError(err).Error("Failed to list logs")
		return nil, 0, err
	}

	return logs, total, nil
}

/**
 * DeleteOldLogs deletes logs older than specified date
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} beforeDate - Delete logs before this date
 * @returns {int64, error} Number of deleted records and error if any
 * @description
 * - Performs cleanup of old log records
 * - Uses database delete operation for bulk deletion
 * - Returns count of deleted records
 * - Logs deletion operation
 * @throws
 * - Database delete errors
 */
func (dao *LogDAO) DeleteOldLogs(ctx context.Context, beforeDate string) (int64, error) {
	if dao.db == nil {
		return 0, fmt.Errorf("Database is not initialized")
	}

	// Parse the before date
	parsedDate, err := time.Parse("2006-01-02", beforeDate)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %w", err)
	}

	// Execute delete operation and get count
	result := dao.db.Where("updated_at < ?", parsedDate).Delete(&models.Log{})
	if result.Error != nil {
		dao.log.WithError(result.Error).Error("Failed to delete old logs")
		return 0, result.Error
	}

	deletedCount := result.RowsAffected

	dao.log.WithFields(logrus.Fields{
		"before_date":   beforeDate,
		"deleted_count": deletedCount,
	}).Info("Successfully deleted old logs")

	return deletedCount, nil
}

/**
* Delete deletes a log record by ID from database
* @param {context.Context} ctx - Context for request cancellation
* @param {uint} id - Log record ID to delete
* @returns {error} Error if any
* @description
* - Deletes a single log record by ID from database only
* - Does not delete the physical file (handled by service layer)
* - Logs deletion operation
* @throws
* - Database delete errors
* - Record not found errors
 */
func (dao *LogDAO) Delete(ctx context.Context, id uint) error {
	if dao.db == nil {
		return fmt.Errorf("Database is not initialized")
	}

	// Delete from database
	err := dao.db.Delete(&models.Log{}, id).Error
	if err != nil {
		dao.log.WithError(err).WithField("id", id).Error("Failed to delete log record")
		return err
	}

	dao.log.WithField("id", id).Info("Successfully deleted log record from database")

	return nil
}

/**
* GetByID retrieves a log record by ID
* @param {context.Context} ctx - Context for request cancellation
* @param {uint} id - Log record ID
* @returns {models.Log, error} Log record and error if any
* @description
* - Retrieves a single log record by ID
* @throws
* - Database query errors
* - Record not found errors
 */
func (dao *LogDAO) GetByID(ctx context.Context, id uint) (models.Log, error) {
	if dao.db == nil {
		return models.Log{}, fmt.Errorf("Database is not initialized")
	}

	var log models.Log
	err := dao.db.Where("id = ?", id).First(&log).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return log, fmt.Errorf("log record not found")
		}
		dao.log.WithError(err).WithField("id", id).Error("Failed to get log record")
		return log, err
	}

	return log, nil
}
