package gmessages

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"jane/pkg/bus"
	"jane/pkg/channels"
	"jane/pkg/logger"
	mediaPkg "jane/pkg/media"

	"go.mau.fi/mautrix-gmessages/pkg/libgm"
	"go.mau.fi/mautrix-gmessages/pkg/libgm/events"
	"go.mau.fi/mautrix-gmessages/pkg/libgm/gmproto"
)

type EventHandler struct {
	Store       *Store
	Channel     *GMessagesChannel
	SessionPath string
	Client      *GMClient
}

func (h *EventHandler) Handle(rawEvt any) {
	switch evt := rawEvt.(type) {
	case *events.ClientReady:
		h.handleClientReady(evt)
	case *libgm.WrappedMessage:
		h.handleMessage(evt)
	case *gmproto.Conversation:
		h.handleConversation(evt)
	case *events.AuthTokenRefreshed:
		h.handleAuthRefresh()
	case *events.ListenFatalError:
		logger.ErrorCF("channels.gmessages", "Listen fatal error", map[string]any{"err": evt.Error})
		h.Channel.Stop(context.Background())
	case *events.ListenTemporaryError:
		logger.WarnCF("channels.gmessages", "Listen temporary error", map[string]any{"err": evt.Error})
	case *events.ListenRecovered:
		logger.InfoCF("channels.gmessages", "Listen recovered", nil)
	case *events.PhoneNotResponding:
		logger.WarnCF("channels.gmessages", "Phone not responding", nil)
	case *events.PhoneRespondingAgain:
		logger.InfoCF("channels.gmessages", "Phone responding again", nil)
	default:
		logger.DebugCF("channels.gmessages", "Unhandled event", map[string]any{"type": fmt.Sprintf("%T", evt)})
	}
}

func (h *EventHandler) handleClientReady(evt *events.ClientReady) {
	logger.InfoCF("channels.gmessages", "Client ready", map[string]any{
		"session_id":    evt.SessionID,
		"conversations": len(evt.Conversations),
	})

	for _, conv := range evt.Conversations {
		h.handleConversation(conv)
	}

	// Fetch contacts in background to populate our DB
	go func() {
		resp, err := h.Client.GM.ListContacts()
		if err != nil {
			logger.WarnCF("channels.gmessages", "Failed to list contacts", map[string]any{"err": err})
			return
		}

		for _, contact := range resp.GetContacts() {
			name := contact.GetName()
			if num := contact.GetNumber(); num != nil && num.GetNumber() != "" {
				h.Store.UpsertContact(num.GetNumber(), name)
			}
		}
		logger.InfoCF("channels.gmessages", "Finished syncing contacts", map[string]any{"count": len(resp.GetContacts())})
	}()
}

