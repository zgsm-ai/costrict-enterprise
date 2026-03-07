package vector

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/store/redis"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"

	"github.com/weaviate/weaviate/entities/vectorindex/dynamic"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	goweaviate "github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type weaviateWrapper struct {
	reranker      Reranker
	embedder      Embedder
	client        *goweaviate.Client
	className     string
	cfg           config.VectorStoreConf
	statusManager *redis.StatusManager
	requestId     string
}

func New(cfg config.VectorStoreConf, embedder Embedder, reranker Reranker) (Store, error) {
	var authConf auth.Config
	if cfg.Weaviate.APIKey != types.EmptyString {
		authConf = auth.ApiKey{Value: cfg.Weaviate.APIKey}
	}
	client, err := goweaviate.NewClient(goweaviate.Config{
		Host:       cfg.Weaviate.Endpoint,
		Scheme:     schemeHttp,
		AuthConfig: authConf,
		Timeout:    cfg.Weaviate.Timeout,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Weaviate client: %w", err)
	}

	store := &weaviateWrapper{
		client:    client,
		className: cfg.Weaviate.ClassName,
		embedder:  embedder,
		reranker:  reranker,
		cfg:       cfg,
	}

	// init class
	err = store.createClassWithAutoTenantEnabled(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	return store, nil
}

// NewWithStatusManager creates a new instance of weaviateWrapper with status manager
func NewWithStatusManager(cfg config.VectorStoreConf, embedder Embedder, reranker Reranker, statusManager *redis.StatusManager, requestId string) (Store, error) {
	var authConf auth.Config
	if cfg.Weaviate.APIKey != types.EmptyString {
		authConf = auth.ApiKey{Value: cfg.Weaviate.APIKey}
	}
	client, err := goweaviate.NewClient(goweaviate.Config{
		Host:       cfg.Weaviate.Endpoint,
		Scheme:     schemeHttp,
		AuthConfig: authConf,
		Timeout:    cfg.Weaviate.Timeout,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Weaviate client: %w", err)
	}

	store := &weaviateWrapper{
		client:        client,
		className:     cfg.Weaviate.ClassName,
		embedder:      embedder,
		reranker:      reranker,
		cfg:           cfg,
		statusManager: statusManager,
		requestId:     requestId,
	}

	// init class
	err = store.createClassWithAutoTenantEnabled(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	return store, nil
}

func (r *weaviateWrapper) GetIndexSummary(ctx context.Context, clientId string, codebasePath string) (*types.EmbeddingSummary, error) {
	start := time.Now()
	// 使用 codebaseId 作为 clientId
	tenantName, err := r.generateTenantName(clientId, codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tenant name: %w", err)
	}

	// Define GraphQL fields using proper Field type
	fields := []graphql.Field{
		{Name: "meta", Fields: []graphql.Field{
			{Name: "count"},
		}},
		{Name: "groupedBy", Fields: []graphql.Field{
			{Name: "path"},
			{Name: "value"},
		}},
	}

	res, err := r.client.GraphQL().Aggregate().
		WithClassName(r.className).
		WithFields(fields...).
		WithGroupBy(MetadataFilePath).
		WithTenant(tenantName).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get index summary: %w", err)
	}

	summary, err := r.unmarshalSummarySearchResponse(res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary response: %w", err)
	}
	tracer.WithTrace(ctx).Infof("embedding getIndexSummary end, cost %d ms on total %d files %d chunks",
		time.Since(start).Milliseconds(), summary.TotalFiles, summary.TotalChunks)
	return summary, nil
}

func (r *weaviateWrapper) GetIndexSummaryWithLanguage(ctx context.Context, clientId string, codebasePath string, language string) (*types.EmbeddingSummary, error) {
	start := time.Now()
	// 使用 codebaseId 作为 clientId
	tenantName, err := r.generateTenantName(clientId, codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tenant name: %w", err)
	}

	// Define GraphQL fields using proper Field type
	fields := []graphql.Field{
		{Name: "meta", Fields: []graphql.Field{
			{Name: "count"},
		}},
		{Name: "groupedBy", Fields: []graphql.Field{
			{Name: "path"},
			{Name: "value"},
		}},
	}

	// 创建基础聚合查询
	aggregateBuilder := r.client.GraphQL().Aggregate().
		WithClassName(r.className).
		WithFields(fields...).
		WithGroupBy(MetadataFilePath).
		WithTenant(tenantName)

	// 如果指定了语言过滤条件，则添加Where过滤器
	if language != "" {
		whereFilter := filters.Where().
			WithPath([]string{MetadataLanguage}). // MetadataLanguage = "language"
			WithOperator(filters.Equal).
			WithValueText(language)

		aggregateBuilder = aggregateBuilder.WithWhere(whereFilter)
	}

	res, err := aggregateBuilder.Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get index summary with language: %w", err)
	}

	summary, err := r.unmarshalSummarySearchResponse(res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary response: %w", err)
	}
	tracer.WithTrace(ctx).Infof("embedding getIndexSummaryWithLanguage end, cost %d ms on total %d files %d chunks with language filter: %s",
		time.Since(start).Milliseconds(), summary.TotalFiles, summary.TotalChunks, language)
	return summary, nil
}

func (r *weaviateWrapper) DeleteCodeChunks(ctx context.Context, chunks []*types.CodeChunk, options Options) error {
	if len(chunks) == 0 {
		return nil // Nothing to delete
	}

	tenant, err := r.generateTenantName(options.ClientId, options.CodebasePath)
	if err != nil {
		return err
	}
	// Build a list of filters, one for each codebaseId and filePath pair
	chunkFilters := make([]*filters.WhereBuilder, len(chunks))
	for i, chunk := range chunks {
		if chunk.CodebaseId == 0 || chunk.FilePath == types.EmptyString {
			return fmt.Errorf("invalid chunk to delete: required codebaseId and filePath")
		}
		chunkFilters[i] = filters.Where().
			WithOperator(filters.And).
			WithOperands([]*filters.WhereBuilder{
				filters.Where().
					WithPath([]string{MetadataCodebaseId}).
					WithOperator(filters.Equal).
					WithValueInt(int64(chunk.CodebaseId)),
				filters.Where().
					WithPath([]string{MetadataFilePath}).
					WithOperator(filters.Equal).
					WithValueText(chunk.FilePath),
			})
	}

	// Combine all chunk filters with OR to support batch deletion of files
	combinedFilter := filters.Where().
		WithOperator(filters.Or).
		WithOperands(chunkFilters)

	do, err := r.client.Batch().ObjectsBatchDeleter().
		WithTenant(tenant).WithWhere(
		combinedFilter,
	).WithClassName(r.className).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to send delete chunks err:%w", err)
	}
	return CheckBatchDeleteErrors(do)
}

func (r *weaviateWrapper) SimilaritySearch(ctx context.Context, query string, numDocuments int, options Options) ([]*types.SemanticFileItem, error) {
	embedQuery, err := r.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	tenantName, err := r.generateTenantName(options.ClientId, options.CodebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tenant name: %w", err)
	}

	// Define GraphQL fields using proper Field type
	fields := []graphql.Field{
		{Name: MetadataCodebaseId},
		{Name: MetadataCodebaseName},
		{Name: MetadataSyncId},
		{Name: MetadataCodebasePath},
		{Name: MetadataFilePath},
		{Name: MetadataLanguage},
		{Name: MetadataRange},
		{Name: MetadataTokenCount},
		{Name: Content},
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "certainty"},
			{Name: "distance"},
			{Name: "id"},
		}},
	}

	// Build GraphQL query with proper tenant filter
	nearVector := r.client.GraphQL().NearVectorArgBuilder().
		WithVector(embedQuery)

	// 创建基础查询
	queryBuilder := r.client.GraphQL().Get().
		WithClassName(r.className).
		WithFields(fields...).
		WithNearVector(nearVector).
		WithLimit(numDocuments).
		WithTenant(tenantName)

	// 如果指定了语言过滤条件，则添加Where过滤器
	if options.Language != "" {
		whereFilter := filters.Where().
			WithPath([]string{MetadataLanguage}). // MetadataLanguage = "language"
			WithOperator(filters.Equal).
			WithValueText(options.Language)

		queryBuilder = queryBuilder.WithWhere(whereFilter)
	}

	res, err := queryBuilder.Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	// Improved error handling for response validation
	if res == nil || res.Data == nil {
		return nil, fmt.Errorf("received empty response from Weaviate")
	}
	if err = CheckGraphQLResponseError(res); err != nil {
		return nil, fmt.Errorf("query weaviate failed: %w", err)
	}

	items, err := r.unmarshalSimilarSearchResponse(res, options.CodebasePath, options.ClientId, options.Authorization)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return items, nil
}

