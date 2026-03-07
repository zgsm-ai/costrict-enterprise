# Codebase Embedder API 文档

## 1. 服务简介
Codebase Embedder 提供代码库嵌入管理、语义搜索和项目摘要功能。主要功能包括：
- 提交代码嵌入任务
- 管理嵌入数据
- 获取代码库摘要
- 执行语义代码搜索
- 服务状态检查

## 2. 认证方法
使用 API Key 认证：
- 在请求头中添加 `Authorization` 字段
- 认证方式：`apiKey`
- 示例：
  ```http
  GET /status HTTP/1.1
  Authorization: your_api_key_here
  ```

## 3. 完整端点列表

| 方法   | 端点路径                                  | 功能描述               |
|--------|-------------------------------------------|------------------------|
| POST   | /codebase-embedder/api/v1/embeddings      | 提交嵌入任务           |
| DELETE | /codebase-embedder/api/v1/embeddings      | 删除嵌入数据           |
| GET    | /codebase-embedder/api/v1/embeddings/summary | 获取代码库摘要信息     |
| GET    | /codebase-embedder/api/v1/status          | 服务状态检查           |
| POST   | /codebase-embedder/api/v1/files/status     | 查询文件处理状态       |
| GET    | /codebase-embedder/api/v1/search/semantic | 执行语义代码搜索       |
| POST   | /codebase-embedder/api/v1/files/upload     | 上传文件               |
| POST   | /codebase-embedder/api/v1/codebase/query  | 查询代码库信息         |

## 4. 端点详细说明

### 4.1 提交嵌入任务 (POST /embeddings)

**请求格式**：`multipart/form-data`

**请求参数**：

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| clientId | string | 是 | 无 | 客户端唯一标识（如MAC地址） | "user_machine_id" |
| codebasePath | string | 是 | 无 | 项目绝对路径 | "/absolute/path/to/project" |
| codebaseName | string | 是 | 无 | 项目名称 | "my_project" |
| uploadToken | string | 否 | "" | 上传令牌 | "upload_token_123" |
| extraMetadata | string | 否 | "" | 额外元数据（JSON字符串） | '{"version": "1.0", "author": "dev"}' |
| fileTotals | int | 是 | 无 | 上传工程文件总数 | 42 |

**请求示例**：
```http
POST /codebase-embedder/api/v1/embeddings

RequestId: xxxxxxxxxxxx
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW



------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="clientId"

user_machine_id
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="codebasePath"

/absolute/path/to/project
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="codebaseName"

my_project
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="fileTotals"

42
------WebKitFormBoundary7MA4YWxkTrZu0gW
```

**成功响应**：
```json
HTTP/1.1 200 OK
{
  "taskId": 12345
}
```

**错误响应**：
```json
HTTP/1.1 400 Bad Request
{
  "code": 400,
  "message": "缺少必需参数: clientId"
}
```

### 4.2 删除嵌入数据 (DELETE /embeddings)

**请求格式**：`application/x-www-form-urlencoded`

**请求参数**：

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| clientId | string | 是 | 无 | 客户端唯一标识（如MAC地址） | "user_machine_id" |
| codebasePath | string | 是 | 无 | 项目绝对路径 | "/absolute/path/to/project" |
| filePaths | string | 否 | 无 | 要删除的文件路径file1.js 如果不传则根据clientId、codebasePath 删除工程 |

**请求示例**：
```http
DELETE /codebase-embedder/api/v1/embeddings
Content-Type: application/x-www-form-urlencoded

clientId=user_machine_id&codebasePath=/project/path&filePaths=file1.js
```

**成功响应**：
```json
HTTP/1.1 200 OK
{}
```

**错误响应**：
```json
HTTP/1.1 400 Bad Request
{
  "code": 400,
  "message": "缺少必需参数: filePaths"
}
```

### 4.3 获取代码库摘要 (GET /embeddings/summary)

**请求格式**：查询参数

**请求参数**：

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| clientId | string | 是 | 无 | 客户端唯一标识（如MAC地址） | "user_machine_id" |
| codebasePath | string | 是 | 无 | 项目绝对路径 | "/absolute/path/to/project" |

**请求示例**：
```http
GET /codebase-embedder/api/v1/embeddings/summary?clientId=user_machine_id&codebasePath=/project/path
```

**成功响应**：
```json
HTTP/1.1 200 OK
{
  "totalFiles": 42,
  "lastSyncAt": "2025-07-28T12:00:00Z",
  "embedding": {
    "status": "completed",
    "updatedAt": "2025-07-28T12:00:00Z",
    "totalFiles": 42,
    "totalChunks": 156
  }
}
```

**响应字段说明**：

| 字段名 | 类型 | 描述 |
|--------|------|------|
| totalFiles | int | 项目总文件数 |
| lastSyncAt | string | 最后同步时间（ISO 8601格式） |
| embedding.status | string | 嵌入状态（pending/processing/completed/failed） |
| embedding.updatedAt | string | 嵌入更新时间（ISO 8601格式） |
| embedding.totalFiles | int | 已嵌入文件总数 |
| embedding.totalChunks | int | 嵌入块总数 |

