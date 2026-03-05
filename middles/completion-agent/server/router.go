package server

import (
	"net/http"
	"time"

	"completion-agent/pkg/logger"
	"completion-agent/pkg/metrics"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建Gin实例
	r := gin.New()

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

	// 补全接口 - 新版本路径（与客户端脚本保持一致）
	api := r.Group("/completion-agent/api/v1")
	api.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})
	api.POST("/completions", Completions)
	api.POST("/logs", logHandler)

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

type LogSettings struct {
	Level string `json:"level"`
}

// logHandler 日志级别设置处理器
// @Summary 设置日志级别
// @Description 设置应用程序的日志级别
// @Tags logs
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
