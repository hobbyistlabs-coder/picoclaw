package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ReplayEventType string

const (
	EventCoT             ReplayEventType = "cot"
	EventToolCall        ReplayEventType = "tool_call"
	EventToolResult      ReplayEventType = "tool_result"
	EventStateTransition ReplayEventType = "state_transition"
	EventError           ReplayEventType = "error"
)

type ReplayErrorCategory string

const (
	CategoryModelFailure          ReplayErrorCategory = "model_failure"
	CategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	CategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	CategoryNone                  ReplayErrorCategory = "none"
)

type ReplayEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type ReplayEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     ReplayEventType     `json:"event_type"`
	Details       ReplayEventDetails  `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var replayMutex sync.Mutex

// LogSessionEvent appends a structured JSON event to the session's replay log.
func LogSessionEvent(workspacePath, sessionID string, eventType ReplayEventType, details ReplayEventDetails, errCat ReplayErrorCategory, errMsg string) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	event := ReplayEvent{
		Timestamp:     time.Now().UTC(),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errCat,
		ErrorMessage:  errMsg,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	data = append(data, '\n')

	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")

	replayMutex.Lock()
	defer replayMutex.Unlock()

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return
	}

	logFile := filepath.Join(logDir, "events.jsonl")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return
	}

	return
}
