package models

import (
	"time"
)

/**
 * Log model represents client log entries
 * @description
 * - Stores log data from clients
 * - Includes client and user identification
 * - Supports structured logging with module information
 */
type Log struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ClientID    string    `json:"client_id" gorm:"index;not null"`
	UserID      string    `json:"user_id" gorm:"index"`
	FileName    string    `json:"file_name" gorm:"index;not null"`
	FirstLineNo int64     `json:"first_line_no"`
	LastLineNo  int64     `json:"end_line_no"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

/**
 * TableName returns the table name for Log model
 * @returns {string} Database table name
 */
func (Log) TableName() string {
	return "logs"
}
