package codebase_context

import (
	"strings"
)

// getCodeFirstNLines 获取代码的前 n 行，忽略空白字符行
func getCodeFirstNLines(s string, n int) []string {
	if s == "" || n <= 0 {
		return []string{}
	}

	var lines []string
	start := 0
	length := len(s)

	for start < length && len(lines) < n {
		end := strings.Index(s[start:], "\n")
		if end == -1 {
			// 没有更多的换行符，把剩下的作为一行加入
			line := strings.TrimSpace(s[start:])
			if line != "" {
				lines = append(lines, line)
			}
			break
		} else {
			// 包含 \n 前的内容
			line := strings.TrimSpace(s[start : start+end])
			if line != "" {
				lines = append(lines, line)
			}
		}
		start += end + 1
	}

	return lines
}

// getCodeLastNLines 获取代码的后 n 行，忽略空白字符行
func getCodeLastNLines(s string, n int) []string {
	if s == "" || n <= 0 {
		return []string{}
	}

	var lines []string
	start := len(s)
	end := start

	for start > 0 && len(lines) < n {
		lastNewline := strings.LastIndex(s[:start], "\n")
		if lastNewline == -1 {
			// 没有更多的换行符，把剩下的作为一行加入
			line := strings.TrimSpace(s[0:end])
			if line != "" {
				lines = append(lines, line)
			}
			break
		} else {
			// 包含 \n 后的内容
			line := strings.TrimSpace(s[lastNewline+1 : end])
			if line != "" {
				lines = append(lines, line)
			}
		}
		start = lastNewline
		end = lastNewline
	}

	// 反转数组，因为我们是从后往前收集的
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	return lines
}

// sliceBeforeNthInstance 字符串第n次出现位置之前的内容，如果没有出现，返回整个字符串
func sliceBeforeNthInstance(s, sub string, n int) string {
	if s == "" || sub == "" || n <= 0 {
		return s
	}

	index := findStrN(s, sub, n)
	if index == -1 {
		return s
	}

	return s[:index]
}

// sliceAfterNthInstance 字符串第n次出现位置之后的内容，如果没有出现，返回整个字符串
func sliceAfterNthInstance(s, sub string, n int) string {
	if s == "" || sub == "" || n <= 0 {
		return s
	}

	index := findStrN(s, sub, n)
	if index == -1 {
		return s
	}

	return s[index+len(sub):]
}

// rSliceBeforeNthInstance 从右搜索，第n次出现位置前的内容，如果没有出现，返回整个字符串
func rSliceBeforeNthInstance(s, sub string, n int) string {
	if s == "" || sub == "" || n <= 0 {
		return s
	}

	index := rFindStrN(s, sub, n)
	if index == -1 {
		return s
	}

	return s[:index]
}

// rSliceAfterNthInstance 从右搜索，第n次出现位置之后的内容，如果没有出现，返回整个字符串
func rSliceAfterNthInstance(s, sub string, n int) string {
	if s == "" || sub == "" || n <= 0 {
		return s
	}

	index := rFindStrN(s, sub, n)
	if index == -1 {
		return s
	}

	return s[index+len(sub):]
}

// rFindStrN 从右查找字符串第n次出现的位置
func rFindStrN(s, sub string, n int) int {
	if s == "" || sub == "" || n <= 0 {
		return -1
	}

	count := 0
	index := -1

	for i := len(s) - len(sub); i >= 0; i-- {
		if s[i:i+len(sub)] == sub {
			count++
			if count == n {
				index = i
				break
			}
		}
	}

	return index
}

// findStrN 查找字符串第n次出现的位置
func findStrN(s, sub string, n int) int {
	if s == "" || sub == "" || n <= 0 {
		return -1
	}

	count := 0
	index := -1

	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			count++
			if count == n {
				index = i
				break
			}
		}
	}

	return index
}

// isEmptyLine 判断是否为空白行
func isEmptyLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

// trimEmptyLines 去除字符串前后的空白行
func trimEmptyLines(s string) string {
	if s == "" {
		return ""
	}

	lines := strings.Split(s, "\n")

	// 去除前面的空白行
	start := 0
	for start < len(lines) && isEmptyLine(lines[start]) {
		start++
	}

	// 去除后面的空白行
	end := len(lines) - 1
	for end >= 0 && isEmptyLine(lines[end]) {
		end--
	}

	if start > end {
		return ""
	}

	return strings.Join(lines[start:end+1], "\n")
}

// getLineCount 获取字符串的行数
func getLineCount(s string) int {
	if s == "" {
		return 0
	}

	count := 1
	for _, char := range s {
		if char == '\n' {
			count++
		}
	}

	return count
}

// truncateLines 截断字符串到指定的行数
func truncateLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}

	return strings.Join(lines[:maxLines], "\n")
}
