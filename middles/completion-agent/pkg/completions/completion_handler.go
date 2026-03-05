package completions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"completion-agent/pkg/config"
	"completion-agent/pkg/env"
	"completion-agent/pkg/model"

	"go.uber.org/zap"
)

/**
 * 补全处理器结构体
 * @description
 * - 封装代码补全的核心处理逻辑
 * - 包含模型配置信息和LLM模型实例
 * - 提供补全请求的预处理和后处理功能
 * - 支持多种编程语言的代码补全
 */
type CompletionHandler struct {
	cfg *config.ModelConfig // 模型配置
	llm model.LLM           // 模型
}

/**
 * 补全上下文结构体
 * @description
 * - 封装补全处理过程中需要的上下文信息
 * - 包含context.Context用于请求控制和超时处理
 * - 包含性能统计信息用于监控补全处理过程
 * - 用于在补全处理的不同阶段传递状态和数据
 * @example
 * perf := &CompletionPerformance{ReceiveTime: time.Now()}
 * ctx := NewCompletionContext(context.Background(), perf)
 */
type CompletionContext struct {
	Ctx  context.Context
	Perf *CompletionPerformance
}

/**
 * 创建新的补全上下文
 * @param {context.Context} ctx - 上下文对象，用于请求控制和超时处理
 * @param {*CompletionPerformance} perf - 性能统计对象，用于记录补全处理性能数据
 * @returns {*CompletionContext} 返回创建的补全上下文对象指针
 * @description
 * - 初始化补全上下文对象
 * - 设置上下文对象和性能统计信息
 * - 用于在补全处理过程中传递状态和数据
 * - 简单的构造函数模式
 * @example
 * perf := &CompletionPerformance{ReceiveTime: time.Now()}
 * ctx := NewCompletionContext(context.Background(), perf)
 */
func NewCompletionContext(ctx context.Context, perf *CompletionPerformance) *CompletionContext {
	return &CompletionContext{
		Ctx:  ctx,
		Perf: perf,
	}
}

/**
 * 创建新的补全处理器
 * @param {model.LLM} m - 大语言模型实例，如果为nil则使用自动选择的模型
 * @returns {*CompletionHandler} 返回初始化好的补全处理器对象指针
 * @description
 * - 创建并初始化补全处理器实例
 * - 如果传入的模型为nil，使用自动选择的模型
 * - 获取模型配置信息并保存到处理器中
 * - 返回可用于处理补全请求的处理器
 * @example
 * handler := NewCompletionHandler(nil)
 * // 使用自动选择的模型
 *
 * customModel := model.GetAutoModel()
 * handler := NewCompletionHandler(customModel)
 * // 使用指定的模型
 */
func NewCompletionHandler(m model.LLM) *CompletionHandler {
	if m == nil {
		m = model.GetAutoModel()
	}
	return &CompletionHandler{
		llm: m,
		cfg: m.Config(),
	}
}

func (h *CompletionHandler) Adapt(input *CompletionInput) *model.CompletionParameter {
	// 3. 补全模型相关的前置处理 （拼接prompt策略，单行/多行补全策略，裁剪过长上下文）
	h.truncatePrompt(h.cfg, input.Prompts)

	// 4. 准备停用词，根据是否单行补全调整停用词
	stopWords := h.prepareStopWords(input)

	// 5. 交给模型处理
	var para model.CompletionParameter
	para.Model = input.Model
	para.ClientID = input.ClientID
	para.CompletionID = input.CompletionID
	para.Language = strings.ToLower(input.LanguageID)
	para.Prefix = input.Prompts.Prefix
	para.Suffix = input.Prompts.Suffix
	para.CodeContext = input.Prompts.CodeContext
	para.Stop = stopWords
	para.MaxTokens = h.cfg.MaxOutput
	para.Temperature = float32(input.Temperature)
	para.Verbose = input.Verbose
	if h.cfg.ModelName != "" {
		para.Model = h.cfg.ModelName
	}
	return &para
}

