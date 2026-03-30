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

// picoSessionPrefix is the key prefix used by the gateway's routing for Pico
// channel sessions. The full key format is:
//
//	agent:main:pico:direct:pico:<session-uuid>
//
// The sanitized filename replaces ':' with '_', so on disk it becomes:
//
//	agent_main_pico_direct_pico_<session-uuid>.json
const (
	picoSessionPrefix          = "agent:main:pico:direct:pico:"
	sanitizedPicoSessionPrefix = "agent_main_pico_direct_pico_"
	maxSessionTitleRunes       = 60
)

// extractPicoSessionID extracts the session UUID from a full session key.
// Returns the UUID and true if the key matches the Pico session pattern.
func extractPicoSessionID(key string) (string, bool) {
	if strings.HasPrefix(key, picoSessionPrefix) {
		return strings.TrimPrefix(key, picoSessionPrefix), true
	}
	return "", false
}

func extractPicoSessionIDFromSanitizedKey(key string) (string, bool) {
	if strings.HasPrefix(key, sanitizedPicoSessionPrefix) {
		return strings.TrimPrefix(key, sanitizedPicoSessionPrefix), true
	}
	return "", false
}

func sanitizeSessionKey(key string) string {
	return strings.ReplaceAll(key, ":", "_")
}

func (h *Handler) readLegacySession(dir, sessionID string) (sessionFile, error) {
	path := filepath.Join(dir, sanitizeSessionKey(picoSessionPrefix+sessionID)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return sessionFile{}, err
	}

	var sess sessionFile
	if err := json.Unmarshal(data, &sess); err != nil {
		return sessionFile{}, err
	}
	return sess, nil
}

func buildSessionListItem(sessionID string, sess sessionFile) sessionListItem {
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
	costKnown := true

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
		} else {
			costKnown = false
		}
	}

	if !hasUsage && metrics.ToolCalls == 0 {
		return nil
	}

	metrics.HasEstimatedCost = hasUsage && costKnown
	if !metrics.HasEstimatedCost {
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
		stored, listErr := store.ListSessions(r.Context(), picoSessionPrefix, limit, offset)
		if listErr != nil {
			http.Error(w, "failed to load sessions", http.StatusInternalServerError)
			return
		}
		items := make([]sessionListItem, 0, len(stored))
		for _, sess := range stored {
			sessionID, ok := extractPicoSessionID(sess.Key)
			if !ok {
				continue
			}
			file := storedToSessionFile(sess)
			if isEmptySession(file) {
				continue
			}
			items = append(items, buildSessionListItem(sessionID, file))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]sessionListItem{})
		return
	}
	items := []sessionListItem{}
	seen := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") ||
			strings.HasSuffix(entry.Name(), ".meta.json") ||
			strings.HasSuffix(entry.Name(), ".migrated") {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			continue
		}
		var sess sessionFile
		if unmarshalErr := json.Unmarshal(data, &sess); unmarshalErr != nil || isEmptySession(sess) {
			continue
		}
		sessionID, ok := extractPicoSessionID(sess.Key)
		if !ok {
			continue
		}
		if _, exists := seen[sessionID]; exists {
			continue
		}
		seen[sessionID] = struct{}{}
		items = append(items, buildSessionListItem(sessionID, sess))
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
		stored, getErr := store.GetSession(r.Context(), picoSessionPrefix+sessionID)
		if getErr == nil {
			sess := storedToSessionFile(stored)
			if isEmptySession(sess) {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			writeSessionResponse(w, sessionID, sess)
			return
		}
		if !errors.Is(getErr, os.ErrNotExist) {
			http.Error(w, "failed to parse session", http.StatusInternalServerError)
			return
		}
	}

	sess, err := h.readLegacySession(dir, sessionID)
	if err != nil || isEmptySession(sess) {
		if errors.Is(err, os.ErrNotExist) || isEmptySession(sess) {
			http.Error(w, "session not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to parse session", http.StatusInternalServerError)
		}
		return
	}
	writeSessionResponse(w, sessionID, sess)
}

func writeSessionResponse(w http.ResponseWriter, sessionID string, sess sessionFile) {
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
		removed, deleteErr := store.DeleteSession(r.Context(), picoSessionPrefix+sessionID)
		if deleteErr != nil {
			http.Error(w, "failed to delete session", http.StatusInternalServerError)
			return
		}
		if removed {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	base := filepath.Join(dir, sanitizeSessionKey(picoSessionPrefix+sessionID))
	removed := false
	for _, path := range []string{base + ".json", base + ".jsonl", base + ".meta.json"} {
		if removeErr := os.Remove(path); removeErr != nil {
			if os.IsNotExist(removeErr) {
				continue
			}
			http.Error(w, "failed to delete session", http.StatusInternalServerError)
			return
		}
		removed = true
	}
	if !removed {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
