package validation

import (
	"context"

	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// FileValidator 文件验证器接口
type FileValidator interface {
	// Validate 验证文件
	Validate(ctx context.Context, params *types.ValidationParams) (*types.ValidationResult, error)
	// SetConfig 设置配置
	SetConfig(config *types.ValidationConfig)
}

// SyncMetadataReader 同步元数据读取器接口
type SyncMetadataReader interface {
	// ReadMetadata 读取元数据
	ReadMetadata(ctx context.Context, path string) (*types.SyncMetadata, error)
	// ValidateMetadata 验证元数据格式
	ValidateMetadata(metadata *types.SyncMetadata) error
	// GetMetadataPath 获取元数据文件路径
	GetMetadataPath(extractPath string) string
}

// FileChecker 文件检查器接口
type FileChecker interface {
	// CheckFileExists 检查文件是否存在
	CheckFileExists(ctx context.Context, filePath string) (bool, error)
	// CheckFileMatch 检查文件是否匹配
	CheckFileMatch(ctx context.Context, expectedPath, actualPath string) (bool, error)
	// GetFileStats 获取文件统计信息
	GetFileStats(ctx context.Context, filePath string) (*types.FileStats, error)
}

// ValidationReporter 验证结果报告器接口
type ValidationReporter interface {
	// Report 报告验证结果
	Report(ctx context.Context, result *types.ValidationResult) error
	// Log 记录验证日志
	Log(ctx context.Context, result *types.ValidationResult) error
}
