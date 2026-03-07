package proxy

import (
	"net/http"
	"strings"
)

// FilterHeaders 过滤请求头
func FilterHeaders(headers http.Header, exclude []string, override map[string]string) http.Header {
	filtered := make(http.Header)

	// 复制原始header
	for key, values := range headers {
		// 检查是否需要排除
		if shouldExclude(key, exclude) {
			continue
		}

		// 复制所有值
		for _, value := range values {
			filtered.Add(key, value)
		}
	}

	// 应用覆盖规则
	for key, value := range override {
		filtered.Set(key, value)
	}

	// 移除连接相关的header
	removeConnectionHeaders(filtered)

	return filtered
}

// shouldExclude 检查header是否应该被排除
func shouldExclude(key string, exclude []string) bool {
	for _, pattern := range exclude {
		// 支持通配符匹配
		if strings.Contains(pattern, "*") {
			pattern = strings.ReplaceAll(pattern, "*", "")
			if strings.Contains(strings.ToLower(key), strings.ToLower(pattern)) {
				return true
			}
		} else if strings.EqualFold(key, pattern) {
			return true
		}
	}
	return false
}

// removeConnectionHeaders 移除连接相关的header
func removeConnectionHeaders(headers http.Header) {
	// 移除标准连接header
	headers.Del("Connection")
	headers.Del("Keep-Alive")
	headers.Del("Proxy-Connection")
	headers.Del("Proxy-Authenticate")
	headers.Del("Proxy-Authorization")
	headers.Del("TE")
	headers.Del("Trailers")
	headers.Del("Transfer-Encoding")
	headers.Del("Upgrade")

	// 移除hop-by-hop header
	connection := headers.Get("Connection")
	if connection != "" {
		for _, h := range strings.Split(connection, ",") {
			headers.Del(strings.TrimSpace(h))
		}
	}
}

// CopyHeaders 复制响应header到目标writer
func CopyHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// IsSensitiveHeader 判断是否为敏感header
func IsSensitiveHeader(key string) bool {
	sensitive := []string{
		"Authorization",
		"Cookie",
		"Set-Cookie",
		"X-Api-Key",
		"X-Auth-Token",
		"X-Csrf-Token",
		"X-Forwarded-For",
		"X-Real-Ip",
	}

	for _, s := range sensitive {
		if strings.EqualFold(key, s) {
			return true
		}
	}
	return false
}
