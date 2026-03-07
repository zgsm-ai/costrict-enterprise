package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

type Database struct {
	db *gorm.DB
}

type DBConfig struct {
	Type         string
	Host         string
	Port         int
	Username     string
	Password     string
	DBName       string
	MaxIdleConns int
	MaxOpenConns int
}

var (
	globalDb *Database
	once     sync.Once
)

func InitGlobalDatabase(cfg *DBConfig) (*Database, error) {
	var err error
	once.Do(func() {
		dbErr := createDatabaseIfNotExists(cfg)
		if dbErr != nil {
			log.Fatal("unable to create database, %v", dbErr.Error())
		}
		globalDb, err = newDatabaseImpl(cfg)
	})
	if err != nil {
		return nil, err
	}
	return globalDb, nil
}

func GetDB() *Database {
	return globalDb
}

// newDatabaseImpl creates a new Database globalDb
func newDatabaseImpl(cfg *DBConfig) (*Database, error) {
	newLogger := logger.New(
		log.NewGormWriter(),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	dialector, err := createDialector(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dialector: %w", err)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
		NowFunc: func() time.Time {
			return time.Now()
		},
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := configureConnection(db, cfg); err != nil {
		return nil, fmt.Errorf("failed to configure connection: %v", err)
	}

	db_ := &Database{db: db}

	if err := db_.HealthCheck(context.Background()); err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	if err := db_.AutoMigrate(
		&AuthUser{},
		&SyncLock{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %v", err)
	}
	return db_, nil
}

func createDialector(cfg *DBConfig) (gorm.Dialector, error) {
	switch cfg.Type {
	case "mysql":
		return createMySQLDialector(cfg)
	case "postgres":
		return createPostgreSQLDialector(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

func createMySQLDialector(cfg *DBConfig) (gorm.Dialector, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	return mysql.Open(dsn), nil
}

func createPostgreSQLDialector(cfg *DBConfig) (gorm.Dialector, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai connect_timeout=30",
		cfg.Host, cfg.Username, cfg.Password, cfg.DBName, cfg.Port)

	return postgres.Open(dsn), nil
}

func createDatabaseIfNotExists(cfg *DBConfig) error {
	switch cfg.Type {
	case "mysql":
		return createMySQLDatabaseIfNotExists(cfg)
	case "postgres":
		return createPostgresDatabaseIfNotExists(cfg)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

func createMySQLDatabaseIfNotExists(cfg *DBConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	dialector := mysql.Open(dsn)
	tempDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server to create database: %w", err)
	}
	defer func() {
		sqlDB, closeErr := tempDB.DB()
		if closeErr == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()

	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", cfg.DBName)
	err = tempDB.Exec(createDBSQL).Error
	if err != nil {
		return fmt.Errorf("failed to execute CREATE DATABASE statement for %s: %w", cfg.DBName, err)
	}
	return nil
}

func createPostgresDatabaseIfNotExists(cfg *DBConfig) error {
	defaultDB := "postgres"
	tempDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai connect_timeout=30",
		cfg.Host, cfg.Username, cfg.Password, defaultDB, cfg.Port)
	dialector := postgres.Open(tempDSN)
	tempDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to default database %s to create target database %s: %w", defaultDB, cfg.DBName, err)
	}
	defer func() {
		sqlDB, closeErr := tempDB.DB()
		if closeErr == nil {
			_ = sqlDB.Close()
		}
	}()

	// Check if database exists
	var exists int64
	err = tempDB.Raw("SELECT 1 FROM pg_database WHERE datname = $1", cfg.DBName).Count(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check if database %s exists: %w", cfg.DBName, err)
	}

	// Create database if it doesn't exist
	if exists == 0 {
		createDBSQL := fmt.Sprintf("CREATE DATABASE %s;", cfg.DBName)
		err = tempDB.Exec(createDBSQL).Error
		if err != nil {
			return fmt.Errorf("failed to execute CREATE DATABASE statement for %s: %w", cfg.DBName, err)
		}
	}

	return nil
}

func (d *Database) AutoMigrate(models ...any) error {
	return d.db.AutoMigrate(models...)
}

func configureConnection(db *gorm.DB, cfg *DBConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	return nil
}

func (d *Database) HealthCheck(ctx context.Context) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("database instance is nil")
	}

	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return nil
	}

	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %v", err)
	}

	return sqlDB.Close()
}
