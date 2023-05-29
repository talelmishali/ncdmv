package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/gtuk/discordwebhook"
)

const (
	apptUrl                = "https://skiptheline.ncdot.gov/"
	makeApptButtonSel      = "button#cmdMakeAppt"
	apptTypeSelector       = `div.QflowObjectItem[data-id="%d"]`
	locationSelector       = apptTypeSelector
	discordWebhookUsername = "ncdmv-bot"
)

func mapToJoinedKeys[K comparable, V any](m map[K]V) string {
	var keys []string
	for k := range m {
		keys = append(keys, fmt.Sprintf("%v", k))
	}
	return strings.Join(keys, ",")
}

// AppointmentType represents the type of appointment.
// The value is the index of the box in the UI (see: "data-id").
type AppointmentType int

func (a AppointmentType) ToSelector() string {
	return fmt.Sprintf(apptTypeSelector, a)
}

const (
	AppointmentTypeDriverLicense          AppointmentType = 1
	AppointmentTypeDriverLicenseDuplicate                 = 2
	AppointmentTypeDriverLicenseRenewal                   = 3
	AppointmentTypePermit                                 = 9
)

var appointmentTypeMap map[string]AppointmentType = map[string]AppointmentType{
	"license":           AppointmentTypeDriverLicense,
	"license-duplicate": AppointmentTypeDriverLicenseDuplicate,
	"license-renewal":   AppointmentTypeDriverLicenseRenewal,
	"permit":            AppointmentTypePermit,
}
var validApptTypes string = mapToJoinedKeys(appointmentTypeMap)

type Location int

const (
	LocationCary         Location = 66
	LocationDurhamEast   Location = 47
	LocationDurhamSouth  Location = 80
	LocationRaleighEast  Location = 181
	LocationRaleighNorth Location = 10
	LocationRaleighWest  Location = 9
)

func (l Location) ToSelector() string {
	return fmt.Sprintf(locationSelector, l)
}

func (l Location) String() string {
	switch l {
	case LocationCary:
		return "cary"
	case LocationDurhamEast:
		return "durham-east"
	case LocationDurhamSouth:
		return "durham-south"
	case LocationRaleighEast:
		return "raleigh-east"
	case LocationRaleighNorth:
		return "raleigh-north"
	case LocationRaleighWest:
		return "raleigh-west"
	}
	return ""
}

var locationMap map[string]Location = map[string]Location{
	LocationCary.String():         LocationCary,
	LocationDurhamEast.String():   LocationDurhamEast,
	LocationDurhamSouth.String():  LocationDurhamSouth,
	LocationRaleighEast.String():  LocationRaleighEast,
	LocationRaleighNorth.String(): LocationRaleighNorth,
	LocationRaleighWest.String():  LocationRaleighWest,
}
var validLocations string = mapToJoinedKeys(locationMap)

func isLocationNodeEnabled(node *cdp.Node) bool {
	return strings.Contains(node.AttributeValue("class"), "Active-Unit")
}

var (
	apptType       = flag.String("appt_type", "permit", fmt.Sprintf("appointment type (one of: %s)", validApptTypes))
	rawLocations   = flag.String("locations", "cary,durham-east", fmt.Sprintf("comma-seperated list of locations to look for (valid options: %s)", validLocations))
	discordWebhook = flag.String("discord_webhook", "", "Discord webhook URL to send notifications to")
	timeout        = flag.Int("timeout", 10, "timeout (seconds)")
	interval       = flag.Int("interval", 3, "interval between checks (minutes)")
	debug          = flag.Bool("debug", false, "enable debug mode")
	headless       = flag.Bool("headless", true, "enable headless browser")
)

