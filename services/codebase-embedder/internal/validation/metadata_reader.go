package validation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

var (
	ErrMetadataNotFound = errors.New("metadata file not found")
	ErrInvalidMetadata  = errors.New("invalid metadata format")
	ErrInvalidFileList  = errors.New("invalid file list in metadata")
	ErrNoTimestampFiles = errors.New("no timestamp files found in .shenma_sync directory")
)

// SyncMetadataReaderImpl 同步元数据读取器实现
type SyncMetadataReaderImpl struct {
	memoryFiles map[string][]byte // 内存中的文件缓存
}

// NewSyncMetadataReader 创建同步元数据读取器
func NewSyncMetadataReader() SyncMetadataReader {
	return &SyncMetadataReaderImpl{
		memoryFiles: make(map[string][]byte),
	}
}

// NewSyncMetadataReaderWithMemory 创建带内存缓存的同步元数据读取器
func NewSyncMetadataReaderWithMemory(files map[string][]byte) SyncMetadataReader {
	return &SyncMetadataReaderImpl{
		memoryFiles: files,
	}
}

// ReadMetadata 读取元数据文件
func (r *SyncMetadataReaderImpl) ReadMetadata(ctx context.Context, path string) (*types.SyncMetadata, error) {
	tracer.WithTrace(ctx).Infof("reading metadata from path: %s", path)

	// 添加诊断日志：记录输入参数
	tracer.WithTrace(ctx).Infof("[DEBUG] ReadMetadata called with path: '%s'", path)
	if absPath, err := filepath.Abs(path); err != nil {
		tracer.WithTrace(ctx).Errorf("[DEBUG] Failed to resolve absolute path for '%s': %v", path, err)
	} else {
		tracer.WithTrace(ctx).Infof("[DEBUG] Absolute path resolved to: '%s'", absPath)
	}

	// 检查路径是否以.shenma_sync结尾，如果不是则修正
	if !strings.HasSuffix(filepath.Clean(path), ".shenma_sync") {
		// 如果是目录路径，添加.shenma_sync
		cleanPath := filepath.Clean(path)
		if fi, err := os.Stat(cleanPath); err == nil && fi.IsDir() {
			correctPath := filepath.Join(cleanPath, ".shenma_sync")
			tracer.WithTrace(ctx).Infof("[DEBUG] Path is directory, correcting from '%s' to '%s'", path, correctPath)
			path = correctPath
		}
	}

	// 添加关键诊断日志：检查内存中的文件
	tracer.WithTrace(ctx).Infof("[DEBUG] CRITICAL DIAGNOSIS: This is a hybrid metadata reader (memory + filesystem)")
	tracer.WithTrace(ctx).Infof("[DEBUG] CRITICAL DIAGNOSIS: Memory files cache contains %d files", len(r.memoryFiles))
	for memPath := range r.memoryFiles {
		tracer.WithTrace(ctx).Infof("[DEBUG] CRITICAL DIAGNOSIS: Memory file available: '%s'", memPath)
	}

	// 首先尝试从内存中读取文件
	var data []byte
	var err error
	var foundInMemory bool

	// 检查是否是.shenma_sync文件或目录
	if strings.Contains(path, ".shenma_sync") {
		tracer.WithTrace(ctx).Infof("[DEBUG] Path contains .shenma_sync, checking memory first")

		// 如果是目录，查找内存中最新的时间戳文件
		if fi, err := os.Stat(path); err == nil && fi.IsDir() {
			tracer.WithTrace(ctx).Infof("[DEBUG] Path is a directory, looking for .shenma_sync files in memory")

			// 在内存中查找.shenma_sync文件
			var memoryTimestampFiles []string
			for memPath := range r.memoryFiles {
				if strings.HasPrefix(memPath, ".shenma_sync/") && !strings.HasSuffix(memPath, "/") {
					memoryTimestampFiles = append(memoryTimestampFiles, memPath)
					tracer.WithTrace(ctx).Infof("[DEBUG] Found .shenma_sync file in memory: '%s'", memPath)
				}
			}

			if len(memoryTimestampFiles) > 0 {
				// 使用第一个找到的文件（在实际应用中可能需要按时间戳排序）
				selectedFile := memoryTimestampFiles[0]
				tracer.WithTrace(ctx).Infof("[DEBUG] Selected memory file: '%s'", selectedFile)
				data = r.memoryFiles[selectedFile]
				foundInMemory = true
				path = selectedFile // 更新路径为内存中的文件路径
			}
		} else {
			// 直接查找内存中的文件
			for memPath, memData := range r.memoryFiles {
				if memPath == path || filepath.Base(memPath) == filepath.Base(path) {
					tracer.WithTrace(ctx).Infof("[DEBUG] Found direct match in memory: '%s'", memPath)
					data = memData
					foundInMemory = true
					path = memPath // 更新路径为内存中的文件路径
					break
				}
			}
		}
	}

	if foundInMemory {
		tracer.WithTrace(ctx).Infof("[DEBUG] Successfully read %d bytes from memory", len(data))
	} else {
		tracer.WithTrace(ctx).Infof("[DEBUG] File not found in memory, trying filesystem")

		// 检查路径是否存在
		if fi, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				tracer.WithTrace(ctx).Errorf("[DEBUG] Path does not exist: '%s'", path)
			} else {
				tracer.WithTrace(ctx).Errorf("[DEBUG] Error checking path existence: '%s', error: %v", path, err)
			}
		} else {
			tracer.WithTrace(ctx).Infof("[DEBUG] Path exists: '%s', is directory: %t, size: %d bytes", path, fi.IsDir(), fi.Size())
		}

		// 如果path是目录，则查找最新的时间戳文件
		if fi, err := os.Stat(path); err == nil && fi.IsDir() {
			tracer.WithTrace(ctx).Infof("[DEBUG] Path is a directory, listing contents...")

			// 列出目录内容
			entries, err := os.ReadDir(path)
			if err != nil {
				tracer.WithTrace(ctx).Errorf("[DEBUG] Failed to read directory '%s': %v", path, err)
			} else {
				tracer.WithTrace(ctx).Infof("[DEBUG] Directory '%s' contains %d entries:", path, len(entries))
				for _, entry := range entries {
					info, _ := entry.Info()
					tracer.WithTrace(ctx).Infof("[DEBUG]  - %s (dir: %t, size: %d, modTime: %v)",
						entry.Name(), entry.IsDir(), info.Size(), info.ModTime())
				}
			}

			timestampFile, err := r.findLatestTimestampFile(path)
			if err != nil {
				tracer.WithTrace(ctx).Errorf("[DEBUG] Failed to find latest timestamp file in directory '%s': %v", path, err)
				return nil, err
			}
			path = filepath.Join(path, timestampFile)
			tracer.WithTrace(ctx).Infof("[DEBUG] Found latest timestamp file, new path: '%s'", path)
			tracer.WithTrace(ctx).Infof("found latest timestamp file: %s", timestampFile)
		}

		// 检查文件是否存在
		tracer.WithTrace(ctx).Infof("[DEBUG] Checking if metadata file exists: '%s'", path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			tracer.WithTrace(ctx).Errorf("[DEBUG] Metadata file NOT found: '%s'", path)
			tracer.WithTrace(ctx).Errorf("[DEBUG] Working directory: '%s'", getWorkingDirectory())
			tracer.WithTrace(ctx).Errorf("[DEBUG] File info check failed with error: %v", err)
			return nil, fmt.Errorf("%w: %s", ErrMetadataNotFound, path)
		}

		// 获取文件信息用于日志
		if fi, err := os.Stat(path); err == nil {
			tracer.WithTrace(ctx).Infof("[DEBUG] Metadata file found: '%s', size: %d bytes, modTime: %v",
				path, fi.Size(), fi.ModTime())
		}

		// 读取文件内容
		tracer.WithTrace(ctx).Infof("[DEBUG] Reading file content from: '%s'", path)
		data, err = os.ReadFile(path)
		if err != nil {
			tracer.WithTrace(ctx).Errorf("[DEBUG] Failed to read file content from '%s': %v", path, err)
			return nil, fmt.Errorf("failed to read metadata file: %w", err)
		}
		tracer.WithTrace(ctx).Infof("[DEBUG] Successfully read %d bytes from file", len(data))
	}

	// 解析JSON
	var metadata types.SyncMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		tracer.WithTrace(ctx).Errorf("[DEBUG] Failed to parse JSON metadata: %v", err)
		tracer.WithTrace(ctx).Errorf("[DEBUG] Raw data (first 200 chars): %s", string(data)[:min(200, len(data))])
		return nil, fmt.Errorf("%w: %v", ErrInvalidMetadata, err)
	}

	tracer.WithTrace(ctx).Infof("[DEBUG] Successfully parsed metadata, ClientId: '%s', CodebasePath: '%s', CodebaseName: '%s'",
		metadata.ClientId, metadata.CodebasePath, metadata.CodebaseName)
	tracer.WithTrace(ctx).Infof("successfully read metadata, files count: %d", len(metadata.FileList))
	return &metadata, nil
}

