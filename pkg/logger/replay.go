package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ReplayErrorCategory string

const (
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
)

type ReplayEventDetails struct {
	CotText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type ReplayEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // "cot", "tool_call", "tool_result", "state_transition", "error"
	Details       ReplayEventDetails  `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var sessionLocks sync.Map

// LogSessionEvent logs a structured event for session replay.
// It is thread-safe and writes to a JSONL file per session.
// Formatting or I/O errors are handled internally on a best-effort basis.
func LogSessionEvent(
	workspacePath string,
	sessionID string,
	eventType string,
	details ReplayEventDetails,
	errorCategory ReplayErrorCategory,
	errorMessage string,
) {
	// Ensure errorCategory defaults to "none" if empty, based on the schema
	if errorCategory == "" {
		errorCategory = ReplayErrorCategoryNone
	}

	event := ReplayEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errorCategory,
		ErrorMessage:  errorMessage,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		ErrorCF("logger", "Failed to marshal session replay event", map[string]any{"error": err.Error()})
		return
	}
	eventBytes = append(eventBytes, '\n')

	// Get or create a mutex for this session
	val, _ := sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	mu := val.(*sync.Mutex)

	mu.Lock()
	defer mu.Unlock()

	// Use filepath.Clean to sanitize sessionID to prevent directory traversal
	safeSessionID := filepath.Base(filepath.Clean(sessionID))

	// Ensure directory exists
	eventsDir := filepath.Join(workspacePath, "logs", safeSessionID, "events")
	if mkdirErr := os.MkdirAll(eventsDir, 0o755); mkdirErr != nil {
		ErrorCF("logger", "Failed to create session replay directory", map[string]any{
			"error": mkdirErr.Error(),
			"dir":   eventsDir,
		})
		return
	}

	eventsFile := filepath.Join(eventsDir, "events.jsonl")
	f, openErr := os.OpenFile(eventsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if openErr != nil {
		ErrorCF("logger", "Failed to open session replay file", map[string]any{
			"error": openErr.Error(),
			"file":  eventsFile,
		})
		return
	}
	defer f.Close()

	if _, writeErr := f.Write(eventBytes); writeErr != nil {
		ErrorCF("logger", "Failed to write session replay event", map[string]any{
			"error": writeErr.Error(),
			"file":  eventsFile,
		})
	}
}
