package completions

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"completion-agent/pkg/config"
	"completion-agent/pkg/logger"

	"go.uber.org/zap"
)

// 拒绝原因枚举
type RejectCode string

const (
	Accepted          RejectCode = "ACCEPTED"
	LowHiddenScore    RejectCode = "LOW_HIDDEN_SCORE"
	AuthFail          RejectCode = "AUTH_FAIL"
	FeatureNotSupport RejectCode = "FEATURE_NOT_SUPPORT"
)

// 补全过滤器接口
type Filter interface {
	Judge(in *CompletionInput) RejectCode
}

// 补全拒绝规则链
type FilterChain struct {
	filters []Filter
}

/**
 * Create new filter chain for completion request processing
 * @param {config.CompletionWrapperConfig} cfg - Configuration wrapper containing filter settings
 * @returns {FilterChain} Returns configured filter chain instance
 * @description
 * - Creates a chain of filters to evaluate completion requests
 * - Adds hidden score filter if not disabled in configuration
 * - Adds language feature filter if not disabled in configuration
 * - Filters are executed in the order they are added
 * @example
 * chain := NewFilterChain(config)
 * err := chain.Handle(request)
 * if err != nil {
 *     // Handle rejection
 * }
 */
func NewFilterChain(cfg *config.WrapperConfig) *FilterChain {
	handlers := make([]Filter, 0)

	if !cfg.Score.Disabled {
		handlers = append(handlers, NewScoreFilter(&cfg.Score))
	}

	if !cfg.Syntax.Disabled {
		handlers = append(handlers, NewSyntaxFilter(&cfg.Syntax))
	}

	return &FilterChain{
		filters: handlers,
	}
}

/**
 * Handle completion request through filter chain
 * @param {CompletionInput} in - Completion request data to be evaluated
 * @returns {error} Returns error if any filter rejects the request, nil if all filters accept
 * @description
 * - Processes completion request through all filters in the chain
 * - Stops processing and returns error on first filter rejection
 * - Request must pass all filters to be accepted
 * - Returns specific error message indicating which filter rejected the request
 * @example
 * err := chain.Handle(request)
 * if err != nil {
 *     log.Printf("Request rejected: %v", err)
 * }
 */
func (c *FilterChain) Handle(in *CompletionInput) error {
	for _, handler := range c.filters {
		if rejectCode := handler.Judge(in); rejectCode != Accepted {
			return fmt.Errorf("%s", rejectCode)
		}
	}
	return nil
}

//------------------------------------------------------------------------------
//	CodeFilters
//------------------------------------------------------------------------------

// 代码过滤器
type CodeFilters struct {
	StrPattern    string
	TreePattern   string
	FIMIndicator  string
	EndTag        string
	MinPromptLine int
}

/**
 * Create language feature filter for completion requests
 * @param {config.SyntaxFilterConfig} cfg - Configuration wrapper containing filter settings
 * @returns {CodeFilters} Returns configured language feature filter instance
 * @description
 * - Creates a language feature filter to determine if code completion should be triggered
 * - Sets up threshold score, string pattern, tree pattern, line count threshold and end tag
 * - Uses default values if not provided in configuration
 * @example
 * filter := NewSyntaxFilter(config)
 * rejectCode := filter.Judge(request)
 * if rejectCode == Accepted {
 *     // Process completion
 * }
 */
func NewSyntaxFilter(cfg *config.SyntaxFilterConfig) *CodeFilters {
	strPattern := cfg.StrPattern
	if strPattern == "" {
		strPattern = `import +.*|from +.*|from +.* import *.*`
	}

	treePattern := cfg.TreePattern
	if treePattern == "" {
		treePattern = `\(comment.*|\(string.*|\(set \(string.*|\(dictionary.*|\(integer.*|\(list.*|\(tuple.*`
	}
	minPromptLine := cfg.MinPromptLine
	if minPromptLine == 0 {
		minPromptLine = 5
	}

	endTag := cfg.EndTag
	if endTag == "" {
		endTag = "('>',';','}',')')"
	}

	return NewCodeFilters(minPromptLine, strPattern, treePattern, endTag)
}

