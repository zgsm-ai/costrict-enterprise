# 集成测试：不同嵌入模型和向量数据库对召回的影响测试

## 测试目标

本测试框架用于测试不同嵌入模型和向量数据库对代码检索召回效果的影响，为系统优化提供数据支持。

## 测试结构

```
test/
├── integration_test.go           # 主要的集成测试文件
├── config.yaml                   # 测试配置文件
├── testdata/                     # 手动准备的测试数据
│   ├── code_samples/             # 标准代码样例
│   │   ├── sample1.go           # Go用户服务实现
│   │   ├── sample2.py           # Python用户服务实现
│   │   └── sample3.js           # JavaScript用户服务实现
│   ├── doc_samples/             # 文档样例（统一目录）
│   │   ├── 自定义知识库功能.md   # 知识库功能需求文档
│   │   ├── index.md             # Roo Code文档首页
│   │   ├── api_documentation.md # API文档示例
│   │   └── swagger.json         # Swagger API定义
│   └── queries.yaml              # 测试查询定义
└── README.md                     # 本说明文件
```

## 测试数据

### 代码样例
- **sample1.go**: Go语言实现的用户服务，包含用户添加、获取、列表功能
- **sample2.py**: Python语言实现的用户服务，包含用户增删改查功能
- **sample3.js**: JavaScript语言实现的用户服务，包含用户管理和Map数据结构使用

### Markdown文档样例
- **自定义知识库功能.md**: 知识库管理和AI检索接口的需求文档，包含功能规格和验收标准
- **index.md**: Roo Code文档首页，介绍产品功能和使用方法
- **api_documentation.md**: API接口文档示例，包含用户认证和数据管理接口
- **swagger.json**: Swagger格式的API定义文件，描述RESTful API接口

### JSON配置样例
- **config.json**: 应用配置文件示例，包含数据库、嵌入模型、向量存储等配置

### 查询配置文件
测试框架支持两种格式的查询配置文件：
- **queries.json**: JSON格式的查询配置文件，支持场景类型区分
- **queries.yaml**: YAML格式的查询配置文件，兼容原有格式

### 场景类型区分
测试框架现在明确区分两种场景类型：
- **代码场景 (scenario_type: "code")**: 专注于代码相关的查询，如函数定义、类声明、方法调用等
- **文档场景 (scenario_type: "doc")**: 专注于文档内容的查询，如功能说明、API文档、配置说明等

这种区分使得测试结果分析更加精确，能够针对不同类型的内容优化检索策略。

### 测试查询
预定义了30个测试查询，涵盖：
- 代码功能查询（如"type UserService struct"）
- 特定语言查询（如"class UserService"、"NewUserService() *UserService"）
- 数据结构查询（如"dictionary operations in Python"）
- 文档内容查询（如"自定义知识库"、"AI统一检索接口"）
- 功能需求查询（如"知识库管理界面功能"）
- 验收标准查询（如"检索准确率验收指标"）
- 产品功能查询（如"What Can Roo Code Do"、"Smart Tools"）
- 配置设置查询（如"custom modes and tools"）

查询配置现在支持更丰富的元数据，包括语言类型、场景类型等，便于进行分类分析和性能评估。

## 运行测试

### 前置条件
1. 确保Go环境已正确配置
2. 安装项目依赖：`go mod tidy`
3. 设置环境变量（可选）：
   ```bash
   export EMBEDDER_API_KEY="your_embedder_api_key"
   export OPENAI_API_KEY="your_openai_api_key"
   export WEAVIATE_ENDPOINT="localhost:8080"
   ```

### 运行命令

```bash
# 运行所有测试
go test -v ./test/integration_test.go

# 运行特定测试
go test -v ./test/integration_test.go -run TestEmbedderComparison
go test -v ./test/integration_test.go -run TestVectorStoreComparison

# 使用自定义配置文件
go test -v ./test/integration_test.go -config=/path/to/config.yaml

# 指定输出目录
go test -v ./test/integration_test.go -output=/path/to/output

# 跳过长时间运行的测试
go test -v ./test/integration_test.go -short

# 指定查询配置文件
go test -v ./test/integration_test.go -queries=/path/to/queries.json

# 启用详细日志
go test -v ./test/integration_test.go -log-level=debug

# 并发运行测试
go test -v ./test/integration_test.go -parallel=4
```

### 新增配置选项

测试框架支持以下新的配置选项：

