package api

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"jane/pkg/auth"
	"jane/pkg/config"
	"jane/pkg/logger"
)

func renderOAuthCallbackPage(w http.ResponseWriter, flowID, status, title, errMsg string) {
	payload := map[string]string{
		"type":   "jane-ai-oauth-result",
		"flowId": flowID,
		"status": status,
	}
	if errMsg != "" {
		payload["error"] = errMsg
	}
	payloadJSON, _ := json.Marshal(payload)

	message := title
	if errMsg != "" {
		message = fmt.Sprintf("%s: %s", title, errMsg)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status == oauthFlowSuccess {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	_, _ = fmt.Fprintf(
		w,
		"<!doctype html><html><head><meta charset=\"utf-8\"><title>Jane AI OAuth</title></head><body><script>(function(){var payload=%s;var hasOpener=false;try{if(window.opener&&!window.opener.closed){window.opener.postMessage(payload,window.location.origin);hasOpener=true}}catch(e){}var target='/credentials?oauth_flow_id='+encodeURIComponent(payload.flowId||'')+'&oauth_status='+encodeURIComponent(payload.status||'');setTimeout(function(){if(hasOpener){window.close();return}window.location.replace(target)},800)})();</script><div style=\"font-family:Inter,system-ui,sans-serif;padding:24px\"><h2>%s</h2><p>%s</p><p>You can close this window.</p></div></body></html>",
		string(payloadJSON),
		html.EscapeString(title),
		html.EscapeString(message),
	)
}

func normalizeOAuthProvider(raw string) (string, error) {
	provider := strings.ToLower(strings.TrimSpace(raw))
	switch provider {
	case "antigravity":
		return oauthProviderGoogleAntigravity, nil
	case oauthProviderOpenAI, oauthProviderAnthropic, oauthProviderGoogleAntigravity:
		return provider, nil
	default:
		return "", fmt.Errorf("unsupported provider %q", raw)
	}
}

func isOAuthMethodSupported(provider, method string) bool {
	methods := oauthProviderMethods[provider]
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func oauthConfigForProvider(provider string) (auth.OAuthProviderConfig, error) {
	switch provider {
	case oauthProviderOpenAI:
		return auth.OpenAIOAuthConfig(), nil
	case oauthProviderGoogleAntigravity:
		return auth.GoogleAntigravityOAuthConfig(), nil
	default:
		return auth.OAuthProviderConfig{}, fmt.Errorf("provider %q does not support browser oauth", provider)
	}
}

func oauthMethodTokenOrOAuth(method string) string {
	if method == oauthMethodToken {
		return oauthMethodToken
	}
	return "oauth"
}

func buildOAuthRedirectURI(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.Split(forwarded, ",")[0]
	}
	return fmt.Sprintf("%s://%s/oauth/callback", scheme, r.Host)
}

func flowToResponse(flow *oauthFlow) oauthFlowResponse {
	resp := oauthFlowResponse{
		FlowID:   flow.ID,
		Provider: flow.Provider,
		Method:   flow.Method,
		Status:   flow.Status,
		Error:    flow.Error,
	}
	if !flow.ExpiresAt.IsZero() {
		resp.ExpiresAt = flow.ExpiresAt.Format(time.RFC3339)
	}
	if flow.Method == oauthMethodDeviceCode {
		resp.UserCode = flow.UserCode
		resp.VerifyURL = flow.VerifyURL
		resp.Interval = flow.Interval
	}
	return resp
}

func (h *Handler) persistCredentialAndConfig(provider, authMethod string, cred *auth.AuthCredential) error {
	if cred == nil {
		return fmt.Errorf("empty credential")
	}

	cp := *cred
	cp.Provider = provider
	if cp.AuthMethod == "" {
		cp.AuthMethod = authMethod
	}

	if provider == oauthProviderGoogleAntigravity {
		if cp.Email == "" {
			email, err := oauthFetchGoogleUserEmailFunc(cp.AccessToken)
			if err != nil {
				logger.WarnCF("oauth", "could not fetch google email", map[string]any{"error": err.Error()})
			} else {
				cp.Email = email
			}
		}
		if cp.ProjectID == "" {
			projectID, err := oauthFetchAntigravityProject(cp.AccessToken)
			if err != nil {
				logger.WarnCF("oauth", "could not fetch antigravity project id", map[string]any{"error": err.Error()})
			} else {
				cp.ProjectID = projectID
			}
		}
	}

	if err := oauthSetCredential(provider, &cp); err != nil {
		return fmt.Errorf("saving credential: %w", err)
	}
	if err := h.syncProviderAuthMethod(provider, authMethod); err != nil {
		return fmt.Errorf("syncing provider auth config: %w", err)
	}
	return nil
}

func (h *Handler) syncProviderAuthMethod(provider, authMethod string) error {
	cfg, err := oauthLoadConfig(h.configPath)
	if err != nil {
		return err
	}

	switch provider {
	case oauthProviderOpenAI:
		cfg.Providers.OpenAI.AuthMethod = authMethod
	case oauthProviderAnthropic:
		cfg.Providers.Anthropic.AuthMethod = authMethod
	case oauthProviderGoogleAntigravity:
		cfg.Providers.Antigravity.AuthMethod = authMethod
	default:
		return fmt.Errorf("unsupported provider %q", provider)
	}

	found := false
	for i := range cfg.ModelList {
		if modelBelongsToProvider(provider, cfg.ModelList[i].Model) {
			cfg.ModelList[i].AuthMethod = authMethod
			found = true
		}
	}

	if !found && authMethod != "" {
		cfg.ModelList = append(cfg.ModelList, defaultModelConfigForProvider(provider, authMethod))
	}

	return oauthSaveConfig(h.configPath, cfg)
}

func modelBelongsToProvider(provider, model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))
	switch provider {
	case oauthProviderOpenAI:
		return lower == "openai" || strings.HasPrefix(lower, "openai/")
	case oauthProviderAnthropic:
		return lower == "anthropic" || strings.HasPrefix(lower, "anthropic/")
	case oauthProviderGoogleAntigravity:
		return lower == "antigravity" ||
			lower == "google-antigravity" ||
			strings.HasPrefix(lower, "antigravity/") ||
			strings.HasPrefix(lower, "google-antigravity/")
	default:
		return false
	}
}

func defaultModelConfigForProvider(provider, authMethod string) config.ModelConfig {
	switch provider {
	case oauthProviderOpenAI:
		return config.ModelConfig{
			ModelName:  "gpt-5.4",
			Model:      "openai/gpt-5.4",
			AuthMethod: authMethod,
		}
	case oauthProviderAnthropic:
		return config.ModelConfig{
			ModelName:  "claude-sonnet-4.6",
			Model:      "anthropic/claude-sonnet-4.6",
			AuthMethod: authMethod,
		}
	case oauthProviderGoogleAntigravity:
		return config.ModelConfig{
			ModelName:  "gemini-flash",
			Model:      "antigravity/gemini-3-flash",
			AuthMethod: authMethod,
		}
	default:
		return config.ModelConfig{}
	}
}

func fetchGoogleUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("userinfo request failed: %s", string(body))
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return "", err
	}
	if userInfo.Email == "" {
		return "", fmt.Errorf("empty email in userinfo response")
	}
	return userInfo.Email, nil
}
