package router

import (
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/router/strategies/priority"
	ssemantic "github.com/zgsm-ai/chat-rag/internal/router/strategies/semantic"
	"go.uber.org/zap"
)

// NewRunner creates a strategy instance based on config
func NewRunner(cfg config.RouterConfig) Strategy {
	switch cfg.Strategy {
	case "semantic", "":
		return ssemantic.New(cfg.Semantic)
	case "priority":
		strategy, err := priority.New(cfg.Priority)
		if err != nil {
			logger.Error("priority router: failed to create strategy",
				zap.Error(err),
			)
			return nil
		}
		return strategy
	default:
		logger.Info("router: no strategy matched",
			zap.String("strategy", cfg.Strategy),
		)
		return nil
	}
}
