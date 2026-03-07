package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/utils/proxy"
)

// DynamicProxyHandler 动态代理处理器
type DynamicProxyHandler struct {
	portManager *proxy.PortManager
	proxyConfig *config.ProxyConfig
}

// NewDynamicProxyHandler 创建动态代理处理器
func NewDynamicProxyHandler(cfg *config.ProxyConfig) *DynamicProxyHandler {
	var portManager *proxy.PortManager

	// 优先使用新的端口管理器配置
	portManager = proxy.NewPortManagerWithConfig(cfg.PortManager)

	return &DynamicProxyHandler{
		portManager: portManager,
		proxyConfig: cfg,
	}
}

// ServeHTTP 处理动态代理请求
func (h *DynamicProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logx.Infof("Received dynamic proxy request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// 读取请求体
	var body []byte
	if r.Method != "GET" {
		// 使用 io.ReadAll 读取完整的请求体，但限制最大大小为 10MB 防止内存问题
		const maxBodySize = 100 * 1024 * 1024 // 10MB
		limitReader := io.LimitReader(r.Body, maxBodySize)
		var err error
		body, err = io.ReadAll(limitReader)
		if err != nil {
			logx.Errorf("Failed to read request body: %v", err)
			h.sendError(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
			return
		}

		// 检查是否超出限制
		if len(body) >= maxBodySize {
			logx.Errorf("Request body too large, exceeds %d bytes", maxBodySize)
			h.sendError(w, fmt.Sprintf("Request body too large, exceeds %d bytes", maxBodySize), http.StatusRequestEntityTooLarge)
			return
		}

		// 重置body以便后续使用
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))

		logx.Infof("Read request body: %d bytes", len(body))
	}

	// 从请求获取端口信息（GET请求从params获取，其他请求从body获取）
	portResp, err := h.portManager.GetPortFromHeaders(ctx, r.Method, r.Header, r.URL.Query(), body)
	if err != nil {
		logx.Errorf("Failed to get port: %v", err)
		h.sendError(w, fmt.Sprintf("Failed to get port: %v", err), http.StatusBadRequest)
		return
	}

	logx.Infof("Forwarding portResp to: %v", portResp)

	// 构建目标URL
	targetURL := h.portManager.BuildTargetURL(portResp)
	logx.Infof("Forwarding request to: %s", targetURL)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// 构建目标请求
	targetReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL+r.URL.Path, r.Body)
	if err != nil {
		logx.Errorf("Failed to create target request: %v", err)
		h.sendError(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	logx.Errorf("create target request response targetReq: %v", targetReq)

	// 复制请求头
	for key, values := range r.Header {
		// 跳过内部使用的头
		if key == "clientId" || key == "appName" {
			continue
		}
		for _, value := range values {
			targetReq.Header.Add(key, value)
		}
	}

	// 复制查询参数
	if r.URL.RawQuery != "" {
		targetReq.URL.RawQuery = r.URL.RawQuery
	}

	logx.Infof("forward request: %v", targetReq.URL.RawQuery)

	// 发送请求
	resp, err := client.Do(targetReq)
	if err != nil {
		logx.Errorf("Failed to forward request: %v", err)
		h.sendError(w, fmt.Sprintf("Failed to forward request: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 复制响应状态码和内容
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				logx.Errorf("Failed to write response: %v", writeErr)
				break
			}
		}
		if err != nil {
			break
		}
	}

	logx.Infof("Successfully handled dynamic proxy request: %s %s -> %d", r.Method, r.URL.Path, resp.StatusCode)
}

// HealthCheck 健康检查
func (h *DynamicProxyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 读取请求体
	var body []byte
	if r.Method != "GET" {
		// 使用 io.ReadAll 读取完整的请求体，但限制最大大小为 10MB 防止内存问题
		const maxBodySize = 100 * 1024 * 1024 // 10MB
		limitReader := io.LimitReader(r.Body, maxBodySize)
		var err error
		body, err = io.ReadAll(limitReader)
		if err != nil {
			logx.Errorf("Failed to read request body in health check: %v", err)
			h.sendHealthCheckResponse(w, false, 0, fmt.Sprintf("Failed to read request body: %v", err))
			return
		}

		// 检查是否超出限制
		if len(body) >= maxBodySize {
			logx.Errorf("Health check request body too large, exceeds %d bytes", maxBodySize)
			h.sendHealthCheckResponse(w, false, 0, fmt.Sprintf("Request body too large, exceeds %d bytes", maxBodySize))
			return
		}

		// 重置body以便后续使用
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))

		logx.Infof("Health check read request body: %d bytes", len(body))
	}

	// 从请求获取端口信息（GET请求从params获取，其他请求从body获取）
	portResp, err := h.portManager.GetPortFromHeaders(ctx, r.Method, r.Header, r.URL.Query(), body)
	if err != nil {
		logx.Errorf("Health check failed to get port: %v", err)
		h.sendHealthCheckResponse(w, false, 0, fmt.Sprintf("Failed to get port: %v", err))
		return
	}

	// 构建目标URL
	targetURL := h.portManager.BuildTargetURL(portResp)
	healthURL := targetURL + "/health"

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送健康检查请求
	start := time.Now()
	resp, err := client.Get(healthURL)
	duration := time.Since(start)

	if err != nil {
		logx.Errorf("Health check failed: %v", err)
		h.sendHealthCheckResponse(w, false, duration, fmt.Sprintf("Health check failed: %v", err))
		return
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	h.sendHealthCheckResponse(w, healthy, duration, "")
}

// sendHealthCheckResponse 发送健康检查响应
func (h *DynamicProxyHandler) sendHealthCheckResponse(w http.ResponseWriter, healthy bool, duration time.Duration, errorMsg string) {
	status := "ok"
	if !healthy {
		status = "unhealthy"
	}

	response := map[string]interface{}{
		"status": status,
		"proxy": map[string]interface{}{
			"mode":             h.proxyConfig.Mode,
			"dynamic_port":     true,
			"port_manager_url": h.proxyConfig.PortManagerURL,
			"reachable":        healthy,
			"response_time_ms": duration.Milliseconds(),
		},
	}

	if errorMsg != "" {
		response["error"] = errorMsg
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendError 发送错误响应
func (h *DynamicProxyHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":      "PROXY_ERROR",
		"message":   message,
		"timestamp": time.Now().UTC(),
	})
}

// Close 关闭处理器
func (h *DynamicProxyHandler) Close() error {
	// 目前没有需要清理的资源
	return nil
}
