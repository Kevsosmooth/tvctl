package tv

import (
	"context"
	"strings"
	"testing"

	"tvremote/internal/adb"
)

// sequenceRunner answers screencap calls from a queue and accepts everything else.
type sequenceRunner struct {
	shots [][]byte
	calls []string
}

func (s *sequenceRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	s.calls = append(s.calls, joined)
	if !strings.Contains(joined, "screencap") {
		return []byte("already connected to tv"), nil
	}
	shot := s.shots[0]
	s.shots = s.shots[1:]
	return shot, nil
}

func TestScreenshotRetriesCutOffImage(t *testing.T) {
	whole := []byte("\x89PNG....data....IEND\xaeB`\x82")
	r := &sequenceRunner{shots: [][]byte{whole[:10], whole}}
	d := New(adb.NewClient(r), "tv:5555")
	png, err := d.Screenshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(png) != string(whole) {
		t.Errorf("got %q", png)
	}
	if !strings.HasPrefix(r.calls[1], "connect") {
		t.Errorf("expected a reconnect between attempts, calls = %v", r.calls)
	}
}

func TestScreenshotNonImageFailsFast(t *testing.T) {
	r := &sequenceRunner{shots: [][]byte{[]byte("Error: permission denied\n")}}
	d := New(adb.NewClient(r), "tv:5555")
	if _, err := d.Screenshot(context.Background()); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("err = %v", err)
	}
	if len(r.calls) != 1 {
		t.Errorf("calls = %v", r.calls)
	}
}
