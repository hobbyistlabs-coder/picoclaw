package memory

import (
	"context"
	"path/filepath"
	"testing"

	"jane/pkg/providers"
)

func TestNewSQLiteStore_UsesDeleteJournalMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.db")
	store, err := NewSQLiteStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	var mode string
	if err := store.db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode error = %v", err)
	}
	if mode != "delete" {
		t.Fatalf("journal_mode = %q, want %q", mode, "delete")
	}
}

func TestSQLiteStore_WritesVisibleAcrossConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.db")
	writer, err := NewSQLiteStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStore(writer) error = %v", err)
	}
	defer writer.Close()

	if err := writer.AddFullMessage(
		context.Background(),
		"agent:main:pico:direct:pico:test",
		providers.Message{Role: "user", Content: "hello"},
	); err != nil {
		t.Fatalf("AddFullMessage() error = %v", err)
	}

	reader, err := NewSQLiteStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStore(reader) error = %v", err)
	}
	defer reader.Close()

	sessions, err := reader.ListSessions(context.Background(), "agent:main:pico:direct:pico:", 20, 0)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}

	history, err := reader.GetHistory(context.Background(), "agent:main:pico:direct:pico:test")
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if len(history) != 1 || history[0].Content != "hello" {
		t.Fatalf("history = %+v, want single visible message", history)
	}
}
