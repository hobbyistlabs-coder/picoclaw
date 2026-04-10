package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"jane/pkg/providers"
)

type StoredSession struct {
	Key       string
	Summary   string
	Messages  []providers.Message
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SQLiteStore struct {
	db *sql.DB
}

func SQLitePath(dir string) string {
	return filepath.Join(dir, "sessions.db")
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	// SECURITY: Restrict directory permissions to 0o700 (user read/write/execute only)
	// to prevent other local users from reading the sensitive sessions SQLite database.
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("memory: create sqlite dir: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("memory: open sqlite: %w", err)
	}
	if _, err = db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("memory: enable foreign keys: %w", err)
	}
	if err = createSQLiteSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func createSQLiteSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		session_key TEXT PRIMARY KEY,
		summary TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_key TEXT NOT NULL,
		payload TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(session_key) REFERENCES sessions(session_key) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_messages_session_key ON messages(session_key, id);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("memory: create sqlite schema: %w", err)
	}
	return nil
}

func (s *SQLiteStore) AddMessage(ctx context.Context, sessionKey, role, content string) error {
	return s.AddFullMessage(ctx, sessionKey, providers.Message{
		Role:    role,
		Content: content,
	})
}

func (s *SQLiteStore) AddFullMessage(ctx context.Context, sessionKey string, msg providers.Message) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memory: begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if err = upsertSession(tx, sessionKey, "", now, now, false); err != nil {
		return err
	}
	if err = insertMessage(tx, sessionKey, msg, now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE sessions SET updated_at = ? WHERE session_key = ?`, now.Format(time.RFC3339Nano), sessionKey); err != nil {
		return fmt.Errorf("memory: touch session: %w", err)
	}
	return tx.Commit()
}

func (s *SQLiteStore) GetHistory(ctx context.Context, sessionKey string) ([]providers.Message, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT payload FROM messages WHERE session_key = ? ORDER BY id`, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("memory: query history: %w", err)
	}
	defer rows.Close()

	msgs := make([]providers.Message, 0)
	for rows.Next() {
		var payload string
		if err = rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("memory: scan history: %w", err)
		}
		var msg providers.Message
		if err = json.Unmarshal([]byte(payload), &msg); err != nil {
			return nil, fmt.Errorf("memory: decode history: %w", err)
		}
		msgs = append(msgs, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("memory: iterate history: %w", err)
	}
	return msgs, nil
}

func (s *SQLiteStore) GetSummary(ctx context.Context, sessionKey string) (string, error) {
	var summary string
	err := s.db.QueryRowContext(ctx, `SELECT summary FROM sessions WHERE session_key = ?`, sessionKey).Scan(&summary)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("memory: get summary: %w", err)
	}
	return summary, nil
}

func (s *SQLiteStore) SetSummary(ctx context.Context, sessionKey, summary string) error {
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (session_key, summary, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(session_key) DO UPDATE SET summary = excluded.summary, updated_at = excluded.updated_at
	`, sessionKey, summary, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("memory: set summary: %w", err)
	}
	return nil
}

func (s *SQLiteStore) TruncateHistory(ctx context.Context, sessionKey string, keepLast int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memory: begin truncate tx: %w", err)
	}
	defer tx.Rollback()

	var deleteQuery string
	var args []any
	if keepLast <= 0 {
		deleteQuery = `DELETE FROM messages WHERE session_key = ?`
		args = []any{sessionKey}
	} else {
		deleteQuery = `
			DELETE FROM messages
			WHERE session_key = ?
			AND id NOT IN (
				SELECT id FROM messages WHERE session_key = ? ORDER BY id DESC LIMIT ?
			)
		`
		args = []any{sessionKey, sessionKey, keepLast}
	}
	if _, err = tx.ExecContext(ctx, deleteQuery, args...); err != nil {
		return fmt.Errorf("memory: truncate history: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE sessions SET updated_at = ? WHERE session_key = ?`, time.Now().UTC().Format(time.RFC3339Nano), sessionKey); err != nil {
		return fmt.Errorf("memory: touch truncated session: %w", err)
	}
	return tx.Commit()
}

