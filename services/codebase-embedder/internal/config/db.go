package config

import "time"

type Database struct {
	Driver      string
	DataSource  string
	AutoMigrate struct {
		Enable bool
	}
	Pool struct {
		MaxIdleConns    int           `json:",default=10"`  // 最大空闲连接数
		MaxOpenConns    int           `json:",default=100"` // 最大打开连接数
		ConnMaxLifetime time.Duration `json:",default=1h"`  // 连接最大生命周期
		ConnMaxIdleTime time.Duration `json:",default=30m"` // 空闲连接最大生命周期
	}
	LogLevel string `json:",default=info"` // 日志级别：silent, error, warn, info
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr              string
	Password          string        `json:",optional"`
	DB                int           `json:",default=0"`
	PoolSize          int           `json:",default=10"`
	MinIdleConn       int           `json:",default=10"`
	ConnectTimeout    time.Duration `json:",default=10s"`
	ReadTimeout       time.Duration `json:",default=10s"`
	WriteTimeout      time.Duration `json:",default=10s"`
	DefaultExpiration time.Duration `json:",default=24h"` // 默认过期时间
}
