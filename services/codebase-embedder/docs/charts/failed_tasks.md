# /api/v1/tasks/failed 接口流程图

## 接口说明
失败任务查询接口用于获取系统中执行失败的任务列表，帮助用户了解失败原因和处理异常情况。

## 请求方式
- 方法：GET
- 路径：/codebase-embedder/api/v1/tasks/failed

## 响应数据
- `Code`: 响应状态码
- `Message`: 响应消息
- `Success`: 是否成功
- `Data`: 失败任务数据
  - `TotalTasks`: 失败任务总数
  - `Tasks`: 任务详情列表

## 接口处理流程

```mermaid
flowchart TD
    A[入口 /codebase-embedder/api/v1/tasks/failed<br>GET 方法] --> B[创建失败任务处理器实例]
    B --> C[调用ServeHTTP方法处理请求]
    C --> D[验证请求方法]
    D --> E{是否为GET方法?}
    E -->|否| F[返回方法不允许错误]
    E -->|是| G[创建失败任务逻辑实例]
    G --> H[执行GetFailedTasks方法]
    H --> I[检查Redis连接<br>checkRedisConnection]
    I --> J[创建带超时的上下文<br>5秒超时]
    J --> K[调用StatusManager检查连接]
    K --> L{Redis连接是否正常?}
    L -->|否| M[返回Redis服务不可用错误]
    L -->|是| N[扫描失败任务<br>scanFailedTasks]
    N --> O[调用StatusManager扫描失败任务]
    O --> P{扫描操作是否成功?}
    P -->|否| Q[返回查询失败任务错误]
    P -->|是| R[构建响应数据]
    R --> S[设置响应状态码为0]
    S --> T[设置响应消息为"ok"]
    T --> U[设置成功标志为true]
    U --> V[设置失败任务总数]
    V --> W[设置失败任务列表]
    W --> X[返回成功响应]
    F --> Y[结束]
    M --> Y
    Q --> Y
    X --> Y
```

## 详细处理步骤

### 1. 请求验证
- 验证HTTP请求方法是否为GET
- 拒绝非GET方法的请求，返回405错误

### 2. 连接检查
- 创建带有5秒超时的上下文
- 检查Redis服务连接状态
- 确保Redis服务可用后继续处理

### 3. 任务扫描
- 调用StatusManager扫描失败的任务
- 获取所有执行失败的任务信息
- 处理扫描过程中的异常情况

### 4. 响应构建
- 统计失败任务的总数
- 构建包含任务详情的响应数据
- 设置响应状态码、消息和成功标志

### 5. 响应返回
- 返回JSON格式的响应数据
- 包含任务总数和任务列表信息

## 任务信息结构

每个失败任务包含以下信息：
- 任务ID
- 任务类型
- 客户端ID
- 代码库路径
- 开始时间
- 失败时间
- 错误信息
- 错误代码
- 重试次数
- 处理的文件数量
- 失败原因分类

## 错误处理
- **方法错误**: 当请求方法不是GET时返回405错误
- **连接错误**: 当Redis服务不可用时返回服务不可用错误
- **查询错误**: 当扫描失败任务失败时返回内部错误

## 性能考虑
- 使用5秒超时上下文，避免长时间阻塞
- Redis扫描操作性能较高，适合频繁查询
- 建议合理设置查询频率，避免过度频繁
- 失败任务数据量通常较少，查询性能较好

## 使用示例

### 请求
```bash
GET /codebase-embedder/api/v1/tasks/failed
```

### 成功响应示例
```json
{
  "Code": 0,
  "Message": "ok",
  "Success": true,
  "Data": {
    "TotalTasks": 3,
    "Tasks": [
      {
        "TaskId": "task_001",
        "TaskType": "embedding",
        "ClientId": "client123",
        "CodebasePath": "/projects/myapp",
        "StartTime": "2025-01-24T09:00:00Z",
        "FailTime": "2025-01-24T09:15:00Z",
        "ErrorCode": "TIMEOUT_ERROR",
        "ErrorMessage": "向量嵌入服务超时，请求超过30秒未响应",
        "RetryCount": 3,
        "ProcessedFiles": 75,
        "ErrorCategory": "service_timeout"
      },
      {
        "TaskId": "task_002",
        "TaskType": "indexing",
        "ClientId": "client456",
        "CodebasePath": "/projects/another",
        "StartTime": "2025-01-24T09:20:00Z",
        "FailTime": "2025-01-24T09:22:00Z",
        "ErrorCode": "PERMISSION_DENIED",
        "ErrorMessage": "没有权限访问指定代码库",
        "RetryCount": 0,
        "ProcessedFiles": 0,
        "ErrorCategory": "authorization"
      },
      {
        "TaskId": "task_003",
        "TaskType": "cleaning",
        "ClientId": "client789",
        "CodebasePath": "/projects/old",
        "StartTime": "2025-01-24T10:00:00Z",
        "FailTime": "2025-01-24T10:01:00Z",
        "ErrorCode": "STORAGE_ERROR",
        "ErrorMessage": "存储空间不足，无法完成清理操作",
        "RetryCount": 1,
        "ProcessedFiles": 50,
        "ErrorCategory": "resource_exhausted"
      }
    ]
  }
}
```

### 错误响应示例
```json
{
  "Code": 500,
  "Message": "Redis服务不可用，请稍后再试",
  "Success": false,
  "Data": null
}
```

## 监控用途
此接口主要用于系统监控和故障排查：
- 快速定位系统中执行失败的任务
- 分析失败原因和模式
- 监控系统健康状态
- 辅助故障诊断和修复
- 统计失败率和错误类型分布

## 错误分类

### 服务超时 (service_timeout)
- 向量嵌入服务响应超时
- 数据库查询超时
- 网络请求超时

### 授权错误 (authorization)
- 权限不足
- 认证失败
- 访问被拒绝

### 资源耗尽 (resource_exhausted)
- 存储空间不足
- 内存不足
- 连接池耗尽

### 数据错误 (data_error)
- 数据格式错误
- 数据完整性问题
- 文件损坏

### 系统错误 (system_error)
- 内部服务异常
- 系统资源不足
- 未知异常

## 数据保留策略
- 失败任务数据在Redis中保留较长时间
- 建议保留足够时间用于故障分析
- 可通过配置文件设置保留时间
- 定期归档历史失败任务数据

## 重试机制
- 系统会自动重试失败的任务
- 重试次数和间隔可配置
- 达到最大重试次数后标记为最终失败
- 某些类型的错误（如权限问题）不会重试

## 告警建议
- 失败任务数量突然增加时触发告警
- 特定类型错误频繁出现时触发告警
- 失败率超过阈值时触发告警
- 建议设置不同级别的告警策略

## 注意事项
- 此接口只返回失败的任务，不包括成功或正在运行的任务
- Redis连接检查使用5秒超时，超时会返回错误
- 任务数据按失败时间倒序排列
- 建议定期调用此接口进行系统健康检查
- 失败任务信息对于系统维护非常重要
- 应及时处理失败任务，避免影响用户体验