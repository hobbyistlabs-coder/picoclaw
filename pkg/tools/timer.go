package tools

import (
	"context"
	"fmt"
	"time"
)

// TimerTool allows the agent to start a background timer.
// When the timer expires, the agent loop will be notified asynchronously.
type TimerTool struct{}

// NewTimerTool creates a new instance of TimerTool.
func NewTimerTool() *TimerTool {
	return &TimerTool{}
}

// Name returns the name of the tool.
func (t *TimerTool) Name() string {
	return "background_timer"
}

// Description returns the description of the tool.
func (t *TimerTool) Description() string {
	return "Starts a background timer that waits for a specified duration in seconds. The agent can continue with other tasks, and will be notified when the timer expires."
}

// Parameters returns the schema for the tool's parameters.
func (t *TimerTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"duration_seconds": map[string]any{
				"type":        "integer",
				"description": "The number of seconds to wait before notifying the agent.",
			},
			"message": map[string]any{
				"type":        "string",
				"description": "An optional message or context to be sent back when the timer expires.",
			},
		},
		"required": []string{"duration_seconds"},
	}
}

// Execute is a fallback for synchronous execution, but this tool should ideally use ExecuteAsync.
func (t *TimerTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	return t.ExecuteAsync(ctx, args, nil)
}

// ExecuteAsync implements the AsyncExecutor interface.
func (t *TimerTool) ExecuteAsync(ctx context.Context, args map[string]any, cb AsyncCallback) *ToolResult {
	durationFloat, ok := args["duration_seconds"].(float64)
	if !ok {
		return ErrorResult("duration_seconds must be a valid number").WithError(fmt.Errorf("duration_seconds must be a valid number"))
	}

	duration := time.Duration(durationFloat) * time.Second

	msg, _ := args["message"].(string)
	if msg == "" {
		msg = fmt.Sprintf("Timer for %d seconds completed.", int(durationFloat))
	} else {
		msg = fmt.Sprintf("Timer for %d seconds completed. Message: %s", int(durationFloat), msg)
	}

	// Start a background goroutine for the timer
	go func() {
		// Wait for the duration or context cancellation
		select {
		case <-time.After(duration):
			if cb != nil {
				// We create a new background context for the callback because the original tool call ctx might be canceled
				// Use NewToolResult for system result and format the message appropriately.
				cb(context.Background(), NewToolResult(fmt.Sprintf("Task 'background_timer' completed.\n\nResult:\n%s", msg)))
			}
		case <-ctx.Done():
			if cb != nil {
				cb(context.Background(), ErrorResult(fmt.Sprintf("Timer cancelled: %v", ctx.Err())).WithError(ctx.Err()))
			}
		}
	}()

	return AsyncResult(fmt.Sprintf("Timer started for %d seconds in the background. You will be notified when it completes.", int(durationFloat)))
}

// RequiresApproval returns false as this is a safe operation.
func (t *TimerTool) RequiresApproval() bool {
	return false
}
