package config

import (
	"errors"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		UserInfoHeader string
	}
	Database    Database
	Redis       RedisConfig
	IndexTask   IndexTaskConf
	VectorStore VectorStoreConf
	Cleaner     CleanerConf
	Validation  ValidationConfig
	TokenLimit  TokenLimitConf
	HealthCheck HealthCheckConf
}

// TokenLimitConf token限流配置
type TokenLimitConf struct {
	MaxRunningTasks int  `json:"max_running_tasks" yaml:"max_running_tasks"`
	Enabled         bool `json:"enabled" yaml:"enabled"`
}

// HealthCheckConf 探活接口配置
type HealthCheckConf struct {
	Enabled bool          `json:"enabled" yaml:"enabled"`
	URL     string        `json:"url" yaml:"url"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// Validate 实现 Validator 接口
func (c Config) Validate() error {
	if len(c.Name) == 0 {
		return errors.New("name 不能为空")
	}
	return nil
}
