package api

import (
	"os"
	"strings"
)

func launcherGatewayBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("JANE_AI_GATEWAY_BASE_URL")), "/")
}
