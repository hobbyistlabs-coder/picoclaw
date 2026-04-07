package boards

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ColumnTodo  = "todo"
	ColumnDoing = "in_progress"
	ColumnDone  = "done"
)

type Board struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Columns     []BoardColumn   `json:"columns"`
	Review      *ReviewSchedule `json:"review,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type BoardColumn struct {
	ID       string      `json:"id"`
	Key      string      `json:"key"`
	Name     string      `json:"name"`
	Position int         `json:"position"`
	Cards    []BoardCard `json:"cards"`
}

type BoardCard struct {
	ID          string    `json:"id"`
	BoardID     string    `json:"board_id"`
	ColumnID    string    `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ReviewSchedule struct {
	BoardID      string    `json:"board_id"`
	Enabled      bool      `json:"enabled"`
	EveryMinutes int       `json:"every_minutes"`
	CronJobID    string    `json:"cron_job_id"`
	Channel      string    `json:"channel"`
	ChatID       string    `json:"chat_id"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateBoardInput struct {
	Name        string
	Description string
	Columns     []BoardColumnInput
}

type UpdateCardInput struct {
	Title       *string
	Description *string
	ColumnID    *string
}

type BoardColumnInput struct {
	Key  string
	Name string
}

type Store struct {
	db *sql.DB
}

func DBPath(workspace string) string {
	return filepath.Join(workspace, "boards", "boards.db")
}

func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("boards: create dir: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("boards: open sqlite: %w", err)
	}
	if _, err = db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("boards: enable foreign keys: %w", err)
	}
	if err = createSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS boards (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS board_columns (
		id TEXT PRIMARY KEY,
		board_id TEXT NOT NULL,
		column_key TEXT NOT NULL,
		name TEXT NOT NULL,
		position INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY(board_id) REFERENCES boards(id) ON DELETE CASCADE,
		UNIQUE(board_id, column_key)
	);
	CREATE TABLE IF NOT EXISTS board_cards (
		id TEXT PRIMARY KEY,
		board_id TEXT NOT NULL,
		column_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		position INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY(board_id) REFERENCES boards(id) ON DELETE CASCADE,
		FOREIGN KEY(column_id) REFERENCES board_columns(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS board_review_schedules (
		board_id TEXT PRIMARY KEY,
		enabled INTEGER NOT NULL,
		every_minutes INTEGER NOT NULL,
		cron_job_id TEXT NOT NULL DEFAULT '',
		channel TEXT NOT NULL DEFAULT '',
		chat_id TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL,
		FOREIGN KEY(board_id) REFERENCES boards(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_board_columns_board ON board_columns(board_id, position);
	CREATE INDEX IF NOT EXISTS idx_board_cards_column ON board_cards(column_id, position);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("boards: create schema: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) EnsureDefaultBoard(ctx context.Context) (*Board, error) {
	boards, err := s.ListBoards(ctx)
	if err != nil {
		return nil, err
	}
	if len(boards) > 0 {
		return s.GetBoard(ctx, boards[0].ID)
	}
	return s.CreateBoard(ctx, CreateBoardInput{Name: "Main Board"})
}

func (s *Store) ListBoards(ctx context.Context) ([]Board, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM boards ORDER BY updated_at DESC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("boards: list boards: %w", err)
	}
	defer rows.Close()

	var out []Board
	for rows.Next() {
		var board Board
		var created string
		var updated string
		if err = rows.Scan(&board.ID, &board.Name, &board.Description, &created, &updated); err != nil {
			return nil, fmt.Errorf("boards: scan board: %w", err)
		}
		board.CreatedAt = parseTime(created)
		board.UpdatedAt = parseTime(updated)
		board.Columns = []BoardColumn{}
		out = append(out, board)
	}
	return out, rows.Err()
}

func (s *Store) CreateBoard(ctx context.Context, input CreateBoardInput) (*Board, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("boards: board name is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("boards: begin create board: %w", err)
	}
	defer tx.Rollback()

	now := nowText()
	boardID := newID()
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO boards (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, boardID, input.Name, input.Description, now, now); err != nil {
		return nil, fmt.Errorf("boards: insert board: %w", err)
	}

	defs := normalizeColumnInputs(input.Columns)
	for i, def := range defs {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO board_columns (id, board_id, column_key, name, position, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, newID(), boardID, def.Key, def.Name, i, now, now); err != nil {
			return nil, fmt.Errorf("boards: seed columns: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("boards: commit create board: %w", err)
	}
	return s.GetBoard(ctx, boardID)
}

func (s *Store) AddColumn(
	ctx context.Context, boardID string, input BoardColumnInput,
) (*BoardColumn, error) {
	name := trimColumnName(input.Name)
	if name == "" {
		return nil, fmt.Errorf("boards: column name is required")
	}
	key := sanitizeColumnKey(input.Key)
	if key == "" {
		key = sanitizeColumnKey(name)
	}
	position, err := s.nextColumnPosition(ctx, boardID)
	if err != nil {
		return nil, err
	}
	col := &BoardColumn{
		ID:       newID(),
		Key:      ensureUniqueColumnKey(ctx, s.db, boardID, key),
		Name:     name,
		Position: position,
		Cards:    []BoardCard{},
	}
	now := nowText()
	if _, err = s.db.ExecContext(ctx, `
		INSERT INTO board_columns (id, board_id, column_key, name, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, col.ID, boardID, col.Key, col.Name, col.Position, now, now); err != nil {
		return nil, fmt.Errorf("boards: add column: %w", err)
	}
	if err = s.touchBoard(ctx, boardID); err != nil {
		return nil, err
	}
	return col, nil
}

