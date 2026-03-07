package response

import (
	"github.com/zgsm-ai/oidc-auth/pkg/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// JSONSuccess returns successful response
func JSONSuccess(c *gin.Context, message, data any) {
	response := gin.H{
		"success":   true,
		"code":      200,
		"message":   message,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// JSONError returns error response (basic version)
func JSONError(c *gin.Context, httpCode int, codeMsg, message string) {
	c.JSON(httpCode, gin.H{
		"code":      codeMsg,
		"success":   false,
		"message":   message,
		"timestamp": time.Now(),
		"data":      "",
	})
}

func HandleError(c *gin.Context, status int, codeMsg string, err error) {
	log.Error(nil, "operation failed: %v", err)
	JSONError(c, status, codeMsg, err.Error())
}
