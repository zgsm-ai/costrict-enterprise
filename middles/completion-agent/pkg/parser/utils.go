package parser

import (
	"os"
	"regexp"
	"strings"
)

/**
 * Compute the longest prefix suffix match length for a string
 * @param {string} content - Input string to compute prefix suffix matches
 * @returns {[]int} Returns array of match lengths for each position
 * @description
 * - Implements KMP algorithm's prefix function to compute longest prefix suffix matches
 * - Returns empty array for empty input
 * - Each position i stores the length of longest proper prefix which is also suffix
 * @example
 * matches := computePrefixSuffixMatchLength("ababc")
 * // matches will be [-1, 0, 0, 1, 2]
 */
func computePrefixSuffixMatchLength(content string) []int {
	if len(content) == 0 {
		return []int{}
	}

	matchLengths := make([]int, len(content))
	matchLengths[0] = -1
	matchIndex := -1

	for i := 1; i < len(content); i++ {
		for matchIndex >= 0 && content[matchIndex+1] != content[i] {
			matchIndex = matchLengths[matchIndex]
		}
		if content[matchIndex+1] == content[i] {
			matchIndex++
		}
		matchLengths[i] = matchIndex
	}

	return matchLengths
}

/**
 * Reverse a string character by character
 * @param {string} s - Input string to reverse
 * @returns {string} Returns reversed string
 * @description
 * - Converts string to rune array for proper Unicode handling
 * - Swaps characters from both ends moving towards center
 * - Handles multi-byte Unicode characters correctly
 * @example
 * reversed := reverseString("hello")
 * // reversed will be "olleh"
 */
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

/**
 * Remove overlapping content between completion text and suffix prefix
 * @param {string} text - Completion text to be processed
 * @param {string} prefix - Prefix text before cursor position (not used in this function)
 * @param {string} suffix - Suffix text after cursor position
 * @param {int} cutLine - Maximum number of lines to check for overlap
 * @param {int} ignoreOverlapLen - Minimum overlap length to ignore (avoid false positives)
 * @returns {string} Returns text with overlapping content removed from end
 * @description
 * - Removes trailing whitespace from completion text
 * - Iteratively checks for overlap with suffix lines
 * - For each iteration, checks overlap with suffix first line
 * - Breaks early if suffix first line is shorter than overlap check
 * - Avoids cutting if overlap content is too short (ignoreOverlapLen)
 * - Processes multiple lines by removing first line from suffix and repeating
 * @example
 * processed := CutSuffixOverlap("completion", "prefix", "pletion suffix", 5, 3)
 * // processed will be "com"
 */
func CutSuffixOverlap(text, prefix, suffix string, cutLine int, ignoreOverlapLen int) string {
	if len(text) == 0 {
		return text
	}

	text = strings.TrimRight(text, " \t\n\r")
	textLen := len(text)
	suffix = strings.TrimSpace(suffix)

	// 循环多次，每次都截掉suffix的首行再进行内容重叠切割
	for i := 0; i < cutLine; i++ {
		suffixLines := strings.Split(suffix, "\n")
		suffixLen := len(suffix)

		if textLen == 0 || suffixLen == 0 {
			return text
		}

		firstLineSuffixLen := len(suffixLines[0])
		maxOverlapLength := min(textLen, suffixLen)

		for j := maxOverlapLength; j > maxOverlapLength/2; j-- {
			// 若suffix首行长度大于判重长度，则直接返回
			if j < firstLineSuffixLen {
				break
			}

			// 若suffix首行长度等于判重长度且首行仅有一个单词，那么无需判重直接返回
			if j == firstLineSuffixLen && len(strings.Split(suffix[:j], " ")) == 1 {
				break
			}

			// 一旦text和suffix存在重叠部分，立刻返回非重叠部分
			if text[textLen-j:] == suffix[:j] {
				// 避免误判切割到有效结束符，因此当需切割的内容长度小于等于某个值时，不进行重叠内容切割
				if len(strings.TrimSpace(suffix[:j])) > ignoreOverlapLen {
					return text[:textLen-j]
				}
			}
		}

		if len(suffixLines) > 1 {
			suffix = strings.Join(suffixLines[1:], "\n")
		} else {
			suffix = ""
		}
	}

	return text
}

