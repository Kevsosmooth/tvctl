// Package adb wraps the adb binary for talking to TVs over the network.
package adb

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes the adb binary. Tests swap in a fake.
type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// ExecRunner runs a real adb binary.
type ExecRunner struct {
	Path string
}

func (r ExecRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	path := r.Path
	if path == "" {
		path = "adb"
	}
	cmd := exec.CommandContext(ctx, path, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		return stdout.Bytes(), fmt.Errorf("adb %s: %w: %s", firstArgs(args), err, msg)
	}
	return stdout.Bytes(), nil
}

// firstArgs keeps error messages short and free of typed text payloads.
func firstArgs(args []string) string {
	if len(args) > 4 {
		args = args[:4]
	}
	return strings.Join(args, " ")
}
