# Chat-RAG 🚀

<div align="center">

[![Go版本](https://img.shields.io/badge/Go-1.24.2-blue.svg)](https://golang.org/doc/go1.24) [![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE) [![Docker](https://img.shields.io/badge/docker-available-blue.svg)](Dockerfile) [![构建状态](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

[English](./README.md) | [中文](#chinese)

</div>

## 🎯 项目概述

Chat-RAG 是一个高性能、企业级的聊天服务，结合了大语言模型（LLM）与检索增强生成（RAG）功能。它为现代 AI 应用提供智能上下文处理、工具集成和流式响应功能。

### 核心特性

- **🧠 智能上下文处理**：先进的提示工程，支持上下文压缩和过滤
- **🔧 工具集成**：无缝集成语义搜索、代码定义查询和知识库查询
- **⚡ 流式支持**：通过服务器发送事件（SSE）实现实时流式响应
- **🛡️ 企业安全**：基于 JWT 的身份验证和请求验证
- **📊 全面监控**：内置指标和日志记录，支持 Prometheus
- **🔄 多模态支持**：支持各种 LLM 模型和函数调用
- **🚀 高性能**：优化的低延迟响应和高吞吐量
 - **🤖 语义路由（来自 ai-llm-router 迁移）**：可选开启，自动按语义选择下游模型；在响应头透出 `x-select-llm`、`x-user-input`

## 🏗️ 架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API 网关      │───▶│  聊天处理器     │───▶│  提示引擎       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   身份验证      │    │  LLM 客户端     │    │  工具执行器     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   指标监控      │    │  Redis 缓存     │    │  搜索工具       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.24.2 或更高版本
- Redis 6.0+（可选，用于缓存）
- Docker（可选，用于容器化部署）

### 安装步骤

```bash
# 克隆仓库
git clone https://github.com/zgsm-ai/chat-rag.git
cd chat-rag

# 安装依赖
make deps

# 构建应用
make build

# 使用默认配置运行
make run
```

### Docker 部署

```bash
# 构建 Docker 镜像
make docker-build

# 运行容器
make docker-run
```

## ⚙️ 配置说明

服务通过 YAML 文件进行配置。查看 [`etc/chat-api.yaml`](etc/chat-api.yaml) 了解默认配置：

```yaml
# 服务
Host: 0.0.0.0
Port: 8080

# LLM 上游（单一端点；具体模型由请求体的 model 字段决定）
LLM:
  Endpoint: "http://localhost:8000/v1/chat/completions"
  # 可选：支持函数调用的模型清单
  FuncCallingModels: ["gpt-4o-mini", "o4-mini"]

# 系统提示词处理（默认 false，保持请求中的 system 消息不变）
promptProcessing:
  modifySystemPrompt: false

# LLM 超时和重试配置（普通模式）
LLMTimeout:
  idleTimeoutMs: 180000          # 单次空闲超时（毫秒），默认 180000ms (180s)
  totalIdleTimeoutMs: 180000     # 总空闲超时预算（毫秒），默认 180000ms (180s)
  maxRetryCount: 1               # 最大重试次数，默认 1（即总共尝试 2 次）
  retryIntervalMs: 5000          # 重试间隔（毫秒），默认 5000ms（5秒）

# 上下文压缩
ContextCompressConfig:
  EnableCompress: true
  TokenThreshold: 5000
  SummaryModel: "deepseek-v3"
  SummaryModelTokenThreshold: 4000
  RecentUserMsgUsedNums: 4

# 工具（RAG 后端）
Tools:
  SemanticSearch:
    SearchEndpoint: "http://localhost:8002/codebase-indexer/api/v1/semantics"
    ApiReadyEndpoint: "http://localhost:8002/healthz"
    TopK: 5
    ScoreThreshold: 0.3
  DefinitionSearch:
    SearchEndpoint: "http://localhost:8002/codebase-indexer/api/v1/definitions"
    ApiReadyEndpoint: "http://localhost:8002/healthz"
  ReferenceSearch:
    SearchEndpoint: "http://localhost:8002/codebase-indexer/api/v1/references"
    ApiReadyEndpoint: "http://localhost:8002/healthz"
  KnowledgeSearch:
    SearchEndpoint: "http://localhost:8003/knowledge/api/v1/search"
    ApiReadyEndpoint: "http://localhost:8003/healthz"
    TopK: 5
    ScoreThreshold: 0.3

# 日志与分类
Log:
  LogFilePath: "logs/chat-rag.log"
  LokiEndpoint: "http://localhost:3100/loki/api/v1/push"
  LogScanIntervalSec: 60
  ClassifyModel: "deepseek-v3"
  EnableClassification: true
  # 错误日志持久化策略：all | sampled | none。
  # sampled 模式在每个时间窗内，每用户每错误类型最多保留 errorLogSampleN 条。
  errorLogMode: sampled
  errorLogSampleN: 1
  errorLogSampleWindowSec: 60

# Redis（可选）
Redis:
  Addr: "127.0.0.1:6379"
  Password: ""
  DB: 0

# 模型选择路由（支持语义路由和优先级路由策略）
router:
  enabled: true
  strategy: semantic  # 可选: semantic（语义路由）, priority（优先级轮询）
  semantic:
    analyzer:
      model: gpt-4o-mini
      timeoutMs: 3000
      # 可为 analyzer 单独覆盖全局 LLM 的端点与令牌
      # endpoint: "http://higress-gateway.costrict.svc.cluster.local/v1/chat/completions"
      # apiToken: "<你的令牌>"
      # 可选高级项：
      # totalTimeoutMs: 5000
      # maxInputBytes: 8192
      # promptTemplate: ""   # 自定义分类 Prompt，不配置则使用内置默认
      # analysisLabels: ["simple_request", "planning_request", "code_modification"]
      # dynamicMetrics:
      #   enabled: false
      #   redisPrefix: "ai_router:metrics:"
      #   metrics: ["error_rate", "p99", "circuit"]
    inputExtraction:
      protocol: openai
      userJoinSep: "\n\n"
      stripCodeFences: true
      codeFenceRegex: ""
      maxUserMessages: 100
      maxHistoryBytes: 4096
    routing:
      candidates:
        - modelName: "gpt-4o-mini"
          enabled: true
          scores:
            simple_request: 10
            planning_request: 5
            code_modification: 3
        - modelName: "o4-mini"
          enabled: true
          scores:
            simple_request: 4
            planning_request: 8
            code_modification: 6
      minScore: 0
      tieBreakOrder: ["o4-mini", "gpt-4o-mini"]
      fallbackModelName: "gpt-4o-mini"

      # 模型降级场景的超时配置（独立于普通模式）
      idleTimeoutMs: 180000        # 单次空闲超时，默认 180000ms (180s)
      totalIdleTimeoutMs: 180000   # 总空闲超时预算，默认 180000ms (180s)

      # 模型降级场景的重试配置（独立于普通模式）
      maxRetryCount: 1             # 最大重试次数，默认 1（总共尝试 2 次）
      retryIntervalMs: 5000        # 重试间隔（毫秒），默认 5000ms
    ruleEngine:
      enabled: false
      inlineRules: []
      bodyPrefix: "body."
      headerPrefix: "header."

  # 优先级轮询策略（semantic 的替代方案）
  # 取消注释以使用优先级策略代替语义路由
  priority:
    candidates:
      - modelName: "gpt-4"
        enabled: true
        priority: 1           # 优先级（数字越小优先级越高，范围 0-999）
        weight: 5             # 权重（同优先级内的负载均衡，范围 1-100）
        minVipLevel: 1        # AUTO 模式可见所需的最低 VIP 等级；0 表示所有用户可见

      - modelName: "claude-3-opus"
        enabled: true
        priority: 1           # 与 gpt-4 同优先级
        weight: 3             # 权重比 gpt-4 低
        minVipLevel: 0

      - modelName: "gpt-3.5-turbo"
        enabled: true
        priority: 2           # 优先级较低，仅在优先级 1 失败时使用
        weight: 10
        minVipLevel: 0

    fallbackModelName: "gpt-3.5-turbo"

    # 超时配置（与语义路由相同）
    idleTimeoutMs: 180000
    totalIdleTimeoutMs: 180000

    # 重试配置（与语义路由相同）
    maxRetryCount: 1
    retryIntervalMs: 5000
```

#### 配置字段详解（节选）

- **LLM**
  - `Endpoint`：统一的 Chat Completions 端点；最终模型名通过请求体 `model` 传递
  - `FuncCallingModels`：具备函数调用能力的模型清单，便于按需启用工具
- **promptProcessing**
  - `modifySystemPrompt`：是否允许规则、工具、过滤器和语言设置改写 system 消息；默认为 `false`
- **LLMTimeout**（普通模式 - 不使用路由或 model != "auto" 时）
  - `idleTimeoutMs`：单次空闲超时（毫秒），默认 180000ms (180s)
  - `totalIdleTimeoutMs`：总空闲超时预算（毫秒），默认 180000ms (180s)
  - `maxRetryCount`：可重试错误的最大重试次数（超时、网络错误），默认 1（总共尝试 2 次）
  - `retryIntervalMs`：重试间隔（毫秒），默认 5000ms（5秒）
- **ContextCompressConfig**
  - `EnableCompress`：是否开启长上下文压缩
  - `TokenThreshold`：超过此阈值触发压缩
  - `SummaryModel` / `SummaryModelTokenThreshold`：用于摘要压缩的模型与阈值
  - `RecentUserMsgUsedNums`：压缩流程中参照的最近用户消息数量
- **Tools**（RAG）
  - 各搜索模块提供 HTTP 端点；`TopK`/`ScoreThreshold` 控制召回数量与质量
- **Log**
  - `LogFilePath`：本地日志文件路径；后台进程会批量上传至 Loki
  - `LokiEndpoint`：Loki Push 端点
  - `LogScanIntervalSec`：日志扫描与上传周期
  - `ClassifyModel` / `EnableClassification`：是否使用 LLM 对日志分类
  - `errorLogMode`：错误日志持久化策略：`all`、`sampled` 或 `none`。为空时回落至已弃用的 `saveErrorLog`
  - `errorLogSampleN` / `errorLogSampleWindowSec`：`sampled` 模式的采样参数（默认 N=1，窗口=60s）
- **Redis**：可选；用于工具状态、路由动态指标等
- **router**（模型选择路由）
  - `enabled` / `strategy`：启用路由；可选策略：`semantic`（语义路由）、`priority`（优先级轮询）
  - **semantic** 策略配置：
    - `analyzer`：分类模型/超时；支持仅对 analyzer 覆盖 endpoint/apiToken；在 auto 模式下使用独立的非流式客户端；可自定义 Prompt 与标签；可选动态指标（Redis）
    - `inputExtraction`：控制用户输入与历史的抽取方式，支持去除代码块、限制历史长度
    - `routing`：候选模型评分表；通过 `tieBreakOrder` 解决同分，`fallbackModelName` 兜底；支持模型降级场景的独立超时和重试配置：
      - `idleTimeoutMs`：降级重试的单次空闲超时（毫秒），默认 180000ms (180s)
      - `totalIdleTimeoutMs`：降级重试的总空闲超时预算（毫秒），默认 180000ms (180s)
      - `maxRetryCount`：降级重试的最大重试次数，默认 1
      - `retryIntervalMs`：降级重试的重试间隔（毫秒），默认 5000ms
    - `ruleEngine`：可选的规则引擎预筛模型，默认关闭
  - **priority** 策略配置（semantic 的替代方案）：
    - 简单、低成本的策略，无需语义分析；根据优先级选择模型（数字越小优先级越高，范围 0-999）
    - 使用平滑加权轮询算法在同优先级组内实现负载均衡
    - 配置字段：
      - `candidates`：候选模型列表，包含 `modelName`、`enabled`、`priority`（0-999）、`weight`（1-100）以及可选的 `minVipLevel`（默认 0）
      - `minVipLevel`：AUTO 模式下可见该候选模型所需的最低 VIP 等级。`0` 或未配置表示所有用户可见。不满足要求的 VIP 模型不会进入 AUTO 首选或降级链路。
      - `fallbackModelName`：所有候选模型失败时的回退模型。priority AUTO 路由中，fallback 必须存在于 `candidates`（配置加载时强校验，缺失则启动失败）；运行时仅当 fallback 对当前用户可见时才追加到降级链，否则记录 warn 日志并跳过。
      - **降级顺序契约**：同 `priority` 内按 `weight` 降序排序；`weight` 相同时保持 `candidates` 配置的原始顺序（稳定排序）。
      - 超时和重试配置（与语义路由相同）：
        - `idleTimeoutMs`：单次空闲超时（毫秒），默认 180000ms (180s)
        - `totalIdleTimeoutMs`：总空闲超时预算（毫秒），默认 180000ms (180s)
        - `maxRetryCount`：最大重试次数，默认 1
        - `retryIntervalMs`：重试间隔（毫秒），默认 5000ms
    - **性能优化**：单模型优先级组使用快速路径，零锁开销

## 📡 API 端点

### 聊天完成（非流式）

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "今天天气怎么样？"}
    ],
    "stream": false
  }'
