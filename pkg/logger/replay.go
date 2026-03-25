package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReplayErrorCategory defines the category of an error for session replay logs.
type ReplayErrorCategory string

const (
	ReplayErrorModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorNone                  ReplayErrorCategory = "none"
)

// ReplayEventType defines the type of event for session replay logs.
type ReplayEventType string

const (
	ReplayEventCoT             ReplayEventType = "cot"
	ReplayEventToolCall        ReplayEventType = "tool_call"
	ReplayEventToolResult      ReplayEventType = "tool_result"
	ReplayEventStateTransition ReplayEventType = "state_transition"
	ReplayEventError           ReplayEventType = "error"
)

// ReplayEventDetails contains specific details for different event types.
type ReplayEventDetails struct {
	CoTText   string `json:"cot_text,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	Inputs    any    `json:"inputs,omitempty"`
	Outputs   any    `json:"outputs,omitempty"`
	FromState string `json:"from_state,omitempty"`
	ToState   string `json:"to_state,omitempty"`
}

// ReplayEvent represents a single event in the session replay log.
type ReplayEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     ReplayEventType     `json:"event_type"`
	Details       ReplayEventDetails  `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var (
	replayMutex sync.Mutex
)

// LogSessionEvent appends a structured JSON event to the session's replay log.
func LogSessionEvent(
	workspacePath string,
	sessionID string,
	eventType ReplayEventType,
	details ReplayEventDetails,
	errorCategory ReplayErrorCategory,
	errorMessage string,
) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	event := ReplayEvent{
		Timestamp:     time.Now().UTC(),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errorCategory,
		ErrorMessage:  errorMessage,
	}

	// Default empty error category to "none" if not provided
	if event.ErrorCategory == "" {
		event.ErrorCategory = ReplayErrorNone
	}

	data, err := json.Marshal(event)
	if err != nil {
		WarnCF("replay", "Failed to marshal replay event", map[string]any{"error": err.Error()})
		return
	}

	// Append a newline for JSONL format
	data = append(data, '\n')

	// Construct file path: {workspacePath}/logs/{session_id}/events/events.jsonl
	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")

	// Ensure thread-safe file operations
	replayMutex.Lock()
	defer replayMutex.Unlock()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		WarnCF("replay", "Failed to create replay log directory", map[string]any{"error": err.Error(), "path": logDir})
		return
	}

	logFile := filepath.Join(logDir, "events.jsonl")

	// Open file in append mode, create if not exists
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		WarnCF("replay", "Failed to open replay log file", map[string]any{"error": err.Error(), "path": logFile})
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		WarnCF("replay", "Failed to write replay event", map[string]any{"error": err.Error()})
	}
}
