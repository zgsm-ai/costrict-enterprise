package server

import (
	"code-completion/pkg/model"
	"code-completion/pkg/stream_controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary sangfor/completions接口的代码补全
// @Description 根据提供的代码上下文生成代码补全建议，该接口使用sangfor/completions接口，请求参数在客户端已经被预处理过了
// @Tags completions
// @Accept json
// @Produce json
// @Param request body model.CompletionParameter true "补全请求"
// @Success 200 {object} completions.CompletionResponse
// @Failure 400 {object} completions.CompletionResponse
// @Failure 500 {object} completions.CompletionResponse
// @Router /code-completion/api/v2/completions [post]
func CompletionsV2(c *gin.Context) {
	var para model.CompletionParameter
	if err := c.ShouldBindJSON(&para); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": model.StatusReqError,
			"error":  err.Error(),
		})
		return
	}
	rsp := stream_controller.Controller.ProcessCompletionV2(c.Request.Context(), &para)
	respCompletion(c, para.ClientID, "sangfor/v2", rsp)
}
