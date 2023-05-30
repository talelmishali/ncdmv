package lib

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/gtuk/discordwebhook"
)

const (
	discordWebhookUsername = "ncdmv-bot"

	makeApptUrl                   = "https://skiptheline.ncdot.gov/"
	makeApptButtonSelector        = "button#cmdMakeAppt"
	locationAvailableClassName    = "Active-Unit"
	appointmentCalendarSelector   = "div.CalendarDateModel.hasDatepicker"
	appointmentDayButtonSelector  = `td[data-handler="selectDay"]`
	appointmentMonthAttributeName = "data-month"
	appointmentYearAttributeName  = "data-year"
	appointmentTimeSelectSelector = "div.AppointmentTime select"
)

// isLocationNodeEnabled returns "true" if the location DOM node is available/clickable.
func isLocationNodeEnabled(node *cdp.Node) bool {
	return strings.Contains(node.AttributeValue("class"), locationAvailableClassName)
}

type Client struct {
	// chromedp browser context.
	ctx                      context.Context
	discordWebhook           string
	stopOnFailure            bool
	appointmentNotifications map[Appointment]bool
}

func NewClient(ctx context.Context, discordWebhook string, headless, debug, stopOnFailure bool) (*Client, context.CancelFunc, error) {
	allocatorOpts := chromedp.DefaultExecAllocatorOptions[:]
	var ctxOpts []chromedp.ContextOption
	if !headless {
		allocatorOpts = append(allocatorOpts, chromedp.Flag("headless", false))
	}
	if debug {
		ctxOpts = append(ctxOpts, chromedp.WithDebugf(log.Printf))
	}

	ctx, cancel1 := chromedp.NewExecAllocator(ctx, allocatorOpts...)
	ctx, cancel2 := chromedp.NewContext(ctx, ctxOpts...)
	cancel := func() { cancel2(); cancel1() }

	// Open the first (empty) tab.
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to open first tab: %w", err)
	}

	return &Client{
		ctx:                      ctx,
		discordWebhook:           discordWebhook,
		stopOnFailure:            stopOnFailure,
		appointmentNotifications: make(map[Appointment]bool),
	}, cancel, nil
}

func isLocationAvailable(ctx context.Context, apptType AppointmentType, location Location, timeout time.Duration) (bool, error) {
	// Navigate to the main page.
	if _, err := chromedp.RunResponse(ctx, chromedp.Navigate(makeApptUrl)); err != nil {
		return false, err
	}

	// Click the "Make Appointment" button once it is visible.
	if _, err := chromedp.RunResponse(ctx, chromedp.Click(makeApptButtonSelector, chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		return false, err
	}

	// Click the appointment type button.
	if _, err := chromedp.RunResponse(ctx, chromedp.Click(apptType.ToSelector(), chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		return false, err
	}

	// Wait for the location and read the node.
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx,
		chromedp.WaitVisible(location.ToSelector(), chromedp.ByQuery),
		chromedp.Nodes(location.ToSelector(), &nodes, chromedp.NodeVisible, chromedp.ByQuery),
	); err != nil {
		return false, err
	}

	if len(nodes) == 0 {
		return false, fmt.Errorf("found no nodes for location %q - is it even valid?", location)
	} else if len(nodes) != 1 {
		return false, fmt.Errorf("found multiple nodes for location %q: %+v", location, nodes)
	}

	return isLocationNodeEnabled(nodes[0]), nil
}

type Appointment struct {
	Location Location
	Time     time.Time
}

func (a Appointment) String() string {
	return fmt.Sprintf("Appointment(location: %q, time: %s)", a.Location, a.Time)
}

// findAvailableAppointmentDates finds all available dates on the location calendar page.
func findAvailableAppointmentDates(ctx context.Context) ([]time.Time, error) {
	var calendarHtml string
	if err := chromedp.Run(ctx,
		// Wait for the time select element to appear.
		chromedp.WaitVisible(appointmentTimeSelectSelector, chromedp.ByQuery),

		// Extract the HTML for the calendar.
		chromedp.InnerHTML(appointmentCalendarSelector, &calendarHtml, chromedp.ByQuery),
	); err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(calendarHtml))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find the available dates.
	var availableDates []time.Time
	doc.Find(appointmentDayButtonSelector).Each(func(i int, s *goquery.Selection) {
		// Extract the day, month, and year from the DOM node.
		dayStr := s.Text()
		monthStr, ok := s.Attr(appointmentMonthAttributeName)
		if !ok {
			log.Print("No month attribute")
			return
		}
		yearStr, ok := s.Attr(appointmentYearAttributeName)
		if !ok {
			log.Print("No year attribute")
			return
		}

		// Parse the date parts.
		day, err := strconv.ParseInt(dayStr, 10, 32)
		if err != nil {
			log.Printf("Invalid day: %s", s.Text())
			return
		}
		month, err := strconv.ParseInt(monthStr, 10, 32)
		if err != nil {
			log.Printf("Invalid month: %s", s.Text())
			return
		}
		year, err := strconv.ParseInt(yearStr, 10, 32)
		if err != nil {
			log.Printf("Invalid year: %s", s.Text())
			return
		}

		// Month is 0-indexed.
		d := time.Date(int(year), time.Month(month+1), int(day), 0, 0, 0, 0, time.UTC)

		availableDates = append(availableDates, d)
	})

	return availableDates, nil
}

