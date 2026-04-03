package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/api/helper"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
)

// IdentityMiddleware is an optional authentication middleware
// It extracts identity information from request headers and stores it in context
func IdentityMiddleware(svcCtx *bootstrap.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract identity information from request headers
		identity := helper.GetIdentityFromHeaders(c)

		// Store identity information in context
		ctxWithIdentity := context.WithValue(c.Request.Context(), model.IdentityContextKey, identity)

		// Also store x-request-id directly in context for logger access
		if identity.RequestID != "" {
			ctxWithIdentity = context.WithValue(ctxWithIdentity, types.HeaderRequestId, identity.RequestID)
		}
		// If request verification is enabled, perform verification
		if svcCtx.Config.RequestVerify.Enabled {
			if err := helper.VerifyRequest(c, identity, svcCtx); err != nil {
				helper.SendErrorResponse(c, 400, err)
				c.Abort()
				return
			}
		}

		c.Request = c.Request.WithContext(ctxWithIdentity)

		// Continue processing the request
		c.Next()
	}
}
