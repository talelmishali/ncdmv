package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/aksiksi/ncdmv/pkg/discord"
	"github.com/bwmarrin/discordgo"
)

var (
	token = flag.String("token", "", "Discord bot token")
)

func main() {
	flag.Parse()

	if *token == "" {
		log.Fatalf("Token must be provided")
	}

	client, err := discord.NewClient(*token, 10*time.Second, false)
	if err != nil {
		log.Fatalf("Failed to create Discord client: %v", err)
	}

	cmd := &discordgo.ApplicationCommand{
		Name:        "basic-command",
		Description: "Basic command",
	}

	client.RegisterCommand(cmd, func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var userID string
		if i.Member != nil {
			userID = i.Member.User.ID
		} else {
			userID = i.User.ID
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hey there! Congratulations, you just executed your first slash command",
			},
		})
		channel, err := s.UserChannelCreate(userID)
		if err != nil {
			return fmt.Errorf("failed to create user channel: %v", err)
		}
		if _, err := s.ChannelMessageSend(channel.ID, "Hello there!"); err != nil {
			return fmt.Errorf("failed to send message over user channel: %v", err)
		}
		return nil
	})

	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer client.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
