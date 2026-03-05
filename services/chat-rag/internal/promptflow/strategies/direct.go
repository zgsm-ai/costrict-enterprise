package strategies

import (
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/promptflow/ds"
	"github.com/zgsm-ai/chat-rag/internal/types"
)

// DirectProcessor directly passes through messages without processing
type DirectProcessor struct {
	identity *model.Identity
}

func NewDirectProcessor(identity *model.Identity) *DirectProcessor {
	return &DirectProcessor{
		identity: identity,
	}
}

// Arrange implements the PromptProcessor interface for DirectProcessor
func (d *DirectProcessor) Arrange(messages []types.Message) (*ds.ProcessedPrompt, error) {
	return &ds.ProcessedPrompt{
		Messages: messages,
	}, nil
}
