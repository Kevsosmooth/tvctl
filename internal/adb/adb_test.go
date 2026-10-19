package adb

import (
	"context"
	"strings"
	"testing"
)

type fakeRunner struct {
	out   string
	calls [][]string
}

func (f *fakeRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	return []byte(f.out), nil
}

func TestParseDevices(t *testing.T) {
	out := "List of devices attached\n" +
		"100.64.0.5:5555 device product:sabrina model:Chromecast device:sabrina transport_id:1\n" +
		"192.168.50.20:5555\tunauthorized transport_id:2\n" +
		"* daemon started successfully\n"
	got := ParseDevices(out)
	if len(got) != 2 {
		t.Fatalf("got %d devices, want 2", len(got))
	}
	if got[0].Serial != "100.64.0.5:5555" || got[0].State != "device" || got[0].Model != "Chromecast" {
		t.Errorf("first device = %+v", got[0])
	}
	if got[1].State != "unauthorized" {
		t.Errorf("second state = %q", got[1].State)
	}
}

func TestConnectReadsOutputText(t *testing.T) {
	cases := map[string]bool{
		"connected to 1.2.3.4:5555":           true,
		"already connected to 1.2.3.4:5555":   true,
		"failed to connect to '1.2.3.4:5555'": false,
		"failed to authenticate to 1.2.3.4":   false,
	}
	for out, ok := range cases {
		c := NewClient(&fakeRunner{out: out})
		err := c.Connect(context.Background(), "1.2.3.4:5555")
		if (err == nil) != ok {
			t.Errorf("Connect with %q: err=%v, want ok=%v", out, err, ok)
		}
	}
}

func TestInstallRequiresSuccess(t *testing.T) {
	c := NewClient(&fakeRunner{out: "Performing Streamed Install\nFailure [INSTALL_FAILED_INSUFFICIENT_STORAGE]"})
	err := c.Install(context.Background(), "s", "app.apk")
	if err == nil || !strings.Contains(err.Error(), "INSUFFICIENT_STORAGE") {
		t.Fatalf("err = %v", err)
	}
}

func TestShellTargetsSerial(t *testing.T) {
	f := &fakeRunner{out: "ok\r\n"}
	out, _ := NewClient(f).Shell(context.Background(), "tv:5555", "echo", "ok")
	if out != "ok" {
		t.Errorf("out = %q", out)
	}
	want := "-s tv:5555 shell echo ok"
	if got := strings.Join(f.calls[0], " "); got != want {
		t.Errorf("args = %q, want %q", got, want)
	}
}