/**
 * Create code filters for completion request evaluation
 * @param {int} MinPromptLine - Minimum line count threshold
 * @param {string} strPattern - String pattern for code analysis
 * @param {string} treePattern - Tree pattern for code analysis
 * @param {string} endTag - End tag pattern for cursor position detection
 * @returns {CodeFilters} Returns configured code filters instance
 * @description
 * - Creates code filters with specified configuration parameters
 * - Sets up patterns and thresholds for code completion evaluation
 * - Initializes FIM indicator for fill-in-middle completion detection
 * @example
 * filters := NewCodeFilters(0.3, 5, "import.*", ".*", "';','}'")
 * needCode := filters.NeedCode(request)
 */
func NewCodeFilters(minPromptLine int, strPattern, treePattern, endTag string) *CodeFilters {
	return &CodeFilters{
		StrPattern:    strPattern,
		TreePattern:   treePattern,
		FIMIndicator:  "<FILL_HERE>",
		EndTag:        endTag,
		MinPromptLine: minPromptLine,
	}
}

/**
 * Determine if code completion is needed for the request
 * @param {CompletionInput} in - Completion request data containing code context
 * @returns {bool} Returns true if code completion is needed, false otherwise
 * @description
 * - Checks if cursor is at the end of line (no completion needed)
 * - Checks if text after fill position starts with a word (no completion needed)
 * - Returns true if none of the rejection conditions are met
 * - Simplified implementation with basic filtering logic
 * @example
 * if filters.NeedCode(request) {
 *     // Process code completion
 * }
 */
func (c *CodeFilters) Judge(in *CompletionInput) RejectCode {
	// 跳过手动触发模式
	mode := strings.ToUpper(in.TriggerMode)
	if mode == "MANUAL" || mode == "CONTINUE" {
		return Accepted
	}
	if c.cursorIsAtTheEnd(in) {
		return FeatureNotSupport
	}

	if c.textAfterFillHereStartWithWord(in) {
		return FeatureNotSupport
	}
	// 简化实现，其他复杂的过滤逻辑暂时关闭
	// 可以根据需要逐步启用其他过滤条件

	return Accepted
}

func (c *CodeFilters) NeedCode(in *CompletionInput) bool {
	// 是否需要触发模型进行自动补全编码

	// 暂时关闭，参考相关文档
	// if c.tooFewLines(in) {
	//     return false
	// }

	if c.cursorIsAtTheEnd(in) {
		return false
	}

	if c.textAfterFillHereStartWithWord(in) {
		return false
	}

	// 简化实现，其他复杂的过滤逻辑暂时关闭
	// 可以根据需要逐步启用其他过滤条件

	return true
}

/**
 * Split prompt into text before and after cursor position
 * @param {string} prompt - Complete prompt text containing FIM indicator
 * @returns {string, string} Returns text before cursor and text after cursor
 * @description
 * - Splits prompt text at FIM indicator position
 * - Returns empty strings if FIM indicator is not found
 * - Handles cases where FIM indicator appears multiple times
 * @example
 * before, after := filters.splitPrompt("code before <FILL_HERE> code after")
 */
func (c *CodeFilters) splitPrompt(prompt string) (string, string) {
	textBeforeCursor, textAfterCursor := "", ""
	if strings.Contains(prompt, c.FIMIndicator) {
		parts := strings.Split(prompt, c.FIMIndicator)
		if len(parts) >= 2 {
			textBeforeCursor = parts[len(parts)-2]
			textAfterCursor = parts[len(parts)-1]
		} else {
			textBeforeCursor = prompt
			textAfterCursor = ""
		}
	}
	return textBeforeCursor, textAfterCursor
}

/**
 * Check if cursor is at the end of a line
 * @param {CompletionInput} in - Completion request data containing prompt
 * @returns {bool} Returns true if cursor is at line end, false otherwise
 * @description
 * - Splits prompt into text before and after cursor
 * - Parses end tags from configuration
 * - Checks if text before cursor ends with any configured end tag
 * - Verifies that text after cursor starts with empty line
 * - Returns true if all conditions indicate cursor is at line end
 * @example
 * if filters.cursorIsAtTheEnd(request) {
 *     // Skip completion
 * }
 */
