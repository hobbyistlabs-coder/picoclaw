package api

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) handleOpenRouterCatalog(w http.ResponseWriter, r *http.Request) {
	models, err := fetchOpenRouterCatalog(
		r.Context(),
		r.URL.Query().Get("output_modalities"),
		r.URL.Query().Get("supported_parameters"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":  models,
		"total": len(models),
	})
}