func (r *weaviateWrapper) unmarshalSimilarSearchResponse(res *models.GraphQLResponse, codebasePath, clientId string, authorization string) ([]*types.SemanticFileItem, error) {
	// Get the data for our class
	data, ok := res.Data["Get"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: 'Get' field not found or has wrong type")
	}

	results, ok := data[r.className].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: class data not found or has wrong type")
	}

	items := make([]*types.SemanticFileItem, 0, len(results))

	// 如果开启获取源码，则收集所有需要获取的代码片段
	var snippets []CodeSnippetRequest

	// 第一遍遍历：收集所有需要获取源码的片段信息
	for _, result := range results {
		obj, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		// content := getStringValue(obj, Content)
		filePath := getStringValue(obj, MetadataFilePath)

		// 如果开启获取源码，则从MetadataRange中提取行号信息
		if r.cfg.FetchSourceCode && filePath != "" && codebasePath != "" {
			// 从MetadataRange中提取startLine和endLine
			var startLine, endLine int
			if rangeValue, ok := obj[MetadataRange].([]interface{}); ok && len(rangeValue) >= 2 {
				if first, ok := rangeValue[0].(float64); ok {
					startLine = int(first)
				}
				if second, ok := rangeValue[2].(float64); ok {
					endLine = int(second)
				}
			}

			// 添加到批量获取列表，拼接成全路径
			fullPath := filepath.Join(codebasePath, filePath)
			snippets = append(snippets, CodeSnippetRequest{
				FilePath:  fullPath,
				StartLine: startLine,
				EndLine:   endLine,
			})
		}
	}

	// 批量获取代码片段内容
	var contentMap map[string]string
	if len(snippets) > 0 && codebasePath != "" {
		var err error
		contentMap, err = fetchCodeContentsBatch(context.Background(), r.cfg, clientId, codebasePath, snippets, authorization)
		if err != nil {
			return nil, fmt.Errorf("批量获取代码片段失败: %w", err)
		}
	}

	// 第二遍遍历：构建最终的SemanticFileItem列表
	for _, result := range results {
		obj, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract additional properties
		additional, ok := obj["_additional"].(map[string]interface{})
		if !ok {
			continue
		}

		content := getStringValue(obj, Content)

		filePath := getStringValue(obj, MetadataFilePath)

		// 从MetadataRange中提取startLine和endLine（用于构建映射键）
		var startLine, endLine int
		if rangeValue, ok := obj[MetadataRange].([]interface{}); ok && len(rangeValue) >= 2 {
			if first, ok := rangeValue[0].(float64); ok {
				startLine = int(first)
			}
			if second, ok := rangeValue[2].(float64); ok {
				endLine = int(second)
			}
		}

		// 如果开启获取源码且有批量获取的内容，则使用获取到的内容
		if r.cfg.FetchSourceCode && filePath != "" && codebasePath != "" {

			// 构建映射键并查找批量获取的内容
			fullPath := filepath.Join(codebasePath, filePath)
			key := fmt.Sprintf("%s:%d-%d", fullPath, startLine, endLine)
			if fetchedContent, exists := contentMap[key]; exists && fetchedContent != "" {
				content = fetchedContent
			}
		}

		// Create SemanticFileItem with proper fields
		item := &types.SemanticFileItem{
			Content:   content,
			FilePath:  filePath,
			StartLine: startLine,
			EndLine:   endLine,
			Score:     float32(getFloatValue(additional, "certainty")), // Convert float64 to float32
		}

		items = append(items, item)
	}

	return items, nil
}