**错误响应**：
```json
HTTP/1.1 404 Not Found
{
  "code": 404,
  "message": "未找到指定的嵌入任务"
}
```

### 4.4 服务状态检查 (GET /status)

**请求格式**：无参数

**请求示例**：
```http
GET /codebase-embedder/api/v1/status
```

**成功响应**：
```json
HTTP/1.1 200 OK
{
  "status": "ok",
  "version": "1.0.0"
}
```

**响应字段说明**：

| 字段名 | 类型 | 描述 |
|--------|------|------|
| status | string | 服务状态（ok/error） |
| version | string | 服务版本号 |

### 4.5 语义代码搜索 (GET /search/semantic)

**请求格式**：查询参数

**请求参数**：

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| clientId | string | 是 | 无 | 客户端唯一标识（如MAC地址） | "user_machine_id" |
| codebasePath | string | 是 | 无 | 项目绝对路径 | "/absolute/path/to/project" |
| query | string | 是 | 无 | 查询内容（需要进行URL编码） | "authentication logic" |
| topK | int | 否 | 10 | 结果返回数量 | 5 |
| scoreThreshold | float32 | 否 | 0.3 | 分数阈值（0-1之间） | 0.5 |

**请求示例**：
```http
GET /codebase-embedder/api/v1/search/semantic?clientId=user_machine_id&codebasePath=/project/path&query=authentication+logic&topK=5&scoreThreshold=0.5
```

