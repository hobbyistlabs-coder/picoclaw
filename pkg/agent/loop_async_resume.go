package agent

import (
	"fmt"
	"strings"

	"jane/pkg/providers"
)

func buildAsyncResumePrompt(
	history []providers.Message,
	senderID string,
	result string,
) string {
	original := latestUserRequest(history)
	if original == "" {
		return fmt.Sprintf("[System: %s] %s", senderID, result)
	}

	return fmt.Sprintf(
		"[System: %s]\nOriginal user request:\n%s\n\nCompleted async task results:\n%s\n\nUse the completed results above to answer the original request exactly. Do not call tools or delegate again.",
		senderID,
		original,
		result,
	)
}

func latestUserRequest(history []providers.Message) string {
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role != "user" {
			continue
		}
		if strings.HasPrefix(msg.Content, "[System: ") {
			continue
		}
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		return msg.Content
	}
	return ""
}