// Helper functions for safe type conversion
func getStringValue(obj map[string]interface{}, key string) string {
	if val, ok := obj[key].(string); ok {
		return val
	}
	return ""
}

func getFloatValue(obj map[string]interface{}, key string) float64 {
	if val, ok := obj[key].(float64); ok {
		return val
	}
	return 0
}

func (r *weaviateWrapper) GetCodebaseRecords(ctx context.Context, clientId string, codebasePath string) ([]*types.CodebaseRecord, error) {
	// 添加调试日志
	fmt.Printf("[DEBUG] GetCodebaseRecords - 开始执行，clientId: %s, codebasePath: %s\n", clientId, codebasePath)

	// 检查输入参数
	if clientId == "" {
		fmt.Printf("[DEBUG] 警告: codebaseId 为 为空字符串\n")
	}
	if codebasePath == "" {
		fmt.Printf("[DEBUG] 警告: codebasePath 为空字符串\n")
	}

	tenantName, err := r.generateTenantName(clientId, codebasePath)
	if err != nil {
		fmt.Printf("[DEBUG] 生成 tenantName 失败: %v\n", err)
		return nil, fmt.Errorf("failed to generate tenant name: %w", err)
	}
	fmt.Printf("[DEBUG] 生成的 tenantName: %s\n", tenantName)

	// 添加调试日志：检查 Weaviate 连接状态
	live, err := r.client.Misc().LiveChecker().Do(ctx)
	if err != nil {
		fmt.Printf("[DEBUG] Weaviate 连接检查失败: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] Weaviate 连接状态: %v\n", live)
	}

	// 定义GraphQL字段
	fields := []graphql.Field{
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "id"},
			{Name: "lastUpdateTimeUnix"},
		}},
		{Name: MetadataFilePath},
		{Name: MetadataLanguage},
		{Name: Content},
		{Name: MetadataRange},
		{Name: MetadataTokenCount},
		{Name: MetadataCodebaseId},
		{Name: MetadataCodebasePath},
		{Name: MetadataCodebaseName},
		{Name: MetadataSyncId},
	}

	// 执行查询，获取所有记录
	var allRecords []*types.CodebaseRecord
	limit := 1000 // 每批获取1000条记录
	offset := 0

	for {
		fmt.Printf("[DEBUG] 执行 GraphQL 查询 - offset: %d, limit: %d\n", offset, limit)
		fmt.Printf("[DEBUG] GraphQL 查询参数 - className: %s, tenant: %s, clientId: %s\n",
			r.className, tenantName, clientId)

		res, err := r.client.GraphQL().Get().
			WithClassName(r.className).
			WithFields(fields...).
			WithLimit(limit).
			WithOffset(offset).
			WithTenant(tenantName).
			Do(ctx)

		if err != nil {
			fmt.Printf("[DEBUG] GraphQL 查询失败: %v\n", err)
			return nil, fmt.Errorf("failed to get codebase records: %w", err)
		}

		if res == nil || res.Data == nil {
			fmt.Printf("[DEBUG] 响应为空，结束查询 - 可能 tenant %s 中没有数据\n", tenantName)
			break
		}

		// 解析响应
		records, err := r.unmarshalCodebaseRecordsResponse(res)
		if err != nil {
			fmt.Printf("[DEBUG] 解析响应失败: %v\n", err)
			return nil, fmt.Errorf("failed to unmarshal records response: %w", err)
		}

		fmt.Printf("[DEBUG] 本批次获取记录数: %d\n", len(records))
		if len(records) == 0 {
			fmt.Printf("[DEBUG] 没有更多记录，结束查询 - tenant %s 中可能没有 clientId %s 的数据\n", tenantName, clientId)
			break
		}

		allRecords = append(allRecords, records...)
		offset += limit

		// 如果获取的记录数小于limit，说明已经获取完所有记录
		if len(records) < limit {
			break
		}
	}

	return allRecords, nil
}

func (r *weaviateWrapper) unmarshalCodebaseRecordsResponse(res *models.GraphQLResponse) ([]*types.CodebaseRecord, error) {
	if len(res.Errors) > 0 {
		var errMsg string
		for _, e := range res.Errors {
			errMsg += e.Message
		}
		return nil, fmt.Errorf("failed to get codebase records: %s", errMsg)
	}

	// 检查响应是否为空
	if res == nil || res.Data == nil {
		fmt.Printf("[DEBUG] 响应为空，返回 nil 记录\n")
		return nil, nil
	}

	// 获取 Get 字段
	data, ok := res.Data["Get"].(map[string]interface{})
	if !ok {
		fmt.Printf("[DEBUG] 响应格式错误：'Get' 字段不存在或类型错误\n")
		return nil, fmt.Errorf("invalid response format: 'Get' field not found or has wrong type")
	}

	// 获取类名对应的数据
	results, ok := data[r.className].([]interface{})
	if !ok {
		fmt.Printf("[DEBUG] 响应格式错误：类数据不存在或类型错误，类名: %s\n", r.className)
		return nil, fmt.Errorf("invalid response format: class data not found or has wrong type")
	}

	fmt.Printf("[DEBUG] 解析响应，原始结果数量: %d\n", len(results))

	records := make([]*types.CodebaseRecord, 0, len(results))
	uniquePaths := make(map[string]int) // 跟踪唯一路径

	for i, result := range results {
		obj, ok := result.(map[string]interface{})
		if !ok {
			fmt.Printf("[DEBUG] 跳过结果 %d：不是有效的 map[string]interface{} 类型\n", i)
			continue
		}

		// 提取附加属性
		additional, ok := obj["_additional"].(map[string]interface{})
		if !ok {
			fmt.Printf("[DEBUG] 跳过结果 %d：_additional 字段不存在或类型错误\n", i)
			continue
		}

		// 解析最后更新时间
		var lastUpdated time.Time
		if lastUpdateUnix, ok := additional["lastUpdateTimeUnix"].(float64); ok {
			lastUpdated = time.Unix(int64(lastUpdateUnix), 0)
		} else {
			lastUpdated = time.Now()
		}

		// 解析范围信息
		var rangeInfo []int
		if rangeData, ok := obj[MetadataRange].([]interface{}); ok {
			rangeInfo = make([]int, len(rangeData))
			for i, v := range rangeData {
				if num, ok := v.(float64); ok {
					rangeInfo[i] = int(num)
				}
			}
		}

		filePath := getStringValue(obj, MetadataFilePath)
		record := &types.CodebaseRecord{
			Id:           getStringValue(additional, "id"),
			FilePath:     filePath,
			Language:     getStringValue(obj, MetadataLanguage),
			Content:      getStringValue(obj, Content),
			Range:        rangeInfo,
			TokenCount:   int(getFloatValue(obj, MetadataTokenCount)),
			LastUpdated:  lastUpdated,
			CodebaseId:   int32(getFloatValue(obj, MetadataCodebaseId)),
			CodebasePath: getStringValue(obj, MetadataCodebasePath),
			CodebaseName: getStringValue(obj, MetadataCodebaseName),
			SyncId:       int32(getFloatValue(obj, MetadataSyncId)),
		}

		records = append(records, record)
		uniquePaths[filePath]++

	}

	return records, nil
}

