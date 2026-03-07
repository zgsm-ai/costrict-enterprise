package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oasdiff/yaml"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/embedding"

	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

var baseDir, _ = filepath.Abs("../")

// TestRunner 测试运行器
// 负责协调整个测试流程，包括文件加载、查询执行、指标计算和结果输出
// 支持嵌入模型和向量数据库的对比测试，以及多种输出格式
type TestRunner struct {
	config     TestConfig                   // 测试配置
	testFiles  map[string]*types.SourceFile // 测试文件映射
	queries    []QueryConfig                // 查询配置列表
	fileLoader *FileLoader                  // 文件加载器
	metrics    *MetricsCalculator           // 指标计算器
	logger     *log.Logger                  // 日志记录器
	mu         sync.RWMutex                 // 读写锁，用于并发安全
}

// NewTestRunner 创建新的测试运行器
// 初始化测试运行器所需的所有组件，包括文件加载器、指标计算器和日志系统
// 参数:
//   - config: 测试配置
//
// 返回:
//   - *TestRunner: 初始化完成的测试运行器
//   - error: 初始化过程中的错误（如果有）
func NewTestRunner(config TestConfig) (*TestRunner, error) {
	// 设置日志
	logx.MustSetup(logx.LogConf{
		Level: "error",
	})
	logx.DisableStat()

	logger := log.New(log.Writer(), "[TestRunner] ", log.LstdFlags|log.Lmsgprefix)

	// 创建文件加载器
	fileLoader := NewFileLoader([]string{
		filepath.Join(baseDir, "test/testdata/code_samples"),
		filepath.Join(baseDir, "test/testdata/doc_samples"),
	})

	// 加载测试文件
	logger.Printf("开始加载测试文件...")
	testFiles, err := fileLoader.LoadFiles()
	if err != nil {
		logger.Printf("加载测试文件失败: %v", err)
		return nil, fmt.Errorf("加载测试文件失败: %w", err)
	}
	logger.Printf("成功加载 %d 个测试文件", len(testFiles))

	// 加载查询配置
	logger.Printf("开始加载查询配置...")
	queries, err := loadQueries(filepath.Join(baseDir, "test/testdata/queries.json"))
	if err != nil {
		logger.Printf("加载查询配置失败: %v", err)
		return nil, fmt.Errorf("加载测试查询失败: %w", err)
	}
	logger.Printf("成功加载 %d 个查询配置", len(queries))

	// 验证查询配置的场景类型分布
	codeQueries := 0
	docQueries := 0
	for _, query := range queries {
		switch query.ScenarioType {
		case "code":
			codeQueries++
		case "doc":
			docQueries++
		}
	}
	logger.Printf("查询分布 - 代码场景: %d, 文档场景: %d", codeQueries, docQueries)

	return &TestRunner{
		config:     config,
		testFiles:  testFiles,
		queries:    queries,
		fileLoader: fileLoader,
		metrics:    &MetricsCalculator{},
		logger:     logger,
	}, nil
}

