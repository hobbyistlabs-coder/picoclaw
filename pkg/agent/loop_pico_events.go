package agent

import (
	"context"
	"encoding/json"
	"strings"

	"jane/pkg/bus"
	"jane/pkg/providers"
	"jane/pkg/tools"
	"jane/pkg/utils"
)

func buildToolEvent(tc providers.ToolCall, status string, result *tools.ToolResult, ms int64) *bus.ToolCallEvent {
	event := &bus.ToolCallEvent{
		ID:        tc.ID,
		Name:      tc.Name,
		Kind:      toolKind(tc.Name),
		Status:    status,
		Label:     toolLabel(tc),
		Arguments: tc.Arguments,
	}
	if result == nil {
		return event
	}
	event.DurationMS = ms
	event.Result = utils.Truncate(toolResultText(result), 240)
	if result.Async {
		event.Summary = "Running in background"
	}
	if result.IsError {
		event.Summary = "Execution failed"
		return event
	}
	if event.Summary == "" {
		event.Summary = "Completed"
	}
	return event
}

func buildSubagentEvent(
	tc providers.ToolCall,
	task *tools.SubagentTask,
	event *tools.SubagentProgressEvent,
) *bus.ToolCallEvent {
	if task == nil || event == nil {
		return nil
	}
	return &bus.ToolCallEvent{
		ID:              tc.ID,
		Name:            tc.Name,
		Kind:            "subagent",
		Status:          task.Status,
		Label:           task.Label,
		Summary:         event.Message,
		Result:          task.LastOutputExcerpt,
		EventType:       event.EventType,
		TaskID:          task.ID,
		Codename:        task.Codename,
		ParentSessionID: task.ParentSessionID,
		LatestEvent:     task.LatestEvent,
		ProgressPercent: task.ProgressPercent,
		Error:           task.Error,
		ToolName:        task.ToolName,
		ToolStatus:      task.ToolStatus,
	}
}

func publishToolEvent(ctx context.Context, al *AgentLoop, opts processOptions, event *bus.ToolCallEvent) {
	if event == nil || opts.Channel != "pico" {
		return
	}
	_ = al.bus.PublishOutbound(ctx, bus.OutboundMessage{
		Channel:   opts.Channel,
		ChatID:    opts.ChatID,
		ToolEvent: event,
	})
}

func toolKind(name string) string {
	switch {
	case name == "spawn" || name == "subagent":
		return "subagent"
	case name == "mcp2cli" || strings.HasPrefix(name, "mcp_"):
		return "mcp"
	default:
		return "tool"
	}
}

func toolLabel(tc providers.ToolCall) string {
	if toolKind(tc.Name) == "mcp" && strings.HasPrefix(tc.Name, "mcp_") {
		return strings.TrimPrefix(tc.Name, "mcp_")
	}
	task := toolTask(tc.Arguments)
	if task == "" {
		return tc.Name
	}
	return task
}

func toolTask(args map[string]any) string {
	raw, _ := args["label"].(string)
	if strings.TrimSpace(raw) != "" {
		return raw
	}
	raw, _ = args["task"].(string)
	return utils.Truncate(strings.TrimSpace(raw), 80)
}

func toolResultText(result *tools.ToolResult) string {
	if result == nil {
		return ""
	}
	if result.ForUser != "" {
		return result.ForUser
	}
	if result.ForLLM != "" {
		return result.ForLLM
	}
	if result.Err != nil {
		return result.Err.Error()
	}
	out, _ := json.Marshal(result)
	return string(out)
}
