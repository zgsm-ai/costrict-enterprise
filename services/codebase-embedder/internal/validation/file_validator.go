package validation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

var (
	ErrValidationFailed = errors.New("file validation failed")
)

// FileValidatorImpl 文件验证器实现
type FileValidatorImpl struct {
	config         *types.ValidationConfig
	metadataReader SyncMetadataReader
	fileChecker    FileChecker
	reporter       ValidationReporter
}

// NewFileValidator 创建文件验证器
func NewFileValidator(config *types.ValidationConfig) FileValidator {
	return &FileValidatorImpl{
		config:         config,
		metadataReader: NewSyncMetadataReader(),
		fileChecker:    NewFileChecker(config.SkipPatterns),
		reporter:       NewValidationReporter(config.LogLevel),
	}
}

// Validate 执行文件验证
func (v *FileValidatorImpl) Validate(ctx context.Context, params *types.ValidationParams) (*types.ValidationResult, error) {
	tracer.WithTrace(ctx).Infof("starting file validation, metadata_path: %s, extract_path: %s",
		params.MetadataPath, params.ExtractPath)

	startTime := time.Now()

	// 如果没有提供元数据路径，尝试从解压路径推导
	if params.MetadataPath == "" {
		params.MetadataPath = v.metadataReader.GetMetadataPath(params.ExtractPath)
		tracer.WithTrace(ctx).Infof("[DEBUG] Generated metadata path from extract path: '%s'", params.MetadataPath)
	} else {
		tracer.WithTrace(ctx).Infof("[DEBUG] Using provided metadata path: '%s'", params.MetadataPath)
	}

	// 强制使用解压路径而不是接口传入的路径
	if params.MetadataPath != params.ExtractPath {
		correctPath := v.metadataReader.GetMetadataPath(params.ExtractPath)
		tracer.WithTrace(ctx).Infof("[DEBUG] Correcting metadata path from '%s' to '%s'", params.MetadataPath, correctPath)
		params.MetadataPath = correctPath
	}

	// 读取元数据
	metadata, err := v.metadataReader.ReadMetadata(ctx, params.MetadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// 验证元数据格式
	if err := v.metadataReader.ValidateMetadata(metadata); err != nil {
		return nil, fmt.Errorf("invalid metadata: %w", err)
	}

	// 初始化验证结果
	result := &types.ValidationResult{
		TotalFiles: len(metadata.FileList),
		Details:    make([]types.ValidationDetail, 0),
		Status:     types.ValidationStatusSuccess,
		Timestamp:  time.Now(),
	}

	// 并发验证文件
	if err := v.validateFilesConcurrently(ctx, metadata, params.ExtractPath, result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 计算最终状态
	v.calculateFinalStatus(result)

	// 记录验证结果
	if err := v.reporter.Report(ctx, result); err != nil {
		tracer.WithTrace(ctx).Errorf("failed to report validation result: %v", err)
	}

	// 记录验证日志
	if err := v.reporter.Log(ctx, result); err != nil {
		tracer.WithTrace(ctx).Errorf("failed to log validation result: %v", err)
	}

	tracer.WithTrace(ctx).Infof("file validation completed, cost: %d ms, status: %s",
		time.Since(startTime).Milliseconds(), result.Status)

	return result, nil
}

// SetConfig 设置配置
func (v *FileValidatorImpl) SetConfig(config *types.ValidationConfig) {
	v.config = config
	v.fileChecker.(*FileCheckerImpl).SetSkipPatterns(config.SkipPatterns)
	v.reporter.(*ValidationReporterImpl).SetLogLevel(config.LogLevel)
}

// validateFilesConcurrently 并发验证文件
func (v *FileValidatorImpl) validateFilesConcurrently(
	ctx context.Context,
	metadata *types.SyncMetadata,
	extractPath string,
	result *types.ValidationResult,
) error {
	maxConcurrency := v.config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 10 // 默认值
	}

	pool, err := ants.NewPool(maxConcurrency)
	if err != nil {
		return fmt.Errorf("create ants pool failed: %w", err)
	}
	defer pool.Release()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		runErrs []error
		done    = make(chan struct{})
	)

	// 为每个文件创建验证任务
	for filePath, expectedStatus := range metadata.FileList {
		select {
		case <-ctx.Done():
			return fmt.Errorf("validation cancelled")
		default:
			wg.Add(1)
			if err := pool.Submit(func() {
				defer wg.Done()

				// 构建完整文件路径
				fullPath := filepath.Join(extractPath, filePath)

				// 验证单个文件
				detail, err := v.validateSingleFile(ctx, fullPath, expectedStatus)

				mu.Lock()
				if err != nil {
					runErrs = append(runErrs, err)
				}
				result.Details = append(result.Details, *detail)

				// 更新统计信息
				switch detail.Status {
				case types.FileStatusMatched:
					result.MatchedFiles++
				case types.FileStatusMismatched:
					result.MismatchedFiles++
				case types.FileStatusSkipped:
					result.SkippedFiles++
				}
				mu.Unlock()
			}); err != nil {
				wg.Done()
				mu.Lock()
				runErrs = append(runErrs, fmt.Errorf("submit validation task failed: %w", err))
				mu.Unlock()
			}
		}
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待任务完成或上下文取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("validation cancelled")
	case <-done:
		if len(runErrs) > 0 {
			return fmt.Errorf("file validation failed: %w", errors.Join(runErrs...))
		}
		return nil
	}
}

