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
	ReplayModelFailure          ReplayErrorCategory = "model_failure"
	ReplayInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayNone                  ReplayErrorCategory = "none"
)

type SessionEventDetails struct {
	CotText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    any            `json:"inputs,omitempty"`
	Outputs   any            `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type SessionEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // "cot", "tool_call", "tool_result", "state_transition", "error"
	Details       SessionEventDetails `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var sessionLocks sync.Map

func getSessionLock(sessionID string) *sync.Mutex {
	lock, _ := sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func CleanupSessionLocks(sessionID string) {
	sessionLocks.Delete(sessionID)
}

func LogSessionEvent(workspacePath string, event SessionEvent) {
	if workspacePath == "" || event.SessionID == "" {
		return
	}

	lock := getSessionLock(event.SessionID)
	lock.Lock()
	defer lock.Unlock()

	dir := filepath.Join(workspacePath, "logs", event.SessionID, "events")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	filePath := filepath.Join(dir, "events.jsonl")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format(time.RFC3339)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	file.Write(data)
	file.WriteString("\n")
}
