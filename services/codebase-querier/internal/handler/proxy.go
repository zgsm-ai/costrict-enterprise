package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// ProxyConfig 代理配置
type ProxyConfig struct {
	Mode    string        `json:"mode" yaml:"mode"` // 代理模式: rewrite, full_path
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
	ErrorCodeInvalidMode       = "PROXY_INVALID_MODE"
)

// 代理模式常量
const (
	ProxyModeRewrite  = "rewrite"
	ProxyModeFullPath = "full_path"
)

// UtilsRewriteRule 路径重写规则
type UtilsRewriteRule struct {
	From string
	To   string
}

// PathBuilder 路径构建器接口
type PathBuilder interface {
	BuildPath(originalPath string) (string, error)
}

// RewritePathBuilder 路径重写构建器
type RewritePathBuilder struct {
	rules []UtilsRewriteRule
}

// FullPathBuilder 全路径构建器
type FullPathBuilder struct {
	targetURL string
}

// ProxyLogic 代理转发逻辑
type ProxyLogic struct {
	cfg         *ProxyConfig
	client      *http.Client
	pathBuilder PathBuilder
}

// NewProxyLogic 创建代理逻辑实例
func NewProxyLogic(cfg *ProxyConfig) *ProxyLogic {
	// 构建重写规则
	rules := make([]UtilsRewriteRule, len(cfg.Rewrite.Rules))
	for i, rule := range cfg.Rewrite.Rules {
		rules[i] = UtilsRewriteRule{
			From: rule.From,
			To:   rule.To,
		}
	}

	// 根据模式创建路径构建器
	var pathBuilder PathBuilder
	if cfg.Mode == ProxyModeFullPath {
		pathBuilder = &FullPathBuilder{targetURL: cfg.Target.URL}
	} else {
		pathBuilder = &RewritePathBuilder{rules: rules}
	}

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
		pathBuilder: pathBuilder,
	}
}

