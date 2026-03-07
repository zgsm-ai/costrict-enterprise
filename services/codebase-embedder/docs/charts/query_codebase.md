# /api/v1/codebase/query 接口流程图

## 接口说明
代码库查询接口用于获取指定代码库的详细信息，包括汇总信息、语言分布、最近文件、索引统计和详细记录等。

## 请求参数
- `ClientId`: 客户端标识
- `CodebasePath`: 代码库路径
- `CodebaseName`: 代码库名称

## 响应数据
- `CodebaseId`: 代码库ID
- `CodebaseName`: 代码库名称
- `CodebasePath`: 代码库路径
- `Summary`: 代码库汇总信息
- `LanguageDist`: 语言分布统计
- `RecentFiles`: 最近文件列表
- `IndexStats`: 索引统计信息
- `Records`: 详细记录列表

## 接口处理流程

```mermaid
flowchart TD
    A[入口 /codebase-embedder/api/v1/codebase/query<br>POST 方法] --> B[解析请求参数<br>CodebaseQueryRequest]
    B --> C[记录请求详情日志<br>Method, URL, Content-Type]
    C --> D{参数解析是否成功?}
    D -->|否| E[记录解析错误日志<br>包含请求体内容]
    E --> F[返回参数错误响应]
    D -->|是| G[验证请求参数<br>validateRequest]
    G --> H{参数验证是否通过?}
    H -->|否| I[记录验证失败日志]
    I --> J[返回参数错误响应]
    H -->|是| K[创建查询代码库逻辑实例]
    K --> L[执行QueryCodebase方法]
    L --> M[二次验证请求参数]
    M --> N{参数验证是否通过?}
    N -->|否| O[返回参数错误响应]
    N -->|是| P[验证代码库权限<br>verifyCodebasePermission]
    P --> Q[查询数据库验证关联关系]
    Q --> R{代码库记录是否存在?}
    R -->|否| S[记录不存在日志]
    S --> T[返回无权限错误]
    R -->|是| U[检查代码库状态]
    U --> V{状态是否为active?}
    V -->|否| W[记录状态异常日志]
    W --> X[返回状态异常错误]
    V -->|是| Y[创建向量查询存储实例]
    Y --> Z[并行查询各种信息]
    Z --> AA[启动goroutine并行查询]
    AA --> AB[查询汇总信息<br>QueryCodebaseStats]
    AB --> AC{汇总查询是否成功?}
    AC -->|否| AD[设置查询错误]
    AC -->|是| AE[查询语言分布<br>QueryLanguageDistribution]
    AE --> AF{语言查询是否成功?}
    AF -->|否| AD
    AF -->|是| AG[查询最近文件<br>QueryRecentFiles]
    AG --> AH{文件查询是否成功?}
    AH -->|否| AD
    AH -->|是| AI[查询索引统计<br>QueryIndexStats]
    AI --> AJ{统计查询是否成功?}
    AJ -->|否| AD
    AJ -->|是| AK[查询详细记录<br>QueryCodebaseRecords]
    AK --> AL{记录查询是否成功?}
    AL -->|否| AD
    AL -->|是| AM[发送完成信号]
    AD --> AM
    AM --> AN[等待所有查询完成]
    AN --> AO{是否有查询错误?}
    AO -->|是| AP[记录查询失败日志]
    AP --> AQ[返回查询失败错误]
    AO -->|否| AR[构建响应数据]
    AR --> AS[记录查询日志]
    AS --> AT[返回成功响应]
    F --> AU[结束]
    J --> AU
    O --> AU
    T --> AU
    X --> AU
    AQ --> AU
    AT --> AU
```

## 详细处理步骤

### 1. 请求解析与验证
- 解析POST请求中的JSON参数
- 记录详细的请求日志（方法、URL、Content-Type）
- 验证必填字段：ClientId、CodebasePath、CodebaseName

### 2. 权限验证
- 查询数据库验证ClientId与Codebase的关联关系
- 检查代码库状态是否为"active"
- 确保用户有权限访问指定代码库

### 3. 并行查询
使用goroutine并行查询以下信息以提高性能：
- **汇总信息**: 代码库的基本统计信息
- **语言分布**: 各种编程语言的占比统计
- **最近文件**: 最近修改或添加的文件列表
- **索引统计**: 向量索引的统计信息
- **详细记录**: 代码库的详细记录列表

### 4. 响应构建
- 整合所有查询结果
- 构建完整的响应数据结构
- 记录查询成功日志

## 错误处理
- **参数错误**: 当必填字段缺失或无效时返回400错误
- **权限错误**: 当代码库不存在或用户无权限访问时返回403错误
- **状态错误**: 当代码库状态不正常时返回相应错误
- **查询错误**: 当向量数据库查询失败时返回500错误

## 性能优化
- 使用goroutine并行查询多个数据源
- 查询失败时立即终止其他查询
- 合理设置最近文件查询数量（默认10个）
- 记录详细的性能日志用于监控

## 使用示例

### 请求示例
```json
{
  "ClientId": "client123",
  "CodebasePath": "/projects/myapp",
  "CodebaseName": "My Application"
}
```

### 响应示例
```json
{
  "CodebaseId": 123,
  "CodebaseName": "My Application",
  "CodebasePath": "/projects/myapp",
  "Summary": {
    "TotalChunks": 1500,
    "TotalFiles": 200,
    "TotalLines": 50000
  },
  "LanguageDist": [
    {
      "Language": "Go",
      "Percentage": 65.5,
      "FileCount": 130
    },
    {
      "Language": "JavaScript",
      "Percentage": 20.3,
      "FileCount": 41
    }
  ],
  "RecentFiles": [
    {
      "Path": "/projects/myapp/src/main.go",
      "LastModified": "2025-01-24T10:30:00Z",
      "Size": 2048
    }
  ],
  "IndexStats": {
    "IndexedFiles": 195,
    "TotalVectors": 1480,
    "LastIndexed": "2025-01-24T09:15:00Z"
  },
  "Records": [
    {
      "FilePath": "/projects/myapp/src/main.go",
      "Language": "Go",
      "ChunkCount": 15,
      "LastIndexed": "2025-01-24T09:15:00Z"
    }
  ]
}