- **queries**: 指定查询配置文件路径，支持JSON和YAML格式
- **log-level**: 设置日志级别（debug、info、warn、error）
- **parallel**: 设置并发测试数量
- **timeout**: 设置单个测试的超时时间
- **retry**: 设置失败测试的重试次数
- **metrics-format**: 指定指标输出格式（json、csv、console）

## 配置说明

### config.yaml 配置文件

主要配置项：

- **嵌入模型配置**: 支持配置多个嵌入模型（如 gte-modernbert-base、text-embedding-ada-002）
- **向量数据库配置**: 支持配置多个向量数据库实例（如不同端口的Weaviate）
- **测试场景配置**: 定义嵌入模型对比和向量数据库对比测试
- **输出配置**: 控制结果输出格式和目录

### 环境变量

- `EMBEDDER_API_KEY`: 嵌入模型API密钥
- `OPENAI_API_KEY`: OpenAI API密钥
- `WEAVIATE_ENDPOINT`: Weaviate服务端点

## 测试结果

### 输出格式
测试结果会以JSON格式保存在 `test/results/` 目录中：
- `embedder_comparison.json`: 嵌入模型对比测试结果
- `vector_store_comparison.json`: 向量数据库对比测试结果

### 评估指标

#### 基础指标
- **准确率 (Precision)**: 检索到的相关文件比例
- **召回率 (Recall)**: 相关文件被检索到的比例
- **F1分数**: 准确率和召回率的调和平均
- **响应时间**: 平均查询响应时间（毫秒）

#### 高级指标
- **平均精度均值 (MAP - Mean Average Precision)**:
  - 衡量检索系统在多个查询上的平均性能
  - 考虑了检索结果的相关性排序
  - 对排序靠前的相关结果给予更高权重
  - 公式: MAP = (1/Q) * Σ(Average Precision for each query)

- **归一化折损累计增益 (NDCG - Normalized Discounted Cumulative Gain)**:
  - 评估检索结果排序质量的高级指标
  - 考虑了结果的位置权重（位置越靠前权重越高）
  - 支持分级相关性（而不仅仅是相关/不相关）
  - 公式: NDCG = DCG / IDCG，其中DCG为实际折损累计增益，IDCG为理想折损累计增益

#### 指标说明
- **MAP** 特别适用于评估多文档检索场景，能够综合反映检索的准确性和排序质量
- **NDCG** 对排序质量敏感，适合评估需要精确排序的应用场景
- **响应时间** 包括网络传输、向量计算和数据库查询的完整时间
- 所有指标都支持按场景类型（代码/文档）分别统计，便于针对性优化

### 结果示例
```json
{
  "test_name": "嵌入模型对比测试",
  "test_description": "使用相同向量数据库，测试不同嵌入模型的召回效果",
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": "2024-01-01T10:05:00Z",
  "results": {
    "gte-modernbert-base": {
      "start_time": "2024-01-01T10:00:00Z",
      "end_time": "2024-01-01T10:05:00Z",
      "query_results": [
        {
          "query_id": "query_1",
          "query": "type UserService struct",
          "metrics": {
            "precision": 1.0,
            "recall": 1.0,
            "f1_score": 1.0,
            "map": 1.0,
            "ndcg": 1.0,
            "response_time": 150.50
          },
          "retrieved": [...],
          "expected": ["sample1.go"]
        }
      ],
      "average_metrics": {
        "precision": 0.8500,
        "recall": 0.9000,
        "f1_score": 0.8742,
        "map": 0.8234,
        "ndcg": 0.8567,
        "response_time": 150.50
      }
    }
  }
}
```

#### 结果说明
- **query_results**: 包含每个查询的详细结果，包括检索到的文件和评估指标
- **average_metrics**: 所有查询的平均指标，用于整体性能评估
- **MAP** 和 **NDCG**: 新增的高级指标，提供更全面的性能评估
- **response_time**: 包含网络传输、向量计算和数据库查询的完整时间
- **retrieved**: 检索到的文件列表，包含文件路径、内容片段和相似度分数
- **expected**: 期望匹配的文件列表，用于计算评估指标

## 测试场景

### 1. 嵌入模型对比测试
- **目的**: 评估不同嵌入模型对召回效果的影响
- **方法**: 使用相同的向量数据库，测试不同的嵌入模型
- **配置**: 在 `scenarios.embedder_comparison` 中配置

