package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleOpenRouterCatalog_ForwardsFilters(t *testing.T) {
	origURL := openRouterModelsURL
	t.Cleanup(func() { openRouterModelsURL = origURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("output_modalities"); got != "image" {
			t.Fatalf("output_modalities = %q, want image", got)
		}
		if got := r.URL.Query().Get("supported_parameters"); got != "tools" {
			t.Fatalf("supported_parameters = %q, want tools", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"id": "openai/gpt-4.1-mini", "name": "GPT-4.1 Mini"}},
		})
	}))
	t.Cleanup(server.Close)
	openRouterModelsURL = server.URL

	h := NewHandler(t.TempDir() + "/config.json")
	req := httptest.NewRequest(http.MethodGet, "/api/models/openrouter/catalog?output_modalities=image&supported_parameters=tools", nil)
	rec := httptest.NewRecorder()

	h.handleOpenRouterCatalog(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
