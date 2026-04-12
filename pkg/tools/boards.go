package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jane/pkg/boards"
	"jane/pkg/cron"
)

type BoardsTool struct {
	store       *boards.Store
	cronService *cron.CronService
}

func NewBoardsTool(store *boards.Store, cronService *cron.CronService) *BoardsTool {
	return &BoardsTool{store: store, cronService: cronService}
}

func (t *BoardsTool) Name() string { return "boards" }

func (t *BoardsTool) Description() string {
	return "Read and update kanban boards. Supports bulk moves to handle multiple tasks efficiently."
}

func (t *BoardsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
				"enum": []string{
					"list_boards", "get_board", "create_board", "add_card",
					"add_column", "update_card", "move_card", "bulk_move_cards",
					"delete_card", "set_review_schedule",
				},
			},
			"board_id":      map[string]any{"type": "string"},
			"card_id":       map[string]any{"type": "string"},
			"card_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"title":         map[string]any{"type": "string"},
			"description":   map[string]any{"type": "string"},
			"column_id":     map[string]any{"type": "string"},
			"enabled":       map[string]any{"type": "boolean"},
			"every_minutes": map[string]any{"type": "integer"},
			"channel":       map[string]any{"type": "string"},
			"chat_id":       map[string]any{"type": "string"},
		},
		"required": []string{"action"},
	}
}

// withRetries handles "database is locked" errors internally to prevent agent loops.
func (t *BoardsTool) withRetries(fn func() (*ToolResult, error)) *ToolResult {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		res, err := fn()
		if err == nil {
			return res
		}
		if strings.Contains(strings.ToLower(err.Error()), "locked") {
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		return ErrorResult(err.Error())
	}
	return ErrorResult("database remained locked after multiple retries")
}

func (t *BoardsTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	action, _ := args["action"].(string)
	switch action {
	case "list_boards":
		return t.listBoards(ctx)
	case "get_board":
		return t.getBoard(ctx, args)
	case "create_board":
		return t.createBoard(ctx, args)
	case "add_card":
		return t.addCard(ctx, args)
	case "add_column":
		return t.addColumn(ctx, args)
	case "update_card":
		return t.withRetries(func() (*ToolResult, error) {
			return t.updateCard(ctx, args, false), nil
		})
	case "move_card":
		return t.withRetries(func() (*ToolResult, error) {
			return t.updateCard(ctx, args, true), nil
		})
	case "bulk_move_cards":
		return t.bulkMoveCards(ctx, args)
	case "delete_card":
		return t.deleteCard(ctx, args)
	case "set_review_schedule":
		return t.setReviewSchedule(ctx, args)
	default:
		return ErrorResult(fmt.Sprintf("unknown boards action: %s", action))
	}
}

func (t *BoardsTool) RequiresApproval() bool { return false }

func (t *BoardsTool) bulkMoveCards(ctx context.Context, args map[string]any) *ToolResult {
	cardIDs, _ := args["card_ids"].([]any)
	columnID, _ := args["column_id"].(string)

	if len(cardIDs) == 0 || columnID == "" {
		return ErrorResult("card_ids and column_id are required for bulk moves")
	}

	return t.withRetries(func() (*ToolResult, error) {
		results := make(map[string]string)
		for _, idAny := range cardIDs {
			id := idAny.(string)
			input := boards.UpdateCardInput{ColumnID: &columnID}
			_, err := t.store.UpdateCard(ctx, id, input)
			if err != nil {
				results[id] = fmt.Sprintf("error: %s", err.Error())
			} else {
				results[id] = "success"
			}
		}
		return jsonResult(results), nil
	})
}

func (t *BoardsTool) listBoards(ctx context.Context) *ToolResult {
	items, err := t.store.ListBoards(ctx)
	if err != nil {
		return ErrorResult(err.Error())
	}
	return jsonResult(items)
}

