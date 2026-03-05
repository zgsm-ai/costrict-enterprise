package completions

import (
	"code-completion/pkg/parser"
	"fmt"
	"strings"
)

/**
 * 补全后置处理器类型
 * @description
 * - 定义后置处理器的类型枚举
 * - TypeDiscarder: 丢弃类型处理器，会完全丢弃补全内容
 * - TypeCutter: 裁剪类型处理器，会部分修改补全内容
 * - 用于区分不同类型的处理逻辑
 * @example
 * var processorType PrunerType = TypeDiscarder
 * if processorType == TypeDiscarder {
 *     // 丢弃处理逻辑
 * }
 */
type PrunerType string

const (
	TypeDiscarder PrunerType = "discarder"
	TypeCutter    PrunerType = "cutter"
)

const (
	DiscardExtremeRepetition string = "discard-extreme_repetition"
	DiscardNotMatchLanguage  string = "discard-not_match_language"
	DiscardInvalidBrackets   string = "discard-invalid_brackets"
	DiscardSyntaxError       string = "discard-syntax_error"
	DicardCssContent         string = "discard-css_content"
	CutSingleLine            string = "cut-single-line"
	CutRepetitiveText        string = "cut-repetitive_text"
	CutPrefixOverlap         string = "cut-prefix_overlap"
	CutSuffixOverlap         string = "cut-suffix_overlap"
	CutSyntaxError           string = "cut-syntax_error"
)

/**
 * 后置处理器定义映射
 * @description
 * - 定义处理器名称到处理器实例的映射关系
 * - 包含所有可用的后置处理器实现
 * - 支持动态查找和创建处理器
 * - 用于构建处理器链
 * @example
 * processor, exists := prunerDefs["discard-extreme_repetition"]
 * if exists {
 *     result := processor.Process(ctx)
 * }
 */
var prunerDefs map[string]Pruner = map[string]Pruner{
	DiscardExtremeRepetition: &ExtremeRepetitionDiscarder{},
	DiscardNotMatchLanguage:  &NotMatchLanguageDiscarder{},
	DiscardSyntaxError:       &SyntaxErrorDiscarder{},
	DiscardInvalidBrackets:   &InvalidBracketsDiscarder{},
	DicardCssContent:         &CssContentDiscarder{},
	CutSingleLine:            &SingleLineCutter{},
	CutRepetitiveText:        &RepetitiveTextCutter{},
	CutPrefixOverlap:         &PrefixOverlapCutter{},
	CutSuffixOverlap:         &SuffixOverlapCutter{},
	CutSyntaxError:           &SyntaxErrorCutter{},
}

/**
 * 补全后置处理器上下文
 * @description
 * - 封装后置处理器需要的上下文信息
 * - 包含语言类型、补全代码、前缀和后缀
 * - 用于在处理器链中传递数据和状态
 * - 处理器可以修改CompletionCode字段
 * @example
 * ctx := &PrunerContext{
 *     Language: "python",
 *     CompletionCode: "def test():\n    return",
 *     Prefix: "def ",
 *     Suffix: "\nprint('hello')",
 * }
 */
type PrunerContext struct {
	CompletionID   string `json:"completion_id"`
	Language       string `json:"language"`
	CompletionCode string `json:"completion_code"`
	Prefix         string `json:"prefix"`
	Suffix         string `json:"suffix"`
}

/**
 * 抽象补全后置处理器接口
 * @description
 * - 定义所有后置处理器必须实现的方法
 * - Process: 处理补全内容，返回是否进行了修改
 * - Name: 返回处理器名称，用于标识和调试
 * - Type: 返回处理器类型，决定处理行为
 * - 支持多种处理器的统一接口
 * @example
 * type MyPruner struct{}
 *
 * func (p *MyPruner) Process(ctx *PrunerContext) bool {
 *     // 处理逻辑
 *     return modified
 * }
 *
 * func (p *MyPruner) Name() string {
 *     return "my-processor"
 * }
 *
 * func (p *MyPruner) Type() PrunerType {
 *     return TypeCutter
 * }
 */
type Pruner interface {
	Process(ctx *PrunerContext) bool
	Name() string
	Type() PrunerType
}

//------------------------------------------------------------------------------
//	PrunerChain
//------------------------------------------------------------------------------

/**
 * 补全后置处理器链
 * @description
 * - 管理一组后置处理器的执行链
 * - 分别管理丢弃器和裁剪器
 * - 记录命中的处理器列表
 * - 按顺序执行处理器，支持提前终止
 * @example
 * chain := NewDefaultPrunerChain()
 * modified := chain.Process(ctx)
 * if modified {
 *     log.Println("补全内容被后置处理器修改")
 * }
 */
