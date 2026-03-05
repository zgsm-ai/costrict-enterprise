package ds

import (
	"github.com/zgsm-ai/chat-rag/internal/types"
)

// ProcessedPrompt contains the result of prompt processing
type ProcessedPrompt struct {
	Messages     []types.Message    `json:"messages"`
	Tools        []types.Function   `json:"tools"`
	Agent        string             `json:"agent"`
	TokenMetrics types.TokenMetrics `json:"token_metrics"`
}
