package config

import "time"

// VectorStoreConf 向量数据库配置
type VectorStoreConf struct {
	Type string // 向量数据库类型
	// 通用配置
	Timeout    time.Duration // 操作超时时间
	MaxRetries int           // 最大重试次数
	Embedder   EmbedderConf
	Reranker   RerankerConf
	// 具体实现配置
	Weaviate        WeaviateConf // Weaviate配置
	FetchSourceCode bool         `json:",default=false"` // 是否获取源码
	StoreSourceCode bool         `json:",default=false"` // 是否存储源码
	BaseURL         string       `json:",optional"`      // 获取代码内容的基础URL
}

// WeaviateConf Weaviate向量数据库配置
type WeaviateConf struct {
	Endpoint     string        // HTTP端点
	APIKey       string        `json:",optional"`    // API密钥
	BatchSize    int           `json:",default=10"`  // 批处理大小
	Timeout      time.Duration `json:",default=10s"` // 超时时间
	ClassName    string
	MaxDocuments int `json:",default=10"`
}

// EmbedderConf 嵌入模型配置
type EmbedderConf struct {
	// 通用配置
	Timeout       time.Duration
	MaxRetries    int
	BatchSize     int
	Model         string // 模型名称（如text-embedding-ada-002）
	APIKey        string // API密钥
	APIBase       string // API基础URL
	StripNewLines bool
}

type RerankerConf struct {
	Timeout    time.Duration
	MaxRetries int
	Model      string // 模型名称（如text-embedding-ada-002）
	APIKey     string // API密钥
	APIBase    string // API基础URL
}
