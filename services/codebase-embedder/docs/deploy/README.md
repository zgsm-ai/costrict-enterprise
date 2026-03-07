# Codebase-Embedder 部署文档

本文档提供了 Codebase-Embedder 项目的完整部署指南，包括开发环境、生产环境和 Kubernetes 环境的部署步骤。

## 目录

- [系统要求](#系统要求)
- [本地开发环境部署](#本地开发环境部署)
- [Docker Compose 部署](#docker-compose-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [配置说明](#配置说明)
- [运维指南](#运维指南)

## 系统要求

### 硬件要求
- CPU: 4 核心及以上
- 内存: 8GB 及以上
- 存储: 50GB 可用空间（用于向量数据库和代码存储）

### 软件要求
- Go 1.24.3 或更高版本（本地开发）
- Docker 20.10 或更高版本
- Docker Compose 2.0 或更高版本
- Kubernetes 1.20 或更高版本（K8s 部署）
- kubectl 命令行工具

### 依赖服务
- PostgreSQL 15+
- Redis 7.0+
- Weaviate 1.30+

## 本地开发环境部署

### 1. 环境准备

```bash
# 克隆项目
git clone https://github.com/zgsm-ai/codebase-embedder.git
cd codebase-embedder

# 安装 Go 依赖
go mod tidy

# 安装开发工具
make init
```

### 2. 启动依赖服务

```bash
# 启动 PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_DB=codebase_indexer \
  -e POSTGRES_USER=shenma \
  -e POSTGRES_PASSWORD=shenma \
  -p 5432:5432 \
  postgres:15-alpine

# 启动 Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:7.2.4

# 启动 Weaviate
docker run -d --name weaviate \
  -p 8080:8080 \
  -p 50051:50051 \
  semitechnologies/weaviate:1.31.0 \
  --host 0.0.0.0 --port 8080 --scheme http
```

### 3. 配置文件

编辑 [`etc/conf.yaml`](../etc/conf.yaml) 文件，确保以下配置正确：

```yaml
Database:
  DataSource: postgres://shenma:shenma@localhost:5432/codebase_indexer?sslmode=disable

Redis:
  Addr: localhost:6379

VectorStore:
  Weaviate:
    Endpoint: "localhost:8080"
```

### 4. 构建和运行

```bash
# 构建项目
make build

# 运行服务
./bin/main -f etc/conf.yaml
```

## Docker Compose 部署

### 1. 快速启动

```bash
# 进入项目根目录
cd /path/to/codebase-embedder

# 使用 Docker Compose 启动所有服务
docker-compose -f deploy/docker-compose.yml up -d
```

### 2. 验证部署

```bash
# 检查服务状态
docker-compose -f deploy/docker-compose.yml ps

# 查看日志
docker-compose -f deploy/docker-compose.yml logs -f codebase-embedder
```

### 3. 数据持久化

Docker Compose 配置已包含数据持久化：
- PostgreSQL 数据存储在 `${HOME}/postgres/data`
- Redis 数据存储在 `${HOME}/redis/data`
- Weaviate 数据存储在 `${HOME}/weaviate/data`

## Kubernetes 部署

### 1. 准备工作

```bash
# 创建命名空间
kubectl create namespace costrict

# 配置 kubectl 上下文
kubectl config set-context --current --namespace=costrict
```

### 2. 部署 Weaviate

```bash
# 部署 Weaviate 服务
kubectl apply -f deploy/weaviate-k8s.yaml

# 等待 Weaviate 就绪
kubectl wait --for=condition=ready pod -l app=weaviate --timeout=300s
```

### 3. 部署 Codebase-Embedder

```bash
# 部署应用
kubectl apply -f deploy/embber.yaml

# 等待应用就绪
kubectl wait --for=condition=ready pod -l app=codebase-embedder --timeout=300s
```

### 4. 验证部署

```bash
# 检查所有资源状态
kubectl get all -n costrict

# 检查 Pod 状态
kubectl get pods -n costrict

# 查看服务
kubectl get svc -n costrict
```

### 5. 访问服务

- Codebase-Embedder API: `http://<node-ip>:32002`
- Weaviate 服务: `http://<node-ip>:32003`
- Weaviate gRPC: `http://<node-ip>:32004`

## 配置说明

### 核心配置项

#### 数据库配置
```yaml
Database:
  Driver: postgres
  DataSource: postgres://username:password@host:port/database?sslmode=disable
  AutoMigrate:
    enable: true
```

#### Redis 配置
```yaml
Redis:
  Addr: host:port
  DefaultExpiration: 1h
```

#### 向量存储配置
```yaml
VectorStore:
  Type: weaviate
  Timeout: 60s
  MaxRetries: 5
  Weaviate:
    Endpoint: "weaviate-host:8080"
    BatchSize: 100
    ClassName: "CodebaseIndex"
```

#### 嵌入模型配置
```yaml
VectorStore:
  Embedder:
    Timeout: 30s
    MaxRetries: 3
    BatchSize: 1
    Model: gte-modernbert-base
    ApiKey: "your-api-key"
    ApiBase: https://your-embedding-api-endpoint/v1/embeddings
```

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `TZ` | `Asia/Shanghai` | 时区设置 |
| `INDEX_NODE` | `1` | 索引节点标识 |
| `MODE` | `dev` | 运行模式 (dev/test/rt/pre/pro) |

## 运维指南

### 1. 健康检查

```bash
# 本地健康检查
curl http://localhost:8888/health

# Kubernetes 健康检查
kubectl get pods -n costrict -l app=codebase-embedder
```

### 2. 日志管理

```bash
# Docker Compose 日志
docker-compose -f deploy/docker-compose.yml logs -f codebase-embedder

# Kubernetes 日志
kubectl logs -f deployment/codebase-embedder -n costrict
```

### 3. 备份和恢复

#### PostgreSQL 备份
```bash
# 备份
docker exec postgres pg_dump -U shenma codebase_indexer > backup.sql

# 恢复
docker exec -i postgres psql -U shenma codebase_indexer < backup.sql
```

#### Weaviate 备份
```bash
# 创建备份
curl -X POST http://localhost:8080/v1/backups -H "Content-Type: application/json" -d '{"id": "my-backup"}'

# 恢复备份
curl -X POST http://localhost:8080/v1/backups/my-backup/restore -H "Content-Type: application/json"
```

### 4. 扩缩容

#### Docker Compose 扩缩容
```bash
# 扩容到 3 个实例
docker-compose -f deploy/docker-compose.yml up -d --scale codebase-embedder=3
```

#### Kubernetes 扩缩容
```bash
# 扩容到 5 个副本
kubectl scale deployment codebase-embedder --replicas=5 -n costrict
```

### 5. 监控指标

应用暴露了以下监控指标：
- HTTP 请求计数和延迟
- 数据库连接池状态
- Redis 连接状态
- 向量嵌入处理时间
- 任务队列状态

### 6. 故障排查

#### 常见问题

1. **服务无法启动**
   - 检查配置文件语法
   - 确认依赖服务（PostgreSQL、Redis、Weaviate）正常运行
   - 查看详细日志

2. **向量嵌入失败**
   - 检查 API 密钥是否有效
   - 确认嵌入服务端点可访问
   - 检查网络连接

3. **数据库连接失败**
   - 验证数据库连接字符串
   - 检查数据库服务状态
   - 确认用户权限

#### 调试模式

```yaml
# 在配置文件中启用调试模式
Log:
  Level: debug
  Mode: console
```

## 版本更新

### Docker Compose 更新

```bash
# 拉取最新镜像
docker-compose -f deploy/docker-compose.yml pull

# 重新构建和启动
docker-compose -f deploy/docker-compose.yml up -d --build
```

### Kubernetes 更新

```bash
# 更新镜像
kubectl set image deployment/codebase-embedder codebase-embedder=zgsm/codebase-embedder:v0.0.22 -n costrict

# 滚动重启
kubectl rollout restart deployment/codebase-embedder -n costrict

# 检查更新状态
kubectl rollout status deployment/codebase-embedder -n costrict
```

## 安全建议

1. **生产环境配置**
   - 修改默认密码
   - 使用 HTTPS 协议
   - 启用认证机制
   - 限制网络访问

2. **密钥管理**
   - 使用 Kubernetes Secrets 管理敏感信息
   - 定期轮换 API 密钥
   - 不要在配置文件中硬编码密钥

3. **网络安全**
   - 使用防火墙规则限制访问
   - 启用网络策略（Network Policy）
   - 定期更新依赖组件

## 联系支持

如需技术支持，请联系：
- 邮箱: support@zgsm-ai.com
- 文档: https://docs.zgsm-ai.com
- 问题反馈: https://github.com/zgsm-ai/codebase-embedder/issues