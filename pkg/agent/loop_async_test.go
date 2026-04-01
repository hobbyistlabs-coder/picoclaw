package agent

import (
	"context"
	"os"
	"testing"

	"jane/pkg/bus"
	"jane/pkg/config"
)

func TestProcessSystemMessage_BatchesAsyncResults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:         tmpDir,
				Model:             "test-model",
				MaxTokens:         4096,
				MaxToolIterations: 10,
			},
		},
	}

	msgBus := bus.NewMessageBus()
	provider := &countingMockProvider{response: "batched summary"}
	al := NewAgentLoop(cfg, msgBus, provider)
	helper := testHelper{al: al}
	batch := map[string]string{
		"async_batch_id":       "batch-1",
		"async_batch_expected": "2",
		"async_tool_name":      "spawn",
	}

	first := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:    "system",
		SenderID:   "async:spawn",
		ChatID:     "telegram:chat-1",
		Content:    "first result",
		SessionKey: "agent:test:session",
		Metadata:   batch,
	})
	if first != "" {
		t.Fatalf("expected first batch item to stay silent, got %q", first)
	}
	if provider.calls != 0 {
		t.Fatalf("expected provider not to run until batch completes, got %d", provider.calls)
	}

	second := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:    "system",
		SenderID:   "async:spawn",
		ChatID:     "telegram:chat-1",
		Content:    "second result",
		SessionKey: "agent:test:session",
		Metadata:   batch,
	})
	if second != "batched summary" {
		t.Fatalf("expected final batched response, got %q", second)
	}
	if provider.calls != 1 {
		t.Fatalf("expected provider to run once after batch completion, got %d", provider.calls)
	}
	defaultAgent := al.registry.GetDefaultAgent()
	history := defaultAgent.Sessions.GetHistory("agent:test:session")
	if len(history) == 0 {
		t.Fatal("expected batched async results to be written to the originating session")
	}
}

func TestProcessSystemMessage_RestoresOriginalPromptForAsyncResume(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:         tmpDir,
				Model:             "test-model",
				MaxTokens:         4096,
				MaxToolIterations: 10,
			},
		},
	}

	msgBus := bus.NewMessageBus()
	provider := &resumePromptMockProvider{
		response:       "final: alpha beta",
		requiredPrompt: "return exactly this format: final: alpha beta",
		requiredResult: "Subagent 'Bean' completed (iterations: 1): alpha",
	}
	al := NewAgentLoop(cfg, msgBus, provider)
	helper := testHelper{al: al}
	sessionKey := "agent:test:resume"
	defaultAgent := al.registry.GetDefaultAgent()
	defaultAgent.Sessions.AddMessage(
		sessionKey,
		"user",
		"Use the spawn tool exactly twice in parallel and return exactly this format: final: alpha beta",
	)

	batch := map[string]string{
		"async_batch_id":       "batch-2",
		"async_batch_expected": "2",
		"async_tool_name":      "spawn",
	}

	first := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:    "system",
		SenderID:   "async:spawn",
		ChatID:     "pico:chat-2",
		Content:    "Subagent 'Petra' completed (iterations: 1): beta",
		SessionKey: sessionKey,
		Metadata:   batch,
	})
	if first != "" {
		t.Fatalf("expected first batch item to stay silent, got %q", first)
	}

	second := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:    "system",
		SenderID:   "async:spawn",
		ChatID:     "pico:chat-2",
		Content:    "Subagent 'Bean' completed (iterations: 1): alpha",
		SessionKey: sessionKey,
		Metadata:   batch,
	})
	if second != "final: alpha beta" {
		t.Fatalf("expected exact resumed answer, got %q", second)
	}
	if provider.calls != 1 {
		t.Fatalf("expected provider to run once after batch completion, got %d", provider.calls)
	}
}
