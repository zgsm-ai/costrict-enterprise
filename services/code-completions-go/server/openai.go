package server

import (
	"code-completion/pkg/model"
	"code-completion/pkg/stream_controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary openai/completions接口的代码补全
// @Description 根据提供的代码上下文生成代码补全建议（OPENAI协议的请求格式）
// @Tags completions
// @Accept json
// @Produce json
// @Param request body model.CompletionParameter true "补全请求"
// @Success 200 {object} completions.CompletionResponse
// @Failure 400 {object} completions.CompletionResponse
// @Failure 500 {object} completions.CompletionResponse
// @Router /api/completions [post]
func CompletionsOpenAI(c *gin.Context) {
	var req model.CompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": model.StatusReqError,
			"error":  err.Error(),
		})
		return
	}
	rsp := stream_controller.Controller.ProcessCompletionOpenAI(c.Request.Context(), &req)
	respCompletion(c, "", "openai", rsp)
}
