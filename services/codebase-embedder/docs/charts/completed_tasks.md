# /api/v1/tasks/completed 接口流程图

## 接口说明
已完成任务查询接口用于获取系统中已完成的任务列表，帮助用户了解历史任务执行情况和结果。

## 请求方式
- 方法：GET
- 路径：/codebase-embedder/api/v1/tasks/completed

## 响应数据
- `Code`: 响应状态码
- `Message`: 响应消息
- `Success`: 是否成功
- `Data`: 已完成任务数据
  - `TotalTasks`: 已完成任务总数
  - `Tasks`: 任务详情列表

## 接口处理流程

```mermaid
flowchart TD
    A[入口 /codebase-embedder/api/v1/tasks/completed<br>GET 方法] --> B[创建已完成任务处理器实例]
    B --> C[调用ServeHTTP方法处理请求]
    C --> D[验证请求方法]
    D --> E{是否为GET方法?}
    E -->|否| F[返回方法不允许错误]
    E -->|是| G[创建已完成任务逻辑实例]
    G --> H[执行GetCompletedTasks方法]
    H --> I[检查Redis连接<br>checkRedisConnection]
    I --> J[创建带超时的上下文<br>5秒超时]
    J --> K[调用StatusManager检查连接]
    K --> L{Redis连接是否正常?}
    L -->|否| M[返回Redis服务不可用错误]
    L -->|是| N[扫描已完成任务<br>scanCompletedTasks]
    N --> O[调用StatusManager扫描已完成任务]
    O --> P{扫描操作是否成功?}
    P -->|否| Q[返回查询已完成任务错误]
    P -->|是| R[构建响应数据]
    R --> S[设置响应状态码为0]
    S --> T[设置响应消息为"ok"]
    T --> U[设置成功标志为true]
    U --> V[设置已完成任务总数]
    V --> W[设置已完成任务列表]
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
- 调用StatusManager扫描已完成的任务
- 获取所有已执行完成的任务信息
- 处理扫描过程中的异常情况

### 4. 响应构建
- 统计已完成任务的总数
- 构建包含任务详情的响应数据
- 设置响应状态码、消息和成功标志

### 5. 响应返回
- 返回JSON格式的响应数据
- 包含任务总数和任务列表信息

## 任务信息结构

每个已完成任务包含以下信息：
- 任务ID
- 任务类型
- 客户端ID
- 代码库路径
- 开始时间
- 完成时间
- 执行结果
- 任务状态（success/failed）
- 处理的文件数量
- 错误信息（如果有）

## 错误处理
- **方法错误**: 当请求方法不是GET时返回405错误
- **连接错误**: 当Redis服务不可用时返回服务不可用错误
- **查询错误**: 当扫描已完成任务失败时返回内部错误

## 性能考虑
- 使用5秒超时上下文，避免长时间阻塞
- Redis扫描操作性能较高，适合频繁查询
- 建议合理设置查询频率，避免过度频繁
- 已完成任务数据可能较多，考虑分页查询

## 使用示例

### 请求
```bash
GET /codebase-embedder/api/v1/tasks/completed
```

### 成功响应示例
```json
{
  "Code": 0,
  "Message": "ok",
  "Success": true,
  "Data": {
    "TotalTasks": 5,
    "Tasks": [
      {
        "TaskId": "task_001",
        "TaskType": "embedding",
        "ClientId": "client123",
        "CodebasePath": "/projects/myapp",
        "StartTime": "2025-01-24T09:00:00Z",
        "EndTime": "2025-01-24T09:15:00Z",
        "Status": "success",
        "ProcessedFiles": 150,
        "ErrorMessage": ""
      },
      {
        "TaskId": "task_002",
        "TaskType": "indexing",
        "ClientId": "client456",
        "CodebasePath": "/projects/another",
        "StartTime": "2025-01-24T09:20:00Z",
        "EndTime": "2025-01-24T09:45:00Z",
        "Status": "failed",
        "ProcessedFiles": 75,
        "ErrorMessage": "向量嵌入服务超时"
      },
      {
        "TaskId": "task_003",
        "TaskType": "cleaning",
        "ClientId": "client789",
        "CodebasePath": "/projects/old",
        "StartTime": "2025-01-24T10:00:00Z",
        "EndTime": "2025-01-24T10:05:00Z",
        "Status": "success",
        "ProcessedFiles": 200,
        "ErrorMessage": ""
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
此接口主要用于系统监控和运维：
- 了解历史任务执行情况
- 统计任务成功率和失败率
- 分析任务执行时间分布
- 识别系统性能瓶颈
- 辅助故障排查和问题定位

## 数据保留策略
- 已完成任务数据在Redis中保留一定时间
- 超过保留时间的任务数据会被自动清理
- 保留时间可通过配置文件设置
- 建议根据业务需求合理设置保留时间

## 注意事项
- 此接口只返回已完成的任务，不包括正在运行的任务
- Redis连接检查使用5秒超时，超时会返回错误
- 任务数据按完成时间倒序排列
- 建议定期调用此接口进行历史数据分析
- 大量任务数据可能影响查询性能，建议使用分页查询