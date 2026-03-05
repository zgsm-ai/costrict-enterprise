package strategies

import (
	"context"
	"fmt"

	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/promptflow/ds"
	"github.com/zgsm-ai/chat-rag/internal/promptflow/processor"
	"github.com/zgsm-ai/chat-rag/internal/tokenizer"
	"github.com/zgsm-ai/chat-rag/internal/types"
)

type RagOnlyProcessor struct {
	ctx          context.Context
	tokenCounter *tokenizer.TokenCounter
	config       config.Config
	identity     *model.Identity

	end *processor.End
}

// NewRagOnlyProcessor creates a new RAG compression processor
func NewRagOnlyProcessor(
	ctx context.Context,
	svcCtx *bootstrap.ServiceContext,
	identity *model.Identity,
) (*RagOnlyProcessor, error) {
	return &RagOnlyProcessor{
		ctx:          ctx,
		config:       svcCtx.Config,
		tokenCounter: svcCtx.TokenCounter,
		identity:     identity,
	}, nil
}

// Arrange processes the prompt with RAG compression
func (p *RagOnlyProcessor) Arrange(messages []types.Message) (*ds.ProcessedPrompt, error) {
	promptMsg, err := processor.NewPromptMsg(messages)
	if err != nil {
		return &ds.ProcessedPrompt{
			Messages: messages,
		}, fmt.Errorf("create prompt message: %w", err)
	}

	if err := p.buildProcessorChain(); err != nil {
		return &ds.ProcessedPrompt{
			Messages: messages,
		}, fmt.Errorf("build processor chain: %w", err)
	}

	// Since semantic search is no longer used, we directly pass to end processor
	p.end.Execute(promptMsg)

	return p.createProcessedPrompt(promptMsg), nil
}

// buildProcessorChain constructs and connects the processor chain
func (p *RagOnlyProcessor) buildProcessorChain() error {
	p.end = processor.NewEndpoint()

	// Since semantic search is no longer used, we only have end processor
	return nil
}

// createProcessedPrompt creates the final processed prompt result
func (p *RagOnlyProcessor) createProcessedPrompt(
	promptMsg *processor.PromptMsg,
) *ds.ProcessedPrompt {
	processor.SetLanguage(p.identity.Language, promptMsg)
	return &ds.ProcessedPrompt{
		Messages: promptMsg.AssemblePrompt(),
	}
}
