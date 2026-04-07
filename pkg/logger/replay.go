package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReplayErrorCategory defines categories for errors logged during session replay.
type ReplayErrorCategory string

const (
	ModelFailure          ReplayErrorCategory = "model_failure"
	InfrastructureFailure ReplayErrorCategory = "infrastructure_failure"
	LogicFailure          ReplayErrorCategory = "logic_failure"
	None                  ReplayErrorCategory = "none"
)

// ReplayEventDetails holds details specific to a replay event.
type ReplayEventDetails struct {
	CoTText   string         `json:"cot_text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Inputs    map[string]any `json:"inputs,omitempty"`
	Outputs   map[string]any `json:"outputs,omitempty"`
	FromState string         `json:"from_state,omitempty"`
	ToState   string         `json:"to_state,omitempty"`
}

// ReplayEvent represents a single logged event for session replay.
type ReplayEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"` // cot, tool_call, tool_result, state_transition, error
	Details       ReplayEventDetails  `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

type sessionFileState struct {
	mu   sync.Mutex
	file *os.File
}

var (
	sessionFiles = sync.Map{} // map[string]*sessionFileState
)

// LogSessionEvent safely appends a JSONL ReplayEvent to {workspacePath}/logs/{session_id}/events/events.jsonl
func LogSessionEvent(workspacePath string, event ReplayEvent) error {
	sessionID := event.SessionID
	if sessionID == "" {
		return fmt.Errorf("session ID is empty")
	}

	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	if event.ErrorCategory == "" {
		event.ErrorCategory = None
	}

	val, ok := sessionFiles.Load(sessionID)
	var state *sessionFileState

	if ok {
		state = val.(*sessionFileState)
	} else {
		state = &sessionFileState{}
		actual, loaded := sessionFiles.LoadOrStore(sessionID, state)
		state = actual.(*sessionFileState)

		if !loaded {
			// We won the race to create it, but we still need to initialize the file safely.
			// The lock guarantees other concurrent requests will wait until we're done opening it.
			state.mu.Lock()
			// Check again if file was somehow initialized (unlikely here, but good practice)
			if state.file == nil {
				// Sanitizing sessionID to avoid Path Traversal Vulnerabilities
				safeSessionID := filepath.Base(filepath.Clean(sessionID))
				if safeSessionID == "." || safeSessionID == ".." {
					state.mu.Unlock()
					sessionFiles.Delete(sessionID)
					return fmt.Errorf("invalid session ID")
				}
				logDir := filepath.Join(workspacePath, "logs", safeSessionID, "events")
				if err := os.MkdirAll(logDir, 0755); err != nil {
					state.mu.Unlock()
					sessionFiles.Delete(sessionID)
					return fmt.Errorf("failed to create log dir %s: %w", logDir, err)
				}
				filePath := filepath.Join(logDir, "events.jsonl")
				f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					state.mu.Unlock()
					sessionFiles.Delete(sessionID)
					return fmt.Errorf("failed to open file %s: %w", filePath, err)
				}
				state.file = f
			}
			state.mu.Unlock()
		}
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.file == nil {
		// Sanitizing sessionID to avoid Path Traversal Vulnerabilities
		safeSessionID := filepath.Base(filepath.Clean(sessionID))
		if safeSessionID == "." || safeSessionID == ".." {
			return fmt.Errorf("invalid session ID")
		}

		// Attempt to reopen the file if it was closed
		logDir := filepath.Join(workspacePath, "logs", safeSessionID, "events")
		if err := os.MkdirAll(logDir, 0755); err == nil {
			filePath := filepath.Join(logDir, "events.jsonl")
			f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				state.file = f
				// Re-store in the map if it was deleted
				sessionFiles.Store(sessionID, state)
			}
		}

		if state.file == nil {
			return fmt.Errorf("file handle is closed for session %s", sessionID)
		}
	}

	_, err = state.file.Write(append(data, '\n'))
	return err
}

// CleanupSessionLocks closes the file handle for a given session and removes it from the map.
func CleanupSessionLocks(sessionID string) {
	val, ok := sessionFiles.LoadAndDelete(sessionID)
	if ok {
		state := val.(*sessionFileState)
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.file != nil {
			_ = state.file.Close()
			state.file = nil
		}
	}
}
