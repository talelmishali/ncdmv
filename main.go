package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	ncdmv "github.com/aksiksi/ncdmv/pkg/lib"
	"github.com/gtuk/discordwebhook"
)

const (
	discordWebhookUsername = "ncdmv-bot"
)

var (
	apptType       = flag.String("appt_type", "permit", fmt.Sprintf("appointment type (one of: %s)", strings.Join(ncdmv.ValidApptTypes(), ",")))
	locations      = flag.String("locations", "cary,durham-east,durham-south", fmt.Sprintf("comma-seperated list of locations to look for (valid options: %s)", strings.Join(ncdmv.ValidLocations(), ",")))
	discordWebhook = flag.String("discord_webhook", "", "Discord webhook URL to send notifications to")
	timeout        = flag.Int("timeout", 10, "timeout (seconds)")
	interval       = flag.Int("interval", 3, "interval between checks (minutes)")
	debug          = flag.Bool("debug", false, "enable debug mode")
	headless       = flag.Bool("headless", true, "enable headless browser")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	if *apptType == "" {
		log.Fatal("appt_type must be specified")
	}
	if *locations == "" {
		log.Fatalf("locations list must be specified: %s", *locations)
	}

	parsedApptType := ncdmv.StringToAppointmentType(*apptType)
	if parsedApptType == ncdmv.AppointmentTypeInvalid {
		log.Fatalf("Invalid appointment type specified: %q", *apptType)
	}

	var parsedLocations []ncdmv.Location
	for _, location := range strings.Split(*locations, ",") {
		parsedLocation := ncdmv.StringToLocation(location)
		if parsedLocation == ncdmv.LocationInvalid {
			log.Fatalf("Invalid location specified: %q", location)
		}
		parsedLocations = append(parsedLocations, parsedLocation)
	}

	client, cancel, err := ncdmv.NewClient(ctx, *headless, *debug)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer cancel()

	for {
		location, err := client.CheckLocations(parsedApptType, parsedLocations, time.Duration(*timeout)*time.Second)
		if err != nil {
			log.Fatalf("Failed to check all locations: %v", err)
		}

		if location != nil {
			log.Printf("Found available location %q", *location)
			if *discordWebhook != "" {
				username := discordWebhookUsername
				content := fmt.Sprintf("Found appointment at location %q", *location)
				if err := discordwebhook.SendMessage(*discordWebhook, discordwebhook.Message{
					Username: &username,
					Content:  &content,
				}); err != nil {
					log.Printf("Failed to send message to Discord webhook %q: %v", *discordWebhook, err)
				}
			}
			break
		}

		interval := time.Duration(*interval) * time.Minute
		log.Printf("Sleeping for %v between location checks...", interval)
		time.Sleep(interval)
	}
}