func (s *SQLiteStore) SetHistory(ctx context.Context, sessionKey string, history []providers.Message) error {
	return s.ImportSession(ctx, StoredSession{
		Key:      sessionKey,
		Messages: history,
	})
}

func (s *SQLiteStore) ImportSession(ctx context.Context, session StoredSession) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memory: begin import tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	createdAt := session.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := session.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	if err = upsertSession(tx, session.Key, session.Summary, createdAt, updatedAt, true); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM messages WHERE session_key = ?`, session.Key); err != nil {
		return fmt.Errorf("memory: clear imported history: %w", err)
	}
	for _, msg := range session.Messages {
		if err = insertMessage(tx, session.Key, msg, updatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) ListSessions(ctx context.Context, prefix string, limit, offset int) ([]StoredSession, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT session_key, summary, created_at, updated_at
		FROM sessions
		WHERE session_key LIKE ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, prefix+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("memory: list sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]StoredSession, 0)
	for rows.Next() {
		var (
			sessionKey string
			summary    string
			createdRaw string
			updatedRaw string
		)
		if err = rows.Scan(&sessionKey, &summary, &createdRaw, &updatedRaw); err != nil {
			return nil, fmt.Errorf("memory: scan session row: %w", err)
		}
		messages, histErr := s.GetHistory(ctx, sessionKey)
		if histErr != nil {
			return nil, histErr
		}
		sessions = append(sessions, StoredSession{
			Key:       sessionKey,
			Summary:   summary,
			Messages:  messages,
			CreatedAt: parseSQLiteTime(createdRaw),
			UpdatedAt: parseSQLiteTime(updatedRaw),
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("memory: iterate sessions: %w", err)
	}
	return sessions, nil
}

func (s *SQLiteStore) GetSession(ctx context.Context, sessionKey string) (StoredSession, error) {
	var (
		summary    string
		createdRaw string
		updatedRaw string
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT summary, created_at, updated_at
		FROM sessions
		WHERE session_key = ?
	`, sessionKey).Scan(&summary, &createdRaw, &updatedRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return StoredSession{}, os.ErrNotExist
	}
	if err != nil {
		return StoredSession{}, fmt.Errorf("memory: get session: %w", err)
	}
	messages, err := s.GetHistory(ctx, sessionKey)
	if err != nil {
		return StoredSession{}, err
	}
	return StoredSession{
		Key:       sessionKey,
		Summary:   summary,
		Messages:  messages,
		CreatedAt: parseSQLiteTime(createdRaw),
		UpdatedAt: parseSQLiteTime(updatedRaw),
	}, nil
}

func (s *SQLiteStore) DeleteSession(ctx context.Context, sessionKey string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE session_key = ?`, sessionKey)
	if err != nil {
		return false, fmt.Errorf("memory: delete session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("memory: delete affected rows: %w", err)
	}
	return n > 0, nil
}

func (s *SQLiteStore) Compact(context.Context, string) error {
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func upsertSession(tx *sql.Tx, sessionKey, summary string, createdAt, updatedAt time.Time, preserveSummary bool) error {
	query := `
		INSERT INTO sessions (session_key, summary, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(session_key) DO UPDATE SET
			summary = CASE WHEN ? THEN COALESCE(NULLIF(excluded.summary, ''), sessions.summary) ELSE sessions.summary END,
			updated_at = excluded.updated_at
	`
	if _, err := tx.Exec(
		query,
		sessionKey,
		summary,
		createdAt.UTC().Format(time.RFC3339Nano),
		updatedAt.UTC().Format(time.RFC3339Nano),
		preserveSummary,
	); err != nil {
		return fmt.Errorf("memory: upsert session: %w", err)
	}
	return nil
}

func insertMessage(tx *sql.Tx, sessionKey string, msg providers.Message, createdAt time.Time) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("memory: encode message: %w", err)
	}
	if _, err = tx.Exec(
		`INSERT INTO messages (session_key, payload, created_at) VALUES (?, ?, ?)`,
		sessionKey,
		string(payload),
		createdAt.UTC().Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("memory: insert message: %w", err)
	}
	return nil
}

func parseSQLiteTime(raw string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err == nil {
		return t
	}
	return time.Time{}
}
