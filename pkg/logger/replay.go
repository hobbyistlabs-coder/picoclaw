package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ReplayEventType string

const (
	ReplayEventTypeCoT             ReplayEventType = "cot"
	ReplayEventTypeToolCall        ReplayEventType = "tool_call"
	ReplayEventTypeToolResult      ReplayEventType = "tool_result"
	ReplayEventTypeStateTransition ReplayEventType = "state_transition"
	ReplayEventTypeError           ReplayEventType = "error"
)

type ReplayErrorCategory string

const (
	ReplayErrorModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorNone                  ReplayErrorCategory = "none"
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

type sessionFileState struct {
	mu   sync.Mutex
	file *os.File
}

var sessionLocks sync.Map

// getSafeSessionID neutralizes path traversal characters while preserving prefixes
func getSafeSessionID(key string) string {
	if key == "" || key == "." || key == ".." {
		return "default_session"
	}
	safe := strings.ReplaceAll(key, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	if safe == "." || safe == ".." {
		return "default_session"
	}
	return safe
}

// LogSessionEvent appends an event to the session's replay log
func LogSessionEvent(workspacePath, sessionID string, event ReplayEvent) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	safeSessionID := getSafeSessionID(sessionID)

	// Create JSON representation of the event
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	event.SessionID = sessionID

	data, err := json.Marshal(event)
	if err != nil {
		WarnCF("replay", "Failed to marshal replay event", map[string]any{"error": err.Error()})
		return
	}
	data = append(data, '\n')

	// Get or create session lock/file state
	val, ok := sessionLocks.Load(safeSessionID)
	if !ok {
		// Ensure the directory exists
		eventsDir := filepath.Join(workspacePath, "logs", safeSessionID, "events")
		if err := os.MkdirAll(eventsDir, 0o700); err != nil {
			WarnCF("replay", "Failed to create events directory", map[string]any{"error": err.Error(), "dir": eventsDir})
			return
		}

		filePath := filepath.Join(eventsDir, "events.jsonl")
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			WarnCF("replay", "Failed to open events file", map[string]any{"error": err.Error(), "file": filePath})
			return
		}

		newState := &sessionFileState{file: f}
		actual, loaded := sessionLocks.LoadOrStore(safeSessionID, newState)
		if loaded {
			f.Close()
			val = actual
		} else {
			val = newState
		}
	}

	state := val.(*sessionFileState)
	state.mu.Lock()
	defer state.mu.Unlock()

	// Reopen file if it was closed
	if state.file == nil {
		eventsDir := filepath.Join(workspacePath, "logs", safeSessionID, "events")
		filePath := filepath.Join(eventsDir, "events.jsonl")
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			WarnCF("replay", "Failed to reopen events file", map[string]any{"error": err.Error(), "file": filePath})
			return
		}
		state.file = f
	}

	if _, err := state.file.Write(data); err != nil {
		WarnCF("replay", "Failed to write replay event", map[string]any{"error": err.Error()})
	}
}

// CleanupSessionLocks closes the file handles for a session
func CleanupSessionLocks(sessionID string) {
	if sessionID == "" {
		return
	}
	safeSessionID := getSafeSessionID(sessionID)
	// We do NOT delete the session lock from the map, because async callbacks
	// might still need to log events for this session later.
	// Instead, we just close the file descriptor and set it to nil.
	// When the next event comes in, LogSessionEvent will re-open the file.
	if val, ok := sessionLocks.Load(safeSessionID); ok {
		state := val.(*sessionFileState)
		state.mu.Lock()
		if state.file != nil {
			state.file.Close()
			state.file = nil
		}
		state.mu.Unlock()
	}
}
