package completions

import (
	"code-completion/pkg/config"

	"go.uber.org/zap"
)

/**
 * 修剪补全结果
 * @param {string} completionText - 原始补全文本内容
 * @param {string} prefix - 代码前缀文本
 * @param {string} suffix - 代码后缀文本
 * @param {string} lang - 编程语言标识符
 * @returns {string} 返回修剪后的补全文本
 * @description
 * - 使用后置处理器链修剪补全结果
 * - 如果配置了自定义修剪器，使用自定义链
 * - 否则使用默认的后置处理器链
 * - 记录修剪过程的调试信息
 * - 用于优化补全结果的质量和格式
 * @example
 * result := handler.pruneCompletionCode(
 *     "function test() {\n    return;\n}\nfunction test2() {}",
 *     "function test() {",
 *     "}",
 *     "javascript"
 * )
 * // 结果可能移除重复的函数定义
 */
func (h *CompletionHandler) pruneCompletionCode(completionText, prefix, suffix, lang string) string {
	prunerContext := &PrunerContext{
		Language:       lang,
		CompletionCode: completionText,
		Prefix:         prefix,
		Suffix:         suffix,
	}
	var chain *PrunerChain
	var err error
	if len(config.Wrapper.Prune.Pruners) > 0 {
		chain, err = NewPrunerChainByNames(config.Wrapper.Prune.Pruners)
		if err != nil {
			zap.L().Error("Invalid config: 'wrapper.prune.pruners' contains invalid pruner names",
				zap.Any("pruners", config.Wrapper.Prune.Pruners))
		}
	}
	if chain == nil {
		chain = NewDefaultPrunerChain()
	}
	if chain.Process(prunerContext) {
		zap.L().Info("Prune by Pruners",
			zap.String("pre", completionText),
			zap.String("post", prunerContext.CompletionCode),
			zap.Any("hits", chain.GetHitProcessors()))
	}
	return prunerContext.CompletionCode
}
