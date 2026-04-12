package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"jane/pkg/boards"
	"jane/pkg/config"
	"jane/pkg/cron"
)

type boardCardRequest struct {
	BoardID     string `json:"board_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ColumnID    string `json:"column_id"`
}

type boardColumnRequest struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type boardCreateRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Columns     []boardColumnRequest `json:"columns"`
}

type boardCardPatchRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ColumnID    *string `json:"column_id"`
}

type boardReviewRequest struct {
	Enabled      bool   `json:"enabled"`
	EveryMinutes int    `json:"every_minutes"`
	Channel      string `json:"channel"`
	ChatID       string `json:"chat_id"`
}

func (h *Handler) registerBoardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/boards", h.handleListBoards)
	mux.HandleFunc("POST /api/boards", h.handleCreateBoard)
	mux.HandleFunc("GET /api/boards/{id}", h.handleGetBoard)
	mux.HandleFunc("POST /api/boards/{id}/columns", h.handleCreateColumn)
	mux.HandleFunc("POST /api/boards/{id}/cards", h.handleCreateCard)
	mux.HandleFunc("POST /api/boards/{id}/cards/{cardID}/run", h.handleRunCardAgent)
	mux.HandleFunc("PATCH /api/boards/{id}/cards/{cardID}", h.handleUpdateCard)
	mux.HandleFunc("DELETE /api/boards/{id}/cards/{cardID}", h.handleDeleteCard)
	mux.HandleFunc("PUT /api/boards/{id}/review", h.handleSetBoardReview)
}

func (h *Handler) handleListBoards(w http.ResponseWriter, r *http.Request) {
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()

	if _, err = store.EnsureDefaultBoard(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items, err := store.ListBoards(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, items)
}

func (h *Handler) handleCreateBoard(w http.ResponseWriter, r *http.Request) {
	var req boardCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()
	columns := make([]boards.BoardColumnInput, 0, len(req.Columns))
	for _, col := range req.Columns {
		columns = append(columns, boards.BoardColumnInput{
			Key: col.Key, Name: col.Name,
		})
	}
	board, err := store.CreateBoard(r.Context(), boards.CreateBoardInput{
		Name: req.Name, Description: req.Description, Columns: columns,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, board)
}

func (h *Handler) handleCreateColumn(w http.ResponseWriter, r *http.Request) {
	var req boardColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()

	column, err := store.AddColumn(r.Context(), r.PathValue("id"), boards.BoardColumnInput{
		Key: req.Key, Name: req.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, column)
}

func (h *Handler) handleGetBoard(w http.ResponseWriter, r *http.Request) {
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()

	boardID := r.PathValue("id")
	var board *boards.Board
	if boardID == "default" {
		board, err = store.EnsureDefaultBoard(r.Context())
	} else {
		board, err = store.GetBoard(r.Context(), boardID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, board)
}

func (h *Handler) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	var req boardCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()

	card, err := store.AddCard(r.Context(), r.PathValue("id"), req.Title, req.Description, req.ColumnID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, card)
}

func (h *Handler) handleUpdateCard(w http.ResponseWriter, r *http.Request) {
	var req boardCardPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()

	input := boards.UpdateCardInput{}
	input.Title = req.Title
	input.Description = req.Description
	input.ColumnID = req.ColumnID
	card, err := store.UpdateCard(r.Context(), r.PathValue("cardID"), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, card)
}

func (h *Handler) handleDeleteCard(w http.ResponseWriter, r *http.Request) {
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()

	if err = store.DeleteCard(r.Context(), r.PathValue("cardID")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) handleSetBoardReview(w http.ResponseWriter, r *http.Request) {
	var req boardReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()

	cronService, err := h.openCronService()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	review, err := boards.SyncReviewSchedule(
		context.Background(),
		store,
		cronService,
		r.PathValue("id"),
		boards.ReviewScheduleInput{
			Enabled: req.Enabled, EveryMinutes: req.EveryMinutes, Channel: req.Channel, ChatID: req.ChatID,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, review)
}

func (h *Handler) openBoardsStore() (*boards.Store, func(), error) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}
	store, err := boards.NewStore(boards.DBPath(cfg.WorkspacePath()))
	if err != nil {
		return nil, nil, err
	}
	return store, func() { store.Close() }, nil
}

func (h *Handler) openCronService() (*cron.CronService, error) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	storePath := filepath.Join(cfg.WorkspacePath(), "cron", "jobs.json")
	return cron.NewCronService(storePath, nil), nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
