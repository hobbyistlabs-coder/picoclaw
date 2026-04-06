package agent

import (
	"strings"
	"sync"
)

type asyncBatchState struct {
	mu       sync.Mutex
	expected int
	results  []string
}

func (al *AgentLoop) startAsyncBatch(batchID string, expected int) {
	if batchID == "" || expected < 1 {
		return
	}
	al.asyncBatches.Store(batchID, &asyncBatchState{expected: expected})
}

func (al *AgentLoop) addAsyncBatchResult(
	batchID string, expected int, result string,
) (string, bool) {
	if batchID == "" {
		return result, true
	}
	value, _ := al.asyncBatches.LoadOrStore(batchID, &asyncBatchState{expected: expected})
	state := value.(*asyncBatchState)
	state.mu.Lock()
	defer state.mu.Unlock()
	if expected > state.expected {
		state.expected = expected
	}
	state.results = append(state.results, result)
	if len(state.results) < state.expected {
		return "", false
	}
	al.asyncBatches.Delete(batchID)

	// Bolt: using strings.Join instead of O(N^2) loop string concatenation for performance
	combined := strings.Join(state.results, "\n\n")

	return combined, true
}
