package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"jane/pkg/bus"
	"jane/pkg/channels"
	"jane/pkg/identity"
	"jane/pkg/logger"
	"jane/pkg/media"
	"jane/pkg/utils"
)

func (c *DiscordChannel) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil || m.Author == nil {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check allowlist first to avoid downloading attachments for rejected users
	sender := bus.SenderInfo{
		Platform:    "discord",
		PlatformID:  m.Author.ID,
		CanonicalID: identity.BuildCanonicalID("discord", m.Author.ID),
		Username:    m.Author.Username,
	}
	// Build display name
	displayName := m.Author.Username
	if m.Author.Discriminator != "" && m.Author.Discriminator != "0" {
		displayName += "#" + m.Author.Discriminator
	}
	sender.DisplayName = displayName

	if !c.IsAllowedSender(sender) {
		logger.DebugCF("discord", "Message rejected by allowlist", map[string]any{
			"user_id": m.Author.ID,
		})
		return
	}

	content := m.Content

	// In guild (group) channels, apply unified group trigger filtering
	// DMs (GuildID is empty) always get a response
	if m.GuildID != "" {
		isMentioned := false
		for _, mention := range m.Mentions {
			if mention.ID == c.botUserID {
				isMentioned = true
				break
			}
		}
		content = c.stripBotMention(content)
		respond, cleaned := c.ShouldRespondInGroup(isMentioned, content)
		if !respond {
			logger.DebugCF("discord", "Group message ignored by group trigger", map[string]any{
				"user_id": m.Author.ID,
			})
			return
		}
		content = cleaned
	} else {
		// DMs: just strip bot mention without filtering
		content = c.stripBotMention(content)
	}

	// Resolve Discord refs in main content before concatenation to avoid
	// double-expanding links that appear in the referenced message.
	content = c.resolveDiscordRefs(s, content, m.GuildID)

	// Prepend referenced (quoted) message content if this is a reply
	if m.MessageReference != nil && m.ReferencedMessage != nil {
		refContent := m.ReferencedMessage.Content
		if refContent != "" {
			refAuthor := "unknown"
			if m.ReferencedMessage.Author != nil {
				refAuthor = m.ReferencedMessage.Author.Username
			}
			refContent = c.resolveDiscordRefs(s, refContent, m.GuildID)
			content = fmt.Sprintf("[quoted message from %s]: %s\n\n%s",
				refAuthor, refContent, content)
		}
	}

	senderID := m.Author.ID

	mediaPaths := make([]string, 0, len(m.Attachments))

	scope := channels.BuildMediaScope("discord", m.ChannelID, m.ID)

	// Helper to register a local file with the media store
	storeMedia := func(localPath, filename string) string {
		if store := c.GetMediaStore(); store != nil {
			ref, err := store.Store(localPath, media.MediaMeta{
				Filename: filename,
				Source:   "discord",
			}, scope)
			if err == nil {
				return ref
			}
		}
		return localPath // fallback
	}

	var contentBuilder strings.Builder
	contentBuilder.WriteString(content)
	for _, attachment := range m.Attachments {
		isAudio := utils.IsAudioFile(attachment.Filename, attachment.ContentType)

		if isAudio {
			localPath := c.downloadAttachment(attachment.URL, attachment.Filename)
			if localPath != "" {
				mediaPaths = append(mediaPaths, storeMedia(localPath, attachment.Filename))
				if contentBuilder.Len() > 0 {
					contentBuilder.WriteByte('\n')
				}
				contentBuilder.WriteString("[audio: ")
				contentBuilder.WriteString(attachment.Filename)
				contentBuilder.WriteByte(']')
			} else {
				logger.WarnCF("discord", "Failed to download audio attachment", map[string]any{
					"url":      attachment.URL,
					"filename": attachment.Filename,
				})
				mediaPaths = append(mediaPaths, attachment.URL)
				if contentBuilder.Len() > 0 {
					contentBuilder.WriteByte('\n')
				}
				contentBuilder.WriteString("[attachment: ")
				contentBuilder.WriteString(attachment.URL)
				contentBuilder.WriteByte(']')
			}
		} else {
			mediaPaths = append(mediaPaths, attachment.URL)
			if contentBuilder.Len() > 0 {
				contentBuilder.WriteByte('\n')
			}
			contentBuilder.WriteString("[attachment: ")
			contentBuilder.WriteString(attachment.URL)
			contentBuilder.WriteByte(']')
		}
	}
	content = contentBuilder.String()
	if content == "" && len(mediaPaths) == 0 {
		return
	}

	if content == "" {
		content = "[media only]"
	}

	logger.DebugCF("discord", "Received message", map[string]any{
		"sender_name": sender.DisplayName,
		"sender_id":   senderID,
		"preview":     utils.Truncate(content, 50),
	})

	peerKind := "channel"
	peerID := m.ChannelID
	if m.GuildID == "" {
		peerKind = "direct"
		peerID = senderID
	}

	peer := bus.Peer{Kind: peerKind, ID: peerID}

	metadata := map[string]string{
		"user_id":      senderID,
		"username":     m.Author.Username,
		"display_name": sender.DisplayName,
		"guild_id":     m.GuildID,
		"channel_id":   m.ChannelID,
		"is_dm":        fmt.Sprintf("%t", m.GuildID == ""),
	}

	c.HandleMessage(c.ctx, peer, m.ID, senderID, m.ChannelID, content, mediaPaths, metadata, sender)
}

// stripBotMention removes the bot mention from the message content.
// Discord mentions have the format <@USER_ID> or <@!USER_ID> (with nickname).
func (c *DiscordChannel) stripBotMention(text string) string {
	if c.botUserID == "" {
		return text
	}
	// Remove both regular mention <@USER_ID> and nickname mention <@!USER_ID>
	text = strings.ReplaceAll(text, fmt.Sprintf("<@%s>", c.botUserID), "")
	text = strings.ReplaceAll(text, fmt.Sprintf("<@!%s>", c.botUserID), "")
	return strings.TrimSpace(text)
}
