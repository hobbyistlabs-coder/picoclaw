package discord

import (
	"context"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"

	"jane/pkg/bus"
	"jane/pkg/channels"
	"jane/pkg/logger"
	"jane/pkg/utils"
)

// SendMedia implements the channels.MediaSender interface.
func (c *DiscordChannel) SendMedia(ctx context.Context, msg bus.OutboundMediaMessage) error {
	if !c.IsRunning() {
		return channels.ErrNotRunning
	}

	channelID := msg.ChatID
	if channelID == "" {
		return fmt.Errorf("channel ID is empty")
	}

	store := c.GetMediaStore()
	if store == nil {
		return fmt.Errorf("no media store available: %w", channels.ErrSendFailed)
	}

	// Collect all files into a single ChannelMessageSendComplex call
	files := make([]*discordgo.File, 0, len(msg.Parts))
	var caption string

	for _, part := range msg.Parts {
		localPath, err := store.Resolve(part.Ref)
		if err != nil {
			logger.ErrorCF("discord", "Failed to resolve media ref", map[string]any{
				"ref":   part.Ref,
				"error": err.Error(),
			})
			continue
		}

		file, err := os.Open(localPath)
		if err != nil {
			logger.ErrorCF("discord", "Failed to open media file", map[string]any{
				"path":  localPath,
				"error": err.Error(),
			})
			continue
		}
		// Note: discordgo reads from the Reader and we can't close it before send

		filename := part.Filename
		if filename == "" {
			filename = "file"
		}

		files = append(files, &discordgo.File{
			Name:        filename,
			ContentType: part.ContentType,
			Reader:      file,
		})

		if part.Caption != "" && caption == "" {
			caption = part.Caption
		}
	}

	if len(files) == 0 {
		return nil
	}

	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content: caption,
			Files:   files,
		})
		done <- err
	}()

	select {
	case err := <-done:
		// Close all file readers
		for _, f := range files {
			if closer, ok := f.Reader.(*os.File); ok {
				closer.Close()
			}
		}
		if err != nil {
			return fmt.Errorf("discord send media: %w", channels.ErrTemporary)
		}
		return nil
	case <-sendCtx.Done():
		// Close all file readers
		for _, f := range files {
			if closer, ok := f.Reader.(*os.File); ok {
				closer.Close()
			}
		}
		return sendCtx.Err()
	}
}

func (c *DiscordChannel) downloadAttachment(url, filename string) string {
	return utils.DownloadFile(url, filename, utils.DownloadOptions{
		LoggerPrefix: "discord",
		ProxyURL:     c.config.Proxy,
	})
}
