package parser

import (
	"strings"
)

type SimpleParser struct {
	language string
}

/**
 * 创建简化版本分析器
 * @param {string} language - 编程语言标识符，用于指定代码分析的语言类型
 * @returns {Parser} 返回Parser接口实现，可用于代码语法分析
 * @description
 * - 根据指定的编程语言创建简化版本的分析器实例
 * - 支持多种编程语言的基本语法检查
 * - 实现Parser接口，提供基础的代码分析功能
 * @example
 * parser := NewSimpleParser("python")
 * isValid := parser.IsCodeSyntax("print('Hello World')")
 */
func NewSimpleParser(language string) Parser {
	return &SimpleParser{
		language: language,
	}
}

/**
 * 检查代码语法（简化实现）
 * @param {string} code - 需要检查语法的代码字符串
 * @returns {boolean} 返回代码语法是否正确，true表示语法正确，false表示语法错误
 * @description
 * - 基于编程语言类型执行基本的语法检查
 * - 支持Python、JavaScript/TypeScript、Go语言的语法验证
 * - 对于不支持的语言默认返回true
 * - 实现Parser接口的IsCodeSyntax方法
 * @example
 * parser := NewSimpleParser("python")
 * isValid := parser.IsCodeSyntax("print('Hello World')")
 * // isValid = true
 */
func (t *SimpleParser) IsCodeSyntax(code string) bool {
	switch t.language {
	case "python":
		return t.checkPythonSyntax(code)
	case "javascript", "typescript":
		return t.checkJavaScriptSyntax(code)
	case "go":
		return t.checkGoSyntax(code)
	default:
		return true // 对于不支持的语言，默认返回true
	}
}

/**
 * 拦截语法错误代码（简化实现）
 * @param {string} choicesText - 候选文本内容，需要从中提取有效代码
 * @param {string} prefix - 代码前缀，用于语法检查的上下文
 * @param {string} suffix - 代码后缀，用于语法检查的上下文
 * @returns {string} 返回经过语法检查和修正的代码片段
 * @description
 * - 通过逐步截断候选文本来找到语法正确的代码片段
 * - 从后向前逐个字符删除，直到找到语法正确的代码
 * - 使用前缀和后缀进行完整的语法检查
 * - 如果无法找到有效代码，返回原始候选文本
 * @example
 * parser := NewSimpleParser("python")
 * result := parser.InterceptSyntaxErrorCode("print('Hello')", "def main():\n", "\nmain()")
 * // result = "print('Hello')"
 */
func (t *SimpleParser) InterceptSyntaxErrorCode(choicesText, prefix, suffix string) string {
	if choicesText == "" {
		return choicesText
	}

	cutCode := choicesText
	maxCutCount := t.GetLastKLineStrLen(cutCode, 1)

	for i := 0; i < maxCutCount; i++ {
		if t.IsCodeSyntax(prefix+cutCode+suffix) && strings.TrimSpace(cutCode) != "" {
			return strings.TrimRight(cutCode, "\n\r\t ")
		}

		if len(cutCode) > 0 {
			cutCode = cutCode[:len(cutCode)-1]
		} else {
			break
		}
	}

	return choicesText
}

/**
 * 提取代码块前后缀（简化实现）
 * @param {string} choicesText - 候选文本内容，用于提取前后缀
 * @param {string} prefix - 代码前缀，用于提取上下文
 * @param {string} suffix - 代码后缀，用于提取上下文
 * @returns {string, string} 返回提取的前缀和后缀字符串
 * @description
 * - 从候选文本中提取有效的代码块前后缀
 * - 使用特殊标记符来识别代码块边界
 * - 简化实现，基于行号提取代码块
 * - 如果提取失败，返回原始前缀和后缀
 * @example
 * parser := NewSimpleParser("python")
 * prefix, suffix := parser.ExtractBlockPrefixSuffix("code", "def main():", "\nreturn")
 */
func (t *SimpleParser) ExtractBlockPrefixSuffix(choicesText, prefix, suffix string) (string, string) {
	const specialMiddleSignal = "<special-middle>"
	code := prefix + specialMiddleSignal + choicesText + specialMiddleSignal + suffix

	startNumber, endNumber := getChoicesTextLineNumber(code, specialMiddleSignal)

	// 简化实现：基于行号提取代码块
	lines := strings.Split(code, "\n")
	if startNumber >= 0 && startNumber < len(lines) && endNumber >= 0 && endNumber < len(lines) {
		blockLines := lines[startNumber : endNumber+1]
		blockCode := strings.Join(blockLines, "\n")
		return isolatedPrefixSuffix(blockCode, specialMiddleSignal)
	}

	return prefix, suffix
}