func (h *EventHandler) handleMessage(evt *libgm.WrappedMessage) {
	msg := evt.Message
	body := ExtractMessageBody(msg)
	senderName, senderNumber := ExtractSenderInfo(msg)

	status := "unknown"
	if ms := msg.GetMessageStatus(); ms != nil {
		status = ms.GetStatus().String()
	}

	dbMsg := &Message{
		MessageID:      msg.GetMessageID(),
		ConversationID: msg.GetConversationID(),
		SenderName:     senderName,
		SenderNumber:   senderNumber,
		Body:           body,
		TimestampMS:    msg.GetTimestamp() / 1000,
		Status:         status,
		IsFromMe:       msg.GetSenderParticipant() != nil && msg.GetSenderParticipant().GetIsMe(),
	}

	var mediaURLs []string

	if media := ExtractMediaInfo(msg); media != nil {
		dbMsg.MediaID = media.MediaID
		dbMsg.MimeType = media.MimeType
		dbMsg.DecryptionKey = hex.EncodeToString(media.DecryptionKey)

		saveToStore := func(data []byte, mimeType string) {
			if h.Channel.GetMediaStore() == nil {
				return
			}

			// Save to a temporary file
			tmpFile, err := os.CreateTemp("", "gmessages-media-*")
			if err != nil {
				logger.ErrorCF("channels.gmessages", "Failed to create temp file for media", map[string]any{"err": err})
				return
			}
			defer tmpFile.Close()

			if _, err := tmpFile.Write(data); err != nil {
				logger.ErrorCF("channels.gmessages", "Failed to write media data to temp file", map[string]any{"err": err})
				return
			}

			scope := channels.BuildMediaScope(h.Channel.Name(), msg.GetConversationID(), msg.GetMessageID())
			meta := mediaPkg.MediaMeta{ContentType: mimeType, Source: "gmessages"}

			ref, err := h.Channel.GetMediaStore().Store(tmpFile.Name(), meta, scope)
			if err == nil {
				mediaURLs = append(mediaURLs, ref)
			} else {
				logger.ErrorCF("channels.gmessages", "Failed to save media to store", map[string]any{"err": err})
			}
		}

		if media.InlineData != nil {
			saveToStore(media.InlineData, media.MimeType)
		} else if media.MediaID != "" && media.DecryptionKey != nil {
			data, err := h.Client.GM.DownloadMedia(media.MediaID, media.DecryptionKey)
			if err == nil {
				saveToStore(data, media.MimeType)
			} else {
				logger.ErrorCF("channels.gmessages", "Failed to download media from Google Messages", map[string]any{"err": err})
			}
		}
	}

	if reactions := ExtractReactions(msg); reactions != nil {
		if b, err := json.Marshal(reactions); err == nil {
			dbMsg.Reactions = string(b)
			// Append reactions to body for bot visibility if body isn't empty
			var reactionStrs []string
			for _, r := range reactions {
				reactionStrs = append(reactionStrs, r.Emoji)
			}
			if len(reactionStrs) > 0 {
				body = body + "\n[Reactions: " + strings.Join(reactionStrs, " ") + "]"
				dbMsg.Body = body
			}
		}
	}
	dbMsg.ReplyToID = ExtractReplyToID(msg)

	if err := h.Store.UpsertMessage(dbMsg); err != nil {
		logger.ErrorCF("channels.gmessages", "Failed to store message", map[string]any{"err": err, "msg_id": dbMsg.MessageID})
		return
	}

	if dbMsg.IsFromMe && !strings.HasPrefix(dbMsg.MessageID, "tmp_") {
		h.Store.DeleteTmpMessages(dbMsg.ConversationID)
		return // we don't dispatch our own messages usually, unless requested
	}

	// Dispatch to bot
	if !evt.IsOld && !dbMsg.IsFromMe {
		senderID := senderNumber
		if senderID == "" {
			senderID = senderName
		}

		// Map ChatID to phone number or conversation ID. In 1:1 we'll use phone number or conv ID
		chatID := msg.GetConversationID() // fallback to group/thread ID
		if chatID == "" {
			chatID = senderID
		}

		// determine if it's a group from DB (we do it in the next steps)

		peer := bus.Peer{
			ID: chatID,
		}

		senderInfo := bus.SenderInfo{
			PlatformID:  senderID,
			CanonicalID: senderID,
			DisplayName: senderName,
		}

		metadata := map[string]string{
			"conversation_id": msg.GetConversationID(),
		}

		h.Channel.HandleMessage(
			context.Background(),
			peer,
			msg.GetMessageID(),
			senderID,
			chatID,
			body,
			mediaURLs,
			metadata,
			senderInfo,
		)
	}

	logger.DebugCF("channels.gmessages", "Stored message", map[string]any{
		"msg_id": dbMsg.MessageID,
		"from":   senderName,
		"is_old": evt.IsOld,
	})
}

func (h *EventHandler) handleConversation(conv *gmproto.Conversation) {
	participantsJSON := "[]"
	if ps := conv.GetParticipants(); len(ps) > 0 {
		type pInfo struct {
			Name   string `json:"name"`
			Number string `json:"number"`
			IsMe   bool   `json:"is_me,omitempty"`
		}
		var infos []pInfo
		for _, p := range ps {
			info := pInfo{
				Name: p.GetFullName(),
				IsMe: p.GetIsMe(),
			}
			if id := p.GetID(); id != nil {
				info.Number = id.GetNumber()
			}
			if info.Number == "" {
				info.Number = p.GetFormattedNumber()
			}
			infos = append(infos, info)
		}
		if b, err := json.Marshal(infos); err == nil {
			participantsJSON = string(b)
		}
	}

	unread := 0
	if conv.GetUnread() {
		unread = 1
	}

	dbConv := &Conversation{
		ConversationID: conv.GetConversationID(),
		Name:           conv.GetName(),
		IsGroup:        conv.GetIsGroupChat(),
		Participants:   participantsJSON,
		LastMessageTS:  conv.GetLastMessageTimestamp() / 1000,
		UnreadCount:    unread,
	}

	if err := h.Store.UpsertConversation(dbConv); err != nil {
		logger.ErrorCF("channels.gmessages", "Failed to store conversation", map[string]any{"err": err, "conv_id": dbConv.ConversationID})
		return
	}
}

