package agent

import (
	"context"
	"os"
	"strings"
	"testing"

	"jane/pkg/bus"
	"jane/pkg/config"
	"jane/pkg/routing"
)

func TestProcessMessage_UsesRouteSessionKey(t *testing.T) {
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
	provider := &simpleMockProvider{response: "ok"}
	al := NewAgentLoop(cfg, msgBus, provider)

	msg := bus.InboundMessage{
		Channel:  "telegram",
		SenderID: "user1",
		ChatID:   "chat1",
		Content:  "hello",
		Peer: bus.Peer{
			Kind: "direct",
			ID:   "user1",
		},
	}

	route := al.registry.ResolveRoute(routing.RouteInput{
		Channel: msg.Channel,
		Peer:    extractPeer(msg),
	})
	sessionKey := route.SessionKey

	defaultAgent := al.registry.GetDefaultAgent()
	if defaultAgent == nil {
		t.Fatal("No default agent found")
	}

	helper := testHelper{al: al}
	_ = helper.executeAndGetResponse(t, context.Background(), msg)

	history := defaultAgent.Sessions.GetHistory(sessionKey)
	if len(history) != 2 {
		t.Fatalf("expected session history len=2, got %d", len(history))
	}
	if history[0].Role != "user" || history[0].Content != "hello" {
		t.Fatalf("unexpected first message in session: %+v", history[0])
	}
}

func TestProcessMessage_CommandOutcomes(t *testing.T) {
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
		Session: config.SessionConfig{
			DMScope: "per-channel-peer",
		},
	}

	msgBus := bus.NewMessageBus()
	provider := &countingMockProvider{response: "LLM reply"}
	al := NewAgentLoop(cfg, msgBus, provider)
	helper := testHelper{al: al}

	baseMsg := bus.InboundMessage{
		Channel:  "whatsapp",
		SenderID: "user1",
		ChatID:   "chat1",
		Peer: bus.Peer{
			Kind: "direct",
			ID:   "user1",
		},
	}

	showResp := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:  baseMsg.Channel,
		SenderID: baseMsg.SenderID,
		ChatID:   baseMsg.ChatID,
		Content:  "/show channel",
		Peer:     baseMsg.Peer,
	})
	if showResp != "Current Channel: whatsapp" {
		t.Fatalf("unexpected /show reply: %q", showResp)
	}
	if provider.calls != 0 {
		t.Fatalf("LLM should not be called for handled command, calls=%d", provider.calls)
	}

	fooResp := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:  baseMsg.Channel,
		SenderID: baseMsg.SenderID,
		ChatID:   baseMsg.ChatID,
		Content:  "/foo",
		Peer:     baseMsg.Peer,
	})
	if fooResp != "LLM reply" {
		t.Fatalf("unexpected /foo reply: %q", fooResp)
	}
	if provider.calls != 1 {
		t.Fatalf(
			"LLM should be called exactly once after /foo passthrough, calls=%d",
			provider.calls,
		)
	}

	newResp := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:  baseMsg.Channel,
		SenderID: baseMsg.SenderID,
		ChatID:   baseMsg.ChatID,
		Content:  "/new",
		Peer:     baseMsg.Peer,
	})
	if newResp != "LLM reply" {
		t.Fatalf("unexpected /new reply: %q", newResp)
	}
	if provider.calls != 2 {
		t.Fatalf("LLM should be called for passthrough /new command, calls=%d", provider.calls)
	}
}

func TestProcessMessage_SwitchModelShowModelConsistency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:         tmpDir,
				Provider:          "openai",
				Model:             "before-switch",
				MaxTokens:         4096,
				MaxToolIterations: 10,
			},
		},
	}

	msgBus := bus.NewMessageBus()
	provider := &countingMockProvider{response: "LLM reply"}
	al := NewAgentLoop(cfg, msgBus, provider)
	helper := testHelper{al: al}

	switchResp := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:  "telegram",
		SenderID: "user1",
		ChatID:   "chat1",
		Content:  "/switch model to after-switch",
		Peer: bus.Peer{
			Kind: "direct",
			ID:   "user1",
		},
	})
	if !strings.Contains(switchResp, "Switched model from before-switch to after-switch") {
		t.Fatalf("unexpected /switch reply: %q", switchResp)
	}

	showResp := helper.executeAndGetResponse(t, context.Background(), bus.InboundMessage{
		Channel:  "telegram",
		SenderID: "user1",
		ChatID:   "chat1",
		Content:  "/show model",
		Peer: bus.Peer{
			Kind: "direct",
			ID:   "user1",
		},
	})
	if !strings.Contains(showResp, "Current Model: after-switch (Provider: openai)") {
		t.Fatalf("unexpected /show model reply after switch: %q", showResp)
	}

	if provider.calls != 0 {
		t.Fatalf("LLM should not be called for /switch and /show, calls=%d", provider.calls)
	}
}