func (c *CodeFilters) cursorIsAtTheEnd(in *CompletionInput) bool {
	// 光标位于有效行行尾的直接不触发补全
	// 行尾定义：光标左侧是'>'、';'、'}'、')'，右侧是换行符号

	textBeforeCursor, textAfterCursor := c.splitPrompt(in.Prompts.Prefix)
	if textBeforeCursor != "" && textAfterCursor != "" {
		// 解析endTag
		endTags := c.parseEndTag()
		for _, tag := range endTags {
			if strings.HasSuffix(strings.ReplaceAll(textBeforeCursor, " ", ""), tag) {
				// 检查右侧是否是空行
				lines := strings.Split(textAfterCursor, "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
					// fmt.Printf("光标位于行尾，跳过自动补全\n")
					return true
				}
			}
		}
	}
	return false
}

/**
 * Parse end tag configuration string into individual tags
 * @returns {[]string} Returns slice of parsed end tags
 * @description
 * - Parses end tag configuration string in format "('>',';','}',')')"
 * - Removes parentheses and quotes from configuration
 * - Splits string by comma separator
 * - Returns slice of cleaned end tags
 * @example
 * tags := filters.parseEndTag()
 * // tags will be ["\u003e", ";", "}", ")"]
 */
func (c *CodeFilters) parseEndTag() []string {
	// 解析endTag配置，格式如 "('>',';','}',')')"
	endTag := strings.TrimSpace(c.EndTag)
	endTag = strings.TrimPrefix(endTag, "(")
	endTag = strings.TrimSuffix(endTag, ")")
	endTag = strings.TrimPrefix(endTag, "'")
	endTag = strings.TrimSuffix(endTag, "'")

	tags := strings.Split(endTag, "','")
	result := make([]string, 0)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

/**
 * Check if text after fill position starts with a word character
 * @param {CompletionInput} in - Completion request data containing prompt
 * @returns {bool} Returns true if text after fill starts with word character, false otherwise
 * @description
 * - Extracts text after cursor position from prompt
 * - Checks if first character is letter (a-z, A-Z) or digit (0-9)
 * - Returns true if text after cursor starts with word character
 * - Used to skip completion when modifying variable names
 * @example
 * if filters.textAfterFillHereStartWithWord(request) {
 *     // Skip completion (likely variable name modification)
 * }
 */
func (c *CodeFilters) textAfterFillHereStartWithWord(in *CompletionInput) bool {
	// 补全后面直接是英文字母开头或数字的不补全，比如修改变量名称的场景
	_, textAfterCursor := c.splitPrompt(in.Prompts.Prefix)
	if textAfterCursor != "" {
		firstChar := textAfterCursor[0]
		if (firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= '0' && firstChar <= '9') {
			// fmt.Printf("光标后面是字符`%c`，跳过自动补全\n", firstChar)
			return true
		}
	}
	return false
}

/**
 * Check if prompt contains too few lines for completion
 * @param {CompletionInput} in - Completion request data containing prompt
 * @returns {bool} Returns true if prompt has too few lines, false otherwise
 * @description
 * - Splits prompt into individual lines
 * - Filters out empty lines from line count
 * - Compares non-empty line count with configured threshold
 * - Returns true if line count is below threshold
 * - Currently disabled implementation (always returns false)
 * @example
 * if filters.tooFewLines(request) {
 *     // Skip completion (insufficient context)
 * }
 */
func (c *CodeFilters) tooFewLines(in *CompletionInput) bool {
	// prompt行数太少不触发补全，排除空行场景
	lines := strings.Split(in.Prompts.Prefix, "\n")
	nonEmptyLines := make([]string, 0)
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	lineCount := len(nonEmptyLines)
	if lineCount < c.MinPromptLine {
		// fmt.Printf("prompt行数%d小于阈值%d，跳过自动补全\n", lineCount, c.MinPromptLine)
		return true
	}
	return false
}

//------------------------------------------------------------------------------
//	HiddenScoreFilter
//------------------------------------------------------------------------------

// 低隐藏分数过滤器
type HiddenScoreFilter struct {
	ThresholdScore                  float64
	ContextualFilterLanguageMap     map[string]int
	ContextualFilterWeights         []float64
	ContextualFilterAcceptThreshold float64
	ContextualFilterIntercept       float64
	ContextualFilterCharacterMap    map[string]int
}

/**
 * Create hidden score filter for completion requests
 * @param {config.CompletionWrapperConfig} cfg - Configuration wrapper containing filter settings
 * @returns {HiddenScoreFilter} Returns configured hidden score filter instance
 * @description
 * - Creates a hidden score filter to evaluate completion request quality
 * - Sets up threshold score for filtering low-quality completions
 * - Initializes hide score configuration with default threshold if not provided
 * @example
 * filter := NewScoreFilter(config)
 * rejectCode := filter.Judge(request)
 * if rejectCode == Accepted {
 *     // Process completion
 * }
 */
func NewScoreFilter(cfg *config.ScoreFilterConfig) *HiddenScoreFilter {
	thresholdScore := cfg.Threshold
	if thresholdScore == 0 {
		thresholdScore = 0.3
	}
	fpath := filepath.Join(config.CostrictDir, "config", "hidden-scores.json")
	return NewHiddenScoreFilter(fpath, thresholdScore)
}

/**
 * Judge if completion request should be accepted based on hidden score
 * @param {CompletionInput} in - Completion request data with score calculation info
 * @returns {RejectCode} Returns AcceptCode if score is above threshold, LowHiddenScore otherwise
 * @description
 * - Skips filtering for manual and continue trigger modes (always accepts)
 * - Calculates hidden score using configured algorithm
 * - Updates request data with calculated score
 * - Rejects completions with scores below threshold
 * - Logs debug information for rejected completions
 * @example
 * rejectCode := filter.Judge(request)
 * if rejectCode == LowHiddenScore {
 *     log.Printf("Completion rejected due to low score")
 * }
 */
func (h *HiddenScoreFilter) Judge(in *CompletionInput) RejectCode {
	// 跳过手动触发和继续补全模式
	mode := strings.ToUpper(in.TriggerMode)
	if mode == "MANUAL" || mode == "CONTINUE" {
		return Accepted
	}

	// 计算隐藏分数
	if in.HideScores == nil {
		return Accepted
	}

	score := 0.0
	if in.HideScores.DocumentLength != 0 {
		score = h.CalculateHideScore(in.HideScores, in.Prompts.Prefix, in.LanguageID)
	}

	// 将分数更新到请求数据中（问题4修复）
	if in.Extra == nil {
		in.Extra = make(map[string]interface{})
	}
	in.Extra["score"] = score

	// 通过配置阈值来过滤隐藏分低的补全
	if score < h.ThresholdScore {
		// 添加日志记录（问题1修复）
		logger.Debug("低隐藏分数拒绝补全",
			zap.Float64("score", score),
			zap.Float64("threshold", h.ThresholdScore),
			zap.String("completion_id", in.CompletionID),
			zap.String("language", in.LanguageID))
		return LowHiddenScore
	}

	return Accepted
}

func loadHiddenScoreFilter(configPath string) *HiddenScoreFilter {
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}
	var c HiddenScoreFilter
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil
	}
	return &c
}

