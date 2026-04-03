package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
)

// MetricsHandler handles Prometheus metrics endpoint
func MetricsHandler(serverCtx *bootstrap.ServiceContext) gin.HandlerFunc {
	handler := promhttp.Handler()
	return gin.WrapH(handler)
}
