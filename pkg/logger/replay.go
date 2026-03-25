package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogSessionEvent appends a structured JSON log entry for session replay
// according to the defined observability schema.
func LogSessionEvent(workspacePath, sessionID, eventType string, details map[string]any, errorCategory, errorMessage string) error {
	if workspacePath == "" || sessionID == "" {
		return nil // skip logging if workspace or session is not set
	}

	// Build the path to the events file
	logDir := filepath.Join(workspacePath, "logs", sessionID, "events")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create replay log directory: %w", err)
	}
	logFile := filepath.Join(logDir, "events.jsonl")

	// Construct the JSON object
	event := map[string]any{
		"timestamp":  time.Now().Format(time.RFC3339),
		"session_id": sessionID,
		"event_type": eventType,
	}

	if details != nil {
		event["details"] = details
	}

	if errorCategory != "" {
		event["error_category"] = errorCategory
	}

	if errorMessage != "" {
		event["error_message"] = errorMessage
	}

	// Serialize to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal replay log event: %w", err)
	}

	// Append to file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open replay log file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to replay log file: %w", err)
	}

	return nil
}
