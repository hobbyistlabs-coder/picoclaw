// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
// Copyright (c) 2026 PicoClaw contributors

package channels

import (
	"context"
	"time"
)

// runTTLJanitor periodically scans the typingStops and placeholders maps
// and evicts entries that have exceeded their TTL. This prevents memory
// accumulation when outbound paths fail to trigger preSend (e.g. LLM errors).
func (m *Manager) runTTLJanitor(ctx context.Context) {
	ticker := time.NewTicker(janitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			m.typingStops.Range(func(key, value any) bool {
				if entry, ok := value.(typingEntry); ok {
					if now.Sub(entry.createdAt) > typingStopTTL {
						if _, loaded := m.typingStops.LoadAndDelete(key); loaded {
							entry.stop() // idempotent, safe
						}
					}
				}
				return true
			})
			m.reactionUndos.Range(func(key, value any) bool {
				if entry, ok := value.(reactionEntry); ok {
					if now.Sub(entry.createdAt) > typingStopTTL {
						if _, loaded := m.reactionUndos.LoadAndDelete(key); loaded {
							entry.undo() // idempotent, safe
						}
					}
				}
				return true
			})
			m.placeholders.Range(func(key, value any) bool {
				if entry, ok := value.(placeholderEntry); ok {
					if now.Sub(entry.createdAt) > placeholderTTL {
						m.placeholders.Delete(key)
					}
				}
				return true
			})
		}
	}
}
