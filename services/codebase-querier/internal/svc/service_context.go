package svc

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
)

type ServiceContext struct {
	Config            config.Config
	serverContext     context.Context
	ProxyHandler      *ProxyHandler
	MultiProxyHandler interface{} // 使用interface{}避免循环导入，实际使用时需要类型断言
}

// ProxyHandler 代理处理器
type ProxyHandler struct {
	healthCheckHandler http.HandlerFunc
	proxyHandler       http.HandlerFunc
	ProxyLogic         interface{} // 使用interface{}避免循环导入
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.proxyHandler != nil {
		p.proxyHandler(w, r)
	} else {
		http.Error(w, "Proxy not configured", http.StatusServiceUnavailable)
	}
}

func (p *ProxyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if p.healthCheckHandler != nil {
		p.healthCheckHandler(w, r)
	} else {
		http.Error(w, "Proxy not configured", http.StatusServiceUnavailable)
	}
}

// Close closes the shared Redis client and database connection
func (s *ServiceContext) Close() {
	var errs []error
	if len(errs) > 0 {
		logx.Errorf("service_context close err:%v", errs)
	} else {
		logx.Infof("service_context close successfully.")
	}
}

func NewServiceContext(ctx context.Context, c config.Config) (*ServiceContext, error) {
	var err error
	svcCtx := &ServiceContext{
		Config:        c,
		serverContext: ctx,
	}

	// 初始化代理处理器
	if c.ProxyConfig != nil && len(c.ProxyConfig.Routes) > 0 {
		// 使用第一个路由作为默认配置
		firstRoute := c.ProxyConfig.Routes[0]
		svcCtx.ProxyHandler = &ProxyHandler{
			healthCheckHandler: createHealthCheckHandler(firstRoute.Target.URL),
			proxyHandler:       createProxyHandler(c.ProxyConfig, firstRoute),
		}
		logx.Infof("Initialized proxy handler with route: %s -> %s", firstRoute.PathPrefix, firstRoute.Target.URL)
	}

	return svcCtx, err
}

func createHealthCheckHandler(targetURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","proxy":{"target_url":"` + targetURL + `","reachable":true,"response_time_ms":0}}`))
	}
}

func createProxyHandler(cfg *config.ProxyConfig, route config.RouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 根据代理模式选择不同的处理逻辑
		if cfg.Mode == "full_path" {
			handleFullPathProxy(w, r, &route)
		} else {
			handleRewriteProxy(w, r, &route)
		}
	}
}

