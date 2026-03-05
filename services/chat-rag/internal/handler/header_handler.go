package handler

import (
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// getHeaderWithDefault retrieves a header value from the request context,
// or returns a default value if the header is not present.
func getHeaderWithDefault(c *gin.Context, headerKey, defaultValue string) string {
	if value := c.GetHeader(headerKey); value != "" {
		return value
	}
	return defaultValue
}

// getIdentityFromHeaders extracts request headers and creates Identity struct
func getIdentityFromHeaders(c *gin.Context) *model.Identity {
	clientIDE := getHeaderWithDefault(c, types.HeaderClientIde, "vscode")
	caller := getHeaderWithDefault(c, types.HeaderCaller, "chat")
	sender := getHeaderWithDefault(c, types.HeaderQuotaIdentity, "system")

	projectPath := c.GetHeader(types.HeaderProjectPath)
	if decodedPath, err := url.PathUnescape(projectPath); err != nil {
		logger.Error("Failed to PathUnescape project path",
			zap.String("projectPath", projectPath),
			zap.Error(err),
		)
	} else {
		projectPath = decodedPath
	}

	jwtToken := c.GetHeader(types.HeaderAuthorization)
	userInfo := model.NewUserInfo(jwtToken)
	logger.Info("User info:", zap.Any("userInfo", userInfo))

	return &model.Identity{
		RequestID:     c.GetHeader(types.HeaderRequestId),
		TaskID:        c.GetHeader(types.HeaderTaskId),
		ClientID:      c.GetHeader(types.HeaderClientId),
		ClientIDE:     clientIDE,
		ClientVersion: c.GetHeader(types.HeaderClientVersion),
		ClientOS:      c.GetHeader(types.HeaderClientOS),
		ProjectPath:   projectPath,
		AuthToken:     jwtToken,
		UserName:      userInfo.Name,
		LoginFrom:     userInfo.ExtractLoginFromToken(),
		Caller:        caller,
		Language:      c.GetHeader(types.HeaderLanguage),
		Sender:        sender,
		UserInfo:      userInfo,
	}
}

// setSSEResponseHeaders sets SSE response headers
func setSSEResponseHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	c.Header("X-Accel-Buffering", "no")
}
