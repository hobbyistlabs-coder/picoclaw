// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package agent

import (
	"context"
	"fmt"
	"strings"

	"jane/pkg/bus"
	"jane/pkg/constants"
	"jane/pkg/logger"
	"jane/pkg/providers"
	"jane/pkg/routing"
	"jane/pkg/utils"
)

func (al *AgentLoop) ProcessDirect(
	ctx context.Context,
	content, sessionKey string,
) (string, error) {
	return al.ProcessDirectWithChannel(ctx, content, sessionKey, "cli", "direct")
}

func (al *AgentLoop) ProcessDirectWithChannel(
	ctx context.Context,
	content, sessionKey, channel, chatID string,
) (string, error) {
	if err := al.ensureMCPInitialized(ctx); err != nil {
		return "", err
	}

	msg := bus.InboundMessage{
		Channel:    channel,
		SenderID:   "cron",
		ChatID:     chatID,
		Content:    content,
		SessionKey: sessionKey,
	}

	return al.processMessage(ctx, msg)
}

// ProcessHeartbeat processes a heartbeat request without session history.
// Each heartbeat is independent and doesn't accumulate context.
func (al *AgentLoop) ProcessHeartbeat(
	ctx context.Context,
	content, channel, chatID string,
) (string, error) {
	agent := al.registry.GetDefaultAgent()
	if agent == nil {
		return "", fmt.Errorf("no default agent for heartbeat")
	}
	return al.runAgentLoop(ctx, agent, processOptions{
		SessionKey:      "heartbeat",
		Channel:         channel,
		ChatID:          chatID,
		UserMessage:     content,
		DefaultResponse: defaultResponse,
		EnableSummary:   false,
		SendResponse:    false,
		NoHistory:       true, // Don't load session history for heartbeat
	})
}

