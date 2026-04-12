package api

import (
	"time"
)

func (h *Handler) storeOAuthFlow(flow *oauthFlow) {
	now := oauthNow()
	h.oauthMu.Lock()
	defer h.oauthMu.Unlock()

	h.gcOAuthFlowsLocked(now)
	h.oauthFlows[flow.ID] = flow
	if flow.OAuthState != "" {
		h.oauthState[flow.OAuthState] = flow.ID
	}
}

func (h *Handler) getOAuthFlow(flowID string) (*oauthFlow, bool) {
	now := oauthNow()
	h.oauthMu.Lock()
	defer h.oauthMu.Unlock()

	h.gcOAuthFlowsLocked(now)
	flow, ok := h.oauthFlows[flowID]
	if !ok {
		return nil, false
	}
	cp := *flow
	return &cp, true
}

func (h *Handler) getOAuthFlowByState(state string) (*oauthFlow, bool) {
	now := oauthNow()
	h.oauthMu.Lock()
	defer h.oauthMu.Unlock()

	h.gcOAuthFlowsLocked(now)
	flowID, ok := h.oauthState[state]
	if !ok {
		return nil, false
	}
	flow, ok := h.oauthFlows[flowID]
	if !ok {
		delete(h.oauthState, state)
		return nil, false
	}
	cp := *flow
	return &cp, true
}

func (h *Handler) setOAuthFlowSuccess(flowID string) {
	now := oauthNow()
	h.oauthMu.Lock()
	defer h.oauthMu.Unlock()

	flow, ok := h.oauthFlows[flowID]
	if !ok {
		return
	}
	flow.Status = oauthFlowSuccess
	flow.Error = ""
	flow.UpdatedAt = now
	if flow.OAuthState != "" {
		delete(h.oauthState, flow.OAuthState)
	}
}

func (h *Handler) setOAuthFlowError(flowID, errMsg string) {
	now := oauthNow()
	h.oauthMu.Lock()
	defer h.oauthMu.Unlock()

	flow, ok := h.oauthFlows[flowID]
	if !ok {
		return
	}
	flow.Status = oauthFlowError
	flow.Error = errMsg
	flow.UpdatedAt = now
	if flow.OAuthState != "" {
		delete(h.oauthState, flow.OAuthState)
	}
}

func (h *Handler) gcOAuthFlowsLocked(now time.Time) {
	for id, flow := range h.oauthFlows {
		if flow.Status == oauthFlowPending && !flow.ExpiresAt.IsZero() &&
			now.After(flow.ExpiresAt) {
			flow.Status = oauthFlowExpired
			flow.Error = "flow expired"
			flow.UpdatedAt = now
			if flow.OAuthState != "" {
				delete(h.oauthState, flow.OAuthState)
			}
		}

		if flow.Status != oauthFlowPending && now.Sub(flow.UpdatedAt) > oauthTerminalFlowGC {
			if flow.OAuthState != "" {
				delete(h.oauthState, flow.OAuthState)
			}
			delete(h.oauthFlows, id)
		}
	}
}
