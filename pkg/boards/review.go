package boards

import (
	"context"
	"fmt"
	"strings"

	"jane/pkg/cron"
)

type ReviewScheduleInput struct {
	Enabled      bool
	EveryMinutes int
	Channel      string
	ChatID       string
}

func SyncReviewSchedule(
	ctx context.Context,
	store *Store,
	cronService *cron.CronService,
	boardID string,
	input ReviewScheduleInput,
) (*ReviewSchedule, error) {
	if input.Enabled && input.EveryMinutes < 5 {
		return nil, fmt.Errorf("boards: review schedule must be at least 5 minutes")
	}
	current, err := store.GetReviewSchedule(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if current != nil && current.CronJobID != "" {
		cronService.RemoveJob(current.CronJobID)
	}

	review := &ReviewSchedule{
		BoardID:      boardID,
		Enabled:      input.Enabled,
		EveryMinutes: input.EveryMinutes,
		Channel:      normalizeChannel(input.Channel),
		ChatID:       normalizeChatID(input.ChatID),
	}
	if input.Enabled {
		everyMS := int64(input.EveryMinutes) * 60 * 1000
		job, err := cronService.AddJob(
			fmt.Sprintf("board-review:%s", boardID),
			cron.CronSchedule{Kind: "every", EveryMS: &everyMS},
			BuildReviewPrompt(boardID),
			false,
			review.Channel,
			review.ChatID,
		)
		if err != nil {
			return nil, fmt.Errorf("boards: add review job: %w", err)
		}
		review.CronJobID = job.ID
	}
	if err := store.SaveReviewSchedule(ctx, *review); err != nil {
		return nil, err
	}
	return store.GetReviewSchedule(ctx, boardID)
}

func BuildReviewPrompt(boardID string) string {
	var b strings.Builder
	b.WriteString("Review the kanban board and update progress where needed.\n")
	b.WriteString("Use the boards tool.\n")
	b.WriteString("Board ID: ")
	b.WriteString(boardID)
	b.WriteString("\n")
	b.WriteString(
		"Summarize overdue or blocked work, then move or annotate cards if the current context justifies it.",
	)
	return b.String()
}

func BuildCardActionPrompt(boardID, cardID string) string {
	var b strings.Builder
	b.WriteString("Review and act on this kanban card now.\n")
	b.WriteString("Use the boards tool.\n")
	b.WriteString("Board ID: ")
	b.WriteString(boardID)
	b.WriteString("\n")
	b.WriteString("Card ID: ")
	b.WriteString(cardID)
	b.WriteString("\n")
	b.WriteString(
		"Update progress, move the card if appropriate, and refine the card details with blockers or next steps when justified by the current context.",
	)
	return b.String()
}

func normalizeChannel(channel string) string {
	if channel == "" {
		return "cli"
	}
	return channel
}

func normalizeChatID(chatID string) string {
	if chatID == "" {
		return "direct"
	}
	return chatID
}