type PrunerChain struct {
	discarders    []Pruner
	cutters       []Pruner
	hitProcessors []string
}

/**
 * 创建新的后置处理器链
 * @param {[]Pruner} discarders - 丢弃类型处理器列表
 * @param {[]Pruner} cutters - 裁剪类型处理器列表
 * @returns {*PrunerChain} 返回初始化好的处理器链
 * @description
 * - 使用提供的丢弃器和裁剪器创建处理器链
 * - 初始化命中处理器列表为空
 * - 返回可执行的处理器链实例
 * - 用于自定义处理器组合
 * @example
 * discarders := []Pruner{&ExtremeRepetitionDiscarder{}}
 * cutters := []Pruner{&RepetitiveTextCutter{}}
 * chain := NewPrunerChain(discarders, cutters)
 */
func NewPrunerChain(discarders, cutters []Pruner) *PrunerChain {
	return &PrunerChain{
		discarders:    discarders,
		cutters:       cutters,
		hitProcessors: make([]string, 0),
	}
}

/**
 * 根据名称创建后置处理器链
 * @param {[]string} names - 处理器名称列表
 * @returns {*PrunerChain, error} 返回处理器链和错误信息
 * @description
 * - 根据处理器名称查找对应的处理器实例
 * - 将处理器按类型分组到丢弃器和裁剪器
 * - 如果遇到无效的处理器名称，返回错误
 * - 使用查找到的处理器创建处理器链
 * @throws
 * - 如果处理器名称无效，返回错误
 * @example
 * names := []string{"discard-extreme_repetition", "cut-repetitive_text"}
 * chain, err := NewPrunerChainByNames(names)
 * if err != nil {
 *     log.Fatal("创建处理器链失败:", err)
 * }
 */
func NewPrunerChainByNames(names []string) (*PrunerChain, error) {
	dicarders := make([]Pruner, 0)
	cutters := make([]Pruner, 0)
	for _, name := range names {
		p, exists := prunerDefs[name]
		if !exists {
			return nil, fmt.Errorf("Invalid Pruner: %s", name)
		}
		if p.Type() == TypeDiscarder {
			dicarders = append(dicarders, p)
		} else {
			cutters = append(cutters, p)
		}
	}
	return NewPrunerChain(dicarders, cutters), nil
}

/**
 * 创建默认的后置处理器链
 * @returns {*PrunerChain} 返回默认配置的处理器链
 * @description
 * - 创建包含标准处理器的默认链
 * - 丢弃器包含：极端重复、语言不匹配、语法错误
 * - 裁剪器包含：重复文本、前缀重叠、后缀重叠、语法错误
 * - 用于大多数常规补全场景
 * @example
 * chain := NewDefaultPrunerChain()
 * result := chain.Process(ctx)
 */
func NewDefaultPrunerChain() *PrunerChain {
	return NewPrunerChain(
		[]Pruner{
			&ExtremeRepetitionDiscarder{},
			&NotMatchLanguageDiscarder{},
			&SyntaxErrorDiscarder{},
		},
		[]Pruner{
			&RepetitiveTextCutter{},
			&PrefixOverlapCutter{},
			&SuffixOverlapCutter{},
			&SyntaxErrorCutter{},
		},
	)
}

/*
*
* 处理丢弃类型的处理器
* @param {*PrunerContext} ctx - 后置处理器上下文
* @returns {bool} 返回是否触发了丢弃处理
* @description
* - 按顺序执行所有丢弃类型处理器
* - 如果任何一个处理器返回true，立即停止处理
* - 记录命中的处理器名称
* - 返回是否触发了丢弃
* - 内部方法，由Process方法调用
* @example
// 通常不直接调用，由Process方法内部使用
*/
func (c *PrunerChain) processDiscard(ctx *PrunerContext) bool {
	for _, dicarder := range c.discarders {
		if dicarder.Process(ctx) {
			c.hitProcessors = append(c.hitProcessors, dicarder.Name())
			return true
		}
	}
	return false
}

