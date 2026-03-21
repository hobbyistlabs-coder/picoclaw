package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type EventType string

const (
	EventTypeCoT             EventType = "cot"
	EventTypeToolCall        EventType = "tool_call"
	EventTypeToolResult      EventType = "tool_result"
	EventTypeStateTransition EventType = "state_transition"
	EventTypeError           EventType = "error"
)

type ReplayErrorCategory string

const (
	ReplayErrorModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorNone                  ReplayErrorCategory = "none"
)

type EventDetails struct {
	CoTText   string `json:"cot_text,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	Inputs    any    `json:"inputs,omitempty"`
	Outputs   any    `json:"outputs,omitempty"`
	FromState string `json:"from_state,omitempty"`
	ToState   string `json:"to_state,omitempty"`
}

type SessionEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     EventType           `json:"event_type"`
	Details       *EventDetails       `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

// LogSessionEvent writes a structured JSON event to {workspacePath}/logs/{session_id}/events/{timestamp}_event.json
func LogSessionEvent(workspacePath string, event SessionEvent) error {
	if workspacePath == "" || event.SessionID == "" {
		return fmt.Errorf("workspacePath and session_id are required")
	}

	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	eventsDir := filepath.Join(workspacePath, "logs", event.SessionID, "events")

	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return fmt.Errorf("failed to create events directory: %w", err)
	}

	// Use UnixNano for uniqueness in filename
	filename := fmt.Sprintf("%d_event.json", time.Now().UnixNano())
	file_path := filepath.Join(eventsDir, filename)

	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := os.WriteFile(file_path, data, 0644); err != nil {
		return fmt.Errorf("failed to write event file: %w", err)
	}

	return nil
}
