package config

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// LoadYAML loads yaml from the specified file path using viper
func LoadYAML[T any](path string) (*T, error) {
	var yaml T

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read YAML: %w", err)
	}

	if err := viper.Unmarshal(&yaml); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &yaml, nil
}

// MustLoadConfig loads configuration and panics if there's an error
func MustLoadConfig(configPath string) Config {
	c, err := LoadYAML[Config](configPath)
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Apply defaults: if fallbackModelName not set, use the first candidate
	if c != nil && c.Router != nil && c.Router.Semantic.Routing.FallbackModelName == "" {
		if len(c.Router.Semantic.Routing.Candidates) > 0 {
			c.Router.Semantic.Routing.FallbackModelName = c.Router.Semantic.Routing.Candidates[0].ModelName
		}
	}

	// Align rule engine prefix defaults with original plugin logic
	if c != nil && c.Router != nil {
		if c.Router.Semantic.RuleEngine.BodyPrefix == "" {
			c.Router.Semantic.RuleEngine.BodyPrefix = "body."
		}
		if c.Router.Semantic.RuleEngine.HeaderPrefix == "" {
			c.Router.Semantic.RuleEngine.HeaderPrefix = "header."
		}
	}

	// Align stripCodeFences default behavior with plugin:
	// default to true when the key is not explicitly set in YAML
	if c != nil && c.Router != nil {
		// inputExtraction.protocol default
		if c.Router.Semantic.InputExtraction.Protocol == "" {
			c.Router.Semantic.InputExtraction.Protocol = "openai"
		}
		// inputExtraction.userJoinSep default
		if c.Router.Semantic.InputExtraction.UserJoinSep == "" {
			c.Router.Semantic.InputExtraction.UserJoinSep = "\n\n"
		}
		// inputExtraction.stripCodeFences default (only when key not set)
		if !viper.IsSet("router.semantic.inputExtraction.stripCodeFences") {
			c.Router.Semantic.InputExtraction.StripCodeFences = true
		}
		// inputExtraction.codeFenceRegex default is empty string (no-op) — keep if unset
		// inputExtraction.maxUserMessages default
		if !viper.IsSet("router.semantic.inputExtraction.maxUserMessages") || c.Router.Semantic.InputExtraction.MaxUserMessages == 0 {
			c.Router.Semantic.InputExtraction.MaxUserMessages = 100
		}
		// inputExtraction.maxHistoryBytes default
		if !viper.IsSet("router.semantic.inputExtraction.maxHistoryBytes") || c.Router.Semantic.InputExtraction.MaxHistoryBytes == 0 {
			c.Router.Semantic.InputExtraction.MaxHistoryBytes = 2048
		}
		// inputExtraction.maxHistoryMessages default
		if !viper.IsSet("router.semantic.inputExtraction.maxHistoryMessages") || c.Router.Semantic.InputExtraction.MaxHistoryMessages == 0 {
			c.Router.Semantic.InputExtraction.MaxHistoryMessages = 5
		}
	}
	// Apply idle timeout defaults
	if c != nil && c.LLMTimeout.IdleTimeoutMs <= 0 {
		c.LLMTimeout.IdleTimeoutMs = 180000
		logger.Info("llm idle timeout not set, using default", zap.Int("idleTimeoutMs", c.LLMTimeout.IdleTimeoutMs))
	}
	if c != nil && c.LLMTimeout.TotalIdleTimeoutMs <= 0 {
		c.LLMTimeout.TotalIdleTimeoutMs = 180000
		logger.Info("llm total idle timeout not set, using default", zap.Int("totalIdleTimeoutMs", c.LLMTimeout.TotalIdleTimeoutMs))
	}

	// Apply retry configuration defaults for regular mode
	if c != nil && c.LLMTimeout.MaxRetryCount < 0 {
		c.LLMTimeout.MaxRetryCount = 1
		logger.Info("llm maxRetryCount not set or negative, using default", zap.Int("maxRetryCount", c.LLMTimeout.MaxRetryCount))
	}
	if c != nil && c.LLMTimeout.RetryIntervalMs <= 0 {
		c.LLMTimeout.RetryIntervalMs = 5000
		logger.Info("llm retryIntervalMs not set, using default", zap.Int("retryIntervalMs", c.LLMTimeout.RetryIntervalMs))
	}

	// Apply forward configuration defaults
	if c != nil {
		// forward.enabled default
		if !viper.IsSet("forward.enabled") {
			c.Forward.Enabled = false
		}
		// forward.defaultTarget default
		if !viper.IsSet("forward.defaultTarget") {
			c.Forward.DefaultTarget = ""
		}
		// vipPriority.enabled default (only when key not set)
		if !viper.IsSet("vipPriority.enabled") {
			c.VIPPriority.Enabled = false
			logger.Info("vipPriority.enabled not set, using default", zap.Bool("enabled", c.VIPPriority.Enabled))
		}
	}

	// Apply timeout and retry defaults for routing (model degradation scenarios)
	ApplyRouterDefaults(c)

	logger.Info("loaded config", zap.Any("config", c))
	return *c
}

