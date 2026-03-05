package completions

import (
	"code-completion/pkg/config"
	"strings"
)

/**
 * 截断超长的提示词(前缀，后缀，上下文)
 * @param {*config.ModelConfig} cfg - 模型配置，包含最大前缀和后缀token限制
 * @param {*PromptOptions} ppt - 提示词选项，包含前缀、后缀和代码上下文
 * @description
 * - 检查并截断超过模型限制的长提示词
 * - 优先保留最靠近补全位置的代码
 * - 如果前缀已超长，完全丢弃上下文
 * - 否则截断上下文以保留前缀
 * - 同时处理后缀的截断
 * @example
 * cfg := &config.ModelConfig{MaxPrefix: 1000, MaxSuffix: 500}
 * ppt := &PromptOptions{
 *     Prefix: "long prefix...",
 *     Suffix: "long suffix...",
 *     CodeContext: "long context...",
 * }
 * handler.truncatePrompt(cfg, ppt)
 * // ppt中的内容会被截断到模型限制范围内
 */
func (h *CompletionHandler) truncatePrompt(cfg *config.ModelConfig, ppt *PromptOptions) {
	tokenizer := h.llm.Tokenizer()
	if tokenizer == nil {
		return
	}

	prefixTokens := tokenizer.Encode(ppt.Prefix)
	prefixTokensNum := len(prefixTokens)

	suffixTokens := tokenizer.Encode(ppt.Suffix)
	suffixTokensNum := len(suffixTokens)

	contextTokens := tokenizer.Encode(ppt.CodeContext)
	contextTokensNum := len(contextTokens)

	// 获取最大模型长度限制
	prefixMax := h.llm.Config().MaxPrefix
	suffixMax := h.llm.Config().MaxSuffix

	// 如果总token数超过限制，需要截断
	if prefixTokensNum+contextTokensNum > prefixMax {
		needCutTokens := prefixTokensNum + contextTokensNum - prefixMax

		// 前缀都已经超长了，就把上下文完全丢弃掉
		if prefixTokensNum >= prefixMax {
			prefixTokens = prefixTokens[prefixTokensNum-prefixMax:]
			ppt.CodeContext = ""
			ppt.Prefix = tokenizer.Decode(prefixTokens)
			ppt.Prefix = h.trimFirstLine(ppt.Prefix)
		} else {
			contextTokens = contextTokens[needCutTokens:]
			ppt.CodeContext = tokenizer.Decode(contextTokens)
		}
	}
	if suffixTokensNum > suffixMax {
		suffixTokens = suffixTokens[:suffixMax]
		ppt.Suffix = tokenizer.Decode(suffixTokens)
		ppt.Suffix = h.trimLastLine(ppt.Suffix)
	}
}

/**
 * 修剪提示词的第一行
 * @param {string} prompt - 要修剪的提示词文本
 * @returns {string} 返回修剪后的提示词文本
 * @description
 * - 从提示词中移除第一行（如果不是以换行符开头）
 * - 使用SplitAfter方法分割文本
 * - 保留除第一行外的所有内容
 * - 用于处理提示词格式，确保正确的代码缩进
 * @example
 * result := handler.trimFirstLine("line1\nline2\nline3")
 * // result = "line2\nline3"
 *
 * result = handler.trimFirstLine("\nline1\nline2")
 * // result = "\nline1\nline2" (第一行以换行符开头，保留)
 */
func (h *CompletionHandler) trimFirstLine(prompt string) string {
	lines := strings.SplitAfter(prompt, "\n")
	if len(lines) > 0 {
		if !strings.HasPrefix(lines[0], "\n") && !strings.HasPrefix(lines[0], "\r\n") {
			lines = lines[1:]
		}
		return strings.Join(lines, "")
	}
	return prompt
}

/**
 * 修剪后缀的最后一行
 * @param {string} suffix - 要修剪的后缀文本
 * @returns {string} 返回修剪后的后缀文本
 * @description
 * - 从后缀中移除最后一行（如果不是以换行符结尾）
 * - 使用SplitAfter方法分割文本
 * - 保留除最后一行外的所有内容
 * - 用于处理后缀格式，确保正确的代码结构
 * @example
 * result := handler.trimLastLine("line1\nline2\nline3")
 * // result = "line1\nline2"
 *
 * result = handler.trimLastLine("line1\nline2\n")
 * // result = "line1\nline2\n" (最后一行以换行符结尾，保留)
 */
func (h *CompletionHandler) trimLastLine(suffix string) string {
	lines := strings.SplitAfter(suffix, "\n")
	if len(lines) > 0 {
		if len(lines) > 1 && !strings.HasSuffix(lines[len(lines)-1], "\n") {
			lines = lines[:len(lines)-1]
		}
		return strings.Join(lines, "")
	}
	return suffix
}

/**
 * 获取提示词的token数量
 * @param {string} prompt - 要计算token数量的提示词文本
 * @returns {int} 返回token数量，如果tokenizer不可用返回0
 * @description
 * - 使用全局tokenizer计算文本的token数量
 * - 如果tokenizer未初始化，返回0
 * - 用于检查提示词长度是否超过模型限制
 * - 在truncatePrompt方法中调用
 * @example
 * count := handler.getTokensCount("function test() { return 'hello'; }")
 * // count = 10 (实际数量取决于tokenizer实现)
 */
func (h *CompletionHandler) getTokensCount(prompt string) int {
	return h.llm.Tokenizer().GetTokenCount(prompt)
}

/**
 * 准备停用词
 * @param {*CompletionInput} input - 补全输入对象，包含请求参数和停用词设置
 * @returns {[]string} 返回停用词列表
 * @description
 * - 合并请求中的停用词和系统默认停用词
 * - 添加默认的FIM停用词"<｜end▁of▁sentence｜>"
 * - 如果后缀为空或只包含空白字符，添加多行停用词
 * - 用于控制补全生成的停止条件
 * @example
 * input := &CompletionInput{
 *     Stop: []string{";", "}"},
 *     Processed: PromptOptions{Suffix: ""},
 * }
 * stopWords := handler.prepareStopWords(input)
 * // stopWords = [";", "}", "<｜end▁of▁sentence｜>", "\n\n", "\n\n\n"]
 */
func (h *CompletionHandler) prepareStopWords(input *CompletionInput) []string {
	var stopWords []string

	// 添加请求中的停用词
	if len(input.Stop) > 0 {
		stopWords = append(stopWords, input.Stop...)
	}

	// 添加默认的FIM停用词
	stopWords = append(stopWords, "<｜end▁of▁sentence｜>")

	// 如果后缀为空，添加系统停用词
	if input.Processed.Suffix == "" || strings.TrimSpace(input.Processed.Suffix) == "" {
		stopWords = append(stopWords, "\n\n", "\n\n\n")
	}

	return stopWords
}
