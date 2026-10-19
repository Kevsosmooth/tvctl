package tv

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tvremote/internal/imgscale"
)

const screenshotAttempts = 3

// ShotWidth is the default saved screenshot width: sharp enough to read every label,
// small enough for an AI agent to view (a Shield captures at 3840 wide).
const ShotWidth = 1280

var (
	pngHeader = []byte("\x89PNG")
	pngEnd    = []byte("IEND")
)

// Screenshot returns a PNG of the screen. DRM video shows as black. Big captures over
// Wi-Fi sometimes fail: the Shield at 4K can cut the stream short while adb reports
// success, or drop the connection for a second or two. Both are retried after a pause
// and a reconnect.
func (d *Device) Screenshot(ctx context.Context) ([]byte, error) {
	var lastErr error
	for attempt := range screenshotAttempts {
		if attempt > 0 {
			if err := d.recover(ctx, time.Duration(attempt)*time.Second); err != nil {
				return nil, err
			}
		}
		png, err := d.client.ExecOut(ctx, d.serial, "screencap", "-p")
		switch {
		case err != nil:
			lastErr = err
		case !bytes.HasPrefix(png, pngHeader):
			return nil, fmt.Errorf("screencap returned no image: %s", bytes.TrimSpace(firstLine(png)))
		case !completePNG(png):
			lastErr = fmt.Errorf("screencap image was cut off (%d bytes)", len(png))
		default:
			return png, nil
		}
	}
	return nil, lastErr
}

// recover waits out a dropped connection and reattaches; a failed reconnect is left
// for the next attempt to report.
func (d *Device) recover(ctx context.Context, wait time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
	}
	_ = d.Connect(ctx)
	return nil
}

func completePNG(png []byte) bool {
	return len(png) >= 12 && bytes.Equal(png[len(png)-8:len(png)-4], pngEnd)
}

// SaveScreenshot writes a PNG of the screen to path, shrunk to maxWidth when it is
// wider (0 keeps full size), creating folders as needed.
func (d *Device) SaveScreenshot(ctx context.Context, path string, maxWidth int) error {
	png, err := d.Screenshot(ctx)
	if err != nil {
		return err
	}
	if png, err = imgscale.FitWidth(png, maxWidth); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, png, 0o644)
}
