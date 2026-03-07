# Codebase Embedder 功能文档

## 1. 系统功能概述

Codebase Embedder 是一个代码库嵌入管理系统，旨在为开发人员提供代码库的智能分析和检索能力。系统通过将代码库转换为向量嵌入（embeddings），实现高效的语义搜索和代码理解功能。

核心功能包括：
- **代码库嵌入管理**：将整个代码库上传并转换为向量表示，为后续的语义搜索提供基础
- **语义代码搜索**：基于自然语言查询，查找语义上相关的代码片段，而不仅仅是关键词匹配
- **代码库摘要**：提供代码库的统计信息和处理状态，帮助用户了解代码库的整体情况
- **文件处理状态监控**：实时追踪代码库中各个文件的处理进度和状态
- **服务健康检查**：提供系统状态检查接口，确保服务正常运行
- **代码库树结构展示**：提供代码库的目录树结构展示功能，帮助用户直观了解项目结构

系统采用微服务架构，主要组件包括：
- **HTTP API 服务**：提供RESTful接口供客户端调用
- **向量存储**：使用Weaviate等向量数据库存储代码嵌入
- **任务队列**：管理嵌入任务的执行和调度
- **分布式锁**：确保同一代码库的并发操作安全
- **状态管理**：使用Redis存储临时状态信息

## 2. 各个接口的功能说明

### 2.1 提交嵌入任务 (POST /codebase-embedder/api/v1/embeddings)

**功能描述**：
提交代码库嵌入任务，将本地代码库上传并转换为向量表示。系统会解析代码库中的文件，提取代码结构和语义信息，生成向量嵌入。上传的ZIP文件必须包含`.shenma_sync`文件夹，该文件夹用于存储同步相关的元数据。

**请求参数**（form-data格式）：
- `clientId` (string, 必填): 客户端唯一标识，用于区分不同用户的代码库
- `codebasePath` (string, 必填): 项目在客户端的绝对路径
- `codebaseName` (string, 必填): 代码库名称
- `uploadToken` (string, 可选): 上传令牌，用于验证上传权限（当前调试阶段使用"xxxx"作为万能令牌）
- `fileTotals` (number, 必填): 上传工程文件总数
- `file` (file, 必填): 代码库的ZIP压缩文件，必须包含`.shenma_sync`文件夹
- `extraMetadata` (string, 可选): 额外元数据（JSON字符串格式）
- `X-Request-ID` (header, 可选): 请求ID，用于跟踪和调试，如果没有提供则系统会自动生成

**处理流程**：
1. **参数验证**：验证必填字段（clientId、codebasePath、codebaseName）
2. **令牌验证**：验证uploadToken的有效性（当前调试阶段跳过验证）
3. **代码库初始化**：在数据库中查找或创建代码库记录，使用clientId和codebasePath作为唯一标识
4. **分布式锁获取**：获取基于codebaseID的分布式锁，防止重复处理，锁超时时间可配置
5. **ZIP文件处理**：
   - 验证上传文件为ZIP格式
   - 创建临时文件存储ZIP内容
   - 检查ZIP文件中必须包含`.shenma_sync`文件夹
   - 遍历ZIP文件，读取所有代码文件内容（跳过`.shenma_sync`文件夹中的文件）
   - 读取`.shenma_sync`文件夹中的元数据文件用于任务管理
   - 同步元数据数据格式如下,文件名为时间戳：
   ```json
    {
    "clientId": ""
    "codebasePath": "",
    "codebaseName": "",
    "extraMetadata":  {},
    "fileList":  {
        "src/main/java/main.java": "add" , //add  modify   delete
      },
    "timestamp": 12334234233
    }
    ```

6. **检查文件**：
   - 验证同步元数据文件中add和modify里面文件是否和解压后文件匹配，并打印匹配结果
7. **数据库更新**：更新代码库的文件数量（file_count）和总大小（total_size）信息
8. **任务提交**：将嵌入任务提交到异步任务队列进行处理
9. **状态初始化**：在Redis中使用requestId作为键初始化文件处理状态

