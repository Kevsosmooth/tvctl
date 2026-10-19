package p2p

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestCodeRoundTrip(t *testing.T) {
	c, err := NewCode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseCode(" " + c.String() + "\n")
	if err != nil || got != c {
		t.Fatalf("got %v, %v", got, err)
	}
	if _, err := ParseCode("not-a-code"); err == nil {
		t.Error("expected an error")
	}
}

func TestNoteRejectsOtherCodeAndStale(t *testing.T) {
	c1, _ := NewCode()
	c2, _ := NewCode()
	a1, _ := c1.aead()
	a2, _ := c2.aead()
	sealed, err := seal(a1, note{Kind: kindOffer, Session: "s1", SDP: "v=0"})
	if err != nil {
		t.Fatal(err)
	}
	if n, err := open(a1, sealed, time.Now()); err != nil || n.SDP != "v=0" {
		t.Fatalf("own note: %v %v", n, err)
	}
	if _, err := open(a2, sealed, time.Now()); err == nil {
		t.Error("note opened with the wrong code")
	}
	if _, err := open(a1, sealed, time.Now().Add(5*time.Minute)); err == nil {
		t.Error("stale note accepted")
	}
}

// local keeps test connections on this machine.
var local = Options{loopbackOnly: true}

// memBoard delivers every post to every watcher of the same topic.
type memBoard struct {
	mu       sync.Mutex
	watchers map[string][]chan string
}

func (b *memBoard) Post(_ context.Context, topic, content string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, w := range b.watchers[topic] {
		w <- content
	}
	return nil
}

func (b *memBoard) Watch(_ context.Context, topic string) (<-chan string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan string, 16)
	if b.watchers == nil {
		b.watchers = map[string][]chan string{}
	}
	b.watchers[topic] = append(b.watchers[topic], ch)
	return ch, nil
}

// echoServer stands in for the TV's adb.
func echoServer(t *testing.T) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() { _, _ = io.Copy(c, c); c.Close() }()
		}
	}()
	return ln.Addr().String()
}

func TestLinkCarriesLargeStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	board := &memBoard{}
	code, _ := NewCode()
	go func() { _ = Serve(ctx, board, code, echoServer(t), local, t.Logf) }()
	time.Sleep(100 * time.Millisecond) // let Serve start watching

	link, err := Dial(ctx, board, code, local)
	if err != nil {
		t.Fatal(err)
	}
	defer link.Close()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = link.Forward(ctx, ln) }()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	payload := make([]byte, 4<<20) // bigger than the send queue limit
	_, _ = rand.Read(payload)
	go func() { _, _ = conn.Write(payload) }()
	got := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("stream came back changed")
	}
	if r := link.Route(); r == "relay" || r == "unknown" {
		t.Errorf("route = %s", r)
	}
}

func TestDialWithWrongCodeTimesOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	board := &memBoard{}
	tvCode, _ := NewCode()
	otherCode, _ := NewCode()
	go func() { _ = Serve(ctx, board, tvCode, echoServer(t), local, t.Logf) }()
	if _, err := Dial(ctx, board, otherCode, local); err == nil {
		t.Fatal("connected with the wrong code")
	}
}
