package mocks

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockDB 封装了 mock 数据库相关的组件
type MockDB struct {
	GormDB *gorm.DB
	Mock   sqlmock.Sqlmock
	sqlDB  *sql.DB
}

// NewMockDB 创建一个新的 mock 数据库实例
func NewMockDB() (*MockDB, error) {
	// 创建 sqlmock
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlmock: %w", err)
	}

	// 创建 GORM 实例
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true, // 必须设置为 true 以兼容 sqlmock
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm db: %w", err)
	}

	return &MockDB{
		GormDB: gormDB,
		Mock:   mock,
		sqlDB:  sqlDB,
	}, nil
}

// Close 关闭数据库连接
func (m *MockDB) Close() error {
	return m.sqlDB.Close()
}

// Begin 开始事务
func (m *MockDB) Begin() {
	m.Mock.ExpectBegin()
}

// Commit 提交事务
func (m *MockDB) Commit() {
	m.Mock.ExpectCommit()
}

// Rollback 回滚事务
func (m *MockDB) Rollback() {
	m.Mock.ExpectRollback()
}

// ExpectationsWereMet 验证所有期望的数据库操作是否都已完成
func (m *MockDB) ExpectationsWereMet() error {
	return m.Mock.ExpectationsWereMet()
}

// MustExpectationsWereMet 验证所有期望的数据库操作是否都已完成，如果失败则调用 t.Fatal
func (m *MockDB) MustExpectationsWereMet(t *testing.T) {
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatalf("Unfulfilled expectations: %v", err)
	}
}
