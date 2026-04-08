package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"jane/pkg/boards"
	"jane/pkg/config"
)

type boardRunResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
}

func (h *Handler) handleRunCardAgent(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load config: %v", err), http.StatusInternalServerError)
		return
	}
	store, cleanup, err := h.openBoardsStore()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanup()
	if err := validateBoardCard(r.Context(), store, r.PathValue("id"), r.PathValue("cardID")); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if !cfg.Channels.Pico.Enabled || strings.TrimSpace(cfg.Channels.Pico.Token) == "" {
		http.Error(w, "pico channel is not configured", http.StatusConflict)
		return
	}

	sessionID, err := h.dispatchBoardPrompt(
		r.Context(),
		cfg,
		boards.BuildCardActionPrompt(r.PathValue("id"), r.PathValue("cardID")),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, boardRunResponse{Status: "queued", SessionID: sessionID})
}

func (h *Handler) dispatchBoardPrompt(
	ctx context.Context,
	cfg *config.Config,
	prompt string,
) (string, error) {
	wsURL, err := h.gatewayPicoWSURL(cfg)
	if err != nil {
		return "", err
	}

	sessionID := generateSecureToken()
	dialURL := wsURL + "?session_id=" + url.QueryEscape(sessionID)
	header := http.Header{"Authorization": []string{"Bearer " + cfg.Channels.Pico.Token}}
	conn, res, err := websocket.DefaultDialer.DialContext(ctx, dialURL, header)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return "", fmt.Errorf("connect to gateway pico websocket: %w", err)
	}

	if err := conn.WriteJSON(map[string]any{
		"type":       "message.send",
		"session_id": sessionID,
		"payload":    map[string]any{"content": prompt},
	}); err != nil {
		conn.Close()
		return "", fmt.Errorf("send board prompt: %w", err)
	}

	go drainPicoConn(conn, 45*time.Second)
	return sessionID, nil
}

func (h *Handler) gatewayPicoWSURL(cfg *config.Config) (string, error) {
	if h.gatewayBaseURL != "" {
		baseURL, err := url.Parse(h.gatewayBaseURL)
		if err != nil {
			return "", fmt.Errorf("invalid gateway base url: %w", err)
		}
		switch baseURL.Scheme {
		case "https":
			baseURL.Scheme = "wss"
		default:
			baseURL.Scheme = "ws"
		}
		baseURL.Path = "/pico/ws"
		baseURL.RawQuery = ""
		baseURL.Fragment = ""
		return baseURL.String(), nil
	}

	host := gatewayProbeHost(h.effectiveGatewayBindHost(cfg))
	return fmt.Sprintf("ws://%s:%d/pico/ws", host, cfg.Gateway.Port), nil
}

func drainPicoConn(conn *websocket.Conn, maxDuration time.Duration) {
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(maxDuration))
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func validateBoardCard(ctx context.Context, store *boards.Store, boardID, cardID string) error {
	board, err := store.GetBoard(ctx, boardID)
	if err != nil {
		return err
	}
	for _, column := range board.Columns {
		for _, card := range column.Cards {
			if card.ID == cardID {
				return nil
			}
		}
	}
	return fmt.Errorf("card %s not found on board %s", cardID, boardID)
}