```

### 启用语义路由（自动选型）

将请求体中的 `model` 置为 `auto`，并在配置中开启 `router.enabled: true`：

```bash
curl -i -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "auto",
    "messages": [
      {"role": "user", "content": "给我一个详细的改造方案并产出代码示例"}
    ],
    "stream": false
  }'
```

响应头将包含：
- `x-select-llm`：最终选择的下游模型名
- `x-user-input`：用于分类的用户输入（已清洗并进行 base64 编码）

### 流式响应

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "写一个 Python 函数"}
    ],
    "stream": true
  }'
```

### 指标监控

Prometheus 指标暴露在 `/metrics`，详见 `METRICS.md`。

## 🔧 开发指南

### 项目结构

```
chat-rag/
├── internal/
│   ├── handler/          # HTTP 处理器
│   ├── logic/           # 业务逻辑
│   ├── router/          # 语义路由（策略 + 工厂）
│   ├── client/          # 外部服务客户端
│   ├── promptflow/      # 提示处理管道
│   ├── functions/       # 工具执行引擎
│   └── config/          # 配置管理
├── etc/                 # 配置文件
├── test/               # 测试文件
└── deploy/             # 部署配置
```

### 可用命令

```bash
make help              # 显示可用命令
make build            # 构建应用
make test             # 运行测试
make fmt              # 格式化代码
make vet              # 检查代码
make docker-build     # 构建 Docker 镜像
make dev              # 运行开发服务器（支持热重载）
```