**ZIP文件结构要求**：
```
project.zip
├── .shenma_sync/          # 必须存在的文件夹
│   ├── 20250728213645    # 同步元数据文件,文件名为时间戳
│   └── ...               # 其他同步相关文件
├── src/
│   ├── main.js
│   └── utils.js
├── package.json
└── ...                   # 其他项目文件
```

**成功响应**：
```json
{
  "taskId": 12345
}
```

**错误响应**：
- 400 Bad Request：缺少必填参数或ZIP文件格式不正确
- 409 Conflict：无法获取分布式锁，任务正在处理中
- 422 Unprocessable Entity：ZIP文件中缺少必需的`.shenma_sync`文件夹

**注意事项**：
- 上传的ZIP文件大小限制为32MB（可在配置中调整）
- 系统会自动跳过`.shenma_sync`文件夹中的文件，这些文件仅用于任务管理
- 任务处理状态可通过文件状态查询接口进行监控
- 每个代码库（由clientId和codebasePath唯一标识）同时只能有一个处理任务

### 2.2 删除嵌入数据 (DELETE /codebase-embedder/api/v1/embeddings)

**功能描述**：
删除指定代码库的嵌入数据，包括向量存储中的嵌入和数据库中的相关记录。

**请求参数**：
- `clientId` (string): 客户端唯一标识
- `projectPath` (string): 项目路径
- `filePaths` (array, optional): 要删除的特定文件路径列表，为空时删除整个代码库

**成功响应**：
```json
{}
```

### 2.3 获取代码库摘要 (GET /codebase-embedder/api/v1/embeddings/summary)

**功能描述**：
获取指定代码库的摘要信息，包括嵌入状态、文件数量、处理进度等。

**请求参数**：
- `clientId` (string): 客户端唯一标识
- `projectPath` (string): 项目路径

**响应数据**：
```json
{
  "embedding": {
    "status": "completed",
    "updatedAt": "2025-07-28T12:00:00Z",
    "totalFiles": 42,
    "totalChunks": 156
  },
  "status": "active",
  "totalFiles": 42
}
```

**字段说明**：
- `embedding.status`: 嵌入处理状态（pending/processing/completed/failed）
- `embedding.updatedAt`: 最后更新时间
- `embedding.totalFiles`: 已处理的文件数量
- `embedding.totalChunks`: 生成的代码块数量
- `totalFiles`: 代码库总文件数量

### 2.4 服务状态检查 (GET /codebase-embedder/api/v1/status)

**功能描述**：
检查服务的健康状态，确认服务是否正常运行。

**成功响应**：
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

### 2.5 语义代码搜索 (GET /codebase-embedder/api/v1/search/semantic)

**功能描述**：
执行语义代码搜索，根据自然语言查询查找相关的代码片段。系统会将查询转换为向量，并在向量空间中查找最相似的代码块。

**请求参数**：
- `clientId` (string): 客户端唯一标识
- `projectPath` (string): 项目路径
- `query` (string): 搜索查询，可以是自然语言描述
- `topK` (number, optional): 返回结果数量，默认为5
- `scoreThreshold` (number, optional): 相似度分数阈值，过滤低相关性结果

**成功响应**：
```json
{
  "list": [
    {
      "content": "function authenticateUser() {...}",
      "filePath": "src/auth.js",
      "score": 0.92
    },
    {
      "content": "class AuthMiddleware {...}",
      "filePath": "middleware/auth.py",
      "score": 0.87
    }
  ]
}
```

**字段说明**：
- `content`: 匹配的代码片段内容
- `filePath`: 代码文件路径
- `score`: 相似度分数（0-1），分数越高表示相关性越强

### 2.6 文件状态查询 (POST /codebase-embedder/api/v1/files/status)

**功能描述**：
查询代码库中文件的处理状态，了解嵌入任务的进度。

**请求参数**：
- `clientId` (string): 客户端唯一标识
- `codebasePath` (string): 代码库路径
- `codebaseName` (string): 代码库名称

