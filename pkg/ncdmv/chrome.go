package ncdmv

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

func NewChromeContext(ctx context.Context, headless, disableGpu, debug bool) (context.Context, context.CancelFunc, error) {
	allocatorOpts := chromedp.DefaultExecAllocatorOptions[:]
	var ctxOpts []chromedp.ContextOption
	if !headless {
		allocatorOpts = append(allocatorOpts, chromedp.Flag("headless", false))
	}
	if disableGpu {
		allocatorOpts = append(allocatorOpts, chromedp.DisableGPU)
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

	return ctx, cancel, nil
}
