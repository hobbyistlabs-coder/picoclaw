package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReplayErrorCategory defines the category of error for session replays.
type ReplayErrorCategory string

const (
	ErrorCategoryModelFailureReplay          ReplayErrorCategory = "model_failure"
	ErrorCategoryInfrastructureFailureReplay ReplayErrorCategory = "infrastructure_failure"
	ErrorCategoryLogicFailureReplay          ReplayErrorCategory = "logic_failure"
	ErrorCategoryNoneReplay                  ReplayErrorCategory = "none"
)

// sessionFileState manages the file handle and mutex for a session's log file.
type sessionFileState struct {
	mu   sync.Mutex
	file *os.File
}

// Global map to hold session states.
var (
	sessionFiles sync.Map // map[string]*sessionFileState
)

// SessionEventDetails holds the structured details of the event.
type SessionEventDetails struct {
	CotText   string `json:"cot_text,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	Inputs    any    `json:"inputs,omitempty"`
	Outputs   any    `json:"outputs,omitempty"`
	FromState string `json:"from_state,omitempty"`
	ToState   string `json:"to_state,omitempty"`
}

// SessionEvent is the top-level structure for observability logs.
type SessionEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // cot, tool_call, tool_result, state_transition, error
	Details       SessionEventDetails `json:"details,omitempty"`
	ErrorCategory string              `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

// LogSessionEvent logs a structured event to a session-specific JSONL file.
func LogSessionEvent(workspacePath, sessionID, eventType string, details SessionEventDetails, errorCategory ReplayErrorCategory, errorMessage string) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	event := SessionEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SessionID:     sessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: string(errorCategory),
		ErrorMessage:  errorMessage,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return // Best-effort logging
	}

	stateVal, ok := sessionFiles.Load(sessionID)
	var state *sessionFileState

	if !ok {
		// Needs initialization
		logDir := filepath.Join(workspacePath, "logs", sessionID, "events")
		err := os.MkdirAll(logDir, 0o755)
		if err != nil {
			return
		}

		logPath := filepath.Join(logDir, "events.jsonl")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return
		}

		newState := &sessionFileState{file: file}
		actualStateVal, loaded := sessionFiles.LoadOrStore(sessionID, newState)
		if loaded {
			// Another goroutine beat us to it, close the file we just opened
			file.Close()
			state = actualStateVal.(*sessionFileState)
		} else {
			state = newState
		}
	} else {
		state = stateVal.(*sessionFileState)
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	// If file was closed for some reason, ignore
	if state.file != nil {
		state.file.Write(eventJSON)
		state.file.WriteString("\n")
	}
}

// CleanupSessionLocks closes the file handle for a given session to prevent memory/descriptor leaks.
func CleanupSessionLocks(sessionID string) {
	stateVal, ok := sessionFiles.LoadAndDelete(sessionID)
	if ok {
		state := stateVal.(*sessionFileState)
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.file != nil {
			state.file.Close()
			state.file = nil
		}
	}
}
