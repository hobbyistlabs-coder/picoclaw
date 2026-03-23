package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ReplayEventCategory represents the type of event in the session replay log.
type ReplayEventCategory string

const (
	EventCategoryCoT             ReplayEventCategory = "cot"
	EventCategoryToolCall        ReplayEventCategory = "tool_call"
	EventCategoryToolResult      ReplayEventCategory = "tool_result"
	EventCategoryStateTransition ReplayEventCategory = "state_transition"
	EventCategoryError           ReplayEventCategory = "error"
)

// ReplayErrorCategory maps to ErrorCategory constants but is specific to the JSON schema.
type ReplayErrorCategory string

const (
	ReplayErrorModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorNone                  ReplayErrorCategory = "none"
)

// ReplayEventDetails contains the specific data for an event.
type ReplayEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

// ReplayEvent represents a single event in the session replay.
type ReplayEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     ReplayEventCategory `json:"event_type"`
	Details       ReplayEventDetails  `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

// LogSessionEvent appends a structured JSON event to the session's event log file.
func LogSessionEvent(workspacePath string, event ReplayEvent) error {
	if workspacePath == "" || event.SessionID == "" {
		return fmt.Errorf("workspacePath and SessionID are required")
	}

	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	if event.ErrorCategory == "" {
		event.ErrorCategory = ReplayErrorNone
	}

	// Define the directory and file path
	dirPath := filepath.Join(workspacePath, "logs", event.SessionID, "events")
	err := os.MkdirAll(dirPath, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", dirPath, err)
	}

	filePath := filepath.Join(dirPath, "events.jsonl")

	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}
	defer file.Close()

	// Serialize the event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Write the JSON followed by a newline (JSONL format)
	_, err = file.Write(append(eventBytes, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write to log file %s: %w", filePath, err)
	}

	return nil
}
