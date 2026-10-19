package script

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"tvremote/internal/tv"
	"tvremote/internal/uitree"
)

// Target is what a script drives; *tv.Device satisfies it.
type Target interface {
	Keys(ctx context.Context, names ...string) error
	Text(ctx context.Context, text string) error
	Click(ctx context.Context, q uitree.Query) (uitree.Element, error)
	WaitFor(ctx context.Context, q uitree.Query, timeout time.Duration) (uitree.Element, error)
	Launch(ctx context.Context, pkg string) error
	Stop(ctx context.Context, pkg string) error
	Install(ctx context.Context, apkPath string) error
	Uninstall(ctx context.Context, pkg string) error
	Shell(ctx context.Context, cmdline string) (string, error)
	SaveScreenshot(ctx context.Context, path string, maxWidth int) error
}

// Env gives steps the device name (for {device} in paths) and the script's folder.
type Env struct {
	Device string
	Dir    string
}

// Run executes every step in order and stops at the first failure.
func (s *Script) Run(ctx context.Context, t Target, env Env) (int, error) {
	for i, st := range s.Steps {
		if err := runStep(ctx, t, st, env); err != nil {
			return i, fmt.Errorf("step %d (%s): %w", i+1, st.Kind(), err)
		}
	}
	return len(s.Steps), nil
}

func runStep(ctx context.Context, t Target, st Step, env Env) error {
	switch st.Kind() {
	case "key":
		return t.Keys(ctx, st.Key...)
	case "text":
		return t.Text(ctx, st.Text)
	case "click":
		_, err := t.Click(ctx, uitree.Query{Text: st.Click})
		return err
	case "wait_for":
		_, err := t.WaitFor(ctx, uitree.Query{Text: st.WaitFor}, timeoutOr(st.Timeout, 20*time.Second))
		return err
	case "wait":
		return sleep(ctx, st.Wait)
	case "launch":
		return t.Launch(ctx, st.Launch)
	case "stop":
		return t.Stop(ctx, st.Stop)
	case "install":
		return t.Install(ctx, resolve(env.Dir, st.Install))
	case "uninstall":
		return t.Uninstall(ctx, st.Uninstall)
	case "shell":
		_, err := t.Shell(ctx, st.Shell)
		return err
	case "screenshot":
		return t.SaveScreenshot(ctx, expand(resolve(env.Dir, st.Screenshot), env), tv.ShotWidth)
	}
	return fmt.Errorf("unknown step")
}

func sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func timeoutOr(d, fallback time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return fallback
}

func resolve(dir, p string) string {
	if filepath.IsAbs(p) || dir == "" {
		return p
	}
	return filepath.Join(dir, p)
}

func expand(p string, env Env) string {
	return strings.ReplaceAll(p, "{device}", env.Device)
}
