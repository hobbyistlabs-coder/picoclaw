package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLogSessionEventAppendsAndSanitizesPath(t *testing.T) {
	dir := t.TempDir()
	event := SessionEvent{
		SessionID: "team/chat-1",
		EventType: EventTypeToolCall,
		Details: SessionEventDetails{
			ToolName: "search",
			Inputs:   map[string]any{"q": "test"},
		},
	}

	LogSessionEvent(dir, event)
	LogSessionEvent(dir, SessionEvent{SessionID: "team/chat-1", EventType: EventTypeToolResult})
	CleanupSessionLocks("team/chat-1")

	path := filepath.Join(dir, "logs", "team_chat-1", "events", "events.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
		var got SessionEvent
		if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if got.SessionID != "team/chat-1" {
			t.Fatalf("SessionID = %q, want %q", got.SessionID, "team/chat-1")
		}
		if got.Timestamp.IsZero() {
			t.Fatal("Timestamp was not populated")
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("line count = %d, want 2", count)
	}
}

func TestCleanupSessionLocksAllowsReopen(t *testing.T) {
	dir := t.TempDir()
	sessionID := "session-1"

	LogSessionEvent(dir, SessionEvent{SessionID: sessionID, EventType: EventTypeCoT})
	CleanupSessionLocks(sessionID)
	LogSessionEvent(dir, SessionEvent{SessionID: sessionID, EventType: EventTypeError})
	CleanupSessionLocks(sessionID)

	path := filepath.Join(dir, "logs", sessionID, "events", "events.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected replay log data")
	}
}