**成功响应**：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "process": "processing",
    "totalProgress": 50,
    "fileList": [
      {
        "path": "src/main/java/main.java",
        "operate":"add",
        "status": "complete"
      },
      {
        "path": "src/main/java/server.java",
        "operate":"modify",
        "status": "complete"
      },
      {
        "path": "src/main/java/server.java",
        "operate":"delete",
        "status": "complete"
      }
    ]
  }
}
```

**字段说明**：
- `process`: 整体处理状态（pending/processing/complete/failed）
- `totalProgress`: 整体处理进度百分比（0-100）
- `fileList`: 文件状态列表
  - `path`: 文件路径
  - `status`: 文件处理状态（pending/processing/complete/failed）

### 2.7 文件上传接口 (POST /codebase-embedder/api/v1/files/upload)

**功能描述**：
分片上传代码库文件，适用于大文件上传场景。
如果uploadToken上传，则刷新一下

**请求参数**：
- `clientId` (string): 客户端唯一标识
- `codebasePath` (string): 项目绝对路径
- `codebaseName` (string): 项目名称
- `uploadToken` (string): 上传令牌
- `extraMetadata` (object, optional): 额外元数据
- `chunkNumber` (number, optional): 当前分片编号，默认为0
- `totalChunks` (number, optional): 分片总数，默认为1
- `fileTotals` (number): 上传工程文件总数
- `file` (file): 当前分片文件

**成功响应**：
```json
{
  "taskId": 12345
}
```

### 2.8 查询代码库信息 (POST /codebase-embedder/api/v1/codebase/query)

**功能描述**：
查询代码库的全面信息，包括统计信息、语言分布、最近文件、索引统计和详细记录。此接口提供代码库的综合视图，帮助用户全面了解代码库的结构和状态。注意：部分功能（如最近文件、详细记录）尚未完全实现，返回的数据可能为空或不完整。

**请求参数**：
- `clientId` (string): 客户端唯一标识
- `codebasePath` (string): 代码库在客户端的绝对路径
- `codebaseName` (string): 代码库名称

**处理流程**：
1. **参数验证**：验证必填字段（clientId、codebasePath、codebaseName）
2. **权限验证**：验证客户端是否有权访问指定代码库
3. **并行查询**：并行执行多个数据库查询以获取代码库信息
   - 查询代码库基本信息（ID、创建时间、更新时间）
   - 查询代码库统计信息（文件总数、总大小、代码行数等）
   - 查询语言分布信息
   - 查询最近修改的文件列表（部分功能未实现）
   - 查询索引统计信息（已处理文件数、待处理文件数等）
   - 查询详细记录（部分功能未实现）
4. **结果整合**：将所有查询结果整合为统一响应格式
5. **日志记录**：记录查询操作和结果摘要

**响应数据**：
```json
{
    "code": 0,
    "message": "ok",
    "success": true,
    "data": {
        "codebaseId": 12,
        "codebaseName": "codebase-embedder",
        "codebasePath": "D:\\workspace\\codebase-embedder",
        "summary": {
            "totalFiles": 100,
            "totalChunks": 907,
            "lastUpdateTime": "2025-08-04T11:09:12.1751558+08:00",
            "indexStatus": "",
            "indexProgress": 0
        },
        "languageDistribution": [],
        "recentFiles": [],
        "indexStats": {
            "averageChunkSize": 0,
            "maxChunkSize": 0,
            "minChunkSize": 0,
            "totalVectors": 0
        },
        "records": [
            {
                "id": "722f9e2f-b575-44a9-9d7f-ac9921b51edc",
                "filePath": "pkg/utils/slice.go",
                "language": "",
                "content": "func SliceContains[T comparable](slice []T, value T) bool {\r\n\tfor _, v := range slice {\r\n\t\tif v == value {\r\n\t\t\treturn true\r\n\t\t}\r\n\t}\r\n\treturn false\r\n}",
                "range": [
                    7,
                    0,
                    14,
                    1
                ],
                "tokenCount": 41,
                "lastUpdated": "2025-08-04T11:09:12.3005914+08:00"
            },
            {
                "id": "9e194109-cc96-41b4-ace5-3c04ff4ede0c",
                "filePath": "internal/parser/testdata/test.ts",
                "language": "",
                "content": "class UIElement implements Drawable, Resizable {\r\n    draw(): void {\r\n        console.log(\"Drawing UI element\");\r\n    }\r\n\r\n    resize(width: number, height: number): void {\r\n        console.log(`Resizing to ${width}x${height}`);\r\n    }\r\n}",
                "range": [
                    158,
                    0,
                    166,
                    1
                ],
                "tokenCount": 54,
                "lastUpdated": "2025-08-04T11:09:12.3005914+08:00"
            }
          ]
    }
```

**错误响应**：
- 400 Bad Request：缺少必填参数或参数格式不正确
- 403 Forbidden：客户端无权访问指定代码库
- 500 Internal Server Error：服务器内部错误，如数据库查询失败

**注意事项**：
- 部分功能（如最近文件、详细记录）尚未完全实现，返回的数据可能为空或不完整
- 该接口可能涉及大量数据查询，在高并发场景下需注意性能影响
- 为了提高查询效率，建议客户端合理使用缓存机制

**使用示例**：
```bash
curl -X POST "http://localhost:8080/codebase-embedder/api/v1/codebase/query" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "user123",
    "codebasePath": "/home/user/myproject",
    "codebaseName": "myproject"
  }'