/**
 * Remove overlapping content between completion text and prefix suffix
 * @param {string} text - The completion text to be processed
 * @param {string} prefix - The prefix text before cursor position
 * @param {string} suffix - The suffix text after cursor position (not used in this function)
 * @param {int} cutLine - Maximum number of lines to check for overlap
 * @returns {string} Processed text with overlapping content removed
 * @description
 * - This function identifies and removes overlapping content between the completion text and the prefix
 * - It uses a sliding window approach to compare lines from the end of prefix with lines from the beginning of completion text
 * - If significant overlap is detected (3+ consecutive lines or 60%+ match ratio), the entire completion is discarded
 * - For short completion texts (<3 lines), it uses a simpler check via judgePrefixFullLineRepetitive
 * @example
 * // If prefix ends with "function test() {" and completion starts with "function test() {",
 * // the completion will be considered overlapping and removed
 * processedText := CutPrefixOverlap(completion, prefix, suffix, 5)
 */
func CutPrefixOverlap(text, prefix, suffix string, cutLine int) string {
	// Remove leading/trailing whitespace from completion text
	stripText := strings.TrimSpace(text)
	if len(stripText) == 0 {
		return text
	}

	// Split completion text into lines
	splitText := strings.Split(stripText, "\n")
	// For short completion texts (<3 lines), use a simpler check
	if len(splitText) < 3 {
		// Check if completion text is completely repetitive with prefix's last line
		if judgePrefixFullLineRepetitive(text, prefix) {
			return ""
		}
		return text
	}

	// Clean and split prefix into lines
	prefix = strings.TrimSpace(prefix)
	splitPrefix := strings.Split(prefix, "\n")

	// Determine the maximum number of lines to compare
	matchLine := min(len(splitPrefix), len(splitText))
	// Take only the first 'matchLine' lines from completion text for comparison
	patternTextList := splitText[:matchLine]

	// Iterate through different offset positions to find overlap
	for i := 0; i < cutLine; i++ {
		// Ensure i doesn't exceed splitPrefix length to avoid negative endIdx
		if i >= len(splitPrefix) {
			break
		}

		// Calculate start and end indices for the sliding window in prefix
		// This creates a window that slides from the end of prefix backward
		startIdx := len(splitPrefix) - matchLine - i
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx := len(splitPrefix) - i

		// Ensure startIdx <= endIdx to avoid invalid slicing
		if startIdx > endIdx {
			continue
		}

		// Extract the current window from prefix for comparison
		curMatchTextList := splitPrefix[startIdx:endIdx]

		// Count matching lines between prefix window and completion text
		matchCount := 0
		continueFlag := true // Tracks if we're still in a consecutive match sequence

		// Compare each line in the windows
		for j := 0; j < len(curMatchTextList) && j < len(patternTextList); j++ {
			if strings.TrimSpace(curMatchTextList[j]) == strings.TrimSpace(patternTextList[j]) {
				matchCount++
				// If we find 3 consecutive matching lines, consider it significant overlap
				if matchCount == 3 && continueFlag {
					return ""
				}
				// If 60% or more lines match, also consider it significant overlap
				if float64(matchCount)/float64(matchLine) >= 0.6 {
					return ""
				}
			} else {
				// Reset consecutive match flag when a mismatch is found
				continueFlag = false
			}
		}
	}

	// If no significant overlap is found, return the original text
	return text
}

