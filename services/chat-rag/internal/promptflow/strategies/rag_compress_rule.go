package strategies

import (
	"context"
	"net/http"

	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/promptflow/processor"
)

type RagWithRuleProcessor struct {
	RagCompressProcessor

	rulesConfig  *config.RulesConfig
	ruleInjector *processor.RulesInjector
}

// NewRagWithRuleProcessor creates a new processor with rule injection
func NewRagWithRuleProcessor(
	ctx context.Context,
	svcCtx *bootstrap.ServiceContext,
	headers *http.Header,
	identity *model.Identity,
	modelName string,
	promoptMode string,
) (*RagWithRuleProcessor, error) {
	ragCompressProcessor, err := NewRagCompressProcessor(ctx, svcCtx, headers, identity, modelName, promoptMode)
	if err != nil {
		return nil, err
	}

	processor := &RagWithRuleProcessor{
		RagCompressProcessor: *ragCompressProcessor,
		rulesConfig:          svcCtx.Config.Rules,
	}

	processor.chainBuilder = processor

	return processor, nil
}

// buildProcessorChain constructs and connects the processor chain
func (r *RagWithRuleProcessor) buildProcessorChain() error {
	// First build the parent chain
	err := r.RagCompressProcessor.buildProcessorChain()
	if err != nil {
		return err
	}

	// Create rule injector
	r.ruleInjector = processor.NewRulesInjector(r.promptMode, r.rulesConfig, r.agentName)

	// Rebuild chain with rule injector inserted at the beginning
	r.xmlToolAdapter.SetNext(r.ruleInjector)
	r.ruleInjector.SetNext(r.end)
	// The rest of the chain remains the same as in parent

	return nil
}
