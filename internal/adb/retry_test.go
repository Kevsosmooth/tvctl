package adb

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// flakyRunner fails the first n device commands with the given error.
type flakyRunner struct {
	failures int
	err      error
	calls    []string
}

func (f *flakyRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	f.calls = append(f.calls, joined)
	if args[0] == "connect" {
		return []byte("already connected to " + args[1]), nil
	}
	if f.failures > 0 {
		f.failures--
		return nil, f.err
	}
	return []byte("ok"), nil
}

func TestShellRetriesWhenDeviceOffline(t *testing.T) {
	r := &flakyRunner{failures: 1, err: errors.New("adb -s tv:5555 shell input: exit status 1: adb: device offline")}
	out, err := NewClient(r).Shell(context.Background(), "tv:5555", "input", "keyevent", "3")
	if err != nil || out != "ok" {
		t.Fatalf("out %q, err %v", out, err)
	}
	if len(r.calls) != 3 || !strings.HasPrefix(r.calls[1], "connect") {
		t.Errorf("calls = %v", r.calls)
	}
}

func TestShellDoesNotRepeatCommandsThatRan(t *testing.T) {
	r := &flakyRunner{failures: 1, err: errors.New("adb -s tv:5555 shell foo: exit status 127: /system/bin/sh: foo: not found")}
	if _, err := NewClient(r).Shell(context.Background(), "tv:5555", "foo"); err == nil {
		t.Fatal("expected error")
	}
	if len(r.calls) != 1 {
		t.Errorf("command was repeated: %v", r.calls)
	}
}