### 测试

```bash
# 运行所有测试
make test

# 运行特定测试
go test -v ./internal/logic/

# 带覆盖率运行
go test -cover ./...
```

## 🔍 高级功能

### 上下文压缩

智能上下文压缩处理长对话：

```yaml
ContextCompressConfig:
  EnableCompress: true
  TokenThreshold: 5000
  SummaryModel: "deepseek-v3"
  SummaryModelTokenThreshold: 4000
  RecentUserMsgUsedNums: 4
```

### 工具集成

支持多种搜索和分析工具：

- **语义搜索**：基于向量的代码和文档搜索
- **定义搜索**：代码定义查询
- **引用搜索**：代码引用分析
- **知识搜索**：文档知识库查询

### 语义路由（来自 ai-llm-router 迁移）

当 `router.enabled: true` 且请求体 `model` 为 `auto` 时，将自动选择最合适的下游模型：

1. 输入抽取：按 `router.semantic.inputExtraction` 提取当前输入与少量历史，可选移除代码块
2. 语义分类：调用 `router.semantic.analyzer.model` 获取标签（默认：simple_request / planning_request / code_modification）
3. 候选打分：在 `routing.candidates` 中按标签取分；支持 `minScore` 和动态指标（可选）
4. Tie-break 与回退：用 `tieBreakOrder` 破同分；失败或低于阈值则使用 `fallbackModelName`
5. 可观测性：在响应头写入 `x-select-llm` 与 `x-user-input`（后者做过清洗并 base64 编码）