/**
 * Check if completion content completely repeats with prefix's last line
 * @param {string} completionText - Completion text to check for repetition
 * @param {string} prefix - Prefix text ending to compare with
 * @returns {bool} Returns true if completion repeats prefix ending, false otherwise
 * @description
 * - Returns false for empty completion or prefix
 * - Combines prefix last line with completion text for comparison
 * - Filters out empty lines from both texts
 * - Returns false if completion has more lines than prefix
 * - Compares completion lines with corresponding prefix ending lines
 * @example
 * if judgePrefixFullLineRepetitive("text", "prefix text") {
 *     // Completion completely repeats prefix ending
 * }
 */
func judgePrefixFullLineRepetitive(completionText, prefix string) bool {
	if len(prefix) == 0 || len(completionText) == 0 {
		return false
	}

	splitPrefixText := strings.Split(prefix, "\n")
	// 若将同行光标前的内容拼接到补全内容中，便于完全匹配
	linePrefixText := splitPrefixText[len(splitPrefixText)-1]
	completionText = linePrefixText + completionText

	splitCompletionText := strings.Split(completionText, "\n")
	var nonEmptyCompletionText []string
	for _, line := range splitCompletionText {
		if strings.TrimSpace(line) != "" {
			nonEmptyCompletionText = append(nonEmptyCompletionText, line)
		}
	}

	splitPrefixText = splitPrefixText[:len(splitPrefixText)-1]
	var nonEmptyPrefixText []string
	for _, line := range splitPrefixText {
		if strings.TrimSpace(line) != "" {
			nonEmptyPrefixText = append(nonEmptyPrefixText, line)
		}
	}

	// 若补全内容行数大于匹配内容行数，则直接返回False
	if len(nonEmptyCompletionText) > len(nonEmptyPrefixText) {
		return false
	}

	if len(nonEmptyPrefixText) >= len(nonEmptyCompletionText) {
		nonEmptyPrefixText = nonEmptyPrefixText[len(nonEmptyPrefixText)-len(nonEmptyCompletionText):]
	}

	for i := 0; i < len(nonEmptyCompletionText); i++ {
		if nonEmptyCompletionText[i] != nonEmptyPrefixText[i] {
			return false
		}
	}

	return true
}

/**
 * Remove repetitive content from completion text
 * @param {string} text - Completion text to process
 * @returns {string} Returns text with repetitive content removed
 * @description
 * - Returns original text if length is 0
 * - Only processes texts with 3 or more lines
 * - Uses internal ratio threshold of 0.15 for repetition detection
 * - Delegates to doCutRepetitiveText for actual processing
 * @example
 * processed := CutRepetitiveText("abc\nabc\nabc\ndef")
 * // processed will remove repetitive "abc" lines
 */
func CutRepetitiveText(text string) string {
	if len(text) == 0 {
		return text
	}

	// 行数超过3才触发去重
	lineCount := len(strings.Split(strings.TrimSpace(text), "\n"))
	if lineCount < 3 {
		return text
	}

	return doCutRepetitiveText(text, 0.15)
}

/**
 * Remove repetitive content if prefix-suffix match ratio exceeds threshold
 * @param {string} text - Text to process for repetitive content
 * @param {float64} ratio - Threshold ratio for detecting repetition
 * @returns {string} Returns text with repetitive content removed
 * @description
 * - Returns original text if empty or only whitespace
 * - Counts trailing newlines to preserve them after processing
 * - Reverses text and computes prefix-suffix match lengths
 * - Removes repetitive portion if ratio threshold is exceeded
 * - Restores original order and trailing newlines
 * @example
 * processed := doCutRepetitiveText("abcabcabc", 0.15)
 * // processed will remove repetitive pattern
 */
