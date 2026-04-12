package gmessages

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"go.mau.fi/mautrix-gmessages/pkg/libgm/gmproto"

	"jane/pkg/bus"
	"jane/pkg/logger"
)

// ContactNumberMysteriousInt is the default value for the MysteriousInt field
const ContactNumberMysteriousInt = 7

func (c *GMessagesChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	client := c.state.client
	store := c.state.store

	convID := msg.ChatID

	var sim *gmproto.SIMPayload
	var participantID string

	if !strings.HasPrefix(convID, "bugle:") {
		existingID, err := store.GetConversationIDByNumber(msg.ChatID)
		if err == nil && existingID != "" {
			convID = existingID
		} else {
			logger.InfoCF("channels.gmessages", "Creating new conversation", map[string]any{"number": msg.ChatID})

			// Build numbers request
			numbers := []*gmproto.ContactNumber{{
				MysteriousInt: ContactNumberMysteriousInt,
				Number:        msg.ChatID,
				Number2:       msg.ChatID,
			}}

			req := &gmproto.GetOrCreateConversationRequest{
				Numbers: numbers,
			}
			newConv, err := client.GM.GetOrCreateConversation(req)
			if err != nil {
				return fmt.Errorf("failed to create conversation for %s: %w", msg.ChatID, err)
			}
			conv := newConv.GetConversation()
			if conv == nil {
				return fmt.Errorf("no conversation returned for %s", msg.ChatID)
			}
			convID = conv.GetConversationID()

			// Extract sim for new conv
			for _, p := range conv.GetParticipants() {
				if p.GetIsMe() {
					if id := p.GetID(); id != nil {
						participantID = id.GetNumber()
					}
					sim = p.GetSimPayload()
					break
				}
			}
			if sim == nil {
				if sc := conv.GetSimCard(); sc != nil {
					sim = sc.GetSIMData().GetSIMPayload()
				}
			}
		}
	}

	content := msg.Content

	tmpID := fmt.Sprintf("tmp_%012d", rand.Int63n(1e12))
	req := &gmproto.SendMessageRequest{
		ConversationID: convID,
		MessagePayload: &gmproto.MessagePayload{
			TmpID:                 tmpID,
			MessagePayloadContent: nil,
			MessageInfo: []*gmproto.MessageInfo{{
				Data: &gmproto.MessageInfo_MessageContent{MessageContent: &gmproto.MessageContent{
					Content: content,
				}},
			}},
			ConversationID: convID,
			ParticipantID:  participantID,
			TmpID2:         tmpID,
		},
		SIMPayload: sim,
		TmpID:      tmpID,
	}
	if msg.ReplyToMessageID != "" {
		req.Reply = &gmproto.ReplyPayload{
			MessageID: msg.ReplyToMessageID,
		}
	}

	_, err := client.GM.SendMessage(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logger.DebugCF("channels.gmessages", "Sent message", map[string]any{"conv_id": convID})
	return nil
}

// SendMedia sends rich media using mautrix library.
func (c *GMessagesChannel) SendMedia(ctx context.Context, msg bus.OutboundMediaMessage) error {
	client := c.state.client
	store := c.state.store

	mediaStore := c.GetMediaStore()
	if mediaStore == nil {
		return fmt.Errorf("no media store available for channel %s", c.Name())
	}

	convID := msg.ChatID

	var sim *gmproto.SIMPayload
	var participantID string

	// Ensure conversation exists or create one (similar logic to Send)
	if !strings.HasPrefix(convID, "bugle:") {
		existingID, err := store.GetConversationIDByNumber(msg.ChatID)
		if err == nil && existingID != "" {
			convID = existingID
		} else {
			numbers := []*gmproto.ContactNumber{{
				MysteriousInt: ContactNumberMysteriousInt,
				Number:        msg.ChatID,
				Number2:       msg.ChatID,
			}}
			req := &gmproto.GetOrCreateConversationRequest{Numbers: numbers}
			newConv, err := client.GM.GetOrCreateConversation(req)
			if err != nil {
				return fmt.Errorf("failed to create conversation for %s: %w", msg.ChatID, err)
			}
			conv := newConv.GetConversation()
			if conv == nil {
				return fmt.Errorf("no conversation returned for %s", msg.ChatID)
			}
			convID = conv.GetConversationID()

			for _, p := range conv.GetParticipants() {
				if p.GetIsMe() {
					if id := p.GetID(); id != nil {
						participantID = id.GetNumber()
					}
					sim = p.GetSimPayload()
					break
				}
			}
			if sim == nil {
				if sc := conv.GetSimCard(); sc != nil {
					sim = sc.GetSIMData().GetSIMPayload()
				}
			}
		}
	}

	for _, part := range msg.Parts {
		localPath, err := mediaStore.Resolve(part.Ref)
		if err != nil {
			logger.ErrorCF(
				"channels.gmessages",
				"Failed to resolve media ref",
				map[string]any{"ref": part.Ref, "err": err},
			)
			continue
		}

		fileData, err := os.ReadFile(localPath)
		if err != nil {
			logger.ErrorCF(
				"channels.gmessages",
				"Failed to read media file",
				map[string]any{"path": localPath, "err": err},
			)
			continue
		}

		mimeType := part.ContentType
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		filename := part.Filename
		if filename == "" {
			filename = fmt.Sprintf("media-%d", time.Now().UnixNano())
		}

		// Retry media upload in case of transient network errors
		var mediaContent *gmproto.MediaContent
		var errUpload error
		for attempt := 1; attempt <= 3; attempt++ {
			mediaContent, errUpload = client.GM.UploadMedia(fileData, mimeType, filename)
			if errUpload == nil {
				break
			}
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
		}

		if errUpload != nil {
			logger.ErrorCF(
				"channels.gmessages",
				"Failed to upload media after retries",
				map[string]any{"err": errUpload},
			)
			continue
		}

		tmpID := fmt.Sprintf("tmp_%012d", rand.Int63n(1e12))
		req := &gmproto.SendMessageRequest{
			ConversationID: convID,
			MessagePayload: &gmproto.MessagePayload{
				TmpID: tmpID,
				MessageInfo: []*gmproto.MessageInfo{{
					Data: &gmproto.MessageInfo_MediaContent{MediaContent: mediaContent},
				}},
				ConversationID: convID,
				ParticipantID:  participantID,
				TmpID2:         tmpID,
			},
			SIMPayload: sim,
			TmpID:      tmpID,
		}

		_, err = client.GM.SendMessage(req)
		if err != nil {
			logger.ErrorCF(
				"channels.gmessages",
				"Failed to send media message",
				map[string]any{"err": err},
			)
			continue
		}
	}

	return nil
}
