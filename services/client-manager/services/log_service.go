package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zgsm-ai/client-manager/dao"
	"github.com/zgsm-ai/client-manager/models"
)

/**
 * LogService handles business logic for log operations
 * @description
 * - Implements log processing business rules
 * - Validates log data
 * - Handles different log types
 */
type LogService struct {
	logDAO     *dao.LogDAO
	log        *logrus.Logger
	maxLogSize int64
}

type UploadLogArgs struct {
	ClientID    string `json:"client_id"`
	UserID      string `json:"user_id"`
	FileName    string `json:"file_name"`
	FirstLineNo int64  `json:"first_line_no"`
	LastLineNo  int64  `json:"end_line_no"`
}

type ListLogsArgs struct {
	ClientId string `form:"client_id"`
	UserId   string `form:"user_id"`
	FileName string `form:"file_name"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

type GetLogArgs struct {
	ClientID string `form:"client_id"`
	UserID   string `form:"user_id"`
	FileName string `form:"file_name"`
}

type LogStats struct {
	FirstLineNo int64 //首行编号
	LastLineNo  int64 //尾行编号
}

type Paginated struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

/**
 * NewLogService creates a new LogService instance
 * @param {dao.LogDAO} logDAO - Log data access object
 * @param {logrus.Logger} log - Logger instance
 * @returns {*LogService} New LogService instance
 */
/**
 * NewLogService creates a new LogService instance
 * @param {dao.LogDAO} logDAO - Log data access object
 * @param {logrus.Logger} log - Logger instance
 * @returns {*LogService} New LogService instance
 * @description
 * - Initializes LogService with default maxLogSize of 100MB (100 * 1024 * 1024 bytes)
 * - Can be configured by setting the maxLogSize field after creation
 */
func NewLogService(logDAO *dao.LogDAO, log *logrus.Logger) *LogService {
	// Default max log size: 50MB
	defaultMaxLogSize := int64(50 * 1024 * 1024)

	return &LogService{
		logDAO:     logDAO,
		log:        log,
		maxLogSize: defaultMaxLogSize,
	}
}

/**
 * UploadLog creates a new log record and saves file
 * @param {context.Context} ctx - Context for request cancellation
 * @param {UploadLogArgs} args - Log metadata arguments
 * @param {io.Reader} file - File content reader
 * @returns {string, error} File path and error if any
 * @description
 * - Validates log data
 * - Creates log record in database
 * - Creates directory and saves file to storage
 * - Logs creation operation
 * @throws
 * - Validation errors for invalid data
 * - Database creation errors
 * - File system errors for directory/file creation
 */
func (s *LogService) UploadLog(ctx context.Context, args *UploadLogArgs, file io.Reader) (string, error) {
	// Validate and extract log data
	err := s.validate(args)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"client_id": args.ClientID,
			"user_id":   args.UserID,
			"file_name": args.FileName,
		}).Error("Invalid arguments")
		return "", err
	}

	// Create file destination path
	destPath := filepath.Join("/data", args.ClientID, args.FileName)

	// Create directory if not exists
	destDir := filepath.Join("/data", args.ClientID)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		s.log.WithError(err).WithField("path", destDir).Error("Failed to create directory")
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file and save content
	destFile, err := os.Create(destPath)
	if err != nil {
		s.log.WithError(err).WithField("path", destPath).Error("Failed to create file")
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()

	// Copy file content
	size, err := io.Copy(destFile, file)
	if err != nil {
		s.log.WithError(err).WithField("path", destPath).Error("Failed to save file content")
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Create log
	log := &models.Log{
		ClientID:    args.ClientID,
		UserID:      args.UserID,
		FileName:    args.FileName,
		FirstLineNo: args.FirstLineNo,
		LastLineNo:  args.LastLineNo,
		Size:        size,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create log in database first
	err = s.logDAO.Upsert(ctx, log)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"client_id": log.ClientID,
			"user_id":   log.UserID,
			"file_name": log.FileName,
		}).Error("Failed to create log")
		return "", err
	}
	s.log.WithFields(logrus.Fields{
		"client_id": log.ClientID,
		"user_id":   log.UserID,
		"file_name": log.FileName,
		"path":      destPath,
	}).Info("Log and file created successfully")

	// Perform rollout cleanup to remove old log files if size exceeds limit
	s.rollout(ctx, args)

	return destPath, nil
}

/**
 * rollout handles log file cleanup based on size limit
 * @param {context.Context} ctx - Context for request cancellation
 * @param {UploadLogArgs} args - Log metadata arguments
 * @description
 * - Retrieves all logs for the same client and user
 * - Sorts logs by updated_at DESC to keep newest records
 * - Deletes old log records and their physical files when total size exceeds maxLogSize
 * - Always preserves the current file being saved (args.FileName)
 * - Prioritizes keeping recently saved records and their log files
 * - Errors are logged but do not affect the cleanup process
 */
func (s *LogService) rollout(ctx context.Context, args *UploadLogArgs) {
	// Retrieve all logs for the same client and user
	logs, _, err := s.logDAO.ListLogs(ctx, args.ClientID, args.UserID, "", 1, 1000)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"client_id": args.ClientID,
			"user_id":   args.UserID,
		}).Error("Failed to retrieve logs for rollout")
		return
	}

	// Calculate total size of all logs
	var totalSize int64
	for i := range logs {
		// Get actual file size if Size is 0
		if logs[i].Size == 0 {
			filePath := filepath.Join("/data", logs[i].ClientID, logs[i].FileName)
			if info, err := os.Stat(filePath); err == nil {
				logs[i].Size = info.Size()
				s.log.WithFields(logrus.Fields{
					"file_path": filePath,
					"size":      info.Size(),
				}).Debug("Updated log size from actual file")
			}
		}
		totalSize += logs[i].Size
	}

	// If total size is within limit, no cleanup needed
	if totalSize <= s.maxLogSize {
		s.log.WithFields(logrus.Fields{
			"client_id":  args.ClientID,
			"user_id":    args.UserID,
			"total_size": totalSize,
			"max_size":   s.maxLogSize,
		}).Debug("Total log size within limit, no cleanup needed")
		return
	}

	// Delete old logs starting from the oldest (end of the list)
	// Note: logs are already sorted by updated_at DESC from DAO
	for i := len(logs) - 1; i >= 0; i-- {
		log := logs[i]

		// Always preserve the current file being saved
		if log.FileName == args.FileName {
			continue
		}

		// Check if total size is still over limit
		if totalSize <= s.maxLogSize {
			break
		}

		// Delete physical file
		filePath := filepath.Join("/data", log.ClientID, log.FileName)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			s.log.WithError(err).WithField("path", filePath).Error("Failed to delete log file during rollout")
		} else if err == nil {
			s.log.WithField("path", filePath).Info("Successfully deleted log file during rollout")
		}

		// Delete from database
		err = s.logDAO.Delete(ctx, log.ID)
		if err != nil {
			s.log.WithError(err).WithField("id", log.ID).Error("Failed to delete log record during rollout")
			continue
		}

		// Update total size
		totalSize -= log.Size

		s.log.WithFields(logrus.Fields{
			"log_id":     log.ID,
			"client_id":  log.ClientID,
			"user_id":    log.UserID,
			"file_name":  log.FileName,
			"file_size":  log.Size,
			"total_size": totalSize,
		}).Info("Successfully deleted old log during rollout")
	}

	s.log.WithFields(logrus.Fields{
		"client_id":  args.ClientID,
		"user_id":    args.UserID,
		"total_size": totalSize,
		"max_size":   s.maxLogSize,
	}).Info("Rollout completed successfully")
}

/**
 * GetLogs retrieves logs for a specific client
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} clientID - Client identifier
 * @param {int} page - Page number
 * @param {int} pageSize - Number of items per page
 * @returns {map[string]interface{}, error} Response containing logs and pagination info
 * @description
 * - Validates client ID and pagination parameters
 * - Retrieves logs from database
 * - Returns structured response with pagination metadata
 * @throws
 * - Validation errors for invalid parameters
 * - Database query errors
 */
func (s *LogService) GetLogs(ctx context.Context, clientID, fname string) (string, error) {
	if clientID == "" {
		return "", &ValidationError{Field: "client_id", Message: "client_id is required"}
	}
	if fname == "" {
		return "", &ValidationError{Field: "file_name", Message: "file_name is required"}
	}

	_, _, err := s.logDAO.ListLogs(ctx, clientID, "", fname, 1, 10)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"client_id": clientID,
			"file_name": fname,
		}).Error("Failed to get logs by client")
		return "", err
	}

	return filepath.Join("/data", clientID, fname), nil
}

func (s *LogService) ListLogs(ctx context.Context, args *ListLogsArgs) (logs []models.Log, paging Paginated, err error) {
	if args.Page < 1 {
		args.Page = 1
	}
	if args.PageSize < 1 || args.PageSize > 100 {
		args.PageSize = 20
	}
	var total int64
	logs, total, err = s.logDAO.ListLogs(ctx, args.ClientId, args.UserId, args.FileName, args.Page, args.PageSize)
	if err != nil {
		s.log.WithError(err).WithFields(logrus.Fields{
			"page":      args.Page,
			"page_size": args.PageSize,
		}).Error("Failed to get logs by user")
		return
	}
	paging.Page = int64(args.Page)
	paging.PageSize = int64(args.PageSize)
	paging.Total = total
	paging.TotalPages = (total + int64(args.PageSize) - 1) / int64(args.PageSize)

	s.log.WithFields(logrus.Fields{
		"user_id":   args.UserId,
		"page":      args.Page,
		"page_size": args.PageSize,
		"total":     total,
	}).Info("Logs retrieved successfully by user")
	return
}

/**
 * DeleteOldLogs deletes logs older than specified date
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} beforeDate - Delete logs before this date
 * @returns {int64, error} Number of deleted records and error if any
 * @description
 * - Validates date parameter
 * - Performs cleanup of old log records
 * - Returns count of deleted records
 * @throws
 * - Validation errors for invalid date
 * - Database deletion errors
 */
func (s *LogService) DeleteOldLogs(ctx context.Context, beforeDate string) (int64, error) {
	// Validate date parameter
	if beforeDate == "" {
		return 0, &ValidationError{Field: "before_date", Message: "before_date is required"}
	}

	// Delete old logs
	count, err := s.logDAO.DeleteOldLogs(ctx, beforeDate)
	if err != nil {
		s.log.WithError(err).WithField("before_date", beforeDate).Error("Failed to delete old logs")
		return 0, err
	}

	s.log.WithFields(logrus.Fields{
		"before_date":   beforeDate,
		"deleted_count": count,
	}).Info("Old logs deleted successfully")

	return count, nil
}

/**
 * validateAndExtractLog validates and extracts log data
 * @param {map[string]interface{}} data - Log data
 * @returns {*models.Log, error} Validated log and error if any
 * @description
 * - Validates required log fields
 * - Extracts log data
 * - Creates log object
 * @throws
 * - Validation errors for missing required fields
 */
func (s *LogService) validate(args *UploadLogArgs) error {
	if args.ClientID == "" {
		return &ValidationError{Field: "client_id", Message: "client_id is required and must be a string"}
	}
	if args.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "user_id is required and must be a string"}
	}
	if args.FileName == "" {
		return &ValidationError{Field: "file_name", Message: "file_name is required and must be a string"}
	}

	return nil
}
