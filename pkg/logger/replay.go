package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ReplayEventType defines the type of event being logged.
type ReplayEventType string

const (
	EventTypeCoT             ReplayEventType = "cot"
	EventTypeToolCall        ReplayEventType = "tool_call"
	EventTypeToolResult      ReplayEventType = "tool_result"
	EventTypeStateTransition ReplayEventType = "state_transition"
	EventTypeError           ReplayEventType = "error"
)

// ReplayErrorCategory categorizes failures.
type ReplayErrorCategory string

const (
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
)

// SessionEventDetails contains specific details for a session event.
type SessionEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

// SessionEvent represents a single logged event in the session replay.
type SessionEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     ReplayEventType     `json:"event_type"`
	Details       SessionEventDetails `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

type sessionFileState struct {
	mu   sync.Mutex
	file *os.File
}

var sessionStates sync.Map

func getSessionState(sessionID string) *sessionFileState {
	state, _ := sessionStates.LoadOrStore(sessionID, &sessionFileState{})
	return state.(*sessionFileState)
}

// CleanupSessionLocks removes the lock and closes the file for a specific session ID to prevent memory and file descriptor leaks.
func CleanupSessionLocks(sessionID string) {
	if stateVal, ok := sessionStates.LoadAndDelete(sessionID); ok {
		state := stateVal.(*sessionFileState)
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.file != nil {
			_ = state.file.Close()
			state.file = nil
		}
	}
}

// LogSessionEvent logs a structured event to a JSONL file per session safely using mutexes.
func LogSessionEvent(workspacePath string, event SessionEvent) {
	if event.SessionID == "" || workspacePath == "" {
		return
	}

	cleanedSessionID := strings.ReplaceAll(event.SessionID, "/", "_")
	cleanedSessionID = strings.ReplaceAll(cleanedSessionID, "\\", "_")
	if cleanedSessionID == "." || cleanedSessionID == ".." {
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	state := getSessionState(event.SessionID)
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.file == nil {
		// Ensure the directory exists
		eventsDir := filepath.Join(workspacePath, "logs", cleanedSessionID, "events")
		if err := os.MkdirAll(eventsDir, 0o755); err != nil {
			return
		}

		// Append to events.jsonl
		filePath := filepath.Join(eventsDir, "events.jsonl")
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return
		}
		state.file = file
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	// Write as JSONL
	_, _ = state.file.Write(append(data, '\n'))
}
