package tv

import (
	"context"
	"strings"
	"testing"

	"tvremote/internal/adb"
)

type scriptedRunner struct {
	replies map[string]string // joined-args substring -> output
	calls   []string
}

func (s *scriptedRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	s.calls = append(s.calls, joined)
	for k, v := range s.replies {
		if strings.Contains(joined, k) {
			return []byte(v), nil
		}
	}
	return nil, nil
}

func newTestDevice(replies map[string]string) (*Device, *scriptedRunner) {
	r := &scriptedRunner{replies: replies}
	return New(adb.NewClient(r), "tv:5555"), r
}

func TestInputTextArg(t *testing.T) {
	cases := map[string]string{
		"stremio":      `'stremio'`,
		"hello world":  `'hello%sworld'`,
		"it's; reboot": `'it'\''s;%sreboot'`,
	}
	for in, want := range cases {
		if got := inputTextArg(in); got != want {
			t.Errorf("inputTextArg(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestKeysValidatesBeforeSending(t *testing.T) {
	d, r := newTestDevice(nil)
	if err := d.Keys(context.Background(), "home", "bogus"); err == nil {
		t.Fatal("expected error for unknown key")
	}
	if len(r.calls) != 0 {
		t.Errorf("no keys should be sent when one is invalid, got %v", r.calls)
	}
	if err := d.Keys(context.Background(), "down", "ok"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(r.calls[1], "input keyevent KEYCODE_DPAD_CENTER") {
		t.Errorf("calls = %v", r.calls)
	}
}

func TestLaunchRejectsBadPackage(t *testing.T) {
	d, r := newTestDevice(nil)
	if err := d.Launch(context.Background(), "com.foo;reboot"); err == nil {
		t.Fatal("expected invalid package error")
	}
	if len(r.calls) != 0 {
		t.Errorf("calls = %v", r.calls)
	}
}

func TestLaunchFallsBackToPhoneLauncher(t *testing.T) {
	d, r := newTestDevice(map[string]string{
		"LEANBACK_LAUNCHER": "No activity found",
		"category.LAUNCHER": "priority=0 preferredOrder=0\ncom.stremio.one/.MainActivity\n",
		"am start":          "Starting: Intent { cmp=com.stremio.one/.MainActivity }",
	})
	if err := d.Launch(context.Background(), "com.stremio.one"); err != nil {
		t.Fatal(err)
	}
	if last := r.calls[len(r.calls)-1]; !strings.HasSuffix(last, "am start -n com.stremio.one/.MainActivity") {
		t.Errorf("calls = %v", r.calls)
	}
}

func TestLaunchWithoutActivityFails(t *testing.T) {
	d, _ := newTestDevice(map[string]string{"resolve-activity": "No activity found"})
	if err := d.Launch(context.Background(), "com.amazon.tv.settings.v2"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParsePackages(t *testing.T) {
	got := parsePackages("package:com.b\r\npackage:com.a\n\n")
	if strings.Join(got, ",") != "com.a,com.b" {
		t.Errorf("got %v", got)
	}
}

func TestParseFocus(t *testing.T) {
	out := "  mCurrentFocus=Window{a1b2c3 u0 com.stremio.one/com.stremio.MainActivity}\n"
	if got := parseFocus(out); got != "com.stremio.one/com.stremio.MainActivity" {
		t.Errorf("got %q", got)
	}
	if got := parseFocus("mCurrentFocus=null"); got != "" {
		t.Errorf("null focus = %q", got)
	}
}

func TestExtractHierarchy(t *testing.T) {
	raw := []byte("<?xml version='1.0'?><hierarchy rotation=\"0\"></hierarchy>UI hierchary dumped to: /dev/tty\n")
	got, err := extractHierarchy(raw)
	if err != nil || !strings.HasSuffix(string(got), "</hierarchy>") {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := extractHierarchy([]byte("ERROR: null root node\n")); err == nil {
		t.Error("expected error")
	}
}
