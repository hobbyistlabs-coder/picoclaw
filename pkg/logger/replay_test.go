package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogSessionEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "replay_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sessionID := "test_session_123"

	event := ReplayEvent{
		Timestamp: time.Now().UTC(),
		SessionID: sessionID,
		EventType: EventTypeCoT,
		Details: ReplayEventDetails{
			CoTText: "Thinking...",
		},
	}

	LogSessionEvent(tempDir, sessionID, event)

	logFile := filepath.Join(tempDir, "logs", sessionID, "events", "events.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var readEvent ReplayEvent
	if err := json.Unmarshal(data, &readEvent); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if readEvent.SessionID != sessionID {
		t.Errorf("Expected session ID %q, got %q", sessionID, readEvent.SessionID)
	}

	if readEvent.EventType != EventTypeCoT {
		t.Errorf("Expected event type %q, got %q", EventTypeCoT, readEvent.EventType)
	}

	if readEvent.Details.CoTText != "Thinking..." {
		t.Errorf("Expected CoT text %q, got %q", "Thinking...", readEvent.Details.CoTText)
	}

	CleanupSessionLocks(sessionID)
	if _, ok := sessionLocks.Load(sessionID); ok {
		t.Errorf("Expected session lock to be removed")
	}
}