```

### 2.9 代码库树结构查询 (POST /codebase-embedder/api/v1/codebase/tree)

**功能描述**：
获取指定代码库的目录树结构，支持自定义最大深度和是否包含文件的选项。此接口通过从向量存储中获取文件路径信息，然后构建层次化的目录树结构，帮助用户直观了解代码库的组织结构。

**请求参数**：
- `clientId` (string, 必填): 客户端唯一标识，用于区分不同用户的代码库
- `codebasePath` (string, 必填): 代码库在客户端的绝对路径，用于确定要查询的代码库
- `codebaseName` (string, 必填): 代码库名称，用于标识代码库
- `maxDepth` (number, 可选): 目录树最大深度，默认为10，设置可以限制返回的目录层级深度
- `includeFiles` (boolean, 可选): 是否包含文件节点，默认为true，设置为false则只显示目录结构

**处理流程**：
1. **参数验证**：验证必填字段（clientId、codebasePath、codebaseName）
2. **权限验证**：验证客户端是否有权访问指定代码库，从数据库查询真实的codebaseId
3. **数据获取**：从向量存储中获取指定代码库的所有文件路径记录
4. **路径规范化**：对获取的文件路径进行标准化处理，统一路径分隔符格式
5. **路径去重**：去除重复的文件路径，确保每个路径只处理一次
6. **根路径提取**：分析所有文件路径，提取公共根路径作为目录树的根节点
7. **目录树构建**：根据文件路径层次结构递归构建目录树节点
8. **深度过滤**：根据maxDepth参数过滤超出指定深度的节点
9. **类型过滤**：根据includeFiles参数决定是否包含文件节点
10. **结果返回**：返回构建完成的目录树结构

**响应数据**：
```json
{
    "code": 0,
    "message": "ok",
    "success": true,
    "data": {
        "name": "project-root",
        "path": "/path/to/project-root",
        "type": "directory",
        "children": [
            {
                "name": "src",
                "path": "/path/to/project-root/src",
                "type": "directory",
                "children": [
                    {
                        "name": "main.js",
                        "path": "/path/to/project-root/src/main.js",
                        "type": "file",
                        "size": 2048,
                        "lastModified": "2025-08-04T12:00:00Z"
                    },
                    {
                        "name": "utils",
                        "path": "/path/to/project-root/src/utils",
                        "type": "directory",
                        "children": [
                            {
                                "name": "helper.js",
                                "path": "/path/to/project-root/src/utils/helper.js",
                                "type": "file",
                                "size": 1024,
                                "lastModified": "2025-08-04T11:30:00Z"
                            }
                        ]
                    }
                ]
            },
            {
                "name": "docs",
                "path": "/path/to/project-root/docs",
                "type": "directory",
                "children": [
                    {
                        "name": "README.md",
                        "path": "/path/to/project-root/docs/README.md",
                        "type": "file",
                        "size": 4096,
                        "lastModified": "2025-08-03T15:45:00Z"
                    }
                ]
            }
        ]
    }
}
```

**字段说明**：
- `data.name`: 节点名称，文件名或目录名
- `data.path`: 节点完整路径，使用系统标准路径分隔符
- `data.type`: 节点类型，"file"表示文件节点，"directory"表示目录节点
- `data.size`: 文件大小（字节），仅文件节点有效
- `data.lastModified`: 最后修改时间，仅文件节点有效
- `data.children`: 子节点列表，仅目录节点有效，包含该目录下的所有文件和子目录

**错误响应**：
- 400 Bad Request：缺少必填参数或参数格式不正确
- 403 Forbidden：客户端无权访问指定代码库
- 404 Not Found：指定的代码库不存在或无数据
- 500 Internal Server Error：服务器内部错误，如数据库查询失败或向量存储连接失败

**注意事项**：
- 目录树构建依赖于向量存储中的文件路径数据，如果向量存储中没有对应数据，将返回空目录树
- 路径规范化确保跨平台兼容性，在不同操作系统上都能正确显示路径结构
- 大型代码库的目录树构建可能需要较长时间，建议合理设置maxDepth参数以控制返回数据量
- 如果includeFiles设置为false，将只显示目录结构，便于快速浏览项目架构

**使用示例**：
```bash
# 查询完整目录树（包含文件）
curl -X POST "http://localhost:8080/codebase-embedder/api/v1/codebase/tree" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "user123",
    "codebasePath": "/home/user/myproject",
    "codebaseName": "myproject"
  }'

