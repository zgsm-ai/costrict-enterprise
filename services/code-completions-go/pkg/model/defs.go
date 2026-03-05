package model

//	前置模块处理完毕后给到模型进行调用的参数信息
type CompletionParameter struct {
	CompletionID string   `json:"completionID"` // 补全请求ID，用于唯一标识一次补全请求
	ClientID     string   `json:"clientID"`     // 用户ID，唯一标识发起补全请求的用户
	Language     string   `json:"language"`     // 编程语言
	Model        string   `json:"model"`        // 模型
	MaxTokens    int      `json:"max_tokens"`   // 回复内容的最大token数
	Temperature  float32  `json:"temperature"`  // 温度
	Stop         []string `json:"stop"`         // 停止符
	Prefix       string   `json:"prefix"`       // 前缀
	Suffix       string   `json:"suffix"`       // 后缀
	CodeContext  string   `json:"context"`      // 上下文
	Verbose      bool     `json:"verbose"`      // 是否需要更详细的回复，帮助调试
}

type CompletionVerbose struct {
	Id     string                 `json:"id"`
	Input  map[string]interface{} `json:"input"`
	Output map[string]interface{} `json:"output,omitempty"`
}

type CompletionStatus string

const (
	StatusSuccess     CompletionStatus = "success"     //补全成功
	StatusReqError    CompletionStatus = "reqError"    //请求存在错误
	StatusServerError CompletionStatus = "serverError" //服务端错误
	StatusModelError  CompletionStatus = "modelError"  //模型响应错误
	StatusEmpty       CompletionStatus = "empty"       //补全结果为空
	StatusRejected    CompletionStatus = "rejected"    //根据规则拒绝补全
	StatusTimeout     CompletionStatus = "timeout"     //补全请求超时
	StatusCanceled    CompletionStatus = "canceled"    //用户取消
	StatusBusy        CompletionStatus = "busy"        //服务端繁忙
)

//	OpenAI v1/completions协议的请求和响应结构定义
//
// 请求体结构(参考：https://api-docs.deepseek.com/zh-cn/api/create-completion)
type CompletionRequest struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      float32  `json:"temperature,omitempty"`
	TopP             float32  `json:"top_p,omitempty"`
	FrequencyPenalty float32  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32  `json:"presence_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	Stream           bool     `json:"stream,omitempty"`
	Echo             bool     `json:"echo,omitempty"`
	Suffix           string   `json:"suffix,omitempty"`
}

type CompletionChoice struct {
	Text         string      `json:"text"`
	Index        int         `json:"index"`
	Logprobs     interface{} `json:"logprobs,omitempty"`
	FinishReason string      `json:"finish_reason"`
}

type CompletionUsage struct {
	PromptTokens            int         `json:"prompt_tokens"`
	CompletionTokens        int         `json:"completion_tokens"`
	TotalTokens             int         `json:"total_tokens"`
	PromptCacheHitTokens    int         `json:"prompt_cache_hit_tokens,omitempty"`
	PromptCacheMissTokens   int         `json:"prompt_cache_miss_tokens,omitempty"`
	CompletionTokensDetails interface{} `json:"completion_tokens_details,omitempty"`
}

// 响应体结构
type CompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int                `json:"created"`
	Model             string             `json:"model"`
	Choices           []CompletionChoice `json:"choices"`
	Usage             CompletionUsage    `json:"usage"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
}
