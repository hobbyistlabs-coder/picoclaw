package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionReplayEvent represents a structured event for session replay
// according to the JSON schema in AGENT_LOOP_IMPROVEMENTS.md.
type SessionReplayEvent struct {
	Timestamp     string         `json:"timestamp"`
	SessionID     string         `json:"session_id"`
	EventType     string         `json:"event_type"` // "cot", "tool_call", "tool_result", "state_transition", "error"
	Details       map[string]any `json:"details,omitempty"`
	ErrorCategory string         `json:"error_category,omitempty"` // "model_failure", "infrastructure_failure", "logic_failure", "none"
	ErrorMessage  string         `json:"error_message,omitempty"`
}

// LogSessionEvent writes a structured JSON log for session replay
// to {workspacePath}/logs/{session_id}/events/{timestamp}_event.json
func LogSessionEvent(workspacePath, sessionID, eventType string, errorCategory ErrorCategory, details map[string]any, errMsg string) {
	if sessionID == "" {
		return
	}

	// Format timestamp
	now := time.Now()
	timestampStr := now.Format(time.RFC3339)
	fileTimestampStr := now.Format("20060102_150405.000000")

	// Determine error category string based on schema
	var errCatStr string
	switch errorCategory {
	case ErrorCategoryModelFailure:
		errCatStr = "model_failure"
	case ErrorCategoryInfrastructureFailure:
		errCatStr = "infrastructure_failure"
	case ErrorCategoryLogicFailure:
		errCatStr = "logic_failure"
	default:
		if errMsg != "" {
			errCatStr = "logic_failure" // Fallback if error is provided but category is not specified
		} else {
			errCatStr = "none"
		}
	}

	event := SessionReplayEvent{
		Timestamp:     timestampStr,
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errCatStr,
		ErrorMessage:  errMsg,
	}

	// Create directory {workspacePath}/logs/{session_id}/events
	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		ErrorCF("replay", "Failed to create session log directory", map[string]any{
			"error": err.Error(),
			"dir":   logDir,
		})
		return
	}

	// Marshal JSON
	data, err := json.Marshal(event)
	if err != nil {
		ErrorCF("replay", "Failed to marshal session event", map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Write to file logs/{session_id}/events/{timestamp}_event.json
	filename := filepath.Join(logDir, fmt.Sprintf("%s_event.json", fileTimestampStr))
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		ErrorCF("replay", "Failed to write session event to file", map[string]any{
			"error": err.Error(),
			"file":  filename,
		})
	}
}
