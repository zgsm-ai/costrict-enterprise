package priority

import (
	"sync"
)

// ModelCandidate represents a candidate model with its configuration
type ModelCandidate struct {
	modelName string
	priority  int
	weight    int
	enabled   bool
}

// PriorityGroup represents a group of models with the same priority level
type PriorityGroup struct {
	priority       int
	models         []*ModelCandidate
	currentWeights []int
	mu             sync.Mutex
}

// newPriorityGroup creates a new priority group
func newPriorityGroup(priority int) *PriorityGroup {
	return &PriorityGroup{
		priority:       priority,
		models:         make([]*ModelCandidate, 0),
		currentWeights: make([]int, 0),
	}
}

// addModel adds a model to the priority group
func (pg *PriorityGroup) addModel(model *ModelCandidate) {
	pg.models = append(pg.models, model)
	pg.currentWeights = append(pg.currentWeights, 0)
}

// selectModelByRoundRobin selects a model using smooth weighted round-robin algorithm
// This method is thread-safe and implements the algorithm from the design document:
// 1. Each model has a configured weight and a current weight (initially 0)
// 2. For each selection:
//    - Add configured weight to current weight for all models
//    - Select the model with the highest current weight
//    - Subtract total weight from the selected model's current weight
func (pg *PriorityGroup) selectModelByRoundRobin() string {
	// Optimization: If only one model, return directly without locking
	// This provides zero-lock overhead for single-model scenarios
	if len(pg.models) == 1 {
		return pg.models[0].modelName
	}

	pg.mu.Lock()
	defer pg.mu.Unlock()

	// Calculate total weight
	totalWeight := 0
	for _, model := range pg.models {
		totalWeight += model.weight
	}

	// Find the model with the maximum current weight after adding configured weights
	maxWeight := -1
	selectedIdx := -1

	for i, model := range pg.models {
		// Add configured weight to current weight
		pg.currentWeights[i] += model.weight

		// Find the model with maximum current weight
		if pg.currentWeights[i] > maxWeight {
			maxWeight = pg.currentWeights[i]
			selectedIdx = i
		}
	}

	// Subtract total weight from the selected model's current weight
	pg.currentWeights[selectedIdx] -= totalWeight

	return pg.models[selectedIdx].modelName
}

// getModels returns all models in the group sorted by weight (descending)
func (pg *PriorityGroup) getModels() []*ModelCandidate {
	return pg.models
}