// RunEmbedderComparison 运行嵌入模型对比测试
func (tr *TestRunner) RunEmbedderComparison(ctx context.Context) (*TestResult, error) {
	tr.logger.Println("开始运行嵌入模型对比测试...")

	scenario := tr.config.Scenarios.EmbedderComparison
	result := &TestResult{
		TestName:        scenario.Name,
		TestDescription: scenario.Description,
		StartTime:       time.Now(),
		Results:         make(map[string]ScenarioResult),
	}

	// 获取向量数据库配置
	vectorStoreConfig := tr.getVectorStoreConfig(scenario.VectorStore)
	if vectorStoreConfig.Name == "" {
		return nil, fmt.Errorf("未找到向量数据库配置: %s", scenario.VectorStore)
	}

	// 测试每个嵌入模型
	totalEmbedders := len(scenario.Embedders)
	for i, embedderName := range scenario.Embedders {
		tr.logger.Printf("测试嵌入模型 [%d/%d]: %s", i+1, totalEmbedders, embedderName)

		embedderConfig := tr.getEmbedderConfig(embedderName)
		if embedderConfig.Model == "" {
			tr.logger.Printf("警告: 未找到嵌入模型配置: %s", embedderName)
			result.Results[embedderName] = ScenarioResult{
				Error: fmt.Errorf("未找到嵌入模型配置: %s", embedderName),
			}
			continue
		}

		scenarioResult, err := tr.runSingleTest(
			ctx, embedderConfig, tr.config.Rerank, vectorStoreConfig, scenario.TopK,
		)

		if err != nil {
			tr.logger.Printf("嵌入模型 %s 测试失败: %v", embedderName, err)
			scenarioResult = ScenarioResult{
				Error: err,
			}
		} else {
			tr.logger.Printf("嵌入模型 %s 测试完成", embedderName)
		}

		result.Results[embedderName] = scenarioResult
	}

	result.EndTime = time.Now()
	tr.logger.Printf("嵌入模型对比测试完成，耗时: %v", result.EndTime.Sub(result.StartTime))
	return result, nil
}

// RunVectorStoreComparison 运行向量数据库对比测试
func (tr *TestRunner) RunVectorStoreComparison(ctx context.Context) (*TestResult, error) {
	tr.logger.Println("开始运行向量数据库对比测试...")

	scenario := tr.config.Scenarios.VectorStoreComparison
	result := &TestResult{
		TestName:        scenario.Name,
		TestDescription: scenario.Description,
		StartTime:       time.Now(),
		Results:         make(map[string]ScenarioResult),
	}

	// 获取嵌入模型配置
	if len(scenario.Embedders) == 0 {
		return nil, fmt.Errorf("未配置嵌入模型")
	}
	embedderConfig := tr.getEmbedderConfig(scenario.Embedders[0])
	if embedderConfig.Model == "" {
		return nil, fmt.Errorf("未找到嵌入模型配置: %s", scenario.Embedders[0])
	}

	// 测试每个向量数据库
	totalVectorStores := len(scenario.VectorStores)
	for i, vectorStoreName := range scenario.VectorStores {
		tr.logger.Printf("测试向量数据库 [%d/%d]: %s", i+1, totalVectorStores, vectorStoreName)

		vectorStoreConfig := tr.getVectorStoreConfig(vectorStoreName)
		if vectorStoreConfig.Name == "" {
			tr.logger.Printf("警告: 未找到向量数据库配置: %s", vectorStoreName)
			result.Results[vectorStoreName] = ScenarioResult{
				Error: fmt.Errorf("未找到向量数据库配置: %s", vectorStoreName),
			}
			continue
		}

		scenarioResult, err := tr.runSingleTest(
			ctx, embedderConfig, tr.config.Rerank, vectorStoreConfig, scenario.TopK,
		)

		if err != nil {
			tr.logger.Printf("向量数据库 %s 测试失败: %v", vectorStoreName, err)
			scenarioResult = ScenarioResult{
				Error: err,
			}
		} else {
			tr.logger.Printf("向量数据库 %s 测试完成", vectorStoreName)
		}

		result.Results[vectorStoreName] = scenarioResult
	}

	result.EndTime = time.Now()
	tr.logger.Printf("向量数据库对比测试完成，耗时: %v", result.EndTime.Sub(result.StartTime))
	return result, nil
}

