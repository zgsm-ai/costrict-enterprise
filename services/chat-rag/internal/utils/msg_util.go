package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

const (
	ContentTypeText     = "text"
	ContentTypeImageURL = "image_url"
)

// GetContentAsString converts content to string without parsing internal structure
func GetContentAsString(content interface{}) string {
	// Returns raw JSON content directly
	con, ok := content.(string)
	if ok {
		return con
	}
	contentListAny, ok := content.([]any)
	if ok {
		var contentStr string
		for _, contentItem := range contentListAny {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			if contentMap["type"] == ContentTypeText {
				if subStr, ok := contentMap["text"].(string); ok {
					contentStr += subStr
				}
			}
		}
		return contentStr
	}

	// compatible Content type
	contentList, ok := content.([]model.Content)
	if ok {
		var contentStr string
		for _, contentItem := range contentList {
			contentStr += contentItem.Text
		}
		return contentStr
	}
	return ""
}

// GetUserMsgs filters out non-system messages
func GetUserMsgs(messages []types.Message) []types.Message {
	filtered := make([]types.Message, 0, len(messages))
	for _, msg := range messages {
		if msg.Role != types.RoleSystem {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// GetSystemMsg returns the first system message from messages
func GetSystemMsg(messages []types.Message) types.Message {
	for _, msg := range messages {
		if msg.Role == types.RoleSystem {
			return msg
		}
	}

	logger.Info("no system message found",
		zap.String("method", "GetSystemMsg"))
	return types.Message{Role: types.RoleSystem, Content: ""}
}

// TruncateContent truncates content to a specified length for logging
func TruncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

// GetLastUserMsgContent gets the newest user message content from message list
func GetLastUserMsgContent(messages []types.Message) (string, error) {
	lastUserMsg, err := GetLastUserMsg(messages)
	if err != nil {
		return "", err
	}

	return GetContentAsString(lastUserMsg.Content), nil
}

// GetLastUserMsg gets the newest user message from message list
func GetLastUserMsg(messages []types.Message) (types.Message, error) {
	latestUserMsg := GetRecentUserMsgsWithNum(messages, 1)
	if len(latestUserMsg) == 0 {
		return types.Message{}, fmt.Errorf("no user message found")
	}

	return latestUserMsg[0], nil
}

// GetOldUserMsgsWithNum returns messages between the first system message and the num-th last user message
func GetOldUserMsgsWithNum(messages []types.Message, num int) []types.Message {
	if num <= 0 {
		return messages
	}

	if num >= len(messages) {
		return []types.Message{}
	}

	// Assume system message is at position 0
	sysPos := 0
	if len(messages) == 0 || messages[0].Role != types.RoleSystem {
		// If not at 0, find the first system message
		for i := 0; i < len(messages); i++ {
			if messages[i].Role == types.RoleSystem {
				sysPos = i
				break
			}
		}
	}

	// Find position of num-th last user or tool message
	userCount := 0
	userPos := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == types.RoleUser || messages[i].Role == types.RoleTool {
			userCount++
			if userCount == num {
				userPos = i
				break
			}
		}
	}

	// If no user message found, return all messages after system
	if userPos == -1 {
		logger.Info("no user message found",
			zap.String("method", "GetOldUserMsgsWithNum"))
		if sysPos >= len(messages)-1 {
			return []types.Message{}
		}
		return messages[sysPos+1:]
	}

	// Return messages between system and user positions
	if sysPos >= userPos {
		return []types.Message{}
	}
	return messages[sysPos+1 : userPos]
}

// GetRecentUserMsgsWithNum gets messages starting from the num-th user message from the end
// Returns messages from the position of the num-th user message from the end
func GetRecentUserMsgsWithNum(messages []types.Message, num int) []types.Message {
	if num <= 0 {
		return messages
	}

	// Find the position of the num-th user message from the end
	userCount := 0
	position := -1

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == types.RoleUser || messages[i].Role == types.RoleTool {
			userCount++
			if userCount == num {
				position = i
				break
			}
		}
	}

	// If we didn't find enough user messages, return empty slice
	if position == -1 {
		logger.Info("no enough user message found",
			zap.String("method", "GetRecentUserMsgsWithNum"))
		return []types.Message{}
	}

	// Return messages from the position onwards
	return messages[position:]
}

// MarshalJSONWithoutEscapeHTML marshals object to JSON string without escaping HTML characters
func MarshalJSONWithoutEscapeHTML(v interface{}) (string, error) {
	// Create a custom JSON encoder with EscapeHTML set to false to avoid escaping HTML characters
	builder := &strings.Builder{}
	encoder := json.NewEncoder(builder)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		return "", err
	}
	jsonResult := builder.String()
	// Remove the newline character automatically added by Encode
	if len(jsonResult) > 0 && jsonResult[len(jsonResult)-1] == '\n' {
		jsonResult = jsonResult[:len(jsonResult)-1]
	}
	return jsonResult, nil
}

// ExtractTextFromContent extracts text string from content item
func ExtractTextFromContent(item interface{}) string {
	contentMap, ok := item.(map[string]interface{})
	if !ok {
		return ""
	}

	text, exists := contentMap["text"]
	if !exists {
		return ""
	}

	textStr, ok := text.(string)
	if !ok {
		return ""
	}

	return textStr
}

// ExtractSystemContent extracts content from system message
func ExtractSystemContent(systemMsg *types.Message) (string, error) {
	var content model.Content
	contents, err := content.ExtractMsgContent(systemMsg)
	if err != nil {
		return "", fmt.Errorf("failed to extract message content: %w", err)
	}

	if len(contents) != 1 {
		return "", fmt.Errorf("expected one system content, got %d", len(contents))
	}

	return contents[0].Text, nil
}
