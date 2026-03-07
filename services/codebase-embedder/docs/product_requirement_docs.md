# 产品需求文档：正在执行任务状态查询GET接口

## 1. 需求概述

本文档基于用户需求"添加一个GET接口，获取正在执行的任务情况，数据从redis里面读取"，使用 EARS（Easy Approach to Requirements Syntax）需求语法详细描述新接口的功能需求。该接口专门用于查询系统中所有客户端的所有正在执行的任务状态，数据存储在Redis中。

## 2. EARS 需求描述

### 2.1 需求类型
**请求驱动型需求 (Request-Driven Requirement)**

### 2.2 详细需求描述

**事件 (Event):**
当客户端发送GET请求查询系统中所有正在执行的任务状态时触发。

**控制条件 (Control):**
系统必须验证Redis服务可用，并能够扫描Redis中所有符合条件的状态数据。

**状态 (State):**
- Redis服务正常运行
- 任务状态数据已在Redis中以"request:id:"为前缀存储
- 系统服务正常运行

**响应 (Response):**
系统必须执行以下操作序列：

1. **Redis连接检查**
   - 验证Redis服务连接状态
   - 如果Redis服务不可用，返回相应错误信息

2. **任务状态扫描**
   - 扫描Redis中所有以"request:id:"为前缀的键
   - 获取所有任务的状态数据
   - 过滤出状态为pending、processing、running的任务

3. **数据聚合与处理**
   - 解析每个任务的状态数据
   - 聚合所有符合条件的任务信息
   - 按照时间或其他逻辑排序任务列表

4. **结果返回**
   - 将所有查询结果结构化为JSON格式
   - 包含任务总数和任务列表
   - 每个任务包含完整的详细信息（任务ID、客户端ID、状态、进度、时间戳等）

## 3. 接口设计

### 3.1 接口基本信息

**接口路径:** `GET /codebase-embedder/api/v1/tasks/running`

**请求格式:** 无请求体

**响应格式:** `application/json`

### 3.2 请求参数

| 参数名 | 类型 | 是否必填 | 默认值 | 描述 | 示例值 |
|--------|------|----------|--------|------|--------|
| 无 | 无 | 无 | 无 | 该接口不需要任何参数 | 无 |

### 3.3 成功响应示例

```json
{
  "code": 0,
  "message": "ok",
  "success": true,
  "data": {
    "totalTasks": 3,
    "tasks": [
      {
        "taskId": "request:id:uuid-generated-task-id-1",
        "clientId": "user_machine_id_1",
        "status": "running",
        "process": "embedding",
        "totalProgress": 65,
        "startTime": "2025-08-06T15:20:00Z",
        "lastUpdateTime": "2025-08-06T15:25:30Z",
        "estimatedCompletionTime": "2025-08-06T15:30:00Z",
        "fileList": [
          {
            "path": "src/main.js",
            "status": "complete",
            "operate": "add"
          },
          {
            "path": "src/utils/helper.js",
            "status": "processing",
            "operate": "modify"
          }
        ]
      },
      {
        "taskId": "request:id:uuid-generated-task-id-2",
        "clientId": "user_machine_id_2",
        "status": "pending",
        "process": "indexing",
        "totalProgress": 0,
        "startTime": "2025-08-06T15:25:00Z",
        "lastUpdateTime": "2025-08-06T15:25:00Z",
        "fileList": []
      },
      {
        "taskId": "request:id:uuid-generated-task-id-3",
        "clientId": "user_machine_id_1",
        "status": "processing",
        "process": "embedding",
        "totalProgress": 30,
        "startTime": "2025-08-06T15:15:00Z",
        "lastUpdateTime": "2025-08-06T15:22:00Z",
        "fileList": [
          {
            "path": "src/components/Header.jsx",
            "status": "complete",
            "operate": "add"
          },
          {
            "path": "src/components/Footer.jsx",
            "status": "pending",
            "operate": "add"
          }
        ]
      }
    ]
  }
}
```

### 3.4 错误响应示例

