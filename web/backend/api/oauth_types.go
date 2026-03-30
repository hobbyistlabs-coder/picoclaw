package api

import (
	"time"
)

type oauthFlow struct {
	ID           string
	Provider     string
	Method       string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ExpiresAt    time.Time
	Error        string
	CodeVerifier string
	OAuthState   string
	RedirectURI  string
	DeviceAuthID string
	UserCode     string
	VerifyURL    string
	Interval     int
}

type oauthProviderStatus struct {
	Provider    string   `json:"provider"`
	DisplayName string   `json:"display_name"`
	Methods     []string `json:"methods"`
	LoggedIn    bool     `json:"logged_in"`
	Status      string   `json:"status"`
	AuthMethod  string   `json:"auth_method,omitempty"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
	AccountID   string   `json:"account_id,omitempty"`
	Email       string   `json:"email,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
}

type oauthFlowResponse struct {
	FlowID    string `json:"flow_id"`
	Provider  string `json:"provider"`
	Method    string `json:"method"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Error     string `json:"error,omitempty"`
	UserCode  string `json:"user_code,omitempty"`
	VerifyURL string `json:"verify_url,omitempty"`
	Interval  int    `json:"interval,omitempty"`
}
