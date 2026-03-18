package gmessages

import (
	"context"

	"jane/pkg/bus"
	"jane/pkg/channels"
	"jane/pkg/config"
)

type GMessagesChannel struct {
	*channels.BaseChannel
	cfg   config.GMessagesConfig
	state *ClientState
}

// ClientState holds the internal dependencies like libgm Client and DB.
// It will be instantiated and injected when the channel starts.
type ClientState struct {
	client *GMClient
	store  *Store
}

func NewGMessagesChannel(cfg config.GMessagesConfig, b *bus.MessageBus) (*GMessagesChannel, error) {
	bc := channels.NewBaseChannel(
		"gmessages",
		cfg,
		b,
		cfg.AllowFrom,
		channels.WithGroupTrigger(cfg.GroupTrigger),
		channels.WithReasoningChannelID(cfg.ReasoningChannelID),
	)

	ch := &GMessagesChannel{
		BaseChannel: bc,
		cfg:         cfg,
		state:       &ClientState{},
	}
	bc.SetOwner(ch)
	return ch, nil
}

func (c *GMessagesChannel) Start(ctx context.Context) error {
	if !c.cfg.Enabled {
		return nil
	}

	c.SetRunning(true)

	// Call the separate client initialization
	if err := c.initClient(ctx); err != nil {
		c.SetRunning(false)
		return err
	}

	return nil
}

func (c *GMessagesChannel) Stop(ctx context.Context) error {
	c.SetRunning(false)

	// Disconnect client if exists
	c.stopClient(ctx)

	return nil
}
