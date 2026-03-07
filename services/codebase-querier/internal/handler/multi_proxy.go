package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
)

// MultiProxyHandler 多路由代理处理器
type MultiProxyHandler struct {
	routeHandlers map[string]*ProxyHandler
	routeConfigs  []config.RouteConfig
	mu            sync.RWMutex
}

// NewMultiProxyHandler 创建多路由代理处理器
func NewMultiProxyHandler(cfg *config.ProxyConfig) *MultiProxyHandler {
	handlers := make(map[string]*ProxyHandler)

	for _, route := range cfg.Routes {
		// 转换重写规则
		rules := make([]RewriteRule, len(cfg.Rewrite.Rules))
		for i, rule := range cfg.Rewrite.Rules {
			rules[i] = RewriteRule{From: rule.From, To: rule.To}
		}

		singleConfig := &ProxyConfig{
			Mode:    cfg.Mode,
			Target:  TargetConfig{URL: route.Target.URL, Timeout: route.Target.Timeout},
			Rewrite: RewriteConfig{Enabled: cfg.Rewrite.Enabled, Rules: rules},
			Headers: HeadersConfig{PassThrough: cfg.Headers.PassThrough, Exclude: cfg.Headers.Exclude, Override: cfg.Headers.Override},
		}

		handlers[route.PathPrefix] = NewProxyHandler(singleConfig)
		logx.Infof("Registered route: %s -> %s", route.PathPrefix, route.Target.URL)
	}

	return &MultiProxyHandler{
		routeHandlers: handlers,
		routeConfigs:  cfg.Routes,
	}
}

// ServeHTTP 处理多路由代理请求
func (h *MultiProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	var matchedPrefix string
	var handler *ProxyHandler

	h.mu.RLock()
	for prefix, h := range h.routeHandlers {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > len(matchedPrefix) {
				matchedPrefix = prefix
				handler = h
			}
		}
	}
	h.mu.RUnlock()

	if handler == nil {
		logx.Errorf("No route found for path: %s", path)
		http.NotFound(w, r)
		return
	}

	logx.Infof("Routing request: %s -> %s (prefix: %s)", path, handler.proxyLogic.GetTargetURL(), matchedPrefix)
	handler.ServeHTTP(w, r)
}

// HealthCheck 健康检查处理器
func (h *MultiProxyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type RouteHealth struct {
		PathPrefix string `json:"path_prefix"`
		TargetURL  string `json:"target_url"`
		Healthy    bool   `json:"healthy"`
		Duration   string `json:"duration,omitempty"`
		Error      string `json:"error,omitempty"`
	}

	var routes []RouteHealth

	h.mu.RLock()
	for _, route := range h.routeConfigs {
		handler, exists := h.routeHandlers[route.PathPrefix]
		if !exists {
			continue
		}

		healthy, duration, err := handler.proxyLogic.HealthCheck(r.Context())
		routeHealth := RouteHealth{
			PathPrefix: route.PathPrefix,
			TargetURL:  handler.proxyLogic.GetTargetURL(),
			Healthy:    healthy,
		}

		if err != nil {
			routeHealth.Error = err.Error()
		} else {
			routeHealth.Duration = duration.String()
		}

		routes = append(routes, routeHealth)
	}
	h.mu.RUnlock()

	response := map[string]interface{}{
		"status": "ok",
		"routes": routes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Close 关闭所有处理器
func (h *MultiProxyHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var errs []error
	for prefix, handler := range h.routeHandlers {
		if err := handler.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close handler for %s: %w", prefix, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing handlers: %v", errs)
	}

	return nil
}
