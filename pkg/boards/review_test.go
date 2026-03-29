package boards

import (
	"strings"
	"testing"
)

func TestBuildCardActionPromptIncludesIDs(t *testing.T) {
	prompt := BuildCardActionPrompt("board-123", "card-456")
	if !strings.Contains(prompt, "Board ID: board-123") {
		t.Fatalf("missing board id in prompt: %q", prompt)
	}
	if !strings.Contains(prompt, "Card ID: card-456") {
		t.Fatalf("missing card id in prompt: %q", prompt)
	}
}
