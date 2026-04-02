package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReplayErrorCategory defines the category of error for session replay logs.
type ReplayErrorCategory string

const (
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
)

// ReplayEvent represents a single event in the session replay log.
type ReplayEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // "cot", "tool_call", "tool_result", "state_transition", "error"
	Details       map[string]any      `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var sessionLocks sync.Map

// getSessionLock retrieves or creates a mutex for the given session ID.
func getSessionLock(sessionID string) *sync.Mutex {
	lock, _ := sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// CleanupSessionLocks removes the lock for a given session ID to prevent memory leaks.
func CleanupSessionLocks(sessionID string) {
	sessionLocks.Delete(sessionID)
}

// LogSessionEvent appends a structured JSON event to the session's event.jsonl file.
func LogSessionEvent(workspacePath, sessionID, eventType string, details map[string]any, errorCategory ReplayErrorCategory, errorMsg string) {
	if workspacePath == "" || sessionID == "" || eventType == "" {
		return
	}

	event := ReplayEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errorCategory,
		ErrorMessage:  errorMsg,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return // Best effort
	}

	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")

	lock := getSessionLock(sessionID)
	lock.Lock()
	defer lock.Unlock()

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return // Best effort
	}

	// We'll just write to a single events.jsonl file for the session for simplicity,
	// rather than rotating by timestamp as suggested by {timestamp}_event.json,
	// since JSONL is better for a continuous stream.
	logFile := filepath.Join(logDir, "events.jsonl")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return // Best effort
	}
	defer f.Close()

	_, _ = f.Write(append(eventData, '\n'))
}
