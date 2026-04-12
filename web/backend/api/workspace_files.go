package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"jane/pkg/config"
)

type workspaceFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Exists  bool   `json:"exists"`
}

type workspaceFilesResponse struct {
	Files []workspaceFile `json:"files"`
}

var workspaceBootstrapFiles = []string{
	"AGENTS.md",
	"IDENTITY.md",
	"SOUL.md",
	"USER.md",
}

func (h *Handler) registerWorkspaceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/workspace/bootstrap", h.handleGetWorkspaceFiles)
	mux.HandleFunc("GET /api/workspace/bootstrap/{name}/history", h.handleGetWorkspaceFileHistory)
	mux.HandleFunc("PUT /api/workspace/bootstrap/{name}", h.handleUpdateWorkspaceFile)
}

func (h *Handler) handleGetWorkspaceFiles(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	files := make([]workspaceFile, 0, len(workspaceBootstrapFiles))
	for _, name := range workspaceBootstrapFiles {
		file, err := readWorkspaceFile(cfg.WorkspacePath(), name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read workspace file: %v", err), http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(workspaceFilesResponse{Files: files})
}

func (h *Handler) handleUpdateWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	name, err := validateWorkspaceFileName(r.PathValue("name"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(cfg.WorkspacePath(), 0o755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to prepare workspace: %v", err), http.StatusInternalServerError)
		return
	}

	targetPath := filepath.Join(cfg.WorkspacePath(), name)
	content := normalizeWorkspaceMarkdown(payload.Content)
	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write workspace file: %v", err), http.StatusInternalServerError)
		return
	}
	if err := appendPromptRevision(cfg.WorkspacePath(), "workspace_file", name, content); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store prompt history: %v", err), http.StatusInternalServerError)
		return
	}

	file, err := readWorkspaceFile(cfg.WorkspacePath(), name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read saved workspace file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(file)
}

func (h *Handler) handleGetWorkspaceFileHistory(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}
	name, err := validateWorkspaceFileName(r.PathValue("name"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	revisions, err := loadPromptRevisions(cfg.WorkspacePath(), "workspace_file", name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load prompt history: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(promptHistoryResponse{Revisions: revisions})
}

func readWorkspaceFile(workspace, name string) (workspaceFile, error) {
	path := filepath.Join(workspace, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return workspaceFile{Name: name, Path: path, Exists: false}, nil
		}
		return workspaceFile{}, err
	}
	return workspaceFile{
		Name:    name,
		Path:    path,
		Content: string(data),
		Exists:  true,
	}, nil
}

func validateWorkspaceFileName(name string) (string, error) {
	for _, allowed := range workspaceBootstrapFiles {
		if name == allowed {
			return name, nil
		}
	}
	return "", fmt.Errorf("unsupported workspace file %q", name)
}

func normalizeWorkspaceMarkdown(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	if normalized != "" && !strings.HasSuffix(normalized, "\n") {
		normalized += "\n"
	}
	return normalized
}
