package model

import (
	"completion-agent/pkg/config"
	"context"
)

type LLM interface {
	Completions(ctx context.Context, param *CompletionParameter) (*CompletionResponse, CompletionStatus, error)
	Config() *config.ModelConfig
}
