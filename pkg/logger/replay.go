package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
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

var sessionLocks sync.Map

// LogSessionEvent appends a JSONL entry to the session replay log file.
func LogSessionEvent(workspacePath, sessionID string, event SessionEvent) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	// Sanitize sessionID to prevent path traversal, keeping composite keys with slashes flat.
	cleanSessionID := strings.ReplaceAll(sessionID, "/", "_")
	cleanSessionID = strings.ReplaceAll(cleanSessionID, "\\", "_")
	if cleanSessionID == "." || cleanSessionID == ".." {
		return
	}

	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if event.SessionID == "" {
		event.SessionID = sessionID
	}
	if event.ErrorCategory == "" {
		event.ErrorCategory = ReplayErrorCategoryNone
	}

	data, err := json.Marshal(event)
	if err != nil {
		ErrorC("replay", fmt.Sprintf("Failed to marshal session event: %v", err))
		return
	}
	data = append(data, '\n')

	lockInterface, _ := sessionLocks.LoadOrStore(cleanSessionID, &sync.Mutex{})
	mu := lockInterface.(*sync.Mutex)

	logDir := filepath.Join(workspacePath, "logs", cleanSessionID, "events")

	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		ErrorC("replay", fmt.Sprintf("Failed to create replay log directory: %v", err))
		return
	}

	logFile := filepath.Join(logDir, "events.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ErrorC("replay", fmt.Sprintf("Failed to open replay log file: %v", err))
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		ErrorC("replay", fmt.Sprintf("Failed to write to replay log file: %v", err))
	}
}

// CleanupSessionLocks removes the session lock from the map to prevent memory leaks.
// This should be called when a session is finalized or closed.
func CleanupSessionLocks(sessionID string) {
	cleanSessionID := strings.ReplaceAll(sessionID, "/", "_")
	cleanSessionID = strings.ReplaceAll(cleanSessionID, "\\", "_")
	sessionLocks.Delete(cleanSessionID)
}