```json
{
  "code": 503,
  "message": "Redis服务不可用，请稍后再试",
  "success": false
}
```

```json
{
  "code": 500,
  "message": "查询任务状态时发生内部错误",
  "success": false
}
```

## 4. 功能边界

### 4.1 包含范围
- 查询系统中所有客户端的所有正在执行任务状态
- 包含状态为pending、processing、running的所有任务
- 返回完整的任务详细信息，包括任务ID、客户端ID、状态、进度、时间戳等
- 提供文件级别的处理状态信息
- 支持按时间或其他逻辑排序任务列表

### 4.2 排除范围
- 不修改任务状态
- 不取消或暂停任务
- 不提供任务历史记录查询
- 不处理任务创建
- 不支持分页功能
- 不支持过滤功能
- 不返回任务执行的详细日志

## 5. 验收标准

### 5.1 功能验收
- [ ] 能够正确检查Redis连接状态
- [ ] 能够正确扫描Redis中所有以"request:id:"为前缀的键
- [ ] 能够正确过滤出状态为pending、processing、running的任务
- [ ] 能够正确解析和返回完整的任务详细信息
- [ ] 能够正确处理各种错误情况

### 5.2 性能验收
- [ ] 对于少量任务（<10个），响应时间应在200ms内
- [ ] 对于中等数量任务（10-50个），响应时间应在500ms内
- [ ] 对于大量任务（>50个），响应时间应在2s内
- [ ] Redis扫描操作不应阻塞其他系统操作

### 5.3 兼容性验收
- [ ] 接口设计符合现有API风格
- [ ] 错误处理机制与现有系统一致
- [ ] 响应格式符合现有系统规范
- [ ] 与现有Redis状态管理机制兼容

## 6. 风险评估

### 6.1 技术风险
- **Redis可用性风险**: Redis服务不可用将导致接口无法正常工作
  - 缓解措施: 实现Redis连接池和重试机制，提供优雅的错误处理
  
- **性能风险**: 大量任务同时存在可能导致Redis扫描操作耗时过长
  - 缓解措施: 优化Redis扫描策略，考虑使用SCAN命令替代KEYS命令

- **内存风险**: 大量任务数据同时加载到内存可能导致内存占用过高
  - 缓解措施: 实现流式处理或分批加载机制

### 6.2 业务风险
- **数据敏感性风险**: 返回所有客户端的任务信息可能涉及数据隐私
  - 缓解措施: 确保接口有适当的访问控制和权限验证

- **系统负载风险**: 频繁查询可能影响系统整体性能
  - 缓解措施: 实现查询频率限制，建议客户端使用合理的轮询间隔

## 7. 后续扩展计划

### 7.1 短期扩展
- 支持按客户端ID过滤任务
- 支持按任务状态过滤任务
- 支持按时间范围过滤任务
- 支持分页查询大量任务

### 7.2 长期扩展
- 支持任务状态变更通知（WebSocket）
- 支持任务执行时间统计
- 支持任务历史记录查询
- 支持任务性能分析和优化建议

## 8. 技术实现考虑

### 8.1 Redis数据结构
- 利用现有的Redis状态管理机制（[`internal/store/redis/status_manager.go`](internal/store/redis/status_manager.go)）
- 使用SCAN命令迭代获取所有以"request:id:"为前缀的键
- 解析每个键对应的JSON数据，构建任务状态对象

### 8.2 过滤逻辑
- 根据任务状态字段过滤出pending、processing、running状态的任务
- 可以考虑在Redis数据结构中添加状态索引以提高过滤效率

### 8.3 数据结构复用
- 集成现有的任务状态数据结构（[`internal/types/status.go`](internal/types/status.go)中的FileStatusResponseData）
- 确保与现有状态更新逻辑的兼容性

### 8.4 API实现
- 创建新的HTTP处理器（参考[`internal/handler/status.go`](internal/handler/status.go)）
- 实现对应的业务逻辑（参考[`internal/logic/status.go`](internal/logic/status.go)）
- 在路由配置中添加新的GET路由（[`internal/handler/routes.go`](internal/handler/routes.go)）