func (s *Store) GetBoard(ctx context.Context, boardID string) (*Board, error) {
	var board Board
	var created string
	var updated string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM boards WHERE id = ?
	`, boardID).Scan(&board.ID, &board.Name, &board.Description, &created, &updated)
	if err != nil {
		return nil, fmt.Errorf("boards: get board: %w", err)
	}
	board.CreatedAt = parseTime(created)
	board.UpdatedAt = parseTime(updated)

	columns, err := s.listColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	cards, err := s.listCards(ctx, boardID)
	if err != nil {
		return nil, err
	}
	for i := range columns {
		columns[i].Cards = cards[columns[i].ID]
		if columns[i].Cards == nil {
			columns[i].Cards = []BoardCard{}
		}
	}
	board.Columns = columns
	board.Review, _ = s.GetReviewSchedule(ctx, boardID)
	return &board, nil
}

func (s *Store) AddCard(
	ctx context.Context, boardID string, title, description string, columnID string,
) (*BoardCard, error) {
	if title == "" {
		return nil, fmt.Errorf("boards: card title is required")
	}
	if columnID == "" {
		var err error
		columnID, err = s.columnIDByKey(ctx, boardID, ColumnTodo)
		if err != nil {
			return nil, err
		}
	}
	position, err := s.nextCardPosition(ctx, columnID)
	if err != nil {
		return nil, err
	}
	card := &BoardCard{
		ID:          newID(),
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       title,
		Description: description,
		Position:    position,
	}
	now := nowText()
	if _, err = s.db.ExecContext(ctx, `
		INSERT INTO board_cards (id, board_id, column_id, title, description, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, card.ID, card.BoardID, card.ColumnID, card.Title, card.Description, card.Position, now, now); err != nil {
		return nil, fmt.Errorf("boards: add card: %w", err)
	}
	s.touchBoard(ctx, boardID)
	card.CreatedAt = parseTime(now)
	card.UpdatedAt = card.CreatedAt
	return card, nil
}

func (s *Store) UpdateCard(ctx context.Context, cardID string, input UpdateCardInput) (*BoardCard, error) {
	card, err := s.getCard(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		card.Title = *input.Title
	}
	if input.Description != nil {
		card.Description = *input.Description
	}
	moved := false
	if input.ColumnID != nil && *input.ColumnID != "" && *input.ColumnID != card.ColumnID {
		card.ColumnID = *input.ColumnID
		card.Position, err = s.nextCardPosition(ctx, card.ColumnID)
		if err != nil {
			return nil, err
		}
		moved = true
	}
	now := nowText()
	if _, err = s.db.ExecContext(ctx, `
		UPDATE board_cards
		SET title = ?, description = ?, column_id = ?, position = ?, updated_at = ?
		WHERE id = ?
	`, card.Title, card.Description, card.ColumnID, card.Position, now, card.ID); err != nil {
		return nil, fmt.Errorf("boards: update card: %w", err)
	}
	if moved {
		s.reindexColumn(ctx, card.ColumnID)
	}
	s.touchBoard(ctx, card.BoardID)
	card.UpdatedAt = parseTime(now)
	return card, nil
}