**成功响应**：
```json
HTTP/1.1 200 OK
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

**响应字段说明**：

| 字段名 | 类型 | 描述 |
|--------|------|------|
| list | array | 检索结果列表 |
| list[].content | string | 代码片段内容 |
| list[].filePath | string | 文件相对路径 |
| list[].score | float32 | 匹配得分（0-1之间） |

**错误响应**：
```json
HTTP/1.1 404 Not Found
{
  "code": 404,
  "message": "未找到指定的项目"
}
```

### 4.6 文件状态查询 (POST /files/status)

**请求格式**：JSON

**请求参数**：

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| clientId | string | 是 | 无 | 客户端唯一标识（如MAC地址） | "user_machine_id" |
| codebasePath | string | 是 | 无 | 项目绝对路径 | "/absolute/path/to/project" |
| codebaseName | string | 是 | 无 | 项目名称 | "project_name" |
| syncId | string | 是 | 无 | 上传接口的RequestId | "xxxxxxx-xxxxxxxxx-xxxxx" |

**请求示例**：
```json
POST /codebase-embedder/api/v1/files/status
{
  "clientId": "user_machine_id",
  "codebasePath": "/absolute/path/to/project",
  "codebaseName": "project_name",
  "chunkNumber": 0,
  "totalChunks": 1
}
```

**成功响应**：
```json
HTTP/1.1 200 OK
{
    "process": "processing",
    "totalProgress": 50,
    "fileList": [
        {
         "path": "src/main/java/main.java",
         "status": "complete",
         "operate": "add"
        },
        {
         "path": "src/main/java/server.java",
         "status": "complete",
         "operate": "modify"
        },
        {
         "path": "src/main/java/old.java",
         "status": "complete",
         "operate": "delete"
        }
   ]
}
```

**响应字段说明**：

| 字段名 | 类型 | 描述 |
|--------|------|------|
| process | string | 整体提取状态（pending/processing/complete/failed） |
| totalProgress | int | 当前分片整体提取进度（百分比，0-100） |
| fileList | array | 文件状态列表 |
| fileList[].path | string | 文件相对路径 |
| fileList[].status | string | 单个文件状态（pending/processing/complete/failed） |
| fileList[].operate | string | 文件操作类型（add/modify/delete） |

**错误响应**：
```json
HTTP/1.1 404 Not Found
{
  "code": 404,
  "message": "未找到指定的嵌入任务"
}
```

### 4.7 代码库查询 (POST /codebase/query)

**请求格式**：JSON

**请求参数**：

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| clientId | string | 是 | 无 | 客户端唯一标识（如MAC地址） | "user_machine_id" |
| codebasePath | string | 是 | 无 | 项目绝对路径 | "/absolute/path/to/project" |
| codebaseName | string | 是 | 无 | 项目名称 | "project_name" |

**请求示例**：
```json
POST /codebase-embedder/api/v1/codebase/query
{
  "clientId": "user_machine_id",
  "codebasePath": "/absolute/path/to/project",
  "codebaseName": "project_name"
}
```

**成功响应**：
```json
HTTP/1.1 200 OK
{
  "codebaseId": 12345,
  "codebaseName": "project_name",
  "codebasePath": "/absolute/path/to/project",
  "summary": {
    "totalFiles": 42,
    "totalChunks": 156,
    "lastUpdateTime": "2025-07-28T12:00:00Z",
    "indexStatus": "completed",
    "indexProgress": 100
  },
  "languageDistribution": [
    {
      "language": "JavaScript",
      "fileCount": 20,
      "chunkCount": 78,
      "percentage": 47.6
    },
    {
      "language": "Python",
      "fileCount": 15,
      "chunkCount": 65,
      "percentage": 35.7
    }
  ],
  "recentFiles": [
    {
      "filePath": "src/main.js",
      "lastIndexed": "2025-07-28T12:00:00Z",
      "chunkCount": 5,
      "fileSize": 2048
    }
  ],
  "indexStats": {
    "averageChunkSize": 256,
    "maxChunkSize": 512,
    "minChunkSize": 128,
    "totalVectors": 156
  },
  "records": [
    {
      "id": "record_001",
      "filePath": "src/auth.js",
      "language": "JavaScript",
      "content": "function authenticateUser() {...}",
      "range": [10, 0, 15, 1],
      "tokenCount": 42,
      "lastUpdated": "2025-07-28T12:00:00Z"
    }
  ]
}
```

**响应字段说明**：

| 字段名 | 类型 | 描述 |
|--------|------|------|
| codebaseId | int | 代码库ID |
| codebaseName | string | 代码库名称 |
| codebasePath | string | 代码库路径 |
| summary | object | 代码库摘要信息 |
| summary.totalFiles | int | 总文件数 |
| summary.totalChunks | int | 总块数 |
| summary.lastUpdateTime | string | 最后更新时间（ISO 8601格式） |
| summary.indexStatus | string | 索引状态（pending/processing/completed/failed） |
| summary.indexProgress | int | 索引进度（百分比，0-100） |
| languageDistribution | array | 语言分布信息 |
| languageDistribution[].language | string | 编程语言 |
| languageDistribution[].fileCount | int | 该语言文件数 |
| languageDistribution[].chunkCount | int | 该语言块数 |
| languageDistribution[].percentage | float | 该语言占比（百分比） |
| recentFiles | array | 最近文件信息 |
| recentFiles[].filePath | string | 文件路径 |
| recentFiles[].lastIndexed | string | 最后索引时间（ISO 8601格式） |
| recentFiles[].chunkCount | int | 文件块数 |
| recentFiles[].fileSize | int | 文件大小（字节） |
| indexStats | object | 索引统计信息 |
| indexStats.averageChunkSize | int | 平均块大小 |
| indexStats.maxChunkSize | int | 最大块大小 |
| indexStats.minChunkSize | int | 最小块大小 |
| indexStats.totalVectors | int | 总向量数 |
| records | array | 详细记录列表 |
| records[].id | string | 记录ID |
| records[].filePath | string | 文件路径 |
| records[].language | string | 编程语言 |
| records[].content | string | 代码内容 |
| records[].range | array | 代码范围 [startLine, startColumn, endLine, endColumn] |
| records[].tokenCount | int | Token数量 |
| records[].lastUpdated | string | 最后更新时间（ISO 8601格式） |

**错误响应**：
```json
HTTP/1.1 404 Not Found
{
  "code": 404,
  "message": "未找到指定的代码库"
}
```

### 4.8 文件上传接口 (POST /files/upload)

**请求格式**：`multipart/form-data`

**请求参数**：

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| clientId | string | 是 | 无 | 客户端唯一标识（如MAC地址） | "user_machine_id" |
| codebasePath | string | 是 | 无 | 项目绝对路径 | "/absolute/path/to/project" |
| codebaseName | string | 是 | 无 | 项目名称 | "project_name" |
| uploadToken | string | 是 | 无 | 上传令牌 | "upload_token_123" |
| extraMetadata | string | 否 | "" | 额外元数据（JSON字符串） | '{"version": "1.0", "author": "dev"}' |
| chunkNumber | int | 否 | 0 | 当前分片编号（从0开始） | 0 |
| totalChunks | int | 否 | 1 | 分片总数 | 1 |
| fileTotals | int | 是 | 无 | 上传工程文件总数 | 42 |
| file | file | 是 | 无 | 要上传的文件 | - |

**请求示例**：
```http
POST /codebase-embedder/api/v1/files/upload
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW

------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="clientId"

user_machine_id
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="codebasePath"

/absolute/path/to/project
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="codebaseName"

project_name
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="uploadToken"

upload_token_123
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="fileTotals"

42
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="file"; filename="example.js"
Content-Type: application/javascript

// JavaScript file content
function example() {
  console.log("Hello, World!");
}
------WebKitFormBoundary7MA4YWxkTrZu0gW--
```

**成功响应**：
```json
HTTP/1.1 200 OK
{
  "taskId": 12345
}
```

**错误响应**：
```json
HTTP/1.1 400 Bad Request
{
  "code": 400,
  "message": "缺少必需参数: clientId"
}
```

## 5. 标准错误码表

| 错误码 | 含义               | 可能原因                     |
|--------|--------------------|------------------------------|
| 400    | 无效请求参数       | 缺少必需参数/参数格式错误    |
| 401    | 未授权访问         | API Key缺失或无效            |
| 404    | 资源不存在         | 项目/嵌入数据不存在          |
| 500    | 服务器内部错误     | 服务端处理异常               |

**错误响应示例**：
```json
HTTP/1.1 400 Bad Request
{
  "code": 400,
  "message": "缺少必需参数: clientId"
}