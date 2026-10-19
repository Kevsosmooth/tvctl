package fanout

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"tvremote/internal/registry"
)

func TestRunKeepsOrderAndIsolatesFailures(t *testing.T) {
	devices := []registry.Device{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	results := Run(context.Background(), devices, 2, func(_ context.Context, d registry.Device) (any, error) {
		if d.Name == "b" {
			return nil, errors.New("offline")
		}
		return d.Name + "-done", nil
	})
	if results[0].Value != "a-done" || results[2].Value != "c-done" {
		t.Errorf("values = %+v", results)
	}
	if results[1].OK || results[1].Error != "offline" {
		t.Errorf("b = %+v", results[1])
	}
	if AllOK(results) {
		t.Error("AllOK should be false")
	}
}

func TestRunRespectsLimit(t *testing.T) {
	devices := make([]registry.Device, 8)
	var running, peak int32
	Run(context.Background(), devices, 3, func(context.Context, registry.Device) (any, error) {
		n := atomic.AddInt32(&running, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if n <= p || atomic.CompareAndSwapInt32(&peak, p, n) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&running, -1)
		return nil, nil
	})
	if peak > 3 {
		t.Errorf("peak concurrency %d > 3", peak)
	}
}
