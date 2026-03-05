package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ModelConfig struct {
	Provider       string        `json:"provider" yaml:"provider"`             // 模型供应商，代表着具体的模型接口/类型
	ModelTitle     string        `json:"modelTitle" yaml:"modelTitle"`         // 模型来源的唯一标识
	ModelName      string        `json:"modelName" yaml:"modelName"`           // 真实的模型名称
	CompletionsUrl string        `json:"completionsUrl" yaml:"completionsUrl"` // 补全地址
	Tags           []string      `json:"tags" yaml:"tags"`                     // 模型标签，用户可以根据标签选择补全模型
	Authorization  string        `json:"authorization" yaml:"authorization"`   // 认证信息
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`               // 超时时间ms
	MaxPrefix      int           `json:"maxPrefix" yaml:"maxPrefix"`           // 最大模型上下文长度:前缀
	MaxSuffix      int           `json:"maxSuffix" yaml:"maxSuffix"`           // 最大模型上下文长度:后缀
	MaxOutput      int           `json:"maxOutput" yaml:"maxOutput"`           // 最大输出token数
	FimMode        bool          `json:"fimMode" yaml:"fimMode"`               // 填充FIM标记的模式
	FimBegin       string        `json:"fimBegin" yaml:"fimBegin"`             // 开始
	FimEnd         string        `json:"fimEnd" yaml:"fimEnd"`                 // 结束
	FimHole        string        `json:"fimHole" yaml:"fimHole"`               // 待补全的空洞位置
	FimStop        []string      `json:"fimStop" yaml:"fimStop"`               // 结束符
	TokenizerPath  string        `json:"tokenizerPath" yaml:"tokenizerPath"`   // tokenizer json 路径
	MaxConcurrent  int           `json:"maxConcurrent" yaml:"maxConcurrent"`   // 每种模型的最大并发数，防止模型过载
	DisablePrune   bool          `json:"disablePrune" yaml:"disablePrune"`     // 禁止后期修剪
	CustomPruners  []string      `json:"customPruners" yaml:"customPruners"`   // 自定义的后期修剪工具
}

/**
 * 关系链查询配置结构体，定义了代码关系查询的相关参数
 * @description
 * - 控制是否启用代码关系链查询功能
 * - 设置关系查询服务的URL地址
 * - 配置查询的层级深度
 * - 控制是否在结果中包含内容详情
 * @example
 * {
 *   "disabled": false,
 *   "url": "http://localhost:8081/relation",
 *   "layer": 3,
 *   "includeContent": true
 * }
 */
type RelationConfig struct {
	Disabled       bool   `json:"disabled" yaml:"disabled"`             // 是否禁用关系链查询
	Url            string `json:"url" yaml:"url"`                       // 关系查询服务地址
	Layer          int    `json:"layer" yaml:"layer"`                   // 查询层级深度
	IncludeContent bool   `json:"includeContent" yaml:"includeContent"` // 是否包含内容详情
}

/**
 * 语义相关性查询配置结构体，定义了语义搜索的相关参数
 * @description
 * - 控制是否启用语义相关性查询功能
 * - 设置语义查询服务的URL地址
 * - 配置返回结果的数量上限(TopK)
 * - 设置结果分数的阈值，过滤低相关性结果
 * @example
 * {
 *   "disabled": false,
 *   "url": "http://localhost:8082/semantic",
 *   "topK": 10,
 *   "scoreThreshold": 0.5
 * }
 */
type SemanticConfig struct {
	Disabled       bool    `json:"disabled" yaml:"disabled"`             // 是否禁用语义查询
	Url            string  `json:"url" yaml:"url"`                       // 语义查询服务地址
	TopK           int     `json:"topK" yaml:"topK"`                     // 返回结果数量上限
	ScoreThreshold float64 `json:"scoreThreshold" yaml:"scoreThreshold"` // 结果分数阈值
}

/**
 * 定义查询配置结构体，定义了代码定义查询的相关参数
 * @description
 * - 控制是否启用代码定义查询功能
 * - 设置定义查询服务的URL地址
 * - 用于获取代码中标识符的定义信息
 * @example
 * {
 *   "disabled": false,
 *   "url": "http://localhost:8083/definition"
 * }
 */
type DefinitionConfig struct {
	Disabled bool   `json:"disabled" yaml:"disabled"` // 是否禁用定义查询
	Url      string `json:"url" yaml:"url"`           // 定义查询服务地址
}

/**
 * 上下文配置结构体，定义了代码补全的上下文获取配置
 * @description
 * - 包含定义查询、语义查询和关系链查询的配置
 * - 设置单个请求的超时时间
 * - 设置整个上下文获取过程的总超时时间
 * - 用于控制代码补全时获取相关代码上下文的行为
 * @example
 * {
 *   "definition": {
 *     "disabled": false,
 *     "url": "http://localhost:8083/definition"
 *   },
 *   "semantic": {
 *     "disabled": false,
 *     "url": "http://localhost:8082/semantic",
 *     "topK": 10,
 *     "scoreThreshold": 0.5
 *   },
 *   "relation": {
 *     "disabled": false,
 *     "url": "http://localhost:8081/relation",
 *     "layer": 3,
 *     "includeContent": true
 *   },
 *   "requestTimeout": "5s",
 *   "totalTimeout": "15s"
 * }
 */
type ContextConfig struct {
	Definition     DefinitionConfig `json:"definition" yaml:"definition"`         // 定义查询配置
	Semantic       SemanticConfig   `json:"semantic" yaml:"semantic"`             // 语义相关性查询配置
	Relation       RelationConfig   `json:"relation" yaml:"relation"`             // 关系链查询配置
	RequestTimeout time.Duration    `json:"requestTimeout" yaml:"requestTimeout"` // 单个请求超时时间
	TotalTimeout   time.Duration    `json:"totalTimeout" yaml:"totalTimeout"`     // 上下文获取总超时时间
}

/**
 * 隐藏分过滤器配置结构体，定义了基于隐藏分数的过滤规则
 * @description
 * - 控制是否启用隐藏分数过滤功能
 * - 设置接受补全的最低分数阈值
 * - 用于过滤低质量的补全建议
 * - 分数基于上下文特征计算，如语言类型、光标位置等
 * @example
 * {
 *   "disabled": false,
 *   "threshold": 0.3
 * }
 */
type ScoreFilterConfig struct {
	Disabled  bool    `json:"disabled" yaml:"disabled"`   // 是否禁用隐藏分过滤
	Threshold float64 `json:"threshold" yaml:"threshold"` // 接受补全的最低分数阈值
}

/**
 * 语法过滤器配置结构体，定义了基于语法特征的过滤规则
 * @description
 * - 控制是否启用语法过滤功能
 * - 设置过滤阈值和各种模式匹配规则
 * - 定义最少提示行数和结束标签
 * - 用于判断是否应该触发代码补全
 * @example
 * {
 *   "disabled": false,
 *   "threshold": 0.5,
 *   "strPattern": "import +.*|from +.*|from +.* import *.*",
 *   "treePattern": "\\(comment.*|\\(string.*|\\(set \\(string.*|\\(dictionary.*|\\(integer.*|\\(list.*|\\(tuple.*",
 *   "minPromptLine": 5,
 *   "endTag": "('>',';','}',')')"
 * }
 */
type SyntaxFilterConfig struct {
	Disabled      bool   `json:"disabled" yaml:"disabled"`           // 是否禁用语法过滤
	StrPattern    string `json:"strPattern" yaml:"strPattern"`       // 字符串匹配模式
	TreePattern   string `json:"treePattern" yaml:"treePattern"`     // 语法树匹配模式
	MinPromptLine int    `json:"minPromptLine" yaml:"minPromptLine"` // 触发补全的最少提示行数
	EndTag        string `json:"endTag" yaml:"endTag"`               // 光标行结束标签
}

/**
 * 后期修剪配置结构体，定义了补全结果的后期处理规则
 * @description
 * - 控制是否启用后期修剪功能
 * - 配置使用的修剪工具列表
 * - 用于对补全结果进行后处理，提高质量
 * @example
 * {
 *   "disabled": false,
 *   "pruners": ["deduplication", "formatting", "validation"]
 * }
 */
type PruneConfig struct {
	Disabled bool     `json:"disabled" yaml:"disabled"` // 是否禁用后期修剪
	Pruners  []string `json:"pruners" yaml:"pruners"`   // 自定义的后期修剪工具列表
}

/**
 * 包装器配置结构体，定义了补全前后处理的各种过滤器配置
 * @description
 * - 包含隐藏分过滤器的配置，用于质量过滤
 * - 包含语法过滤器的配置，用于语法判断
 * - 包含后期修剪的配置，用于结果优化
 * - 包含分词器的配置，用于文本预处理
 * - 用于控制补全请求的前后处理流程
 * @example
 * {
 *   "score": {
 *     "disabled": false,
 *     "threshold": 0.3
 *   },
 *   "syntax": {
 *     "disabled": false,
 *     "threshold": 0.5,
 *     "strPattern": "import +.*|from +.*|from +.* import *.*",
 *     "treePattern": "\\(comment.*|\\(string.*|\\(set \\(string.*|\\(dictionary.*|\\(integer.*|\\(list.*|\\(tuple.*",
 *     "minPromptLine": 5,
 *     "endTag": "('>',';','}',')')"
 *   },
 *   "prune": {
 *     "disabled": false,
 *     "pruners": ["deduplication", "formatting", "validation"]
 *   },
 *   "tokenizer": {
 *     "path": "/path/to/tokenizer"
 *   }
 * }
 */
type WrapperConfig struct {
	Score  ScoreFilterConfig  `json:"score" yaml:"score"`   // 隐藏分过滤器配置
	Syntax SyntaxFilterConfig `json:"syntax" yaml:"syntax"` // 语法过滤器配置
	Prune  PruneConfig        `json:"prune" yaml:"prune"`   // 后期修剪配置
}

type StreamControllerConfig struct {
	MaintainInterval  time.Duration `json:"maintainInterval" yaml:"maintainInterval"`   // 定时维护的间隔
	CleanOlderThan    time.Duration `json:"cleanOlderThan" yaml:"cleanOlderThan"`       // 清理过期客户端的最大间隔
	CompletionTimeout time.Duration `json:"completionTimeout" yaml:"completionTimeout"` // 一个补全请求的最大超时
	QueueTimeout      time.Duration `json:"queueTimeout" yaml:"queueTimeout"`           // 排队超时
}

type SoftwareConfig struct {
	Models           []ModelConfig          `json:"models" yaml:"models"`                     // AI模型配置列表
	Context          ContextConfig          `json:"context" yaml:"context"`                   // 上下文获取配置
	Wrapper          WrapperConfig          `json:"wrapper" yaml:"wrapper"`                   // 补全前后处理配置
	StreamController StreamControllerConfig `json:"streamController" yaml:"streamController"` // 全局流控配置
}

var Config = &SoftwareConfig{}
var Context *ContextConfig = &Config.Context
var Wrapper *WrapperConfig = &Config.Wrapper

func resetDefValues(c *SoftwareConfig) {
	if c.StreamController.QueueTimeout == 0 {
		c.StreamController.QueueTimeout = 200 * time.Millisecond
	}
	if c.StreamController.CompletionTimeout == 0 {
		c.StreamController.CompletionTimeout = 2500 * time.Millisecond
	}
	if c.StreamController.CleanOlderThan == 0 {
		c.StreamController.CleanOlderThan = 1 * time.Hour
	}
}

func init() {
	// 读取配置文件
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		return
	}
	configFileStr := strings.ReplaceAll(string(configFile), "\r\n", "\n")

	// 解析 YAML 配置
	err = yaml.Unmarshal([]byte(configFileStr), Config)
	if err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
		panic(err)
	}
	resetDefValues(Config)
	data, _ := json.MarshalIndent(Config, "", "  ")
	fmt.Printf("配置文件加载成功:\n%s\n", string(data))
}
