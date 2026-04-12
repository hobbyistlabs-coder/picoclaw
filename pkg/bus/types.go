package bus

// Peer identifies the routing peer for a message (direct, group, channel, etc.)
type Peer struct {
	Kind string `json:"kind"` // "direct" | "group" | "channel" | ""
	ID   string `json:"id"`
}

// SenderInfo provides structured sender identity information.
type SenderInfo struct {
	Platform    string `json:"platform,omitempty"`     // "telegram", "discord", "slack", ...
	PlatformID  string `json:"platform_id,omitempty"`  // raw platform ID, e.g. "123456"
	CanonicalID string `json:"canonical_id,omitempty"` // "platform:id" format
	Username    string `json:"username,omitempty"`     // username (e.g. @alice)
	DisplayName string `json:"display_name,omitempty"` // display name
}

type InboundMessage struct {
	Channel    string            `json:"channel"`
	SenderID   string            `json:"sender_id"`
	Sender     SenderInfo        `json:"sender"`
	ChatID     string            `json:"chat_id"`
	Content    string            `json:"content"`
	Media      []string          `json:"media,omitempty"`
	Peer       Peer              `json:"peer"`                  // routing peer
	MessageID  string            `json:"message_id,omitempty"`  // platform message ID
	MediaScope string            `json:"media_scope,omitempty"` // media lifecycle scope
	SessionKey string            `json:"session_key"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type OutboundMessage struct {
	Channel          string          `json:"channel"`
	ChatID           string          `json:"chat_id"`
	Content          string          `json:"content"`
	ReplyToMessageID string          `json:"reply_to_message_id,omitempty"`
	Metrics          *MessageMetrics `json:"metrics,omitempty"`
	ReasoningContent string          `json:"reasoning_content,omitempty"`
	ToolEvent        *ToolCallEvent  `json:"tool_event,omitempty"`
}

type OutboundStreamMessage struct {
	Channel     string `json:"channel"`
	ChatID      string `json:"chat_id"`
	Content     string `json:"content"`
	IsReasoning bool   `json:"is_reasoning"`
}

type ToolCallEvent struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Kind            string         `json:"kind,omitempty"`
	Status          string         `json:"status"`
	Label           string         `json:"label,omitempty"`
	Arguments       map[string]any `json:"arguments,omitempty"`
	Summary         string         `json:"summary,omitempty"`
	Result          string         `json:"result,omitempty"`
	DurationMS      int64          `json:"duration_ms,omitempty"`
	EventType       string         `json:"event_type,omitempty"`
	TaskID          string         `json:"task_id,omitempty"`
	Codename        string         `json:"codename,omitempty"`
	ParentSessionID string         `json:"parent_session_id,omitempty"`
	LatestEvent     string         `json:"latest_event,omitempty"`
	ProgressPercent int            `json:"progress_percent,omitempty"`
	Error           string         `json:"error,omitempty"`
	ToolName        string         `json:"tool_name,omitempty"`
	ToolStatus      string         `json:"tool_status,omitempty"`
}

// MediaPart describes a single media attachment to send.
type MediaPart struct {
	Type        string `json:"type"`                   // "image" | "audio" | "video" | "file"
	Ref         string `json:"ref"`                    // media store ref, e.g. "media://abc123"
	Caption     string `json:"caption,omitempty"`      // optional caption text
	Filename    string `json:"filename,omitempty"`     // original filename hint
	ContentType string `json:"content_type,omitempty"` // MIME type hint
}

// OutboundMediaMessage carries media attachments from Agent to channels via the bus.
type OutboundMediaMessage struct {
	Channel string      `json:"channel"`
	ChatID  string      `json:"chat_id"`
	Parts   []MediaPart `json:"parts"`
}
