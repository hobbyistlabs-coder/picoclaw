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
	ReplayErrorModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorNone                  ReplayErrorCategory = "none"
)

type ReplayEventType string

const (
	ReplayEventCoT             ReplayEventType = "cot"
	ReplayEventToolCall        ReplayEventType = "tool_call"
	ReplayEventToolResult      ReplayEventType = "tool_result"
	ReplayEventStateTransition ReplayEventType = "state_transition"
	ReplayEventError           ReplayEventType = "error"
)

type SessionEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type SessionEvent struct {
	Timestamp     time.Time            `json:"timestamp"`
	SessionID     string               `json:"session_id"`
	EventType     ReplayEventType      `json:"event_type"`
	Details       *SessionEventDetails `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory  `json:"error_category,omitempty"`
	ErrorMessage  string               `json:"error_message,omitempty"`
}

var sessionLocks sync.Map

func getSessionLock(sessionID string) *sync.Mutex {
	lock, _ := sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func LogSessionEvent(workspacePath string, event SessionEvent) {
	if workspacePath == "" || event.SessionID == "" {
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if event.ErrorCategory == "" {
		event.ErrorCategory = ReplayErrorNone
	}

	lock := getSessionLock(event.SessionID)
	lock.Lock()
	defer lock.Unlock()

	logDir := filepath.Join(workspacePath, "logs", event.SessionID, "events")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return
	}

	logFile := filepath.Join(logDir, "events.jsonl")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	_, _ = file.Write(append(data, '\n'))
}