// 辅助函数：获取工作目录
func getWorkingDirectory() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "unknown"
}

// 辅助函数：返回两个整数中的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidateMetadata 验证元数据格式
func (r *SyncMetadataReaderImpl) ValidateMetadata(metadata *types.SyncMetadata) error {
	if metadata == nil {
		return ErrInvalidMetadata
	}

	// 验证必要字段
	if metadata.ClientId == "" {
		return fmt.Errorf("%w: client_id is required", ErrInvalidMetadata)
	}

	if metadata.CodebasePath == "" {
		return fmt.Errorf("%w: codebase_path is required", ErrInvalidMetadata)
	}

	if metadata.CodebaseName == "" {
		return fmt.Errorf("%w: codebase_name is required", ErrInvalidMetadata)
	}

	if metadata.FileList == nil {
		return fmt.Errorf("%w: file_list is required", ErrInvalidMetadata)
	}

	// 验证文件列表
	for filePath, status := range metadata.FileList {
		if filePath == "" {
			return fmt.Errorf("%w: empty file path in file list", ErrInvalidFileList)
		}
		if status == "" {
			return fmt.Errorf("%w: empty status for file: %s", ErrInvalidFileList, filePath)
		}
		// 只允许 add 和 modify 状态
		if status != "add" && status != "modify" {
			return fmt.Errorf("%w: invalid status '%s' for file: %s, only 'add' and 'modify' are allowed",
				ErrInvalidFileList, status, filePath)
		}
	}

	return nil
}

// GetMetadataPath 获取元数据文件路径
func (r *SyncMetadataReaderImpl) GetMetadataPath(extractPath string) string {
	return filepath.Join(extractPath, ".shenma_sync")
}

// findLatestTimestampFile 查找目录下的最新文件
func (r *SyncMetadataReaderImpl) findLatestTimestampFile(dirPath string) (string, error) {
	// 读取目录中的所有文件
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var files []struct {
		name    string
		modTime int64
	}

	// 遍历目录中的文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			// 如果获取文件信息失败，跳过这个文件
			continue
		}

		files = append(files, struct {
			name    string
			modTime int64
		}{
			name:    entry.Name(),
			modTime: info.ModTime().Unix(),
		})
	}

	// 如果没有找到文件，返回错误
	if len(files) == 0 {
		return "", ErrNoTimestampFiles
	}

	// 按修改时间降序排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	// 返回最新修改的文件名
	return files[0].name, nil
}
