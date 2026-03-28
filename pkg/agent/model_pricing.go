package agent

import (
	"strings"

	"jane/pkg/providers"
)

type tokenRates struct {
	inputUSDPerMTok  float64
	outputUSDPerMTok float64
}

var tokenRatesByModel = map[string]tokenRates{
	"gpt-5.4":           {inputUSDPerMTok: 2.5, outputUSDPerMTok: 15},
	"gpt-5-4":           {inputUSDPerMTok: 2.5, outputUSDPerMTok: 15},
	"gpt-5.4-mini":      {inputUSDPerMTok: 0.75, outputUSDPerMTok: 4.5},
	"gpt-5-4-mini":      {inputUSDPerMTok: 0.75, outputUSDPerMTok: 4.5},
	"gpt-5.4-nano":      {inputUSDPerMTok: 0.2, outputUSDPerMTok: 1.25},
	"gpt-5-4-nano":      {inputUSDPerMTok: 0.2, outputUSDPerMTok: 1.25},
	"claude-opus-4.6":   {inputUSDPerMTok: 5, outputUSDPerMTok: 25},
	"claude-opus-4-6":   {inputUSDPerMTok: 5, outputUSDPerMTok: 25},
	"claude-sonnet-4.6": {inputUSDPerMTok: 3, outputUSDPerMTok: 15},
	"claude-sonnet-4-6": {inputUSDPerMTok: 3, outputUSDPerMTok: 15},
	"claude-haiku-4.5":  {inputUSDPerMTok: 1, outputUSDPerMTok: 5},
	"claude-haiku-4-5":  {inputUSDPerMTok: 1, outputUSDPerMTok: 5},
}

func enrichUsageWithCost(model string, usage *providers.UsageInfo) *providers.UsageInfo {
	if usage == nil {
		return nil
	}
	enriched := *usage
	if enriched.HasEstimatedCost {
		return &enriched
	}

	rates, ok := tokenRatesByModel[normalizePricedModel(model)]
	if !ok {
		return &enriched
	}

	enriched.EstimatedCostUSD =
		(float64(enriched.PromptTokens)*rates.inputUSDPerMTok +
			float64(enriched.CompletionTokens)*rates.outputUSDPerMTok) / 1_000_000
	enriched.HasEstimatedCost = true
	return &enriched
}

func normalizePricedModel(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	for _, prefix := range []string{"openai/", "anthropic/", "claude-cli/"} {
		normalized = strings.TrimPrefix(normalized, prefix)
	}
	return normalized
}
