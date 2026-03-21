package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EventType represents the type of a session event.
type EventType string

const (
	EventTypeCoT             EventType = "cot"
	EventTypeToolCall        EventType = "tool_call"
	EventTypeToolResult      EventType = "tool_result"
	EventTypeStateTransition EventType = "state_transition"
	EventTypeError           EventType = "error"
)

// EventDetails holds the specific details for an event.
type EventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

// SessionEvent represents a structured log event for session replay.
type SessionEvent struct {
	Timestamp     time.Time     `json:"timestamp"`
	SessionID     string        `json:"session_id"`
	EventType     EventType     `json:"event_type"`
	Details       EventDetails  `json:"details,omitempty"`
	ErrorCategory ErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string        `json:"error_message,omitempty"`
}

// LogSessionEvent writes a structured JSON event to {workspacePath}/logs/{session_id}/events/{timestamp}_event.json
func LogSessionEvent(
	workspacePath string,
	sessionID string,
	eventType EventType,
	details EventDetails,
	errorCategory ErrorCategory,
	errorMessage string,
) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	event := SessionEvent{
		Timestamp:     time.Now().UTC(),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errorCategory,
		ErrorMessage:  errorMessage,
	}

	// Create directory structure: {workspacePath}/logs/{session_id}/events/
	eventsDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		ErrorCF("logger", "Failed to create session events directory", map[string]any{
			"path":  eventsDir,
			"error": err.Error(),
		})
		return
	}

	// Filename: {timestamp}_event.json
	// Using format that avoids colons in filename for Windows compatibility
	filename := fmt.Sprintf("%s_event.json", event.Timestamp.Format("20060102_150405_.000000"))
	filePath := filepath.Join(eventsDir, filename)

	data, err := json.Marshal(event)
	if err != nil {
		ErrorCF("logger", "Failed to marshal session event", map[string]any{
			"error": err.Error(),
		})
		return
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		ErrorCF("logger", "Failed to write session event file", map[string]any{
			"path":  filePath,
			"error": err.Error(),
		})
	}
}
