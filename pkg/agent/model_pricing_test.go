package agent

import (
	"testing"

	"jane/pkg/config"
	"jane/pkg/providers"
)

func TestEnrichUsageWithCost(t *testing.T) {
	usage := enrichUsageWithCost(&config.Config{
		ModelList: []config.ModelConfig{{
			ModelName:      "gpt-5.4",
			Model:          "openai/gpt-5.4",
			PricePerMToken: 10,
		}},
	}, "gpt-5.4", &providers.UsageInfo{
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

func TestEnrichUsageWithSplitCost(t *testing.T) {
	usage := enrichUsageWithCost(&config.Config{
		ModelList: []config.ModelConfig{{
			ModelName:            "miniMax",
			Model:                "openrouter/minimax/minimax-m2.7",
			InputPricePerMToken:  1.3,
			OutputPricePerMToken: 2.6,
		}},
	}, "miniMax", &providers.UsageInfo{
		PromptTokens:     1000,
		CompletionTokens: 500,
		TotalTokens:      1500,
	})

	want := (1000*1.3 + 500*2.6) / 1_000_000
	if usage == nil || !usage.HasEstimatedCost {
		t.Fatalf("usage = %+v, want estimated cost", usage)
	}
	if usage.EstimatedCostUSD != want {
		t.Fatalf("EstimatedCostUSD = %v, want %v", usage.EstimatedCostUSD, want)
	}
}