// findAvailableAppointments finds all available appointment dates for the given location.
//
// NOTE: Currently does not parse the appointment time slots - just dates. Also, this does not look at
// later months.
func findAvailableAppointments(ctx context.Context, apptType AppointmentType, location Location, timeout time.Duration) (appointments []*Appointment, _ error) {
	isAvailable, err := isLocationAvailable(ctx, apptType, location, timeout)
	if err != nil {
		return nil, err
	}
	if !isAvailable {
		return nil, nil
	}

	// At this point, we are on the locations page.
	// Click the location button.
	if _, err := chromedp.RunResponse(ctx, chromedp.Click(location.ToSelector()), chromedp.Sleep(3*time.Second)); err != nil {
		return nil, err
	}

	availableDates, err := findAvailableAppointmentDates(ctx)
	if err != nil {
		return nil, err
	}

	for _, d := range availableDates {
		appointments = append(appointments, &Appointment{
			Location: location,
			Time:     d,
		})
	}

	return appointments, nil
}

// RunForLocations finds all available appointments across the given locations.
//
// NOTE: For now, this only looks at _appointment dates_ and only considers the first available month.
func (c Client) RunForLocations(apptType AppointmentType, locations []Location, timeout time.Duration) ([]*Appointment, error) {
	// Common timeout for all locations.
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()

	// Setup a seperate tab context for each location. The tabs will be closed when this function
	// returns.
	var locationCtxs []context.Context
	for range locations {
		ctx, cancel := chromedp.NewContext(ctx)
		defer cancel()
		locationCtxs = append(locationCtxs, ctx)
	}

	type locationResult struct {
		idx          int
		appointments []*Appointment
		err          error
	}
	resultChan := make(chan locationResult)

	// Spawn a goroutine for each location. Each locations is processed in a separate
	// browser tab.
	for i, location := range locations {
		i := i
		location := location
		ctx := locationCtxs[i]
		go func() {
			appointments, err := findAvailableAppointments(ctx, apptType, location, timeout)
			resultChan <- locationResult{
				idx:          i,
				appointments: appointments,
				err:          err,
			}
		}()
	}

	// Extract appointments from all of the locations.
	var appointments []*Appointment
	for i := 0; i < len(locations); i++ {
		result := <-resultChan
		location := locations[result.idx]

		if result.err != nil {
			return nil, result.err
		}
		if len(result.appointments) == 0 {
			log.Printf("Location %q has no appointments available", location)
		}
		appointments = append(appointments, result.appointments...)
	}

	return appointments, nil
}

// Start runs the NC DMV client for the given locations. A search will be run for all locations based on
// the specified interval.
//
// Note that this method will block indefinitely. If you want to just run a single search, use RunForLocations.
//
// If "stopOnFailure" is true for this client, this method will return any error encountered.
func (c Client) Start(apptType AppointmentType, locations []Location, timeout, interval time.Duration) error {
	appointments, err := c.RunForLocations(apptType, locations, timeout)
	if err != nil {
		if c.stopOnFailure {
			return fmt.Errorf("failed to check locations: %w", err)
		} else {
			log.Printf("Failed to check locations: %v", err)
		}
	}

	for _, appointment := range appointments {
		log.Printf("Found appointment: %q", appointment)

		// Send a notification for this appointment if we haven't already done so.
		if !c.appointmentNotifications[*appointment] {
			if c.discordWebhook != "" {
				username := discordWebhookUsername
				content := fmt.Sprintf("Found appointment: %q", appointment)
				if err := discordwebhook.SendMessage(c.discordWebhook, discordwebhook.Message{
					Username: &username,
					Content:  &content,
				}); err != nil {
					log.Printf("Failed to send message to Discord webhook %q: %v", c.discordWebhook, err)
				}
			}
			c.appointmentNotifications[*appointment] = true
		}

		log.Printf("Sleeping for %v between checks...", interval)
		time.Sleep(interval)
	}

	return nil
}
