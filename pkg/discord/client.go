package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

const (
	GuildIDGlobal = ""
)

// commandInfo represents the state for a single Discord command.
type commandInfo struct {
	command *discordgo.ApplicationCommand
	handler func(context.Context, *Client, *discordgo.Session, *discordgo.InteractionCreate) error
	created bool
}

type Client struct {
	session        *discordgo.Session
	commandTimeout time.Duration
	cleanupOnStop  bool
	commands       map[ /* guild */ string]map[ /* name */ string]*commandInfo
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
		commands:       make(map[string]map[string]*commandInfo),
	}, nil
}

func (c *Client) appID() string {
	return c.session.State.User.ID
}

func (c *Client) RegisterCommand(cmd *discordgo.ApplicationCommand, guildID string, handler func(ctx context.Context, client *Client, s *discordgo.Session, i *discordgo.InteractionCreate) error) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.commands[guildID]; !ok {
		c.commands[guildID] = make(map[string]*commandInfo)
	} else if c.commands[guildID][cmd.Name] != nil {
		return fmt.Errorf("command %q already registered for guild %q", cmd.Name, guildID)
	}
	c.commands[guildID][cmd.Name] = &commandInfo{
		command: cmd,
		handler: handler,
	}
	return nil
}

func (c *Client) startInternal() error {
	var err error

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

	// Figure out what commands we already have across all known guilds.
	numExistingCommands := 0
	for guildID, commands := range c.commands {
		existingCommands, err := c.session.ApplicationCommands(c.appID(), guildID)
		if err != nil {
			return fmt.Errorf("failed to list existing commands for guild %q: %w", guildID, err)
		}
		for _, cmd := range existingCommands {
			if commands[cmd.Name] == nil {
				continue
			}
			commands[cmd.Name].command = cmd
			commands[cmd.Name].created = true
			numExistingCommands++
		}
	}
	slog.Info("Found existing commands", "count", numExistingCommands)

	// Create any commands that do not already exist.
	numCommands := 0
	for guildID, commands := range c.commands {
		for name := range commands {
			cmd := commands[name]
			if cmd.created {
				slog.Info("Skipping created command", "name", name, "guildID", guildID)
				continue
			}
			newCmd, err := c.session.ApplicationCommandCreate(c.appID(), guildID, cmd.command)
			if err != nil {
				return fmt.Errorf("failed to register command %v for guild %q: %w", cmd, guildID, err)
			}
			commands[name].command = newCmd
			commands[name].created = true
			numCommands++
		}
	}
	slog.Info("Created commands", "count", numCommands)

	return nil
}

// Start runs the the Discord API.
func (c *Client) Start() (err error) {
	ctx := context.Background()

	// The commands init logic is wrapped in a function partially for "cleanliness", but more
	// importantly to ensure that we don't recursively take a writer and a reader lock.
	// This could happen if the handler receives a command _before_ releasing the writer lock.
	if err := c.startInternal(); err != nil {
		return err
	}

	c.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		ctx := NewCommandContext(ctx, s, i)
		id := GetRequestID(ctx)
		name := GetCommandName(ctx)
		guildID := GetGuildID(ctx)

		ctx, cancel := context.WithTimeout(ctx, c.commandTimeout)
		defer cancel()

		// Take a reader lock to safely access client state and the commands map.
		c.lock.RLock()
		defer c.lock.RUnlock()

		if !c.started {
			slog.Warn("Client has already been stopped; exiting...")
			return
		}

		var cmd *commandInfo
		if guildID != "" && c.commands[guildID] != nil && c.commands[guildID][name] != nil {
			// Try to check if we have a command for the current guild ID.
			cmd = c.commands[guildID][name]
		} else if c.commands[GuildIDGlobal] != nil && c.commands[GuildIDGlobal][name] != nil {
			// Otherwise, assume it's a global command.
			cmd = c.commands[GuildIDGlobal][name]
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Unknown command: %q (id=%s)", name, id),
				},
			})
			slog.Warn("Unknown command", "name", name, "id", id, "guildID", guildID)
			return
		}

		slog.Info("Handling command...", "name", name, "id", id, "guildID", guildID)
		if err := cmd.handler(ctx, c, s, i); err != nil {
			slog.Error("Command failed", "name", name, "id", id, "guildID", guildID, "err", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Command failed to process (id=%s)", id),
				},
			})
			return
		}
		slog.Info("Completed command", "name", name, "id", id, "guildID", guildID)
	})

	slog.Info("Registered session handler")

	return nil
}

func (c *Client) Stop() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	slog.Info("Stopping client...", "cleanupOnStop", c.cleanupOnStop)

	if c.started {
		if err := c.session.Close(); err != nil {
			return fmt.Errorf("failed to close session: %w", err)
		}
		c.started = false
	}

	if c.cleanupOnStop {
		// Determine which commands need to be deleted.
		commandsToCleanup := make(map[ /* guildID */ string][]*discordgo.ApplicationCommand)
		numCommands := 0
		for guildID, commands := range c.commands {
			for _, cmd := range commands {
				if !cmd.created {
					continue
				}
				commandsToCleanup[guildID] = append(commandsToCleanup[guildID], cmd.command)
				numCommands++
			}
		}
		// Delete the commands.
		for guildID, commands := range commandsToCleanup {
			for _, cmd := range commands {
				if err := c.session.ApplicationCommandDelete(c.appID(), guildID, cmd.ID); err != nil {
					slog.Info("Command deletion failed", "err", err)
					if strings.Contains(err.Error(), "404: Not Found") {
						slog.Warn("Command already deleted", "guildID", guildID, "id", cmd.ID, "name", cmd.Name)
						continue
					} else {
						return fmt.Errorf("failed to delete command %q for guild %q: %w", cmd.Name, guildID, err)
					}
				}
			}
			delete(c.commands, guildID)
		}
		slog.Info("Cleaned up commands", "count", numCommands)
	}

	return nil
}