// Forward 执行请求转发
func (l *ProxyLogic) Forward(ctx context.Context, original *http.Request) (*http.Response, error) {
	logx.Infof("Starting to forward request: %s %s (mode: %s)", original.Method, original.URL.Path, l.cfg.Mode)

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
	logx.Infof("[PROXY_LOGIC_DEBUG] === Building Target Request ===")
	logx.Infof("[PROXY_LOGIC_DEBUG] Original request path: %s", original.URL.Path)
	logx.Infof("[PROXY_LOGIC_DEBUG] Original full URL: %s", original.URL.String())
	logx.Infof("[PROXY_LOGIC_DEBUG] Proxy mode: %s", l.cfg.Mode)

	var fullURL string
	var err error

	if l.cfg.Mode == ProxyModeFullPath {
		// 全路径模式：使用FullPathBuilder
		fullURL, err = l.pathBuilder.BuildPath(original.URL.Path)
		if err != nil {
			return nil, newInternalError("failed to build full path: " + err.Error())
		}
	} else {
		// rewrite模式：使用传统方式
		targetPath := original.URL.Path
		logx.Infof("[PROXY_LOGIC_DEBUG] Rewrite mode - initial path: %s", targetPath)

		// 移除代理前缀
		if strings.HasPrefix(targetPath, "/proxy") {
			targetPath = strings.TrimPrefix(targetPath, "/proxy")
			logx.Infof("[PROXY_LOGIC_DEBUG] After removing /proxy prefix: %s", targetPath)
		}

		// 应用路径重写规则
		if l.cfg.Rewrite.Enabled {
			logx.Infof("[PROXY_LOGIC_DEBUG] Rewrite enabled, applying %d rules", len(l.cfg.Rewrite.Rules))
			rules := make([]UtilsRewriteRule, len(l.cfg.Rewrite.Rules))
			for i, rule := range l.cfg.Rewrite.Rules {
				rules[i] = UtilsRewriteRule{
					From: rule.From,
					To:   rule.To,
				}
			}
			originalPath := targetPath
			targetPath = rewritePath(targetPath, rules)
			logx.Infof("[PROXY_LOGIC_DEBUG] Path before rewrite: %s", originalPath)
			logx.Infof("[PROXY_LOGIC_DEBUG] Path after rewrite: %s", targetPath)
		}

		// 清理路径
		targetPath = cleanPath(targetPath)
		logx.Infof("[PROXY_LOGIC_DEBUG] After cleaning path: %s", targetPath)

		fullURL = joinPath(l.cfg.Target.URL, targetPath)
	}

	if original.URL.RawQuery != "" {
		fullURL += "?" + original.URL.RawQuery
	}
	logx.Infof("[PROXY_LOGIC_DEBUG] Final target URL: %s", fullURL)
	logx.Infof("[PROXY_LOGIC_DEBUG] === End Building Target Request ===")

	// 创建新请求
	targetURL, err := url.Parse(fullURL)
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

// BuildPath 构建路径（RewritePathBuilder实现）
func (b *RewritePathBuilder) BuildPath(originalPath string) (string, error) {
	if len(b.rules) == 0 {
		return originalPath, nil
	}

	for _, rule := range b.rules {
		if strings.HasPrefix(originalPath, rule.From) {
			newPath := strings.Replace(originalPath, rule.From, rule.To, 1)
			logx.Infof("Path rewritten: %s -> %s", originalPath, newPath)
			return newPath, nil
		}
	}

	return originalPath, nil
}

// BuildPath 构建路径（FullPathBuilder实现）
func (b *FullPathBuilder) BuildPath(originalPath string) (string, error) {
	if originalPath == "" {
		originalPath = "/"
	}

	// 确保路径格式正确
	if !strings.HasPrefix(originalPath, "/") {
		originalPath = "/" + originalPath
	}

	// 解析目标URL
	target, err := url.Parse(b.targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid target URL: %w", err)
	}

	// 构建完整URL
	fullURL := *target
	fullURL.Path = path.Join(target.Path, originalPath)
	fullURL.RawQuery = "" // 查询参数将在转发时处理

	return fullURL.String(), nil
}

// GetTargetURL 获取目标URL
func (l *ProxyLogic) GetTargetURL() string {
	return l.cfg.Target.URL
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

// ProxyHandler 代理处理器
type ProxyHandler struct {
	proxyLogic *ProxyLogic
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(cfg *ProxyConfig) *ProxyHandler {
	return &ProxyHandler{
		proxyLogic: NewProxyLogic(cfg),
	}
}

// ServeHTTP 处理代理请求
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 记录请求日志
	logx.Infof("Received proxy request: %s %s from %s (mode: %s)", r.Method, r.URL.Path, r.RemoteAddr, h.proxyLogic.cfg.Mode)

	// 验证请求
	if err := h.validateRequest(r); err != nil {
		logx.Errorf("Invalid request: %v", err)
		h.sendError(w, err, http.StatusBadRequest)
		return
	}

	// 执行转发
	resp, err := h.proxyLogic.Forward(ctx, r)
	if err != nil {
		logx.Errorf("Failed to forward request: %v", err)
		h.handleForwardError(w, err)
		return
	}
	defer resp.Body.Close()

	// 复制响应
	if err := h.copyResponse(w, resp); err != nil {
		logx.Errorf("Failed to copy response: %v", err)
		h.sendError(w, err, http.StatusInternalServerError)
		return
	}

	logx.Infof("Successfully handled proxy request: %s %s -> %d (mode: %s)", r.Method, r.URL.Path, resp.StatusCode, h.proxyLogic.cfg.Mode)
}

// HealthCheck 健康检查处理器
func (h *ProxyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	healthy, duration, err := h.proxyLogic.HealthCheck(ctx)
	if err != nil {
		logx.Errorf("Health check failed: %v", err)
		httpx.OkJson(w, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	status := "ok"
	if !healthy {
		status = "unhealthy"
	}

	httpx.OkJson(w, map[string]interface{}{
		"status": status,
		"proxy": map[string]interface{}{
			"mode":             h.proxyLogic.cfg.Mode,
			"target_url":       h.proxyLogic.GetTargetURL(),
			"reachable":        healthy,
			"response_time_ms": duration.Milliseconds(),
		},
	})
}

// validateRequest 验证请求
func (h *ProxyHandler) validateRequest(r *http.Request) error {
	// 验证URL长度
	if len(r.URL.String()) > 65536 {
		return &ProxyError{
			Code:    "PROXY_URL_TOO_LONG",
			Message: "Request URL too long",
			Details: "URL length exceeds 64KB limit",
		}
	}

	// 验证Header大小
	headerSize := 0
	for key, values := range r.Header {
		headerSize += len(key)
		for _, value := range values {
			headerSize += len(value)
		}
	}
	if headerSize > 1024*1024 { // 1MB
		return &ProxyError{
			Code:    "PROXY_HEADERS_TOO_LARGE",
			Message: "Request headers too large",
			Details: "Headers size exceeds 1MB limit",
		}
	}

	// 验证HTTP方法
	switch r.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodHead, http.MethodOptions:
		// 有效方法
	default:
		return &ProxyError{
			Code:    "PROXY_METHOD_NOT_ALLOWED",
			Message: "HTTP method not allowed",
			Details: "Method " + r.Method + " is not supported",
		}
	}

	return nil
}

// copyResponse 复制响应
func (h *ProxyHandler) copyResponse(dst http.ResponseWriter, src *http.Response) error {
	// 复制状态码
	dst.WriteHeader(src.StatusCode)

	// 复制Header
	for key, values := range src.Header {
		for _, value := range values {
			dst.Header().Add(key, value)
		}
	}

	// 复制Body
	_, err := io.Copy(dst, src.Body)
	return err
}

// handleForwardError 处理转发错误
func (h *ProxyHandler) handleForwardError(w http.ResponseWriter, err error) {
	if proxyErr, ok := err.(*ProxyError); ok {
		switch proxyErr.Code {
		case ErrorCodeTimeout:
			h.sendError(w, err, http.StatusGatewayTimeout)
		case ErrorCodeTargetUnreachable:
			h.sendError(w, err, http.StatusServiceUnavailable)
		case ErrorCodeBadRequest:
			h.sendError(w, err, http.StatusBadRequest)
		default:
			h.sendError(w, err, http.StatusInternalServerError)
		}
	} else {
		h.sendError(w, err, http.StatusInternalServerError)
	}
}

// sendError 发送错误响应
func (h *ProxyHandler) sendError(w http.ResponseWriter, err error, statusCode int) {
	if proxyErr, ok := err.(*ProxyError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(proxyErr)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":      "PROXY_ERROR",
			"message":   err.Error(),
			"timestamp": time.Now().UTC(),
		})
	}
}

// Close 关闭处理器
func (h *ProxyHandler) Close() error {
	return h.proxyLogic.Close()
}

// 工具函数实现
func rewritePath(path string, rules []UtilsRewriteRule) string {
	logx.Infof("[REWRITE_DEBUG] === Path Rewrite Started ===")
	logx.Infof("[REWRITE_DEBUG] Input path: %s", path)
	logx.Infof("[REWRITE_DEBUG] Number of rules: %d", len(rules))

	for i, rule := range rules {
		logx.Infof("[REWRITE_DEBUG] Checking rule %d: From='%s', To='%s'", i+1, rule.From, rule.To)
		logx.Infof("[REWRITE_DEBUG] Path starts with '%s': %v", rule.From, strings.HasPrefix(path, rule.From))

		if strings.HasPrefix(path, rule.From) {
			newPath := strings.Replace(path, rule.From, rule.To, 1)
			logx.Infof("[REWRITE_DEBUG] Rule %d matched! Path before: %s", i+1, path)
			logx.Infof("[REWRITE_DEBUG] Rule %d matched! Path after: %s", i+1, newPath)
			logx.Infof("[REWRITE_DEBUG] === Path Rewrite Completed (Match Found) ===")
			return newPath
		}
	}

	logx.Infof("[REWRITE_DEBUG] No rules matched, returning original path: %s", path)
	logx.Infof("[REWRITE_DEBUG] === Path Rewrite Completed (No Match) ===")
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
		if strings.Contains(pattern, "*") {
			// 支持通配符匹配
			pattern = strings.ReplaceAll(pattern, "*", "")
			if strings.Contains(key, pattern) {
				return true
			}
		} else if key == pattern {
			return true
		}
	}
	return false
}

func removeConnectionHeaders(headers http.Header) {
	headers.Del("Connection")
	headers.Del("Keep-Alive")
	headers.Del("Proxy-Connection")
	headers.Del("Upgrade")
	headers.Del("Transfer-Encoding")
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
		Message:   "Target service unreachable",
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
