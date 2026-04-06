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
	ReplayErrorCategoryModelFailure          ReplayErrorCategory = "model_failure"
	ReplayErrorCategoryInfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	ReplayErrorCategoryLogicFailure          ReplayErrorCategory = "logic_failure"
	ReplayErrorCategoryNone                  ReplayErrorCategory = "none"
)

type SessionEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

type SessionEvent struct {
	Timestamp     time.Time           `json:"timestamp"`
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
	sessionFiles sync.Map
)

func safeSessionKey(key string) string {
	clean := filepath.Clean(key)
	flat := strings.ReplaceAll(clean, "/", "_")
	flat = strings.ReplaceAll(flat, "\\", "_")
	if flat == "." || flat == ".." {
		return "invalid_session"
	}
	return flat
}

func LogSessionEvent(workspacePath, sessionID string, event SessionEvent) {
	if sessionID == "" {
		return
	}
	event.Timestamp = time.Now()
	event.SessionID = sessionID

	safeSession := safeSessionKey(sessionID)
	eventsDir := filepath.Join(workspacePath, "logs", safeSession, "events")

	err := os.MkdirAll(eventsDir, 0755)
	if err != nil {
		WarnCF("replay", "failed to create events dir", map[string]any{"error": err.Error(), "dir": eventsDir})
		return
	}

	eventsFile := filepath.Join(eventsDir, "events.jsonl")

	var state *sessionFileState
	val, ok := sessionFiles.Load(sessionID)
	if !ok {
		f, err := os.OpenFile(eventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			WarnCF("replay", "failed to open events file", map[string]any{"error": err.Error(), "file": eventsFile})
			return
		}
		newState := &sessionFileState{file: f}
		val, _ = sessionFiles.LoadOrStore(sessionID, newState)
		state = val.(*sessionFileState)
		if state != newState {
			f.Close() // Another goroutine won the race, close ours
		}
	} else {
		state = val.(*sessionFileState)
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	b, err := json.Marshal(event)
	if err != nil {
		WarnCF("replay", "failed to marshal event", map[string]any{"error": err.Error()})
		return
	}

	_, err = state.file.Write(append(b, '\n'))
	if err != nil {
		WarnCF("replay", "failed to write event", map[string]any{"error": err.Error()})
	}
}

func CleanupSessionLocks(sessionID string) {
	if val, ok := sessionFiles.LoadAndDelete(sessionID); ok {
		state := val.(*sessionFileState)
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.file != nil {
			state.file.Close()
		}
	}
}
