package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ReplayErrorCategory categorizes errors for structured logging
type ReplayErrorCategory string

const (
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
)

// SessionEvent represents a structured log event for session replay
type SessionEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // cot, tool_call, tool_result, state_transition, error
	Details       map[string]any      `json:"details"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

// LogSessionEvent appends a structured JSON event to the session replay log
func LogSessionEvent(workspacePath, sessionID, eventType string, details map[string]any, errCat ReplayErrorCategory, errMsg string) error {
	if workspacePath == "" || sessionID == "" {
		return nil // skip if not configured properly
	}

	eventsDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return fmt.Errorf("failed to create events directory: %w", err)
	}

	event := SessionEvent{
		Timestamp:     time.Now().UTC(),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errCat,
		ErrorMessage:  errMsg,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal session event: %w", err)
	}

	// Append to a single events.jsonl file per session to avoid filename collisions
	logFile := filepath.Join(eventsDir, "events.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open events file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}
