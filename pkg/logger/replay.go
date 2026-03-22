package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ReplayErrorCategory string

const (
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
)

type EventType string

const (
	EventTypeCOT             EventType = "cot"
	EventTypeToolCall        EventType = "tool_call"
	EventTypeToolResult      EventType = "tool_result"
	EventTypeStateTransition EventType = "state_transition"
	EventTypeError           EventType = "error"
)

type EventDetails struct {
	CotText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type SessionEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     EventType           `json:"event_type"`
	Details       EventDetails        `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

// LogSessionEvent writes a structured JSON event to {workspacePath}/logs/{session_id}/events/
func LogSessionEvent(workspacePath string, event SessionEvent) error {
	if workspacePath == "" {
		workspacePath = "."
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	if event.ErrorCategory == "" {
		event.ErrorCategory = ReplayErrorCategoryNone
	}

	dirPath := filepath.Join(workspacePath, "logs", event.SessionID, "events")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create session log directory: %w", err)
	}

	filePath := filepath.Join(dirPath, "events.jsonl")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open session log file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(event); err != nil {
		return fmt.Errorf("failed to write session event: %w", err)
	}

	return nil
}
