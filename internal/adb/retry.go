package adb

import (
	"context"
	"regexp"
	"time"
)

const deviceAttempts = 3

// onDevice runs an adb command against one device. TVs on Wi-Fi drop off adb for a
// second or two now and then; adb then refuses the command before anything reaches the
// TV, so it is safe to wait, reconnect and send it again.
func (c *Client) onDevice(ctx context.Context, serial string, args ...string) ([]byte, error) {
	full := append([]string{"-s", serial}, args...)
	var out []byte
	var err error
	for attempt := range deviceAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
			if cerr := c.Connect(ctx, serial); cerr != nil {
				continue
			}
		}
		out, err = c.runner.Run(ctx, full...)
		if err == nil || !notReachable(err) {
			return out, err
		}
	}
	return out, err
}

// notReachable matches only adb's own errors for a device it could not send the command
// to, never a failure from a command that ran on the TV (which may not be safe to repeat).
var notReachablePattern = regexp.MustCompile(`(adb|error): device offline|error: device '[^']*' not found`)

func notReachable(err error) bool {
	return notReachablePattern.MatchString(err.Error())
}
