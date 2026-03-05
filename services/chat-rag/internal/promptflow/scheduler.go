package promptflow

import (
	"context"
	"net/http"

	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/promptflow/ds"
	"github.com/zgsm-ai/chat-rag/internal/promptflow/strategies"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// PromptArranger is the interface for processing chat prompts
type PromptArranger interface {
	Arrange(messages []types.Message) (*ds.ProcessedPrompt, error)
}

// NewPromptProcessor creates a new processor based on chat type
func NewPromptProcessor(
	ctx context.Context,
	svcCtx *bootstrap.ServiceContext,
	promptMode types.PromptMode,
	headers *http.Header,
	identity *model.Identity,
	modelName string,
) PromptArranger {
	const fallbackMsg = "falling back to DirectProcessor"

	type processorCreator func() (PromptArranger, error)

	var creator processorCreator
	var modeName string

	switch promptMode {
	case types.Raw:
		modeName = "Direct chat mode"
		creator = func() (PromptArranger, error) {
			return strategies.NewDirectProcessor(identity), nil
		}

	case types.Performance:
		modeName = "RagOnlyProcessor mode"
		creator = func() (PromptArranger, error) {
			return strategies.NewRagOnlyProcessor(ctx, svcCtx, identity)
		}

	case types.Strict:
		modeName = "Strict workflow mode"
		creator = func() (PromptArranger, error) {
			return strategies.NewRagWithRuleProcessor(
				ctx, svcCtx, headers, identity,
				modelName, string(promptMode))
		}

	case types.Cost, types.Balanced, types.Auto:
		fallthrough
	default:
		modeName = "Default processing mode"
		creator = func() (PromptArranger, error) {
			return strategies.NewRagWithRuleProcessor(
				ctx, svcCtx, headers, identity,
				modelName, string(promptMode))
		}
	}

	logger.Info(modeName+" activated",
		zap.String("mode", string(promptMode)),
	)

	processor, err := creator()
	if err != nil {
		logger.Error("Failed to create processor: "+modeName,
			zap.Error(err),
			zap.String("fallback", fallbackMsg),
		)
		return strategies.NewDirectProcessor(identity)
	}

	return processor
}
