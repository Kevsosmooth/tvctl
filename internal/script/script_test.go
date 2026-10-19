package script

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tvremote/internal/uitree"
)

const sample = `
name: stremio
steps:
  - key: [wake, home]
  - install: apks/stremio.apk
  - launch: com.stremio.one
  - wait: 10ms
  - wait_for: Sign in
    timeout: 5s
  - click: Sign in
  - text: user@example.com
  - screenshot: shots/{device}.png
`

type fakeTarget struct {
	log      []string
	failStep string
}

func (f *fakeTarget) rec(s string) error {
	f.log = append(f.log, s)
	if f.failStep != "" && strings.HasPrefix(s, f.failStep) {
		return errors.New("boom")
	}
	return nil
}

func (f *fakeTarget) Keys(_ context.Context, n ...string) error {
	return f.rec("key " + strings.Join(n, ","))
}
func (f *fakeTarget) Text(_ context.Context, s string) error { return f.rec("text " + s) }
func (f *fakeTarget) Click(_ context.Context, q uitree.Query) (uitree.Element, error) {
	return uitree.Element{}, f.rec("click " + q.Text)
}
func (f *fakeTarget) WaitFor(_ context.Context, q uitree.Query, d time.Duration) (uitree.Element, error) {
	return uitree.Element{}, f.rec("wait_for " + q.Text + " " + d.String())
}
func (f *fakeTarget) Launch(_ context.Context, p string) error    { return f.rec("launch " + p) }
func (f *fakeTarget) Stop(_ context.Context, p string) error      { return f.rec("stop " + p) }
func (f *fakeTarget) Install(_ context.Context, p string) error   { return f.rec("install " + p) }
func (f *fakeTarget) Uninstall(_ context.Context, p string) error { return f.rec("uninstall " + p) }
func (f *fakeTarget) Shell(_ context.Context, c string) (string, error) {
	return "", f.rec("shell " + c)
}
func (f *fakeTarget) SaveScreenshot(_ context.Context, path string, _ int) error {
	return f.rec("screenshot " + path)
}

func TestParseAndRun(t *testing.T) {
	s, err := Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	f := &fakeTarget{}
	n, err := s.Run(context.Background(), f, Env{Device: "smith-tv", Dir: dir})
	if err != nil || n != 8 {
		t.Fatalf("ran %d steps, err %v", n, err)
	}
	want := []string{
		"key wake,home",
		"install " + filepath.Join(dir, "apks/stremio.apk"),
		"launch com.stremio.one",
		"wait_for Sign in 5s",
		"click Sign in",
		"text user@example.com",
		"screenshot " + filepath.Join(dir, "shots/smith-tv.png"),
	}
	if strings.Join(f.log, "|") != strings.Join(want, "|") {
		t.Errorf("log = %v", f.log)
	}
}

func TestRunStopsAtFirstFailure(t *testing.T) {
	s, _ := Parse([]byte(sample))
	f := &fakeTarget{failStep: "launch"}
	n, err := s.Run(context.Background(), f, Env{Dir: t.TempDir()})
	if err == nil || n != 2 || !strings.Contains(err.Error(), "step 3 (launch)") {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestParseRejectsBadSteps(t *testing.T) {
	bad := map[string]string{
		"two actions":   "steps:\n  - key: [home]\n    text: hi\n",
		"unknown field": "steps:\n  - reboot: true\n",
		"empty":         "name: x\n",
	}
	for name, src := range bad {
		if _, err := Parse([]byte(src)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
