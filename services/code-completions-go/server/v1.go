package server

import (
	"code-completion/pkg/completions"
	"code-completion/pkg/model"
	"code-completion/pkg/stream_controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary 兼容千流补全接口的代码补全
// @Description 根据提供的代码上下文生成代码补全建议
// @Tags completions
// @Accept json
// @Produce json
// @Param request body completions.CompletionRequest true "补全请求"
// @Success 200 {object} completions.CompletionResponse
// @Failure 400 {object} completions.CompletionResponse
// @Failure 500 {object} completions.CompletionResponse
// @Router /code-completion/api/v1/completions [post]
func CompletionsV1(c *gin.Context) {
	var req completions.CompletionInput
	if err := c.ShouldBindJSON(&req.CompletionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": model.StatusReqError,
			"error":  err.Error(),
		})
		return
	}
	req.Headers = c.Request.Header

	rsp := stream_controller.Controller.ProcessCompletionV1(c.Request.Context(), &req)
	respCompletion(c, req.ClientID, "sangfor/v1", rsp)
}
