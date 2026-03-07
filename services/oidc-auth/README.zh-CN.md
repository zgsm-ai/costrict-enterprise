# OIDC Authentication Server

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](Dockerfile)

一个基于 Casdoor 的现代化 OIDC 认证服务器，提供企业级用户认证和授权功能。

## ✨ 功能特性

- 🔐 **OIDC 标准认证** - 基于 OpenID Connect 协议的安全认证
- 🌟 **GitHub 集成** - 支持 GitHub Star 同步和用户关联
- 📱 **短信验证** - 集成短信服务，支持验证码发送
- 🗄️ **多数据库支持** - 兼容 MySQL 8.0+ 和 PostgreSQL 13+
- 🐳 **容器化部署** - 完整的 Docker 和 Kubernetes 支持
- ⚡ **高性能** - 优化的连接池和并发处理
- 🛡️ **安全中间件** - 完整的安全头和请求日志

## 🚀 快速开始

### 环境要求
- Go 1.23.0+
- MySQL 8.0+ 或 PostgreSQL 13+
- Docker (可选)

### 本地开发

1. **克隆项目**
```bash
git clone https://github.com/zgsm-ai/oidc-auth.git
cd oidc-auth
```

2. **安装依赖**
```bash
go mod tidy
```

3. **配置文件**
```bash
cp config/config.yaml config/config.yaml.local
# 根据实际情况编辑配置文件
```

4. **运行服务**
```bash
go run cmd/main.go serve --config config/config.yaml
```

服务将在 `http://localhost:8080` 启动。

### Docker 部署

1. **构建镜像**
```bash
docker build -t oidc-auth:latest .
```

2. **运行容器**
```bash
docker run -d \
  --name oidc-auth \
  -p 8080:8080 \
  -e SERVER_BASEURL="<server_base_url>" \
  -e PROVIDERS_CASDOOR_CLIENTID="<casdoor_client_id>" \
  -e PROVIDERS_CASDOOR_CLIENTSECRET="<casdoor_client_secret>" \
  -e PROVIDERS_CASDOOR_BASEURL="<casdoor_base_url>" \
  -e PROVIDERS_CASDOOR_INTERNALURL="<casdoor_base_url>" \
  -e SMS_ENABLEDTEST="false" \
  -e SYNCSTAR_ENABLED="false" \
  -e DATABASE_HOST="<database_host>" \
  -e DATABASE_PASSWORD="<database_password>" \
  -e ENCRYPT_AESKEY="<aes_key>" \
  oidc-auth:latest
```

## ⚙️ 配置说明

### 环境变量配置

支持完整的环境变量配置，便于容器化部署：

| 配置分类              | 环境变量 | 描述                    | 默认值 |
|-------------------|---------|-----------------------|--------|
| **服务器配置**         | `SERVER_SERVERPORT` | 服务端口                  | `8080` |
|                   | `SERVER_BASEURL` | 服务基础URL               | `http://localhost:8080` |
|                   | `SERVER_ISPRIVATE` | 内网模式                  | `false` |
| **认证提供商**         | `PROVIDERS_CASDOOR_CLIENTID` | Casdoor 客户端ID         | - |
|                   | `PROVIDERS_CASDOOR_CLIENTSECRET` | Casdoor 客户端密钥         | - |
|                   | `PROVIDERS_CASDOOR_BASEURL` | Casdoor 服务地址          | - |
|                   | `PROVIDERS_CASDOOR_INTERNALURL` | Casdoor 服务内部地址        |-|
| **数据库配置**         | `DATABASE_TYPE` | 数据库类型                 | `postgres` |
|                   | `DATABASE_HOST` | 数据库主机                 | `localhost` |
|                   | `DATABASE_PORT` | 数据库端口                 | `5432` |
|                   | `DATABASE_USERNAME` | 数据库用户名                | `postgres` |
|                   | `DATABASE_PASSWORD` | 数据库密码                 | - |
|                   | `DATABASE_DBNAME` | 数据库名                  | `auth` |
|                   | `DATABASE_MAXIDLECONNS` | 最大空闲连接                | `50` |
|                   | `DATABASE_MAXOPENCONNS` | 最大连接数                 | `300` |
| **短信服务**          | `SMS_ENABLEDTEST` | 测试模式                  | `true` |
|                   | `SMS_CLIENTID` | 短信客户端ID               | - |
|                   | `SMS_CLIENTSECRET` | 短信客户端密钥               | - |
|                   | `SMS_TOKENURL` | Token 获取地址            | - |
|                   | `SMS_SENDURL` | 短信发送地址                | - |
| **GitHub star同步** | `SYNCSTAR_ENABLED` | 启用 Star 同步            | `true` |
|                   | `SYNCSTAR_PERSONALTOKEN` | GitHub Personal Token | - |
|                   | `SYNCSTAR_OWNER` | 仓库所有者                 | `zgsm-ai` |
|                   | `SYNCSTAR_REPO` | 仓库名称                  | `zgsm` |
|                   | `SYNCSTAR_INTERVAL` | 同步间隔(分钟)              | `1` |
| **加密配置**          | `ENCRYPT_AESKEY` | AES 密钥(32位)           | - |
|                   | `ENCRYPT_ENABLERSA` | 启用 RSA                | `false` |
|                   | `ENCRYPT_PRIVATEKEY` | RSA 私钥文件路径            | `config/private.pem` |
|                   | `ENCRYPT_PUBLICKEY` | RSA 公钥文件路径            | `config/public.pem` |
| **配额管理器**         | `QUOTAMANAGER_BASEURL` | 配额管理器服务基础URL        | - |
| **日志配置**          | `LOG_LEVEL` | 日志级别                  | `info` |
|                   | `LOG_FILENAME` | 日志文件路径                | `logs/app.log` |
|                   | `LOG_MAXSIZE` | 日志文件大小限制(MB)          | `100` |
|                   | `LOG_MAXBACKUPS` | 备份文件数量                | `10` |
|                   | `LOG_MAXAGE` | 日志保留天数                | `30` |
|                   | `LOG_COMPRESS` | 压缩旧日志                 | `true` |


## Kubernetes 部署

```bash
cp ./charts/oidc-auth/values.yaml /your/path/values.yaml
# modify /your/path/values.yaml
helm install -n oidc-auth oidc-auth ./charts/oidc-auth \
  --set replicaCount=1 \
  --set autoscaling.enabled=true \
  --set resources.requests.memory=512Mi \
  --create-namespace \
  -f /your/path/values.yaml
```

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 贡献

欢迎提交 Issue 和 Pull Request！

### 贡献指南
1. Fork 项目
2. 创建特性分支
3. 提交变更
4. 推送到分支
5. 创建 Pull Request

## 支持

如有问题或建议，请创建 [Issue](https://github.com/zgsm-ai/oidc-auth/issues)。
