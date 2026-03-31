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
	ReplayErrorModel          ReplayErrorCategory = "model_failure"
	ReplayErrorInfrastructure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorLogic          ReplayErrorCategory = "logic_failure"
	ReplayErrorNone           ReplayErrorCategory = "none"
)

type ReplayEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // cot, tool_call, tool_result, state_transition, error
	Details       map[string]any      `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var sessionLocks sync.Map

func getSessionLock(sessionID string) *sync.Mutex {
	lock, _ := sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func LogSessionEvent(
	workspacePath string,
	sessionID string,
	eventType string,
	details map[string]any,
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

	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	data = append(data, '\n')

	dirPath := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return
	}

	filePath := filepath.Join(dirPath, "events.jsonl")

	lock := getSessionLock(sessionID)
	lock.Lock()
	defer lock.Unlock()

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(data)
}
