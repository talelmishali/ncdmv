package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

type commandDataKey struct{}

type CommandData struct {
	RequestID   string
	UserID      string
	CommandName string
	GuildID     string
}

func NewCommandContext(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) context.Context {
	var userID string
	if i.Member != nil {
		userID = i.Member.User.ID
	} else {
		userID = i.User.ID
	}
	return context.WithValue(ctx, commandDataKey{}, &CommandData{
		RequestID:   uuid.New().String(),
		UserID:      userID,
		CommandName: i.ApplicationCommandData().Name,
		GuildID:     i.GuildID,
	})
}

func GetRequestID(ctx context.Context) string {
	if d, ok := ctx.Value(commandDataKey{}).(*CommandData); ok {
		return d.RequestID
	}
	return ""
}

func GetUserID(ctx context.Context) string {
	if d, ok := ctx.Value(commandDataKey{}).(*CommandData); ok {
		return d.UserID
	}
	return ""
}

func GetCommandName(ctx context.Context) string {
	if d, ok := ctx.Value(commandDataKey{}).(*CommandData); ok {
		return d.CommandName
	}
	return ""
}

func GetGuildID(ctx context.Context) string {
	if d, ok := ctx.Value(commandDataKey{}).(*CommandData); ok {
		return d.GuildID
	}
	return ""
}
