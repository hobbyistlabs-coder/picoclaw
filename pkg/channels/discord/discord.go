package discord

import (
	"context"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"

	"jane/pkg/bus"
	"jane/pkg/channels"
	"jane/pkg/config"
	"jane/pkg/logger"
)

type DiscordChannel struct {
	*channels.BaseChannel
	session    *discordgo.Session
	config     config.DiscordConfig
	ctx        context.Context
	cancel     context.CancelFunc
	typingMu   sync.Mutex
	typingStop map[string]chan struct{} // chatID → stop signal
	botUserID  string                   // stored for mention checking
}

func NewDiscordChannel(cfg config.DiscordConfig, bus *bus.MessageBus) (*DiscordChannel, error) {
	discordgo.Logger = logger.NewLogger("discord").
		WithLevels(map[int]logger.LogLevel{
			discordgo.LogError:         logger.ERROR,
			discordgo.LogWarning:       logger.WARN,
			discordgo.LogInformational: logger.INFO,
			discordgo.LogDebug:         logger.DEBUG,
		}).Log

	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	if err := applyDiscordProxy(session, cfg.Proxy); err != nil {
		return nil, err
	}
	base := channels.NewBaseChannel("discord", cfg, bus, cfg.AllowFrom,
		channels.WithMaxMessageLength(2000),
		channels.WithGroupTrigger(cfg.GroupTrigger),
		channels.WithReasoningChannelID(cfg.ReasoningChannelID),
	)

	return &DiscordChannel{
		BaseChannel: base,
		session:     session,
		config:      cfg,
		ctx:         context.Background(),
		typingStop:  make(map[string]chan struct{}),
	}, nil
}

func (c *DiscordChannel) Start(ctx context.Context) error {
	logger.InfoC("discord", "Starting Discord bot")

	c.ctx, c.cancel = context.WithCancel(ctx)

	// Get bot user ID before opening session to avoid race condition
	botUser, err := c.session.User("@me")
	if err != nil {
		return fmt.Errorf("failed to get bot user: %w", err)
	}
	c.botUserID = botUser.ID

	c.session.AddHandler(c.handleMessage)

	if err := c.session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	c.SetRunning(true)

	logger.InfoCF("discord", "Discord bot connected", map[string]any{
		"username": botUser.Username,
		"user_id":  botUser.ID,
	})

	return nil
}

func (c *DiscordChannel) Stop(ctx context.Context) error {
	logger.InfoC("discord", "Stopping Discord bot")
	c.SetRunning(false)

	// Stop all typing goroutines before closing session
	c.typingMu.Lock()
	for chatID, stop := range c.typingStop {
		close(stop)
		delete(c.typingStop, chatID)
	}
	c.typingMu.Unlock()

	// Cancel our context so typing goroutines using c.ctx.Done() exit
	if c.cancel != nil {
		c.cancel()
	}

	if err := c.session.Close(); err != nil {
		return fmt.Errorf("failed to close discord session: %w", err)
	}

	return nil
}
