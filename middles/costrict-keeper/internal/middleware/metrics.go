package middleware

import (
	"time"

	"costrict-keeper/services"

	"github.com/gin-gonic/gin"
)

/**
 * HTTP请求统计中间件
 * @description
 * - 统计HTTP服务器收到的请求数量
 * - 记录请求处理时间
 * - 区分成功和失败的请求
 * - 为健康检查接口提供请求数据
 */
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算请求处理时间
		duration := time.Since(start).Seconds()

		// 获取请求状态码
		statusCode := c.Writer.Status()

		// 构造服务名称（使用请求路径作为服务名称标识）
		serviceName := c.FullPath()
		if serviceName == "" {
			serviceName = "unknown"
		}

		// 增加请求计数
		services.IncrementRequestCount(serviceName)

		// 记录请求持续时间
		services.RecordRequestDuration(serviceName, duration)

		// 如果是错误请求（状态码 >= 400），增加错误请求计数
		if statusCode >= 400 {
			services.IncrementErrorCount(serviceName)
		}
	}
}

/**
 * 获取总请求数
 * @returns {int64} 返回总请求数
 * @description
 * - 从Prometheus指标中获取总请求数
 * - 用于健康检查接口
 */
func GetTotalRequests() int64 {
	// 这里需要从Prometheus指标中获取总请求数
	// 由于Prometheus客户端API的限制，我们需要维护一个本地计数器
	return services.GetTotalRequestCount()
}

/**
 * 获取错误请求数
 * @returns {int64} 返回错误请求数
 * @description
 * - 从Prometheus指标中获取错误请求数
 * - 用于健康检查接口
 */
func GetErrorRequests() int64 {
	// 这里需要从Prometheus指标中获取错误请求数
	// 由于Prometheus客户端API的限制，我们需要维护一个本地计数器
	return services.GetTotalErrorCount()
}
