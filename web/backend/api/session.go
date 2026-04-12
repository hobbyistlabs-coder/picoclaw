package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"jane/pkg/bus"
	"jane/pkg/config"
	"jane/pkg/memory"
	"jane/pkg/providers"
	"jane/pkg/routing"
	"jane/pkg/runtimepaths"
)

// registerSessionRoutes binds session list and detail endpoints to the ServeMux.
func (h *Handler) registerSessionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sessions", h.handleListSessions)
	mux.HandleFunc("GET /api/sessions/{id}", h.handleGetSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", h.handleDeleteSession)
}

// sessionFile mirrors the on-disk session JSON structure from pkg/session.
type sessionFile struct {
	Key      string              `json:"key"`
	Messages []providers.Message `json:"messages"`
	Summary  string              `json:"summary,omitempty"`
	Created  time.Time           `json:"created"`
	Updated  time.Time           `json:"updated"`
}

// sessionListItem is a lightweight summary returned by GET /api/sessions.
type sessionListItem struct {
	AgentID      string `json:"agent_id"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	Preview      string `json:"preview"`
	MessageCount int    `json:"message_count"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

type sessionToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Kind      string         `json:"kind,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

const (
	picoSessionPrefix    = "agent:main:pico:direct:pico:"
	maxSessionTitleRunes = 60
)

type legacySessionRecord struct {
	basePath  string
	agentID   string
	sessionID string
	session   sessionFile
}

func extractPicoSessionMeta(key string) (string, string, bool) {
	parsed := routing.ParseAgentSessionKey(key)
	if parsed == nil {
		return "", "", false
	}
	if !strings.HasPrefix(parsed.Rest, "pico:direct:pico:") {
		return "", "", false
	}
	return parsed.AgentID, strings.TrimPrefix(parsed.Rest, "pico:direct:pico:"), true
}

func picoSessionKey(agentID, sessionID string) string {
	return strings.ToLower(routing.BuildAgentPeerSessionKey(routing.SessionKeyParams{
		AgentID: agentID,
		Channel: "pico",
		Peer:    &routing.RoutePeer{Kind: "direct", ID: "pico:" + sessionID},
		DMScope: routing.DMScopePerChannelPeer,
	}))
}

func sanitizeSessionKey(key string) string {
	key = strings.ReplaceAll(key, ":", "_")
	return filepath.Base(filepath.Clean(key))
}

func (h *Handler) readLegacySessions(dir string) ([]legacySessionRecord, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	records := make([]legacySessionRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || strings.HasSuffix(entry.Name(), ".meta.json") || strings.HasSuffix(entry.Name(), ".migrated") {
			continue
		}

		basePath := filepath.Join(dir, strings.TrimSuffix(entry.Name(), ".json"))
		data, readErr := os.ReadFile(basePath + ".json")
		if readErr != nil {
			continue
		}

		var sess sessionFile
		if unmarshalErr := json.Unmarshal(data, &sess); unmarshalErr != nil || isEmptySession(sess) {
			continue
		}

		agentID, sessionID, ok := extractPicoSessionMeta(sess.Key)
		if !ok {
			continue
		}

		records = append(records, legacySessionRecord{
			basePath:  basePath,
			agentID:   agentID,
			sessionID: sessionID,
			session:   sess,
		})
	}

	return records, nil
}

func buildSessionListItem(sessionID, agentID string, sess sessionFile) sessionListItem {
	preview := ""
	for _, msg := range sess.Messages {
		if msg.Role == "user" && strings.TrimSpace(msg.Content) != "" {
			preview = msg.Content
			break
		}
	}
	title := strings.TrimSpace(sess.Summary)
	if title == "" {
		title = preview
	}

	title = truncateRunes(title, maxSessionTitleRunes)
	preview = truncateRunes(preview, maxSessionTitleRunes)

	if preview == "" {
		preview = "(empty)"
	}
	if title == "" {
		title = preview
	}

	validMessageCount := 0
	for _, msg := range sess.Messages {
		if (msg.Role == "user" || msg.Role == "assistant") && strings.TrimSpace(msg.Content) != "" {
			validMessageCount++
		}
	}

	return sessionListItem{
		AgentID:      agentID,
		ID:           sessionID,
		Title:        title,
		Preview:      preview,
		MessageCount: validMessageCount,
		Created:      sess.Created.Format(time.RFC3339),
		Updated:      sess.Updated.Format(time.RFC3339),
	}
}

func isEmptySession(sess sessionFile) bool {
	return len(sess.Messages) == 0 && strings.TrimSpace(sess.Summary) == ""
}

func truncateRunes(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "..."
}

func buildSessionMetrics(messages []providers.Message) *bus.MessageMetrics {
	metrics := &bus.MessageMetrics{}
	hasUsage := false
	hasEstimatedCost := false

	for _, msg := range messages {
		metrics.ToolCalls += len(msg.ToolCalls)
		if msg.Usage == nil {
			continue
		}

		hasUsage = true
		metrics.PromptTokens += msg.Usage.PromptTokens
		metrics.CompletionTokens += msg.Usage.CompletionTokens
		metrics.TotalTokens += msg.Usage.TotalTokens
		if msg.Usage.HasEstimatedCost {
			metrics.EstimatedCostUSD += msg.Usage.EstimatedCostUSD
			hasEstimatedCost = true
		}
	}

	if !hasUsage && metrics.ToolCalls == 0 {
		return nil
	}

	metrics.HasEstimatedCost = hasEstimatedCost
	if !hasEstimatedCost {
		metrics.EstimatedCostUSD = 0
	}

	return metrics
}

func (h *Handler) sqliteSessionStore(dir string) (*memory.SQLiteStore, error) {
	store, err := memory.NewSQLiteStore(memory.SQLitePath(dir))
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if _, err = memory.MigrateFromJSON(ctx, dir, store); err != nil {
		store.Close()
		return nil, err
	}
	if _, err = memory.MigrateFromJSONL(ctx, dir, store); err != nil {
		store.Close()
		return nil, err
	}
	return store, nil
}

func storedToSessionFile(session memory.StoredSession) sessionFile {
	return sessionFile{
		Key:      session.Key,
		Messages: session.Messages,
		Summary:  session.Summary,
		Created:  session.CreatedAt,
		Updated:  session.UpdatedAt,
	}
}

// sessionsDir resolves the path to the gateway's session storage directory.
// It reads the workspace from config, falling back to the resolved app home workspace.
func (h *Handler) sessionsDir() (string, error) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		return "", err
	}

	workspace := cfg.Agents.Defaults.Workspace
	if workspace == "" {
		workspace = filepath.Join(runtimepaths.HomeDir(), "workspace")
	}

	// Expand ~ prefix
	if len(workspace) > 0 && workspace[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(workspace) > 1 && workspace[1] == '/' {
			workspace = home + workspace[1:]
		} else {
			workspace = home
		}
	}

	return filepath.Join(workspace, "sessions"), nil
}

// handleListSessions returns a list of Pico session summaries.
//
//	GET /api/sessions
func (h *Handler) handleListSessions(w http.ResponseWriter, r *http.Request) {
	dir, err := h.sessionsDir()
	if err != nil {
		http.Error(w, "failed to resolve sessions directory", http.StatusInternalServerError)
		return
	}

	offset := 0
	limit := 20
	if val, convErr := strconv.Atoi(r.URL.Query().Get("offset")); convErr == nil && val >= 0 {
		offset = val
	}
	if val, convErr := strconv.Atoi(r.URL.Query().Get("limit")); convErr == nil && val > 0 {
		limit = val
	}

	store, err := h.sqliteSessionStore(dir)
	if err == nil {
		defer store.Close()
		stored, listErr := store.ListSessions(r.Context(), "agent:", 1000, 0)
		if listErr != nil {
			http.Error(w, "failed to load sessions", http.StatusInternalServerError)
			return
		}
		items := make([]sessionListItem, 0, len(stored))
		for _, sess := range stored {
			agentID, sessionID, ok := extractPicoSessionMeta(sess.Key)
			if !ok {
				continue
			}
			file := storedToSessionFile(sess)
			if isEmptySession(file) {
				continue
			}
			items = append(items, buildSessionListItem(sessionID, agentID, file))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Updated > items[j].Updated })
		if offset >= len(items) {
			items = []sessionListItem{}
		} else {
			end := offset + limit
			if end > len(items) {
				end = len(items)
			}
			items = items[offset:end]
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
		return
	}

	records, err := h.readLegacySessions(dir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]sessionListItem{})
		return
	}
	items := []sessionListItem{}
	seen := make(map[string]struct{})
	for _, record := range records {
		if _, exists := seen[record.sessionID]; exists {
			continue
		}
		seen[record.sessionID] = struct{}{}
		items = append(items, buildSessionListItem(record.sessionID, record.agentID, record.session))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Updated > items[j].Updated })
	if offset >= len(items) {
		items = []sessionListItem{}
	} else {
		end := offset + limit
		if end > len(items) {
			end = len(items)
		}
		items = items[offset:end]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// handleGetSession returns the full message history for a specific session.
//
//	GET /api/sessions/{id}
func (h *Handler) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}

	dir, err := h.sessionsDir()
	if err != nil {
		http.Error(w, "failed to resolve sessions directory", http.StatusInternalServerError)
		return
	}

	store, err := h.sqliteSessionStore(dir)
	if err == nil {
		defer store.Close()
		stored, listErr := store.ListSessions(r.Context(), "agent:", 1000, 0)
		if listErr == nil {
			for _, item := range stored {
				agentID, currentSessionID, ok := extractPicoSessionMeta(item.Key)
				if !ok || currentSessionID != sessionID {
					continue
				}
				sess := storedToSessionFile(item)
				if isEmptySession(sess) {
					continue
				}
				writeSessionResponse(w, sessionID, agentID, sess)
				return
			}
		} else if !errors.Is(listErr, os.ErrNotExist) {
			http.Error(w, "failed to parse session", http.StatusInternalServerError)
			return
		}
	}

	records, readErr := h.readLegacySessions(dir)
	if readErr == nil {
		for _, record := range records {
			if record.sessionID != sessionID {
				continue
			}
			writeSessionResponse(w, sessionID, record.agentID, record.session)
			return
		}
	}
	http.Error(w, "session not found", http.StatusNotFound)
}

func writeSessionResponse(w http.ResponseWriter, sessionID, agentID string, sess sessionFile) {
	type chatMessage struct {
		Role             string              `json:"role"`
		Content          string              `json:"content"`
		ReasoningContent string              `json:"reasoning_content,omitempty"`
		ToolCalls        []sessionToolCall   `json:"tool_calls,omitempty"`
		Metrics          *bus.MessageMetrics `json:"metrics,omitempty"`
	}

	messages := make([]chatMessage, 0, len(sess.Messages))
	var pendingReasoning string
	var pendingTools []sessionToolCall
	for _, msg := range sess.Messages {
		if msg.Role == "assistant" {
			if msg.ReasoningContent != "" {
				pendingReasoning = msg.ReasoningContent
			}
			if len(msg.ToolCalls) > 0 {
				pendingTools = append(pendingTools, buildToolCalls(msg.ToolCalls)...)
			}
		}

		if (msg.Role == "user" || msg.Role == "assistant") && strings.TrimSpace(msg.Content) != "" {
			var metrics *bus.MessageMetrics
			if msg.Usage != nil {
				metrics = &bus.MessageMetrics{
					ToolCalls:        len(msg.ToolCalls),
					PromptTokens:     msg.Usage.PromptTokens,
					CompletionTokens: msg.Usage.CompletionTokens,
					TotalTokens:      msg.Usage.TotalTokens,
					EstimatedCostUSD: msg.Usage.EstimatedCostUSD,
					HasEstimatedCost: msg.Usage.HasEstimatedCost,
				}
			}
			messages = append(messages, chatMessage{
				Role:             msg.Role,
				Content:          msg.Content,
				ReasoningContent: pendingReasoning,
				ToolCalls:        pendingTools,
				Metrics:          metrics,
			})
			pendingReasoning = ""
			pendingTools = nil
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"agent_id": agentID,
		"id":       sessionID,
		"messages": messages,
		"metrics":  buildSessionMetrics(sess.Messages),
		"summary":  sess.Summary,
		"created":  sess.Created.Format(time.RFC3339),
		"updated":  sess.Updated.Format(time.RFC3339),
	})
}

func buildToolCalls(calls []providers.ToolCall) []sessionToolCall {
	out := make([]sessionToolCall, 0, len(calls))
	for _, call := range calls {
		args := call.Arguments
		name := call.Name
		if name == "" && call.Function != nil {
			name = call.Function.Name
		}
		if args == nil && call.Function != nil && call.Function.Arguments != "" {
			var parsed map[string]any
			if err := json.Unmarshal([]byte(call.Function.Arguments), &parsed); err == nil {
				args = parsed
			}
		}
		out = append(out, sessionToolCall{
			ID:        call.ID,
			Name:      name,
			Kind:      sessionToolKind(name),
			Arguments: args,
		})
	}
	return out
}

func sessionToolKind(name string) string {
	switch {
	case name == "spawn" || name == "subagent":
		return "subagent"
	case name == "mcp2cli" || strings.HasPrefix(name, "mcp_"):
		return "mcp"
	default:
		return "tool"
	}
}

// handleDeleteSession deletes a specific session.
//
//	DELETE /api/sessions/{id}
func (h *Handler) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}

	dir, err := h.sessionsDir()
	if err != nil {
		http.Error(w, "failed to resolve sessions directory", http.StatusInternalServerError)
		return
	}

	store, err := h.sqliteSessionStore(dir)
	if err == nil {
		defer store.Close()
		stored, listErr := store.ListSessions(r.Context(), "agent:", 1000, 0)
		if listErr != nil && !errors.Is(listErr, os.ErrNotExist) {
			http.Error(w, "failed to delete session", http.StatusInternalServerError)
			return
		}
		removedAny := false
		for _, item := range stored {
			_, currentSessionID, ok := extractPicoSessionMeta(item.Key)
			if !ok || currentSessionID != sessionID {
				continue
			}
			removed, deleteErr := store.DeleteSession(r.Context(), item.Key)
			if deleteErr != nil {
				http.Error(w, "failed to delete session", http.StatusInternalServerError)
				return
			}
			removedAny = removedAny || removed
		}
		if removedAny {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	removed := false
	records, _ := h.readLegacySessions(dir)
	for _, record := range records {
		if record.sessionID != sessionID {
			continue
		}
		for _, path := range []string{
			record.basePath + ".json",
			record.basePath + ".jsonl",
			record.basePath + ".meta.json",
		} {
			if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
				http.Error(w, "failed to delete session", http.StatusInternalServerError)
				return
			}
			removed = true
		}
	}
	if !removed {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
