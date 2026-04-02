package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLogSessionEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "replay-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sessionID := "test-session-123"

	event := SessionEvent{
		EventType: EventTypeCoT,
		Details: &EventDetails{
			CoTText: "This is a reasoning step.",
		},
	}

	LogSessionEvent(tempDir, sessionID, event)

	logFilePath := filepath.Join(tempDir, "logs", sessionID, "events", "events.jsonl")
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var loggedEvent SessionEvent
	if err := json.Unmarshal(data, &loggedEvent); err != nil {
		t.Fatalf("Failed to unmarshal logged event: %v", err)
	}

	if loggedEvent.EventType != EventTypeCoT {
		t.Errorf("Expected event type %s, got %s", EventTypeCoT, loggedEvent.EventType)
	}
	if loggedEvent.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, loggedEvent.SessionID)
	}
	if loggedEvent.Timestamp == "" {
		t.Errorf("Expected timestamp to be populated")
	}
	if loggedEvent.Details == nil || loggedEvent.Details.CoTText != "This is a reasoning step." {
		t.Errorf("Expected valid details, got %v", loggedEvent.Details)
	}
}

func TestLogSessionEvent_PathTraversalProtection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "replay-test-traversal-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sessionID := "../../../etc/passwd"

	event := SessionEvent{
		EventType: EventTypeCoT,
	}

	LogSessionEvent(tempDir, sessionID, event)

	expectedCleanID := ".._.._.._etc_passwd"
	logFilePath := filepath.Join(tempDir, "logs", expectedCleanID, "events", "events.jsonl")

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Fatalf("Expected file to exist at %s, but it didn't", logFilePath)
	}
}

func TestLogSessionEvent_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "replay-test-concurrent-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sessionID := "concurrent-session"

	done := make(chan bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		go func(idx int) {
			LogSessionEvent(tempDir, sessionID, SessionEvent{
				EventType: EventTypeToolCall,
			})
			done <- true
		}(i)
	}

	for i := 0; i < iterations; i++ {
		<-done
	}

	logFilePath := filepath.Join(tempDir, "logs", sessionID, "events", "events.jsonl")
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := 0
	for _, char := range data {
		if char == '\n' {
			lines++
		}
	}

	if lines != iterations {
		t.Errorf("Expected %d lines, got %d", iterations, lines)
	}
}
