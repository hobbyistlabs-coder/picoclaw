package agent

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"strconv"
	"sync"
	"time"

	"jane/pkg/bus"
	"jane/pkg/logger"
	"jane/pkg/providers"
	"jane/pkg/tools"
	"jane/pkg/utils"
)

var (
	metricsToolExecutionDuration = expvar.NewFloat("agentloop_tool_execution_duration_seconds")
)

type indexedAgentResult struct {
	result *tools.ToolResult
	tc     providers.ToolCall
}

// executeToolBatch runs multiple tools concurrently in goroutines and returns their results
// in the same order they were requested.
func (al *AgentLoop) executeToolBatch(
	ctx context.Context,
	agent *AgentInstance,
	opts processOptions,
	normalizedToolCalls []providers.ToolCall,
	iteration int,
) ([]indexedAgentResult, bool) {
	agentResults := make([]indexedAgentResult, len(normalizedToolCalls))
	var wg sync.WaitGroup
	asyncCount := 0
	for _, tc := range normalizedToolCalls {
		tool, ok := agent.Tools.Get(tc.Name)
		if ok {
			if _, ok := tool.(tools.AsyncExecutor); ok {
				asyncCount++
			}
		}
	}
	asyncBatchID := ""
	if asyncCount > 0 {
		asyncBatchID = fmt.Sprintf("%s:%d:%d", opts.SessionKey, iteration, time.Now().UnixNano())
		al.startAsyncBatch(asyncBatchID, asyncCount)
	}

	for i, tc := range normalizedToolCalls {
		agentResults[i].tc = tc

		wg.Add(1)
		go func(idx int, tc providers.ToolCall) {
			defer wg.Done()

			// Panic recovery for robust tool execution
			defer func() {
				if r := recover(); r != nil {
					errStr := fmt.Sprintf("Tool execution panicked: %v", r)
					logger.ErrorCF("agent", "Tool panic recovered", map[string]any{
						"agent_id": agent.ID,
						"tool":     tc.Name,
						"panic":    r,
					})
					agentResults[idx].result = &tools.ToolResult{
						ForLLM: errStr,
						Err:    fmt.Errorf("%s", errStr),
					}
				}
			}()

			argsJSON, _ := json.Marshal(tc.Arguments)
			argsPreview := utils.Truncate(string(argsJSON), 200)
			logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
				map[string]any{
					"agent_id":  agent.ID,
					"tool":      tc.Name,
					"iteration": iteration,
				})

			// Log tool call for observability
			var inputs any
			json.Unmarshal(argsJSON, &inputs) // Unmarshal back to any for clean JSON logging
			logger.LogSessionEvent(
				agent.Workspace,
				opts.SessionKey,
				"tool_call",
				logger.SessionEventDetails{
					ToolName: tc.Name,
					Inputs:   inputs,
				},
				logger.ErrorCategoryNoneReplay,
				"",
			)

			publishToolEvent(ctx, al, opts, buildToolEvent(tc, "started", nil, 0))
			startToolTime := time.Now()

			// Create async callback for tools that implement AsyncExecutor.
			// When the background work completes, this publishes the result
			// as an inbound system message so processSystemMessage routes it
			// back to the user via the normal agent loop.
			asyncCallback := func(_ context.Context, result *tools.ToolResult) {
				if tc.Name != "spawn" {
					publishToolEvent(context.Background(), al, opts,
						buildToolEvent(tc, "completed", result, time.Since(startToolTime).Milliseconds()))
				}

				// Determine content for the agent loop (ForLLM or error).
				content := result.ForLLM
				if content == "" && result.Err != nil {
					content = result.Err.Error()
				}
				if content == "" {
					return
				}

				logger.InfoCF("agent", "Async tool completed, publishing result",
					map[string]any{
						"tool":        tc.Name,
						"content_len": len(content),
						"channel":     opts.Channel,
					})

				pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer pubCancel()
				_ = al.bus.PublishInbound(pubCtx, bus.InboundMessage{
					Channel:    "system",
					SenderID:   fmt.Sprintf("async:%s", tc.Name),
					ChatID:     fmt.Sprintf("%s:%s", opts.Channel, opts.ChatID),
					Content:    content,
					SessionKey: opts.SessionKey,
					Metadata: map[string]string{
						"async_batch_id":       asyncBatchID,
						"async_batch_expected": strconv.Itoa(asyncCount),
						"async_tool_name":      tc.Name,
						"async_tool_call_id":   tc.ID,
					},
				})
			}

			progressCallback := func(task *tools.SubagentTask, event *tools.SubagentProgressEvent) {
				publishToolEvent(context.Background(), al, opts, buildSubagentEvent(tc, task, event))
			}
			toolCtx := tools.WithToolSessionKey(ctx, opts.SessionKey)
			toolCtx = tools.WithToolCallID(toolCtx, tc.ID)
			toolCtx = tools.WithToolAsyncBatchID(toolCtx, asyncBatchID)
			toolCtx = tools.WithSubagentProgress(toolCtx, progressCallback)
			toolResult := agent.Tools.ExecuteWithContext(
				toolCtx,
				tc.Name,
				tc.Arguments,
				opts.Channel,
				opts.ChatID,
				asyncCallback,
			)
			metricsToolExecutionDuration.Add(time.Since(startToolTime).Seconds())
			agentResults[idx].result = toolResult

			// Log tool result for observability
			errorCategory := logger.ErrorCategoryNoneReplay
			errMsg := ""
			if toolResult.IsError || toolResult.Err != nil {
				errorCategory = logger.ErrorCategoryLogicFailureReplay
				if toolResult.Err != nil {
					errMsg = toolResult.Err.Error()
				} else {
					errMsg = toolResult.ForLLM
				}
			}

			logger.LogSessionEvent(
				agent.Workspace,
				opts.SessionKey,
				"tool_result",
				logger.SessionEventDetails{
					ToolName: tc.Name,
					Outputs:  toolResult.ForLLM,
				},
				errorCategory,
				errMsg,
			)

			if !toolResult.Async {
				publishToolEvent(ctx, al, opts,
					buildToolEvent(tc, "completed", toolResult, time.Since(startToolTime).Milliseconds()))
			}
		}(i, tc)
	}
	wg.Wait()

	return agentResults, asyncCount > 0
}
