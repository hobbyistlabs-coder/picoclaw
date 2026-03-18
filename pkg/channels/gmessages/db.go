package gmessages

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Message struct {
	MessageID      string
	ConversationID string
	SenderName     string
	SenderNumber   string
	Body           string
	MediaID        string
	MimeType       string
	DecryptionKey  string
	Reactions      string
	ReplyToID      string
	TimestampMS    int64
	Status         string
	IsFromMe       bool
}

type Conversation struct {
	ConversationID string
	Name           string
	IsGroup        bool
	Participants   string
	LastMessageTS  int64
	UnreadCount    int
}

type Contact struct {
	Number string
	Name   string
}

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	// Use modernc.org/sqlite (pure Go) and WAL mode for concurrency
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		message_id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL,
		sender_name TEXT,
		sender_number TEXT,
		body TEXT,
		media_id TEXT,
		mime_type TEXT,
		decryption_key TEXT,
		reactions TEXT,
		reply_to_id TEXT,
		timestamp_ms INTEGER NOT NULL,
		status TEXT,
		is_from_me BOOLEAN
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp_ms DESC);
	CREATE INDEX IF NOT EXISTS idx_messages_body ON messages(body) WHERE body IS NOT NULL;

	CREATE TABLE IF NOT EXISTS conversations (
		conversation_id TEXT PRIMARY KEY,
		name TEXT,
		is_group BOOLEAN,
		participants TEXT,
		last_message_ts INTEGER,
		unread_count INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_conversations_last_message_ts ON conversations(last_message_ts DESC);

	CREATE TABLE IF NOT EXISTS contacts (
		number TEXT PRIMARY KEY,
		name TEXT
	);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *Store) UpsertMessage(m *Message) error {
	query := `
	INSERT INTO messages (
		message_id, conversation_id, sender_name, sender_number, body,
		media_id, mime_type, decryption_key, reactions, reply_to_id,
		timestamp_ms, status, is_from_me
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(message_id) DO UPDATE SET
		status=excluded.status,
		reactions=excluded.reactions,
		body=excluded.body,
		media_id=excluded.media_id,
		mime_type=excluded.mime_type
	`
	_, err := s.db.Exec(query,
		m.MessageID, m.ConversationID, m.SenderName, m.SenderNumber, m.Body,
		m.MediaID, m.MimeType, m.DecryptionKey, m.Reactions, m.ReplyToID,
		m.TimestampMS, m.Status, m.IsFromMe,
	)
	return err
}

func (s *Store) DeleteTmpMessages(conversationID string) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM messages WHERE conversation_id = ? AND message_id LIKE 'tmp_%'`, conversationID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) UpsertConversation(c *Conversation) error {
	query := `
	INSERT INTO conversations (
		conversation_id, name, is_group, participants, last_message_ts, unread_count
	) VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(conversation_id) DO UPDATE SET
		name=excluded.name,
		is_group=excluded.is_group,
		participants=excluded.participants,
		last_message_ts=excluded.last_message_ts,
		unread_count=excluded.unread_count
	`
	_, err := s.db.Exec(query, c.ConversationID, c.Name, c.IsGroup, c.Participants, c.LastMessageTS, c.UnreadCount)
	return err
}

func (s *Store) GetConversationIDByNumber(number string) (string, error) {
	// Look for a 1:1 conversation that has this number in participants
	query := `
	SELECT conversation_id
	FROM conversations
	WHERE is_group = false
	AND participants LIKE ?
	LIMIT 1
	`
	var convID string
	err := s.db.QueryRow(query, "%"+number+"%").Scan(&convID)
	if err == sql.ErrNoRows {
		return "", nil // not found
	}
	return convID, err
}

func (s *Store) UpsertContact(number, name string) error {
	query := `
	INSERT INTO contacts (number, name) VALUES (?, ?)
	ON CONFLICT(number) DO UPDATE SET name=excluded.name
	`
	_, err := s.db.Exec(query, number, name)
	return err
}
