# 错误码文档索引

本文档提供了 Codebase Embedder 项目错误码文档的快速导航。

## 文档结构

### 📖 主要文档

- **[README.md](./README.md)** - 完整的错误码文档
  - 包含所有错误码的详细说明
  - 错误处理最佳实践
  - 错误码使用示例

- **[cheatsheet.md](./cheatsheet.md)** - 错误码速查表
  - 快速参考代码示例
  - 常见错误处理模式
  - 错误码映射表

## 错误码分类导航

### 🔢 HTTP 状态码错误
- **400 Bad Request** - 参数错误
- **401 Unauthorized** - 认证错误
- **403 Forbidden** - 权限错误
- **429 Too Many Requests** - 限流错误

### 📊 系统响应码
- **0 (CodeOK)** - 成功状态码
- **-1 (CodeError)** - 通用错误状态码

### 🗄️ 数据库错误
- **InsertDatabaseFailed** - 数据库插入失败

### 📝 参数错误
- **NewInvalidParamErr** - 无效参数错误
- **NewRecordNotFoundErr** - 记录未找到错误
- **NewMissingParamError** - 缺少必需参数错误

### ⚡ 任务错误
- **FileNotFound** - 文件或目录未找到
- **ReadTimeout** - 读取超时
- **RunTimeout** - 运行超时

### 🔍 向量存储错误
- **ErrInvalidCodebasePath** - 无效的代码库路径
- **ErrInvalidClientId** - 无效的客户端ID
- **ErrEmptyResponse** - 响应为空
- **ErrInvalidResponse** - 响应无效
- **CheckBatchErrors** - 批量操作错误检查
- **CheckGraphQLResponseError** - GraphQL响应错误检查
- **CheckBatchDeleteErrors** - 批量删除错误检查

## 快速查找

### 按错误类型查找

| 错误类型 | 文档位置 | 代码位置 |
|----------|----------|----------|
| HTTP 状态码错误 | [README.md](./README.md#http-状态码错误) | [`internal/response/code_msg.go`](../../internal/response/code_msg.go) |
| 系统响应码 | [README.md](./README.md#系统响应码) | [`internal/response/resp.go`](../../internal/response/resp.go) |
| 数据库错误 | [README.md](./README.md#数据库错误) | [`internal/errs/database.go`](../../internal/errs/database.go) |
| 参数错误 | [README.md](./README.md#参数错误) | [`internal/errs/param.go`](../../internal/errs/param.go) |
| 任务错误 | [README.md](./README.md#任务错误) | [`internal/errs/task.go`](../../internal/errs/task.go) |
| 向量存储错误 | [README.md](./README.md#向量存储错误) | [`internal/store/vector/error.go`](../../internal/store/vector/error.go) |

### 按使用场景查找

| 场景 | 推荐错误码 | 文档位置 |
|------|------------|----------|
| API 参数验证 | 400, NewInvalidParamErr | [cheatsheet.md](./cheatsheet.md#1-api-参数验证) |
| 用户认证 | 401 | [cheatsheet.md](./cheatsheet.md#http-状态码错误) |
| 权限控制 | 403 | [cheatsheet.md](./cheatsheet.md#http-状态码错误) |
| 数据库操作 | InsertDatabaseFailed | [cheatsheet.md](./cheatsheet.md#2-数据库操作) |
| 文件操作 | FileNotFound, ReadTimeout | [cheatsheet.md](./cheatsheet.md#3-文件操作) |
| 向量存储 | 向量存储错误系列 | [cheatsheet.md](./cheatsheet.md#4-向量存储操作) |

## 相关资源

### 代码文件
- [`internal/response/code_msg.go`](../../internal/response/code_msg.go) - HTTP 状态码错误定义
- [`internal/response/resp.go`](../../internal/response/resp.go) - 系统响应码定义
- [`internal/errs/database.go`](../../internal/errs/database.go) - 数据库错误定义
- [`internal/errs/param.go`](../../internal/errs/param.go) - 参数错误定义
- [`internal/errs/task.go`](../../internal/errs/task.go) - 任务错误定义
- [`internal/store/vector/error.go`](../../internal/store/vector/error.go) - 向量存储错误定义

### 其他文档
- [API 文档](../api_documentation.md) - API 接口文档
- [技术文档](../technical.md) - 技术实现文档
- [测试计划](../test_plan_final.md) - 测试相关文档

## 使用指南

### 新手入门
1. 首先阅读 [README.md](./README.md) 了解错误码的整体架构
2. 查看 [cheatsheet.md](./cheatsheet.md) 获取快速参考
3. 根据具体场景选择合适的错误码

### 有经验的开发者
1. 直接查看 [cheatsheet.md](./cheatsheet.md) 获取代码示例
2. 参考 [错误处理最佳实践](./README.md#错误处理最佳实践) 优化错误处理逻辑
3. 使用索引快速定位到具体的错误码定义

### 维护者
1. 定期更新错误码文档
2. 添加新的错误码时同步更新文档
3. 维护错误码的一致性和规范性

## 贡献指南

如果您发现错误码文档有遗漏或错误，或者需要添加新的错误码，请：

1. 检查相关的代码文件
2. 更新相应的文档
3. 提交 Pull Request

## 版本信息

- **文档版本**: 1.0.0
- **最后更新**: 2025-08-25
- **维护者**: Codebase Embedder Team