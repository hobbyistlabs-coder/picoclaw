package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (p *AntigravityProvider) parseSSEResponse(body string) (*LLMResponse, error) {
	var contentParts []string
	var toolCalls []ToolCall
	var usage *UsageInfo
	var finishReason string

	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		// v1internal SSE wraps the Gemini response in a "response" field
		var sseChunk struct {
			Response antigravityJSONResponse `json:"response"`
		}
		if err := json.Unmarshal([]byte(data), &sseChunk); err != nil {
			continue
		}
		resp := sseChunk.Response

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					contentParts = append(contentParts, part.Text)
				}
				if part.FunctionCall != nil {
					argumentsJSON, _ := json.Marshal(part.FunctionCall.Args)
					toolCalls = append(toolCalls, ToolCall{
						ID:        fmt.Sprintf("call_%s_%d", part.FunctionCall.Name, time.Now().UnixNano()),
						Name:      part.FunctionCall.Name,
						Arguments: part.FunctionCall.Args,
						Function: &FunctionCall{
							Name:      part.FunctionCall.Name,
							Arguments: string(argumentsJSON),
							ThoughtSignature: extractPartThoughtSignature(
								part.ThoughtSignature,
								part.ThoughtSignatureSnake,
							),
						},
					})
				}
			}
			if candidate.FinishReason != "" {
				finishReason = candidate.FinishReason
			}
		}

		if resp.UsageMetadata.TotalTokenCount > 0 {
			usage = &UsageInfo{
				PromptTokens:     resp.UsageMetadata.PromptTokenCount,
				CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      resp.UsageMetadata.TotalTokenCount,
			}
		}
	}

	mappedFinish := "stop"
	if len(toolCalls) > 0 {
		mappedFinish = "tool_calls"
	}
	if finishReason == "MAX_TOKENS" {
		mappedFinish = "length"
	}

	return &LLMResponse{
		Content:      strings.Join(contentParts, ""),
		ToolCalls:    toolCalls,
		FinishReason: mappedFinish,
		Usage:        usage,
	}, nil
}

func extractPartThoughtSignature(thoughtSignature string, thoughtSignatureSnake string) string {
	if thoughtSignature != "" {
		return thoughtSignature
	}
	if thoughtSignatureSnake != "" {
		return thoughtSignatureSnake
	}
	return ""
}

func (p *AntigravityProvider) parseAntigravityError(statusCode int, body []byte) error {
	var errResp struct {
		Error struct {
			Code    int              `json:"code"`
			Message string           `json:"message"`
			Status  string           `json:"status"`
			Details []map[string]any `json:"details"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("antigravity API error (HTTP %d): %s", statusCode, truncateString(string(body), 500))
	}

	msg := errResp.Error.Message
	if statusCode == 429 {
		// Try to extract quota reset info
		for _, detail := range errResp.Error.Details {
			if typeVal, ok := detail["@type"].(string); ok && strings.HasSuffix(typeVal, "ErrorInfo") {
				if metadata, ok := detail["metadata"].(map[string]any); ok {
					if delay, ok := metadata["quotaResetDelay"].(string); ok {
						return fmt.Errorf("antigravity rate limit exceeded: %s (reset in %s)", msg, delay)
					}
				}
			}
		}
		return fmt.Errorf("antigravity rate limit exceeded: %s", msg)
	}

	return fmt.Errorf("antigravity API error (%s): %s", errResp.Error.Status, msg)
}