func (s *Store) DeleteCard(ctx context.Context, cardID string) error {
	card, err := s.getCard(ctx, cardID)
	if err != nil {
		return err
	}
	if _, err = s.db.ExecContext(ctx, `DELETE FROM board_cards WHERE id = ?`, cardID); err != nil {
		return fmt.Errorf("boards: delete card: %w", err)
	}
	s.reindexColumn(ctx, card.ColumnID)
	return s.touchBoard(ctx, card.BoardID)
}

func (s *Store) GetReviewSchedule(ctx context.Context, boardID string) (*ReviewSchedule, error) {
	var review ReviewSchedule
	var updated string
	err := s.db.QueryRowContext(ctx, `
		SELECT board_id, enabled, every_minutes, cron_job_id, channel, chat_id, updated_at
		FROM board_review_schedules WHERE board_id = ?
	`, boardID).Scan(
		&review.BoardID, &review.Enabled, &review.EveryMinutes, &review.CronJobID,
		&review.Channel, &review.ChatID, &updated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("boards: get review schedule: %w", err)
	}
	review.UpdatedAt = parseTime(updated)
	return &review, nil
}

func (s *Store) SaveReviewSchedule(ctx context.Context, review ReviewSchedule) error {
	now := nowText()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO board_review_schedules
		(board_id, enabled, every_minutes, cron_job_id, channel, chat_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(board_id) DO UPDATE SET
			enabled = excluded.enabled,
			every_minutes = excluded.every_minutes,
			cron_job_id = excluded.cron_job_id,
			channel = excluded.channel,
			chat_id = excluded.chat_id,
			updated_at = excluded.updated_at
	`, review.BoardID, review.Enabled, review.EveryMinutes, review.CronJobID, review.Channel, review.ChatID, now)
	if err != nil {
		return fmt.Errorf("boards: save review schedule: %w", err)
	}
	return s.touchBoard(ctx, review.BoardID)
}

func (s *Store) listColumns(ctx context.Context, boardID string) ([]BoardColumn, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, column_key, name, position
		FROM board_columns WHERE board_id = ?
		ORDER BY position ASC, created_at ASC
	`, boardID)
	if err != nil {
		return nil, fmt.Errorf("boards: list columns: %w", err)
	}
	defer rows.Close()
	var cols []BoardColumn
	for rows.Next() {
		var col BoardColumn
		if err = rows.Scan(&col.ID, &col.Key, &col.Name, &col.Position); err != nil {
			return nil, fmt.Errorf("boards: scan column: %w", err)
		}
		col.Cards = []BoardCard{}
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func (s *Store) listCards(ctx context.Context, boardID string) (map[string][]BoardCard, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, board_id, column_id, title, description, position, created_at, updated_at
		FROM board_cards WHERE board_id = ?
		ORDER BY position ASC, created_at ASC
	`, boardID)
	if err != nil {
		return nil, fmt.Errorf("boards: list cards: %w", err)
	}
	defer rows.Close()
	out := make(map[string][]BoardCard)
	for rows.Next() {
		var card BoardCard
		var created string
		var updated string
		if err = rows.Scan(
			&card.ID, &card.BoardID, &card.ColumnID, &card.Title,
			&card.Description, &card.Position, &created, &updated,
		); err != nil {
			return nil, fmt.Errorf("boards: scan card: %w", err)
		}
		card.CreatedAt = parseTime(created)
		card.UpdatedAt = parseTime(updated)
		out[card.ColumnID] = append(out[card.ColumnID], card)
	}
	for key := range out {
		sort.SliceStable(out[key], func(i, j int) bool {
			return out[key][i].Position < out[key][j].Position
		})
	}
	return out, rows.Err()
}

func (s *Store) getCard(ctx context.Context, cardID string) (*BoardCard, error) {
	var card BoardCard
	var created string
	var updated string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, board_id, column_id, title, description, position, created_at, updated_at
		FROM board_cards WHERE id = ?
	`, cardID).Scan(
		&card.ID, &card.BoardID, &card.ColumnID, &card.Title,
		&card.Description, &card.Position, &created, &updated,
	)
	if err != nil {
		return nil, fmt.Errorf("boards: get card: %w", err)
	}
	card.CreatedAt = parseTime(created)
	card.UpdatedAt = parseTime(updated)
	return &card, nil
}

func (s *Store) columnIDByKey(ctx context.Context, boardID, key string) (string, error) {
	var columnID string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM board_columns WHERE board_id = ? AND column_key = ?
	`, boardID, key).Scan(&columnID)
	if err != nil {
		return "", fmt.Errorf("boards: resolve column: %w", err)
	}
	return columnID, nil
}

func (s *Store) nextCardPosition(ctx context.Context, columnID string) (int, error) {
	var pos int
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), -1) + 1 FROM board_cards WHERE column_id = ?
	`, columnID).Scan(&pos)
	if err != nil {
		return 0, fmt.Errorf("boards: next position: %w", err)
	}
	return pos, nil
}

