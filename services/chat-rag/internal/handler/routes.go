package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
)

func RegisterHandlers(router *gin.Engine, serverCtx *bootstrap.ServiceContext) {
	apiGroup := router.Group("/chat-rag/api")
	{
		// 为需要身份验证的路由应用中间件
		apiGroup.POST("/v1/chat/completions", IdentityMiddleware(serverCtx), ChatCompletionHandler(serverCtx))
		apiGroup.GET("/v1/chat/requests/:requestId/status", ChatStatusHandler(serverCtx))

		// 添加转发接口 - 支持所有HTTP方法（仅在启用时注册）
		if serverCtx.Config.Forward.Enabled {
			apiGroup.Any("/forward/*path", ForwardHandler(serverCtx))
		}
	}

	// 添加健康检查端点 - 用于K8s liveness probe
	router.GET("/health", HealthHandler(serverCtx))

	// 添加就绪检查端点 - 用于K8s readiness probe
	router.GET("/ready", ReadyHandler(serverCtx))

	// 指标端点
	router.GET("/metrics", MetricsHandler(serverCtx))
}

// HealthHandler 处理健康检查请求
func HealthHandler(ctx *bootstrap.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "chat-rag",
		})
	}
}

// ReadyHandler 处理就绪检查请求
func ReadyHandler(ctx *bootstrap.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查关键依赖服务是否就绪
		if ctx.RedisClient == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not ready",
				"reason":    "Redis connection not established",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// 如果有其他关键依赖，可以在这里添加检查
		// 例如：数据库连接、外部API等

		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Unix(),
			"service":   "chat-rag",
		})
	}
}
