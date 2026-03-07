# 全路径转发功能使用指南

## 概述

全路径转发功能允许代理服务将原始请求路径完整转发到目标服务，不进行任何路径重写或前缀移除。这与传统的路径重写模式形成对比。

## 功能对比

| 特性 | 路径重写模式 (rewrite) | 全路径转发模式 (full_path) |
|------|----------------------|------------------------|
| 路径处理 | 移除代理前缀，应用重写规则 | 保持原始路径不变 |
| 路由匹配 | `/api/v1/proxy/*path` | `/*path` |
| 适用场景 | API网关、路径转换 | 透明代理、反向代理 |
| 配置复杂度 | 需要配置重写规则 | 无需额外配置 |

## 配置方法

### 1. 配置文件设置

在 `bin/etc/conf.yaml` 中配置代理模式：

```yaml
# 全路径转发模式配置
proxy_config:
  mode: "full_path"  # 设置为全路径转发模式
  target:
    url: "http://localhost:11380"  # 目标服务地址
    timeout: 30s
  rewrite:
    enabled: false  # 全路径模式下禁用路径重写
    rules: []
  headers:
    pass_through: true
    exclude:
      - "X-Internal-*"
      - "Authorization"
    override:
      Host: "localhost:11380"
```

### 2. 环境变量设置

也可以通过环境变量设置：

```bash
export PROXY_MODE=full_path
export PROXY_TARGET_URL=http://localhost:11380
```

## 使用示例

### 场景1：透明代理到后端服务

假设：
- 代理服务运行在 `http://localhost:1010`
- 目标服务运行在 `http://localhost:11380`

**请求示例：**
```bash
# 原始请求
curl http://localhost:1010/codebase-indexer/api/v1/register

# 转发到目标服务
# 目标服务接收到的请求路径：/codebase-indexer/api/v1/register
```

### 场景2：微服务代理

**配置：**
```yaml
proxy_config:
  mode: "full_path"
  target:
    url: "http://user-service:8080"
```

**请求示例：**
```bash
# 请求用户服务
curl http://localhost:1010/users/profile/123

# 转发到用户服务
# 用户服务接收到的请求路径：/users/profile/123
```

## 调试指南

### 启用调试日志

在配置文件中设置日志级别为 `debug`：

```yaml
Log:
  Level: debug
```

### 日志输出示例

```
[PROXY_LOGIC_DEBUG] === Building Target Request ===
[PROXY_LOGIC_DEBUG] Original request path: /api/v1/users/123
[PROXY_LOGIC_DEBUG] Proxy mode: full_path
[PROXY_LOGIC_DEBUG] Final target URL: http://localhost:11380/api/v1/users/123
```

## 常见问题

### Q1: 如何切换回路径重写模式？

**A:** 修改配置文件中的 `mode` 为 `rewrite`：

```yaml
proxy_config:
  mode: "rewrite"
```

### Q2: 全路径模式下是否支持路径重写？

**A:** 不支持。全路径模式会保持原始路径不变，忽略所有重写规则。

### Q3: 如何处理查询参数？

**A:** 查询参数会自动透传，无需额外配置。

**示例：**
```bash
curl "http://localhost:1010/search?q=golang&limit=10"
# 转发到：http://localhost:11380/search?q=golang&limit=10
```

### Q4: 如何验证配置是否生效？

**A:** 访问健康检查端点：

```bash
curl http://localhost:1010/codebase-indexer/api/v1/proxy/health
```

返回示例：
```json
{
  "status": "ok",
  "proxy": {
    "mode": "full_path",
    "target_url": "http://localhost:11380",
    "reachable": true,
    "response_time_ms": 5
  }
}
```

## 性能考虑

- 全路径转发模式比路径重写模式性能略高（减少了路径处理开销）
- 建议在高并发场景下使用全路径模式以获得更好的性能

## 安全建议

1. **路径验证**：确保目标服务能够处理所有可能的路径
2. **访问控制**：结合认证中间件限制访问
3. **速率限制**：配置适当的速率限制防止滥用

## 迁移指南

从路径重写模式迁移到全路径模式：

1. 备份当前配置
2. 修改 `mode` 为 `full_path`
3. 禁用 `rewrite.enabled`
4. 清空 `rewrite.rules`
5. 重启服务
6. 验证功能正常

## 故障排除

### 问题：404 Not Found

**可能原因：**
- 目标服务路径不匹配
- 路由配置错误

**解决方案：**
- 检查目标服务是否监听正确路径
- 验证配置文件中的 `target.url`

### 问题：连接超时

**可能原因：**
- 目标服务不可达
- 网络配置问题

**解决方案：**
- 检查目标服务状态
- 验证网络连通性
- 调整 `timeout` 配置