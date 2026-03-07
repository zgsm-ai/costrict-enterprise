package config

import "time"

type IndexTaskConf struct {
	// Topic             string
	ConsumerGroup     string `json:",default=codebase_embedder"` // 消费者组名称（用于Streams）
	PoolSize          int
	QueueSize         int
	LockTimeout       time.Duration `json:",default=300s"`
	EmbeddingTask     EmbeddingTaskConf
	GraphTask         GraphTaskConf
	FileValidation    FileValidationConf
	MsgMaxFailedTimes int `json:",default=3"`
}

type EmbeddingTaskConf struct {
	MaxConcurrency int  `json:",default=5"`
	Enabled        bool `json:",default=true"`
	Timeout        time.Duration
	// 滑动窗口重叠token数
	OverlapTokens         int
	MaxTokensPerChunk     int
	EnableMarkdownParsing bool `json:",default=false"` // 是否启用markdown文件解析
	EnableOpenAPIParsing  bool `json:",default=false"` // 是否启用OpenAPI文档解析
}

type GraphTaskConf struct {
	MaxConcurrency int  `json:",default=5"`
	Enabled        bool `json:",default=true"`
	Timeout        time.Duration
	ConfFile       string `json:",default=etc/codegraph.yaml"`
}

type FileValidationConf struct {
	Enabled        bool     `json:",default=true"`
	MaxConcurrency int      `json:",default=10"`
	FailOnMismatch bool     `json:",default=false"`
	CheckContent   bool     `json:",default=false"`
	SkipPatterns   []string `json:",default=[]"`
	LogLevel       string   `json:",default=\"info\""`
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
