package vector

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// CodebaseQueryStore 向量数据库查询存储
type CodebaseQueryStore struct {
	store Store
	logx.Logger
}

// NewCodebaseQueryStore 创建新的向量查询存储
func NewCodebaseQueryStore(store Store, logger logx.Logger) *CodebaseQueryStore {
	return &CodebaseQueryStore{
		store:  store,
		Logger: logger,
	}
}

// QueryCodebaseStats 查询代码库统计信息
func (s *CodebaseQueryStore) QueryCodebaseStats(ctx context.Context, clientId string, codebasePath string) (*types.CodebaseSummary, error) {
	// 使用vector.Store接口的GetIndexSummary方法获取统计信息
	embeddingSummary, err := s.store.GetIndexSummary(ctx, clientId, codebasePath)
	if err != nil {
		s.Errorf("查询代码库统计信息失败, clientId: %s, codebasePath: %s, error: %v", clientId, codebasePath, err)
		return nil, fmt.Errorf("查询代码库统计信息失败: %w", err)
	}

	// 检查响应是否为空
	if embeddingSummary == nil {
		return &types.CodebaseSummary{
			TotalFiles:     0,
			TotalChunks:    0,
			LastUpdateTime: time.Now(),
			IndexStatus:    "not_found",
			IndexProgress:  0,
		}, nil
	}

	// 转换时间格式
	lastUpdateTime, err := time.Parse(time.RFC3339, embeddingSummary.UpdatedAt)
	if err != nil {
		// 如果解析失败，使用当前时间
		lastUpdateTime = time.Now()
	}

	// 构造CodebaseSummary
	summary := &types.CodebaseSummary{
		TotalFiles:     int32(embeddingSummary.TotalFiles),
		TotalChunks:    int32(embeddingSummary.TotalChunks),
		LastUpdateTime: lastUpdateTime,
		IndexStatus:    embeddingSummary.Status,
		IndexProgress:  0, // IndexProgress在EmbeddingSummary中未提供，暂时设为0
	}

	return summary, nil
}

// QueryLanguageDistribution 查询语言分布信息
func (s *CodebaseQueryStore) QueryLanguageDistribution(ctx context.Context, codebaseId int32) ([]types.LanguageDistribution, error) {
	// TODO: 实现语言分布查询逻辑
	// 当前版本返回空数组，后续需要根据vector.Store接口或其它方式实现
	s.Errorf("QueryLanguageDistribution 方法尚未实现，返回空的语言分布数据")
	return []types.LanguageDistribution{}, nil
}

// QueryRecentFiles 查询最近更新的文件
func (s *CodebaseQueryStore) QueryRecentFiles(ctx context.Context, codebaseId int32, limit int) ([]types.RecentFileInfo, error) {
	// TODO: 实现最近文件查询逻辑
	// 当前版本返回空数组，后续需要根据vector.Store接口或其它方式实现
	s.Errorf("QueryRecentFiles 方法尚未实现，返回空的最近文件数据")
	return []types.RecentFileInfo{}, nil
}

// QueryIndexStats 查询索引统计信息
func (s *CodebaseQueryStore) QueryIndexStats(ctx context.Context, codebaseId int32) (*types.IndexStatistics, error) {
	// TODO: 实现索引统计查询逻辑
	// 当前版本返回默认值，后续需要根据vector.Store接口或其它方式实现
	s.Errorf("QueryIndexStats 方法尚未实现，返回默认的索引统计数据")
	return &types.IndexStatistics{
		AverageChunkSize: 0,
		MaxChunkSize:     0,
		MinChunkSize:     0,
		TotalVectors:     0,
	}, nil
}

// QueryCodebaseRecords 查询代码库详细记录
func (s *CodebaseQueryStore) QueryCodebaseRecords(ctx context.Context, clientId string, codebasePath string) ([]types.CodebaseRecord, error) {
	records, err := s.store.GetCodebaseRecords(ctx, clientId, codebasePath)
	if err != nil {
		s.Errorf("查询代码库详细记录失败, clientId: %s, codebasePath: %s, error: %v",
			clientId, codebasePath, err)
		return nil, fmt.Errorf("查询代码库详细记录失败: %w", err)
	}

	// 转换类型
	result := make([]types.CodebaseRecord, len(records))
	for i, record := range records {
		result[i] = *record
	}

	return result, nil
}

// QueryDictionaryRecords 查询指定目录的详细记录，通过匹配filePath的前缀
func (s *CodebaseQueryStore) QueryDictionaryRecords(ctx context.Context, clientId string, codebasePath string, dictionary string) ([]types.CodebaseRecord, error) {
	records, err := s.store.GetDictionaryRecords(ctx, clientId, codebasePath, dictionary)
	if err != nil {
		s.Errorf("查询目录详细记录失败, clientId: %s, codebasePath: %s, dictionary: %s, error: %v",
			clientId, codebasePath, dictionary, err)
		return nil, fmt.Errorf("查询目录详细记录失败: %w", err)
	}

	// 转换类型
	result := make([]types.CodebaseRecord, len(records))
	for i, record := range records {
		result[i] = *record
	}

	return result, nil
}

// QueryFileRecords 查询指定文件的详细记录
func (s *CodebaseQueryStore) QueryFileRecords(ctx context.Context, clientId string, codebasePath string, filePath string) ([]types.CodebaseRecord, error) {
	records, err := s.store.GetFileRecords(ctx, clientId, codebasePath, filePath)
	if err != nil {
		s.Errorf("查询文件详细记录失败, clientId: %s, codebasePath: %s, filePath: %s, error: %v",
			clientId, codebasePath, filePath, err)
		return nil, fmt.Errorf("查询文件详细记录失败: %w", err)
	}

	// 转换类型
	result := make([]types.CodebaseRecord, len(records))
	for i, record := range records {
		result[i] = *record
	}

	return result, nil
}
