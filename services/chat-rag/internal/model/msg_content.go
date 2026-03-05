package model

import (
	"fmt"

	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// ContentTextType defines different types of content in prompt message
type ContentTextType string

const (
	// ContTypeText content type
	ContTypeText ContentTextType = "text"
)

type Content struct {
	Type         ContentTextType `json:"type"`
	Text         string          `json:"text"`
	CacheControl any             `json:"cache_control,omitempty"`
}

// ExtractMsgContent extracts and normalizes system content from message
func (p *Content) ExtractMsgContent(msg *types.Message) ([]Content, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	switch v := msg.Content.(type) {
	case string:
		logger.Info("message content is string type",
			zap.String("method", "ExtractMsgContent"),
		)
		content := []Content{
			{
				Type: ContTypeText,
				Text: v,
			},
		}
		return content, nil

	case []interface{}:
		return p.extractFromContentList(v)

	case []Content:
		logger.Info("message content is []Content type",
			zap.String("method", "ExtractMsgContent"),
		)
		return v, nil

	default:
		return nil, fmt.Errorf("unsupported content type: %T", msg.Content)
	}
}

// extractFromContentList extracts content from []interface{} type
func (p *Content) extractFromContentList(contentList []interface{}) ([]Content, error) {
	logger.Info("converted content to []Content type from []interface{}",
		zap.String("method", "extractFromContentList"),
	)
	var systemContents []Content

	for _, item := range contentList {
		contentMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		text, ok := contentMap["text"].(string)
		if !ok {
			continue
		}

		content := Content{
			Type: ContTypeText,
			Text: text,
		}

		if cacheControl, exists := contentMap["cache_control"]; exists {
			content.CacheControl = cacheControl
		}

		systemContents = append(systemContents, content)
	}

	return systemContents, nil
}
