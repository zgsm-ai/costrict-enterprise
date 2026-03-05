package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

/**
 * 模型配置结构体，定义了单个AI模型的配置参数
 * @description
 * - 包含模型的基本信息如提供商、名称、标题等
 * - 定义了模型请求的URL和认证信息
 * - 设置了模型请求的各种限制参数
 * - 支持FIM(Fill in the Middle)模式的配置
 * @example
 * {
 *   "provider": "openai",
 *   "modelTitle": "GPT-4",
 *   "modelName": "gpt-4",
 *   "completionsUrl": "https://api.openai.com/v1/completions",
 *   "tags": ["code", "general"],
 *   "authorization": "Bearer your-api-key",
 *   "timeout": "30s",
 *   "maxPrefix": 2048,
 *   "maxSuffix": 2048,
 *   "maxOutput": 256,
 *   "fimMode": true,
 *   "fimBegin": "<|fim_prefix|>",
 *   "fimEnd": "<|fim_suffix|>",
 *   "fimHole": "<|fim_middle|>",
 *   "fimStop": ["<|endoftext|>"]
 * }
 */
type ModelConfig struct {
	Provider       string   `json:"provider"`                // 模型供应商，代表着具体的模型接口/类型
	ModelTitle     string   `json:"modelTitle,omitempty"`    // 模型的标题，方便用户区分不同的模型来源
	ModelName      string   `json:"modelName"`               // 真实的模型名称
	CompletionsUrl string   `json:"completionsUrl"`          // 补全地址
	Tags           []string `json:"tags"`                    // 模型标签，用户可以根据标签选择补全模型
	Authorization  string   `json:"authorization,omitempty"` // 认证信息
	Timeout        duration `json:"timeout"`                 // 超时时间ms
	MaxPrefix      int      `json:"maxPrefix"`               // 最大前缀token数
	MaxSuffix      int      `json:"maxSuffix"`               // 最大后缀token数
	MaxOutput      int      `json:"maxOutput"`               // 最大输出token数
	FimMode        bool     `json:"fimMode,omitempty"`       // 填充FIM标记的模式
	FimBegin       string   `json:"fimBegin,omitempty"`      // 开始
	FimEnd         string   `json:"fimEnd,omitempty"`        // 结束
	FimHole        string   `json:"fimHole,omitempty"`       // 待补全的空洞位置
	FimStop        []string `json:"fimStop,omitempty"`       // 结束符
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
	Disabled       bool   `json:"disabled"`       // 是否禁用关系链查询
	Url            string `json:"url"`            // 关系查询服务地址
	Layer          int    `json:"layer"`          // 查询层级深度
	IncludeContent bool   `json:"includeContent"` // 是否包含内容详情
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
	Disabled       bool    `json:"disabled"`       // 是否禁用语义查询
	Url            string  `json:"url"`            // 语义查询服务地址
	TopK           int     `json:"topK"`           // 返回结果数量上限
	ScoreThreshold float64 `json:"scoreThreshold"` // 结果分数阈值
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
	Disabled bool   `json:"disabled"` // 是否禁用定义查询
	Url      string `json:"url"`      // 定义查询服务地址
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
	Definition     DefinitionConfig `json:"definition"`     // 定义查询配置
	Semantic       SemanticConfig   `json:"semantic"`       // 语义相关性查询配置
	Relation       RelationConfig   `json:"relation"`       // 关系链查询配置
	RequestTimeout duration         `json:"requestTimeout"` // 单个请求超时时间
	TotalTimeout   duration         `json:"totalTimeout"`   // 上下文获取总超时时间
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
	Disabled  bool    `json:"disabled"`  // 是否禁用隐藏分过滤
	Threshold float64 `json:"threshold"` // 接受补全的最低分数阈值
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
	Disabled      bool   `json:"disabled"`      // 是否禁用语法过滤
	StrPattern    string `json:"strPattern"`    // 字符串匹配模式
	TreePattern   string `json:"treePattern"`   // 语法树匹配模式
	MinPromptLine int    `json:"minPromptLine"` // 触发补全的最少提示行数
	EndTag        string `json:"endTag"`        // 光标行结束标签
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
	Disabled bool     `json:"disabled"` // 是否禁用后期修剪
	Pruners  []string `json:"pruners"`  // 自定义的后期修剪工具列表
}