// runSingleTest 运行单个测试
func (tr *TestRunner) runSingleTest(
	ctx context.Context,
	embedderConfig EmbedderConf,
	rerankConfig RerankerConf,
	vectorStoreConfig VectorStoreConf,
	topK int,
) (ScenarioResult, error) {
	startTime := time.Now()
	tr.logger.Printf("开始单个测试 - 嵌入模型: %s, 向量数据库: %s", embedderConfig.Model, vectorStoreConfig.Name)

	// 转换配置格式
	embedderConf := config.EmbedderConf{
		Timeout:       embedderConfig.Timeout,
		MaxRetries:    embedderConfig.MaxRetries,
		BatchSize:     embedderConfig.BatchSize,
		Model:         embedderConfig.Model,
		APIKey:        embedderConfig.APIKey,
		APIBase:       embedderConfig.APIBase,
		StripNewLines: embedderConfig.StripNewLines,
	}

	// 创建嵌入器
	tr.logger.Printf("创建嵌入器: %s", embedderConfig.Model)
	embedder, err := vector.NewEmbedder(embedderConf)
	if err != nil {
		tr.logger.Printf("创建嵌入器失败: %v", err)
		return ScenarioResult{}, fmt.Errorf("创建嵌入器失败: %w", err)
	}

	rerankConf := config.RerankerConf{
		Timeout:    rerankConfig.Timeout,
		MaxRetries: rerankConfig.MaxRetries,
		Model:      rerankConfig.Model,
		APIKey:     rerankConfig.APIKey,
		APIBase:    rerankConfig.APIBase,
	}
	reranker := vector.NewReranker(rerankConf)

	// 创建向量存储
	tr.logger.Printf("创建向量存储: %s", vectorStoreConfig.Name)
	vectorStore, err := vector.NewVectorStore(
		config.VectorStoreConf{
			Timeout:    vectorStoreConfig.Timeout,
			MaxRetries: vectorStoreConfig.MaxRetries,
			Type:       vectorStoreConfig.Type,
			Embedder: config.EmbedderConf{
				Timeout:       embedderConf.Timeout,
				MaxRetries:    embedderConf.MaxRetries,
				BatchSize:     embedderConf.BatchSize,
				Model:         embedderConf.Model,
				APIKey:        embedderConf.APIKey,
				APIBase:       embedderConf.APIBase,
				StripNewLines: embedderConf.StripNewLines,
			},
			Reranker: rerankConf,
			Weaviate: config.WeaviateConf{
				Endpoint:     vectorStoreConfig.Endpoint,
				BatchSize:    vectorStoreConfig.BatchSize,
				Timeout:      vectorStoreConfig.Timeout,
				ClassName:    vectorStoreConfig.ClassName,
				MaxDocuments: vectorStoreConfig.MaxDocuments,
			},
			FetchSourceCode: false,
			StoreSourceCode: true,
		}, embedder, reranker)
	if err != nil {
		tr.logger.Printf("创建向量存储失败: %v", err)
		return ScenarioResult{}, fmt.Errorf("创建向量存储失败: %w", err)
	}
	defer func() {
		vectorStore.Close()
		tr.logger.Printf("向量存储已关闭")
	}()

	// 准备代码块
	tr.logger.Printf("准备代码块...")
	codeChunks, err := tr.prepareCodeChunksOptimized()
	if err != nil {
		tr.logger.Printf("准备代码块失败: %v", err)
		return ScenarioResult{}, fmt.Errorf("准备代码块失败: %w", err)
	}
	tr.logger.Printf("生成了 %d 个代码块", len(codeChunks))

	options := vector.Options{
		CodebaseId:   1,
		CodebasePath: "/test/codebase",
		CodebaseName: "test_codebase",
		ClientId:     uuid.New().String(),
	}

	// 存储代码块
	tr.logger.Printf("存储代码块到向量数据库...")
	if err := vectorStore.InsertCodeChunks(ctx, codeChunks, options); err != nil {
		tr.logger.Printf("存储代码块失败: %v", err)
		return ScenarioResult{}, fmt.Errorf("存储代码块失败: %w", err)
	}

	// 验证存储结果
	summary, err := vectorStore.GetIndexSummary(ctx, options.ClientId, options.CodebasePath)
	if err != nil {
		tr.logger.Printf("获取索引摘要失败: %v", err)
		return ScenarioResult{}, fmt.Errorf("获取索引摘要失败: %w", err)
	}
	if summary.TotalChunks != len(codeChunks) {
		err := fmt.Errorf("chunks in store is %d, want %d", summary.TotalChunks, len(codeChunks))
		tr.logger.Printf("代码块数量不匹配: %v", err)
		return ScenarioResult{}, err
	}
	tr.logger.Printf("成功存储 %d 个代码块", summary.TotalChunks)

	// 执行查询测试
	tr.logger.Printf("开始执行查询测试...")
	queryResults := make([]QueryResult, 0, len(tr.queries))
	totalQueries := len(tr.queries)

	for i, query := range tr.queries {
		tr.logger.Printf("执行查询 [%d/%d]: %s", i+1, totalQueries, query.Text[:min(50, len(query.Text))])

		queryStart := time.Now()
		items, err := vectorStore.Query(ctx, query.Text, topK, options)
		if err != nil {
			tr.logger.Printf("查询失败: %v", err)
			return ScenarioResult{}, fmt.Errorf("查询失败: %w", err)
		}

		queryTime := time.Since(queryStart)

		if len(items) == 0 {
			tr.logger.Printf("查询结果为空, query: %s", query.Text)
		} else {
			tr.logger.Printf("查询返回 %d 个结果", len(items))
		}

		// 计算评估指标
		metrics := tr.metrics.CalculateWithContent(query.Expected, query.ExpectedContents, items, queryTime)

		// 生成内容片段匹配结果
		contentMatches := tr.generateContentMatches(query.ExpectedContents, items)

		queryResults = append(queryResults, QueryResult{
			QueryID:           query.ID,
			Query:             query.Text,
			Metrics:           metrics,
			Retrieved:         items,
			Expected:          query.Expected,
			ExpectedContents:  query.ExpectedContents,
			ScenarioDimension: query.ScenarioDimension,
			ContentMatches:    contentMatches,
		})
	}

	// 计算平均指标
	avgMetrics := tr.metrics.CalculateAverage(queryResults)

	endTime := time.Now()
	tr.logger.Printf("单个测试完成，耗时: %v", endTime.Sub(startTime))

	return ScenarioResult{
		StartTime:      startTime,
		EndTime:        endTime,
		QueryResults:   queryResults,
		AverageMetrics: avgMetrics,
	}, nil
}