func doCutRepetitiveText(text string, ratio float64) string {
	if strings.TrimSpace(text) == "" {
		return text
	}

	// 计算text末尾有多少\n
	lastLineCount := 0
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == '\n' {
			lastLineCount++
		} else {
			break
		}
	}

	// 逆转补全文本
	reversedText := reverseString(strings.TrimRight(text, " \t\n\r"))

	// 计算当前逆转后的补全文本的最长前后缀长度
	matchLengths := computePrefixSuffixMatchLength(reversedText)
	maxMatchLengths := 0
	for _, length := range matchLengths {
		if length > maxMatchLengths {
			maxMatchLengths = length
		}
	}

	// 若最长前后缀长度/补全内容长度大于等于ratio则判断为重复内容
	if maxMatchLengths > 0 && float64(maxMatchLengths)/float64(len(reversedText)) >= ratio {
		if maxMatchLengths+1 < len(reversedText) {
			reversedText = reversedText[maxMatchLengths+1:]
		} else {
			reversedText = ""
		}
	}

	// 还原字符串
	result := reverseString(reversedText)

	// 还原\n
	for i := 0; i < lastLineCount; i++ {
		result += "\n"
	}

	return result
}

/**
 * Calculate the longest common substring between two strings
 * @param {string} a - First string for comparison
 * @param {string} b - Second string for comparison
 * @returns {string} Returns longest common substring, empty string if no common substring
 * @description
 * - Returns empty string if either input is empty
 * - Uses dynamic programming approach with O(m*n) complexity
 * - Tracks maximum length and ending position of common substring
 * - Extracts and returns the longest common substring
 * @example
 * lcs := longestCommonSubstring("abcdef", "xcdefx")
 * // lcs will be "cdef"
 */
func longestCommonSubstring(a, b string) string {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return ""
	}

	prev := make([]int, n+1)
	maxLen := 0
	end := 0

	for i := 1; i <= m; i++ {
		current := make([]int, n+1)
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				current[j] = prev[j-1] + 1
				if current[j] > maxLen {
					maxLen = current[j]
					end = i
				}
			}
		}
		prev = current
	}

	if maxLen == 0 {
		return ""
	}

	return a[end-maxLen : end]
}

/**
 * Check for extreme repetition patterns in code
 * @param {string} code - Code content to check for extreme repetition
 * @returns {bool, string, int} Returns (hasExtremeRepetition, repeatedPattern, repetitionCount)
 * @description
 * - Returns (false, "", 0) for empty code or insufficient lines
 * - Filters out empty lines before analysis
 * - Requires at least 5 non-empty lines for analysis
 * - Finds longest common substring between consecutive lines
 * - Checks if LCS length is significant (> 5 chars and >= half line length)
 * - Counts occurrences with same position in subsequent lines
 * - Returns true if repetition count > 8 or > half of total lines
 * @example
 * hasRepetition, pattern, count := IsExtremeRepetition("line1\nline1\nline1")
 * if hasRepetition {
 *     fmt.Printf("Pattern '%s' repeated %d times", pattern, count)
 * }
 */
func IsExtremeRepetition(code string) (bool, string, int) {
	if len(code) == 0 {
		return false, "", 0
	}

	// 去除空行并获取非空行
	lines := strings.Split(code, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			nonEmptyLines = append(nonEmptyLines, trimmedLine)
		}
	}

	if len(nonEmptyLines) < 5 {
		return false, "", 0
	}

	n := len(nonEmptyLines)

	for i := 0; i < n-1; i++ {
		lcs := longestCommonSubstring(nonEmptyLines[i], nonEmptyLines[i+1])

		// 如果最长公共子串长度大于5且不小于行数的一半，则进行匹配过程
		if len(lcs) > 5 && len(lcs) >= len(nonEmptyLines[i])/2 {
			// 查找lcs在第一个字符串中的位置
			firstLineLcsIndex := strings.Index(nonEmptyLines[i], lcs)
			if firstLineLcsIndex == -1 {
				continue
			}

			// 统计重复次数
			count := 0
			for k := i + 1; k < n; k++ {
				if strings.Contains(nonEmptyLines[k], lcs) {
					if strings.Index(nonEmptyLines[k], lcs) == firstLineLcsIndex {
						count++
					}
				}
			}

			// 如果重复次数超过8或超过总行数的一半，则认为存在极端重复
			if count > 8 || count > n/2 {
				return true, lcs, count
			}
		}
	}

	return false, "", 0
}