// validateSingleFile 验证单个文件
func (v *FileValidatorImpl) validateSingleFile(
	ctx context.Context,
	filePath string,
	expectedStatus string,
) (*types.ValidationDetail, error) {
	detail := &types.ValidationDetail{
		FilePath: filePath,
		Expected: expectedStatus,
	}

	// 添加详细的文件存在性检查日志
	tracer.WithTrace(ctx).Infof("[DEBUG] VALIDATION_DETAIL: Starting file validation for path: '%s'", filePath)
	tracer.WithTrace(ctx).Infof("[DEBUG] VALIDATION_DETAIL: Expected file status: '%s'", expectedStatus)

	// 检查文件路径是否包含 .shenma_sync
	if strings.Contains(filePath, ".shenma_sync") {
		tracer.WithTrace(ctx).Errorf("[DEBUG] VALIDATION_CRITICAL: File path contains .shenma_sync: '%s'", filePath)
		tracer.WithTrace(ctx).Errorf("[DEBUG] VALIDATION_CRITICAL: This indicates validation is looking for .shenma_sync files on disk")
		tracer.WithTrace(ctx).Errorf("[DEBUG] VALIDATION_CRITICAL: But these files might only exist in memory")
	}

	// 检查文件是否存在
	tracer.WithTrace(ctx).Infof("[DEBUG] VALIDATION_DETAIL: About to check file existence for: '%s'", filePath)
	exists, err := v.fileChecker.CheckFileExists(ctx, filePath)

	if err != nil {
		tracer.WithTrace(ctx).Errorf("[DEBUG] VALIDATION_ERROR: File existence check failed for '%s': %v", filePath, err)
		detail.Status = types.FileStatusMissing
		detail.Actual = "missing"
		detail.Error = err.Error()
		return detail, err
	}

	tracer.WithTrace(ctx).Infof("[DEBUG] VALIDATION_DETAIL: File existence check result for '%s': exists=%v", filePath, exists)

	if !exists {
		tracer.WithTrace(ctx).Errorf("[DEBUG] VALIDATION_CRITICAL: File not found on disk: '%s'", filePath)
		tracer.WithTrace(ctx).Errorf("[DEBUG] VALIDATION_CRITICAL: This confirms the mismatch between memory storage and disk validation")
		detail.Status = types.FileStatusMissing
		detail.Actual = "missing"
		detail.Error = "file not found"
		return detail, nil
	}

	tracer.WithTrace(ctx).Infof("[DEBUG] VALIDATION_DETAIL: File successfully found on disk: '%s'", filePath)

	// 如果配置了内容检查，则进行内容匹配验证
	if v.config.CheckContent {
		// 在MVP版本中，我们只检查文件是否存在
		// 实际内容匹配检查可以在后续版本中实现
		tracer.WithTrace(ctx).Debugf("content check not implemented in MVP, skipping for file: %s", filePath)
	}

	// 文件存在且状态匹配
	detail.Status = types.FileStatusMatched
	detail.Actual = expectedStatus

	return detail, nil
}

// calculateFinalStatus 计算最终验证状态
func (v *FileValidatorImpl) calculateFinalStatus(result *types.ValidationResult) {
	if result.SkippedFiles == result.TotalFiles {
		result.Status = types.ValidationStatusSkipped
		return
	}

	if result.MismatchedFiles == 0 && result.MatchedFiles > 0 {
		result.Status = types.ValidationStatusSuccess
		return
	}

	if result.MismatchedFiles > 0 {
		if v.config.FailOnMismatch {
			result.Status = types.ValidationStatusFailed
		} else {
			result.Status = types.ValidationStatusPartial
		}
		return
	}

	// 默认状态
	result.Status = types.ValidationStatusPartial
}
