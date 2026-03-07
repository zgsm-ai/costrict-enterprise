package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/oidc-auth/internal/repository"
	"github.com/zgsm-ai/oidc-auth/pkg/response"
)

func readinessHandler(c *gin.Context) {
	if repository.GetDB() == nil {
		response.JSONError(c, http.StatusInternalServerError, "", "Database not initialized")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Application ready"})
}
