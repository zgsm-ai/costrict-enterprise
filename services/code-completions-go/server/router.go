package server

import (
	"net/http"
	"time"

	"code-completion/pkg/logger"
	"code-completion/pkg/metrics"
	"code-completion/pkg/stream_controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建Gin实例
	r := gin.New()

	// 使用自定义日志中间件
	r.Use(ginLogger())

	// 使用恢复中间件，防止panic导致服务器崩溃
	r.Use(gin.Recovery())

	// 健康检查接口
	r.GET("/healthz", healthCheck)

	// Prometheus指标接口
	r.GET("/metrics", func(c *gin.Context) {
		metrics.GetMetricsHandler().ServeHTTP(c.Writer, c.Request)
	})

	// Swagger文档接口
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由组 - 统一设置JSON响应头
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})
	api.POST("/logs", logHandler)
	api.GET("/stats", statsHandler)
	api.GET("/details", detailsHandler)

	// 支持OPENAI标准的补全接口，默认并不开放
	api.POST("/completions", CompletionsOpenAI)
	// 补全接口 - 新版本路径（与客户端脚本保持一致）
	completionRouter := r.Group("/code-completion")
	completionRouter.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})
	completionRouter.POST("/api/v1/completions", CompletionsV1)
	completionRouter.POST("/api/v2/completions", CompletionsV2)

	return r
}

// healthCheck 健康检查处理器
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /healthz [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// statsHandler 统计信息处理器
// @Summary 获取统计信息
// @Description 获取代码补全服务的统计信息
// @Tags debug
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/stats [get]
func statsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
		"data":    stream_controller.Controller.GetStats(),
	})
}

// detailsHandler 详细信息处理器
// @Summary 获取详细信息
// @Description 获取代码补全服务的详细信息
// @Tags debug
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/details [get]
func detailsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
		"data":    stream_controller.Controller.GetDetails(),
	})
}

type LogSettings struct {
	Level string `json:"level"`
}

// logHandler 日志级别设置处理器
// @Summary 设置日志级别
// @Description 设置应用程序的日志级别
// @Tags debug
// @Accept json
// @Produce json
// @Param request body LogSettings true "日志级别设置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/logs [post]
func logHandler(c *gin.Context) {
	var req LogSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.SetLevel(req.Level)

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"level":  req.Level,
	})
}

// gin日志中间件
func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		zap.L().Debug("HTTP Request", zap.Int("status", statusCode),
			zap.String("method", method), zap.String("path", path),
			zap.String("ip", clientIP), zap.Duration("latency", latency),
			zap.Int("bodySize", bodySize),
		)
	}
}
