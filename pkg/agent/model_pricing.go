package agent

import (
	"strings"

	"jane/pkg/config"
	"jane/pkg/providers"
)

type tokenPrice struct {
	input  float64
	output float64
}

func enrichUsageWithCost(cfg *config.Config, model string, usage *providers.UsageInfo) *providers.UsageInfo {
	if usage == nil {
		return nil
	}
	enriched := *usage
	if enriched.HasEstimatedCost {
		return &enriched
	}

	pricing := lookupTokenPrice(cfg, model)
	if pricing.input <= 0 && pricing.output <= 0 {
		return &enriched
	}

	enriched.EstimatedCostUSD =
		(float64(enriched.PromptTokens) * pricing.input / 1_000_000) +
			(float64(enriched.CompletionTokens) * pricing.output / 1_000_000)
	if enriched.EstimatedCostUSD == 0 && enriched.TotalTokens > 0 {
		fallback := pricing.input
		if fallback <= 0 {
			fallback = pricing.output
		}
		enriched.EstimatedCostUSD = float64(enriched.TotalTokens) * fallback / 1_000_000
	}
	enriched.HasEstimatedCost = true
	return &enriched
}

func lookupTokenPrice(cfg *config.Config, model string) tokenPrice {
	if cfg == nil {
		return tokenPrice{}
	}

	normalized := normalizePricedModel(model)
	for i := range cfg.ModelList {
		candidate := cfg.ModelList[i]
		if candidate.ModelName == model || normalizePricedModel(candidate.Model) == normalized {
			return modelTokenPrice(candidate)
		}
	}

	if mc, err := cfg.GetModelConfig(model); err == nil && mc != nil {
		return modelTokenPrice(*mc)
	}
	return tokenPrice{}
}

func modelTokenPrice(mc config.ModelConfig) tokenPrice {
	input := mc.InputPricePerMToken
	output := mc.OutputPricePerMToken
	if input <= 0 {
		input = mc.PricePerMToken
	}
	if output <= 0 {
		output = mc.PricePerMToken
	}
	return tokenPrice{
		input:  input,
		output: output,
	}
}

func normalizePricedModel(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	for _, prefix := range []string{"openai/", "anthropic/", "claude-cli/"} {
		normalized = strings.TrimPrefix(normalized, prefix)
	}
	return normalized
}
