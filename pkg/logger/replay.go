package logger

import (
	"encoding/json"
	"fmt"
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

type ReplayEventType string

const (
	ReplayEventTypeCoT             ReplayEventType = "cot"
	ReplayEventTypeToolCall        ReplayEventType = "tool_call"
	ReplayEventTypeToolResult      ReplayEventType = "tool_result"
	ReplayEventTypeStateTransition ReplayEventType = "state_transition"
	ReplayEventTypeError           ReplayEventType = "error"
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
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     ReplayEventType     `json:"event_type"`
	Details       SessionEventDetails `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var (
	replayLogMu sync.Mutex
)

// LogSessionEvent thread-safely appends a JSON event to {workspacePath}/logs/{session_id}/events/events.jsonl
func LogSessionEvent(workspacePath, sessionID string, event SessionEvent) error {
	if workspacePath == "" || sessionID == "" {
		return fmt.Errorf("workspacePath and sessionID must be provided")
	}

	// Ensure timestamp and session ID are set correctly
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	event.SessionID = sessionID

	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")

	replayLogMu.Lock()
	defer replayLogMu.Unlock()

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create replay log directory: %w", err)
	}

	logFile := filepath.Join(logDir, "events.jsonl")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open replay log file: %w", err)
	}
	defer f.Close()

	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal replay event: %w", err)
	}

	if _, err := f.Write(append(b, '\n')); err != nil {
		return fmt.Errorf("failed to write replay event: %w", err)
	}

	return nil
}
