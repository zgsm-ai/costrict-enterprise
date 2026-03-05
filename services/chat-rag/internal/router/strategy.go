package router

import (
	"context"
	"net/http"

	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/types"
)

// Strategy defines a routing strategy that selects a model
type Strategy interface {
	Name() string
	// Run performs semantic routing and returns the top selected model, the current user input snapshot,
	// and an ordered list of candidate model names (best to worst) for degradation attempts.
	Run(ctx context.Context, svcCtx *bootstrap.ServiceContext, headers *http.Header, req *types.ChatCompletionRequest) (selectedModel string, currentUserInput string, orderedCandidates []string, err error)
}
