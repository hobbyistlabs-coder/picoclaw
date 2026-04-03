package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type EventType string

const (
	EventTypeCoT             EventType = "cot"
	EventTypeToolCall        EventType = "tool_call"
	EventTypeToolResult      EventType = "tool_result"
	EventTypeStateTransition EventType = "state_transition"
	EventTypeError           EventType = "error"
)

type EventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type SessionEvent struct {
	Timestamp     string       `json:"timestamp"`
	SessionID     string       `json:"session_id"`
	EventType     EventType    `json:"event_type"`
	Details       EventDetails `json:"details,omitempty"`
	ErrorCategory string       `json:"error_category,omitempty"`
	ErrorMessage  string       `json:"error_message,omitempty"`
}

type sessionFileState struct {
	mu   sync.Mutex
	file *os.File
}

var (
	sessionStates sync.Map // map[string]*sessionFileState
	workspaceDir  string
)

// SetWorkspaceDir initializes the workspace directory used for storing logs
func SetWorkspaceDir(dir string) {
	workspaceDir = dir
}

// LogSessionEvent logs a structured event for session replay
func LogSessionEvent(sessionID string, event SessionEvent) {
	if workspaceDir == "" {
		return // Ignore if workspace is not set
	}

	event.Timestamp = time.Now().Format(time.RFC3339)
	event.SessionID = sessionID

	data, err := json.Marshal(event)
	if err != nil {
		WarnCF("logger", "failed to marshal session event", map[string]any{"error": err.Error()})
		return
	}

	logDir := filepath.Join(workspaceDir, "logs", sessionID, "events")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		WarnCF("logger", "failed to create session log directory", map[string]any{"error": err.Error()})
		return
	}

	filePath := filepath.Join(logDir, "events.jsonl")

	stateAny, _ := sessionStates.LoadOrStore(sessionID, &sessionFileState{})
	state := stateAny.(*sessionFileState)

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.file == nil {
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			WarnCF("logger", "failed to open session event log file", map[string]any{"error": err.Error()})
			return
		}
		state.file = file
	}

	_, err = state.file.Write(append(data, '\n'))
	if err != nil {
		WarnCF("logger", "failed to write to session event log file", map[string]any{"error": err.Error()})
	}
}

// CleanupSessionLocks closes the file handle and removes the session state
func CleanupSessionLocks(sessionID string) {
	if stateAny, ok := sessionStates.LoadAndDelete(sessionID); ok {
		state := stateAny.(*sessionFileState)
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.file != nil {
			_ = state.file.Close()
			state.file = nil
		}
	}
}