// prepareCodeChunksOptimized 优化后的代码块准备函数
func (tr *TestRunner) prepareCodeChunksOptimized() ([]*types.CodeChunk, error) {
	// 创建代码分割器
	splitter, err := embedding.NewCodeSplitter(embedding.SplitOptions{
		MaxTokensPerChunk:          1000,
		SlidingWindowOverlapTokens: 200,
		EnableMarkdownParsing:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("创建代码分割器失败: %w", err)
	}

	// 使用缓冲通道并发处理文件
	chunkChan := make(chan []*types.CodeChunk, len(tr.testFiles))
	errChan := make(chan error, len(tr.testFiles))

	// 并发处理文件
	for fileName, sourceFile := range tr.testFiles {
		go func(name string, file *types.SourceFile) {
			chunks, err := splitter.Split(file)
			if err != nil {
				errChan <- fmt.Errorf("split file %s err:%w", name, err)
				return
			}

			// 为每个chunk添加文件路径信息
			for _, chunk := range chunks {
				chunk.FilePath = name
			}

			chunkChan <- chunks
		}(fileName, sourceFile)
	}

	// 收集结果
	var allChunks []*types.CodeChunk
	completed := 0

	for completed < len(tr.testFiles) {
		select {
		case chunks := <-chunkChan:
			allChunks = append(allChunks, chunks...)
			completed++
		case err := <-errChan:
			return nil, err
		}
	}

	return allChunks, nil
}

// prepareCodeChunks 准备代码块，复用项目的CodeChunk结构（保留原函数以保持兼容性）
func (tr *TestRunner) prepareCodeChunks() []*types.CodeChunk {
	var chunks []*types.CodeChunk

	for fileName, sourceFile := range tr.testFiles {
		// 每个文件作为一个块
		lines := strings.Split(string(sourceFile.Content), "\n")
		chunk := &types.CodeChunk{
			CodebaseId:   sourceFile.CodebaseId,
			CodebasePath: sourceFile.CodebasePath,
			CodebaseName: sourceFile.CodebaseName,
			Language:     sourceFile.Language,
			Content:      sourceFile.Content,
			FilePath:     fileName,
			Range:        []int{0, 0, 0, len(lines), 0},
			TokenCount:   len(strings.Split(string(sourceFile.Content), " ")),
		}
		chunks = append(chunks, chunk)
	}

	return chunks
}

// extractFilePaths 提取检索结果的文件路径
func (tr *TestRunner) extractFilePaths(items []*types.SemanticFileItem) []string {
	paths := make([]string, len(items))
	for i, item := range items {
		paths[i] = item.FilePath
	}
	return paths
}

// getEmbedderConfig 获取嵌入模型配置
func (tr *TestRunner) getEmbedderConfig(name string) EmbedderConf {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	for _, embedder := range tr.config.Embedders {
		if embedder.Model == name {
			return embedder
		}
	}
	tr.logger.Printf("警告: 未找到嵌入模型配置: %s", name)
	return EmbedderConf{}
}

// getVectorStoreConfig 获取向量数据库配置
func (tr *TestRunner) getVectorStoreConfig(name string) VectorStoreConf {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	for _, vs := range tr.config.VectorStores {
		if vs.Name == name {
			return vs
		}
	}
	tr.logger.Printf("警告: 未找到向量数据库配置: %s", name)
	return VectorStoreConf{}
}

// generateContentMatches 生成内容片段匹配结果
func (tr *TestRunner) generateContentMatches(expectedContents []string, retrieved []*types.SemanticFileItem) []ContentMatch {
	if len(expectedContents) == 0 {
		return []ContentMatch{}
	}

	contentMatches := make([]ContentMatch, 0, len(expectedContents))

	for _, expectedContent := range expectedContents {
		match := ContentMatch{
			Content:        expectedContent,
			Matched:        false,
			MatchScore:     0.0,
			FoundInFiles:   []string{},
			MatchPositions: []MatchPosition{},
		}

		// 在检索结果中查找匹配
		for _, item := range retrieved {
			if tr.isContentMatchInItem(expectedContent, item, &match) {
				match.Matched = true
				match.FoundInFiles = append(match.FoundInFiles, item.FilePath)
			}
		}

		// 计算匹配分数
		if match.Matched {
			match.MatchScore = tr.calculateMatchScore(expectedContent, retrieved)
		}

		contentMatches = append(contentMatches, match)
	}

	return contentMatches
}

// isContentMatchInItem 检查期望内容是否在单个检索项中匹配
func (tr *TestRunner) isContentMatchInItem(expectedContent string, item *types.SemanticFileItem, match *ContentMatch) bool {
	expectedLower := strings.ToLower(expectedContent)
	contentLower := strings.ToLower(item.Content)

	// 检查精确匹配
	if strings.Contains(contentLower, expectedLower) {
		// 添加匹配位置信息
		lines := strings.Split(item.Content, "\n")
		for lineNum, line := range lines {
			if strings.Contains(strings.ToLower(line), expectedLower) {
				col := strings.Index(strings.ToLower(line), expectedLower)
				if col >= 0 {
					match.MatchPositions = append(match.MatchPositions, MatchPosition{
						FilePath: item.FilePath,
						Line:     lineNum + 1,
						Column:   col + 1,
						Context:  strings.TrimSpace(line),
					})
				}
			}
		}
		return true
	}

	// 检查词汇匹配
	expectedWords := strings.Fields(expectedLower)
	if len(expectedWords) == 0 {
		return false
	}

	contentWords := strings.Fields(contentLower)
	matchedWords := 0

	for _, expectedWord := range expectedWords {
		for _, contentWord := range contentWords {
			if strings.Contains(contentWord, expectedWord) || expectedWord == contentWord {
				matchedWords++
				break
			}
		}
	}

	// 如果超过60%的词汇匹配，则认为内容匹配
	if float64(matchedWords)/float64(len(expectedWords)) >= 0.6 {
		// 添加匹配位置信息
		lines := strings.Split(item.Content, "\n")
		for lineNum, line := range lines {
			lineLower := strings.ToLower(line)
			for _, expectedWord := range expectedWords {
				if strings.Contains(lineLower, expectedWord) {
					col := strings.Index(lineLower, expectedWord)
					if col >= 0 {
						match.MatchPositions = append(match.MatchPositions, MatchPosition{
							FilePath: item.FilePath,
							Line:     lineNum + 1,
							Column:   col + 1,
							Context:  strings.TrimSpace(line),
						})
					}
				}
			}
		}
		return true
	}

	return false
}

// calculateMatchScore 计算匹配分数
func (tr *TestRunner) calculateMatchScore(expectedContent string, retrieved []*types.SemanticFileItem) float64 {
	expectedLower := strings.ToLower(expectedContent)
	expectedWords := strings.Fields(expectedLower)
	if len(expectedWords) == 0 {
		return 0.0
	}

	totalScore := 0.0
	maxScore := 0.0

	for _, item := range retrieved {
		contentLower := strings.ToLower(item.Content)
		contentWords := strings.Fields(contentLower)

		itemScore := 0.0
		itemMaxScore := 0.0

		for _, expectedWord := range expectedWords {
			itemMaxScore += 1.0
			for _, contentWord := range contentWords {
				if strings.Contains(contentWord, expectedWord) || expectedWord == contentWord {
					itemScore += 1.0
					break
				}
			}
		}

		if itemMaxScore > 0 {
			totalScore += itemScore / itemMaxScore
			maxScore += 1.0
		}
	}

	if maxScore == 0 {
		return 0.0
	}

	return totalScore / maxScore
}

// loadQueries 加载查询配置
// 支持JSON和YAML格式的查询配置文件
// 自动根据文件扩展名选择合适的解析器
// 参数:
//   - path: 查询配置文件路径
//
// 返回:
//   - []QueryConfig: 查询配置列表
//   - error: 加载或解析过程中的错误
func loadQueries(path string) ([]QueryConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取查询配置文件失败: %w", err)
	}

	// 根据文件扩展名选择解析方式
	ext := filepath.Ext(path)
	var queries []QueryConfig

	switch ext {
	case ".json":
		var config struct {
			Queries []QueryConfig `json:"queries"`
		}
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("解析JSON查询配置文件失败: %w", err)
		}
		queries = config.Queries

	case ".yaml", ".yml":
		var config struct {
			Queries []QueryConfig `yaml:"queries"`
		}
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("解析YAML查询配置文件失败: %w", err)
		}
		queries = config.Queries

	default:
		return nil, fmt.Errorf("不支持的查询配置文件格式: %s", ext)
	}

	if len(queries) == 0 {
		return nil, fmt.Errorf("查询配置文件中没有查询配置")
	}

	// 验证查询配置的完整性
	for i, query := range queries {
		if query.ID == "" {
			return nil, fmt.Errorf("查询配置[%d]缺少ID字段", i)
		}
		if query.Text == "" {
			return nil, fmt.Errorf("查询配置[%d]缺少Text字段", i)
		}
		if query.ScenarioType == "" {
			return nil, fmt.Errorf("查询配置[%d]缺少ScenarioType字段", i)
		}
		if query.ScenarioType != "code" && query.ScenarioType != "doc" {
			return nil, fmt.Errorf("查询配置[%d]的ScenarioType字段值无效: %s", i, query.ScenarioType)
		}
	}

	return queries, nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
