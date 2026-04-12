package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var openRouterModelsURL = "https://openrouter.ai/api/v1/models"

func fetchOpenRouterCatalog(ctx context.Context, output, supported string) ([]openRouterModel, error) {
	u, err := url.Parse(openRouterModelsURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if output != "" {
		q.Set("output_modalities", output)
	}
	if supported != "" {
		q.Set("supported_parameters", supported)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouter models request failed: %s", resp.Status)
	}

	var payload openRouterCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Data, nil
}