/*
*
* 处理裁剪类型的处理器
* @param {*PrunerContext} ctx - 后置处理器上下文
* @returns {bool} 返回是否进行了裁剪修改
* @description
* - 按顺序执行所有裁剪类型处理器
* - 记录所有命中的处理器名称
* - 返回是否进行了任何裁剪修改
* - 即使一个处理器修改了内容，仍会继续执行其他处理器
* - 内部方法，由Process方法调用
* @example
// 通常不直接调用，由Process方法内部使用
*/
func (c *PrunerChain) processCut(ctx *PrunerContext) bool {
	result := false
	for _, cutter := range c.cutters {
		if cutter.Process(ctx) {
			c.hitProcessors = append(c.hitProcessors, cutter.Name())
			result = true
		}
	}
	return result
}

/**
 * 执行完整的后置处理流程
 * @param {*PrunerContext} ctx - 后置处理器上下文
 * @returns {bool} 返回是否对补全内容进行了修改
 * @description
 * - 首先执行丢弃类型处理器
 * - 如果触发丢弃，清空补全内容并返回true
 * - 否则执行裁剪类型处理器
 * - 最后去除补全内容末尾的空白字符
 * - 返回是否进行了任何修改
 * @example
 * chain := NewDefaultPrunerChain()
 * ctx := &PrunerContext{
 *     CompletionCode: "  function test() { return; }  ",
 *     Language: "javascript",
 * }
 * modified := chain.Process(ctx)
 * // ctx.CompletionCode = "function test() { return; }" (去除末尾空白)
 * // modified = true
 */
func (c *PrunerChain) Process(ctx *PrunerContext) bool {
	// 先处理内容丢弃情况，再处理内容裁剪情况
	if c.processDiscard(ctx) {
		ctx.CompletionCode = ""
		return true
	}

	result := c.processCut(ctx)

	// 后置验证：去除补全内容末尾的空格
	if ctx.CompletionCode != "" {
		ctx.CompletionCode = strings.TrimRight(ctx.CompletionCode, " \t\n\r")
	}

	return result
}

/**
 * 获取命中的处理器列表
 * @returns {[]string} 返回命中的处理器名称列表
 * @description
 * - 返回在Process执行过程中命中的处理器名称
 * - 包含所有返回true的处理器
 * - 用于调试和监控处理器的执行情况
 * - 每次Process调用后会重置
 * @example
 * chain.Process(ctx)
 * hitProcessors := chain.GetHitProcessors()
 * fmt.Println("命中的处理器:", hitProcessors)
 * // 输出: [discard-extreme_repetition cut-repetitive_text]
 */
func (c *PrunerChain) GetHitProcessors() []string {
	return c.hitProcessors
}

// ------------------------------------------------------------------------------
//
//	Pruners
//
// ------------------------------------------------------------------------------

/**
 * 丢弃器基类结构体
 * @description
 * - 实现Pruner接口的Type方法
 * - 返回TypeDiscarder类型
 * - 用作其他丢弃处理器的嵌入基类
 * - 提供默认的类型实现
 * @example
 * type MyDiscarder struct{ Discarder }
 *
 * func (p *MyDiscarder) Process(ctx *PrunerContext) bool {
 *     // 自定义丢弃逻辑
 *     return shouldDiscard
 * }
 *
 * func (p *MyDiscarder) Name() string {
 *     return "my-discarder"
 * }
 *
 * // Type方法由嵌入的Discarder提供
 */
type Discarder struct{}

func (p *Discarder) Type() PrunerType {
	return TypeDiscarder
}

/**
 * 裁剪器基类结构体
 * @description
 * - 实现Pruner接口的Type方法
 * - 返回TypeCutter类型
 * - 用作其他裁剪处理器的嵌入基类
 * - 提供默认的类型实现
 * @example
 * type MyCutter struct{ Cutter }
 *
 * func (p *MyCutter) Process(ctx *PrunerContext) bool {
 *     // 自定义裁剪逻辑
 *     return modified
 * }
 *
 * func (p *MyCutter) Name() string {
 *     return "my-cutter"
 * }
 *
 * // Type方法由嵌入的Cutter提供
 */
type Cutter struct{}

func (p *Cutter) Type() PrunerType {
	return TypeCutter
}

/**
 * 极端重复内容丢弃处理器
 * @description
 * - 检测并丢弃包含极端重复内容的补全
 * - 使用isExtremeRepetition函数检测重复模式
 * - 如果检测到极端重复，清空补全内容
 * - 继承自Discarder基类
 * @example
 * processor := &ExtremeRepetitionDiscarder{}
 * ctx := &PrunerContext{
 *     CompletionCode: "hello hello hello hello hello",
 * }
 * discarded := processor.Process(ctx)
 * // 如果检测到极端重复，ctx.CompletionCode = ""，discarded = true
 */
