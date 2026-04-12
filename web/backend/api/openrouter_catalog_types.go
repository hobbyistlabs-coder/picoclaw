package api

type openRouterCatalogResponse struct {
	Data []openRouterModel `json:"data"`
}

type openRouterModel struct {
	ID                  string                 `json:"id"`
	CanonicalSlug       string                 `json:"canonical_slug"`
	Name                string                 `json:"name"`
	Created             int64                  `json:"created"`
	Description         string                 `json:"description"`
	ContextLength       int                    `json:"context_length"`
	Architecture        openRouterArchitecture `json:"architecture"`
	Pricing             openRouterPricing      `json:"pricing"`
	TopProvider         openRouterTopProvider  `json:"top_provider"`
	SupportedParameters []string               `json:"supported_parameters"`
	DefaultParameters   map[string]any         `json:"default_parameters"`
	PerRequestLimits    map[string]any         `json:"per_request_limits"`
	ExpirationDate      string                 `json:"expiration_date"`
}

type openRouterArchitecture struct {
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Tokenizer        string   `json:"tokenizer"`
	InstructType     string   `json:"instruct_type"`
}

type openRouterPricing struct {
	Prompt            string `json:"prompt"`
	Completion        string `json:"completion"`
	Request           string `json:"request"`
	Image             string `json:"image"`
	WebSearch         string `json:"web_search"`
	InternalReasoning string `json:"internal_reasoning"`
	InputCacheRead    string `json:"input_cache_read"`
	InputCacheWrite   string `json:"input_cache_write"`
}

type openRouterTopProvider struct {
	ContextLength       int  `json:"context_length"`
	MaxCompletionTokens int  `json:"max_completion_tokens"`
	IsModerated         bool `json:"is_moderated"`
}