func (al *AgentLoop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	// Add message preview to log (show full content for error messages)
	var logContent string
	if strings.Contains(msg.Content, "Error:") || strings.Contains(msg.Content, "error") {
		logContent = msg.Content // Full content for errors
	} else {
		logContent = utils.Truncate(msg.Content, 80)
	}
	logger.InfoCF(
		"agent",
		fmt.Sprintf("Processing message from %s:%s: %s", msg.Channel, msg.SenderID, logContent),
		map[string]any{
			"channel":     msg.Channel,
			"chat_id":     msg.ChatID,
			"sender_id":   msg.SenderID,
			"session_key": msg.SessionKey,
		},
	)

	var hadAudio bool
	msg, hadAudio = al.transcribeAudioInMessage(ctx, msg)

	// For audio messages the placeholder was deferred by the channel.
	// Now that transcription (and optional feedback) is done, send it.
	if hadAudio && al.channelManager != nil {
		al.channelManager.SendPlaceholder(ctx, msg.Channel, msg.ChatID)
	}

	// Route system messages to processSystemMessage
	if msg.Channel == "system" {
		return al.processSystemMessage(ctx, msg)
	}

	route, agent, routeErr := al.resolveMessageRoute(msg)
	if routeErr != nil {
		return "", routeErr
	}

	// Reset message-tool state for this round so we don't skip publishing due to a previous round.
	if tool, ok := agent.Tools.Get("message"); ok {
		if resetter, ok := tool.(interface{ ResetSentInRound() }); ok {
			resetter.ResetSentInRound()
		}
	}

	// Resolve session key from route, while preserving explicit agent-scoped keys.
	scopeKey := resolveScopeKey(route, msg.SessionKey)
	sessionKey := scopeKey

	// Check if this session has a pending human-in-the-loop approval
	if stateRaw, ok := al.pendingApprovals.Load(sessionKey); ok {
		state := stateRaw.(pendingApprovalState)
		content := strings.TrimSpace(strings.ToLower(msg.Content))

		isYes := content == "yes" || content == "y" || content == "approve"
		isNo := content == "no" || content == "n" || content == "deny"

		if isYes || isNo {
			// Valid response, consume the state
			al.pendingApprovals.Delete(sessionKey)

			// DO NOT add the user's message ("yes"/"no") to the session history.
			// LLM APIs strictly enforce that an `assistant` message with tool calls must be
			// immediately followed by `tool` role messages. Interleaving a `user` message
			// will cause a 400 Bad Request error.

			if isYes {
				logger.InfoCF("agent", "User approved tool execution", map[string]any{"session_key": sessionKey})

				// Execute the pending tools
				agentResults := al.executeToolBatch(ctx, state.Agent, state.Opts, state.NormalizedToolCalls, state.Iteration)

				// Process tool results similarly to runLLMIteration
				for _, r := range agentResults {
					if !r.result.Silent && r.result.ForUser != "" && state.Opts.SendResponse {
						al.bus.PublishOutbound(ctx, bus.OutboundMessage{
							Channel: state.Opts.Channel,
							ChatID:  state.Opts.ChatID,
							Content: r.result.ForUser,
						})
					}

					if len(r.result.Media) > 0 {
						parts := make([]bus.MediaPart, 0, len(r.result.Media))
						for _, ref := range r.result.Media {
							part := bus.MediaPart{Ref: ref}
							if al.mediaStore != nil {
								if _, meta, err := al.mediaStore.ResolveWithMeta(ref); err == nil {
									part.Filename = meta.Filename
									part.ContentType = meta.ContentType
									part.Type = inferMediaType(meta.Filename, meta.ContentType)
								}
							}
							parts = append(parts, part)
						}
						al.bus.PublishOutboundMedia(ctx, bus.OutboundMediaMessage{
							Channel: state.Opts.Channel,
							ChatID:  state.Opts.ChatID,
							Parts:   parts,
						})
					}

					contentForLLM := r.result.ForLLM
					if contentForLLM == "" && r.result.Err != nil {
						contentForLLM = r.result.Err.Error()
					}

					if r.result.IsError {
						logger.ErrorCF("agent", "Tool execution failed", map[string]any{
							"tool":           r.tc.Name,
							"error":          contentForLLM,
							"error_category": "logic_failure",
						})
					}

					toolResultMsg := providers.Message{
						Role:       "tool",
						Content:    contentForLLM,
						ToolCallID: r.tc.ID,
					}

					*state.Messages = append(*state.Messages, toolResultMsg)
					state.Agent.Sessions.AddFullMessage(state.SessionKey, toolResultMsg)
				}

				state.Agent.Tools.TickTTL()
			} else {
				logger.InfoCF("agent", "User denied tool execution", map[string]any{"session_key": sessionKey})

				for _, tc := range state.NormalizedToolCalls {
					toolResultMsg := providers.Message{
						Role:       "tool",
						Content:    "User denied permission to execute this tool.",
						ToolCallID: tc.ID,
					}
					*state.Messages = append(*state.Messages, toolResultMsg)
					state.Agent.Sessions.AddFullMessage(state.SessionKey, toolResultMsg)
				}
			}

			// Resume loop using runAgentLoop which continues until max iterations are reached

			// Replace the state's opts.UserMessage with empty string because the user's approval message
			// was just a confirmation of the tool, not a new query we want the LLM to process directly
			// as a user turn before evaluating the tool result. The user message is already saved in history.
			opts := state.Opts
			opts.UserMessage = ""

			// runAgentLoop expects to build context messages itself, but since we already appended
			// the tool results, we need a way to pass them. Actually, runAgentLoop rebuilds messages from
			// history. Since we called AddFullMessage on the session, runAgentLoop will naturally
			// load the pending tool results we just injected into the session history.
			return al.runAgentLoop(ctx, state.Agent, opts)
		} else {
			// Invalid response, remind user
			remindMsg := "⚠️ Awaiting approval. Please reply 'Yes' to execute or 'No' to cancel."
			al.bus.PublishOutbound(ctx, bus.OutboundMessage{
				Channel: msg.Channel,
				ChatID:  msg.ChatID,
				Content: remindMsg,
			})
			return "", nil
		}
	}

	logger.InfoCF("agent", "Routed message",
		map[string]any{
			"agent_id":      agent.ID,
			"scope_key":     scopeKey,
			"session_key":   sessionKey,
			"matched_by":    route.MatchedBy,
			"route_agent":   route.AgentID,
			"route_channel": route.Channel,
		})

	opts := processOptions{
		SessionKey:      sessionKey,
		Channel:         msg.Channel,
		ChatID:          msg.ChatID,
		UserMessage:     msg.Content,
		Media:           msg.Media,
		DefaultResponse: defaultResponse,
		EnableSummary:   true,
		SendResponse:    false,
		Stream:          true, // Hardcoded for Phase 1 streaming implementation
	}

	// Medical Persona specific routing interception
	if agent != nil && agent.ID == "the-clinician" {
		return al.processMedicalRequest(ctx, agent, opts)
	}

	// context-dependent commands check their own Runtime fields and report
	// "unavailable" when the required capability is nil.
	if response, handled := al.handleCommand(ctx, msg, agent, &opts); handled {
		return response, nil
	}

	return al.runAgentLoop(ctx, agent, opts)
}