/**
 * Return the minimum of two integers
 * @param {int} a - First integer to compare
 * @param {int} b - Second integer to compare
 * @returns {int} Returns the smaller of the two integers
 * @description
 * - Simple utility function for finding minimum value
 * - Used throughout the codebase for boundary checks
 * @example
 * smaller := min(5, 10)
 * // smaller will be 5
 */
// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// 判断是否为Python文本
func IsPythonText(text string) bool {
	pythonTextRules := os.Getenv("PYTHON_TEXT_RULES")
	if pythonTextRules == "" {
		pythonTextRules = "return self.name"
	}
	rules := strings.Split(pythonTextRules, ",")
	for _, rule := range rules {
		if strings.Contains(text, rule) {
			return true
		}
	}
	return false
}

// CSS相关正则表达式
var (
	cssPropertyPattern          = regexp.MustCompile(`^\s*[a-zA-Z-]+\s*:\s*[^;]+;\s*$`)
	cssSelectorPattern          = regexp.MustCompile(`^\s*[.#]?[a-zA-Z0-9_-]+\s*\{`)
	cssCommentPattern           = regexp.MustCompile(`/\*.*?\*/`)
	multilineCssPropertyPattern = regexp.MustCompile(`^\s*[a-zA-Z-]+\s*:\s*[^;]+;\s*$`)
)

// FrontLanguageEnum 前端语言枚举
type FrontLanguageEnum string

const (
	FrontLanguageVue  FrontLanguageEnum = "vue"
	FrontLanguageHTML FrontLanguageEnum = "html"
	FrontLanguageTS   FrontLanguageEnum = "typescript"
	FrontLanguageCSS  FrontLanguageEnum = "css"
)

var frontendLanguages []FrontLanguageEnum = []FrontLanguageEnum{FrontLanguageVue, FrontLanguageHTML, FrontLanguageTS, FrontLanguageCSS}

// 判断文本是否为css样式
func JudgeCss(language string, text string, ratio float64) bool {
	// 检查语言是否为前端语言
	isFrontLanguage := false
	for _, lang := range frontendLanguages {
		if language == string(lang) {
			isFrontLanguage = true
			break
		}
	}

	if !isFrontLanguage {
		return false
	}

	if strings.Count(text, "\n") == 0 {
		return false
	}

	// 判断是否为多行CSS属性
	if multilineCssPropertyPattern.MatchString(text) {
		return true
	}

	count := 0
	lineCount := 0
	// 判断单行是否为CSS属性
	for _, line := range strings.Split(text, "\n") {
		if line == "\n" || line == "" {
			continue
		}
		if includeCss(line) {
			count++
		}
		lineCount++
	}

	if lineCount > 0 && float64(count)/float64(lineCount) > ratio {
		return true
	}
	return false
}

// 包含css样式
func includeCss(line string) bool {
	// 去除CSS注释
	line = cssCommentPattern.ReplaceAllString(line, "")

	// 检查是否包含CSS属性
	if cssPropertyPattern.MatchString(line) {
		return true
	}

	// 检查是否包含CSS选择器
	if cssSelectorPattern.MatchString(line) {
		return true
	}

	return false
}

// 用于判断text字符串中括号是否完整
func IsValidBrackets(text string) bool {
	stack := make([]rune, 0)
	mapping := map[rune]rune{
		')': '(',
		'}': '{',
		']': '[',
	}

	for _, char := range text {
		if closer, ok := mapping[char]; ok {
			if len(stack) == 0 || stack[len(stack)-1] != closer {
				return false
			}
			stack = stack[:len(stack)-1]
		} else if strings.ContainsRune("({[", char) {
			stack = append(stack, char)
		}
	}

	return len(stack) == 0
}