type ExtremeRepetitionDiscarder struct{ Discarder }

func (p *ExtremeRepetitionDiscarder) Process(ctx *PrunerContext) bool {
	// 极端重复内容丢弃
	flag, _, _ := isExtremeRepetition(ctx.CompletionCode)
	if !flag {
		return false
	}
	ctx.CompletionCode = ""
	return true
}

func (p *ExtremeRepetitionDiscarder) Name() string {
	return string(DiscardExtremeRepetition)
}

/**
 * 非匹配语言补全丢弃处理器
 * @description
 * - 检测并丢弃与目标语言不匹配的补全
 * - 特别检测非Python语言但生成Python代码的情况
 * - 使用IsPythonText函数判断是否为Python代码
 * - 如果语言不匹配，清空补全内容
 * - 继承自Discarder基类
 * @example
 * processor := &NotMatchLanguageDiscarder{}
 * ctx := &PrunerContext{
 *     Language: "javascript",
 *     CompletionCode: "def python_function():\n    pass",
 * }
 * discarded := processor.Process(ctx)
 * // 检测到JavaScript语言但生成Python代码，ctx.CompletionCode = ""，discarded = true
 */
type NotMatchLanguageDiscarder struct{ Discarder }

func (p *NotMatchLanguageDiscarder) Process(ctx *PrunerContext) bool {
	// 非python语言但是python代码，则丢弃补全内容
	if strings.ToLower(ctx.Language) != "python" && IsPythonText(ctx.CompletionCode) {
		ctx.CompletionCode = ""
		return true
	}
	return false
}

func (p *NotMatchLanguageDiscarder) Name() string {
	return string(DiscardNotMatchLanguage)
}

/**
 * 重复文本裁剪处理器
 * @description
 * - 检测并裁剪补全中的重复文本
 * - 使用cutRepetitiveText函数处理重复内容
 * - 如果检测到重复并进行裁剪，返回true
 * - 继承自Cutter基类
 * @example
 * processor := &RepetitiveTextCutter{}
 * ctx := &PrunerContext{
 *     CompletionCode: "function test() { return; return; return; }",
 * }
 * modified := processor.Process(ctx)
 * // 裁剪重复内容后，ctx.CompletionCode = "function test() { return; }"，modified = true
 */
type RepetitiveTextCutter struct{ Cutter }

func (p *RepetitiveTextCutter) Process(ctx *PrunerContext) bool {
	processedCode := cutRepetitiveText(ctx.CompletionCode)
	if processedCode != ctx.CompletionCode {
		ctx.CompletionCode = processedCode
		return true
	}
	return false
}

func (p *RepetitiveTextCutter) Name() string {
	return string(CutRepetitiveText)
}

/**
 * 前缀重叠裁剪处理器
 * @description
 * - 检测并裁剪与前缀重叠的补全内容
 * - 使用cutPrefixOverlap函数处理重叠部分
 * - 默认使用3行的cutLine参数
 * - 如果检测到重叠并进行裁剪，返回true
 * - 继承自Cutter基类
 * @example
 * processor := &PrefixOverlapCutter{}
 * ctx := &PrunerContext{
 *     Prefix: "function test() {",
 *     CompletionCode: "function test() { return; }",
 * }
 * modified := processor.Process(ctx)
 * // 裁剪重叠部分后，ctx.CompletionCode = "return; }"，modified = true
 */
type PrefixOverlapCutter struct{ Cutter }

func (p *PrefixOverlapCutter) Process(ctx *PrunerContext) bool {
	// 补全内容前缀重复处理
	// 使用默认的cutLine参数值3
	processedCode := cutPrefixOverlap(ctx.CompletionCode, ctx.Prefix, ctx.Suffix, 3)
	if processedCode != ctx.CompletionCode {
		ctx.CompletionCode = processedCode
		return true
	}
	return false
}

func (p *PrefixOverlapCutter) Name() string {
	return string(CutPrefixOverlap)
}

/**
 * 后缀重叠裁剪处理器
 * @description
 * - 检测并裁剪与后缀重叠的补全内容
 * - 使用cutSuffixOverlap函数处理重叠部分
 * - 默认使用3行的cutLine参数和8的ignoreOverlapLen参数
 * - 如果检测到重叠并进行裁剪，返回true
 * - 继承自Cutter基类
 * @example
 * processor := &SuffixOverlapCutter{}
 * ctx := &PrunerContext{
 *     Suffix: "\n    return; }",
 *     CompletionCode: "function test() {\n    return; }",
 * }
 * modified := processor.Process(ctx)
 * // 裁剪重叠部分后，ctx.CompletionCode = "function test() {"，modified = true
 */
