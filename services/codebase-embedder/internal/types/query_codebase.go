package types

import "time"

// CodebaseQueryRequest 查询代码库请求参数
type CodebaseQueryRequest struct {
	ClientId     string `json:"clientId"`
	CodebasePath string `json:"codebasePath"`
	CodebaseName string `json:"codebaseName"`
}

// CodebaseQueryResponse 查询代码库响应
type CodebaseQueryResponse struct {
	CodebaseId   int32                  `json:"codebaseId"`
	CodebaseName string                 `json:"codebaseName"`
	CodebasePath string                 `json:"codebasePath"`
	Summary      *CodebaseSummary       `json:"summary"`
	LanguageDist []LanguageDistribution `json:"languageDistribution"`
	RecentFiles  []RecentFileInfo       `json:"recentFiles"`
	IndexStats   *IndexStatistics       `json:"indexStats"`
	Records      []CodebaseRecord       `json:"records"` // 新增：详细记录列表
}

// CodebaseRecord 代码库详细记录
type CodebaseRecord struct {
	Id           string    `json:"id"`
	FilePath     string    `json:"filePath"`
	Language     string    `json:"language"`
	Content      string    `json:"content"`
	Range        []int     `json:"range"` // [startLine, startColumn, endLine, endColumn]
	TokenCount   int       `json:"tokenCount"`
	LastUpdated  time.Time `json:"lastUpdated"`
	CodebaseId   int32     `json:"codebaseId"`
	CodebasePath string    `json:"codebasePath"`
	CodebaseName string    `json:"codebaseName"`
	SyncId       int32     `json:"syncId"`
}

// CodebaseSummary 代码库摘要信息
type CodebaseSummary struct {
	TotalFiles     int32     `json:"totalFiles"`
	TotalChunks    int32     `json:"totalChunks"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
	IndexStatus    string    `json:"indexStatus"`
	IndexProgress  int32     `json:"indexProgress"`
}

// LanguageDistribution 语言分布信息
type LanguageDistribution struct {
	Language   string  `json:"language"`
	FileCount  int32   `json:"fileCount"`
	ChunkCount int32   `json:"chunkCount"`
	Percentage float64 `json:"percentage"`
}

// RecentFileInfo 最近文件信息
type RecentFileInfo struct {
	FilePath    string    `json:"filePath"`
	LastIndexed time.Time `json:"lastIndexed"`
	ChunkCount  int32     `json:"chunkCount"`
	FileSize    int64     `json:"fileSize"`
}

// IndexStatistics 索引统计信息
type IndexStatistics struct {
	AverageChunkSize int32 `json:"averageChunkSize"`
	MaxChunkSize     int32 `json:"maxChunkSize"`
	MinChunkSize     int32 `json:"minChunkSize"`
	TotalVectors     int32 `json:"totalVectors"`
}