### 2. 向量数据库对比测试
- **目的**: 评估不同向量数据库配置对召回效果的影响
- **方法**: 使用相同的嵌入模型，测试不同的向量数据库
- **配置**: 在 `scenarios.vector_store_comparison` 中配置

## 故障排除

### 常见问题

1. **编译错误**: 确保所有依赖已正确安装
2. **连接失败**: 检查向量数据库服务是否正常运行
3. **API密钥错误**: 确保环境变量中的API密钥正确
4. **测试数据缺失**: 确保 `testdata/` 目录中的文件存在

### 调试方法

1. **查看详细日志**: 使用 `-v` 参数运行测试
2. **运行单个测试**: 使用 `-run` 参数指定特定测试
3. **检查配置**: 验证 `config.yaml` 配置是否正确
4. **验证环境**: 确保外部服务（如Weaviate）可访问

## 扩展测试

### 添加新的嵌入模型
1. 在 `config.yaml` 的 `embedders` 部分添加新配置
2. 在 `scenarios.embedder_comparison.embedders` 中添加模型名称
3. 确保API密钥和端点配置正确

### 添加新的向量数据库
1. 在 `config.yaml` 的 `vector_stores` 部分添加新配置
2. 在 `scenarios.vector_store_comparison.vector_stores` 中添加数据库名称
3. 配置相应的连接参数和认证信息

### 添加新的测试数据
1. **代码文件**: 添加到 `testdata/code_samples/` 目录
2. **文档文件**: 添加到 `testdata/doc_samples/` 目录，支持Markdown和JSON格式
3. 确保文件大小在限制范围内（默认10MB）
4. 文件命名应具有描述性，便于识别和测试

### 添加新的测试查询
1. 在 `testdata/queries.json` 或 `testdata/queries.yaml` 的 `queries` 部分添加新查询
2. 指定期望的文件匹配结果，支持跨文件类型的查询
3. 必须指定场景类型（`scenario_type: "code"` 或 `scenario_type: "doc"`）
4. 可以指定特定语言类型的查询（如 `language: "markdown"`、`language: "json"`）
5. 保持代码用例与文档用例的平衡比例
6. 查询文本应具有代表性，覆盖实际使用场景

### 添加新的场景类型
测试框架支持扩展新的场景类型，步骤如下：

1. **定义场景类型**:
   ```yaml
   scenarios:
     new_scenario_type:
       name: "新场景类型测试"
       description: "测试新场景类型的检索效果"
       vector_store: "weaviate-default"
       embedders: ["gte-modernbert-base"]
       top_k: 5
       scenario_filter: "custom_filter"  # 可选的场景过滤器
   ```

2. **实现场景过滤器**（可选）:
   - 在 `test_runner.go` 中添加自定义过滤逻辑
   - 根据业务需求筛选特定的查询或文件

3. **更新测试结果分析**:
   - 在 `metrics_calculator.go` 中添加针对性的指标计算
   - 更新结果报告格式以包含新场景类型

4. **添加配置验证**:
   - 在配置加载时验证新场景类型的参数
   - 提供清晰的错误提示和默认值

### 自定义评估指标
如果需要添加新的评估指标：

1. **扩展指标结构**:
   ```go
   type Metrics struct {
       Precision    float64 `json:"precision"`
       Recall       float64 `json:"recall"`
       F1Score      float64 `json:"f1_score"`
       MAP          float64 `json:"map"`
       NDCG         float64 `json:"ndcg"`
       ResponseTime float64 `json:"response_time"`
       // 添加新指标
       CustomMetric float64 `json:"custom_metric"`
   }
   ```

2. **实现计算逻辑**:
   - 在 `metrics_calculator.go` 中添加计算方法
   - 确保计算逻辑与业务需求一致

3. **更新结果展示**:
   - 修改结果输出格式以包含新指标
   - 更新文档说明新指标的含义和计算方法

## 注意事项

1. **测试时间**: 完整测试可能需要较长时间，建议在非生产环境运行
2. **资源消耗**: 测试会消耗嵌入模型API配额和向量数据库资源
3. **数据清理**: 测试完成后，建议清理向量数据库中的测试数据
4. **结果分析**: 结合业务需求分析测试结果，选择最适合的配置

## 贡献指南

如需改进测试框架，请：
1. 确保新功能向后兼容
2. 添加相应的测试用例
3. 更新文档说明
4. 遵循项目代码规范