package codebase_context

import (
	"path/filepath"
	"strings"
)

// CommentFunc 注释函数类型
type CommentFunc func(string) string

// commentWithHash 适用于 # 注释风格的语言
func commentWithHash(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "# "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithSlash 适用于 // 注释风格的语言
func commentWithSlash(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "// "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithDash 适用于 Lua 的 -- 注释风格
func commentWithDash(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "-- "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithDoubleDash 适用于 SQL、Haskell 等使用 '--' 的语言
func commentWithDoubleDash(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "-- "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithDoubleHash 适用于 Dockerfile 使用 ## 注释（虽非标准）
func commentWithDoubleHash(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "## "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithExclamation 适用于 Batch 文件使用 @REM 注释
func commentWithExclamation(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "@REM "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithPercent 适用于 TeX/LaTeX 注释 %
func commentWithPercent(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "% "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithSemicolon 适用于 Lisp、Prolog、INI 等用 ; 注释的语言
func commentWithSemicolon(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "; "+line)
		}
	}

	return strings.Join(result, "\n")
}

// commentWithStar 适用于多行注释风格如 /* ... */ 的语言（简单前缀添加）
func commentWithStar(code string) string {
	if code == "" {
		return ""
	}

	return "/*\n" + code + "\n*/"
}

// commentWithMarkdown 适用于 Markdown 等使用 Markdown 语法的语言
func commentWithMarkdown(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else {
			result = append(result, "<!-- "+line+" -->")
		}
	}

	return strings.Join(result, "\n")
}

// defaultCommenter 默认处理函数
func defaultCommenter(code string) string {
	return commentWithHash(code)
}

// extMap 文件扩展名到注释函数的映射
var extMap = map[string]CommentFunc{
	// 使用 '#' 注释的语言
	".py":  commentWithHash,
	".sh":  commentWithHash,
	".rb":  commentWithHash,
	".pl":  commentWithHash, // Perl
	".tcl": commentWithHash, // Tcl
	".r":   commentWithHash, // R script
	".R":   commentWithHash,
	".mak": commentWithHash, // Makefile

	// 使用 '//' 注释的语言
	".c":      commentWithSlash,
	".h":      commentWithSlash,
	".cpp":    commentWithSlash,
	".cc":     commentWithSlash,
	".hpp":    commentWithSlash,
	".java":   commentWithSlash,
	".js":     commentWithSlash,
	".ts":     commentWithSlash,
	".go":     commentWithSlash,
	".rs":     commentWithSlash,
	".kt":     commentWithSlash,
	".swift":  commentWithSlash,
	".cs":     commentWithSlash, // C#
	".m":      commentWithSlash, // Objective-C
	".scala":  commentWithSlash, // Scala
	".groovy": commentWithSlash, // Groovy

	// 使用 '--' 注释的语言
	".lua": commentWithDash,

	// 使用 '--' 或类似单行注释的语言
	".sql": commentWithDoubleDash, // SQL
	".hs":  commentWithDoubleDash, // Haskell
	".vhd": commentWithDoubleDash, // VHDL

	// 使用 '##' 或特殊注释的语言（自定义）
	".dockerfile": commentWithDoubleHash,

	// 使用 '@REM' 注释的语言（Windows Batch）
	".bat": commentWithExclamation,
	".cmd": commentWithExclamation,

	// 使用 '%' 注释的语言
	".tex": commentWithPercent,
	".sty": commentWithPercent,
	".cls": commentWithPercent,

	// 使用 ';' 注释的语言
	".lisp": commentWithSemicolon,
	".el":   commentWithSemicolon, // Emacs Lisp
	".pro":  commentWithSemicolon, // Prolog
	".ini":  commentWithSemicolon,
	".cfg":  commentWithSemicolon,

	// 多行注释风格（简单模拟）
	".php":  commentWithStar, // PHP 支持 /* */
	".css":  commentWithStar, // CSS
	".scss": commentWithStar, // Sass/SCSS

	// 标签注释
	".html": commentWithMarkdown,
	".md":   commentWithMarkdown,
}

// getComment 根据文件扩展名选择合适的注释函数，并返回注释后的代码
func getComment(filePath string, code string) string {
	if code == "" {
		return ""
	}

	ext := filepath.Ext(filePath)
	if ext == "" {
		// 如果没有扩展名，尝试从文件名中提取
		if strings.Contains(filePath, ".") {
			parts := strings.Split(filePath, ".")
			if len(parts) > 1 {
				ext = "." + parts[len(parts)-1]
			}
		}
	}

	commentFunc, exists := extMap[ext]
	if !exists {
		commentFunc = defaultCommenter
	}

	return commentFunc(code)
}
