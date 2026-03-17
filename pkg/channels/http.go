// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
// Copyright (c) 2026 PicoClaw contributors

package channels

import (
	"net/http"
	"time"

	"jane/pkg/health"
	"jane/pkg/logger"
)

// SetupHTTPServer creates a shared HTTP server with the given listen address.
// It registers health endpoints from the health server and discovers channels
// that implement WebhookHandler and/or HealthChecker to register their handlers.
func (m *Manager) SetupHTTPServer(addr string, healthServer *health.Server) {
	m.mux = http.NewServeMux()

	// Register health endpoints
	if healthServer != nil {
		healthServer.RegisterOnMux(m.mux)
	}

	// Discover and register webhook handlers and health checkers
	for name, ch := range m.channels {
		if wh, ok := ch.(WebhookHandler); ok {
			m.mux.Handle(wh.WebhookPath(), wh)
			logger.InfoCF("channels", "Webhook handler registered", map[string]any{
				"channel": name,
				"path":    wh.WebhookPath(),
			})
		}
		if hc, ok := ch.(HealthChecker); ok {
			m.mux.HandleFunc(hc.HealthPath(), hc.HealthHandler)
			logger.InfoCF("channels", "Health endpoint registered", map[string]any{
				"channel": name,
				"path":    hc.HealthPath(),
			})
		}
	}

	m.httpServer = &http.Server{
		Addr:              addr,
		Handler:           m.mux,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
