# /api/v1/files/status 接口流程图

## 接口说明
任务状态查询接口用于查询文件处理状态，通过SyncId获取文件上传和处理状态信息。

## 请求参数
- `SyncId`: 同步任务ID

## 响应数据
- `Process`: 处理状态（success/failed/processing）
- `TotalProgress`: 总进度（0-100）
- `FileList`: 文件状态列表

## 接口处理流程

```mermaid
flowchart TD
    A[入口 /codebase-embedder/api/v1/files/status<br>POST 方法] --> B[创建状态处理器实例]
    B --> C[调用ServeHTTP方法处理请求]
    C --> D[解析请求参数<br>FileStatusRequest]
    D --> E{参数解析是否成功?}
    E -->|否| F[返回参数错误响应]
    E -->|是| G[创建状态查询逻辑实例]
    G --> H[执行GetFileStatus方法]
    H --> I[获取状态管理器<br>StatusManager]
    I --> J[提取SyncId作为请求ID]
    J --> K[从Redis查询文件状态<br>GetFileStatus]
    K --> L{Redis查询是否成功?}
    L -->|否| M[返回Redis查询失败错误]
    L -->|是| N{是否找到状态记录?}
    N -->|是| O[返回Redis中的状态数据]
    N -->|否| P[构建默认失败状态响应]
    P --> Q[设置Process为"failed"]
    Q --> R[设置TotalProgress为0]
    R --> S[设置空文件列表]
    S --> T[返回默认响应]
    O --> U[返回成功响应]
    F --> V[结束]
    M --> V
    U --> V
    T --> V
```

## 详细处理步骤

### 1. 请求解析与验证
- 解析POST请求中的JSON参数
- 提取SyncId作为查询键
- 验证参数格式是否正确

### 2. 状态查询
- 获取Redis状态管理器实例
- 使用SyncId作为键查询Redis中的状态信息
- 检查查询操作是否成功

### 3. 状态处理
- 如果Redis中存在状态记录，直接返回该状态
- 如果Redis中不存在状态记录，返回默认的失败状态
- 默认状态包含：Process="failed"、TotalProgress=0、空文件列表

### 4. 响应构建
- 构建包含状态信息的响应数据
- 返回JSON格式的响应

## 状态值说明

### Process字段
- `success`: 文件处理成功完成
- `failed`: 文件处理失败
- `processing`: 文件正在处理中

### TotalProgress字段
- 范围：0-100
- 0表示未开始或失败
- 100表示完成
- 中间值表示处理进度百分比

### FileList字段
- 包含每个文件的处理状态
- 为空数组表示没有文件信息或查询失败

## 错误处理
- **参数错误**: 当请求参数解析失败时返回400错误
- **Redis错误**: 当Redis查询操作失败时返回500错误
- **数据不存在**: 当未找到状态记录时返回默认失败状态

## 使用示例

### 请求示例
```json
{
  "SyncId": "sync_1234567890"
}
```

### 成功响应示例（找到状态记录）
```json
{
  "Process": "success",
  "TotalProgress": 100,
  "FileList": [
    {
      "FilePath": "/projects/myapp/src/main.go",
      "Status": "success",
      "Progress": 100,
      "Message": "处理完成"
    },
    {
      "FilePath": "/projects/myapp/src/utils.go",
      "Status": "success",
      "Progress": 100,
      "Message": "处理完成"
    }
  ]
}
```

### 默认响应示例（未找到状态记录）
```json
{
  "Process": "failed",
  "TotalProgress": 0,
  "FileList": []
}
```

## 性能考虑
- 使用Redis作为状态存储，查询速度快
- 简单的键值查询，性能开销小
- 建议合理设置状态记录的过期时间，避免Redis内存占用过大

## 状态管理
- 状态信息由其他处理流程写入Redis
- 使用SyncId作为唯一标识符
- 状态信息包含处理进度、文件列表等详细信息
- 默认失败状态用于处理查询不到记录的情况

## 集成说明
- 此接口通常与文件上传接口配合使用
- 上传接口返回SyncId，客户端使用此ID查询处理状态
- 支持轮询查询，直到获取最终状态（success/failed）