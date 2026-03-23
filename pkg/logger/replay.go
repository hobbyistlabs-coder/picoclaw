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

type EventDetails struct {
	COTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type SessionEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"`
	Details       *EventDetails       `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

var replayMu sync.Mutex

func LogSessionEvent(workspacePath, sessionID, eventType string, details *EventDetails, errorCategory ReplayErrorCategory, errorMessage string) {
	if workspacePath == "" || sessionID == "" || eventType == "" {
		return
	}

	event := SessionEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errorCategory,
		ErrorMessage:  errorMessage,
	}

	b, err := json.Marshal(event)
	if err != nil {
		WarnCF("replay", "Failed to marshal session event", map[string]any{"error": err.Error()})
		return
	}

	eventsDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		WarnCF("replay", "Failed to create events directory", map[string]any{"error": err.Error(), "path": eventsDir})
		return
	}

	eventsFile := filepath.Join(eventsDir, "events.jsonl")

	replayMu.Lock()
	defer replayMu.Unlock()

	f, err := os.OpenFile(eventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		WarnCF("replay", "Failed to open events file", map[string]any{"error": err.Error(), "path": eventsFile})
		return
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%s\n", string(b)); err != nil {
		WarnCF("replay", "Failed to write event to file", map[string]any{"error": err.Error(), "path": eventsFile})
	}
}
