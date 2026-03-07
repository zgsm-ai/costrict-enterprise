# HTTP代理转发功能

基于Go-zero框架实现的轻量级HTTP反向代理功能，支持请求转发、响应透传和错误处理。

## 功能特性

- ✅ **核心转发功能**：支持所有HTTP方法的请求转发
- ✅ **路径重写**：支持基于配置的路径重写规则
- ✅ **Header过滤**：支持Header的排除和覆盖配置
- ✅ **错误处理**：统一的错误响应格式
- ✅ **健康检查**：代理服务和目标服务的健康状态检查
- ✅ **性能优化**：连接池、超时控制、内存优化
- ✅ **配置管理**：YAML配置文件支持，环境变量覆盖

## 快速开始

### 1. 配置代理

编辑配置文件 `conf/proxy.yaml`：

```yaml
Proxy:
  Target:
    URL: "http://your-target-service:8080"
    Timeout: 30s
  Rewrite:
    Enabled: true
    Rules:
      - From: "/api/v1/proxy"
        To: ""
  Headers:
    PassThrough: true
    Exclude:
      - "X-Internal-*"
    Override:
      Host: "your-target-service"
```

### 2. 启动服务

```bash
# 使用go-zero启动
go run cmd/main.go

# 或者使用Docker
docker-compose up
```

### 3. 使用代理

所有以 `/proxy/*` 开头的请求都会被转发到配置的目标服务：

```bash
# 转发GET请求
curl http://localhost:8888/proxy/api/users

# 转发POST请求
curl -X POST http://localhost:8888/proxy/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}'
```

## API接口

### 代理转发接口

- **路径**: `/proxy/{path...}`
- **方法**: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- **描述**: 将请求转发到配置的目标地址
- **示例**: `GET /proxy/api/users` -> `GET http://target-service/api/users`

### 健康检查接口

- **路径**: `/health/proxy`
- **方法**: GET
- **响应**:
  ```json
  {
    "status": "ok",
    "proxy": {
      "target_url": "http://target-service:8080",
      "reachable": true,
      "response_time_ms": 45
    }
  }
  ```

## 配置说明

### 目标服务配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Target.URL` | string | - | 目标服务地址 |
| `Target.Timeout` | duration | 30s | 请求超时时间 |

### 路径重写配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Rewrite.Enabled` | bool | false | 是否启用路径重写 |
| `Rewrite.Rules` | array | [] | 重写规则列表 |

### Header配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Headers.PassThrough` | bool | true | 是否透传所有header |
| `Headers.Exclude` | array | [] | 需要排除的header列表 |
| `Headers.Override` | map | {} | 需要覆盖的header键值对 |

## 环境变量

支持通过环境变量覆盖配置：

| 环境变量 | 说明 | 示例 |
|----------|------|------|
| `PROXY_TARGET_URL` | 覆盖目标地址 | `http://new-target:8080` |
| `PROXY_TIMEOUT` | 覆盖超时时间 | `60s` |
| `PROXY_REWRITE_ENABLED` | 启用/禁用重写 | `true` |

## 错误处理

所有错误都返回统一的JSON格式：

```json
{
  "code": "PROXY_ERROR_CODE",
  "message": "错误描述",
  "details": "详细错误信息",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 错误码说明

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| `PROXY_BAD_REQUEST` | 400 | 请求格式错误 |
| `PROXY_TARGET_UNREACHABLE` | 503 | 目标服务不可达 |
| `PROXY_TIMEOUT` | 504 | 请求超时 |
| `PROXY_INTERNAL_ERROR` | 500 | 内部错误 |

## 性能指标

- **单请求延迟**: < 100ms（本地网络）
- **并发能力**: 支持100并发请求，成功率 > 99%
- **内存使用**: 稳定，无明显泄漏

## 开发指南

### 项目结构

```
internal/
├── config/          # 配置模块
├── handler/         # HTTP处理器
├── logic/           # 业务逻辑
└── utils/proxy/     # 工具函数
```

### 添加新功能

1. **配置扩展**: 在 `internal/config/proxy.go` 中添加新配置项
2. **业务逻辑**: 在 `internal/logic/proxy.go` 中实现新功能
3. **处理器**: 在 `internal/handler/proxy.go` 中添加新接口
4. **工具函数**: 在 `internal/utils/proxy/` 中添加通用工具

## 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试
go test -tags=integration ./...

# 性能测试
go test -bench=. ./...
```

## 监控

集成Prometheus监控指标：

- `proxy_requests_total`: 总请求数
- `proxy_request_duration_seconds`: 请求延迟
- `proxy_errors_total`: 错误总数
- `proxy_target_connection_status`: 目标连接状态

## 许可证

MIT License
