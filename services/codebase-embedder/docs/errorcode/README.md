# Codebase Embedder 错误码文档

本文档详细描述了 Codebase Embedder 项目中使用的所有错误码和错误类型。

## 目录

- [错误码分类](#错误码分类)
- [HTTP 状态码错误](#http-状态码错误)
- [系统响应码](#系统响应码)
- [数据库错误](#数据库错误)
- [参数错误](#参数错误)
- [任务错误](#任务错误)
- [向量存储错误](#向量存储错误)
- [错误处理最佳实践](#错误处理最佳实践)

## 错误码分类

项目中的错误码系统分为以下几个主要类别：

1. **HTTP 状态码错误** - 基于 HTTP 标准状态码的错误
2. **系统响应码** - 应用层级的响应状态码
3. **数据库错误** - 数据库操作相关的错误
4. **参数错误** - 请求参数验证相关的错误
5. **任务错误** - 任务执行过程中的错误
6. **向量存储错误** - 向量数据库操作相关的错误

## HTTP 状态码错误

这些错误码基于 HTTP 标准状态码，用于表示客户端请求的状态。

| 状态码 | 错误类型 | 说明 | 使用场景 |
|--------|----------|------|----------|
| 400 | Bad Request | 参数错误 | 请求参数格式错误、缺少必需参数等 |
| 401 | Unauthorized | 认证错误 | 用户未认证或认证失败 |
| 403 | Forbidden | 权限错误 | 用户无权限访问资源 |
| 429 | Too Many Requests | 限流错误 | 请求频率超过限制 |

### 代码示例

```go
// 创建参数错误
err := response.NewParamError("invalid request parameter")

// 创建认证错误
err := response.NewAuthError("authentication failed")

// 创建权限错误
err := response.NewPermissionError("access denied")

// 创建限流错误
err := response.NewRateLimitError("rate limit exceeded")
```

## 系统响应码

应用层级的响应状态码，用于表示业务逻辑的执行结果。

| 状态码 | 常量名 | 说明 | 使用场景 |
|--------|--------|------|----------|
| 0 | CodeOK | 成功 | 请求成功处理 |
| -1 | CodeError | 通用错误 | 未分类的通用错误 |

### 代码示例

```go
// 成功响应
response.Ok(w)

// 错误响应
response.Error(w, err)
```

## 数据库错误

数据库操作相关的错误定义。

| 错误名称 | 说明 | 使用场景 |
|----------|------|----------|
| InsertDatabaseFailed | 数据库插入失败 | 向数据库插入数据时发生错误 |

### 代码示例

```go
// 数据库插入失败
if err := db.Insert(data); err != nil {
    return errs.InsertDatabaseFailed
}
```

## 参数错误

请求参数验证相关的错误，提供动态错误信息生成。

| 错误函数 | 说明 | 参数 |
|----------|------|------|
| NewInvalidParamErr | 无效参数错误 | name: 参数名, value: 参数值 |
| NewRecordNotFoundErr | 记录未找到错误 | name: 记录类型, value: 查询条件 |
| NewMissingParamError | 缺少必需参数错误 | name: 缺失的参数名 |

### 代码示例

```go
// 无效参数错误
err := errs.NewInvalidParamErr("userId", "invalid-uuid")

// 记录未找到错误
err := errs.NewRecordNotFoundErr("user", "id=123")

// 缺少必需参数错误
err := errs.NewMissingParamError("token")
```

## 任务错误

任务执行过程中的错误定义。

| 错误名称 | 说明 | 使用场景 |
|----------|------|----------|
| FileNotFound | 文件或目录未找到 | 文件系统操作时文件不存在 |
| ReadTimeout | 读取超时 | 读取文件或数据时超时 |
| RunTimeout | 运行超时 | 任务执行超时 |

### 代码示例

```go
// 文件未找到
if _, err := os.Stat(filepath); os.IsNotExist(err) {
    return errs.FileNotFound
}

// 读取超时
if err := readWithTimeout(); err != nil {
    return errs.ReadTimeout
}

// 运行超时
if err := executeWithTimeout(); err != nil {
    return errs.RunTimeout
}
```

## 向量存储错误

向量数据库操作相关的错误定义和错误检查函数。

### 静态错误

| 错误名称 | 说明 | 使用场景 |
|----------|------|----------|
| ErrInvalidCodebasePath | 无效的代码库路径 | 代码库路径格式错误或不存在 |
| ErrInvalidClientId | 无效的客户端ID | 客户端ID格式错误或不存在 |
| ErrEmptyResponse | 响应为空 | 向量数据库返回空响应 |
| ErrInvalidResponse | 响应无效 | 向量数据库返回的响应格式错误 |

### 错误检查函数

| 函数名 | 说明 | 返回值 |
|--------|------|--------|
| CheckBatchErrors | 检查批量操作错误 | 第一个错误或 nil |
| CheckGraphQLResponseError | 检查 GraphQL 响应错误 | 第一个错误或 nil |
| CheckBatchDeleteErrors | 检查批量删除错误 | 第一个错误或 nil |

### 代码示例

```go
// 静态错误使用
if !isValidCodebasePath(path) {
    return vector.ErrInvalidCodebasePath
}

// 批量操作错误检查
if err := vector.CheckBatchErrors(responses); err != nil {
    return err
}

// GraphQL 响应错误检查
if err := vector.CheckGraphQLResponseError(graphqlResp); err != nil {
    return err
}

// 批量删除错误检查
if err := vector.CheckBatchDeleteErrors(deleteResp); err != nil {
    return err
}
```

## 错误处理最佳实践

### 1. 错误码使用原则

- **HTTP 状态码错误**：用于 API 响应，表示请求的状态
- **系统响应码**：用于应用层级的业务逻辑响应
- **具体错误类型**：用于具体的错误场景，提供详细的错误信息

### 2. 错误处理流程

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

### 3. 错误日志记录

```go
func handleError(w http.ResponseWriter, err error) {
    // 记录错误日志
    logx.Errorf("request failed: %v", err)
    
    // 根据错误类型返回相应的响应
    if codeMsg, ok := err.(*response.codeMsg); ok {
        httpx.WriteJson(w, http.StatusBadRequest, codeMsg)
    } else {
        response.Error(w, err)
    }
}
```

### 4. 错误信息国际化

错误信息应该支持国际化，建议在错误信息中使用错误码而非具体的错误消息：

```go
// 前端根据错误码显示对应的国际化消息
{
    "code": 400001,
    "message": "invalid_parameter",
    "data": null
}
```

### 5. 错误恢复策略

对于可恢复的错误，应该提供重试机制：

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

## 更新日志

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| 1.0.0 | 2025-08-25 | 初始版本，整理所有错误码定义 |