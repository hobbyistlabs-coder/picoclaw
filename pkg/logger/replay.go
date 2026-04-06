package logger

import (
	"context"
	"encoding/json"
	"fmt"
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
	EventType     string              `json:"event_type"` // cot, tool_call, tool_result, state_transition, error
	Details       SessionEventDetails `json:"details"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

type sessionFileState struct {
	mu   sync.Mutex
	file *os.File
}

var (
	sessionLocks sync.Map // map[string]*sessionFileState
	WorkspaceDir string
)

// SetWorkspaceDir sets the base directory for replay logs
func SetWorkspaceDir(dir string) {
	WorkspaceDir = dir
}

// getSessionDir returns the directory for a specific session's events
func getSessionDir(sessionID string) string {
	base := WorkspaceDir
	if base == "" {
		// Use default if not set
		home, err := os.UserHomeDir()
		if err == nil {
			base = filepath.Join(home, ".jane-ai", "workspace")
		} else {
			base = "."
		}
	}
	return filepath.Join(base, "logs", sanitizeSessionKey(sessionID), "events")
}

// sanitizeSessionKey prevents path traversal but preserves structured keys like group:-100/12
func sanitizeSessionKey(key string) string {
	if key == "" || key == "." || key == ".." {
		return "default_session"
	}
	s := filepath.Clean(key)
	// Replace slashes with underscores to flatten the path
	s = string(os.PathSeparator) + s
	s = filepath.ToSlash(s)
	s = s[1:]

	// manually replace / and \ with _
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}

func LogSessionEvent(ctx context.Context, sessionID string, event SessionEvent) {
	if sessionID == "" {
		return
	}

	stateRaw, _ := sessionLocks.LoadOrStore(sessionID, &sessionFileState{})
	state := stateRaw.(*sessionFileState)

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.file == nil {
		dir := getSessionDir(sessionID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			ErrorC("logger", fmt.Sprintf("Failed to create session log dir: %v", err))
			return
		}

		filePath := filepath.Join(dir, "events.jsonl")
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			ErrorC("logger", fmt.Sprintf("Failed to open session log file: %v", err))
			return
		}
		state.file = f
	}

	event.Timestamp = time.Now().UTC()
	if event.SessionID == "" {
		event.SessionID = sessionID
	}

	data, err := json.Marshal(event)
	if err != nil {
		ErrorC("logger", fmt.Sprintf("Failed to marshal session event: %v", err))
		return
	}

	data = append(data, '\n')
	if _, err := state.file.Write(data); err != nil {
		ErrorC("logger", fmt.Sprintf("Failed to write session event: %v", err))
	}
}

func CleanupSessionLocks(sessionID string) {
	if stateRaw, ok := sessionLocks.Load(sessionID); ok {
		state := stateRaw.(*sessionFileState)
		state.mu.Lock()
		if state.file != nil {
			state.file.Close()
			state.file = nil
		}
		state.mu.Unlock()
		sessionLocks.Delete(sessionID)
	}
}
