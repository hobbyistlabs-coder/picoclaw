package agent

import (
	"jane/pkg/bus"
	"jane/pkg/providers"
)

type turnMetrics struct {
	toolCalls        int
	promptTokens     int
	completionTokens int
	totalTokens      int
	estimatedCostUSD float64
	hasEstimatedCost bool
}

func (m *turnMetrics) addUsage(usage *providers.UsageInfo) {
	if usage == nil {
		return
	}
	m.promptTokens += usage.PromptTokens
	m.completionTokens += usage.CompletionTokens
	m.totalTokens += usage.TotalTokens
	if usage.HasEstimatedCost {
		m.estimatedCostUSD += usage.EstimatedCostUSD
		m.hasEstimatedCost = true
	}
}

func (m turnMetrics) usage() *providers.UsageInfo {
	if m.promptTokens == 0 && m.completionTokens == 0 && m.totalTokens == 0 {
		return nil
	}
	return &providers.UsageInfo{
		PromptTokens:     m.promptTokens,
		CompletionTokens: m.completionTokens,
		TotalTokens:      m.totalTokens,
		EstimatedCostUSD: m.estimatedCostUSD,
		HasEstimatedCost: m.hasEstimatedCost,
	}
}

func (m turnMetrics) outbound() *bus.MessageMetrics {
	if m.toolCalls == 0 && m.promptTokens == 0 && m.completionTokens == 0 && m.totalTokens == 0 {
		return nil
	}
	return &bus.MessageMetrics{
		ToolCalls:        m.toolCalls,
		PromptTokens:     m.promptTokens,
		CompletionTokens: m.completionTokens,
		TotalTokens:      m.totalTokens,
		EstimatedCostUSD: m.estimatedCostUSD,
		HasEstimatedCost: m.hasEstimatedCost,
	}
}
