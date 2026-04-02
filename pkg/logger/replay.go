package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReplayEventType represents the type of an event in the session replay log
type ReplayEventType string

const (
	EventTypeCoT             ReplayEventType = "cot"
	EventTypeToolCall        ReplayEventType = "tool_call"
	EventTypeToolResult      ReplayEventType = "tool_result"
	EventTypeStateTransition ReplayEventType = "state_transition"
	EventTypeError           ReplayEventType = "error"
)

// ReplayErrorCategory represents the category of an error in the session replay log
type ReplayErrorCategory string

const (
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
)

// ReplayEventDetails holds specific details based on the event type
type ReplayEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

// ReplayEvent represents a single event in the session replay log
type ReplayEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     ReplayEventType     `json:"event_type"`
	Details       ReplayEventDetails  `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var sessionLocks sync.Map

func getSessionLock(sessionID string) *sync.Mutex {
	lock, _ := sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// CleanupSessionLocks removes the lock for a given session ID
func CleanupSessionLocks(sessionID string) {
	sessionLocks.Delete(sessionID)
}

// LogSessionEvent logs an event to the session replay log file
func LogSessionEvent(workspacePath string, sessionID string, event ReplayEvent) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	lock := getSessionLock(sessionID)
	lock.Lock()
	defer lock.Unlock()

	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return
	}

	logFile := filepath.Join(logDir, "events.jsonl")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	_, _ = f.Write(data)
	_, _ = f.Write([]byte("\n"))
}
