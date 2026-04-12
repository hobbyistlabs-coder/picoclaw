package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"jane/pkg/config"
)

func TestWorkspaceBootstrapRoundTrip(t *testing.T) {
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = workspace
	configPath := filepath.Join(dir, "config.json")
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "AGENTS.md"), []byte("# Agent"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	handler := NewHandler(configPath)
	mux := http.NewServeMux()
	handler.registerWorkspaceRoutes(mux)

	getReq := httptest.NewRequest(http.MethodGet, "/api/workspace/bootstrap", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status: %d body=%s", getRec.Code, getRec.Body.String())
	}

	var listed workspaceFilesResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(listed.Files) != 4 {
		t.Fatalf("expected 4 files, got %d", len(listed.Files))
	}

	body := bytes.NewBufferString(`{"content":"# Soul\r\nHello"}`)
	putReq := httptest.NewRequest(http.MethodPut, "/api/workspace/bootstrap/SOUL.md", body)
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	mux.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("put status: %d body=%s", putRec.Code, putRec.Body.String())
	}

	saved, err := os.ReadFile(filepath.Join(workspace, "SOUL.md"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(saved) != "# Soul\nHello\n" {
		t.Fatalf("unexpected saved content: %q", string(saved))
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/workspace/bootstrap/SOUL.md/history", nil)
	historyRec := httptest.NewRecorder()
	mux.ServeHTTP(historyRec, historyReq)
	if historyRec.Code != http.StatusOK {
		t.Fatalf("history status: %d body=%s", historyRec.Code, historyRec.Body.String())
	}

	var history promptHistoryResponse
	if err := json.Unmarshal(historyRec.Body.Bytes(), &history); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if len(history.Revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(history.Revisions))
	}
	if history.Revisions[0].Content != "# Soul\nHello\n" {
		t.Fatalf("unexpected history content: %q", history.Revisions[0].Content)
	}
}

func TestWorkspaceBootstrapRejectsUnknownFile(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	configPath := filepath.Join(dir, "config.json")
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	handler := NewHandler(configPath)
	mux := http.NewServeMux()
	handler.registerWorkspaceRoutes(mux)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/workspace/bootstrap/not-allowed.md",
		bytes.NewBufferString(`{"content":"x"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