func (t *BoardsTool) getBoard(ctx context.Context, args map[string]any) *ToolResult {
	boardID, _ := args["board_id"].(string)
	var board *boards.Board
	var err error

	if boardID == "" {
		board, err = t.store.EnsureDefaultBoard(ctx)
	} else {
		board, err = t.store.GetBoard(ctx, boardID)
	}

	if err != nil {
		return ErrorResult(err.Error())
	}
	return jsonResult(board)
}

func (t *BoardsTool) createBoard(ctx context.Context, args map[string]any) *ToolResult {
	title, _ := args["title"].(string)
	desc, _ := args["description"].(string)
	board, err := t.store.CreateBoard(ctx, boards.CreateBoardInput{Name: title, Description: desc})
	if err != nil {
		return ErrorResult(err.Error())
	}
	return jsonResult(board)
}

func (t *BoardsTool) addColumn(ctx context.Context, args map[string]any) *ToolResult {
	boardID, _ := args["board_id"].(string)
	title, _ := args["title"].(string)
	desc, _ := args["description"].(string)
	column, err := t.store.AddColumn(ctx, boardID, boards.BoardColumnInput{
		Key: desc, Name: title,
	})
	if err != nil {
		return ErrorResult(err.Error())
	}
	return jsonResult(column)
}

func (t *BoardsTool) addCard(ctx context.Context, args map[string]any) *ToolResult {
	boardID, _ := args["board_id"].(string)
	if boardID == "" {
		board, err := t.store.EnsureDefaultBoard(ctx)
		if err != nil {
			return ErrorResult(err.Error())
		}
		boardID = board.ID
	}
	title, _ := args["title"].(string)
	desc, _ := args["description"].(string)
	columnID, _ := args["column_id"].(string)
	card, err := t.store.AddCard(ctx, boardID, title, desc, columnID)
	if err != nil {
		return ErrorResult(err.Error())
	}
	return jsonResult(card)
}

func (t *BoardsTool) updateCard(ctx context.Context, args map[string]any, moveOnly bool) *ToolResult {
	cardID, _ := args["card_id"].(string)
	if cardID == "" {
		return ErrorResult("card_id is required")
	}
	input := boards.UpdateCardInput{}
	if !moveOnly {
		if title, ok := args["title"].(string); ok {
			input.Title = &title
		}
		if desc, ok := args["description"].(string); ok {
			input.Description = &desc
		}
	}
	if columnID, ok := args["column_id"].(string); ok && columnID != "" {
		input.ColumnID = &columnID
	}
	card, err := t.store.UpdateCard(ctx, cardID, input)
	if err != nil {
		return ErrorResult(err.Error())
	}
	return jsonResult(card)
}

func (t *BoardsTool) deleteCard(ctx context.Context, args map[string]any) *ToolResult {
	cardID, _ := args["card_id"].(string)
	if err := t.store.DeleteCard(ctx, cardID); err != nil {
		return ErrorResult(err.Error())
	}
	return SilentResult("card deleted")
}

func (t *BoardsTool) setReviewSchedule(ctx context.Context, args map[string]any) *ToolResult {
	boardID, _ := args["board_id"].(string)
	if boardID == "" {
		board, err := t.store.EnsureDefaultBoard(ctx)
		if err != nil {
			return ErrorResult(err.Error())
		}
		boardID = board.ID
	}
	enabled, _ := args["enabled"].(bool)
	every, _ := args["every_minutes"].(float64)
	channel, _ := args["channel"].(string)
	chatID, _ := args["chat_id"].(string)
	review, err := boards.SyncReviewSchedule(ctx, t.store, t.cronService, boardID, boards.ReviewScheduleInput{
		Enabled: enabled, EveryMinutes: int(every), Channel: channel, ChatID: chatID,
	})
	if err != nil {
		return ErrorResult(err.Error())
	}
	return jsonResult(review)
}

func jsonResult(v any) *ToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ErrorResult(err.Error())
	}
	return SilentResult(string(data))
}