/**
 * Create hide score configuration for completion filtering
 * @param {string} configPath - Path to configuration file (currently unused)
 * @param {float64} thresholdScore - Threshold score for filtering completions
 * @returns {HideScoreConfig} Returns configured hide score configuration
 * @description
 * - Creates hide score configuration with default language and character mappings
 * - Sets up weights for contextual filtering algorithm
 * - Configures acceptance threshold and intercept values
 * - Uses default threshold of 0.3 if not provided
 * @example
 * config := NewHiddenScoreFilter("config.yml", 0.3)
 * score := config.CalculateHideScore(request, prefix, "python")
 */
func NewHiddenScoreFilter(configPath string, thresholdScore float64) *HiddenScoreFilter {
	if thresholdScore == 0.0 {
		thresholdScore = 0.3
	}
	if filter := loadHiddenScoreFilter(configPath); filter != nil {
		filter.ThresholdScore = thresholdScore
		return filter
	}
	// 默认配置，模拟YAML文件中的配置
	filter := &HiddenScoreFilter{
		ContextualFilterLanguageMap: map[string]int{
			"python": 0, "javascript": 1, "typescript": 2, "java": 3, "go": 4,
			"c": 5, "cpp": 6, "csharp": 7, "php": 8, "ruby": 9,
			"rust": 10, "kotlin": 11, "scala": 12, "swift": 13, "objective-c": 14,
		},
		ContextualFilterWeights: []float64{
			0.99,   // 上一个标签的权重
			0.7,    // 当前行光标后为空的权重
			-0.17,  // 时间间隔的权重
			-0.22,  // 前缀尾行长度的权重
			0.13,   // 后缀长度的权重
			-0.007, // 文档长度的权重
			0.005,  // 光标所在文档位置的权重
			0.41,   // 光标位置与文档长度比值的权重
			// 语言权重，从第8个位置开始，每种语言一个权重
			-0.1, -0.08, -0.06, -0.04, -0.02, 0.0, 0.02, 0.04, 0.06, 0.08, 0.1, 0.12, 0.14, 0.16, 0.18,
			// 字符权重，从第29个位置开始
		},
		ContextualFilterAcceptThreshold: 0.5,
		ContextualFilterIntercept:       -0.3,
		ContextualFilterCharacterMap: map[string]int{
			" ": 0, "\t": 1, "\n": 2, "(": 3, ")": 4, "[": 5, "]": 6, "{": 7, "}": 8,
			",": 9, ";": 10, ":": 11, ".": 12, "=": 13, "+": 14, "-": 15, "*": 16, "/": 17,
			"\\": 18, "\"": 19, "'": 20, "<": 21, ">": 22, "?": 23, "!": 24, "@": 25, "#": 26,
			"$": 27, "%": 28, "^": 29, "&": 30, "|": 31, "~": 32, "`": 33,
		},
	}
	filter.ThresholdScore = thresholdScore
	return filter
}

