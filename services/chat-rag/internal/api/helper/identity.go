package helper

import (
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// GetIdentityFromHeaders extracts request headers and creates Identity struct
func GetIdentityFromHeaders(c *gin.Context) *model.Identity {
	caller := getHeaderWithDefault(c, types.HeaderCaller, "chat")
	sender := getHeaderWithDefault(c, types.HeaderQuotaIdentity, "system")

	clientIDE := c.GetHeader(types.HeaderClientIde)
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

// getHeaderWithDefault retrieves a header value from the request context,
// or returns a default value if the header is not present.
func getHeaderWithDefault(c *gin.Context, headerKey, defaultValue string) string {
	if value := c.GetHeader(headerKey); value != "" {
		return value
	}
	return defaultValue
}
