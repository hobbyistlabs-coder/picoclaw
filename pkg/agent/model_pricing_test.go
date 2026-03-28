package agent

import (
	"testing"

	"jane/pkg/providers"
)

func TestEnrichUsageWithCost(t *testing.T) {
	usage := enrichUsageWithCost("gpt-5.4", &providers.UsageInfo{
		PromptTokens:     1000,
		CompletionTokens: 500,
		TotalTokens:      1500,
	})

	if usage == nil || !usage.HasEstimatedCost {
		t.Fatalf("usage = %+v, want estimated cost", usage)
	}
	if usage.EstimatedCostUSD <= 0 {
		t.Fatalf("EstimatedCostUSD = %v, want > 0", usage.EstimatedCostUSD)
	}
}