func (r *weaviateWrapper) Close() {
}

func (r *weaviateWrapper) DeleteByCodebase(ctx context.Context, clientId string, codebasePath string) error {

	// 使用 codebaseId 作为 clientId
	tenant, err := r.generateTenantName(clientId, codebasePath)
	if err != nil {
		return err
	}

	// 构建过滤器：根据 codebasePath 删除所有相关的 chunks
	filter := filters.Where().
		WithOperator(filters.And).
		WithOperands([]*filters.WhereBuilder{
			filters.Where().
				WithPath([]string{MetadataCodebasePath}).
				WithOperator(filters.Equal).
				WithValueText(codebasePath),
		})

	do, err := r.client.Batch().ObjectsBatchDeleter().
		WithTenant(tenant).WithWhere(filter).WithClassName(r.className).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to send delete codebase chunks, err:%w", err)
	}
	return CheckBatchDeleteErrors(do)
}

func (r *weaviateWrapper) UpsertCodeChunks(ctx context.Context, docs []*types.CodeChunk, options Options) error {
	if len(docs) == 0 {
		return nil
	}
	// TODO 事务保障
	// 先删除已有的相同codebaseId和FilePath的数据，避免重复
	//TODO 启动一个定时任务，清理重复数据。根据CodebaseId、FilePaths、Content 去重。
	// TODO 区分添加、修改、删除场景， 只有修改/删除需要先delete，添加不用。
	err := r.DeleteCodeChunks(ctx, docs, options)
	if err != nil {
		tracer.WithTrace(ctx).Errorf("[%s]failed to delete existing code chunks before upsert: %v", docs[0].CodebasePath, err)
	}

	return r.InsertCodeChunks(ctx, docs, options)
}

// UpdateCodeChunksPaths 直接更新代码块的文件路径，而不是删除再插入
func (r *weaviateWrapper) UpdateCodeChunksPaths(ctx context.Context, updates []*types.CodeChunkPathUpdate, options Options) error {

	if len(updates) == 0 {
		tracer.WithTrace(ctx).Errorf("UpdateCodeChunksPaths len(updates): %v", len(updates))
		return nil
	}

	tenantName, err := r.generateTenantName(options.ClientId, options.CodebasePath)
	if err != nil {
		return err
	}

	// 对于每个更新，使用GraphQL更新操作
	for _, update := range updates {
		if update.OldFilePath == types.EmptyString || update.NewFilePath == types.EmptyString || update.CodebaseId == 0 {
			return fmt.Errorf("invalid chunk path update: required fields: CodebaseId, OldFilePath, NewFilePath")
		}

		// 首先获取要更新的对象ID和完整数据
		records, err := r.getRecordsByPath(ctx, update.OldFilePath, tenantName)
		if err != nil {
			return fmt.Errorf("failed to get records for path %s: %w", update.OldFilePath, err)
		}

		tracer.WithTrace(ctx).Errorf(" 更新找到记录数: %v", len(records))

		if len(records) == 0 {
			// 没有找到要更新的对象，跳过
			continue
		}

		// 对每个记录执行更新操作
		for _, record := range records {
			tracer.WithTrace(ctx).Errorf("%s  ->  %s  updateObjectPath: %v", update.OldFilePath, update.NewFilePath, record.Id)

			err := r.updateObjectPath(ctx, record.Id, update.NewFilePath, tenantName, record)
			if err != nil {
				return fmt.Errorf("failed to update object %s path from %s to %s: %w", record.Id, update.OldFilePath, update.NewFilePath, err)
			}
		}
	}

	tracer.WithTrace(ctx).Infof("updated %d chunk paths for codebase %s successfully", len(updates), options.CodebasePath)
	return nil
}

// getRecordsByPath 根据文件路径获取完整的记录
func (r *weaviateWrapper) getRecordsByPath(ctx context.Context, filePath string, tenantName string) ([]*types.CodebaseRecord, error) {
	// 定义GraphQL字段
	fields := []graphql.Field{
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "id"},
			{Name: "lastUpdateTimeUnix"},
		}},
		{Name: MetadataFilePath},
		{Name: MetadataLanguage},
		{Name: Content},
		{Name: MetadataRange},
		{Name: MetadataTokenCount},
		{Name: MetadataCodebaseId},
		{Name: MetadataCodebasePath},
		{Name: MetadataCodebaseName},
		{Name: MetadataSyncId},
	}

	// 构建过滤器
	filter := filters.Where().
		WithOperator(filters.And).
		WithOperands([]*filters.WhereBuilder{
			filters.Where().
				WithPath([]string{MetadataFilePath}).
				WithOperator(filters.Equal).
				WithValueText(filePath),
		})

	// 执行查询
	res, err := r.client.GraphQL().Get().
		WithClassName(r.className).
		WithFields(fields...).
		WithWhere(filter).
		WithTenant(tenantName).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query objects by path: %w", err)
	}

	if res == nil || res.Data == nil {
		return nil, nil
	}

	// 解析响应获取记录
	return r.unmarshalRecordsResponse(res)
}

