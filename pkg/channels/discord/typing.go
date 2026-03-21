package discord

import (
	"context"
	"time"

	"jane/pkg/logger"
)

// startTyping starts a continuous typing indicator loop for the given chatID.
// It stops any existing typing loop for that chatID before starting a new one.
func (c *DiscordChannel) startTyping(chatID string) {
	c.typingMu.Lock()
	// Stop existing loop for this chatID if any
	if stop, ok := c.typingStop[chatID]; ok {
		close(stop)
	}
	stop := make(chan struct{})
	c.typingStop[chatID] = stop
	c.typingMu.Unlock()

	go func() {
		if err := c.session.ChannelTyping(chatID); err != nil {
			logger.DebugCF("discord", "ChannelTyping error", map[string]any{"chatID": chatID, "err": err})
		}
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()
		timeout := time.After(5 * time.Minute)
		for {
			select {
			case <-stop:
				return
			case <-timeout:
				return
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				if err := c.session.ChannelTyping(chatID); err != nil {
					logger.DebugCF("discord", "ChannelTyping error", map[string]any{"chatID": chatID, "err": err})
				}
			}
		}
	}()
}

// stopTyping stops the typing indicator loop for the given chatID.
func (c *DiscordChannel) stopTyping(chatID string) {
	c.typingMu.Lock()
	defer c.typingMu.Unlock()
	if stop, ok := c.typingStop[chatID]; ok {
		close(stop)
		delete(c.typingStop, chatID)
	}
}

// StartTyping implements channels.TypingCapable.
// It starts a continuous typing indicator and returns an idempotent stop function.
func (c *DiscordChannel) StartTyping(ctx context.Context, chatID string) (func(), error) {
	c.startTyping(chatID)
	return func() { c.stopTyping(chatID) }, nil
}
