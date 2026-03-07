package handler

import (
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
)

// proxyHealthCheckHandler 代理健康检查处理器
func proxyHealthCheckHandler(serverCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 兼容旧版本
		if serverCtx.ProxyHandler != nil {
			serverCtx.ProxyHandler.HealthCheck(w, r)
			return
		}

		http.Error(w, "No proxy handler configured", http.StatusNotImplemented)
	}
}

// proxyHandler 代理处理器（兼容旧版本）
func proxyHandler(serverCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if serverCtx.ProxyHandler != nil {
			serverCtx.ProxyHandler.ServeHTTP(w, r)
			return
		}

		http.Error(w, "No proxy handler configured", http.StatusNotImplemented)
	}
}
