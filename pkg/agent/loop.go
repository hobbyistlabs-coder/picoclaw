// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package agent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"jane/pkg/bus"
	"jane/pkg/channels"
	"jane/pkg/commands"
	"jane/pkg/config"
	"jane/pkg/media"
	"jane/pkg/providers"
	"jane/pkg/state"
	"jane/pkg/tools"
	"jane/pkg/voice"
)

type AgentLoop struct {
	bus              *bus.MessageBus
	cfg              *config.Config
	registry         *AgentRegistry
	state            *state.Manager
	running          atomic.Bool
	summarizing      sync.Map
	asyncBatches     sync.Map
	pendingApprovals sync.Map // Tracks state for Human-in-the-Loop approvals
	summaryJobs      chan summaryJob
	wg               sync.WaitGroup
	fallback         *providers.FallbackChain
	provider         providers.LLMProvider
	channelManager   *channels.Manager
	mediaStore       media.MediaStore
	transcriber      voice.Transcriber
	cmdRegistry      *commands.Registry
	mcp              mcpRuntime
	configPath       string
	configModTime    time.Time
	reloadMu         sync.Mutex
}

type pendingApprovalState struct {
	agent               *AgentInstance
	opts                processOptions
	normalizedToolCalls []providers.ToolCall
	messages            []providers.Message
	iteration           int
	activeCandidates    []providers.FallbackCandidate
	activeModel         string
}

// processOptions configures how a message is processed
type summaryJob struct {
	agent      *AgentInstance
	sessionKey string
}

type processOptions struct {
	SessionKey      string   // Session identifier for history/context
	Channel         string   // Target channel for tool execution
	ChatID          string   // Target chat ID for tool execution
	UserMessage     string   // User message content (may include prefix)
	Media           []string // media:// refs from inbound message
	DefaultResponse string   // Response when LLM returns empty
	EnableSummary   bool     // Whether to trigger summarization
	SendResponse    bool     // Whether to send response via bus
	NoHistory       bool     // If true, don't load session history (for heartbeat)
	Stream          bool     // Whether to stream LLM generation
}

const (
	defaultResponse           = "I've completed processing but have no response to give. Increase `max_tool_iterations` in config.json."
	sessionKeyAgentPrefix     = "agent:"
	metadataKeyAccountID      = "account_id"
	metadataKeyGuildID        = "guild_id"
	metadataKeyTeamID         = "team_id"
	metadataKeyParentPeerKind = "parent_peer_kind"
	metadataKeyParentPeerID   = "parent_peer_id"
)

// DispatchSubagent implements tools.AgentDispatcher.
func (al *AgentLoop) DispatchSubagent(
	ctx context.Context,
	agentID, task, originChannel, originChatID, sessionKey string,
) (*tools.ToolResult, error) {
	agent, ok := al.registry.GetAgent(agentID)
	if !ok {
		return nil, fmt.Errorf("agent '%s' not found", agentID)
	}

	opts := processOptions{
		SessionKey:      sessionKey,
		Channel:         originChannel,
		ChatID:          originChatID,
		UserMessage:     task,
		DefaultResponse: "Subagent task completed.",
		EnableSummary:   false, // Keep subagent session isolated from auto-summary for now
		SendResponse:    false, // Do not send intermediate bus messages to user for delegated tasks
		Stream:          false,
	}

	content, err := al.runAgentLoop(ctx, agent, opts)
	if err != nil {
		return nil, err
	}

	return &tools.ToolResult{
		ForLLM:  content,
		ForUser: content,
		Silent:  true,
		IsError: false,
		Async:   false,
	}, nil
}
