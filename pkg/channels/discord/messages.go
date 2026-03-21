package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"jane/pkg/bus"
	"jane/pkg/channels"
)

func (c *DiscordChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return channels.ErrNotRunning
	}

	channelID := msg.ChatID
	if channelID == "" {
		return fmt.Errorf("channel ID is empty")
	}

	if len([]rune(msg.Content)) == 0 {
		return nil
	}

	return c.sendChunk(ctx, channelID, msg.Content, msg.ReplyToMessageID)
}

// EditMessage implements channels.MessageEditor.
func (c *DiscordChannel) EditMessage(ctx context.Context, chatID string, messageID string, content string) error {
	_, err := c.session.ChannelMessageEdit(chatID, messageID, content)
	return err
}

// SendPlaceholder implements channels.PlaceholderCapable.
// It sends a placeholder message that will later be edited to the actual
// response via EditMessage (channels.MessageEditor).
func (c *DiscordChannel) SendPlaceholder(ctx context.Context, chatID string) (string, error) {
	if !c.config.Placeholder.Enabled {
		return "", nil
	}

	text := c.config.Placeholder.Text
	if text == "" {
		text = "Thinking... 💭"
	}

	msg, err := c.session.ChannelMessageSend(chatID, text)
	if err != nil {
		return "", err
	}

	return msg.ID, nil
}

func (c *DiscordChannel) sendChunk(ctx context.Context, channelID, content, replyToID string) error {
	// Use the passed ctx for timeout control
	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		var err error

		// If we have an ID, we send the message as "Reply"
		if replyToID != "" {
			_, err = c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Content: content,
				Reference: &discordgo.MessageReference{
					MessageID: replyToID,
					ChannelID: channelID,
				},
			})
		} else {
			// Otherwise, we send a normal message
			_, err = c.session.ChannelMessageSend(channelID, content)
		}

		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("discord send: %w", channels.ErrTemporary)
		}
		return nil
	case <-sendCtx.Done():
		return sendCtx.Err()
	}
}

// resolveDiscordRefs resolves channel references (<#id> → #channel-name) and
// expands Discord message links to show the linked message content.
// Only links pointing to the same guild are expanded to prevent cross-guild leakage.
func (c *DiscordChannel) resolveDiscordRefs(s *discordgo.Session, text string, guildID string) string {
	// 1. Resolve channel references: <#id> → #channel-name
	text = channelRefRe.ReplaceAllStringFunc(text, func(match string) string {
		parts := channelRefRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		// Prefer session state cache to avoid API calls
		if ch, err := s.State.Channel(parts[1]); err == nil {
			return "#" + ch.Name
		}
		if ch, err := s.Channel(parts[1]); err == nil {
			return "#" + ch.Name
		}
		return match
	})

	// 2. Expand Discord message links (max 3, same guild only)
	matches := msgLinkRe.FindAllStringSubmatch(text, 3)
	if len(matches) == 0 {
		return text
	}
	var sb strings.Builder
	sb.WriteString(text)
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		linkGuildID, channelID, messageID := m[1], m[2], m[3]
		// Security: only expand links from the same guild
		if linkGuildID != guildID {
			continue
		}
		msg, err := s.ChannelMessage(channelID, messageID)
		if err != nil || msg == nil || msg.Content == "" {
			continue
		}
		author := "unknown"
		if msg.Author != nil {
			author = msg.Author.Username
		}
		sb.WriteString("\n[linked message from ")
		sb.WriteString(author)
		sb.WriteString("]: ")
		sb.WriteString(msg.Content)
	}

	return sb.String()
}