func (s *Store) nextColumnPosition(ctx context.Context, boardID string) (int, error) {
	var pos int
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), -1) + 1 FROM board_columns WHERE board_id = ?
	`, boardID).Scan(&pos)
	if err != nil {
		return 0, fmt.Errorf("boards: next column position: %w", err)
	}
	return pos, nil
}

func (s *Store) reindexColumn(ctx context.Context, columnID string) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id FROM board_cards WHERE column_id = ?
		ORDER BY position ASC, created_at ASC
	`, columnID)
	if err != nil {
		return
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		var cardID string
		if rows.Scan(&cardID) == nil {
			s.db.ExecContext(ctx, `UPDATE board_cards SET position = ? WHERE id = ?`, i, cardID)
			i++
		}
	}
	_ = rows.Err()
}

func (s *Store) touchBoard(ctx context.Context, boardID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE boards SET updated_at = ? WHERE id = ?
	`, nowText(), boardID)
	if err != nil {
		return fmt.Errorf("boards: touch board: %w", err)
	}
	return nil
}

func parseTime(v string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return time.Time{}
	}
	return t
}

func nowText() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func normalizeColumnInputs(inputs []BoardColumnInput) []BoardColumnInput {
	if len(inputs) == 0 {
		return []BoardColumnInput{
			{Key: ColumnTodo, Name: "Todo"},
			{Key: ColumnDoing, Name: "In Progress"},
			{Key: ColumnDone, Name: "Done"},
		}
	}
	seen := map[string]struct{}{}
	out := make([]BoardColumnInput, 0, len(inputs))
	for _, input := range inputs {
		name := trimColumnName(input.Name)
		if name == "" {
			continue
		}
		key := sanitizeColumnKey(input.Key)
		if key == "" {
			key = sanitizeColumnKey(name)
		}
		if key == "" {
			continue
		}
		key = dedupeColumnKey(seen, key)
		out = append(out, BoardColumnInput{Key: key, Name: name})
	}
	if len(out) == 0 {
		return normalizeColumnInputs(nil)
	}
	return out
}

func trimColumnName(name string) string {
	return strings.TrimSpace(name)
}

func sanitizeColumnKey(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func dedupeColumnKey(seen map[string]struct{}, key string) string {
	if _, ok := seen[key]; !ok {
		seen[key] = struct{}{}
		return key
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", key, i)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		return candidate
	}
}

func ensureUniqueColumnKey(ctx context.Context, db *sql.DB, boardID, key string) string {
	base := key
	for i := 1; ; i++ {
		var existing string
		err := db.QueryRowContext(ctx, `
			SELECT column_key FROM board_columns WHERE board_id = ? AND column_key = ?
		`, boardID, key).Scan(&existing)
		if err == sql.ErrNoRows {
			return key
		}
		if err != nil {
			return fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
		}
		key = fmt.Sprintf("%s-%d", base, i+1)
	}
}