/**
 * 分词器配置结构体，定义了文本分词的相关参数
 * @description
 * - 设置分词器文件的路径
 * - 用于代码文本的预处理和tokenization
 * - 是补全模型输入处理的重要组件
 * @example
 * {
 *   "path": "/path/to/tokenizer"
 * }
 */
type TokenizerConfig struct {
	Path string `json:"path"` // 分词器文件路径
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
	Score     ScoreFilterConfig  `json:"score"`     // 隐藏分过滤器配置
	Syntax    SyntaxFilterConfig `json:"syntax"`    // 语法过滤器配置
	Prune     PruneConfig        `json:"prune"`     // 后期修剪配置
	Tokenizer TokenizerConfig    `json:"tokenizer"` // 分词器配置
}

/**
 * 软件配置结构体，定义了整个应用程序的配置
 * @description
 * - 包含所有AI模型的配置列表
 * - 包含上下文获取的相关配置
 * - 包含补全前后处理的过滤器配置
 * - 是应用程序的主要配置结构
 * @example
 * {
 *   "models": [
 *     {
 *       "provider": "openai",
 *       "modelTitle": "GPT-4",
 *       "modelName": "gpt-4",
 *       "completionsUrl": "https://api.openai.com/v1/completions",
 *       "tags": ["code", "general"],
 *       "authorization": "Bearer your-api-key",
 *       "timeout": "30s",
 *       "maxPrefix": 2048,
 *       "maxSuffix": 2048,
 *       "maxOutput": 256,
 *       "fimMode": true,
 *       "fimBegin": "<|fim_prefix|>",
 *       "fimEnd": "<|fim_suffix|>",
 *       "fimHole": "<|fim_middle|>",
 *       "fimStop": ["<|endoftext|>"]
 *     }
 *   ],
 *   "context": {
 *     "definition": {
 *       "disabled": false,
 *       "url": "http://localhost:8083/definition"
 *     },
 *     "semantic": {
 *       "disabled": false,
 *       "url": "http://localhost:8082/semantic",
 *       "topK": 10,
 *       "scoreThreshold": 0.5
 *     },
 *     "relation": {
 *       "disabled": false,
 *       "url": "http://localhost:8081/relation",
 *       "layer": 3,
 *       "includeContent": true
 *     },
 *     "requestTimeout": "5s",
 *     "totalTimeout": "15s"
 *   },
 *   "wrapper": {
 *     "score": {
 *       "disabled": false,
 *       "threshold": 0.3
 *     },
 *     "syntax": {
 *       "disabled": false,
 *       "threshold": 0.5,
 *       "strPattern": "import +.*|from +.*|from +.* import *.*",
 *       "treePattern": "\\(comment.*|\\(string.*|\\(set \\(string.*|\\(dictionary.*|\\(integer.*|\\(list.*|\\(tuple.*",
 *       "minPromptLine": 5,
 *       "endTag": "('>',';','}',')')"
 *     },
 *     "prune": {
 *       "disabled": false,
 *       "pruners": ["deduplication", "formatting", "validation"]
 *     },
 *     "tokenizer": {
 *       "path": "/path/to/tokenizer"
 *     }
 *   }
 * }
 */
type SoftwareConfig struct {
	Models  []ModelConfig `json:"models"`  // AI模型配置列表
	Context ContextConfig `json:"context"` // 上下文获取配置
	Wrapper WrapperConfig `json:"wrapper"` // 补全前后处理配置
}

/**
 * duration是time.Duration的包装器，支持从字符串反序列化JSON
 * @description
 * - 封装time.Duration类型以支持JSON反序列化
 * - 实现json.Unmarshaler接口以自定义反序列化逻辑
 * - 支持字符串和浮点数两种格式的输入
 * - 用于配置文件中的时间持续时间设置
 * @example
 * // 在配置文件中使用
 * {
 *   "timeout": "30s"  // 字符串格式
 *   // 或
 *   "timeout": 30000  // 毫秒数
 * }
 */
type duration struct {
	dur time.Duration
}

