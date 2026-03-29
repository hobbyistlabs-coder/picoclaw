package providers

import (
	"encoding/json"
	"regexp"
	"strings"
)

var internalExecutionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)\[TOOL_CALL\].*?\[/TOOL_CALL\]`),
	regexp.MustCompile(`(?is)<tool_call>.*?</tool_call>`),
	regexp.MustCompile(`(?is)<minimax:tool_call>.*?</minimax:tool_call>`),
	regexp.MustCompile(`(?im)^\s*\{[\s\S]*"tool"\s*:\s*"[^"]+"[\s\S]*"parameters"\s*:\s*\{[\s\S]*\}\s*\}\s*$`),
	regexp.MustCompile(`(?im)^\s*<invoke\b.*$`),
	regexp.MustCompile(`(?im)^\s*</invoke>\s*$`),
}

// extractToolCallsFromText parses tool call JSON from response text.
// Both ClaudeCliProvider and CodexCliProvider use this to extract
// tool calls that the model outputs in its response text.
func extractToolCallsFromText(text string) []ToolCall {
	start := strings.Index(text, `{"tool_calls"`)
	if start == -1 {
		return nil
	}

	end := findMatchingBrace(text, start)
	if end == start {
		return nil
	}

	jsonStr := text[start:end]

	var wrapper struct {
		ToolCalls []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
		return nil
	}

	var result []ToolCall
	for _, tc := range wrapper.ToolCalls {
		var args map[string]any
		json.Unmarshal([]byte(tc.Function.Arguments), &args)

		result = append(result, ToolCall{
			ID:        tc.ID,
			Type:      tc.Type,
			Name:      tc.Function.Name,
			Arguments: args,
			Function: &FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}

	return result
}

// stripToolCallsFromText removes tool call JSON from response text.
func stripToolCallsFromText(text string) string {
	start := strings.Index(text, `{"tool_calls"`)
	if start != -1 {
		end := findMatchingBrace(text, start)
		if end != start {
			text = strings.TrimSpace(text[:start] + text[end:])
		}
	}
	return sanitizeProviderText(text)
}

func sanitizeProviderText(text string) string {
	cleaned := text
	for _, pattern := range internalExecutionPatterns {
		cleaned = pattern.ReplaceAllString(cleaned, "")
	}
	lines := strings.Split(cleaned, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[/tool_id]") || strings.HasPrefix(trimmed, "</tool_id>") {
			continue
		}
		if trimmed == "" {
			continue
		}
		kept = append(kept, strings.TrimRight(line, " \t"))
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}