### 基于代理的处理

可配置的代理匹配，用于专门任务：

```yaml
AgentsMatch:
  - AgentName: "strict"
    MatchKey: "a strict strategic workflow controller"
  - AgentName: "code"
    MatchKey: "a highly skilled software engineer"
```

## 📊 监控和可观测性

### 指标监控

服务在 `/metrics` 端点暴露 Prometheus 指标：

- 请求计数和延迟
- Token 使用统计
- 工具执行指标
- 错误率和类型

### 日志记录

使用 Zap 记录器进行结构化日志记录：

- 请求/响应日志记录
- 错误跟踪
- 性能指标
- 调试信息

## 🔒 安全特性

- 基于 JWT 的身份验证
- 请求验证和清理
- 速率限制支持
- 安全头部处理

## 🚢 部署方案

### 生产部署

```bash
# 构建生产版本
CGO_ENABLED=0 GOOS=linux go build -o chat-rag .

# 使用生产配置运行
./chat-rag -f etc/prod.yaml
```

### Kubernetes 部署

查看 [`deploy/`](deploy/) 目录中的 Kubernetes 清单和 Helm 图表。

## 🤝 贡献指南

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add some amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 打开拉取请求

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

如需支持和提问：
- 在 GitHub 仓库中创建问题
- 联系维护者

---

<div align="center">
  <b>⭐ 如果这个项目对你有帮助，请给我们一个星标！</b>
</div>

## S3 日志归档

`log.storageType: s3` 使用 AWS SDK Go v2。配置示例（保留部署中的原 bucket、endpoint 和凭证）：

```yaml
log:
  storageType: s3
  s3:
    endpoint: "s3.example.com:9000" # host[:port]，不带协议和路径
    bucket: "existing-bucket"
    accessKey: "<access-key>"
    secretKey: "<secret-key>"
    region: "us-east-1"
    useSSL: true
    skipSSLVerify: false
    skipBucketCheck: true
    forcePathStyle: true
    disableRequestChecksum: true
```

新增 `forcePathStyle` 和 `disableRequestChecksum` 均默认 true，可显式设为 false。
前者使用 `/bucket/key` 请求地址，不改变桶内对象 key；后者仅在接口要求时计算请求校验和，避免兼容服务错误处理 CRC trailer，不关闭请求签名或 TLS 验证。
`skipSSLVerify` 默认 false，仅 HTTPS 下生效；`skipBucketCheck` 默认 false，默认检查桶，仅不存在时尝试创建。桶已预建且凭证权限有限时可跳过检查。
`region` 为空时使用 `us-east-1`，非该区域的服务需填写实际区域。
对象目录、文件名、JSON 正文及末尾换行、Content-Type、返回的对象路径和无引号 ETag 保持不变，不添加 prefix。