// unmarshalRecordsResponse 从GraphQL响应中解析CodebaseRecord列表
func (r *weaviateWrapper) unmarshalRecordsResponse(res *models.GraphQLResponse) ([]*types.CodebaseRecord, error) {
	if len(res.Errors) > 0 {
		var errMsg string
		for _, e := range res.Errors {
			errMsg += e.Message
		}
		return nil, fmt.Errorf("failed to get records: %s", errMsg)
	}

	if res == nil || res.Data == nil {
		return nil, nil
	}

	data, ok := res.Data["Get"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: 'Get' field not found or has wrong type")
	}

	results, ok := data[r.className].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: class data not found or has wrong type")
	}

	var records []*types.CodebaseRecord
	for _, result := range results {
		obj, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		additional, ok := obj["_additional"].(map[string]interface{})
		if !ok {
			continue
		}

		id, ok := additional["id"].(string)
		if !ok {
			continue
		}

		lastUpdateTimeUnix, ok := additional["lastUpdateTimeUnix"].(float64)
		if !ok {
			lastUpdateTimeUnix = 0
		}

		filePath, ok := obj[MetadataFilePath].(string)
		if !ok {
			continue
		}

		language, ok := obj[MetadataLanguage].(string)
		if !ok {
			language = ""
		}

		content, ok := obj[Content].(string)
		if !ok {
			content = ""
		}

		rangeData, ok := obj[MetadataRange].([]interface{})
		var rangeInt []int
		if ok {
			for _, r := range rangeData {
				if val, ok := r.(float64); ok {
					rangeInt = append(rangeInt, int(val))
				}
			}
		}

		tokenCount, ok := obj[MetadataTokenCount].(float64)
		if !ok {
			tokenCount = 0
		}

		// 解析新增的字段
		codebaseId, _ := obj[MetadataCodebaseId].(float64)
		codebasePath, _ := obj[MetadataCodebasePath].(string)
		codebaseName, _ := obj[MetadataCodebaseName].(string)

		syncId, _ := obj[MetadataSyncId].(float64)

		record := &types.CodebaseRecord{
			Id:          id,
			FilePath:    filePath,
			Language:    language,
			Content:     content,
			Range:       rangeInt,
			TokenCount:  int(tokenCount),
			LastUpdated: time.Unix(int64(lastUpdateTimeUnix), 0),
			// 新增字段
			CodebaseId:   int32(codebaseId),
			CodebasePath: codebasePath,
			CodebaseName: codebaseName,
			SyncId:       int32(syncId),
		}

		records = append(records, record)
	}

	return records, nil
}

// getObjectIdsByPath 根据文件路径获取对象ID
func (r *weaviateWrapper) getObjectIdsByPath(ctx context.Context, codebaseId int32, filePath string, tenantName string) ([]string, error) {
	// 定义GraphQL字段
	fields := []graphql.Field{
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "id"},
		}},
	}

	// 构建过滤器
	filter := filters.Where().
		WithOperator(filters.And).
		WithOperands([]*filters.WhereBuilder{
			filters.Where().
				WithPath([]string{MetadataCodebaseId}).
				WithOperator(filters.Equal).
				WithValueInt(int64(codebaseId)),
			filters.Where().
				WithPath([]string{MetadataFilePath}).
				WithOperator(filters.Equal).
				WithValueText(filePath),
		})

	// 执行查询
	res, err := r.client.GraphQL().Get().
		WithClassName(r.className).
		WithFields(fields...).
		WithWhere(filter).
		WithTenant(tenantName).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query objects by path: %w", err)
	}

	if res == nil || res.Data == nil {
		return nil, nil
	}

	// 解析响应获取ID
	return r.unmarshalIdsResponse(res)
}

// unmarshalIdsResponse 从GraphQL响应中解析ID列表
func (r *weaviateWrapper) unmarshalIdsResponse(res *models.GraphQLResponse) ([]string, error) {
	if len(res.Errors) > 0 {
		var errMsg string
		for _, e := range res.Errors {
			errMsg += e.Message
		}
		return nil, fmt.Errorf("failed to get object ids: %s", errMsg)
	}

	if res == nil || res.Data == nil {
		return nil, nil
	}

	data, ok := res.Data["Get"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: 'Get' field not found or has wrong type")
	}

	results, ok := data[r.className].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: class data not found or has wrong type")
	}

	var ids []string
	for _, result := range results {
		obj, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		additional, ok := obj["_additional"].(map[string]interface{})
		if !ok {
			continue
		}

		if id, ok := additional["id"].(string); ok {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// updateObjectPath 更新单个对象的路径
func (r *weaviateWrapper) updateObjectPath(ctx context.Context, id string, newFilePath string, tenantName string, record *types.CodebaseRecord) error {
	// 使用Weaviate的REST API直接更新对象
	// 构建更新请求体，包含所有原有属性
	updateData := map[string]interface{}{
		MetadataFilePath:     newFilePath,
		MetadataLanguage:     record.Language,
		Content:              record.Content,
		MetadataRange:        record.Range,
		MetadataTokenCount:   record.TokenCount,
		MetadataCodebaseId:   record.CodebaseId,
		MetadataCodebasePath: record.CodebasePath,
		MetadataCodebaseName: record.CodebaseName,
		MetadataSyncId:       record.SyncId,
	}

	// 执行更新
	err := r.client.Data().Updater().
		WithID(id).
		WithClassName(r.className).
		WithTenant(tenantName).
		WithProperties(updateData).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to update object path: %w", err)
	}

	return nil
}

func (r *weaviateWrapper) InsertCodeChunks(ctx context.Context, docs []*types.CodeChunk, options Options) error {
	if len(docs) == 0 {
		return nil
	}
	tenantName, err := r.generateTenantName(options.ClientId, docs[0].CodebasePath)
	if err != nil {
		return err
	}

	tracer.WithTrace(ctx).Infof("InsertCodeChunks options.RequestId: %s ", options.RequestId)
	// 如果有状态管理器和请求ID，则使用带有状态管理器的 embedder
	var chunks []*CodeChunkEmbedding
	if r.statusManager != nil && options.RequestId != "" {
		// 创建带有状态管理器的临时 embedder
		embedderWithStatus, err := NewEmbedderWithStatusManager(r.cfg.Embedder, r.statusManager, options.RequestId, options.TotalFiles)
		if err != nil {
			return fmt.Errorf("failed to create embedder with status manager: %w", err)
		}
		chunks, err = embedderWithStatus.EmbedCodeChunks(ctx, docs)
	} else {
		// 使用原有的 embedder
		chunks, err = r.embedder.EmbedCodeChunks(ctx, docs)
	}

	if err != nil {
		return err
	}
	tracer.WithTrace(ctx).Infof("embedded %d chunks for codebase %s successfully", len(chunks), docs[0].CodebaseName)

	objs := make([]*models.Object, len(chunks), len(chunks))
	for i, c := range chunks {
		if c.FilePath == types.EmptyString || c.CodebaseId == 0 || c.CodebasePath == types.EmptyString {
			return fmt.Errorf("invalid chunk to write: required fields: CodebaseId, CodebasePath, FilePaths")
		}

		// 根据配置决定是否存储Content代码片段
		properties := map[string]any{
			MetadataFilePath:     c.FilePath,
			MetadataLanguage:     c.Language,
			MetadataCodebaseId:   c.CodebaseId,
			MetadataCodebasePath: options.CodebasePath,
			MetadataCodebaseName: options.CodebaseName,
			MetadataSyncId:       options.SyncId,
			MetadataRange:        c.Range,
			MetadataTokenCount:   c.TokenCount,
			Content:              "",
		}

		// 如果配置中启用了StoreSourceCode，则存储源码
		if r.cfg.StoreSourceCode {
			properties[Content] = string(c.Content)
		}

		objs[i] = &models.Object{
			ID:         strfmt.UUID(uuid.New().String()),
			Class:      r.className,
			Tenant:     tenantName,
			Vector:     c.Embedding,
			Properties: properties,
		}
	}
	tracer.WithTrace(ctx).Infof("start to save %d chunks for codebase %s successfully", len(docs), docs[0].CodebaseName)
	resp, err := r.client.Batch().ObjectsBatcher().WithObjects(objs...).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to send batch to Weaviate: %w", err)
	}
	if err = CheckBatchErrors(resp); err != nil {
		return fmt.Errorf("failed to send batch to Weaviate: %w", err)
	}
	tracer.WithTrace(ctx).Infof("save %d chunks for codebase %s successfully", len(docs), docs[0].CodebaseName)
	return nil
}