/**
 * 提取准确的代码块前后缀（简化实现）
 * @param {string} prefix - 代码前缀，用于提取上下文
 * @param {string} suffix - 代码后缀，用于提取上下文
 * @returns {string, string} 返回提取的前缀和后缀字符串
 * @description
 * - 使用特殊标记符来准确定位代码块边界
 * - 基于行号提取代码块，获取前后缀内容
 * - 选取标记符周围的代码行作为提取范围
 * - 如果提取失败，返回原始前缀和后缀
 * @example
 * parser := NewSimpleParser("python")
 * prefix, suffix := parser.ExtractAccurateBlockPrefixSuffix("def main():", "\nreturn")
 */
func (t *SimpleParser) ExtractAccurateBlockPrefixSuffix(prefix, suffix string) (string, string) {
	const specialMiddleSignal = "<special-middle>"
	code := prefix + specialMiddleSignal + suffix
	lineNum, _ := getChoicesTextLineNumber(code, specialMiddleSignal)

	// 简化实现：基于行号提取代码块
	lines := strings.Split(code, "\n")
	if lineNum >= 0 && lineNum < len(lines) {
		// 提取当前行所在的代码块
		startLine := max(0, lineNum-2)
		endLine := min(len(lines), lineNum+3)

		blockLines := lines[startLine:endLine]
		blockCode := strings.Join(blockLines, "\n")
		return isolatedPrefixSuffix(blockCode, specialMiddleSignal)
	}

	return prefix, suffix
}

/**
 * 查找最近的代码块（简化实现）
 * @param {string} code - 完整的代码字符串，用于查找代码块
 * @param {int} startNumber - 起始行号，用于指定代码块的起始位置
 * @param {int} endNumber - 结束行号，用于指定代码块的结束位置
 * @returns {string} 返回找到的代码块字符串
 * @description
 * - 根据指定的行号范围提取代码块
 * - 简化实现，直接按行号截取代码
 * - 如果行号超出范围，返回完整代码
 * - 实现Parser接口的FindNearestBlock方法
 * @example
 * parser := NewSimpleParser("python")
 * block := parser.FindNearestBlock("line1\nline2\nline3", 0, 1)
 * // block = "line1\nline2"
 */
func (t *SimpleParser) FindNearestBlock(code string, startNumber, endNumber int) string {
	lines := strings.Split(code, "\n")
	if startNumber >= 0 && startNumber < len(lines) && endNumber >= 0 && endNumber < len(lines) {
		blockLines := lines[startNumber : endNumber+1]
		return strings.Join(blockLines, "\n")
	}
	return code
}

/**
 * 按行号查找第二层节点（简化实现）
 * @param {string} code - 完整的代码字符串，用于查找节点
 * @param {int} lineNum - 行号，用于指定要查找的节点位置
 * @returns {string} 返回找到的节点内容字符串
 * @description
 * - 根据指定的行号查找对应的代码节点
 * - 简化实现，直接返回指定行的内容
 * - 如果行号超出范围，返回空字符串
 * - 实现Parser接口的FindSecondLevelNodeByLineNum方法
 * @example
 * parser := NewSimpleParser("python")
 * node := parser.FindSecondLevelNodeByLineNum("line1\nline2\nline3", 1)
 * // node = "line2"
 */
func (t *SimpleParser) FindSecondLevelNodeByLineNum(code string, lineNum int) string {
	lines := strings.Split(code, "\n")
	if lineNum >= 0 && lineNum < len(lines) {
		return lines[lineNum]
	}
	return ""
}

/**
 * 查找指定行号的最近节点（简化实现）
 * @param {string} code - 完整的代码字符串，用于查找节点
 * @param {int} lineNum - 行号，用于指定要查找的节点位置
 * @returns {string, string} 返回前缀节点和后缀节点字符串
 * @description
 * - 根据指定的行号分割代码，获取前后节点
 * - 前缀节点包含指定行号之前的所有代码
 * - 后缀节点包含指定行号之后的所有代码
 * - 实现Parser接口的FindSecondLevelNearestNodeByLineNum方法
 * @example
 * parser := NewSimpleParser("python")
 * prefix, suffix := parser.FindSecondLevelNearestNodeByLineNum("line1\nline2\nline3", 1)
 * // prefix = "line1", suffix = "line3"
 */
