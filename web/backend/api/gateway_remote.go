package api

import (
	"encoding/json"
	"net/http"
	"time"
)

func (h *Handler) managesGatewayProcess() bool {
	return h.gatewayBaseURL == ""
}

func (h *Handler) remoteGatewayStatus() map[string]any {
	data := map[string]any{
		"gateway_status":        "stopped",
		"gateway_start_allowed": false,
		"gateway_start_reason":  "gateway is managed externally",
		"logs":                  []string{},
		"log_total":             0,
		"log_run_id":            0,
	}
	if h.gatewayBaseURL == "" {
		return data
	}

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(h.gatewayBaseURL + "/health")
	if err != nil {
		return data
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data["gateway_status"] = "error"
		data["status_code"] = resp.StatusCode
		return data
	}

	var healthData map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&healthData); err != nil {
		data["gateway_status"] = "error"
		return data
	}

	for key, value := range healthData {
		data[key] = value
	}
	data["gateway_status"] = "running"
	return data
}
