package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReplayErrorCategory categorizes errors for session replay logging.
type ReplayErrorCategory string

const (
	ErrorCategoryModel          ReplayErrorCategory = "model_failure"
	ErrorCategoryInfrastructure ReplayErrorCategory = "infrastructure_failure"
	ErrorCategoryLogic          ReplayErrorCategory = "logic_failure"
	ErrorCategoryNone           ReplayErrorCategory = "none"
)

// EventType represents the type of an event in the session replay log.
type EventType string

const (
	EventTypeCoT             EventType = "cot"
	EventTypeToolCall        EventType = "tool_call"
	EventTypeToolResult      EventType = "tool_result"
	EventTypeStateTransition EventType = "state_transition"
	EventTypeError           EventType = "error"
)

// EventDetails holds the context-specific details of a replay event.
type EventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

// ReplayEvent represents a single event in the session replay log.
type ReplayEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     EventType           `json:"event_type"`
	Details       EventDetails        `json:"details"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var sessionLocks sync.Map

// getSessionLock retrieves or creates a mutex for a specific session ID.
func getSessionLock(sessionID string) *sync.Mutex {
	lock, _ := sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// CleanupSessionLock removes the lock for a specific session ID to prevent memory leaks.
// This should be called when a session is explicitly deleted or closed.
func CleanupSessionLock(sessionID string) {
	sessionLocks.Delete(sessionID)
}

// LogSessionEvent appends a ReplayEvent to the session's log file.
// It uses a per-session lock to prevent write conflicts and logs errors internally.
func LogSessionEvent(workspacePath string, event ReplayEvent) {
	if workspacePath == "" || event.SessionID == "" {
		return
	}

	lock := getSessionLock(event.SessionID)
	lock.Lock()
	defer lock.Unlock()

	logDir := filepath.Join(workspacePath, "logs", event.SessionID, "events")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		WarnCF("logger", "Failed to create session replay log directory", map[string]any{
			"error": err.Error(),
			"dir":   logDir,
		})
		return
	}

	logFile := filepath.Join(logDir, "events.jsonl")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		WarnCF("logger", "Failed to open session replay log file", map[string]any{
			"error": err.Error(),
			"file":  logFile,
		})
		return
	}
	defer file.Close()

	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format(time.RFC3339)
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(event); err != nil {
		WarnCF("logger", "Failed to encode session replay event", map[string]any{
			"error": err.Error(),
		})
	}
}
