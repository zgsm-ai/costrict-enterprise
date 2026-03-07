package logic

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// ProxyConfig 代理配置
type ProxyConfig struct {
	Target  TargetConfig  `json:"target" yaml:"target"`
	Rewrite RewriteConfig `json:"rewrite" yaml:"rewrite"`
	Headers HeadersConfig `json:"headers" yaml:"headers"`
}

// TargetConfig 目标服务配置
type TargetConfig struct {
	URL     string        `json:"url" yaml:"url"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// RewriteConfig 路径重写配置
type RewriteConfig struct {
	Enabled bool          `json:"enabled" yaml:"enabled"`
	Rules   []RewriteRule `json:"rules" yaml:"rules"`
}

// RewriteRule 重写规则
type RewriteRule struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// HeadersConfig Header配置
type HeadersConfig struct {
	PassThrough bool              `json:"pass_through" yaml:"pass_through"`
	Exclude     []string          `json:"exclude" yaml:"exclude"`
	Override    map[string]string `json:"override" yaml:"override"`
}

// ProxyError 代理错误响应
type ProxyError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Error 实现error接口
func (e *ProxyError) Error() string {
	return e.Message
}

// 错误码常量
const (
	ErrorCodeBadRequest        = "PROXY_BAD_REQUEST"
	ErrorCodeTargetUnreachable = "PROXY_TARGET_UNREACHABLE"
	ErrorCodeTimeout           = "PROXY_TIMEOUT"
	ErrorCodeInternalError     = "PROXY_INTERNAL_ERROR"
)

// RewriteRule 路径重写规则
type UtilsRewriteRule struct {
	From string
	To   string
}

// ProxyLogic 代理转发逻辑
type ProxyLogic struct {
	cfg    *ProxyConfig
	client *http.Client
}

// NewProxyLogic 创建代理逻辑实例
func NewProxyLogic(cfg *ProxyConfig) *ProxyLogic {
	return &ProxyLogic{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Target.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Forward 执行请求转发
func (l *ProxyLogic) Forward(ctx context.Context, original *http.Request) (*http.Response, error) {
	logx.Infof("Starting to forward request: %s %s", original.Method, original.URL.Path)

	targetReq, err := l.buildTargetRequest(ctx, original)
	if err != nil {
		logx.Errorf("Failed to build target request: %v", err)
		return nil, err
	}

	resp, err := l.client.Do(targetReq)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				return nil, newTimeoutError(urlErr.Error())
			}
		}
		return nil, newTargetUnreachableError(err.Error())
	}

	logx.Infof("Successfully forwarded request, status: %d", resp.StatusCode)
	return resp, nil
}

// buildTargetRequest 构建目标请求
func (l *ProxyLogic) buildTargetRequest(ctx context.Context, original *http.Request) (*http.Request, error) {
	// 解析目标URL
	targetURL, err := url.Parse(l.cfg.Target.URL)
	if err != nil {
		return nil, newInternalError("invalid target URL: " + err.Error())
	}

	// 构建目标路径
	targetPath := original.URL.Path

	// 移除代理前缀
	if strings.HasPrefix(targetPath, "/proxy") {
		targetPath = strings.TrimPrefix(targetPath, "/proxy")
	}

	// 应用路径重写规则
	if l.cfg.Rewrite.Enabled {
		rules := make([]UtilsRewriteRule, len(l.cfg.Rewrite.Rules))
		for i, rule := range l.cfg.Rewrite.Rules {
			rules[i] = UtilsRewriteRule{
				From: rule.From,
				To:   rule.To,
			}
		}
		targetPath = rewritePath(targetPath, rules)
	}

	// 清理路径
	targetPath = cleanPath(targetPath)

	// 构建完整URL
	fullURL := joinPath(targetURL.String(), targetPath)
	if original.URL.RawQuery != "" {
		fullURL += "?" + original.URL.RawQuery
	}

	// 创建新请求
	targetURL, err = url.Parse(fullURL)
	if err != nil {
		return nil, newInternalError("failed to parse target URL: " + err.Error())
	}

	targetReq, err := http.NewRequestWithContext(ctx, original.Method, targetURL.String(), original.Body)
	if err != nil {
		return nil, newInternalError("failed to create target request: " + err.Error())
	}

	// 复制并过滤header
	filteredHeaders := filterHeaders(
		original.Header,
		l.cfg.Headers.Exclude,
		l.cfg.Headers.Override,
	)

	// 设置Host header为目标地址
	if host := targetURL.Host; host != "" {
		filteredHeaders.Set("Host", host)
	}

	targetReq.Header = filteredHeaders

	return targetReq, nil
}

// HealthCheck 检查目标服务健康状态
func (l *ProxyLogic) HealthCheck(ctx context.Context) (bool, time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", l.cfg.Target.URL+"/health", nil)
	if err != nil {
		return false, 0, err
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, duration, nil
	}

	return false, duration, nil
}

// Close 关闭连接池
func (l *ProxyLogic) Close() error {
	l.client.CloseIdleConnections()
	return nil
}

// 工具函数实现
func rewritePath(path string, rules []UtilsRewriteRule) string {
	for _, rule := range rules {
		if strings.HasPrefix(path, rule.From) {
			return strings.Replace(path, rule.From, rule.To, 1)
		}
	}
	return path
}

func cleanPath(path string) string {
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

func joinPath(baseURL, path string) string {
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

func filterHeaders(headers http.Header, exclude []string, override map[string]string) http.Header {
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

func newTimeoutError(details string) *ProxyError {
	return &ProxyError{
		Code:      ErrorCodeTimeout,
		Message:   "Request timeout",
		Details:   details,
		Timestamp: time.Now().UTC(),
	}
}

func newTargetUnreachableError(details string) *ProxyError {
	return &ProxyError{
		Code:      ErrorCodeTargetUnreachable,
		Message:   "Target service is unreachable",
		Details:   details,
		Timestamp: time.Now().UTC(),
	}
}

func newInternalError(details string) *ProxyError {
	return &ProxyError{
		Code:      ErrorCodeInternalError,
		Message:   "Internal server error",
		Details:   details,
		Timestamp: time.Now().UTC(),
	}
}
