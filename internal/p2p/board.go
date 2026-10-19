package p2p

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// Board is where the two ends leave notes for each other.
type Board interface {
	Post(ctx context.Context, topic, content string) error
	Watch(ctx context.Context, topic string) (<-chan string, error)
}

// DefaultRelays are public Nostr relays. Notes go to all of them, so one relay
// being down or slow does not matter.
var DefaultRelays = []string{
	"wss://relay.damus.io",
	"wss://nos.lol",
	"wss://relay.primal.net",
	"wss://nostr.mom",
	"wss://relay.snort.social",
}

// noteKind is in Nostr's ephemeral range: relays pass these on without storing them.
const noteKind = 21877

// Nostr posts notes with a throwaway key, so relays cannot link sessions together.
type Nostr struct {
	pool    *nostr.SimplePool
	relays  []string
	signKey string
	pubkey  string
}

// NewNostr connects lazily to the given relays.
func NewNostr(ctx context.Context, relays []string) (*Nostr, error) {
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return nil, err
	}
	return &Nostr{pool: nostr.NewSimplePool(ctx), relays: relays, signKey: sk, pubkey: pk}, nil
}

// Post succeeds when at least one relay accepted the note.
func (n *Nostr) Post(ctx context.Context, topic, content string) error {
	ev := nostr.Event{
		PubKey:    n.pubkey,
		CreatedAt: nostr.Now(),
		Kind:      noteKind,
		Tags:      nostr.Tags{{"x", topic}},
		Content:   content,
	}
	if err := ev.Sign(n.signKey); err != nil {
		return err
	}
	var lastErr error
	accepted := 0
	for res := range n.pool.PublishMany(ctx, n.relays, ev) {
		if res.Error != nil {
			lastErr = res.Error
			continue
		}
		accepted++
	}
	if accepted == 0 {
		return fmt.Errorf("no relay accepted the note: %v", lastErr)
	}
	return nil
}

// Watch streams the content of new notes on topic until ctx ends.
func (n *Nostr) Watch(ctx context.Context, topic string) (<-chan string, error) {
	since := nostr.Now() - 30
	events := n.pool.SubscribeMany(ctx, n.relays, nostr.Filter{
		Kinds: []int{noteKind},
		Tags:  nostr.TagMap{"x": {topic}},
		Since: &since,
	})
	out := make(chan string)
	go func() {
		defer close(out)
		seen := map[string]bool{}
		for ev := range events {
			if seen[ev.ID] || ev.PubKey == n.pubkey {
				continue
			}
			seen[ev.ID] = true
			select {
			case out <- ev.Content:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}