func (t *SimpleParser) FindSecondLevelNearestNodeByLineNum(code string, lineNum int) (string, string) {
	lines := strings.Split(code, "\n")
	var prefixNode, suffixNode string

	if lineNum > 0 && lineNum < len(lines) {
		prefixNode = strings.Join(lines[:lineNum], "\n")
		suffixNode = strings.Join(lines[lineNum+1:], "\n")
	}

	return prefixNode, suffixNode
}

/**
 * 获取代码最后k行字符串长度（简化实现）
 * @param {string} code - 完整的代码字符串，用于计算长度
 * @param {int} k - 要计算的行数，从最后一行开始计算
 * @returns {int} 返回最后k行字符串的总长度，包括换行符
 * @description
 * - 从代码的最后一行开始，计算指定行数的字符长度
 * - 跳过空行，只计算非空行的长度
 * - 包含换行符的长度计算
 * - 如果代码行数不足k行，计算所有非空行的长度
 * @example
 * parser := NewSimpleParser("python")
 * length := parser.GetLastKLineStrLen("line1\nline2\nline3", 2)
 * // length = 10 (line2 + line3 + 换行符)
 */
func (t *SimpleParser) GetLastKLineStrLen(code string, k int) int {
	lines := strings.Split(code, "\n")
	var lastKLines []string

	for i := len(lines) - 1; i >= 0; i-- {
		if len(lines[i]) == 0 {
			continue
		}
		if len(lastKLines) == k {
			break
		}
		lastKLines = append(lastKLines, lines[i])
	}

	totalLen := 0
	for _, line := range lastKLines {
		totalLen += len(line)
	}

	// 加上换行符的长度
	newlineCount := max(k-1, len(lastKLines)-1)
	return totalLen + newlineCount
}

/**
 * 检查Python语法（简化实现）
 * @param {string} code - 需要检查语法的Python代码字符串
 * @returns {boolean} 返回Python代码语法是否正确，true表示语法正确，false表示语法错误
 * @description
 * - 检查Python代码的缩进是否符合语法规则
 * - 使用缩进栈来跟踪代码块的缩进级别
 * - 验证冒号后面的行是否正确缩进
 * - 跳过空行，只检查非空行的缩进
 * @example
 * parser := NewSimpleParser("python")
 * isValid := parser.checkPythonSyntax("if True:\n    print('Hello')")
 * // isValid = true
 */
func (t *SimpleParser) checkPythonSyntax(code string) bool {
	lines := strings.Split(code, "\n")
	indentStack := []int{0}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// 计算当前行的缩进级别
		currentIndent := 0
		for _, char := range line {
			if char == ' ' || char == '\t' {
				currentIndent++
			} else {
				break
			}
		}

		// 检查缩进是否合理
		if currentIndent > indentStack[len(indentStack)-1] && !strings.HasSuffix(trimmed, ":") {
			return false
		}

		// 更新缩进栈
		if strings.HasSuffix(trimmed, ":") {
			indentStack = append(indentStack, currentIndent+4)
		} else {
			for len(indentStack) > 1 && currentIndent <= indentStack[len(indentStack)-2] {
				indentStack = indentStack[:len(indentStack)-1]
			}
		}
	}

	return true
}

/**
 * 检查JavaScript语法（简化实现）
 * @param {string} code - 需要检查语法的JavaScript代码字符串
 * @returns {boolean} 返回JavaScript代码语法是否正确，true表示语法正确，false表示语法错误
 * @description
 * - 检查JavaScript代码的括号匹配情况
 * - 验证大括号{}、圆括号()、方括号[]是否成对匹配
 * - 如果任一类型的括号出现不匹配，返回false
 * - 简化实现，仅检查括号匹配，不检查其他语法规则
 * @example
 * parser := NewSimpleParser("javascript")
 * isValid := parser.checkJavaScriptSyntax("function test() { return; }")
 * // isValid = true
 */
func (t *SimpleParser) checkJavaScriptSyntax(code string) bool {
	bracketCount := 0
	parenCount := 0
	bracketSquareCount := 0

	for _, char := range code {
		switch char {
		case '{':
			bracketCount++
		case '}':
			bracketCount--
			if bracketCount < 0 {
				return false
			}
		case '(':
			parenCount++
		case ')':
			parenCount--
			if parenCount < 0 {
				return false
			}
		case '[':
			bracketSquareCount++
		case ']':
			bracketSquareCount--
			if bracketSquareCount < 0 {
				return false
			}
		}
	}

	return bracketCount == 0 && parenCount == 0 && bracketSquareCount == 0
}

