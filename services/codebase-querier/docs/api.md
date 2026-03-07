# Codebase Querier API 接口文档

## 概述
Codebase Querier 是一个代码库查询和分析工具，提供代码结构分析、语义搜索、关系追踪等功能。所有接口均以 `/codebase-indexer` 为前缀。

## 基础信息
- **Base URL**: `http://localhost:8080/codebase-indexer`
- **数据格式**: JSON
- **编码**: UTF-8

## 通用参数说明
- `clientId`: 用户机器ID，用于标识不同用户
- `codebasePath`: 项目绝对路径，用于定位代码库

---

## 1. 代码关系查询接口

### 1.1 查询符号关系
获取指定代码符号的引用关系树。

**接口地址**: `GET /api/v1/search/relation`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| filePath | string | 是 | 文件相对路径 |
| startLine | int | 是 | 开始行（从1开始） |
| startColumn | int | 是 | 开始列（从1开始） |
| endLine | int | 是 | 结束行（从1开始） |
| endColumn | int | 是 | 结束列（从1开始） |
| symbolName | string | 否 | 符号名（可选） |
| includeContent | int | 否 | 是否返回代码内容（1=是，0=否，默认0） |
| maxLayer | int | 否 | 最大层级数（默认10） |

**响应示例**:
```json
{
  "list": [
    {
      "content": "function getUserData() { ... }",
      "nodeType": "definition",
      "filePath": "src/user.js",
      "position": {
        "startLine": 10,
        "startColumn": 1,
        "endLine": 20,
        "endColumn": 5
      },
      "children": [
        {
          "content": "getUserData()",
          "nodeType": "reference",
          "filePath": "src/app.js",
          "position": {
            "startLine": 50,
            "startColumn": 15,
            "endLine": 50,
            "endColumn": 28
          },
          "children": []
        }
      ]
    }
  ]
}
```

---

## 2. 定义查询接口

### 2.1 查询符号定义
根据代码片段或位置查询符号的定义信息。

**接口地址**: `GET /api/v1/search/definition`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| filePath | string | 否 | 文件相对路径 |
| startLine | int | 否 | 开始行 |
| endLine | int | 否 | 结束行 |
| codeSnippet | string | 否 | 代码内容 |

**响应示例**:
```json
{
  "list": [
    {
      "name": "getUserData",
      "content": "function getUserData(userId) { ... }",
      "type": "function",
      "filePath": "src/user.js",
      "position": {
        "startLine": 10,
        "startColumn": 1,
        "endLine": 20,
        "endColumn": 5
      }
    }
  ]
}
```

---

## 3. 文件结构查询接口

### 3.1 获取文件结构
获取指定文件的结构信息，包括类、函数、变量等定义。

**接口地址**: `GET /api/v1/file/structure`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| filePath | string | 是 | 文件相对路径 |

**响应示例**:
```json
{
  "list": [
    {
      "name": "UserService",
      "type": "class",
      "position": {
        "startLine": 5,
        "startColumn": 1,
        "endLine": 50,
        "endColumn": 3
      },
      "content": "class UserService { ... }"
    },
    {
      "name": "getUserById",
      "type": "method",
      "position": {
        "startLine": 15,
        "startColumn": 3,
        "endLine": 25,
        "endColumn": 5
      },
      "content": "getUserById(id) { ... }"
    }
  ]
}
```

---

## 4. 文件内容接口

### 4.1 获取文件内容
获取指定文件的指定行范围内容。

**接口地址**: `GET /api/v1/files/content`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| filePath | string | 是 | 文件相对路径 |
| startLine | int | 否 | 开始行（默认1） |
| endLine | int | 否 | 结束行（默认100，-1=全部） |

**响应示例**:
```
// 返回原始文件内容
function getUserData(userId) {
  return fetch(`/api/users/${userId}`)
    .then(res => res.json());
}
```

---

## 5. 代码库管理接口

### 5.1 获取代码库目录树
获取代码库的目录结构信息。

**接口地址**: `GET /api/v1/codebases/directory`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| subDir | string | 否 | 子目录路径 |
| depth | int | 否 | 目录深度 |
| includeFiles | int | 否 | 是否包含文件 |

