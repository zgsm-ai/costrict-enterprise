package completions

import (
	"completion-agent/pkg/metrics"
	"completion-agent/pkg/model"
	"fmt"
	"time"
)

/**
 * 补全使用情况统计结构体
 * @description
 * - 记录补全请求的token使用情况
 * - 包含输入token数、输出token数和总token数
 * - 用于监控和统计API使用情况
 * - 嵌入到CompletionResponse中返回给客户端
 */
type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

/**
 * 补全选择结构体
 * @description
 * - 表示补全请求的一个选择结果
 * - 包含生成的文本内容
 * - 支持多个选择结果，按优先级排序
 * - 用于向客户端返回补全建议
 */
type CompletionChoice struct {
	Text string `json:"text"`
}

/**
 * 补全性能统计结构体(Compatible with openai/v1 CompletionUsage)
 * @description
 * - 记录补全请求各阶段的性能数据
 * - 包含接收时间、上下文获取时间、排队时间、LLM处理时间和总时间
 * - 记录token使用统计信息
 * - 用于性能监控和优化分析
 */
type CompletionPerformance struct {
	ReceiveTime      time.Time `json:"receive_time"`      //收到请求的时间
	ContextDuration  int64     `json:"context_duration"`  //获取上下文的时长(毫秒)
	LLMDuration      int64     `json:"llm_duration"`      //调用大语言模型耗用的时长(毫秒)
	TotalDuration    int64     `json:"total_duration"`    //总时长(毫秒)
	PromptTokens     int       `json:"prompt_tokens"`     //提示词token数
	CompletionTokens int       `json:"completion_tokens"` //补全结果token数
	TotalTokens      int       `json:"total_tokens"`      //总token数
}

/**
 * 补全响应结构体
 * @description
 * - 表示补全请求的完整响应
 * - 包含响应ID、模型名称、补全选择列表、使用统计和状态
 * - 支持错误信息和详细输出
 * - 用于向客户端返回补全结果
 */
type CompletionResponse struct {
	ID      string                   `json:"id"`
	Model   string                   `json:"model"`
	Object  string                   `json:"object"`
	Choices []CompletionChoice       `json:"choices"`
	Created int                      `json:"created"`
	Usage   CompletionPerformance    `json:"usage"`
	Status  model.CompletionStatus   `json:"status"`
	Error   string                   `json:"error,omitempty"`
	Verbose *model.CompletionVerbose `json:"verbose,omitempty"`
}

/**
 * 记录补全性能指标
 * @param {string} modelName - 模型名称，用于指标分类
 * @param {string} status - 补全状态字符串，用于结果分类
 * @param {*CompletionPerformance} perf - 性能统计对象，包含各阶段耗时和token使用情况
 * @description
 * - 记录补全请求的各阶段耗时指标
 * - 记录补全请求计数指标
 * - 记录输入和输出token使用指标
 * - 使用metrics包进行指标上报
 * - 用于监控补全服务的性能和资源使用情况
 */
func Metrics(modelName string, status string, perf *CompletionPerformance) {
	metrics.RecordCompletionDuration(modelName, status,
		0, perf.ContextDuration, perf.LLMDuration, perf.TotalDuration)
	metrics.IncrementCompletionRequests(modelName, status)
	metrics.RecordCompletionTokens(modelName, metrics.TokenTypeInput, perf.PromptTokens)
	metrics.RecordCompletionTokens(modelName, metrics.TokenTypeOutput, perf.CompletionTokens)
}

/**
 * 创建错误响应
 * @param {string} completionId - 补全请求ID
 * @param {string} modelName - 模型名称
 * @param {model.CompletionStatus} status - 补全状态，表示错误类型
 * @param {*CompletionPerformance} perf - 性能统计对象，包含耗时和token信息
 * @param {*model.CompletionVerbose} verbose - 详细输出信息
 * @param {error} err - 错误对象，包含错误详情
 * @returns {*CompletionResponse} 返回错误响应对象
 * @description
 * - 创建表示错误的补全响应
 * - 如果错误为nil，使用状态字符串作为错误信息
 * - 记录性能指标到监控系统
 * - 设置空的选择结果
 * - 包含错误详情和性能统计信息
 */
func ErrorResponse(completionId, modelName string, status model.CompletionStatus,
	perf *CompletionPerformance, verbose *model.CompletionVerbose, err error) *CompletionResponse {
	if err == nil {
		err = fmt.Errorf("%s", string(status))
	}
	perf.TotalDuration = time.Since(perf.ReceiveTime).Milliseconds()
	Metrics(modelName, string(status), perf)
	return &CompletionResponse{
		ID:      completionId,
		Model:   modelName,
		Object:  "text_completion",
		Choices: []CompletionChoice{{Text: ""}}, // 使用后置处理后的补全结果
		Created: int(perf.ReceiveTime.Unix()),
		Usage:   *perf,
		Status:  status,
		Error:   err.Error(),
		Verbose: verbose,
	}
}

/**
 * 创建成功响应
 * @param {string} completionId - 补全请求ID
 * @param {string} modelName - 模型名称
 * @param {string} completionText - 补全文本内容，表示生成的代码
 * @param {*CompletionPerformance} perf - 性能统计对象，包含耗时和token信息
 * @param {*model.CompletionVerbose} verbose - 详细输出信息
 * @returns {*CompletionResponse} 返回成功响应对象
 * @description
 * - 创建表示成功的补全响应
 * - 设置状态为成功
 * - 记录性能指标到监控系统
 * - 包含补全文本和性能统计信息
 * - 不包含错误信息
 */
func SuccessResponse(completionId, modelName, completionText string, perf *CompletionPerformance,
	verbose *model.CompletionVerbose) *CompletionResponse {

	perf.TotalDuration = time.Since(perf.ReceiveTime).Milliseconds()
	Metrics(modelName, string(model.StatusSuccess), perf)
	return &CompletionResponse{
		ID:      completionId,
		Model:   modelName,
		Object:  "text_completion",
		Choices: []CompletionChoice{{Text: completionText}}, // 使用后置处理后的补全结果
		Created: int(perf.ReceiveTime.Unix()),
		Usage:   *perf,
		Status:  model.StatusSuccess,
		Verbose: verbose,
	}
}

/**
 * 创建取消请求响应
 * @param {string} completionId - 补全请求ID
 * @param {string} modelName - 模型名称
 * @param {*CompletionPerformance} perf - 性能统计对象，包含耗时和token信息
 * @param {model.CompletionStatus} status - 补全状态，表示取消类型
 * @param {error} err - 错误对象，包含取消原因
 * @returns {*CompletionResponse} 返回取消请求响应对象
 * @description
 * - 创建表示请求取消的补全响应
 * - 根据错误类型判断是超时还是主动取消
 * - 计算总耗时并记录性能指标
 * - 设置空的选择结果
 * - 包含错误详情和性能统计信息
 */
func CancelRequest(completionId, modelName string, perf *CompletionPerformance,
	status model.CompletionStatus, err error) *CompletionResponse {
	perf.TotalDuration = time.Since(perf.ReceiveTime).Milliseconds()
	Metrics(modelName, string(status), perf)
	return &CompletionResponse{
		ID:      completionId,
		Model:   modelName,
		Object:  "text_completion",
		Choices: []CompletionChoice{{Text: ""}},
		Created: int(perf.ReceiveTime.Unix()),
		Usage:   *perf,
		Status:  status,
		Error:   err.Error(),
	}
}