# 查询只包含目录结构的树（不包含文件）
curl -X POST "http://localhost:8080/codebase-embedder/api/v1/codebase/tree" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "user123",
    "codebasePath": "/home/user/myproject",
    "codebaseName": "myproject",
    "includeFiles": false
  }'

# 限制目录树深度为3层
curl -X POST "http://localhost:8080/codebase-embedder/api/v1/codebase/tree" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "user123",
    "codebasePath": "/home/user/myproject",
    "codebaseName": "myproject",
    "maxDepth": 3
  }'
```

## 3. 核心逻辑函数说明

### 3.1 CodebaseTreeLogic 核心逻辑函数

#### 3.1.1 GetCodebaseTree 函数

**功能概述**：
CodebaseTreeLogic的核心方法，负责协调整个目录树构建流程，包括参数验证、权限验证、数据获取和目录树构建等步骤。

**接口定义**：
```go
func (l *CodebaseTreeLogic) GetCodebaseTree(req *types.CodebaseTreeRequest) (*types.CodebaseTreeResponse, error)
```

**输入参数**：
- `req *types.CodebaseTreeRequest`: 包含客户端ID、代码库路径、名称、最大深度和是否包含文件等参数

**输出参数**：
- `*types.CodebaseTreeResponse`: 包含响应码、消息、成功状态和目录树数据的结构体
- `error`: 错误信息，处理失败时返回具体错误

**核心逻辑说明**：
1. **参数验证**：调用`validateRequest`方法检查必填参数是否完整
2. **权限验证**：调用`verifyCodebasePermission`方法验证客户端权限并获取codebaseId
3. **目录树构建**：调用`buildDirectoryTree`方法从向量存储获取数据并构建目录树
4. **结果封装**：将构建的目录树封装为标准响应格式返回

**错误处理**：
- 参数验证失败时返回`errs.FileNotFound`
- 权限验证失败时返回`errs.FileNotFound`
- 目录树构建失败时返回具体的构建错误信息

#### 3.1.2 validateRequest 函数

**功能概述**：
验证请求参数的完整性和有效性，确保所有必填字段都已提供。

**接口定义**：
```go
func (l *CodebaseTreeLogic) validateRequest(req *types.CodebaseTreeRequest) error
```

**输入参数**：
- `req *types.CodebaseTreeRequest`: 请求参数结构体

**输出参数**：
- `error`: 验证失败时返回具体的错误信息，验证成功时返回nil

**核心逻辑说明**：
1. 检查`ClientId`字段是否为空
2. 检查`CodebasePath`字段是否为空
3. 检查`CodebaseName`字段是否为空
4. 所有检查通过则返回nil，表示验证成功

#### 3.1.3 verifyCodebasePermission 函数

**功能概述**：
验证客户端对指定代码库的访问权限，同时获取对应的codebaseId。

**接口定义**：
```go
func (l *CodebaseTreeLogic) verifyCodebasePermission(req *types.CodebaseTreeRequest) (int32, error)
```

**输入参数**：
- `req *types.CodebaseTreeRequest`: 请求参数结构体

**输出参数**：
- `int32`: 代码库ID，验证成功时返回有效的ID
- `error`: 权限验证失败时返回错误信息

**核心逻辑说明**：
1. 根据ClientId和CodebasePath查询数据库中的代码库记录
2. 如果找到匹配记录，返回真实的codebaseId
3. 如果未找到记录（MVP版本），返回模拟的codebaseId（值为1）
4. 记录详细的调试日志用于问题诊断

#### 3.1.4 buildDirectoryTree 函数

**功能概述**：
从向量存储获取文件路径数据，并构建完整的目录树结构。

**接口定义**：
```go
func (l *CodebaseTreeLogic) buildDirectoryTree(codebaseId int32, req *types.CodebaseTreeRequest) (*types.TreeNode, error)
```

**输入参数**：
- `codebaseId int32`: 代码库ID，用于从向量存储获取对应的数据
- `req *types.CodebaseTreeRequest`: 请求参数结构体

**输出参数**：
- `*types.TreeNode`: 构建完成的目录树根节点
- `error`: 构建过程中出现的错误信息

**核心逻辑说明**：
1. **数据获取**：调用向量存储的`GetCodebaseRecords`方法获取所有文件路径记录
2. **数据分析**：对获取的记录进行详细分析，包括路径格式、语言分布、内容长度等
3. **路径提取**：从记录中提取文件路径列表
4. **路径处理**：对路径进行规范化、去重等预处理
5. **目录树构建**：调用`BuildDirectoryTree`函数构建最终的目录树结构
6. **结果验证**：验证构建结果，确保所有文件都正确包含在树中

### 3.2 目录树构建工具函数

#### 3.2.1 BuildDirectoryTree 函数

**功能概述**：
核心的目录树构建函数，将文件路径列表转换为层次化的目录树结构。

**接口定义**：
```go
func BuildDirectoryTree(filePaths []string, maxDepth int, includeFiles bool) (*types.TreeNode, error)
```

**输入参数**：
- `filePaths []string`: 文件路径列表，从向量存储获取
- `maxDepth int`: 目录树最大深度限制
- `includeFiles bool`: 是否包含文件节点的标志

**输出参数**：
- `*types.TreeNode`: 构建完成的目录树根节点
- `error`: 构建过程中出现的错误信息

**核心逻辑说明**：
1. **路径预处理**：对输入路径进行规范化处理，统一路径分隔符
2. **路径去重**：去除重复的文件路径，确保每个路径只处理一次
3. **根路径提取**：调用`extractRootPath`函数提取所有路径的公共根路径
4. **根节点创建**：根据根路径创建目录树的根节点
5. **节点构建**：遍历所有文件路径，构建相应的目录和文件节点
6. **层次关系处理**：正确处理节点的父子关系，构建完整的层次结构
7. **深度过滤**：根据maxDepth参数过滤超出指定深度的节点
8. **类型过滤**：根据includeFiles参数决定是否包含文件节点

#### 3.2.2 extractRootPath 函数

**功能概述**：
从文件路径列表中提取公共根路径，作为目录树的根节点路径。

**接口定义**：
```go
func extractRootPath(filePaths []string) string
```

**输入参数**：
- `filePaths []string`: 文件路径列表

**输出参数**：
- `string`: 提取的公共根路径

**核心逻辑说明**：
1. **边界检查**：检查文件路径列表是否为空
2. **路径分析**：分析所有路径的深度分布和格式特征
3. **公共前缀计算**：使用`findCommonPrefix`函数计算所有路径的公共前缀
4. **路径修正**：根据公共前缀提取正确的根路径，处理边界情况
5. **返回结果**：返回提取的根路径，如果找不到合适的根路径则返回"."

#### 3.2.3 findCommonPrefix 函数

**功能概述**：
找到两个路径的公共前缀，用于根路径提取。

**接口定义**：
```go
func findCommonPrefix(path1, path2 string) string
```

**输入参数**：
- `path1 string`: 第一个路径
- `path2 string`: 第二个路径

**输出参数**：
- `string`: 两个路径的公共前缀

**核心逻辑说明**：
1. **路径分割**：使用系统路径分隔符分割两个路径为组件数组
2. **前缀匹配**：逐个比较路径组件，找到最后一个相同的组件
3. **结果构建**：将匹配的组件重新组合为公共前缀路径

#### 3.2.4 normalizePath 函数

**功能概述**：
统一路径格式，确保所有路径都使用系统标准的路径分隔符。

**接口定义**：
```go
func normalizePath(path string) string
```

**输入参数**：
- `path string`: 需要规范化的路径

**输出参数**：
- `string`: 规范化后的路径

**核心逻辑说明**：
1. **基本清理**：使用`filepath.Clean`进行基本路径清理
2. **分隔符统一**：使用`filepath.FromSlash`确保使用系统标准路径分隔符
3. **跨平台兼容**：在不同操作系统上都能正确处理路径分隔符

#### 3.2.5 createFileNode 函数

**功能概述**：
创建文件节点，包含文件的基本信息和元数据。

**接口定义**：
```go
func createFileNode(filePath string) (*types.TreeNode, error)
```

**输入参数**：
- `filePath string`: 文件路径

**输出参数**：
- `*types.TreeNode`: 创建的文件节点
- `error`: 创建过程中出现的错误信息

**核心逻辑说明**：
1. **路径规范化**：对输入文件路径进行规范化处理
2. **节点信息设置**：设置文件节点的名称、路径、类型等信息
3. **元数据模拟**：模拟文件大小和修改时间等元数据
4. **结果返回**：返回创建完成的文件节点

## 4. 使用场景示例

### 4.1 新项目初始化

**场景描述**：
开发人员开始一个新项目，希望快速了解项目结构和关键功能。

**操作步骤**：
1. 使用 `POST /embeddings` 接口提交项目嵌入任务
```bash
curl -X POST "http://localhost:8080/codebase-embedder/api/v1/embeddings" \
  -H "Authorization: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "user123",
    "projectPath": "/home/user/myproject",
    "codebaseName": "myproject",
    "uploadToken": "xxxx",
    "fileTotals": 42
  }' \
  --form "file=@myproject.zip"