type SuffixOverlapCutter struct{ Cutter }

func (p *SuffixOverlapCutter) Process(ctx *PrunerContext) bool {
	// 使用默认的cutLine参数值3和ignoreOverlapLen参数值8
	processedCode := cutSuffixOverlap(ctx.CompletionCode, ctx.Prefix, ctx.Suffix, 3, 8)
	if processedCode != ctx.CompletionCode {
		ctx.CompletionCode = processedCode
		return true
	}
	return false
}

func (p *SuffixOverlapCutter) Name() string {
	return string(CutSuffixOverlap)
}

// 以下是工具函数的占位符，后续需要从common.py移植实现

/**
 * 无效括号丢弃处理器
 * @description
 * - 检测并丢弃包含无效括号的补全
 * - 使用IsValidBrackets函数验证括号匹配
 * - 如果括号不匹配，清空补全内容
 * - 继承自Discarder基类
 * @example
 * processor := &InvalidBracketsDiscarder{}
 * ctx := &PrunerContext{
 *     CompletionCode: "function test() { return; ",
 * }
 * discarded := processor.Process(ctx)
 * // 检测到括号不匹配，ctx.CompletionCode = ""，discarded = true
 */
type InvalidBracketsDiscarder struct{ Discarder }

func (p *InvalidBracketsDiscarder) Process(ctx *PrunerContext) bool {
	if !IsValidBrackets(ctx.CompletionCode) {
		return true
	}
	return false
}

func (p *InvalidBracketsDiscarder) Name() string {
	return string(DiscardInvalidBrackets)
}

/**
 * CSS内容丢弃处理器
 * @description
 * - 检测并丢弃非CSS语言中的CSS内容
 * - 使用JudgeCss函数判断是否为CSS内容
 * - 如果非CSS语言但包含CSS内容，清空补全内容
 * - 使用0.7的置信度阈值
 * - 继承自Discarder基类
 * @example
 * processor := &CssContentDiscarder{}
 * ctx := &PrunerContext{
 *     Language: "javascript",
 *     CompletionCode: ".css-class { color: red; }",
 * }
 * discarded := processor.Process(ctx)
 * // 检测到JavaScript语言包含CSS内容，ctx.CompletionCode = ""，discarded = true
 */
type CssContentDiscarder struct{ Discarder }

func (p *CssContentDiscarder) Process(ctx *PrunerContext) bool {
	// 如果是非CSS语言但是包含CSS内容，则去除CSS内容
	if strings.ToLower(ctx.Language) != "css" && JudgeCss(ctx.Language, ctx.CompletionCode, 0.7) {
		ctx.CompletionCode = ""
		return true
	}
	return false
}

func (p *CssContentDiscarder) Name() string {
	return string(DicardCssContent)
}

/**
 * 语法错误丢弃处理器
 * @description
 * - 检测并丢弃包含语法错误的补全
 * - 使用isCodeSyntax函数验证语法正确性
 * - 考虑前缀和后缀的上下文进行语法检查
 * - 如果存在语法错误，清空补全内容
 * - 继承自Discarder基类
 * @example
 * processor := &SyntaxErrorDiscarder{}
 * ctx := &PrunerContext{
 *     Language: "python",
 *     CompletionCode: "def test(\n    return",
 *     Prefix: "def ",
 *     Suffix: "\nprint('hello')",
 * }
 * discarded := processor.Process(ctx)
 * // 检测到语法错误，ctx.CompletionCode = ""，discarded = true
 */
type SyntaxErrorDiscarder struct{ Discarder }

func (p *SyntaxErrorDiscarder) Process(ctx *PrunerContext) bool {
	if !isCodeSyntax(ctx.Language, ctx.CompletionCode, ctx.Prefix, ctx.Suffix) {
		ctx.CompletionCode = ""
		return true
	}
	return false
}

func (p *SyntaxErrorDiscarder) Name() string {
	return string(DiscardSyntaxError)
}