func (r *weaviateWrapper) Query(ctx context.Context, query string, topK int, options Options) ([]*types.SemanticFileItem, error) {
	documents, err := r.SimilaritySearch(ctx, query, r.cfg.Weaviate.MaxDocuments, options)

	if err != nil {
		return nil, err
	}
	//  调用reranker模型进行重排
	rerankedDocs, err := r.reranker.Rerank(ctx, query, documents)
	if err != nil {
		tracer.WithTrace(ctx).Errorf("failed customReranker docs: %v", err)
	}
	if len(rerankedDocs) == 0 {
		rerankedDocs = documents
	}
	// topK
	rerankedDocs = rerankedDocs[:int(math.Min(float64(topK), float64(len(rerankedDocs))))]
	return rerankedDocs, nil
}

// CodeSnippetRequest 代码片段请求结构
type CodeSnippetRequest struct {
	FilePath  string `json:"filePath"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

// CodeSnippetsBatchRequest 批量代码片段请求结构
type CodeSnippetsBatchRequest struct {
	ClientId      string               `json:"clientId"`
	WorkspacePath string               `json:"workspacePath"`
	CodeSnippets  []CodeSnippetRequest `json:"codeSnippets"`
}

// CodeSnippetResponse 代码片段响应结构
type CodeSnippetResponse struct {
	FilePath  string `json:"filePath"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Content   string `json:"content"`
}

// CodeSnippetsBatchResponse 批量代码片段响应结构
type CodeSnippetsBatchResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		List []CodeSnippetResponse `json:"list"`
	} `json:"data"`
}

