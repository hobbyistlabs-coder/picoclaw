// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ReplayErrorCategory represents the categorization of errors for session replay.
type ReplayErrorCategory string

const (
	ReplayModelFailure          ReplayErrorCategory = "model_failure"
	ReplayInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayNone                  ReplayErrorCategory = "none"
)

// SessionEventDetails holds the granular details of a session event.
type SessionEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

// SessionEvent represents a structured log event for session replay.
type SessionEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // "cot", "tool_call", "tool_result", "state_transition", "error"
	Details       SessionEventDetails `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

// LogSessionEvent writes a structured JSON event to the specified workspace logs directory.
func LogSessionEvent(workspacePath, sessionID string, event SessionEvent) {
	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if event.SessionID == "" {
		event.SessionID = sessionID
	}

	// Default to "none" if ErrorCategory is missing and EventType is not "error"
	if event.ErrorCategory == "" && event.EventType != "error" {
		event.ErrorCategory = ReplayNone
	}

	// Create directory: {workspacePath}/logs/{session_id}/events/
	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		WarnCF("logger", "Failed to create session replay log directory", map[string]any{"error": err.Error(), "path": logDir})
		return
	}

	// Format filename: {timestamp}_event.json
	filename := fmt.Sprintf("%s_event.json", event.Timestamp.Format("20060102150405.000000"))
	filePath := filepath.Join(logDir, filename)

	// Serialize event
	b, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		WarnCF("logger", "Failed to marshal session event", map[string]any{"error": err.Error(), "event": event})
		return
	}

	// Write file
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		WarnCF("logger", "Failed to write session event file", map[string]any{"error": err.Error(), "path": filePath})
	}
}
