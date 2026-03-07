package validation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

var (
	ErrFileAccessFailed = errors.New("file access failed")
	ErrPathTraversal    = errors.New("potential path traversal attack")
)

// FileCheckerImpl 文件检查器实现
type FileCheckerImpl struct {
	skipPatterns []string
}

// NewFileChecker 创建文件检查器
func NewFileChecker(skipPatterns []string) FileChecker {
	return &FileCheckerImpl{
		skipPatterns: skipPatterns,
	}
}

// CheckFileExists 检查文件是否存在
func (c *FileCheckerImpl) CheckFileExists(ctx context.Context, filePath string) (bool, error) {
	tracer.WithTrace(ctx).Debugf("checking file exists: %s", filePath)

	// 检查路径遍历攻击
	if err := c.validatePath(filePath); err != nil {
		return false, err
	}

	// 检查是否应该跳过该文件
	if c.shouldSkipFile(filePath) {
		tracer.WithTrace(ctx).Debugf("skipping file: %s", filePath)
		return false, nil
	}

	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %v", ErrFileAccessFailed, err)
	}

	return true, nil
}

// CheckFileMatch 检查文件是否匹配
func (c *FileCheckerImpl) CheckFileMatch(ctx context.Context, expectedPath, actualPath string) (bool, error) {
	tracer.WithTrace(ctx).Debugf("checking file match: expected=%s, actual=%s", expectedPath, actualPath)

	// 检查路径遍历攻击
	if err := c.validatePath(expectedPath); err != nil {
		return false, err
	}
	if err := c.validatePath(actualPath); err != nil {
		return false, err
	}

	// 获取文件信息进行对比
	expectedInfo, err := os.Stat(expectedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %v", ErrFileAccessFailed, err)
	}

	actualInfo, err := os.Stat(actualPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %v", ErrFileAccessFailed, err)
	}

	// 比较文件大小
	if expectedInfo.Size() != actualInfo.Size() {
		tracer.WithTrace(ctx).Debugf("file size mismatch: expected=%d, actual=%d",
			expectedInfo.Size(), actualInfo.Size())
		return false, nil
	}

	// 比较修改时间（允许一定差异）
	timeDiff := expectedInfo.ModTime().Sub(actualInfo.ModTime())
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// 如果时间差异超过1秒，认为不匹配
	if timeDiff > time.Second {
		tracer.WithTrace(ctx).Debugf("file mod time mismatch: expected=%v, actual=%v",
			expectedInfo.ModTime(), actualInfo.ModTime())
		return false, nil
	}

	return true, nil
}

// GetFileStats 获取文件统计信息
func (c *FileCheckerImpl) GetFileStats(ctx context.Context, filePath string) (*types.FileStats, error) {
	tracer.WithTrace(ctx).Debugf("getting file stats: %s", filePath)

	// 检查路径遍历攻击
	if err := c.validatePath(filePath); err != nil {
		return nil, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFileAccessFailed, err)
	}

	return &types.FileStats{
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// validatePath 验证路径安全性，防止路径遍历攻击
func (c *FileCheckerImpl) validatePath(filePath string) error {
	// 清理路径
	cleanPath := filepath.Clean(filePath)

	// 检查是否包含路径遍历序列
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("%w: %s", ErrPathTraversal, filePath)
	}

	// 检查是否为绝对路径（根据安全策略可能需要限制）
	if filepath.IsAbs(cleanPath) {
		// 在MVP版本中，我们允许绝对路径，但记录警告
		// 在生产环境中可能需要更严格的限制
		tracer.WithTrace(context.Background()).Errorf("absolute path detected: %s", filePath)
	}

	return nil
}

// shouldSkipFile 检查是否应该跳过该文件
func (c *FileCheckerImpl) shouldSkipFile(filePath string) bool {
	if c.skipPatterns == nil {
		return false
	}

	for _, pattern := range c.skipPatterns {
		matched, err := filepath.Match(pattern, filePath)
		if err != nil {
			// 如果模式无效，记录错误但不跳过
			tracer.WithTrace(context.Background()).Errorf("invalid skip pattern: %s, error: %v", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

// SetSkipPatterns 设置跳过模式
func (c *FileCheckerImpl) SetSkipPatterns(patterns []string) {
	c.skipPatterns = patterns
}