/**
 * Calculate hide score for completion request
 * @param {HiddenScoreOptions} scores - Score calculation parameters
 * @param {string} language - Programming language identifier
 * @returns {float64} Returns calculated hide score between 0 and 1
 * @description
 * - Calculates probability of completion acceptance based on contextual features
 * - Considers previous label, whitespace after cursor, time since last completion
 * - Analyzes prefix and suffix lengths, document length, and cursor position
 * - Applies language-specific weights and character-specific weights
 * - Uses logistic function to convert weighted sum to probability
 * @example
 * score := filter.CalculateHideScore(request, prefix, "python")
 * if score < 0.3 {
 *     // Reject completion
 * }
 */
func (h *HiddenScoreFilter) CalculateHideScore(scores *HiddenScoreOptions, prefix, language string) float64 {
	// 判断光标权重
	whitespaceAfterCursor := 0.0
	if scores.IsWhitespaceAfterCursor {
		whitespaceAfterCursor = 1.0
	}

	// 触发时间间隔
	timeSincePreviousLabel := float64(time.Now().Unix()*1000-scores.PreviousLabelTimestamp) / 1000.0

	// 3.6最小值参考copilot的设置
	timeSincePreviousLabelLog := math.Log(1.0 + math.Max(3.6, timeSincePreviousLabel))

	prefixLengthLog := 0.0
	prefixLastCharWeight := 0
	prefixStr := prefix

	if prefixStr != "" {
		prefixLengthLog = math.Log(1.0 + float64(h.getLastLineLength(prefixStr)))
		prefixLastChar := prefixStr[len(prefixStr)-1:]
		if weight, exists := h.ContextualFilterCharacterMap[prefixLastChar]; exists {
			prefixLastCharWeight = weight
		}
	}

	suffixLengthLog := 0.0
	suffixLastCharWeight := 0

	// 参考const g = h.trimEnd(); 应该把换行符号也删掉
	trimmedSuffixStr := strings.TrimRight(prefixStr, " \t\n\r")
	if trimmedSuffixStr != "" {
		suffixLengthLog = math.Log(1.0 + float64(h.getLastLineLength(trimmedSuffixStr)))
		suffixLastChar := trimmedSuffixStr[len(trimmedSuffixStr)-1:]
		if weight, exists := h.ContextualFilterCharacterMap[suffixLastChar]; exists {
			suffixLastCharWeight = weight
		}
	}

	documentLengthLog := math.Log(1.0 + math.Max(float64(scores.DocumentLength), 0.0))
	promptEndPosLog := math.Log(1.0 + math.Max(float64(scores.PromptEndPos), 0.0))
	promptEndPosRatio := (float64(scores.PromptEndPos) + 0.5) / (1.0 + float64(scores.DocumentLength))

	// 若不支持该语言，默认走python
	languageWeight := 4 // python的默认值
	if weight, exists := h.ContextualFilterLanguageMap[language]; exists {
		languageWeight = weight
	}

	// 初始值-0.3
	score := h.ContextualFilterIntercept

	// 上一个标签的权重(上一次接受的话，下一次基本都会给予补全) +0.99
	if len(h.ContextualFilterWeights) > 0 {
		score += h.ContextualFilterWeights[0] * float64(scores.PreviousLabel)
	}

	// 当前行光标后为空的话倾向补全 + 0.7
	if len(h.ContextualFilterWeights) > 1 {
		score += h.ContextualFilterWeights[1] * whitespaceAfterCursor
	}

	// 时间间隔的权重，上一次触发的时间越久越不补全 - 0.17
	if len(h.ContextualFilterWeights) > 2 {
		score += h.ContextualFilterWeights[2] * timeSincePreviousLabelLog
	}

	// 前缀尾行长度的权重，尾行越长越不补全 - 0.22
	if len(h.ContextualFilterWeights) > 3 {
		score += h.ContextualFilterWeights[3] * prefixLengthLog
	}

	// 前缀去除空行或者空格后尾行长度的权重（去除空格或空行后的尾行），后缀越长越补全 + 0.13
	if len(h.ContextualFilterWeights) > 4 {
		score += h.ContextualFilterWeights[4] * suffixLengthLog
	}

	// 文档长度的权重，越长越不补 - 0.007
	if len(h.ContextualFilterWeights) > 5 {
		score += h.ContextualFilterWeights[5] * documentLengthLog
	}

	// 光标所在文档位置的权重，越靠后越补 + 0.005
	if len(h.ContextualFilterWeights) > 6 {
		score += h.ContextualFilterWeights[6] * promptEndPosLog
	}

	// 光标位置与文档长度的比值的权重，越靠后越补 + 0.41
	if len(h.ContextualFilterWeights) > 7 {
		score += h.ContextualFilterWeights[7] * promptEndPosRatio
	}

	// 语言权重
	languageWeightIndex := 8 + languageWeight
	if len(h.ContextualFilterWeights) > int(languageWeightIndex) {
		score += h.ContextualFilterWeights[languageWeightIndex]
	}

	// 前缀的最后一个字符的权重
	prefixCharWeightIndex := 29 + prefixLastCharWeight
	if len(h.ContextualFilterWeights) > int(prefixCharWeightIndex) {
		score += h.ContextualFilterWeights[prefixCharWeightIndex]
	}

	// 前缀最后一个有效行的最后一个字符的权重
	suffixCharWeightIndex := 125 + suffixLastCharWeight
	if len(h.ContextualFilterWeights) > int(suffixCharWeightIndex) {
		score += h.ContextualFilterWeights[suffixCharWeightIndex]
	}

	probabilityAccept := 1.0 / (1.0 + math.Exp(-score))
	return probabilityAccept
}

/**
 * Get length of the last line in text
 * @param {string} text - Input text to analyze
 * @returns {int} Returns length of last line, 0 if text is empty
 * @description
 * - Splits text into individual lines
 * - Returns length of the last line in the text
 * - Handles empty text by returning 0
 * - Used for calculating prefix and suffix lengths in hide score calculation
 * @example
 * length := filter.getLastLineLength("line1\nline2\nlast")
 * // length will be 4
 */
func (h *HiddenScoreFilter) getLastLineLength(text string) int {
	if text == "" {
		return 0
	}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 0
	}
	return len(lines[len(lines)-1])
}