// fetchCodeContentsBatch 批量获取代码片段Content
func fetchCodeContentsBatch(ctx context.Context, cfg config.VectorStoreConf, clientId, codebasePath string, snippets []CodeSnippetRequest, authorization string) (map[string]string, error) {
	if len(snippets) == 0 {
		return nil, nil
	}

	// 构建请求体
	request := CodeSnippetsBatchRequest{
		ClientId:      clientId,
		WorkspacePath: codebasePath,
		CodeSnippets:  snippets,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建API请求URL
	apiURL := cfg.BaseURL

	tracer.WithTrace(ctx).Infof("fetchCodeContentsBatch: %s", apiURL)

	// 创建HTTP请求
	tracer.WithTrace(ctx).Infof("fetchCodeContentsBatch: jsonData %s", string(jsonData))
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Costrict-Version", "v1.6.0")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	// 发送HTTP POST请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch code contents batch: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		// 读取错误响应体
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code: %d, failed to read error response: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(errorBody))
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var batchResponse CodeSnippetsBatchResponse
	if err := json.Unmarshal(body, &batchResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !batchResponse.Success {
		return nil, fmt.Errorf("API request failed: %s", batchResponse.Message)
	}

	// 构建filePath到content的映射
	contentMap := make(map[string]string)
	for _, snippet := range batchResponse.Data.List {
		key := fmt.Sprintf("%s:%d-%d", snippet.FilePath, snippet.StartLine, snippet.EndLine)
		contentMap[key] = snippet.Content
	}

	return contentMap, nil
}

// fetchCodeContent 通过API获取代码片段的Content
func fetchCodeContent(ctx context.Context, cfg config.VectorStoreConf, clientId, codebasePath, filePath string, startLine, endLine int, authorization string) (string, error) {
	// 构建API请求URL
	baseURL := cfg.BaseURL

	// 对参数进行URL编码
	encodedCodebasePath := url.QueryEscape(codebasePath)

	// 如果filePath是全路径，则与codebasePath拼接处理
	var processedFilePath string
	if strings.HasPrefix(filePath, "/") {
		// filePath是全路径，直接使用
		processedFilePath = filePath
	} else {
		// filePath是相对路径，与codebasePath拼接
		processedFilePath = fmt.Sprintf("%s/%s", strings.TrimSuffix(codebasePath, "/"), filePath)
	}

	// 检查操作系统类型，如果是Windows则将路径转换为Windows格式
	if runtime.GOOS == "windows" {
		// 将Unix风格的路径转换为Windows风格
		processedFilePath = filepath.FromSlash(processedFilePath)
		// 确保路径是绝对路径格式
		if !strings.HasPrefix(processedFilePath, "\\") && !strings.Contains(processedFilePath, ":") {
			// 如果不是网络路径也不是驱动器路径，添加当前驱动器
			processedFilePath = filepath.Join(filepath.VolumeName("."), processedFilePath)
		}
	}

	encodedFilePath := url.QueryEscape(processedFilePath)

	// 构建完整的请求URL
	requestURL := fmt.Sprintf("%s?clientId=%s&codebasePath=%s&filePath=%s&startLine=%d&endLine=%d",
		baseURL, clientId, encodedCodebasePath, encodedFilePath, startLine, endLine)

	tracer.WithTrace(ctx).Infof("fetchCodeContent %s: ", requestURL)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 添加Authorization头
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	// 发送HTTP GET请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch code content: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func (r *weaviateWrapper) createClassWithAutoTenantEnabled(client *goweaviate.Client) error {
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelFunc()
	tracer.WithTrace(timeout).Infof("start to create weaviate class %s", r.className)
	res, err := client.Schema().ClassExistenceChecker().WithClassName(r.className).Do(timeout)
	if err != nil {
		tracer.WithTrace(timeout).Errorf("check weaviate class exists err:%v", err)
	}
	if err == nil && res {
		tracer.WithTrace(timeout).Infof("weaviate class %s already exists, not create.", r.className)
		return nil
	}

	// 定义类的属性并配置索引
	dynamicConf := dynamic.NewDefaultUserConfig()
	class := &models.Class{
		Class:      r.className,
		Properties: classProperties, // fields
		// auto create tenant
		MultiTenancyConfig: &models.MultiTenancyConfig{
			Enabled:            true,
			AutoTenantCreation: true,
		},
		VectorIndexType:   dynamicConf.IndexType(),
		VectorIndexConfig: dynamicConf,
	}

	tracer.WithTrace(timeout).Infof("class info:%v", class)
	err = client.Schema().ClassCreator().WithClass(class).Do(timeout)
	// TODO skip already exists err
	if err != nil && strings.Contains(err.Error(), "already exists") {
		tracer.WithTrace(timeout).Infof("weaviate class %s already exists, not create.", r.className)
		return nil
	}
	tracer.WithTrace(timeout).Infof("weaviate class %s end.", r.className)
	return err
}

// generateTenantName 使用 MD5 哈希生成合规租户名（32字符，纯十六进制）
func (r *weaviateWrapper) generateTenantName(clientId string, codebasePath string) (string, error) {
	// 添加调试日志
	logx.Debugf("[DEBUG] generateTenantName - 输入 clientId: %s, codebasePath: %s\n", clientId, codebasePath)

	if codebasePath == types.EmptyString {
		logx.Debugf("[DEBUG] generateTenantName - codebasePath 为空字符串\n")
		return types.EmptyString, ErrInvalidCodebasePath
	}
	if clientId == types.EmptyString {
		logx.Debugf("[DEBUG] generateTenantName - clientId 为空字符串\n")
		return types.EmptyString, ErrInvalidClientId
	}

	// 将 clientId 和 codebasePath 组合起来生成哈希
	combined := clientId + ":" + codebasePath
	hash := md5.Sum([]byte(combined))         // 计算 MD5 哈希
	tenantName := hex.EncodeToString(hash[:]) // 转为32位十六进制字符串

	logx.Debugf("[DEBUG] generateTenantName - 生成的 tenantName: %s\n", tenantName)
	return tenantName, nil
}

func (r *weaviateWrapper) unmarshalSummarySearchResponse(res *models.GraphQLResponse) (*types.EmbeddingSummary, error) {
	if len(res.Errors) > 0 {
		var errMsg string
		for _, e := range res.Errors {
			errMsg += e.Message
		}
		return nil, fmt.Errorf("failed to get embedding summary: %s", errMsg)
	}
	// 检查响应是否为空
	if res == nil || res.Data == nil {
		return nil, fmt.Errorf("received empty response from Weaviate")
	}

	// 获取 Aggregate 字段
	data, ok := res.Data["Aggregate"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: 'Aggregate' field not found or has wrong type")
	}

	// 获取类名对应的数据
	results, ok := data[r.className].([]interface{})
	// if !ok || len(results) == 0 {
	if !ok {
		return nil, fmt.Errorf("invalid response format: class data not found or has wrong type：%s", reflect.TypeOf(results).String())
	}
	var totalChunks, totalFiles int
	for _, v := range results {
		// 获取 meta 字段
		result, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invaid response format, result has wrong type: %s", reflect.TypeOf(result).String())
		}
		meta, ok := result["meta"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response format: 'meta' field not found or has wrong type:%s", reflect.TypeOf(meta).String())
		}

		// 获取总数
		count, ok := meta["count"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid response format: 'count' field not found or has wrong type:%s", reflect.TypeOf(count).String())
		}
		totalChunks += int(count)
		totalFiles++

	}

	return &types.EmbeddingSummary{
		TotalFiles:  totalFiles,
		TotalChunks: totalChunks,
	}, nil
}

// GetFileRecords 根据文件路径获取代码记录
func (r *weaviateWrapper) GetFileRecords(ctx context.Context, clientId string, codebasePath string, filePath string) ([]*types.CodebaseRecord, error) {
	// 生成租户名称
	// 使用默认的 clientId，因为函数参数中没有提供
	tenantName, err := r.generateTenantName(clientId, codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tenant name: %w", err)
	}

	// 调用内部方法获取记录
	return r.getRecordsByPath(ctx, filePath, tenantName)
}

