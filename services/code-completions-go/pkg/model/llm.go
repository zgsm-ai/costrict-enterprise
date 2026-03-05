package model

import (
	"code-completion/pkg/config"
	"code-completion/pkg/tokenizers"
	"context"
)

type LLM interface {
	Completions(ctx context.Context, param *CompletionParameter) (*CompletionResponse, *CompletionVerbose, CompletionStatus, error)
	Config() *config.ModelConfig
	Tokenizer() *tokenizers.Tokenizer
}