/**
 * 实现json.Unmarshaler接口的UnmarshalJSON方法
 * @param {[]byte} data - JSON数据字节切片，包含要反序列化的持续时间值
 * @returns {error} 返回反序列化过程中的错误，成功返回nil
 * @description
 * - 支持从字符串和浮点数两种格式解析持续时间
 * - 字符串格式使用time.ParseDuration解析，支持"ns", "us", "ms", "s", "m", "h"等单位
 * - 浮点数格式直接转换为毫秒数
 * - 对于不支持的格式返回错误
 * @throws
 * - JSON解析失败时返回错误
 * - 字符串时间格式无效时返回错误
 * @example
 * var dur duration
 * err := dur.UnmarshalJSON([]byte("\"30s\""))
 * // dur.dur = 30 * time.Second
 */
func (d *duration) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		var err error
		d.dur, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	case float64:
		d.dur = time.Duration(value)
	default:
		return fmt.Errorf("invalid duration: %#v", v)
	}

	return nil
}

/**
 * 返回底层time.Duration值
 * @returns {time.Duration} 返回封装的time.Duration值
 * @description
 * - 提供访问底层time.Duration的方法
 * - 简单的getter方法
 * - 用于获取实际的持续时间值
 * @example
 * var dur duration
 * dur.dur = 30 * time.Second
 * actualDuration := dur.Duration()
 * // actualDuration = 30 * time.Second
 */
func (d duration) Duration() time.Duration {
	return d.dur
}

// Must run config.LoadConfig() first
var Config *SoftwareConfig
var Context *ContextConfig
var Wrapper *WrapperConfig

/**
 * 获取costrict目录结构设定
 * @returns {string} 返回costrict配置目录的完整路径
 * @description
 * - 获取用户主目录路径
 * - 在用户主目录下创建.costrict子目录
 * - 用于存储配置文件和相关数据
 * - 如果获取用户主目录失败，返回当前目录
 * @example
 * dir := getCostrictDir()
 * // 返回类似: "C:\\Users\\username\\.costrict" (Windows)
 * // 或 "/home/username/.costrict" (Linux/Mac)
 */
func getCostrictDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".costrict")
}

/**
 * 加载本地配置
 * @returns {*SoftwareConfig, error} 返回加载的配置对象和错误，成功时错误为nil
 * @description
 * - 构建配置文件的完整路径
 * - 读取配置文件内容
 * - 将JSON内容反序列化为SoftwareConfig对象
 * - 对配置进行本地化处理
 * - 打印配置信息用于调试
 * - 用于从本地文件加载应用程序配置
 * @throws
 * - 读取文件失败时返回错误
 * - JSON反序列化失败时返回错误
 * @example
 * config, err := loadLocalConfig()
 * if err != nil {
 *     log.Fatalf("加载配置失败: %v", err)
 * }
 */
func loadLocalConfig() (*SoftwareConfig, error) {
	fname := filepath.Join(getCostrictDir(), "config", "completion-agent.json")

	bytes, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("load 'completion-agent.json' failed: %v", err)
	}
	var c SoftwareConfig
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil, fmt.Errorf("unmarshal 'completion-agent.json' failed: %v", err)
	}
	localize(&c)
	fmt.Printf("Config: %+v", &c)
	return &c, nil
}

/**
 * 加载本地配置（单例模式）
 * @returns {error} 返回加载过程中的错误，成功返回nil
 * @description
 * - 检查全局配置对象是否已初始化
 * - 如果已初始化，直接返回nil
 * - 否则加载本地配置文件
 * - 记录配置加载失败的日志信息
 * - 用于应用程序启动时加载配置
 * @throws
 * - 配置文件读取失败时返回错误
 * - JSON反序列化失败时返回错误
 * @example
 * if err := LoadConfig(); err != nil {
 *     log.Fatalf("程序启动失败: %v", err)
 * }
 */
func LoadConfig() error {
	if Config != nil {
		return nil
	}
	cfg, err := loadLocalConfig()
	if err != nil {
		log.Printf("Load failed: %v", err)
		return err
	}
	Config = cfg
	Context = &cfg.Context
	Wrapper = &cfg.Wrapper
	return nil
}