// GetDictionaryRecords 获取指定目录的记录，通过匹配filePath的前缀
func (r *weaviateWrapper) GetDictionaryRecords(ctx context.Context, clientId string, codebasePath string, dictionary string) ([]*types.CodebaseRecord, error) {
	// 生成租户名称
	// 使用默认的 clientId，因为函数参数中没有提供
	tenantName, err := r.generateTenantName(clientId, codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tenant name: %w", err)
	}

	// 调用内部方法获取目录记录
	return r.getRecordsByPathPrefix(ctx, dictionary, tenantName)
}

// getRecordsByPathPrefix 根据路径前缀获取记录
func (r *weaviateWrapper) getRecordsByPathPrefix(ctx context.Context, pathPrefix string, tenantName string) ([]*types.CodebaseRecord, error) {
	// 确保路径前缀以/结尾，以便正确匹配子目录和文件
	if pathPrefix != "" && !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}

	// 定义GraphQL字段
	fields := []graphql.Field{
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "id"},
			{Name: "lastUpdateTimeUnix"},
		}},
		{Name: MetadataFilePath},
		{Name: MetadataLanguage},
		{Name: Content},
		{Name: MetadataRange},
		{Name: MetadataTokenCount},
		{Name: MetadataCodebaseId},
		{Name: MetadataCodebasePath},
		{Name: MetadataCodebaseName},
		{Name: MetadataSyncId},
	}

	// 构建过滤器：使用Like操作符匹配filePath前缀
	filter := filters.Where().
		WithOperator(filters.And).
		WithOperands([]*filters.WhereBuilder{
			filters.Where().
				WithPath([]string{MetadataFilePath}).
				WithOperator(filters.Like).
				WithValueText(pathPrefix + "%"),
		})

	// 执行查询
	res, err := r.client.GraphQL().Get().
		WithClassName(r.className).
		WithFields(fields...).
		WithWhere(filter).
		WithTenant(tenantName).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query objects by path prefix: %w", err)
	}

	if res == nil || res.Data == nil {
		return nil, nil
	}

	// 解析响应获取记录
	return r.unmarshalRecordsResponse(res)
}

// DeleteDictionary 删除指定目录的记录，通过匹配filePath的前缀
func (r *weaviateWrapper) DeleteDictionary(ctx context.Context, dictionary string, options Options) error {
	// 生成租户名称
	tenantName, err := r.generateTenantName(options.ClientId, options.CodebasePath)
	if err != nil {
		return fmt.Errorf("failed to generate tenant name: %w", err)
	}

	// 确保路径前缀以/结尾，以便正确匹配子目录和文件
	if dictionary != "" && !strings.HasSuffix(dictionary, "/") {
		dictionary += "/"
	}

	// 获取所有匹配目录前缀的记录
	records, err := r.getRecordsByPathPrefix(ctx, dictionary, tenantName)
	if err != nil {
		return fmt.Errorf("failed to get records by path prefix: %w", err)
	}

	if len(records) == 0 {
		// 没有找到需要删除的记录
		return nil
	}

	// 将记录转换为CodeChunk格式以便删除
	chunks := make([]*types.CodeChunk, 0, len(records))
	for _, record := range records {
		chunk := &types.CodeChunk{
			CodebaseId: record.CodebaseId,
			FilePath:   record.FilePath,
		}
		chunks = append(chunks, chunk)
	}

	if len(chunks) > 0 {
		chunkFilters := make([]*filters.WhereBuilder, len(chunks))
		for i, chunk := range chunks {
			if chunk.FilePath == types.EmptyString {
				return fmt.Errorf("invalid chunk to delete: and filePath")
			}
			chunkFilters[i] = filters.Where().
				WithOperator(filters.And).
				WithOperands([]*filters.WhereBuilder{
					filters.Where().
						WithPath([]string{MetadataCodebaseId}).
						WithOperator(filters.Equal).
						WithValueInt(int64(chunk.CodebaseId)),
					filters.Where().
						WithPath([]string{MetadataFilePath}).
						WithOperator(filters.Equal).
						WithValueText(chunk.FilePath),
				})
		}

		// Combine all chunk filters with OR to support batch deletion of files
		combinedFilter := filters.Where().
			WithOperator(filters.Or).
			WithOperands(chunkFilters)

		do, err := r.client.Batch().ObjectsBatchDeleter().
			WithTenant(tenantName).WithWhere(
			combinedFilter,
		).WithClassName(r.className).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to send delete chunks err:%w", err)
		}
		return CheckBatchDeleteErrors(do)
	}

	// 批量删除记录
	return nil
}

// UpdateCodeChunksDictionary 更新代码块的目录路径，通过匹配filePath的前缀
func (r *weaviateWrapper) UpdateCodeChunksDictionary(ctx context.Context, clientId string, codebasePath string, dictionary string, newDictionary string) error {
	// 生成租户名称
	tenantName, err := r.generateTenantName(clientId, codebasePath)
	if err != nil {
		return fmt.Errorf("failed to generate tenant name: %w", err)
	}

	// 确保路径前缀以/结尾，以便正确匹配子目录和文件
	if dictionary != "" && !strings.HasSuffix(dictionary, "/") {
		dictionary += "/"
	}
	if newDictionary != "" && !strings.HasSuffix(newDictionary, "/") {
		newDictionary += "/"
	}

	// 获取所有匹配原目录前缀的记录
	records, err := r.getRecordsByPathPrefix(ctx, dictionary, tenantName)
	if err != nil {
		return fmt.Errorf("failed to get records by path prefix: %w", err)
	}

	if len(records) == 0 {
		// 没有找到需要更新的记录
		return nil
	}

	// 批量更新记录路径
	for _, record := range records {
		// 构建新的文件路径：将原目录前缀替换为新目录前缀
		newFilePath := strings.Replace(record.FilePath, dictionary, newDictionary, 1)

		// 获取对象的ID
		objectIds, err := r.getObjectIdsByPath(ctx, record.CodebaseId, record.FilePath, tenantName)
		if err != nil {
			return fmt.Errorf("failed to get object ids by path: %w", err)
		}

		// 更新每个对象的路径
		for _, objectId := range objectIds {
			err = r.updateObjectPath(ctx, objectId, newFilePath, tenantName, record)
			if err != nil {
				return fmt.Errorf("failed to update object path: %w", err)
			}
		}
	}

	return nil
}
