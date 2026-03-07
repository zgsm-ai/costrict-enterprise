# /codebase-embedder/api/v1/embeddings 接口文档

## 接口概述

更新嵌入路径接口，用于修改代码库中文件或目录的路径，并更新相应的向量嵌入数据。

## 基本信息

- **接口名称**: 更新嵌入路径
- **请求方法**: PUT
- **接口路径**: `/codebase-embedder/api/v1/embeddings`
- **内容类型**: application/json

## 请求参数

### 请求体

```json
{
  "clientId": "string",
  "codebasePath": "string", 
  "oldPath": "string",
  "newPath": "string"
}
```

### 参数说明

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| clientId | string | 是 | 客户端唯一标识（如MAC地址） |
| codebasePath | string | 是 | 项目绝对路径 |
| oldPath | string | 是 | 旧路径（文件或目录的相对路径） |
| newPath | string | 是 | 新路径（文件或目录的相对路径） |

## 响应结果

### 成功响应

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "modifiedFiles": [
      "src/example/file1.js",
      "src/example/file2.js"
    ],
    "totalFiles": 2
  }
}
```

### 响应参数说明

| 参数名 | 类型 | 说明 |
|--------|------|------|
| code | int | 响应状态码，0表示成功 |
| msg | string | 响应消息 |
| data | object | 响应数据 |

#### data 对象说明

| 参数名 | 类型 | 说明 |
|--------|------|------|
| modifiedFiles | string[] | 修改的文件列表 |
| totalFiles | int | 总共修改的文件数 |

### 错误响应

```json
{
  "code": 400,
  "msg": "missing required parameter: clientId"
}
```

### 错误码说明

| 错误码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | missing required parameter: xxx | 缺少必填参数 |
| 500 | internal server error | 服务器内部错误 |

## 请求示例

### cURL 示例

```bash
curl -X PUT "http://localhost:8080/codebase-embedder/api/v1/embeddings" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "00:1A:2B:3C:4D:5E",
    "codebasePath": "/home/user/my-project",
    "oldPath": "src/old-dir",
    "newPath": "src/new-dir"
  }'
```

### JavaScript 示例

```javascript
const response = await fetch('http://localhost:8080/codebase-embedder/api/v1/embeddings', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    clientId: '00:1A:2B:3C:4D:5E',
    codebasePath: '/home/user/my-project',
    oldPath: 'src/old-dir',
    newPath: 'src/new-dir'
  })
});

const result = await response.json();
console.log(result);
```

## 注意事项

1. 所有路径参数都会被规范化为使用正斜杠（`/`）作为分隔符
2. `oldPath` 和 `newPath` 是相对于 `codebasePath` 的路径
3. 该接口会更新向量数据库中与指定路径相关的嵌入数据
4. 如果路径不存在或没有权限，接口会返回相应的错误信息

## 处理流程

1. 验证必填参数（clientId、codebasePath、oldPath、newPath）
2. 规范化路径格式
3. 在向量数据库中查找并更新与旧路径相关的嵌入记录
4. 返回修改的文件列表和总数

## 相关接口

- DELETE `/codebase-embedder/api/v1/embeddings` - 删除嵌入数据
- GET `/codebase-embedder/api/v1/embeddings/summary` - 获取嵌入数据摘要
- GET `/codebase-embedder/api/v1/embeddings/vectors-summary` - 获取向量汇总信息