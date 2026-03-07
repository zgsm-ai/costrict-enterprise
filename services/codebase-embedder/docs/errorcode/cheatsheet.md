# Codebase Embedder 错误码速查表

本文档提供了项目中所有错误码的快速参考，方便开发人员在编码时快速查找和使用。

## HTTP 状态码错误

### 400 - 参数错误
```go
// 创建参数错误
err := response.NewParamError("invalid request parameter")
```

### 401 - 认证错误
```go
// 创建认证错误
err := response.NewAuthError("authentication failed")
```

### 403 - 权限错误
```go
// 创建权限错误
err := response.NewPermissionError("access denied")
```

### 429 - 限流错误
```go
// 创建限流错误
err := response.NewRateLimitError("rate limit exceeded")
```

## 系统响应码

### 0 - 成功
```go
// 成功响应
response.Ok(w)
response.Json(w, data)
```

### -1 - 通用错误
```go
// 错误响应
response.Error(w, err)
```

## 数据库错误

### 数据库插入失败
```go
// 数据库插入失败
if err := db.Insert(data); err != nil {
    return errs.InsertDatabaseFailed
}
```

## 参数错误

### 无效参数错误
```go
// 无效参数错误
err := errs.NewInvalidParamErr("userId", "invalid-uuid")
// 输出: "invalid request params: userId invalid-uuid"
```

### 记录未找到错误
```go
// 记录未找到错误
err := errs.NewRecordNotFoundErr("user", "id=123")
// 输出: "user not found by id=123"
```

### 缺少必需参数错误
```go
// 缺少必需参数错误
err := errs.NewMissingParamError("token")
// 输出: "missing required param: token"
```

## 任务错误

### 文件未找到
```go
// 文件未找到
if _, err := os.Stat(filepath); os.IsNotExist(err) {
    return errs.FileNotFound
}
```

### 读取超时
```go
// 读取超时
if err := readWithTimeout(); err != nil {
    return errs.ReadTimeout
}
```

### 运行超时
```go
// 运行超时
if err := executeWithTimeout(); err != nil {
    return errs.RunTimeout
}
```

## 向量存储错误

### 静态错误
```go
// 无效的代码库路径
if !isValidCodebasePath(path) {
    return vector.ErrInvalidCodebasePath
}

// 无效的客户端ID
if !isValidClientId(clientId) {
    return vector.ErrInvalidClientId
}

// 响应为空
if resp == nil {
    return vector.ErrEmptyResponse
}

// 响应无效
if !isValidResponse(resp) {
    return vector.ErrInvalidResponse
}
```

### 批量操作错误检查
```go
// 检查批量操作错误
if err := vector.CheckBatchErrors(responses); err != nil {
    return err
}
```

### GraphQL 响应错误检查
```go
// 检查 GraphQL 响应错误
if err := vector.CheckGraphQLResponseError(graphqlResp); err != nil {
    return err
}
```

### 批量删除错误检查
```go
// 检查批量删除错误
if err := vector.CheckBatchDeleteErrors(deleteResp); err != nil {
    return err
}
```

## 错误处理模式

### 基本错误处理模式
```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 1. 参数验证
    if err := validateParams(r); err != nil {
        response.Error(w, response.NewParamError(err.Error()))
        return
    }
    
    // 2. 业务逻辑处理
    result, err := processBusinessLogic(r)
    if err != nil {
        // 根据错误类型返回相应的错误码
        switch err {
        case errs.FileNotFound:
            response.Error(w, response.NewError(404, "file not found"))
        case errs.InsertDatabaseFailed:
            response.Error(w, response.NewError(500, "database operation failed"))
        default:
            response.Error(w, response.NewError(500, "internal server error"))
        }
        return
    }
    
    // 3. 成功响应
    response.Json(w, result)
}
```

### 错误包装模式
```go
func processUserData(userId string) (*User, error) {
    // 验证用户ID
    if !isValidUUID(userId) {
        return nil, errs.NewInvalidParamErr("userId", userId)
    }
    
    // 查询用户
    user, err := db.GetUser(userId)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errs.NewRecordNotFoundErr("user", fmt.Sprintf("id=%s", userId))
        }
        return nil, errs.InsertDatabaseFailed
    }
    
    return user, nil
}
```

### 错误恢复模式
```go
func retryOperation(maxRetries int, operation func() error) error {
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        // 对于可重试的错误，等待后重试
        if isRetryableError(err) {
            time.Sleep(time.Second * time.Duration(i+1))
            continue
        }
        
        return err
    }
    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

## 错误码映射表

| 错误类型 | 错误码 | HTTP状态码 | 说明 |
|----------|--------|------------|------|
| 成功 | 0 | 200 | 请求成功处理 |
| 参数错误 | 400 | 400 | 请求参数错误 |
| 认证错误 | 401 | 401 | 认证失败 |
| 权限错误 | 403 | 403 | 权限不足 |
| 限流错误 | 429 | 429 | 请求频率超限 |
| 文件未找到 | 404 | 404 | 文件或目录不存在 |
| 数据库错误 | 500 | 500 | 数据库操作失败 |
| 任务超时 | 504 | 504 | 任务执行超时 |
| 通用错误 | -1 | 500 | 未分类的通用错误 |

## 常见错误场景

### 1. API 参数验证
```go
func validateAPIParams(req *APIRequest) error {
    if req.UserID == "" {
        return errs.NewMissingParamError("userId")
    }
    
    if !isValidUUID(req.UserID) {
        return errs.NewInvalidParamErr("userId", req.UserID)
    }
    
    return nil
}
```

### 2. 数据库操作
```go
func createUser(user *User) error {
    if err := db.Insert(user); err != nil {
        logx.Errorf("failed to create user: %v", err)
        return errs.InsertDatabaseFailed
    }
    return nil
}
```

### 3. 文件操作
```go
func readFileContent(filepath string) ([]byte, error) {
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        return nil, errs.FileNotFound
    }
    
    data, err := os.ReadFile(filepath)
    if err != nil {
        return nil, errs.ReadTimeout
    }
    
    return data, nil
}
```

### 4. 向量存储操作
```go
func storeEmbeddings(embeddings []Embedding) error {
    // 验证代码库路径
    if !isValidCodebasePath(codebasePath) {
        return vector.ErrInvalidCodebasePath
    }
    
    // 批量存储
    responses, err := vectorClient.BatchStore(embeddings)
    if err != nil {
        return err
    }
    
    // 检查批量操作错误
    if err := vector.CheckBatchErrors(responses); err != nil {
        return err
    }
    
    return nil
}
```

## 最佳实践

### 1. 错误信息规范
- 使用清晰的错误描述
- 包含足够的上下文信息
- 避免暴露敏感信息

### 2. 错误处理原则
- 及时处理错误，不要忽略
- 根据错误类型采取不同的处理策略
- 记录错误日志便于排查问题

### 3. 错误恢复策略
- 对于临时性错误，实现重试机制
- 对于永久性错误，快速失败并返回明确的错误信息
- 提供错误恢复的指导信息

### 4. 错误监控
- 记录错误发生的频率和模式
- 设置错误告警阈值
- 定期分析错误日志，优化系统稳定性