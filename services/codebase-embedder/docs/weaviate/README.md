# Weaviate 向量存储文档

## 概述

Weaviate 是一个开源的向量搜索引擎，本系统使用 Weaviate 作为向量存储后端，用于存储和检索代码库的嵌入向量。通过 Weaviate，系统能够实现高效的语义搜索和相似性查询。

## 架构设计

### 核心组件

1. **weaviateWrapper**: Weaviate 的封装实现，实现了 `Store` 接口
2. **多租户支持**: 使用 MD5 哈希生成租户名称，实现不同代码库的数据隔离
3. **Schema 设计**: 定义了代码块存储的结构和属性

### 数据模型

Weaviate 中存储的每个代码块包含以下属性：

- `codebase_id`: 代码库ID
- `codebase_name`: 代码库名称
- `sync_id`: 同步ID
- `codebase_path`: 代码库路径
- `file_path`: 文件路径
- `language`: 编程语言
- `range`: 代码块范围（起始行和结束行）
- `token_count`: Token数量
- `content`: 代码块内容

## 配置

### 基本配置

```go
type VectorStoreConf struct {
    Type     string
    Weaviate struct {
        Endpoint   string
        APIKey     string
        ClassName  string
        Timeout    time.Duration
        MaxDocuments int
    }
    FetchSourceCode bool
    BaseURL        string
    Embedder       EmbedderConf
}
```

### 配置示例

```yaml
vector_store:
  type: "weaviate"
  weaviate:
    endpoint: "localhost:8080"
    api_key: ""  # 可选，如果Weaviate需要认证
    class_name: "CodeChunk"
    timeout: 30s
    max_documents: 100
  fetch_source_code: true
  base_url: "http://localhost:8081/api/code-snippet"
```

## 主要操作

### 1. 初始化 Weaviate 客户端

```go
// 创建基本的 Weaviate 客户端
store, err := New(cfg, embedder, reranker)
if err != nil {
    log.Fatalf("Failed to create Weaviate client: %v", err)
}

// 创建带有状态管理器的 Weaviate 客户端
storeWithStatus, err := NewWithStatusManager(cfg, embedder, reranker, statusManager, requestId)
```

### 2. 插入代码块

```go
// 插入新的代码块
docs := []*types.CodeChunk{
    {
        CodebaseId:   1,
        CodebasePath: "/path/to/codebase",
        CodebaseName: "example-project",
        FilePath:     "src/main.go",
        Content:      []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}"),
        Language:     "go",
        Range:        []int{1, 5},
        TokenCount:   20,
    },
}

err := store.InsertCodeChunks(ctx, docs, Options{
    CodebaseId:    1,
    CodebasePath:  "/path/to/codebase",
    CodebaseName:  "example-project",
    ClientId:      "client-123",
    Authorization: "Bearer token",
})
```

### 3. 更新代码块

```go
// 更新代码块（先删除再插入）
err := store.UpsertCodeChunks(ctx, docs, Options{
    CodebaseId:   1,
    CodebasePath: "/path/to/codebase",
    CodebaseName: "example-project",
})

// 仅更新文件路径
updates := []*types.CodeChunkPathUpdate{
    {
        CodebaseId:   1,
        OldFilePath:  "src/old_name.go",
        NewFilePath:  "src/new_name.go",
    },
}

err := store.UpdateCodeChunksPaths(ctx, updates, Options{
    CodebaseId:   1,
    CodebasePath: "/path/to/codebase",
})
```

### 4. 删除代码块

```go
// 删除特定代码块
chunks := []*types.CodeChunk{
    {
        CodebaseId: 1,
        FilePath:   "src/file_to_delete.go",
    },
}

err := store.DeleteCodeChunks(ctx, chunks, Options{
    CodebaseId:   1,
    CodebasePath: "/path/to/codebase",
})

// 删除整个代码库
err := store.DeleteByCodebase(ctx, 1, "/path/to/codebase")
```

### 5. 语义搜索

```go
// 执行相似性搜索
results, err := store.SimilaritySearch(ctx, "如何实现快速排序", 10, Options{
    CodebaseId:   1,
    CodebasePath: "/path/to/codebase",
    ClientId:     "client-123",
    Authorization: "Bearer token",
})

// 执行完整查询（包括重排序）
results, err := store.Query(ctx, "如何实现快速排序", 5, Options{
    CodebaseId:   1,
    CodebasePath: "/path/to/codebase",
    ClientId:     "client-123",
    Authorization: "Bearer token",
})
```

### 6. 获取索引摘要

```go
// 获取代码库的索引摘要
summary, err := store.GetIndexSummary(ctx, 1, "/path/to/codebase")
if err != nil {
    log.Fatalf("Failed to get index summary: %v", err)
}

fmt.Printf("总文件数: %d\n", summary.TotalFiles)
fmt.Printf("总代码块数: %d\n", summary.TotalChunks)
```

