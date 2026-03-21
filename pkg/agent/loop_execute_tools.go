package agent

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
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
) []indexedAgentResult {
	agentResults := make([]indexedAgentResult, len(normalizedToolCalls))
	var wg sync.WaitGroup

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
					logger.LogSessionEvent(
						agent.Workspace,
						opts.SessionKey,
						logger.EventTypeError,
						logger.EventDetails{
							ToolName: tc.Name,
						},
						logger.ErrorCategoryLogicFailure,
						errStr,
					)
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

			var inputsMap map[string]any
			if tc.Arguments != nil {
				inputsMap = tc.Arguments
			} else {
				inputsMap = make(map[string]any)
			}

			logger.LogSessionEvent(
				agent.Workspace,
				opts.SessionKey,
				logger.EventTypeToolCall,
				logger.EventDetails{
					ToolName: tc.Name,
					Inputs:   inputsMap,
				},
				"none",
				"",
			)

			// Create async callback for tools that implement AsyncExecutor.
			// When the background work completes, this publishes the result
			// as an inbound system message so processSystemMessage routes it
			// back to the user via the normal agent loop.
			asyncCallback := func(_ context.Context, result *tools.ToolResult) {
				// Send ForUser content directly to the user (immediate feedback),
				// mirroring the synchronous tool execution path.
				if !result.Silent && result.ForUser != "" {
					outCtx, outCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer outCancel()
					_ = al.bus.PublishOutbound(outCtx, bus.OutboundMessage{
						Channel: opts.Channel,
						ChatID:  opts.ChatID,
						Content: result.ForUser,
					})
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
					Channel:  "system",
					SenderID: fmt.Sprintf("async:%s", tc.Name),
					ChatID:   fmt.Sprintf("%s:%s", opts.Channel, opts.ChatID),
					Content:  content,
				})
			}

			startToolTime := time.Now()
			toolResult := agent.Tools.ExecuteWithContext(
				ctx,
				tc.Name,
				tc.Arguments,
				opts.Channel,
				opts.ChatID,
				asyncCallback,
			)
			metricsToolExecutionDuration.Add(time.Since(startToolTime).Seconds())
			agentResults[idx].result = toolResult

			// Log tool execution result
			if toolResult.Err != nil {
				logger.LogSessionEvent(
					agent.Workspace,
					opts.SessionKey,
					logger.EventTypeError,
					logger.EventDetails{
						ToolName: tc.Name,
					},
					logger.ErrorCategoryLogicFailure,
					toolResult.Err.Error(),
				)
			} else {
				logger.LogSessionEvent(
					agent.Workspace,
					opts.SessionKey,
					logger.EventTypeToolResult,
					logger.EventDetails{
						ToolName: tc.Name,
						Outputs: map[string]any{
							"for_llm":  toolResult.ForLLM,
							"for_user": toolResult.ForUser,
							"silent":   toolResult.Silent,
						},
					},
					"none",
					"",
				)
			}
		}(i, tc)
	}
	wg.Wait()

	return agentResults
}
