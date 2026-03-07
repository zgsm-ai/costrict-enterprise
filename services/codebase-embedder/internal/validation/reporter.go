package validation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// ValidationReporterImpl 验证结果报告器实现
type ValidationReporterImpl struct {
	logLevel string
}

// NewValidationReporter 创建验证结果报告器
func NewValidationReporter(logLevel string) ValidationReporter {
	return &ValidationReporterImpl{
		logLevel: logLevel,
	}
}

// Report 报告验证结果
func (r *ValidationReporterImpl) Report(ctx context.Context, result *types.ValidationResult) error {
	tracer.WithTrace(ctx).Infof("validation report: total=%d, matched=%d, mismatched=%d, skipped=%d, status=%s",
		result.TotalFiles, result.MatchedFiles, result.MismatchedFiles, result.SkippedFiles, result.Status)

	// 记录详细信息
	if r.logLevel == "debug" || r.logLevel == "trace" {
		for _, detail := range result.Details {
			if detail.Status != types.FileStatusMatched {
				tracer.WithTrace(ctx).Infof("file validation detail: path=%s, status=%s, expected=%s, actual=%s, error=%s",
					detail.FilePath, detail.Status, detail.Expected, detail.Actual, detail.Error)
			}
		}
	}

	return nil
}

// Log 记录验证日志
func (r *ValidationReporterImpl) Log(ctx context.Context, result *types.ValidationResult) error {
	// 记录验证结果摘要
	tracer.WithTrace(ctx).Infof("file validation completed - Status: %s", result.Status)
	tracer.WithTrace(ctx).Infof("validation summary - Total: %d, Matched: %d, Mismatched: %d, Skipped: %d",
		result.TotalFiles, result.MatchedFiles, result.MismatchedFiles, result.SkippedFiles)

	// 如果有不匹配的文件，记录详细信息
	if result.MismatchedFiles > 0 {
		tracer.WithTrace(ctx).Errorf("found %d mismatched files", result.MismatchedFiles)
		for _, detail := range result.Details {
			if detail.Status == types.FileStatusMismatched {
				tracer.WithTrace(ctx).Errorf("mismatched file: %s (expected: %s, actual: %s)",
					detail.FilePath, detail.Expected, detail.Actual)
			}
		}
	}

	// 如果有缺失的文件，记录详细信息
	if result.TotalFiles-result.MatchedFiles-result.MismatchedFiles-result.SkippedFiles > 0 {
		missingCount := result.TotalFiles - result.MatchedFiles - result.MismatchedFiles - result.SkippedFiles
		tracer.WithTrace(ctx).Errorf("found %d missing files", missingCount)
		for _, detail := range result.Details {
			if detail.Status == types.FileStatusMissing {
				tracer.WithTrace(ctx).Errorf("missing file: %s (expected: %s)",
					detail.FilePath, detail.Expected)
			}
		}
	}

	// 记录JSON格式的结果（用于调试和分析）
	if r.logLevel == "debug" {
		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			tracer.WithTrace(ctx).Errorf("failed to marshal validation result: %v", err)
			return err
		}
		tracer.WithTrace(ctx).Debugf("validation result JSON: %s", string(jsonResult))
	}

	return nil
}

// GenerateSummary 生成验证摘要
func (r *ValidationReporterImpl) GenerateSummary(result *types.ValidationResult) string {
	var summary string

	switch result.Status {
	case types.ValidationStatusSuccess:
		summary = fmt.Sprintf("✅ Validation successful: All %d files matched", result.MatchedFiles)
	case types.ValidationStatusPartial:
		summary = fmt.Sprintf("⚠️ Validation partial: %d matched, %d mismatched, %d skipped out of %d total files",
			result.MatchedFiles, result.MismatchedFiles, result.SkippedFiles, result.TotalFiles)
	case types.ValidationStatusFailed:
		summary = fmt.Sprintf("❌ Validation failed: %d matched, %d mismatched out of %d total files",
			result.MatchedFiles, result.MismatchedFiles, result.TotalFiles)
	case types.ValidationStatusSkipped:
		summary = fmt.Sprintf("ℹ️ Validation skipped: %d files skipped", result.SkippedFiles)
	default:
		summary = fmt.Sprintf("❓ Validation unknown status: %s", result.Status)
	}

	return summary
}

// SetLogLevel 设置日志级别
func (r *ValidationReporterImpl) SetLogLevel(level string) {
	r.logLevel = level
}