**响应示例**:
```json
{
  "codebaseId": 123,
  "name": "my-project",
  "rootPath": "/path/to/project",
  "totalFiles": 150,
  "totalSize": 1048576,
  "directoryTree": {
    "name": "my-project",
    "isDir": true,
    "path": "/",
    "children": [
      {
        "name": "src",
        "isDir": true,
        "path": "/src",
        "children": [
          {
            "name": "index.js",
            "isDir": false,
            "path": "/src/index.js",
            "size": 1024,
            "language": "javascript"
          }
        ]
      }
    ]
  }
}
```

### 5.2 获取代码库哈希值
获取代码库所有文件的哈希值，用于对比代码库变化。

**接口地址**: `GET /api/v1/codebases/hash`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |

**响应示例**:
```json
{
  "list": [
    {
      "path": "src/index.js",
      "hash": "a1b2c3d4e5f6"
    },
    {
      "path": "src/utils.js",
      "hash": "b2c3d4e5f6g7"
    }
  ]
}
```

---

## 6. 语义搜索接口

### 6.1 语义代码搜索
基于自然语言查询进行语义代码搜索。

**接口地址**: `GET /api/v1/search/semantic`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| query | string | 是 | 查询内容 |
| topK | int | 否 | 结果返回数量（默认10） |

**响应示例**:
```json
{
  "list": [
    {
      "content": "function validateUserInput(data) {\n  if (!data.email) return false;\n  return data.email.includes('@');\n}",
      "filePath": "src/validation.js",
      "score": 0.95
    },
    {
      "content": "const emailRegex = /^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/;",
      "filePath": "src/utils.js",
      "score": 0.87
    }
  ]
}
```

---

## 7. 索引管理接口

### 7.1 获取索引摘要
获取代码库的索引状态摘要信息。

**接口地址**: `GET /api/v1/index/summary`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |

**响应示例**:
```json
{
  "totalFiles": 150,
  "lastSyncAt": "2024-01-15T10:30:00Z",
  "embedding": {
    "status": "completed",
    "lastSyncAt": "2024-01-15T10:30:00Z",
    "totalFiles": 150,
    "totalChunks": 1200
  },
  "codegraph": {
    "status": "completed",
    "lastSyncAt": "2024-01-15T10:25:00Z",
    "totalFiles": 150
  }
}
```

### 7.2 创建索引任务
创建代码库索引任务。

**接口地址**: `POST /api/v1/index/task`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| indexType | string | 是 | 索引类型（embedding\|codegraph\|all） |
| fileMap | object | 否 | 文件映射 |

**响应示例**:
```json
{
  "taskId": 12345
}
```

### 7.3 删除索引
删除指定类型的索引。

**接口地址**: `DELETE /api/v1/index`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |
| indexType | string | 是 | 索引类型（embedding\|codegraph\|all） |

**响应示例**:
```json
{}
```

### 7.4 删除代码库
删除整个代码库及其索引。

**接口地址**: `DELETE /api/v1/codebase`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 用户机器ID |
| codebasePath | string | 是 | 项目绝对路径 |

**响应示例**:
```json
{}
```

---

## 8. 文件上传接口

### 8.1 同步文件
上传并同步代码库文件。

**接口地址**: `POST /api/v1/files/upload`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 客户ID |
| codebasePath | string | 是 | 项目路径 |
| codebaseName | string | 是 | 项目名称 |
| extraMetadata | string | 否 | 额外元数据（JSON字符串） |

**响应示例**:
```json
{
  "message": "Files uploaded successfully"
}
```

---

## 错误处理

### 错误响应格式
```json
{
  "code": 400,
  "message": "Invalid request parameters",
  "details": "codebasePath is required"
}
```

### 常见错误码
| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 404 | 资源未找到 |
| 500 | 服务器内部错误 |

---

## 使用示例

### 1. 查询函数定义
```bash
curl -X GET "http://localhost:8080/codebase-indexer/api/v1/search/definition?clientId=mac123&codebasePath=/path/to/project&filePath=src/main.js&startLine=10&endLine=20"
```

### 2. 语义搜索
```bash
curl -X GET "http://localhost:8080/codebase-indexer/api/v1/search/semantic?clientId=mac123&codebasePath=/path/to/project&query=用户认证功能&topK=5"
```

### 3. 获取文件内容
```bash
curl -X GET "http://localhost:8080/codebase-indexer/api/v1/files/content?clientId=mac123&codebasePath=/path/to/project&filePath=src/utils.js&startLine=1&endLine=50"