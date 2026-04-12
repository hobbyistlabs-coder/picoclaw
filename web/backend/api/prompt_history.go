package api

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"jane/pkg/config"
)

type promptRevision struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

type promptHistoryResponse struct {
	Revisions []promptRevision `json:"revisions"`
}

var promptHistorySanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func promptHistoryDir(workspace string) string {
	return filepath.Join(workspace, ".jane", "prompt-history")
}

func promptHistoryPath(workspace, kind, name string) string {
	safeKind := promptHistorySanitizer.ReplaceAllString(kind, "-")
	safeName := promptHistorySanitizer.ReplaceAllString(name, "-")
	return filepath.Join(promptHistoryDir(workspace), safeKind+"__"+safeName+".jsonl")
}

func appendPromptRevision(workspace, kind, name, content string) error {
	if strings.TrimSpace(workspace) == "" {
		return nil
	}
	path := promptHistoryPath(workspace, kind, name)
	existing, err := loadPromptRevisions(workspace, kind, name)
	if err != nil {
		return err
	}
	if len(existing) > 0 && existing[0].Content == content {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	entry := promptRevision{
		ID:        randomRevisionID(),
		Kind:      kind,
		Name:      name,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Content:   content,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(line, '\n'))
	return err
}

func loadPromptRevisions(workspace, kind, name string) ([]promptRevision, error) {
	path := promptHistoryPath(workspace, kind, name)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []promptRevision{}, nil
		}
		return nil, err
	}
	defer file.Close()

	revisions := []promptRevision{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry promptRevision
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, err
		}
		revisions = append(revisions, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	slices.Reverse(revisions)
	return revisions, nil
}

func randomRevisionID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func recordPersonaPromptRevisions(prevCfg, nextCfg *config.Config) error {
	workspace := nextCfg.WorkspacePath()
	prevByID := map[string]string{}
	for _, persona := range prevCfg.Agents.List {
		prevByID[persona.ID] = persona.SystemPrompt
	}
	for _, persona := range nextCfg.Agents.List {
		if prevByID[persona.ID] == persona.SystemPrompt {
			continue
		}
		if err := appendPromptRevision(
			workspace,
			"persona_system_prompt",
			persona.ID,
			persona.SystemPrompt,
		); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) handleGetPersonaPromptHistory(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}
	personaID := strings.TrimSpace(r.PathValue("id"))
	if personaID == "" {
		http.Error(w, "persona id is required", http.StatusBadRequest)
		return
	}
	revisions, err := loadPromptRevisions(
		cfg.WorkspacePath(),
		"persona_system_prompt",
		personaID,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load prompt history: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(promptHistoryResponse{Revisions: revisions})
}
