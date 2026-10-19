// Package fanout runs one action on many devices in parallel with a concurrency limit.
package fanout

import (
	"context"
	"sync"
	"time"

	"tvremote/internal/registry"
)

// Result is one device's outcome. Value carries command-specific output.
type Result struct {
	Device   string `json:"device"`
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
	Value    any    `json:"value,omitempty"`
	Duration string `json:"duration"`
}

// Action runs on a single device.
type Action func(ctx context.Context, d registry.Device) (any, error)

// Run executes action on every device, at most limit at a time, and returns results in
// the input order. One device failing never stops the others.
func Run(ctx context.Context, devices []registry.Device, limit int, action Action) []Result {
	if limit < 1 {
		limit = 1
	}
	results := make([]Result, len(devices))
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i, d := range devices {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = runOne(ctx, d, action)
		}()
	}
	wg.Wait()
	return results
}

func runOne(ctx context.Context, d registry.Device, action Action) Result {
	start := time.Now()
	value, err := action(ctx, d)
	r := Result{Device: d.Name, OK: err == nil, Value: value, Duration: time.Since(start).Round(time.Millisecond).String()}
	if err != nil {
		r.Error = err.Error()
	}
	return r
}

// AllOK reports whether every device succeeded.
func AllOK(results []Result) bool {
	for _, r := range results {
		if !r.OK {
			return false
		}
	}
	return true
}
