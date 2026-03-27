package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"jane/pkg/auth"
)

func (h *Handler) handleListOAuthProviders(w http.ResponseWriter, r *http.Request) {
	providersResp := make([]oauthProviderStatus, 0, len(oauthProviderOrder))

	for _, provider := range oauthProviderOrder {
		cred, err := oauthGetCredential(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to load credentials: %v", err), http.StatusInternalServerError)
			return
		}

		item := oauthProviderStatus{
			Provider:    provider,
			DisplayName: oauthProviderLabels[provider],
			Methods:     oauthProviderMethods[provider],
			Status:      "not_logged_in",
		}
		if cred != nil {
			item.LoggedIn = true
			item.AuthMethod = cred.AuthMethod
			item.AccountID = cred.AccountID
			item.Email = cred.Email
			item.ProjectID = cred.ProjectID
			if !cred.ExpiresAt.IsZero() {
				item.ExpiresAt = cred.ExpiresAt.Format(time.RFC3339)
			}
			switch {
			case cred.IsExpired():
				item.Status = "expired"
			case cred.NeedsRefresh():
				item.Status = "needs_refresh"
			default:
				item.Status = "connected"
			}
		}

		providersResp = append(providersResp, item)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"providers": providersResp,
	})
}

func (h *Handler) handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Provider string `json:"provider"`
		Method   string `json:"method"`
		Token    string `json:"token"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	provider, err := normalizeOAuthProvider(req.Provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	method := strings.ToLower(strings.TrimSpace(req.Method))
	if !isOAuthMethodSupported(provider, method) {
		http.Error(
			w,
			fmt.Sprintf("unsupported login method %q for provider %q", method, provider),
			http.StatusBadRequest,
		)
		return
	}

	switch method {
	case oauthMethodToken:
		token := strings.TrimSpace(req.Token)
		if token == "" {
			http.Error(w, "token is required", http.StatusBadRequest)
			return
		}

		cred := &auth.AuthCredential{
			AccessToken: token,
			Provider:    provider,
			AuthMethod:  oauthMethodToken,
		}
		if err := h.persistCredentialAndConfig(provider, oauthMethodToken, cred); err != nil {
			http.Error(w, fmt.Sprintf("token login failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"provider": provider,
			"method":   method,
		})
		return

	case oauthMethodDeviceCode:
		cfg := auth.OpenAIOAuthConfig()
		info, err := oauthRequestDeviceCode(cfg)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to request device code: %v", err), http.StatusInternalServerError)
			return
		}

		now := oauthNow()
		flow := &oauthFlow{
			ID:           newOAuthFlowID(),
			Provider:     provider,
			Method:       method,
			Status:       oauthFlowPending,
			CreatedAt:    now,
			UpdatedAt:    now,
			ExpiresAt:    now.Add(oauthDeviceCodeFlowTTL),
			DeviceAuthID: info.DeviceAuthID,
			UserCode:     info.UserCode,
			VerifyURL:    info.VerifyURL,
			Interval:     info.Interval,
		}
		h.storeOAuthFlow(flow)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":     "ok",
			"provider":   provider,
			"method":     method,
			"flow_id":    flow.ID,
			"user_code":  flow.UserCode,
			"verify_url": flow.VerifyURL,
			"interval":   flow.Interval,
			"expires_at": flow.ExpiresAt.Format(time.RFC3339),
		})
		return

	case oauthMethodBrowser:
		cfg, err := oauthConfigForProvider(provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pkce, err := oauthGeneratePKCE()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate PKCE: %v", err), http.StatusInternalServerError)
			return
		}
		state, err := oauthGenerateState()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate state: %v", err), http.StatusInternalServerError)
			return
		}

		redirectURI := buildOAuthRedirectURI(r)
		authURL := oauthBuildAuthorizeURL(cfg, pkce, state, redirectURI)

		now := oauthNow()
		flow := &oauthFlow{
			ID:           newOAuthFlowID(),
			Provider:     provider,
			Method:       method,
			Status:       oauthFlowPending,
			CreatedAt:    now,
			UpdatedAt:    now,
			ExpiresAt:    now.Add(oauthBrowserFlowTTL),
			CodeVerifier: pkce.CodeVerifier,
			OAuthState:   state,
			RedirectURI:  redirectURI,
		}
		h.storeOAuthFlow(flow)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":     "ok",
			"provider":   provider,
			"method":     method,
			"flow_id":    flow.ID,
			"auth_url":   authURL,
			"expires_at": flow.ExpiresAt.Format(time.RFC3339),
		})
		return
	default:
		http.Error(w, "unsupported login method", http.StatusBadRequest)
	}
}

func (h *Handler) handleGetOAuthFlow(w http.ResponseWriter, r *http.Request) {
	flowID := strings.TrimSpace(r.PathValue("id"))
	if flowID == "" {
		http.Error(w, "missing flow id", http.StatusBadRequest)
		return
	}

	flow, ok := h.getOAuthFlow(flowID)
	if !ok {
		http.Error(w, "flow not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(flowToResponse(flow))
}

func (h *Handler) handlePollOAuthFlow(w http.ResponseWriter, r *http.Request) {
	flowID := strings.TrimSpace(r.PathValue("id"))
	if flowID == "" {
		http.Error(w, "missing flow id", http.StatusBadRequest)
		return
	}

	flow, ok := h.getOAuthFlow(flowID)
	if !ok {
		http.Error(w, "flow not found", http.StatusNotFound)
		return
	}

	if flow.Method != oauthMethodDeviceCode {
		http.Error(w, "flow does not support polling", http.StatusBadRequest)
		return
	}
	if flow.Status != oauthFlowPending {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(flowToResponse(flow))
		return
	}

	cfg := auth.OpenAIOAuthConfig()
	cred, err := oauthPollDeviceCodeOnce(cfg, flow.DeviceAuthID, flow.UserCode)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "pending") {
			updated, _ := h.getOAuthFlow(flowID)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(flowToResponse(updated))
			return
		}
		h.setOAuthFlowError(flowID, fmt.Sprintf("device code poll failed: %v", err))
		updated, _ := h.getOAuthFlow(flowID)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(flowToResponse(updated))
		return
	}
	if cred == nil {
		updated, _ := h.getOAuthFlow(flowID)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(flowToResponse(updated))
		return
	}

	if err := h.persistCredentialAndConfig(flow.Provider, oauthMethodTokenOrOAuth(flow.Method), cred); err != nil {
		h.setOAuthFlowError(flowID, fmt.Sprintf("failed to save credential: %v", err))
		updated, _ := h.getOAuthFlow(flowID)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(flowToResponse(updated))
		return
	}

	h.setOAuthFlowSuccess(flowID)
	updated, _ := h.getOAuthFlow(flowID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(flowToResponse(updated))
}

func (h *Handler) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" {
		renderOAuthCallbackPage(w, "", oauthFlowError, "Missing state", "missing_state")
		return
	}

	flow, ok := h.getOAuthFlowByState(state)
	if !ok {
		renderOAuthCallbackPage(w, "", oauthFlowError, "OAuth flow not found", "flow_not_found")
		return
	}

	if flow.Status != oauthFlowPending {
		renderOAuthCallbackPage(w, flow.ID, flow.Status, "Flow already completed", flow.Error)
		return
	}

	if errMsg := strings.TrimSpace(r.URL.Query().Get("error")); errMsg != "" {
		if desc := strings.TrimSpace(r.URL.Query().Get("error_description")); desc != "" {
			errMsg += ": " + desc
		}
		h.setOAuthFlowError(flow.ID, errMsg)
		renderOAuthCallbackPage(w, flow.ID, oauthFlowError, "Authorization failed", errMsg)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		h.setOAuthFlowError(flow.ID, "missing authorization code")
		renderOAuthCallbackPage(w, flow.ID, oauthFlowError, "Missing authorization code", "missing_code")
		return
	}

	cfg, err := oauthConfigForProvider(flow.Provider)
	if err != nil {
		h.setOAuthFlowError(flow.ID, err.Error())
		renderOAuthCallbackPage(w, flow.ID, oauthFlowError, "Unsupported provider", err.Error())
		return
	}

	cred, err := oauthExchangeCodeForTokens(cfg, code, flow.CodeVerifier, flow.RedirectURI)
	if err != nil {
		h.setOAuthFlowError(flow.ID, fmt.Sprintf("token exchange failed: %v", err))
		renderOAuthCallbackPage(w, flow.ID, oauthFlowError, "Token exchange failed", err.Error())
		return
	}

	if err := h.persistCredentialAndConfig(flow.Provider, oauthMethodTokenOrOAuth(flow.Method), cred); err != nil {
		h.setOAuthFlowError(flow.ID, fmt.Sprintf("failed to save credential: %v", err))
		renderOAuthCallbackPage(w, flow.ID, oauthFlowError, "Failed to save credential", err.Error())
		return
	}

	h.setOAuthFlowSuccess(flow.ID)
	renderOAuthCallbackPage(w, flow.ID, oauthFlowSuccess, "Authentication successful", "")
}

func (h *Handler) handleOAuthLogout(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Provider string `json:"provider"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	provider, err := normalizeOAuthProvider(req.Provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := oauthDeleteCredential(provider); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete credential: %v", err), http.StatusInternalServerError)
		return
	}
	if err := h.syncProviderAuthMethod(provider, ""); err != nil {
		http.Error(w, fmt.Sprintf("failed to update config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"provider": provider,
	})
}