func (al *AgentLoop) resolveMessageRoute(msg bus.InboundMessage) (routing.ResolvedRoute, *AgentInstance, error) {
	route := al.registry.ResolveRoute(routing.RouteInput{
		Channel:    msg.Channel,
		AccountID:  inboundMetadata(msg, metadataKeyAccountID),
		Peer:       extractPeer(msg),
		ParentPeer: extractParentPeer(msg),
		GuildID:    inboundMetadata(msg, metadataKeyGuildID),
		TeamID:     inboundMetadata(msg, metadataKeyTeamID),
	})

	agent, ok := al.registry.GetAgent(route.AgentID)
	if !ok {
		agent = al.registry.GetDefaultAgent()
	}
	if agent == nil {
		return routing.ResolvedRoute{}, nil, fmt.Errorf("no agent available for route (agent_id=%s)", route.AgentID)
	}

	return route, agent, nil
}

func resolveScopeKey(route routing.ResolvedRoute, msgSessionKey string) string {
	if msgSessionKey != "" && strings.HasPrefix(msgSessionKey, sessionKeyAgentPrefix) {
		return msgSessionKey
	}
	return route.SessionKey
}

func (al *AgentLoop) processSystemMessage(
	ctx context.Context,
	msg bus.InboundMessage,
) (string, error) {
	if msg.Channel != "system" {
		return "", fmt.Errorf(
			"processSystemMessage called with non-system message channel: %s",
			msg.Channel,
		)
	}

	logger.InfoCF("agent", "Processing system message",
		map[string]any{
			"sender_id": msg.SenderID,
			"chat_id":   msg.ChatID,
		})

	// Parse origin channel from chat_id (format: "channel:chat_id")
	var originChannel, originChatID string
	if idx := strings.Index(msg.ChatID, ":"); idx > 0 {
		originChannel = msg.ChatID[:idx]
		originChatID = msg.ChatID[idx+1:]
	} else {
		originChannel = "cli"
		originChatID = msg.ChatID
	}

	// Extract subagent result from message content
	// Format: "Task 'label' completed.\n\nResult:\n<actual content>"
	content := msg.Content
	if idx := strings.Index(content, "Result:\n"); idx >= 0 {
		content = content[idx+8:] // Extract just the result part
	}

	// Skip internal channels - only log, don't send to user
	if constants.IsInternalChannel(originChannel) {
		logger.InfoCF("agent", "Subagent completed (internal channel)",
			map[string]any{
				"sender_id":   msg.SenderID,
				"content_len": len(content),
				"channel":     originChannel,
			})
		return "", nil
	}

	// Use default agent for system messages
	agent := al.registry.GetDefaultAgent()
	if agent == nil {
		return "", fmt.Errorf("no default agent for system message")
	}

	// Use the origin session for context
	sessionKey := routing.BuildAgentMainSessionKey(agent.ID)

	return al.runAgentLoop(ctx, agent, processOptions{
		SessionKey:      sessionKey,
		Channel:         originChannel,
		ChatID:          originChatID,
		UserMessage:     fmt.Sprintf("[System: %s] %s", msg.SenderID, msg.Content),
		DefaultResponse: "Background task completed.",
		EnableSummary:   false,
		SendResponse:    true,
	})
}
