package logger

import (
	"encoding/json"
	"fmt"
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
	ErrorCategoryModel          ReplayErrorCategory = "model_failure"
	ErrorCategoryInfrastructure ReplayErrorCategory = "infrastructure_failure"
	ErrorCategoryLogic          ReplayErrorCategory = "logic_failure"
	ErrorCategoryNone           ReplayErrorCategory = "none"
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
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     ReplayEventType     `json:"event_type"`
	Details       ReplayEventDetails  `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var (
	replayMutex sync.Mutex
)

// LogSessionEvent appends a session replay event to the structured JSONL log file.
// Path: {workspacePath}/logs/{sessionID}/events/events.jsonl
func LogSessionEvent(workspacePath, sessionID string, event ReplayEvent) error {
	if workspacePath == "" || sessionID == "" {
		return nil // Missing context, skip logging
	}

	event.Timestamp = time.Now().UTC().Format(time.RFC3339)

	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create replay log dir: %w", err)
	}

	logFile := filepath.Join(logDir, "events.jsonl")

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal replay event: %w", err)
	}

	data = append(data, '\n')

	replayMutex.Lock()
	defer replayMutex.Unlock()

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open replay log file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write replay event: %w", err)
	}

	return nil
}