```

2. 使用 `GET /embeddings/summary` 接口检查处理进度
```bash
curl -X GET "http://localhost:8080/codebase-embedder/api/v1/embeddings/summary?clientId=user123&projectPath=/home/user/myproject" \
  -H "Authorization: your_api_key"
```

3. 项目处理完成后，使用 `GET /search/semantic` 进行语义搜索
```bash
curl -X GET "http://localhost:8080/codebase-embedder/api/v1/search/semantic?clientId=user123&projectPath=/home/user/myproject&query=authentication%20logic&topK=5" \
  -H "Authorization: your_api_key"
```

4. 使用目录树接口查看项目结构
```bash
curl -X POST "http://localhost:8080/codebase-embedder/api/v1/codebase/tree" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "user123",
    "codebasePath": "/home/user/myproject",
    "codebaseName": "myproject"
  }'
```

### 4.2 代码库迁移

**场景描述**：
团队将旧项目迁移到新服务器，需要验证迁移后的代码库是否完整。

**操作步骤**：
1. 提交新代码库的嵌入任务
2. 使用 `POST /files/status` 接口监控处理进度
3. 比较新旧代码库的摘要信息，确保文件数量和结构一致
4. 执行相同的语义搜索查询，验证搜索结果的一致性
5. 使用目录树接口对比项目结构是否一致

### 4.3 代码审查辅助

**场景描述**：
进行代码审查时，需要快速了解相关代码的上下文。

**操作步骤**：
1. 使用语义搜索查找与审查功能相关的代码
2. 根据搜索结果中的文件路径，快速定位相关文件
3. 查看搜索结果中的代码片段，了解功能实现的上下文
4. 使用目录树接口了解项目的整体结构
5. 使用不同的查询词进行多轮搜索，全面了解代码库

### 4.4 项目结构分析

**场景描述**：
开发人员需要快速了解大型项目的结构，找到关键模块和文件。

**操作步骤**：
1. 使用目录树接口获取项目的完整结构
```bash
curl -X POST "http://localhost:8080/codebase-embedder/api/v1/codebase/tree" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "user123",
    "codebasePath": "/path/to/large/project",
    "codebaseName": "large-project",
    "includeFiles": false
  }'
