package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// IdentityMiddleware is an optional authentication middleware
// It extracts identity information from request headers and stores it in context
func IdentityMiddleware(svcCtx *bootstrap.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract identity information from request headers
		identity := getIdentityFromHeaders(c)

		// Store identity information in context
		ctxWithIdentity := context.WithValue(c.Request.Context(), model.IdentityContextKey, identity)

		// Also store x-request-id directly in context for logger access
		if identity.RequestID != "" {
			ctxWithIdentity = context.WithValue(ctxWithIdentity, types.HeaderRequestId, identity.RequestID)
		}
		// If request verification is enabled, perform verification
		if svcCtx.Config.RequestVerify.Enabled {
			if err := verifyRequest(c, identity, svcCtx); err != nil {
				sendErrorResponse(c, http.StatusBadRequest, err)
				c.Abort()
				return
			}
		}

		c.Request = c.Request.WithContext(ctxWithIdentity)

		// Continue processing the request
		c.Next()
	}
}

func verifyRequest(c *gin.Context, identity *model.Identity, svcCtx *bootstrap.ServiceContext) error {
	// verify x-request-id
	verifyTime := false
	if identity == nil {
		// jump verification if identity is nil
		return nil
	}
	if svcCtx != nil {
		verifyTime = svcCtx.Config.RequestVerify.EnabledTimeVerify
	}
	if !uuidV7Verify(identity.RequestID, verifyTime) {
		logger.Warn("invalid x-request-id", zap.String("request-id", identity.RequestID))
		return fmt.Errorf("请使用官方 CoStrict 客户端访问模型服务 | Please use the official CoStrict client to access the model service")
	}
	return nil
}

// uuidV7Verify validates if the given string is a valid UUID v7
// and optionally verifies if the timestamp is within 5 minutes
func uuidV7Verify(id string, verifyTimestamp bool) bool {
	// Parse UUID string
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return false
	}

	// Verify UUID version is 7
	if parsedUUID.Version() != 7 {
		return false
	}

	// If timestamp verification is not required, return true
	if !verifyTimestamp {
		return true
	}

	// Extract timestamp from UUID v7
	// UUID v7 format: 48-bit timestamp (milliseconds) in the first 6 bytes
	uuidBytes := parsedUUID[:]

	// Extract 48-bit timestamp (first 6 bytes)
	timestampMs := int64(uuidBytes[0])<<40 |
		int64(uuidBytes[1])<<32 |
		int64(uuidBytes[2])<<24 |
		int64(uuidBytes[3])<<16 |
		int64(uuidBytes[4])<<8 |
		int64(uuidBytes[5])

	// Convert milliseconds to time.Time
	timestamp := time.UnixMilli(timestampMs)

	// Get current time
	now := time.Now()

	// Calculate time difference
	diff := now.Sub(timestamp)

	// Verify if timestamp is within 5 minutes (300 seconds)
	// Allow both past and future timestamps within the range
	if diff < -5*time.Minute || diff > 5*time.Minute {
		return false
	}

	return true
}
