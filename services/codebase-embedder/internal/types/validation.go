package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// MetadataValue 表示元数据值的自定义类型
type MetadataValue struct {
	StringValue  string
	NumberValue  float64
	BoolValue    bool
	StringValues []string
	NumberValues []float64
	IsArray      bool
}

// ValidationStatus 验证状态
type ValidationStatus string

const (
	ValidationStatusSuccess ValidationStatus = "success"
	ValidationStatusFailed  ValidationStatus = "failed"
	ValidationStatusPartial ValidationStatus = "partial"
	ValidationStatusSkipped ValidationStatus = "skipped"
)

// FileStatus 文件状态
type FileStatus string

const (
	FileStatusMatched    FileStatus = "matched"
	FileStatusMismatched FileStatus = "mismatched"
	FileStatusMissing    FileStatus = "missing"
	FileStatusSkipped    FileStatus = "skipped"
)

// ValidationResult 验证结果
type ValidationResult struct {
	TotalFiles      int                `json:"total_files"`
	MatchedFiles    int                `json:"matched_files"`
	MismatchedFiles int                `json:"mismatched_files"`
	SkippedFiles    int                `json:"skipped_files"`
	Details         []ValidationDetail `json:"details"`
	Status          ValidationStatus   `json:"status"`
	Timestamp       time.Time          `json:"timestamp"`
}

// ValidationDetail 单个文件验证详情
type ValidationDetail struct {
	FilePath string     `json:"file_path"`
	Status   FileStatus `json:"status"`
	Expected string     `json:"expected"` // 元数据中的状态
	Actual   string     `json:"actual"`   // 实际状态
	Error    string     `json:"error,omitempty"`
}

// SyncMetadata 同步元数据结构
type SyncMetadata struct {
	ClientId      string                   `json:"clientId"`
	CodebasePath  string                   `json:"codebasePath"`
	CodebaseName  string                   `json:"codebaseName"`
	ExtraMetadata map[string]MetadataValue `json:"extraMetadata"`
	FileList      map[string]string        `json:"fileList"`                // 文件路径 -> 状态（兼容格式一）
	FileListItems []FileListItem           `json:"fileListItems,omitempty"` // 文件列表项（格式二）
	Timestamp     int64                    `json:"timestamp"`
}

// FileListItem 文件列表项（数组格式）
type FileListItem struct {
	Path       string `json:"path"`       // 源文件路径
	TargetPath string `json:"targetPath"` // 目标文件路径（用于rename操作）
	Hash       string `json:"hash"`       // 文件哈希值
	Status     string `json:"status"`     // 操作类型：add/modify/delete/rename
	Operate    string `json:"operate"`    // 操作类型（备用字段）
	RequestId  string `json:"requestId"`  // 请求ID
}

// FileStats 文件统计信息
type FileStats struct {
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

// ValidationParams 验证参数
type ValidationParams struct {
	MetadataPath string            `json:"metadata_path"` // 元数据文件路径
	ExtractPath  string            `json:"extract_path"`  // 解压文件路径
	SkipPatterns []string          `json:"skip_patterns"` // 跳过文件模式
	Config       *ValidationConfig `json:"config"`        // 验证配置
}

// NewStringMetadataValue 创建字符串类型的元数据值
func NewStringMetadataValue(value string) MetadataValue {
	return MetadataValue{
		StringValue: value,
	}
}

// NewNumberMetadataValue 创建数字类型的元数据值
func NewNumberMetadataValue(value float64) MetadataValue {
	return MetadataValue{
		NumberValue: value,
	}
}

// NewBoolMetadataValue 创建布尔类型的元数据值
func NewBoolMetadataValue(value bool) MetadataValue {
	return MetadataValue{
		BoolValue: value,
	}
}

// NewStringArrayMetadataValue 创建字符串数组类型的元数据值
func NewStringArrayMetadataValue(values []string) MetadataValue {
	return MetadataValue{
		StringValues: values,
		IsArray:      true,
	}
}

// NewNumberArrayMetadataValue 创建数字数组类型的元数据值
func NewNumberArrayMetadataValue(values []float64) MetadataValue {
	return MetadataValue{
		NumberValues: values,
		IsArray:      true,
	}
}

// MarshalJSON 实现MetadataValue的JSON序列化
func (mv MetadataValue) MarshalJSON() ([]byte, error) {
	if mv.IsArray {
		if len(mv.StringValues) > 0 {
			return json.Marshal(mv.StringValues)
		} else if len(mv.NumberValues) > 0 {
			return json.Marshal(mv.NumberValues)
		}
		return json.Marshal([]interface{}{})
	} else {
		if mv.StringValue != "" {
			return json.Marshal(mv.StringValue)
		} else if mv.NumberValue != 0 {
			return json.Marshal(mv.NumberValue)
		} else {
			return json.Marshal(mv.BoolValue)
		}
	}
}

// UnmarshalJSON 实现MetadataValue的JSON反序列化
func (mv *MetadataValue) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		mv.StringValue = strVal
		mv.IsArray = false
		return nil
	}

	// 尝试解析为数字
	var numVal float64
	if err := json.Unmarshal(data, &numVal); err == nil {
		mv.NumberValue = numVal
		mv.IsArray = false
		return nil
	}

	// 尝试解析为布尔值
	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		mv.BoolValue = boolVal
		mv.IsArray = false
		return nil
	}

	// 尝试解析为字符串数组
	var strSlice []string
	if err := json.Unmarshal(data, &strSlice); err == nil {
		mv.StringValues = strSlice
		mv.IsArray = true
		return nil
	}

	// 尝试解析为数字数组
	var numSlice []float64
	if err := json.Unmarshal(data, &numSlice); err == nil {
		mv.NumberValues = numSlice
		mv.IsArray = true
		return nil
	}

	return fmt.Errorf("无法解析为MetadataValue: %s", string(data))
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	CheckContent   bool     `json:"check_content"`    // 是否检查文件内容
	FailOnMismatch bool     `json:"fail_on_mismatch"` // 不匹配时是否失败
	LogLevel       string   `json:"log_level"`        // 日志级别
	MaxConcurrency int      `json:"max_concurrency"`  // 最大并发数
	Enabled        bool     `json:"enabled"`          // 是否启用文件验证
	SkipPatterns   []string `json:"skip_patterns"`    // 跳过文件模式
}