/**
 * 调用大模型，处理补全请求
 * @param {*CompletionContext} c - 补全上下文，包含请求上下文和性能统计信息
 * @param {*CompletionInput} input - 补全输入，包含请求参数和预处理后的数据
 * @returns {*CompletionResponse} 返回补全响应对象，包含补全结果或错误信息
 * @description
 * - 执行补全请求的完整处理流程
 * - 对输入进行截断处理，确保不超过模型最大长度
 * - 准备停用词列表，控制补全生成
 * - 调用LLM模型进行补全生成
 * - 记录模型处理时间和token使用情况
 * - 对生成的补全结果进行后处理和修剪
 * - 构建并返回最终的补全响应
 * @throws
 * - 模型响应失败时返回错误响应
 * - 补全结果为空时返回空状态响应
 * @example
 * ctx := NewCompletionContext(context.Background(), &CompletionPerformance{})
 * input := &CompletionInput{...}
 * response := handler.CallLLM(ctx, input)
 */
func (h *CompletionHandler) CallLLM(c *CompletionContext, para *model.CompletionParameter) *CompletionResponse {
	modelStartTime := time.Now().Local()
	rsp, completionStatus, err := h.llm.Completions(c.Ctx, para)
	modelEndTime := time.Now().Local()
	c.Perf.LLMDuration = modelEndTime.Sub(modelStartTime).Milliseconds()

	var verbose *model.CompletionVerbose
	if rsp != nil {
		verbose = rsp.Verbose
	}
	if completionStatus != model.StatusSuccess {
		c.Perf.PromptTokens = h.getTokensCount(para.Prefix) + h.getTokensCount(para.CodeContext)
		return ErrorResponse(para.CompletionID, para.Model, completionStatus, c.Perf, verbose, err)
	}

	// 6. 补全后置处理
	var completionText string
	if len(rsp.Choices) > 0 {
		completionText = rsp.Choices[0].Text
	}
	if completionText != "" && !config.Wrapper.Prune.Disabled {
		completionText = h.pruneCompletionCode(completionText, para.Prefix, para.Suffix, para.Language)
	}
	c.Perf.PromptTokens = rsp.Usage.PromptTokens
	c.Perf.CompletionTokens = rsp.Usage.CompletionTokens
	c.Perf.TotalTokens = c.Perf.CompletionTokens + c.Perf.PromptTokens

	if completionText == "" {
		return ErrorResponse(para.CompletionID, para.Model, model.StatusEmpty, c.Perf, verbose, fmt.Errorf("empty"))
	}
	// 7. 构建响应
	if !para.Verbose {
		verbose = nil
	}
	return SuccessResponse(para.CompletionID, para.Model, completionText, c.Perf, verbose)
}

/**
 * 完整处理补全请求
 * @param {*CompletionContext} c - 补全上下文，包含请求上下文和性能统计信息
 * @param {*CompletionInput} input - 补全输入，包含请求参数和预处理后的数据
 * @returns {*CompletionResponse} 返回补全响应对象，包含补全结果或错误信息
 * @description
 * - 提供补全请求的完整处理入口
 * - 首先调用输入的预处理方法进行前置处理
 * - 如果预处理返回响应（如错误或拒绝），直接返回
 * - 否则调用CallLLM方法进行实际的补全处理
 * - 是补全处理的主要入口点
 * @example
 * ctx := NewCompletionContext(context.Background(), &CompletionPerformance{})
 * input := &CompletionInput{...}
 * response := handler.HandleCompletion(ctx, input)
 */
func (h *CompletionHandler) HandleCompletion(c *CompletionContext, input *CompletionInput) *CompletionResponse {
	rsp := input.Preprocess(c)
	if rsp != nil {
		return rsp
	}
	para := h.Adapt(input)
	rsp = h.CallLLM(c, para)
	if env.DebugMode {
		zap.L().Debug("completion input", zap.Any("input", input))
	}
	if rsp.Status != model.StatusSuccess {
		zap.L().Warn("completion failed",
			zap.String("status", string(rsp.Status)),
			zap.Any("request", para),
			zap.Any("response", rsp))
	} else {
		zap.L().Info("completion succeeded",
			zap.Any("request", para),
			zap.Any("response", rsp))
	}
	return rsp
}
