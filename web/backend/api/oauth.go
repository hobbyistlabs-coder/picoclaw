package api

import (
	"net/http"
)

// registerOAuthRoutes binds OAuth login/logout endpoints to the ServeMux.
func (h *Handler) registerOAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/oauth/providers", h.handleListOAuthProviders)
	mux.HandleFunc("POST /api/oauth/login", h.handleOAuthLogin)
	mux.HandleFunc("GET /api/oauth/flows/{id}", h.handleGetOAuthFlow)
	mux.HandleFunc("POST /api/oauth/flows/{id}/poll", h.handlePollOAuthFlow)
	mux.HandleFunc("POST /api/oauth/logout", h.handleOAuthLogout)
	mux.HandleFunc("GET /oauth/callback", h.handleOAuthCallback)
}
