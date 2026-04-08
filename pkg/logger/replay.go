package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ReplayErrorCategory string

const (
	ModelFailure          ReplayErrorCategory = "model_failure"
	InfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	LogicFailure          ReplayErrorCategory = "logic_failure"
	None                  ReplayErrorCategory = "none"
)

type SessionEventDetails struct {
	CotText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
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

type sessionFileState struct {
	mu   sync.Mutex
	file *os.File
}

var (
	sessionLocks sync.Map
)

func sanitizeSessionKey(key string) string {
	if key == "" || key == "." || key == ".." {
		return "default_session"
	}
	cleanKey := strings.ReplaceAll(key, "/", "_")
	cleanKey = strings.ReplaceAll(cleanKey, "\\", "_")
	return cleanKey
}

func LogSessionEvent(workspacePath, sessionID, eventType string, details SessionEventDetails, errCategory ReplayErrorCategory, errMsg string) {
	safeSessionID := sanitizeSessionKey(sessionID)

	event := SessionEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		SessionID:     safeSessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errCategory,
		ErrorMessage:  errMsg,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		WarnCF("replay", "Failed to marshal session event", map[string]any{"error": err.Error()})
		return
	}

	eventJSON = append(eventJSON, '\n')

	dirPath := filepath.Join(workspacePath, "logs", safeSessionID, "events")
	if err := os.MkdirAll(dirPath, 0o700); err != nil {
		WarnCF("replay", "Failed to create events directory", map[string]any{"error": err.Error(), "path": dirPath})
		return
	}

	filePath := filepath.Join(dirPath, "events.jsonl")

	val, _ := sessionLocks.LoadOrStore(safeSessionID, &sessionFileState{})
	state := val.(*sessionFileState)

	state.mu.Lock()
	defer state.mu.Unlock()

	// Handle race condition: state might be marked for cleanup or already closed.
	// In that case, we open the file, write, and close it immediately.
	if state.file == nil {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			WarnCF("replay", "Failed to open events file", map[string]any{"error": err.Error(), "path": filePath})
			return
		}
		state.file = f
	}

	if _, err := state.file.Write(eventJSON); err != nil {
		WarnCF("replay", "Failed to write event", map[string]any{"error": err.Error(), "path": filePath})
	}
}

func CleanupSessionLocks(sessionID string) {
	safeSessionID := sanitizeSessionKey(sessionID)
	if val, ok := sessionLocks.LoadAndDelete(safeSessionID); ok {
		state := val.(*sessionFileState)
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.file != nil {
			state.file.Close()
			state.file = nil
		}
	}
}
