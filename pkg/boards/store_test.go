package boards

import (
	"context"
	"path/filepath"
	"testing"

	"jane/pkg/cron"
)

func TestStoreDefaultBoardAndCards(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "boards.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	board, err := store.EnsureDefaultBoard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(board.Columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(board.Columns))
	}

	card, err := store.AddCard(context.Background(), board.ID, "Ship board", "wire api", "")
	if err != nil {
		t.Fatal(err)
	}
	next := board.Columns[1].ID
	title := "Ship kanban board"
	desc := "wire api and UI"
	card, err = store.UpdateCard(context.Background(), card.ID, UpdateCardInput{
		Title: &title, Description: &desc, ColumnID: &next,
	})
	if err != nil {
		t.Fatal(err)
	}
	if card.ColumnID != next {
		t.Fatalf("expected card move to %s, got %s", next, card.ColumnID)
	}
}

func TestSyncReviewSchedule(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "boards.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	board, err := store.EnsureDefaultBoard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cronService := cron.NewCronService(filepath.Join(dir, "jobs.json"), nil)

	review, err := SyncReviewSchedule(ctx, store, cronService, board.ID, ReviewScheduleInput{
		Enabled: true, EveryMinutes: 15,
	})
	if err != nil {
		t.Fatal(err)
	}
	if review.CronJobID == "" {
		t.Fatal("expected cron job id")
	}

	review, err = SyncReviewSchedule(ctx, store, cronService, board.ID, ReviewScheduleInput{
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if review.Enabled {
		t.Fatal("expected review schedule disabled")
	}
}
