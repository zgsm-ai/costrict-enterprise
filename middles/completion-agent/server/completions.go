package server

import (
	"completion-agent/pkg/completions"
	"completion-agent/pkg/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Completions 补全接口路由处理
// @Summary 代码补全
// @Description 根据提供的代码上下文生成代码补全建议
// @Tags completions
// @Accept json
// @Produce json
// @Param request body completions.CompletionRequest true "补全请求"
// @Success 200 {object} completions.CompletionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /completion-agent/api/v1/completions [post]
func Completions(c *gin.Context) {
	var req completions.CompletionInput
	if err := c.ShouldBindJSON(&req.CompletionRequest); err != nil {
		zap.L().Error("Completions error", zap.Any("body", c.Request.Form), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status": model.StatusReqError,
			"error":  err.Error(),
		})
		return
	}
	req.Headers = c.Request.Header

	handler := completions.NewCompletionHandler(nil)
	perf := &completions.CompletionPerformance{
		ReceiveTime: time.Now().Local(),
	}
	rc := completions.NewCompletionContext(c.Request.Context(), perf)
	rsp := handler.HandleCompletion(rc, &req)
	respCompletion(c, &req.CompletionRequest, rsp)
}

/**
 * 处理补全响应
 * @param {*gin.Context} c - Gin上下文对象，用于HTTP响应
 * @param {*completions.CompletionRequest} req - 补全请求对象，包含请求参数
 * @param {*completions.CompletionResponse} rsp - 补全响应对象，包含处理结果
 * @description
 * - 根据补全响应的状态记录相应的日志信息
 * - 成功时记录info级别日志，失败时记录warn级别日志
 * - 根据响应状态映射到对应的HTTP状态码
 * - 将响应对象以JSON格式返回给客户端
 * - 支持多种状态码：200(成功)、408(超时)、504(网关超时)、503(服务不可用)等
 * @example
 * req := &completions.CompletionRequest{...}
 * rsp := &completions.CompletionResponse{...}
 * respCompletion(c, req, rsp)
 */
func respCompletion(c *gin.Context, req *completions.CompletionRequest, rsp *completions.CompletionResponse) {
	statusCode := http.StatusOK
	switch rsp.Status {
	case model.StatusSuccess, model.StatusEmpty:
		statusCode = http.StatusOK
	case model.StatusCanceled:
		statusCode = http.StatusRequestTimeout
	case model.StatusTimeout:
		statusCode = http.StatusGatewayTimeout
	case model.StatusBusy:
		statusCode = http.StatusServiceUnavailable
	case model.StatusReqError, model.StatusRejected:
		statusCode = http.StatusBadRequest
	case model.StatusServerError, model.StatusModelError:
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusInternalServerError
	}
	c.JSON(statusCode, rsp)
}
