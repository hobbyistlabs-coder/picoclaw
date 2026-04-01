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

var sessionLocks sync.Map

type sessionEvent struct {
	Timestamp     string              `json:"timestamp"`
	SessionID     string              `json:"session_id"`
	EventType     string              `json:"event_type"`
	Details       map[string]any      `json:"details,omitempty"`
	ErrorCategory ReplayErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
}

// sanitizeSessionKey extracts a safe base name to prevent path traversal
func sanitizeSessionKey(key string) string {
	base := filepath.Base(filepath.Clean(key))
	if base == "." || base == ".." || base == "/" || base == "\\" {
		return ""
	}
	return base
}

// LogSessionEvent appends a session event to the event stream JSONL file.
// Errors are suppressed (best-effort) to avoid interrupting the agent loop.
func LogSessionEvent(workspacePath, sessionID, eventType string, details map[string]any, errorCategory ReplayErrorCategory, errorMsg string) {
	if workspacePath == "" || sessionID == "" {
		return
	}

	sanitizedSessionID := sanitizeSessionKey(sessionID)
	if sanitizedSessionID == "" {
		return // Ignore invalid session ID
	}

	event := sessionEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		SessionID:     sanitizedSessionID,
		EventType:     eventType,
		Details:       details,
		ErrorCategory: errorCategory,
		ErrorMessage:  errorMsg,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return
	}
	eventJSON = append(eventJSON, '\n')

	// Get or create mutex for this session
	lockIface, _ := sessionLocks.LoadOrStore(sanitizedSessionID, &sync.Mutex{})
	lock := lockIface.(*sync.Mutex)

	lock.Lock()
	defer lock.Unlock()

	// Safe directory traversal validation
	absWorkspace, err := filepath.Abs(workspacePath)
	if err != nil {
		return
	}

	eventsDir := filepath.Join(absWorkspace, "logs", sanitizedSessionID, "events")
	eventsDirAbs, err := filepath.Abs(eventsDir)
	if err != nil {
		return
	}

	// Verify that the target directory is still under the workspace directory
	rel, err := filepath.Rel(absWorkspace, eventsDirAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") || strings.HasPrefix(rel, "..\\") {
		return // Path traversal attempt
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(eventsDirAbs, 0o755); err != nil {
		return
	}

	eventsFile := filepath.Join(eventsDirAbs, "events.jsonl")

	// Append JSONL to file
	f, err := os.OpenFile(eventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(eventJSON)
}