/**
 * 语法错误裁剪处理器
 * @description
 * - 检测并裁剪包含语法错误的补全内容
 * - 使用TreeSitter进行语法分析和错误拦截
 * - 通过InterceptSyntaxErrorCode方法裁剪错误部分
 * - 如果进行了裁剪，返回true
 * - 继承自Cutter基类
 * @example
 * processor := &SyntaxErrorCutter{}
 * ctx := &PrunerContext{
 *     Language: "python",
 *     CompletionCode: "def test():\n    return\n    invalid line\n    return",
 *     Prefix: "def ",
 *     Suffix: "\nprint('hello')",
 * }
 * modified := processor.Process(ctx)
 * // 裁剪语法错误部分后，ctx.CompletionCode可能被修改，modified = true
 */
type SyntaxErrorCutter struct{ Cutter }

func (p *SyntaxErrorCutter) Process(ctx *PrunerContext) bool {
	// 进行语法错误拦截和代码裁剪
	tsUtil := parser.NewSimpleParser(ctx.Language)
	if tsUtil == nil {
		return false
	}

	processedCode := tsUtil.InterceptSyntaxErrorCode(ctx.CompletionCode, ctx.Prefix, ctx.Suffix)
	if processedCode != ctx.CompletionCode {
		ctx.CompletionCode = processedCode
		return true
	}
	return false
}

func (p *SyntaxErrorCutter) Name() string {
	return string(CutSyntaxError)
}

type SingleLineCutter struct{ Cutter }

func (p *SingleLineCutter) Process(ctx *PrunerContext) bool {
	processedCode := pruneSingleLine(ctx.CompletionCode, ctx.Prefix, ctx.Suffix, ctx.Language)
	if processedCode != ctx.CompletionCode {
		ctx.CompletionCode = processedCode
		return true
	}
	return false
}

func (p *SingleLineCutter) Name() string {
	return string(CutSingleLine)
}

/**
 * 检查代码语法是否正确
 * @param {string} language - 编程语言标识符
 * @param {string} code - 要检查的代码内容
 * @param {string} prefix - 代码前缀，用于上下文
 * @param {string} suffix - 代码后缀，用于上下文
 * @returns {bool} 返回语法是否正确
 * @description
 * - 创建指定语言的简单语法分析器
 * - 提取准确的代码块前后缀
 * - 将前缀、代码和后缀组合进行语法检查
 * - 如果分析器创建失败，默认返回true
 * - 用于语法错误处理器的语法验证
 * @example
 * valid := isCodeSyntax("python", "    return", "def ", "\nprint('hello')")
 * // valid = true
 *
 * invalid := isCodeSyntax("python", "def test(\n    return", "def ", "\nprint('hello')")
 * // invalid = false (语法错误)
 */
func isCodeSyntax(language, code, prefix, suffix string) bool {
	tsUtil := parser.NewSimpleParser(language)
	if tsUtil == nil {
		return true
	}

	// 提取准确的代码块前后缀
	newPrefix, newSuffix := tsUtil.ExtractAccurateBlockPrefixSuffix(prefix, suffix)

	// 检查语法
	return tsUtil.IsCodeSyntax(newPrefix + code + newSuffix)
}

// 判断是否为单行补全
func needSingleLine(linePrefix, lineSuffix, language string) bool {
	// 简化的单行补全判断逻辑
	// 可以根据实际需求扩展，参考Python代码中的CompletionLineHandler逻辑

	// 如果光标前缀不为空且光标后缀为空，可能是单行补全
	if linePrefix != "" && lineSuffix == "" {
		return true
	}

	// 如果光标前缀以特定字符结尾，可能是单行补全
	if strings.HasSuffix(linePrefix, ".") ||
		strings.HasSuffix(linePrefix, " ") ||
		strings.HasSuffix(linePrefix, "\t") {
		return true
	}

	// 根据语言类型判断
	switch strings.ToLower(language) {
	case "python", "javascript", "typescript", "java", "c", "cpp":
		// 对于这些语言，如果光标在行首，可能是单行补全
		if strings.TrimSpace(linePrefix) == "" {
			return true
		}
	}

	return false
}

func pruneSingleLine(completionText, prefix, suffix, lang string) string {
	var linePrefix, lineSuffix string
	lines := strings.Split(prefix, "\n")
	if len(lines) > 0 {
		linePrefix = lines[len(lines)-1]
	}
	lines = strings.Split(suffix, "\n")
	if len(lines) > 0 {
		lineSuffix = lines[0]
		if len(lines) > 1 {
			lineSuffix += "\n"
		}
	}
	if needSingleLine(linePrefix, lineSuffix, lang) {
		lines := strings.Split(completionText, "\n")
		if len(lines) <= 1 {
			return completionText
		}
		if lines[0] == "" {
			return "\n" + lines[1]
		}
		return lines[0]
	}
	return completionText
}
