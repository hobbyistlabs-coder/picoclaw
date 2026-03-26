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
	ReplayErrorModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorNone                  ReplayErrorCategory = "none"
)

type SessionEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"`
	Details       SessionEventDetails `json:"details"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

type SessionEventDetails struct {
	CotText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

var replayMu sync.Mutex

func LogSessionEvent(workspacePath, sessionID, eventType string, details SessionEventDetails, errorCategory ReplayErrorCategory, errorMessage string) error {
	event := SessionEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errorCategory,
		ErrorMessage:  errorMessage,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal session event: %w", err)
	}

	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")

	replayMu.Lock()
	defer replayMu.Unlock()

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("failed to create session log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "events.jsonl")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open session log file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(eventData, '\n')); err != nil {
		return fmt.Errorf("failed to write session event: %w", err)
	}

	return nil
}