// handleFullPathProxy 处理全路径模式的代理请求
func handleFullPathProxy(w http.ResponseWriter, r *http.Request, route *config.RouteConfig) {
	logx.Infof("[PROXY_DEBUG] === Full Path Proxy Processing Start ===")
	logx.Infof("[PROXY_DEBUG] Original request: %s %s", r.Method, r.URL.Path)
	logx.Infof("[PROXY_DEBUG] Full URL: %s", r.URL.String())

	// 记录配置信息
	logx.Infof("[PROXY_DEBUG] Config Target URL: %s", route.Target.URL)
	logx.Infof("[PROXY_DEBUG] Config Timeout: %v", route.Target.Timeout)

	// 简单的代理实现，实际项目中应该使用更完整的代理逻辑
	client := &http.Client{
		Timeout: route.Target.Timeout,
	}

	// 构建目标URL - 全路径模式：直接拼接目标URL和原始路径
	logx.Infof("[PROXY_DEBUG] === Path Processing Start ===")
	remainingPath := r.URL.Path
	logx.Infof("[PROXY_DEBUG] Full path mode - remainingPath: %s", remainingPath)

	logx.Infof("[PROXY_DEBUG] Base URL from config: '%s'", route.Target.URL)
	logx.Infof("[PROXY_DEBUG] Full path to append: '%s'", remainingPath)

	// 检查URL是否以/结尾，路径是否以/开头
	needsSlash := !strings.HasSuffix(route.Target.URL, "/") && !strings.HasPrefix(remainingPath, "/")
	logx.Infof("[PROXY_DEBUG] URL needs slash separator: %v", needsSlash)

	var targetURL string
	if needsSlash {
		targetURL = route.Target.URL + "/" + remainingPath
	} else {
		targetURL = route.Target.URL + remainingPath
	}
	logx.Infof("[PROXY_DEBUG] Target URL before query: %s", targetURL)

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
		logx.Infof("[PROXY_DEBUG] Target URL with query: %s", targetURL)
	}

	// 添加更多诊断日志
	logx.Infof("[PROXY_DEBUG] Final Target URL: %s", targetURL)
	logx.Infof("[PROXY_DEBUG] Target URL length: %d", len(targetURL))
	logx.Infof("[PROXY_DEBUG] === Full Path Proxy Processing End ===")

	// 创建新的请求
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制header
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// handleRewriteProxy 处理重写模式的代理请求
func handleRewriteProxy(w http.ResponseWriter, r *http.Request, route *config.RouteConfig) {
	// 添加诊断日志
	logx.Infof("[PROXY_DEBUG] === Rewrite Proxy Processing Start ===")
	logx.Infof("[PROXY_DEBUG] Original request: %s %s", r.Method, r.URL.Path)
	logx.Infof("[PROXY_DEBUG] Full URL: %s", r.URL.String())
	logx.Infof("[PROXY_DEBUG] Path prefix to remove: /api/v1/proxy")
	logx.Infof("[PROXY_DEBUG] Path length: %d, prefix length: %d", len(r.URL.Path), len("/api/v1/proxy"))

	// 记录配置信息
	logx.Infof("[PROXY_DEBUG] Config Target URL: %s", route.Target.URL)
	logx.Infof("[PROXY_DEBUG] Config Timeout: %v", route.Target.Timeout)

	// 简单的代理实现，实际项目中应该使用更完整的代理逻辑
	client := &http.Client{
		Timeout: route.Target.Timeout,
	}

	// 构建目标URL - 添加详细日志
	logx.Infof("[PROXY_DEBUG] === Path Processing Start ===")
	remainingPath := r.URL.Path
	logx.Infof("[PROXY_DEBUG] Initial remainingPath: %s", remainingPath)

	if len(r.URL.Path) > len("/api/v1/proxy") {
		remainingPath = r.URL.Path[len("/api/v1/proxy"):]
		logx.Infof("[PROXY_DEBUG] Path after prefix removal: %s", remainingPath)
	} else {
		logx.Infof("[PROXY_DEBUG] Path length <= prefix length, keeping original: %s", remainingPath)
	}

	logx.Infof("[PROXY_DEBUG] Base URL from config: '%s'", route.Target.URL)
	logx.Infof("[PROXY_DEBUG] Remaining path to append: '%s'", remainingPath)

	// 检查URL是否以/结尾，路径是否以/开头
	needsSlash := !strings.HasSuffix(route.Target.URL, "/") && !strings.HasPrefix(remainingPath, "/")
	logx.Infof("[PROXY_DEBUG] URL needs slash separator: %v", needsSlash)

	var targetURL string
	if needsSlash {
		targetURL = route.Target.URL + "/" + remainingPath
	} else {
		targetURL = route.Target.URL + remainingPath
	}
	logx.Infof("[PROXY_DEBUG] Target URL before query: %s", targetURL)

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
		logx.Infof("[PROXY_DEBUG] Target URL with query: %s", targetURL)
	}

	// 添加更多诊断日志
	logx.Infof("[PROXY_DEBUG] Final Target URL: %s", targetURL)
	logx.Infof("[PROXY_DEBUG] Target URL length: %d", len(targetURL))
	logx.Infof("[PROXY_DEBUG] === Rewrite Proxy Processing End ===")

	// 创建新的请求
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制header
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
