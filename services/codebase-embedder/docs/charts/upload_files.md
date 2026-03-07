# /api/v1/files/upload 接口流程图

## .shenma_sync 元数据格式

`.shenma_sync` 文件是 ZIP 包中的同步元数据文件，用于描述文件操作类型和相关信息。支持两种不同的 JSON 格式：

## 格式一：对象格式

```json
{
  "clientId": "客户端标识",
  "codebasePath": "代码库路径",
  "codebaseName": "代码库名称",
  "extraMetadata": {
    "自定义键值对": "值"
  },
  "fileList": {
    "文件路径1": "add",
    "文件路径2": "modify",
    "文件路径3": "delete"
  },
  "timestamp": 时间戳
}
```

### 格式二：数组格式

```json
{
  "clientId": "客户端标识",
  "codebasePath": "代码库路径",
  "codebaseName": "代码库名称",
  "extraMetadata": {
    "自定义键值对": "值"
  },
  "fileList": [
    {
      "path": "文件路径1",
      "targetPath": "",
      "hash": "文件哈希值1",
      "status": "add",
      "requestId": ""
    },
    {
      "path": "文件路径2",
      "targetPath": "",
      "hash": "文件哈希值2",
      "status": "modify",
      "requestId": ""
    },
    {
      "path": "文件路径3",
      "targetPath": "",
      "hash": "文件哈希值3",
      "status": "delete",
      "requestId": ""
    }
    {
      "path": "test/codegraph/java_test.go",
      "targetPath": "test/codegraph/java_test2.go",
      "hash": "1754987375674",
      "status": "rename",
      "requestId": ""
    },
    {
      "path": "test/codegraph",
      "targetPath": "test/codegraph2",
      "hash": "1754987375674",
      "status": "rename",
      "requestId": ""
    }
  ],

  "timestamp": 时间戳
}
```

### fileList 支持的操作类型：
- **add**: 添加新文件到向量数据库
- **modify**: 修改现有文件的向量数据
- **delete**: 从向量数据库删除文件
- **rename**: 重命名文件或目录，包括删除源文件向量并为目标文件创建新向量

### 格式说明：
- **格式一** 使用键值对方式，键为文件路径，值为操作类型
- **格式二** 使用数组方式，每个元素包含文件详细信息：
  - `path`: 文件路径（必需）
  - `targetPath`: 目标路径（可选）
  - `hash`: 文件哈希值（可选）
  - `status`: 操作类型（优先使用）
  - `operate`: 操作类型（备用字段，当status不存在时使用）
  - `requestId`: 请求ID（可选）

## 接口处理流程

```mermaid
flowchart TD
    A[入口 /codebase-embedder/api/v1/files/upload<br>POST 方法] --> B[解析multipart/form-data]
    B --> C[提取请求参数]
    C --> D[验证必填字段]
    D --> E{验证是否通过?}
    E -->|否| F[返回错误响应]
    E -->|是| G[验证uploadToken]
    G --> H{Token验证是否通过?}
    H -->|否| I[返回错误响应]
    H -->|是| J[解析用户信息]
    J --> K[初始化代码库记录]
    K --> L{初始化是否成功?}
    L -->|否| M[返回错误响应]
    L -->|是| N[获取分布式锁]
    N --> O{获取锁是否成功?}
    O -->|否| P[返回错误响应]
    O -->|是| Q[处理上传的ZIP文件]
    Q --> R{ZIP处理是否成功?}
    R -->|否| S[返回错误响应]
    R -->|是| T[解析.shenma_sync元数据]
    T --> U[分类任务: 添加/删除/修改/重命名]
    U --> V[初始化任务状态为处理中]
    V --> W{是否有删除任务?}
    W -->|是| X[从向量数据库删除文件]
    W -->|否| DD[检查是否有重命名任务?]
    X --> DD
    DD -->|是| EE[执行重命名任务]
    DD -->|否| Y[更新代码库信息]
    EE --> FF[重命名任务详细流程]
    FF --> GG[循环处理每个重命名任务]
    GG --> HH[删除源文件向量]
    HH --> II{目标文件是否在ZIP中?}
    II -->|是| JJ[为目标文件创建新向量]
    II -->|否| KK[仅删除源文件向量]
    JJ --> LL[更新任务状态为完成]
    KK --> LL
    LL --> MM{是否还有重命名任务?}
    MM -->|是| GG
    MM -->|否| Y[更新代码库信息]
    Y --> Z{文件数量是否为0?}
    Z -->|是| AA[直接标记任务完成]
    Z -->|否| BB[提交索引任务到任务池]
    AA --> CC[返回任务ID响应]
    BB --> CC
```