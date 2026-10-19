package tv

import (
	"context"
	"fmt"
	"time"

	"tvremote/internal/uitree"
)

// Shell runs a raw command line through the device shell. Operator-only: tvctl never
// builds this string from data read off a TV.
func (d *Device) Shell(ctx context.Context, cmdline string) (string, error) {
	if cmdline == "" {
		return "", fmt.Errorf("shell command is empty")
	}
	return d.shell(ctx, cmdline)
}

// WaitFor polls the screen until an element matching q appears or timeout passes.
func (d *Device) WaitFor(ctx context.Context, q uitree.Query, timeout time.Duration) (uitree.Element, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		elements, err := d.UI(ctx)
		if err == nil {
			el, findErr := uitree.Find(elements, q)
			if findErr == nil {
				return el, nil
			}
			err = findErr
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return uitree.Element{}, ctx.Err()
		case <-time.After(700 * time.Millisecond):
		}
	}
	return uitree.Element{}, fmt.Errorf("timed out after %s: %w", timeout, lastErr)
}
