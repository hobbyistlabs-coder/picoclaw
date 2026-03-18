package gmessages

import (
	"jane/pkg/bus"
	"jane/pkg/channels"
	"jane/pkg/config"
)

func init() {
	channels.RegisterFactory("gmessages", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewGMessagesChannel(cfg.Channels.GMessages, b)
	})
}
