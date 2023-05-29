package lib

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

const (
	apptUrl                    = "https://skiptheline.ncdot.gov/"
	makeApptButtonSel          = "button#cmdMakeAppt"
	locationAvailableClassName = "Active-Unit"
)

func isLocationNodeEnabled(node *cdp.Node) bool {
	return strings.Contains(node.AttributeValue("class"), locationAvailableClassName)
}

type Client struct {
	// chromedp browser context.
	ctx context.Context
}

func NewClient(ctx context.Context, headless, debug bool) (*Client, context.CancelFunc, error) {
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

	return &Client{ctx}, cancel, nil
}

func isLocationAvailable(ctx context.Context, apptType AppointmentType, location Location, timeout time.Duration) (bool, error) {
	// Navigate to the main page.
	if _, err := chromedp.RunResponse(ctx, chromedp.Navigate(apptUrl)); err != nil {
		return false, err
	}

	// Click the "Make Appointment" button once it is visible.
	if _, err := chromedp.RunResponse(ctx, chromedp.Click(makeApptButtonSel, chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		return false, err
	}

	// Click the appointment type.
	if _, err := chromedp.RunResponse(ctx, chromedp.Click(apptType.ToSelector(), chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		return false, err
	}

	// Wait for the location and read the node.
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx,
		//browser.GrantPermissions([]browser.PermissionType{browser.PermissionTypeGeolocation}).WithOrigin(apptUrl),
		chromedp.WaitVisible(location.ToSelector(), chromedp.ByQuery),
		chromedp.Nodes(location.ToSelector(), &nodes, chromedp.NodeVisible, chromedp.ByQuery),
	); err != nil {
		return false, err
	}

	// if screenshotPath != "" {
	// 	var screenshotBuf []byte
	// 	if err := chromedp.Run(ctx, chromedp.FullScreenshot(&screenshotBuf, 90)); err != nil {
	// 		return false, err
	// 	}
	// 	if err := os.WriteFile(screenshotPath, screenshotBuf, 0o644); err != nil {
	// 		return false, fmt.Errorf("failed to write screenshot to path %q: %w", screenshotPath, err)
	// 	}
	// }

	if len(nodes) == 0 {
		return false, fmt.Errorf("found no nodes for location %q - is it even valid?", location)
	} else if len(nodes) != 1 {
		return false, fmt.Errorf("found multiple nodes for location %q: %+v", location, nodes)
	}

	isAvailable := isLocationNodeEnabled(nodes[0])

	return isAvailable, nil
}

func (c Client) CheckLocations(apptType AppointmentType, locations []Location, timeout time.Duration) (*Location, error) {
	// Timeout for all locations.
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()

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
			isAvailable, err := isLocationAvailable(ctx, apptType, location, timeout)
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
			return nil, result.err
		}
		if result.isAvailable {
			// We found a location!
			return &location, nil
		}
		log.Printf("Location %q has no appointments available!", location)
	}

	return nil, nil
}
