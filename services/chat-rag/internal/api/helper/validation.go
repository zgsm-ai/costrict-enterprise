package helper

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"go.uber.org/zap"
)

// VerifyRequest verifies the request
func VerifyRequest(c *gin.Context, identity *model.Identity, svcCtx *bootstrap.ServiceContext) error {
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
