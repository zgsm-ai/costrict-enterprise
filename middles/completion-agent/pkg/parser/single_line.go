package parser

import "strings"

/**
 * 判断是否应该使用单行补全逻辑
 * @param cursorLinePrefix 光标行前缀文本
 * @param cursorLineSuffix 光标行后缀文本
 * @param language 编程语言类型
 * @return bool true-使用单行补全，false-使用多行补全
 * @description
 * 光标所在行后缀非空，则走单行补全（便于语法修复）
 * 若光标行前非空 且 光标所在行后缀为空 且 首单词和次首单词前缀不包含关键词 且 行间单词不包含关键词 则走单行补全
 * ref: https://docs.atrust.sangfor.com/pages/viewpage.action?pageId=361621625
 */
func NeedSingleCompletion(cursorLinePrefix, cursorLineSuffix, language string) bool {
	// 光标所在行后缀非空，单行
	if strings.TrimSpace(cursorLineSuffix) != "" {
		return true
	}

	// 去除前后空格的光标行前缀
	cursorLinePrefixStripped := strings.TrimSpace(cursorLinePrefix)
	// 光标前为空，多行
	if cursorLinePrefixStripped == "" {
		return false
	}

	// 分割单词
	words := strings.Fields(cursorLinePrefixStripped)
	// 获取当前语言的关键词列表
	keywords := getCodeBlockKeywords(language)
	// 检查关键词匹配
	for _, keyword := range keywords {
		// 检查首单词前缀
		if len(words) >= 1 && strings.HasPrefix(words[0], keyword) {
			return false
		}

		// 检查次首单词前缀
		if len(words) >= 2 && strings.HasPrefix(words[1], keyword) {
			return false
		}

		// 检查行间是否包含关键词
		for _, word := range words {
			if word == keyword {
				return false
			}
		}
	}

	return true
}

/**
 * getCodeBlockKeywords 获取指定语言的关键词列表
 * @param language 编程语言类型
 * @return []string 关键词列表
 */
func getCodeBlockKeywords(language string) []string {
	// 如果找不到指定语言的关键词，返回其他语言的关键词作为默认
	keywords, exists := codeBlockKeywordsMap[language]
	if !exists {
		return codeBlockKeywordsMap["other"]
	}
	return keywords
}

var codeBlockKeywordsMap = map[string][]string{
	"python": {
		"if", "else", "elif", "for", "while", "try", "except",
		"finally", "def", "class", "with", "async", "match",
	},
	"go": {
		"func", "if", "else", "for", "switch", "case", "type",
		"select", "defer", "go",
	},
	"c": {
		"typedef", "if", "for", "do", "while", "switch", "case", "void",
	},
	"cpp": {
		"if", "else", "while", "do", "for", "switch", "case", "default",
		"try", "catch", "struct", "enum", "class", "union", "public", "typedef",
	},
	"javascript": {
		"if", "else", "for", "while", "do", "switch", "try", "catch",
		"finally", "function", "class", "with",
	},
	"typescript": {
		"function", "class", "if", "for", "try", "interface",
		"private", "switch", "case",
	},
	"vue": {
		"methods:", "try", "if", "switch", "case", "for",
		"<ix-", "<sf-", "<lx-", "<el-",
	},
	"other": {
		"if", "else", "for", "while", "do", "try", "catch", "finally",
	},
}
