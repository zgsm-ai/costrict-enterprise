package completions

import (
	"completion-agent/pkg/codebase_context"
	"completion-agent/pkg/config"
	"completion-agent/pkg/model"
	"fmt"
	"net/http"
	"time"
)

/**
 * 补全输入结构体
 * @description
 * - 封装补全请求的所有输入信息
 * - 包含原始请求、HTTP头部、处理后的提示词和选择的模型
 * - 提供预处理方法用于准备补全请求
 * - 用于在补全处理流程中传递数据
 * @example
 * input := &CompletionInput{
 *     CompletionRequest: CompletionRequest{...},
 *     Headers: http.Header{},
 * }
 * ctx := NewCompletionContext(context.Background(), &CompletionPerformance{})
 * response := input.Preprocess(ctx)
 */
type CompletionInput struct {
	CompletionRequest             //原始请求中的BODY
	Headers           http.Header //原始请求中的头部
}

/**
 * 代码上下文客户端实例
 * @description
 * - 全局单例，用于获取代码上下文信息
 * - 在GetContext方法中延迟初始化
 * - 提供代码库上下文查询功能
 * - 用于增强补全请求的上下文信息
 */
var contextClient *codebase_context.ContextClient

/**
 * 处理补全请求
 * @param {*CompletionContext} c - 补全上下文，包含请求上下文和性能统计信息
 * @returns {*CompletionResponse} 返回补全响应对象，如果预处理失败则返回错误响应
 * @description
 * - 执行补全请求的预处理流程
 * - 首先通过过滤器链处理补全拒绝规则
 * - 如果拒绝规则匹配，返回拒绝响应
 * - 解析请求参数获取提示词
 * - 获取代码上下文信息
 * - 是补全处理的第一步
 * @throws
 * - 如果过滤器链处理失败，返回拒绝响应
 * @example
 * input := &CompletionInput{...}
 * ctx := NewCompletionContext(context.Background(), &CompletionPerformance{})
 * response := input.Preprocess(ctx)
 * if response != nil {
 *     // 预处理失败或被拒绝
 * }
 */
func (in *CompletionInput) Preprocess(c *CompletionContext) *CompletionResponse {
	if err := in.GetPrompts(); err != nil {
		return CancelRequest(in.CompletionID, in.Model, c.Perf, model.StatusRejected, err)
	}
	// 1. 补全拒绝规则链处理
	err := NewFilterChain(config.Wrapper).Handle(in)
	if err != nil {
		return CancelRequest(in.CompletionID, in.Model, c.Perf, model.StatusRejected, err)
	}
	// 2. 获取上下文信息
	in.GetContext(c)
	return nil
}

/**
 * 获取上下文信息
 * @param {*CompletionContext} c - 补全上下文，包含请求上下文和性能统计信息
 * @description
 * - 如果代码上下文已存在，直接返回
 * - 延迟初始化上下文客户端
 * - 调用上下文客户端获取代码上下文
 * - 记录获取上下文的耗时
 * - 用于增强补全请求的上下文信息
 */
func (in *CompletionInput) GetContext(c *CompletionContext) {
	if in.Prompts.CodeContext != "" {
		return
	}
	if contextClient == nil {
		contextClient = codebase_context.NewContextClient()
	}
	in.Prompts.CodeContext = contextClient.GetContext(
		c.Ctx,
		in.ClientID,
		in.Prompts.ProjectPath,
		in.Prompts.FileProjectPath,
		in.Prompts.Prefix,
		in.Prompts.Suffix,
		in.Prompts.ImportContent,
		in.Headers,
	)
	c.Perf.ContextDuration = time.Since(c.Perf.ReceiveTime).Milliseconds()
}

/**
 * 解析提示词
 * @description
 * - 从请求中解析提示词选项
 * - 如果请求中包含PromptOptions，直接使用
 * - 否则从简单提示词中提取前缀
 * - 如果行前缀为空，从前缀中提取最后一行
 * - 如果行后缀为空，从后缀中提取第一行
 * - 用于预处理补全请求的提示词
 */
func (in *CompletionInput) GetPrompts() error {
	if in.Prompts == nil {
		return fmt.Errorf("missing 'prompt_options'")
	}
	return nil
}