### 7. 获取代码库记录

```go
// 获取代码库中的所有记录
records, err := store.GetCodebaseRecords(ctx, 1, "/path/to/codebase")
if err != nil {
    log.Fatalf("Failed to get codebase records: %v", err)
}

for _, record := range records {
    fmt.Printf("文件: %s, 语言: %s, Token数: %d\n", 
        record.FilePath, record.Language, record.TokenCount)
}
```

## 多租户支持

系统使用 MD5 哈希将代码库路径转换为租户名称，实现数据隔离：

```go
func (r *weaviateWrapper) generateTenantName(codebasePath string) (string, error) {
    if codebasePath == types.EmptyString {
        return types.EmptyString, ErrInvalidCodebasePath
    }
    hash := md5.Sum([]byte(codebasePath))
    tenantName := hex.EncodeToString(hash[:])
    return tenantName, nil
}
```

例如：
- 代码库路径 `/path/to/project1` → 租户名 `a1b2c3d4e5f6...`
- 代码库路径 `/path/to/project2` → 租户名 `f9e8d7c6b5a4...`

## 代码片段获取

系统支持从原始文件中获取代码片段内容：

```go
// 批量获取代码片段
snippets := []CodeSnippetRequest{
    {
        FilePath:  "src/main.go",
        StartLine: 1,
        EndLine:   10,
    },
}

contentMap, err := fetchCodeContentsBatch(ctx, cfg, clientId, codebasePath, snippets, authorization)
```

## 错误处理

系统提供了完善的错误处理机制：

```go
// 检查批量操作错误
if err := CheckBatchErrors(resp); err != nil {
    return fmt.Errorf("failed to send batch to Weaviate: %w", err)
}

// 检查批量删除错误
if err := CheckBatchDeleteErrors(do); err != nil {
    return fmt.Errorf("failed to delete objects: %w", err)
}
```

## 性能优化

### 批量操作

系统支持批量插入、更新和删除操作，以提高性能：

```go
// 批量插入
objs := make([]*models.Object, len(chunks))
for i, c := range chunks {
    objs[i] = &models.Object{
        ID:     strfmt.UUID(uuid.New().String()),
        Class:  r.className,
        Tenant: tenantName,
        Vector: c.Embedding,
        Properties: map[string]any{
            MetadataFilePath:     c.FilePath,
            MetadataLanguage:     c.Language,
            // ... 其他属性
        },
    }
}

resp, err := r.client.Batch().ObjectsBatcher().WithObjects(objs...).Do(ctx)
```

### 分页查询

对于大量数据的查询，系统使用分页机制：

```go
// 分页获取代码库记录
limit := 1000
offset := 0
for {
    res, err := r.client.GraphQL().Get().
        WithClassName(r.className).
        WithFields(fields...).
        WithWhere(codebaseFilter).
        WithLimit(limit).
        WithOffset(offset).
        WithTenant(tenantName).
        Do(ctx)
    
    // 处理结果...
    
    if len(records) < limit {
        break
    }
    offset += limit
}
```

## 监控和日志

系统提供了详细的日志记录和监控功能：

```go
// 记录操作耗时
start := time.Now()
// 执行操作...
tracer.WithTrace(ctx).Infof("operation completed, cost %d ms", time.Since(start).Milliseconds())

// 调试日志
fmt.Printf("[DEBUG] GetCodebaseRecords - 开始执行，codebaseId: %d, codebasePath: %s\n", 
    codebaseId, codebasePath)
```

## 最佳实践

1. **连接管理**: 使用连接池管理 Weaviate 连接
2. **错误重试**: 实现适当的重试机制处理暂时性错误
3. **批量操作**: 尽量使用批量操作减少网络开销
4. **内存管理**: 及时释放不再使用的资源
5. **监控告警**: 监控关键指标如查询延迟、错误率等

## 故障排除

### 常见问题

1. **连接失败**
   - 检查 Weaviate 服务是否正常运行
   - 验证网络连接和防火墙设置
   - 确认认证配置是否正确

2. **查询结果为空**
   - 检查租户名称是否正确生成
   - 确认过滤器条件是否合适
   - 验证数据是否已正确插入

3. **性能问题**
   - 检查查询复杂度和返回结果数量
   - 优化索引配置
   - 考虑使用缓存机制

### 调试技巧

启用调试日志获取详细信息：

```go
// 检查 Weaviate 连接状态
live, err := r.client.Misc().LiveChecker().Do(ctx)
if err != nil {
    fmt.Printf("[DEBUG] Weaviate 连接检查失败: %v\n", err)
} else {
    fmt.Printf("[DEBUG] Weaviate 连接状态: %v\n", live)
}