func (h *EventHandler) handleAuthRefresh() {
	if h.Client == nil || h.SessionPath == "" {
		return
	}
	sessionData, err := h.Client.SessionData()
	if err != nil {
		logger.ErrorCF("channels.gmessages", "Failed to get session data for save", map[string]any{"err": err})
		return
	}
	if err := saveSession(h.SessionPath, sessionData); err != nil {
		logger.ErrorCF("channels.gmessages", "Failed to save refreshed session", map[string]any{"err": err})
		return
	}
	logger.DebugCF("channels.gmessages", "Saved refreshed auth token", nil)
}

// Extractors mapped from libgm

func ExtractMessageBody(msg *gmproto.Message) string {
	for _, info := range msg.GetMessageInfo() {
		if mc := info.GetMessageContent(); mc != nil {
			return mc.GetContent()
		}
	}
	return ""
}

type MediaInfo struct {
	MediaID                string
	MimeType               string
	MediaName              string
	DecryptionKey          []byte
	Size                   int64
	ThumbnailMediaID       string
	ThumbnailDecryptionKey []byte
	InlineData             []byte
}

func ExtractMediaInfo(msg *gmproto.Message) *MediaInfo {
	for _, info := range msg.GetMessageInfo() {
		if mc := info.GetMediaContent(); mc != nil {
			mime := mc.GetMimeType()
			if mime == "" {
				switch {
				case mc.GetFormat() >= 1 && mc.GetFormat() <= 7:
					mime = "image/jpeg"
				default:
					mime = "application/octet-stream"
				}
			}

			mi := &MediaInfo{
				MediaID:                mc.GetMediaID(),
				MimeType:               mime,
				MediaName:              mc.GetMediaName(),
				DecryptionKey:          mc.GetDecryptionKey(),
				Size:                   mc.GetSize(),
				ThumbnailMediaID:       mc.GetThumbnailMediaID(),
				ThumbnailDecryptionKey: mc.GetThumbnailDecryptionKey(),
				InlineData:             mc.GetMediaData(),
			}

			if mi.MediaID == "" && mi.ThumbnailMediaID != "" {
				mi.MediaID = mi.ThumbnailMediaID
				mi.DecryptionKey = mi.ThumbnailDecryptionKey
			}

			return mi
		}
	}
	return nil
}

type Reaction struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

func ExtractReactions(msg *gmproto.Message) []Reaction {
	entries := msg.GetReactions()
	if len(entries) == 0 {
		return nil
	}
	var reactions []Reaction
	for _, entry := range entries {
		if data := entry.GetData(); data != nil {
			emoji := data.GetUnicode()
			if emoji == "" {
				continue
			}
			reactions = append(reactions, Reaction{
				Emoji: emoji,
				Count: len(entry.GetParticipantIDs()),
			})
		}
	}
	if len(reactions) == 0 {
		return nil
	}
	return reactions
}

func ExtractReplyToID(msg *gmproto.Message) string {
	if rm := msg.GetReplyMessage(); rm != nil {
		return rm.GetMessageID()
	}
	return ""
}

func ExtractSenderInfo(msg *gmproto.Message) (name, number string) {
	if p := msg.GetSenderParticipant(); p != nil {
		name = p.GetFullName()
		if name == "" {
			name = p.GetFirstName()
		}
		if id := p.GetID(); id != nil {
			number = id.GetNumber()
		}
		if number == "" {
			number = p.GetFormattedNumber()
		}
	}
	return
}
