package discord

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

// commandInfo represents the state for a single Discord command.
type commandInfo struct {
	command *discordgo.ApplicationCommand
	handler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
	created bool
}

type Client struct {
	session        *discordgo.Session
	commandTimeout time.Duration
	cleanupOnStop  bool
	commands       map[ /* name */ string]*commandInfo
	started        bool
	lock           sync.RWMutex
}

func NewClient(token string, commandTimeout time.Duration, cleanupOnStop bool) (*Client, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}
	return &Client{
		session:        session,
		commandTimeout: commandTimeout,
		cleanupOnStop:  cleanupOnStop,
		commands:       make(map[string]*commandInfo),
	}, nil
}

func (c *Client) appID() string {
	if !c.started {
		return ""
	}
	return c.session.State.User.ID
}

func (c *Client) RegisterCommand(cmd *discordgo.ApplicationCommand, handler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.commands[cmd.Name] != nil {
		return fmt.Errorf("command %q already registered", cmd.Name)
	}
	c.commands[cmd.Name] = &commandInfo{
		command: cmd,
		handler: handler,
	}
	return nil
}

// Start runs the the Discord API.
func (c *Client) Start() (err error) {
	ctx := context.Background()

	// The commands init block is wrapped in a function to ensure that we don't recursively
	// take a writer and a reader lock. This could happen if the handler receives a command
	// _before_ we release the writer lock.
	if err := func() error {
		c.lock.Lock()
		defer c.lock.Unlock()

		if len(c.commands) == 0 {
			return fmt.Errorf("no commands registered")
		}

		if err := c.session.Open(); err != nil {
			return fmt.Errorf("failed to open Discord session: %w", err)
		}
		c.started = true
		defer func() {
			// Close the session on error.
			if err != nil {
				slog.Warn("Closing session due to error", "result", c.session.Close())
			}
		}()
		slog.Info("Opened Discord client")

		// Figure out what commands we already have.
		existingCommands, err := c.session.ApplicationCommands(c.appID(), "")
		if err != nil {
			return fmt.Errorf("failed to list existing commands: %w", err)
		}
		for _, cmd := range existingCommands {
			if c.commands[cmd.Name] == nil {
				continue
			}
			c.commands[cmd.Name].created = true
		}

		for name := range c.commands {
			cmd := c.commands[name]
			if cmd.created {
				slog.Info("Skipping created command", "name", name)
				continue
			}
			newCmd, err := c.session.ApplicationCommandCreate(c.appID(), "" /* guildID */, cmd.command)
			if err != nil {
				return fmt.Errorf("failed to register command %v: %w", cmd, err)
			}
			c.commands[name].command = newCmd
			c.commands[name].created = true
		}
		slog.Info("Added commands", "count", len(c.commands))

		return nil
	}(); err != nil {
		return err
	}

	c.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		id := uuid.New()
		name := i.ApplicationCommandData().Name
		ctx, cancel := context.WithTimeout(ctx, c.commandTimeout)
		defer cancel()

		// Take a reader lock to safely access the commands map.
		c.lock.RLock()
		defer c.lock.RUnlock()
		if c, ok := c.commands[name]; ok {
			slog.Info("Handling command...", "id", id, "name", name)
			if err := c.handler(ctx, s, i); err != nil {
				slog.Error("Command failed", "id", id, "name", name, "err", err)
				return
			}
			slog.Info("Command completed", "id", id, "name", name)
		} else {
			slog.Warn("Unknown command", "id", id, "name", name)
		}
	})
	slog.Info("Registered session handler")

	return nil
}

func (c *Client) Stop() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.started {
		if err := c.session.Close(); err != nil {
			return fmt.Errorf("failed to close session: %w", err)
		}
		c.started = false
	}

	if c.cleanupOnStop {
		var commandsToCleanup []*discordgo.ApplicationCommand
		for _, cmd := range c.commands {
			if !cmd.created {
				continue
			}
			commandsToCleanup = append(commandsToCleanup, cmd.command)
		}
		for _, cmd := range commandsToCleanup {
			if err := c.session.ApplicationCommandDelete(c.appID(), "" /* guildID */, cmd.ID); err != nil {
				return fmt.Errorf("failed to delete command %q: %w", cmd.Name, err)
			}
			delete(c.commands, cmd.Name)
		}
	}

	return nil
}
