package agent

import (
	"strings"

	"jane/pkg/config"
	"jane/pkg/providers"
)

func enrichUsageWithCost(cfg *config.Config, model string, usage *providers.UsageInfo) *providers.UsageInfo {
	if usage == nil {
		return nil
	}
	enriched := *usage
	if enriched.HasEstimatedCost {
		return &enriched
	}

	rate := lookupTokenPrice(cfg, model)
	if rate <= 0 {
		return &enriched
	}

	enriched.EstimatedCostUSD = float64(enriched.TotalTokens) * rate / 1_000_000
	enriched.HasEstimatedCost = true
	return &enriched
}

func lookupTokenPrice(cfg *config.Config, model string) float64 {
	if cfg == nil {
		return 0
	}

	normalized := normalizePricedModel(model)
	for i := range cfg.ModelList {
		candidate := cfg.ModelList[i]
		if candidate.PricePerMToken <= 0 {
			continue
		}
		if candidate.ModelName == model || normalizePricedModel(candidate.Model) == normalized {
			return candidate.PricePerMToken
		}
	}

	if mc, err := cfg.GetModelConfig(model); err == nil && mc != nil {
		return mc.PricePerMToken
	}
	return 0
}

func normalizePricedModel(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	for _, prefix := range []string{"openai/", "anthropic/", "claude-cli/"} {
		normalized = strings.TrimPrefix(normalized, prefix)
	}
	return normalized
}
