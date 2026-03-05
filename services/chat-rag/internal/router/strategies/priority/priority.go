package priority

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// Strategy implements the priority-based round-robin routing strategy
type Strategy struct {
	cfg             config.PriorityConfig
	priorityGroups  map[int]*PriorityGroup
	mu              sync.RWMutex
	lowestPriority  int // Track the highest priority (lowest number)
}

// New creates a new priority strategy instance
func New(cfg config.PriorityConfig) (*Strategy, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid priority config: %w", err)
	}

	s := &Strategy{
		cfg:            cfg,
		priorityGroups: make(map[int]*PriorityGroup),
		lowestPriority: 999,
	}

	// Initialize priority groups from config
	if err := s.initializePriorityGroups(); err != nil {
		return nil, fmt.Errorf("failed to initialize priority groups: %w", err)
	}

	return s, nil
}

// Name returns the strategy name
func (s *Strategy) Name() string {
	return "priority"
}

// Run implements the Strategy interface
func (s *Strategy) Run(
	ctx context.Context,
	svcCtx *bootstrap.ServiceContext,
	headers *http.Header,
	req *types.ChatCompletionRequest,
) (string, string, []string, error) {
	if req == nil || len(req.Messages) == 0 {
		return "", "", nil, nil
	}

	// Only trigger when request model is "auto"
	if !strings.EqualFold(req.Model, "auto") {
		return "", "", nil, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. Filter enabled candidates
	enabledCandidates := s.filterEnabledCandidates()
	if len(enabledCandidates) == 0 {
		logger.WarnC(ctx, "priority router: no enabled candidates found")
		if s.cfg.FallbackModelName != "" {
			return s.cfg.FallbackModelName, "", []string{s.cfg.FallbackModelName}, nil
		}
		return "", "", nil, errors.New("no enabled candidates and no fallback configured")
	}

	// 2. Select the highest priority group (lowest priority number)
	highestPriorityGroup := s.priorityGroups[s.lowestPriority]
	if highestPriorityGroup == nil {
		logger.ErrorC(ctx, "priority router: highest priority group not found",
			zap.Int("priority", s.lowestPriority))
		if s.cfg.FallbackModelName != "" {
			return s.cfg.FallbackModelName, "", []string{s.cfg.FallbackModelName}, nil
		}
		return "", "", nil, errors.New("priority group not found")
	}

	// 3. Select model using round-robin within the priority group
	selectedModel := highestPriorityGroup.selectModelByRoundRobin()

	logger.InfoC(ctx, "priority router: model selected",
		zap.String("selectedModel", selectedModel),
		zap.Int("priority", s.lowestPriority),
		zap.Int("groupSize", len(highestPriorityGroup.models)),
	)

	// 4. Build ordered candidates list (selectedModel first, then by priority and weight)
	orderedCandidates := s.buildOrderedCandidates(selectedModel)

	return selectedModel, "", orderedCandidates, nil
}

// initializePriorityGroups initializes priority groups from configuration
func (s *Strategy) initializePriorityGroups() error {
	for _, candidate := range s.cfg.Candidates {
		if !candidate.Enabled {
			continue
		}

		// Get or create priority group
		group, exists := s.priorityGroups[candidate.Priority]
		if !exists {
			group = newPriorityGroup(candidate.Priority)
			s.priorityGroups[candidate.Priority] = group
		}

		// Add model to group
		model := &ModelCandidate{
			modelName: candidate.ModelName,
			priority:  candidate.Priority,
			weight:    candidate.Weight,
			enabled:   candidate.Enabled,
		}
		group.addModel(model)

		// Track lowest priority number (highest priority)
		if candidate.Priority < s.lowestPriority {
			s.lowestPriority = candidate.Priority
		}
	}

	if len(s.priorityGroups) == 0 {
		return errors.New("no enabled candidates configured")
	}

	return nil
}

// filterEnabledCandidates returns all enabled candidates
func (s *Strategy) filterEnabledCandidates() []*ModelCandidate {
	candidates := make([]*ModelCandidate, 0)
	for _, group := range s.priorityGroups {
		candidates = append(candidates, group.getModels()...)
	}
	return candidates
}

// buildOrderedCandidates builds an ordered list of candidates for degradation
// The selectedModel is always first, followed by other models sorted by priority (ascending)
// and weight (descending) within the same priority
func (s *Strategy) buildOrderedCandidates(selectedModel string) []string {
	result := []string{selectedModel}

	// Get all priority levels sorted (ascending order - lower number = higher priority)
	priorities := make([]int, 0, len(s.priorityGroups))
	for priority := range s.priorityGroups {
		priorities = append(priorities, priority)
	}
	sort.Ints(priorities)

	// Build ordered list
	for _, priority := range priorities {
		group := s.priorityGroups[priority]
		models := group.getModels()

		// Sort by weight descending within same priority
		sort.Slice(models, func(i, j int) bool {
			return models[i].weight > models[j].weight
		})

		for _, model := range models {
			// Skip selectedModel (already first)
			if model.modelName != selectedModel {
				result = append(result, model.modelName)
			}
		}
	}

	// Append fallback model if not already in list
	if s.cfg.FallbackModelName != "" {
		found := false
		for _, m := range result {
			if m == s.cfg.FallbackModelName {
				found = true
				break
			}
		}
		if !found {
			result = append(result, s.cfg.FallbackModelName)
		}
	}

	return result
}

// validateConfig validates the priority configuration
func validateConfig(cfg config.PriorityConfig) error {
	if len(cfg.Candidates) == 0 {
		return errors.New("no candidates configured")
	}

	hasEnabled := false
	for _, candidate := range cfg.Candidates {
		if candidate.ModelName == "" {
			return errors.New("candidate modelName is empty")
		}
		if candidate.Priority < 0 || candidate.Priority > 999 {
			return fmt.Errorf("priority must be between 0 and 999 for model %s", candidate.ModelName)
		}
		if candidate.Weight < 1 || candidate.Weight > 100 {
			return fmt.Errorf("weight must be between 1 and 100 for model %s", candidate.ModelName)
		}
		if candidate.Enabled {
			hasEnabled = true
		}
	}

	if !hasEnabled {
		return errors.New("at least one candidate must be enabled")
	}

	return nil
}
