# /api/v1/tasks/running 接口流程图

## 接口说明
运行中任务查询接口用于获取当前系统中正在运行的任务列表，帮助用户了解系统当前的负载情况。

## 请求方式
- 方法：GET
- 路径：/codebase-embedder/api/v1/tasks/running

## 响应数据
- `Code`: 响应状态码
- `Message`: 响应消息
- `Success`: 是否成功
- `Data`: 运行中任务数据
  - `TotalTasks`: 运行中任务总数
  - `Tasks`: 任务详情列表

## 接口处理流程

```mermaid
flowchart TD
    A[入口 /codebase-embedder/api/v1/tasks/running<br>GET 方法] --> B[创建运行中任务处理器实例]
    B --> C[调用ServeHTTP方法处理请求]
    C --> D[验证请求方法]
    D --> E{是否为GET方法?}
    E -->|否| F[返回方法不允许错误]
    E -->|是| G[创建运行中任务逻辑实例]
    G --> H[执行GetRunningTasks方法]
    H --> I[检查Redis连接<br>checkRedisConnection]
    I --> J[创建带超时的上下文<br>5秒超时]
    J --> K[调用StatusManager检查连接]
    K --> L{Redis连接是否正常?}
    L -->|否| M[返回Redis服务不可用错误]
    L -->|是| N[扫描运行中任务<br>scanRunningTasks]
    N --> O[调用StatusManager扫描任务]
    O --> P{扫描操作是否成功?}
    P -->|否| Q[返回查询任务状态错误]
    P -->|是| R[构建响应数据]
    R --> S[设置响应状态码为0]
    S --> T[设置响应消息为"ok"]
    T --> U[设置成功标志为true]
    U --> V[设置任务总数]
    V --> W[设置任务列表]
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
- 调用StatusManager扫描运行中的任务
- 获取所有正在执行的任务信息
- 处理扫描过程中的异常情况

### 4. 响应构建
- 统计运行中任务的总数
- 构建包含任务详情的响应数据
- 设置响应状态码、消息和成功标志

### 5. 响应返回
- 返回JSON格式的响应数据
- 包含任务总数和任务列表信息

## 任务信息结构

每个运行中任务包含以下信息：
- 任务ID
- 任务类型
- 开始时间
- 执行进度
- 相关代码库信息
- 任务状态详情

## 错误处理
- **方法错误**: 当请求方法不是GET时返回405错误
- **连接错误**: 当Redis服务不可用时返回服务不可用错误
- **查询错误**: 当扫描任务状态失败时返回内部错误

## 性能考虑
- 使用5秒超时上下文，避免长时间阻塞
- Redis扫描操作性能较高，适合频繁查询
- 建议合理设置查询频率，避免过度频繁

## 使用示例

### 请求
```bash
GET /codebase-embedder/api/v1/tasks/running
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
        "StartTime": "2025-01-24T09:15:00Z",
        "Progress": 45,
        "Status": "running"
      },
      {
        "TaskId": "task_002",
        "TaskType": "indexing",
        "ClientId": "client456",
        "CodebasePath": "/projects/another",
        "StartTime": "2025-01-24T09:20:00Z",
        "Progress": 78,
        "Status": "running"
      },
      {
        "TaskId": "task_003",
        "TaskType": "cleaning",
        "ClientId": "client789",
        "CodebasePath": "/projects/old",
        "StartTime": "2025-01-24T09:25:00Z",
        "Progress": 12,
        "Status": "running"
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
- 了解当前系统负载情况
- 监控长时间运行的任务
- 识别潜在的性能瓶颈
- 辅助系统容量规划

## 注意事项
- 此接口只返回正在运行的任务，不包括已完成或失败的任务
- Redis连接检查使用5秒超时，超时会返回错误
- 建议定期调用此接口进行系统监控
- 任务信息实时性依赖Redis中的状态更新