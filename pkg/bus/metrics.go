package bus

type MessageMetrics struct {
	ToolCalls        int     `json:"tool_calls"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd,omitempty"`
	HasEstimatedCost bool    `json:"has_estimated_cost,omitempty"`
}