func isLocationAvailable(ctx context.Context, apptType AppointmentType, location Location, timeout time.Duration, screenshotPath string) (bool, error) {
	// Navigate to the main page.
	if _, err := chromedp.RunResponse(ctx, chromedp.Navigate(apptUrl)); err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Click the "Make Appointment" button once it is visible.
	if _, err := chromedp.RunResponse(ctx, chromedp.Click(makeApptButtonSel, chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		return false, err
	}

	// Click the appointment type.
	if _, err := chromedp.RunResponse(ctx, chromedp.Click(apptType.ToSelector(), chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		return false, err
	}

	// Wait for the location and read the node.
	var screenshotBuf []byte
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx,
		//browser.GrantPermissions([]browser.PermissionType{browser.PermissionTypeGeolocation}).WithOrigin(apptUrl),
		chromedp.WaitVisible(location.ToSelector(), chromedp.ByQuery),
		chromedp.Nodes(location.ToSelector(), &nodes, chromedp.NodeVisible, chromedp.ByQuery),
	); err != nil {
		return false, err
	}

	if screenshotPath != "" {
		if err := chromedp.Run(ctx, chromedp.FullScreenshot(&screenshotBuf, 90)); err != nil {
			return false, err
		}
		if err := os.WriteFile(screenshotPath, screenshotBuf, 0o644); err != nil {
			return false, fmt.Errorf("failed to write screenshot to path %q: %w", screenshotPath, err)
		}
	}

	if len(nodes) == 0 {
		return false, fmt.Errorf("found no nodes for location %q - is it even valid?", location)
	} else if len(nodes) != 1 {
		return false, fmt.Errorf("found multiple nodes for location %q: %+v", location, nodes)
	}

	isAvailable := isLocationNodeEnabled(nodes[0])

	return isAvailable, nil
}

func checkAllLocations(ctx context.Context, apptType AppointmentType, locations []Location, timeout time.Duration, screenshotPath string) (bool, *Location, error) {
	// Setup a seperate tab context for each location.
	var locationCtxs []context.Context
	for range locations {
		ctx, cancel := chromedp.NewContext(ctx)
		defer cancel()
		locationCtxs = append(locationCtxs, ctx)
	}

	type locationResult struct {
		idx         int
		isAvailable bool
		err         error
	}
	resultChan := make(chan locationResult)

	// Spawn a goroutine for each location.
	for i, location := range locations {
		i := i
		location := location
		ctx := locationCtxs[i]
		go func() {
			isAvailable, err := isLocationAvailable(ctx, apptType, location, timeout, screenshotPath)
			resultChan <- locationResult{
				idx:         i,
				isAvailable: isAvailable,
				err:         err,
			}
		}()
	}

	for i := 0; i < len(locations); i++ {
		result := <-resultChan
		location := locations[result.idx]

		log.Printf("Processing result for location index %q...", location)
		if result.err != nil {
			return false, nil, result.err
		}
		if result.isAvailable {
			// We found a location!
			return true, &location, nil
		}
		log.Printf("Location %q has no appointments available!", location)
	}

	return false, nil, nil
}

func main() {
	flag.Parse()

	ctx := context.Background()

	if *apptType == "" {
		log.Fatal("appt_type must be specified")
	}
	if *rawLocations == "" || !strings.Contains(*rawLocations, ",") {
		log.Fatalf("invalid locations: %s", *rawLocations)
	}

	apptType := appointmentTypeMap[*apptType]
	var locations []Location
	for _, location := range strings.Split(*rawLocations, ",") {
		locations = append(locations, locationMap[location])
	}

	allocatorOpts := chromedp.DefaultExecAllocatorOptions[:]
	var ctxOpts []chromedp.ContextOption
	if !*headless {
		allocatorOpts = append(allocatorOpts, chromedp.Flag("headless", false))
	}
	if *debug {
		ctxOpts = append(ctxOpts, chromedp.WithDebugf(log.Printf))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, allocatorOpts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx, ctxOpts...)
	defer cancel()

	// Open the first (empty) tab.
	if err := chromedp.Run(ctx); err != nil {
		log.Fatalf("Failed to open first tab: %v", err)
	}

	for {
		isAvailable, location, err := checkAllLocations(ctx, apptType, locations, time.Duration(*timeout)*time.Second, "" /* screenshotPath */)
		if err != nil {
			log.Fatalf("Failed to check all locations: %v", err)
		}
		if isAvailable {
			log.Printf("Found available location: %v", *location)

			if *discordWebhook != "" {
				username := discordWebhookUsername
				content := fmt.Sprintf("Found appointment at location %q", location)
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
