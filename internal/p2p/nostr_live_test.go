package p2p

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// TestRelaysLive checks each public relay accepts and passes on our notes.
// Run with TVREMOTE_LIVE=1; it needs the internet.
func TestRelaysLive(t *testing.T) {
	if os.Getenv("TVREMOTE_LIVE") == "" {
		t.Skip("set TVREMOTE_LIVE=1 to talk to public relays")
	}
	UseResolverFallback()
	for _, url := range DefaultRelays {
		t.Run(url, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			watcher, err := NewNostr(ctx, []string{url})
			if err != nil {
				t.Fatal(err)
			}
			poster, err := NewNostr(ctx, []string{url})
			if err != nil {
				t.Fatal(err)
			}
			topic := nostr.GeneratePrivateKey()[:32]
			notes, _ := watcher.Watch(ctx, topic)
			time.Sleep(2 * time.Second)
			start := time.Now()
			if err := poster.Post(ctx, topic, "ping"); err != nil {
				t.Fatalf("post: %v", err)
			}
			select {
			case got := <-notes:
				t.Logf("delivered %q in %s", got, time.Since(start).Round(time.Millisecond))
			case <-ctx.Done():
				t.Fatal("posted but never delivered")
			}
		})
	}
}
