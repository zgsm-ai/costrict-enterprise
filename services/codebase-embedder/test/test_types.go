package test

import (
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// TestConfig 测试配置结构
// 包含测试运行所需的所有配置信息，支持嵌入模型和向量数据库的对比测试
type TestConfig struct {
	Test struct {
		Name        string `yaml:"name"`        // 测试名称
		Description string `yaml:"description"` // 测试描述
	} `yaml:"test"`

	Queries struct {
		FilePath string `yaml:"file_path"` // 查询配置文件路径，支持JSON和YAML格式
	} `yaml:"queries"`

	Runtime struct {
		LogLevel      string        `yaml:"log_level"`      // 日志级别：debug, info, warn, error
		Parallel      int           `yaml:"parallel"`       // 并发测试数量
		Timeout       time.Duration `yaml:"timeout"`        // 单个测试超时时间
		Retry         int           `yaml:"retry"`          // 失败测试重试次数
		MetricsFormat []string      `yaml:"metrics_format"` // 指标输出格式：json, csv, console
	} `yaml:"runtime"`

	Embedders    []EmbedderConf    `yaml:"embedders"`     // 嵌入模型配置列表
	Rerank       RerankerConf      `yaml:"rerank"`        // 重排序模型配置
	VectorStores []VectorStoreConf `yaml:"vector_stores"` // 向量数据库配置列表
	Scenarios    struct {
		EmbedderComparison    ScenarioConfig `yaml:"embedder_comparison"`     // 嵌入模型对比测试配置
		VectorStoreComparison ScenarioConfig `yaml:"vector_store_comparison"` // 向量数据库对比测试配置
	} `yaml:"scenarios"`

	Output struct {
		Format             []string `yaml:"format"`               // 输出格式列表
		Directory          string   `yaml:"directory"`            // 输出目录
		IncludeDetails     bool     `yaml:"include_details"`      // 是否包含详细查询结果
		SeparateByScenario bool     `yaml:"separate_by_scenario"` // 按场景类型分别输出结果
		MetricsOnly        bool     `yaml:"metrics_only"`         // 是否仅输出指标（不含原始数据）
	} `yaml:"output"`
}

// ScenarioConfig 场景配置
// 定义测试场景的具体配置参数，包括测试名称、描述和相关组件配置
type ScenarioConfig struct {
	Name         string   `yaml:"name"`          // 场景名称
	Description  string   `yaml:"description"`   // 场景描述
	VectorStore  string   `yaml:"vector_store"`  // 使用的向量数据库配置名称
	Embedders    []string `yaml:"embedders"`     // 测试的嵌入模型列表
	VectorStores []string `yaml:"vector_stores"` // 测试的向量数据库列表
	TopK         int      `yaml:"top_k"`         // 检索时返回的top-k结果数量
}

// QueryConfig 查询配置
// 定义单个测试查询的配置信息，包括查询文本、期望结果和元数据
type QueryConfig struct {
	ID                string   `yaml:"id"`                 // 查询唯一标识符
	Text              string   `yaml:"text"`               // 查询文本内容
	Expected          []string `yaml:"expected_files"`     // 期望匹配的文件名列表
	ExpectedContents  []string `yaml:"expected_contents"`  // 期望查询到的内容片段
	ScenarioDimension string   `yaml:"scenario_dimension"` // 场景维度（关键词、同义词、近义词、中英文）
	Language          string   `yaml:"language"`           // 查询语言类型（go, python, javascript, markdown等）
	ScenarioType      string   `yaml:"scenario_type"`      // 场景类型（code或doc）
}

// TestResult 测试结果
// 包含整个测试的汇总信息，包括测试名称、时间戳和各场景的测试结果
type TestResult struct {
	TestName        string                    `json:"test_name"`        // 测试名称
	TestDescription string                    `json:"test_description"` // 测试描述
	StartTime       time.Time                 `json:"start_time"`       // 测试开始时间
	EndTime         time.Time                 `json:"end_time"`         // 测试结束时间
	Results         map[string]ScenarioResult `json:"results"`          // 各场景的测试结果映射
}

// ScenarioResult 场景测试结果
// 包含单个测试场景的详细结果，包括查询结果和平均指标
type ScenarioResult struct {
	StartTime      time.Time     `json:"start_time"`      // 场景测试开始时间
	EndTime        time.Time     `json:"end_time"`        // 场景测试结束时间
	QueryResults   []QueryResult `json:"query_results"`   // 所有查询的详细结果
	AverageMetrics Metrics       `json:"average_metrics"` // 平均评估指标
	Error          error         `json:"error,omitempty"` // 错误信息（如果有）
}

// QueryResult 查询结果
// 包含单个查询的测试结果，包括检索到的文件和评估指标
type QueryResult struct {
	QueryID           string                    `json:"query_id"`           // 查询唯一标识符
	Query             string                    `json:"query"`              // 查询文本内容
	Metrics           Metrics                   `json:"metrics"`            // 评估指标
	Retrieved         []*types.SemanticFileItem `json:"retrieved"`          // 检索到的文件列表
	Expected          []string                  `json:"expected"`           // 期望匹配的文件列表
	ExpectedContents  []string                  `json:"expected_contents"`  // 期望查询到的内容片段
	ScenarioDimension string                    `json:"scenario_dimension"` // 场景维度
	ContentMatches    []ContentMatch            `json:"content_matches"`    // 内容片段匹配结果
}

// ContentMatch 内容片段匹配结果
// 包含期望内容片段的匹配信息
type ContentMatch struct {
	Content        string          `json:"content"`         // 期望内容片段
	Matched        bool            `json:"matched"`         // 是否匹配
	MatchScore     float64         `json:"match_score"`     // 匹配分数
	FoundInFiles   []string        `json:"found_in_files"`  // 在哪些文件中找到
	MatchPositions []MatchPosition `json:"match_positions"` // 匹配位置信息
}

// MatchPosition 匹配位置信息
type MatchPosition struct {
	FilePath string `json:"file_path"` // 文件路径
	Line     int    `json:"line"`      // 行号
	Column   int    `json:"column"`    // 列号
	Context  string `json:"context"`   // 上下文内容
}

// Metrics 评估指标
// 包含检索系统的核心性能指标，用于评估检索效果和性能
type Metrics struct {
	Precision    float64 `json:"precision"`     // 准确率：检索到的相关文件比例
	Recall       float64 `json:"recall"`        // 召回率：相关文件被检索到的比例
	F1Score      float64 `json:"f1_score"`      // F1分数：准确率和召回率的调和平均
	ResponseTime float64 `json:"response_time"` // 响应时间：查询的完整响应时间（毫秒）
}

// VectorStoreConf 向量数据库配置
// 定义向量数据库的连接参数和操作配置
type VectorStoreConf struct {
	Name         string        `yaml:"name"`          // 向量数据库配置名称
	Type         string        `yaml:"type"`          // 向量数据库类型（如weaviate）
	Endpoint     string        `yaml:"endpoint"`      // 数据库服务端点URL
	BatchSize    int           `yaml:"batch_size"`    // 批量操作大小
	Timeout      time.Duration `yaml:"timeout"`       // 操作超时时间
	MaxRetries   int           `yaml:"max_retries"`   // 最大重试次数
	ClassName    string        `yaml:"class_name"`    // 向量数据库类名
	MaxDocuments int           `yaml:"max_documents"` // 最大文档数量限制
}

// EmbedderConf 嵌入模型配置
// 定义嵌入模型的API配置和操作参数
type EmbedderConf struct {
	// 通用配置
	Timeout       time.Duration `yaml:"timeout"`         // API请求超时时间
	MaxRetries    int           `yaml:"max_retries"`     // 最大重试次数
	BatchSize     int           `yaml:"batch_size"`      // 批量处理大小
	Model         string        `yaml:"model"`           // 模型名称（如gte-modernbert-base）
	APIKey        string        `yaml:"api_key"`         // API密钥
	APIBase       string        `yaml:"api_base"`        // API基础URL
	StripNewLines bool          `yaml:"strip_new_lines"` // 是否移除换行符
}

// RerankerConf 重排序模型配置
// 定义重排序模型的API配置和操作参数
type RerankerConf struct {
	Timeout    time.Duration `yaml:"timeout"`     // API请求超时时间
	MaxRetries int           `yaml:"max_retries"` // 最大重试次数
	Model      string        `yaml:"model"`       // 模型名称（如gte-reranker-modernbert-base）
	APIKey     string        `yaml:"api_key"`     // API密钥
	APIBase    string        `yaml:"api_base"`    // API基础URL
}
