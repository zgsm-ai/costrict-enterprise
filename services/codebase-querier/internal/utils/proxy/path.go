package proxy

import (
	"strings"
)

// RewritePath 应用路径重写规则
func RewritePath(path string, rules []RewriteRule) string {
	for _, rule := range rules {
		if strings.HasPrefix(path, rule.From) {
			return strings.Replace(path, rule.From, rule.To, 1)
		}
	}
	return path
}

// RewriteRule 路径重写规则
type RewriteRule struct {
	From string
	To   string
}

// CleanPath 清理路径，移除多余斜杠
func CleanPath(path string) string {
	if path == "" {
		return "/"
	}

	// 确保以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 移除重复的/
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// 移除末尾的/（除非是根路径）
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return path
}

// JoinPath 安全地连接基础URL和路径
func JoinPath(baseURL, path string) string {
	if baseURL == "" {
		return path
	}

	// 确保baseURL不以/结尾
	baseURL = strings.TrimRight(baseURL, "/")

	// 确保path以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return baseURL + path
}
