// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
// Copyright (c) 2026 PicoClaw contributors

package channels

import (
	"context"

	"jane/pkg/bus"
	"jane/pkg/constants"
	"jane/pkg/logger"
)

func dispatchLoop[M any](
	ctx context.Context,
	m *Manager,
	subscribe func(context.Context) (M, bool),
	getChannel func(M) string,
	enqueue func(context.Context, *channelWorker, M) bool,
	startMsg, stopMsg, unknownMsg, noWorkerMsg string,
) {
	logger.InfoC("channels", startMsg)

	for {
		msg, ok := subscribe(ctx)
		if !ok {
			logger.InfoC("channels", stopMsg)
			return
		}

		channel := getChannel(msg)

		// Silently skip internal channels
		if constants.IsInternalChannel(channel) {
			continue
		}

		m.mu.RLock()
		_, exists := m.channels[channel]
		w, wExists := m.workers[channel]
		m.mu.RUnlock()

		if !exists {
			logger.WarnCF("channels", unknownMsg, map[string]any{"channel": channel})
			continue
		}

		if wExists && w != nil {
			if !enqueue(ctx, w, msg) {
				return
			}
		} else if exists {
			logger.WarnCF("channels", noWorkerMsg, map[string]any{"channel": channel})
		}
	}
}

func (m *Manager) dispatchOutbound(ctx context.Context) {
	dispatchLoop(
		ctx, m,
		m.bus.SubscribeOutbound,
		func(msg bus.OutboundMessage) string { return msg.Channel },
		func(ctx context.Context, w *channelWorker, msg bus.OutboundMessage) bool {
			select {
			case w.queue <- msg:
				return true
			case <-ctx.Done():
				return false
			}
		},
		"Outbound dispatcher started",
		"Outbound dispatcher stopped",
		"Unknown channel for outbound message",
		"Channel has no active worker, skipping message",
	)
}

func (m *Manager) dispatchOutboundMedia(ctx context.Context) {
	dispatchLoop(
		ctx, m,
		m.bus.SubscribeOutboundMedia,
		func(msg bus.OutboundMediaMessage) string { return msg.Channel },
		func(ctx context.Context, w *channelWorker, msg bus.OutboundMediaMessage) bool {
			select {
			case w.mediaQueue <- msg:
				return true
			case <-ctx.Done():
				return false
			}
		},
		"Outbound media dispatcher started",
		"Outbound media dispatcher stopped",
		"Unknown channel for outbound media message",
		"Channel has no active worker, skipping media message",
	)
}

func (m *Manager) dispatchOutboundStream(ctx context.Context) {
	dispatchLoop(
		ctx, m,
		m.bus.SubscribeOutboundStream,
		func(msg bus.OutboundStreamMessage) string { return msg.Channel },
		func(ctx context.Context, w *channelWorker, msg bus.OutboundStreamMessage) bool {
			select {
			case w.streamQueue <- msg:
				return true
			case <-ctx.Done():
				return false
			}
		},
		"Outbound stream dispatcher started",
		"Outbound stream dispatcher stopped",
		"Unknown channel for outbound stream message",
		"Channel has no active worker, skipping stream message",
	)
}