// ApplyRouterDefaults applies default values to router configuration
// This function can be called after loading config from file or Nacos
func ApplyRouterDefaults(c *Config) {
	if c == nil || c.Router == nil || !c.Router.Enabled {
		return
	}

	if c.Router.Strategy == "semantic" {
		// Timeout defaults for semantic strategy
		if c.Router.Semantic.Routing.IdleTimeoutMs <= 0 {
			c.Router.Semantic.Routing.IdleTimeoutMs = 180000 // Same as regular mode: 180s
			logger.Info("router idle timeout not set, using default", zap.Int("idleTimeoutMs", c.Router.Semantic.Routing.IdleTimeoutMs))
		}
		if c.Router.Semantic.Routing.TotalIdleTimeoutMs <= 0 {
			c.Router.Semantic.Routing.TotalIdleTimeoutMs = 180000
			logger.Info("router total idle timeout not set, using default", zap.Int("totalIdleTimeoutMs", c.Router.Semantic.Routing.TotalIdleTimeoutMs))
		}

		// Retry defaults for semantic strategy
		if c.Router.Semantic.Routing.MaxRetryCount < 0 {
			c.Router.Semantic.Routing.MaxRetryCount = 1
			logger.Info("router maxRetryCount not set or negative, using default", zap.Int("maxRetryCount", c.Router.Semantic.Routing.MaxRetryCount))
		}
		if c.Router.Semantic.Routing.RetryIntervalMs <= 0 {
			c.Router.Semantic.Routing.RetryIntervalMs = 5000
			logger.Info("router retryIntervalMs not set, using default", zap.Int("retryIntervalMs", c.Router.Semantic.Routing.RetryIntervalMs))
		}
	} else if c.Router.Strategy == "priority" {
		// Timeout defaults for priority strategy
		if c.Router.Priority.IdleTimeoutMs <= 0 {
			c.Router.Priority.IdleTimeoutMs = 180000 // Same as regular mode: 180s
			logger.Info("priority router idle timeout not set, using default", zap.Int("idleTimeoutMs", c.Router.Priority.IdleTimeoutMs))
		}
		if c.Router.Priority.TotalIdleTimeoutMs <= 0 {
			c.Router.Priority.TotalIdleTimeoutMs = 180000
			logger.Info("priority router total idle timeout not set, using default", zap.Int("totalIdleTimeoutMs", c.Router.Priority.TotalIdleTimeoutMs))
		}

		// Retry defaults for priority strategy
		if c.Router.Priority.MaxRetryCount < 0 {
			c.Router.Priority.MaxRetryCount = 1
			logger.Info("priority router maxRetryCount not set or negative, using default", zap.Int("maxRetryCount", c.Router.Priority.MaxRetryCount))
		}
		if c.Router.Priority.RetryIntervalMs <= 0 {
			c.Router.Priority.RetryIntervalMs = 5000
			logger.Info("priority router retryIntervalMs not set, using default", zap.Int("retryIntervalMs", c.Router.Priority.RetryIntervalMs))
		}
	}
}