/**
 * 检查Go语法（简化实现）
 * @param {string} code - 需要检查语法的Go代码字符串
 * @returns {boolean} 返回Go代码语法是否正确，true表示语法正确，false表示语法错误
 * @description
 * - 检查Go代码的括号匹配情况
 * - 验证大括号{}、圆括号()、方括号[]是否成对匹配
 * - 如果任一类型的括号出现不匹配，返回false
 * - 简化实现，仅检查括号匹配，不检查其他语法规则
 * @example
 * parser := NewSimpleParser("go")
 * isValid := parser.checkGoSyntax("func test() { return }")
 * // isValid = true
 */
func (t *SimpleParser) checkGoSyntax(code string) bool {
	bracketCount := 0
	parenCount := 0
	bracketSquareCount := 0

	for _, char := range code {
		switch char {
		case '{':
			bracketCount++
		case '}':
			bracketCount--
			if bracketCount < 0 {
				return false
			}
		case '(':
			parenCount++
		case ')':
			parenCount--
			if parenCount < 0 {
				return false
			}
		case '[':
			bracketSquareCount++
		case ']':
			bracketSquareCount--
			if bracketSquareCount < 0 {
				return false
			}
		}
	}

	return bracketCount == 0 && parenCount == 0 && bracketSquareCount == 0
}

/**
 * 获取补全内容在代码中的行号
 * @param {string} code - 完整的代码字符串，包含模式匹配的标记
 * @param {string} pattern - 要查找的模式字符串，用于定位补全内容
 * @returns {int, int} 返回起始行号和结束行号
 * @description
 * - 在代码中查找指定模式的第一次和第二次出现位置
 * - 第一次出现作为起始行号，第二次出现作为结束行号
 * - 逐行扫描代码，使用字符串包含检查来定位模式
 * - 如果模式只出现一次，结束行号与起始行号相同
 * @example
 * start, end := getChoicesTextLineNumber("line1\n<special-middle>\nline3\n<special-middle>", "<special-middle>")
 * // start = 1, end = 3
 */
func getChoicesTextLineNumber(code, pattern string) (int, int) {
	codeSplit := strings.Split(code, "\n")
	startNumber := 0
	endNumber := 0

	for i := range codeSplit {
		if strings.Contains(codeSplit[i], pattern) && startNumber == 0 {
			startNumber = i
			continue
		}

		if strings.Contains(codeSplit[i], pattern) && startNumber != 0 {
			endNumber = i
			break
		}
	}

	return startNumber, endNumber
}

/**
 * 分离前后缀
 * @param {string} code - 包含模式标记的代码字符串
 * @param {string} pattern - 用于分离前后缀的模式字符串
 * @returns {string, string} 返回分离后的前缀和后缀字符串
 * @description
 * - 使用模式字符串将代码分割成多个部分
 * - 第一部分作为前缀，最后一部分作为后缀
 * - 如果代码为空，返回空字符串
 * - 如果分割部分不足，返回空字符串
 * @example
 * prefix, suffix := isolatedPrefixSuffix("prefix<middle>content<middle>suffix", "<middle>")
 * // prefix = "prefix", suffix = "suffix"
 */
func isolatedPrefixSuffix(code, pattern string) (string, string) {
	if code == "" {
		return "", ""
	}

	split := strings.Split(code, pattern)
	if len(split) >= 2 {
		return split[0], split[len(split)-1]
	}

	return "", ""
}

/**
 * 返回两个整数中的较大值
 * @param {int} a - 第一个整数
 * @param {int} b - 第二个整数
 * @returns {int} 返回两个整数中的较大值
 * @description
 * - 比较两个整数的大小
 * - 返回较大的那个整数
 * - 如果两个整数相等，返回其中任意一个
 * - 辅助函数，用于代码中的数值比较
 * @example
 * result := max(5, 3)
 * // result = 5
 */
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

/**
 * 返回两个整数中的较小值
 * @param {int} a - 第一个整数
 * @param {int} b - 第二个整数
 * @returns {int} 返回两个整数中的较小值
 * @description
 * - 比较两个整数的大小
 * - 返回较小的那个整数
 * - 如果两个整数相等，返回其中任意一个
 * - 辅助函数，用于代码中的数值比较
 * @example
 * result := min(5, 3)
 * // result = 3
 */
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
