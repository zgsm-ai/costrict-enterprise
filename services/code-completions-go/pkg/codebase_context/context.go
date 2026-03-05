package codebase_context

import (
	"code-completion/pkg/config"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// ContextClient 上下文客户端
type ContextClient struct {
	apiClient *APIClient
}

/**
 * Create new context client for codebase operations
 * @returns {ContextClient} Returns initialized context client instance
 * @description
 * - Creates a new context client with API client
 * - Initializes underlying HTTP client for API communication
 * - Used for searching codebase context and retrieving related information
 * @example
 * client := NewContextClient()
 * result := client.RequestContext(ctx, "client-id", "/path", "file.go", []string{"code"}, []string{"query"}, headers)
 */
func NewContextClient() *ContextClient {
	return &ContextClient{
		apiClient: NewAPIClient(),
	}
}

// SearchResult 搜索结果
type SearchResult struct {
	DefinitionResults []*ResponseData
	SemanticResults   []*ResponseData
	RelationResults   []*ResponseData
}

/**
 * Asynchronously search for code definitions
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {string} clientID - Client identifier for the request
 * @param {string} codebasePath - Path to the codebase being searched
 * @param {string} filePath - Path to the file containing the code snippet
 * @param {string} codeSnippet - Code snippet to search for definitions
 * @param {http.Header} headers - HTTP headers for the request
 * @param {sync.WaitGroup} wg - Wait group for synchronization
 * @param {[]*ResponseData} results - Slice to store search results
 * @param {int} idx - Index in results slice to store the result
 * @description
 * - Performs asynchronous definition search for code snippet
 * - Updates results slice at specified index with search result
 * - Handles errors by storing error result in results slice
 * - Signals completion via done() on wait group
 * @example
 * wg.Add(1)
 * go client.searchDefinitionAsync(ctx, "client-id", "/codebase", "file.go", "func test()", headers, &wg, results, 0)
 */
func (c *ContextClient) searchDefinitionAsync(ctx context.Context, clientID, codebasePath, filePath, codeSnippet string,
	headers http.Header, wg *sync.WaitGroup, results []*ResponseData, idx int) {
	defer wg.Done()

	data, err := c.searchDefinition(ctx, clientID, codebasePath, filePath, codeSnippet, headers)
	if err != nil {
		results[idx] = data
	}
}

/**
 * Asynchronously search for code relations
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {string} clientID - Client identifier for the request
 * @param {string} codebasePath - Path to the codebase being searched
 * @param {string} filePath - Path to the file containing the code snippet
 * @param {string} codeSnippet - Code snippet to search for relations
 * @param {http.Header} headers - HTTP headers for the request
 * @param {sync.WaitGroup} wg - Wait group for synchronization
 * @param {[]*ResponseData} results - Slice to store search results
 * @param {int} idx - Index in results slice to store the result
 * @description
 * - Performs asynchronous relation search for code snippet
 * - Updates results slice at specified index with search result
 * - Handles errors by storing error result in results slice
 * - Signals completion via done() on wait group
 * @example
 * wg.Add(1)
 * go client.searchRelationAsync(ctx, "client-id", "/codebase", "file.go", "func test()", headers, &wg, results, 1)
 */
func (c *ContextClient) searchRelationAsync(ctx context.Context, clientID, codebasePath, filePath, codeSnippet string,
	headers http.Header, wg *sync.WaitGroup, results []*ResponseData, idx int) {
	defer wg.Done()

	data, err := c.searchRelation(ctx, clientID, codebasePath, filePath, codeSnippet, headers)
	if err != nil {
		results[idx] = data
	}
}

/**
 * Asynchronously search for semantic code matches
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {string} clientID - Client identifier for the request
 * @param {string} codebasePath - Path to the codebase being searched
 * @param {string} query - Semantic query string to search for
 * @param {http.Header} headers - HTTP headers for the request
 * @param {sync.WaitGroup} wg - Wait group for synchronization
 * @param {[]*ResponseData} results - Slice to store search results
 * @param {int} idx - Index in results slice to store the result
 * @description
 * - Performs asynchronous semantic search for code
 * - Updates results slice at specified index with search result
 * - Handles errors by storing error result in results slice
 * - Signals completion via done() on wait group
 * @example
 * wg.Add(1)
 * go client.searchSemanticAsync(ctx, "client-id", "/codebase", "database query", headers, &wg, results, 2)
 */
func (c *ContextClient) searchSemanticAsync(ctx context.Context, clientID, codebasePath, query string, headers http.Header,
	wg *sync.WaitGroup, results []*ResponseData, idx int) {
	defer wg.Done()

	data, err := c.searchSemantic(ctx, clientID, codebasePath, query, headers)
	if err != nil {
		results[idx] = data
	}
}

/**
 * Request context information from multiple sources
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {string} clientID - Client identifier for the request
 * @param {string} codebasePath - Path to the codebase being searched
 * @param {string} filePath - Path to the file being analyzed
 * @param {[]string} codeSnippets - Array of code snippets for definition and relation search
 * @param {[]string} queries - Array of semantic queries for semantic search
 * @param {http.Header} headers - HTTP headers for the requests
 * @returns {SearchResult} Returns search results containing definition, semantic and relation results
 * @description
 * - Performs parallel searches for definitions, relations and semantic matches
 * - Uses goroutines for concurrent execution of different search types
 * - Sets up timeout context based on configuration
 * - Returns partial results if context timeout occurs
 * - Respects configuration flags for enabling/disabling specific search types
 * @example
 * result := client.RequestContext(ctx, "client-id", "/codebase", "file.go",
 *     []string{"func test()"}, []string{"database query"}, headers)
 */
func (c *ContextClient) RequestContext(ctx context.Context, clientID, codebasePath, filePath string,
	codeSnippets []string, queries []string, headers http.Header) *SearchResult {
	if clientID == "" || codebasePath == "" || filePath == "" {
		return &SearchResult{}
	}

	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(ctx, config.Context.TotalTimeout)
	defer cancel()

	var wg sync.WaitGroup
	// 初始化结果数组
	definitionResults := make([]*ResponseData, len(codeSnippets))
	relationResults := make([]*ResponseData, len(codeSnippets))
	semanticResults := make([]*ResponseData, len(queries))

	// 定义检索
	if len(codeSnippets) > 0 && !config.Context.Definition.Disabled {
		for i, codeSnippet := range codeSnippets {
			if codeSnippet == "" {
				continue
			}
			wg.Add(1)
			go c.searchDefinitionAsync(ctx, clientID, codebasePath, filePath, codeSnippet, headers, &wg, definitionResults, i)
		}
	}
	// 调用链检索
	if len(codeSnippets) > 0 && !config.Context.Relation.Disabled {
		for i, codeSnippet := range codeSnippets {
			if codeSnippet == "" {
				continue
			}
			wg.Add(1)
			go c.searchRelationAsync(ctx, clientID, codebasePath, filePath, codeSnippet, headers, &wg, relationResults, i)
		}
	}

	// 语义检索
	if len(queries) > 0 && !config.Context.Semantic.Disabled {
		for i, query := range queries {
			if query == "" {
				continue
			}
			wg.Add(1)
			go c.searchSemanticAsync(ctx, clientID, codebasePath, query, headers, &wg, semanticResults, i)
		}
	}

	// 等待所有请求完成或上下文取消
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待完成或上下文取消
	select {
	case <-done: // 所有请求完成
	case <-ctx.Done(): // 上下文取消，直接返回已收集的结果
		zap.L().Warn("Context timeout, returning partial results", zap.Error(ctx.Err()))
	}
	return &SearchResult{
		DefinitionResults: definitionResults,
		SemanticResults:   semanticResults,
		RelationResults:   relationResults,
	}
}

// 获取上下文信息
func (c *ContextClient) GetContext(ctx context.Context, clientID, projectPath, filePath, prefix, suffix, importContent string, headers http.Header) string {
	if clientID == "" || projectPath == "" || filePath == "" || (prefix == "" && suffix == "") {
		return ""
	}

	// 构建完整文件路径
	fullFilePath := filepath.Join(projectPath, filePath)

	// Windows路径处理
	if len(projectPath) > 1 && projectPath[1:3] == ":\\" {
		fullFilePath = strings.ReplaceAll(fullFilePath, "/", "\\")
	}

	// 获取语义搜索内容（前缀最后几行）
	semanticSearchContent := rSliceAfterNthInstance(prefix, "\n", 4)

	// 定义检索代码片段
	definitionCodeSnaps := []string{
		fmt.Sprintf("%s%s%s", importContent, prefix, suffix),
	}

	searchResult := c.RequestContext(ctx, clientID, projectPath, fullFilePath,
		definitionCodeSnaps, []string{semanticSearchContent}, headers)

	// 解析语义检索结果
	semanticCodes := parseSemantic(searchResult.SemanticResults)

	// 解析定义检索结果
	defCodes := parseDefinition(searchResult.DefinitionResults)

	// 解析关系检索结果
	relationCodes := parseRelation(searchResult.RelationResults)

	var allCodes []string

	// 合并定义检索结果
	for _, item := range defCodes {
		allCodes = append(allCodes, item.FilePath, item.Content)
	}

	// 合并语义检索结果
	for _, item := range semanticCodes {
		allCodes = append(allCodes, item.FilePath, item.Content)
	}

	// 合并关系检索结果
	for _, item := range relationCodes {
		allCodes = append(allCodes, item.FilePath, item.Content)
	}

	// 合并所有结果
	semanticResult := strings.Join(allCodes, "\n")

	// 添加注释
	return getComment(fullFilePath, semanticResult)
}

// 搜索代码定义
func (c *ContextClient) searchDefinition(ctx context.Context, clientID, codebasePath, filePath, codeSnippet string, headers http.Header) (*ResponseData, error) {
	params := RequestParam{
		ClientID:     clientID,
		CodebasePath: codebasePath,
		FilePath:     filePath,
		CodeSnippet:  codeSnippet,
	}

	return c.apiClient.DoRequest(ctx, config.Context.Definition.Url, params, headers, "GET")
}

// 语义搜索
func (c *ContextClient) searchSemantic(ctx context.Context, clientID, codebasePath, query string, headers http.Header) (*ResponseData, error) {
	params := RequestParam{
		ClientID:       clientID,
		CodebasePath:   codebasePath,
		Query:          query,
		TopK:           config.Context.Semantic.TopK,
		ScoreThreshold: config.Context.Semantic.ScoreThreshold,
	}

	return c.apiClient.DoRequest(ctx, config.Context.Semantic.Url, params, headers, "POST")
}

// 关系检索
func (c *ContextClient) searchRelation(ctx context.Context, clientID, codebasePath, filePath, codeSnippet string, headers http.Header) (*ResponseData, error) {
	params := RequestParam{
		ClientID:       clientID,
		CodebasePath:   codebasePath,
		FilePath:       filePath,
		CodeSnippet:    codeSnippet,
		MaxLayer:       config.Context.Relation.Layer,
		IncludeContent: config.Context.Relation.IncludeContent,
	}

	return c.apiClient.DoRequest(ctx, config.Context.Relation.Url, params, headers, "GET")
}