```

2. 根据目录结构找到关键模块
3. 使用语义搜索查找特定功能的实现
4. 结合目录树和搜索结果全面了解项目架构

## 5. 用户操作流程

### 5.1 初始设置

1. **获取API密钥**：联系系统管理员获取访问API的密钥
2. **准备代码库**：将要分析的代码库压缩为ZIP文件
3. **确定客户端ID**：为当前设备或用户分配唯一的客户端标识

### 5.2 提交嵌入任务

1. **调用嵌入接口**：使用 `POST /embeddings` 提交代码库
   - 确保提供正确的 `clientId`、`projectPath` 和 `codebaseName`
   - 上传完整的代码库ZIP文件
   - 提供正确的 `uploadToken`

2. **监控处理进度**：
   - 使用 `GET /status` 确认服务正常
   - 使用 `GET /embeddings/summary` 或 `POST /files/status` 查询处理进度
   - 处理状态会经历 `pending` → `processing` → `completed` 的变化

3. **处理完成**：
   - 当状态变为 `completed` 时，嵌入任务完成
   - 可以开始进行语义搜索和其他操作

### 5.3 日常使用

1. **语义搜索**：
   - 使用自然语言描述要查找的功能
   - 调整 `topK` 参数控制返回结果数量
   - 根据 `score` 字段评估结果的相关性

2. **目录树浏览**：
   - 使用 `POST /codebase/tree` 获取项目结构
   - 根据需要调整 `maxDepth` 和 `includeFiles` 参数
   - 通过目录树快速定位文件和模块

3. **状态管理**：
   - 定期检查代码库状态，确保数据最新
   - 如果代码库有重大更新，重新提交嵌入任务

4. **数据清理**：
   - 使用 `DELETE /embeddings` 删除不再需要的代码库数据
   - 可以选择删除整个代码库或特定文件

### 5.4 故障排除

**常见问题及解决方案**：

1. **提交任务失败**：
   - 检查API密钥是否正确
   - 确认 `uploadToken` 有效
   - 验证代码库ZIP文件格式正确

2. **处理进度停滞**：
   - 检查服务日志，确认没有错误
   - 使用 `GET /status` 确认服务正常运行
   - 重启嵌入任务

3. **搜索结果不相关**：
   - 尝试不同的查询词
   - 降低 `scoreThreshold` 以获得更多信息
   - 重新提交嵌入任务，确保代码库已完全处理

4. **目录树显示异常**：
   - 检查向量存储中是否有数据
   - 验证 `clientId` 和 `codebasePath` 是否正确
   - 确认代码库已成功处理完成

5. **权限问题**：
   - 检查 `clientId` 是否有效
   - 确认有权限访问指定的代码库
   - 联系管理员确认权限配置

通过遵循上述操作流程，用户可以充分利用Codebase Embedder的功能，提高代码理解和开发效率。