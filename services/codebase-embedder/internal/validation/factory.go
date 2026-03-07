package validation

import (
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// NewValidationCreator 创建验证组件的工厂函数
func NewValidationCreator(config *types.ValidationConfig) (FileValidator, error) {
	if config == nil {
		// 使用默认配置
		config = &types.ValidationConfig{
			Enabled:        true,
			MaxConcurrency: 10,
			FailOnMismatch: false,
			CheckContent:   false,
			LogLevel:       "info",
		}
	}

	// 创建文件验证器
	validator := NewFileValidator(config)

	return validator, nil
}

// NewDefaultValidationConfig 创建默认验证配置
func NewDefaultValidationConfig() *types.ValidationConfig {
	return &types.ValidationConfig{
		Enabled:        true,
		MaxConcurrency: 10,
		FailOnMismatch: false,
		CheckContent:   false,
		LogLevel:       "info",
		SkipPatterns:   []string{},
	}
}
