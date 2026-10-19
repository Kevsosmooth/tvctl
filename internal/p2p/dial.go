package p2p

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

// Link is the helper's end of a direct connection to one TV.
type Link struct {
	pc     *webrtc.PeerConnection
	lost   chan struct{}
	lostMu sync.Once
}

// Dial runs on the helper's computer. It posts an offer for the TV holding code
// and returns once the two ends are connected.
func Dial(ctx context.Context, board Board, code Code, opt Options) (*Link, error) {
	watchCtx, stopWatch := context.WithCancel(ctx)
	defer stopWatch()
	notes, err := board.Watch(watchCtx, code.topic())
	if err != nil {
		return nil, err
	}
	pc, err := newPeer(opt)
	if err != nil {
		return nil, err
	}
	l := &Link{pc: pc, lost: make(chan struct{})}
	connected := make(chan struct{})
	var connectedOnce sync.Once
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		switch s {
		case webrtc.PeerConnectionStateConnected:
			connectedOnce.Do(func() { close(connected) })
		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed:
			l.lostMu.Do(func() { close(l.lost) })
		}
	})
	session, err := l.offer(ctx, board, code)
	if err != nil {
		_ = pc.Close()
		return nil, err
	}
	if err := l.awaitAnswer(ctx, notes, session, code); err != nil {
		_ = pc.Close()
		return nil, err
	}
	select {
	case <-connected:
		return l, nil
	case <-l.lost:
		return nil, errors.New("the TV answered but no direct path worked; this network needs a relay (TURN)")
	case <-ctx.Done():
		_ = pc.Close()
		return nil, fmt.Errorf("connecting to the TV: %w", ctx.Err())
	}
}

func (l *Link) offer(ctx context.Context, board Board, code Code) (string, error) {
	// A data channel must exist before the offer, or it carries no data section.
	if _, err := l.pc.CreateDataChannel("control", nil); err != nil {
		return "", err
	}
	desc, err := l.pc.CreateOffer(nil)
	if err != nil {
		return "", err
	}
	gctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	sdp, err := gatheredSDP(gctx, l.pc, desc)
	if err != nil {
		return "", err
	}
	id := make([]byte, 6)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	session := hex.EncodeToString(id)
	a, err := code.aead()
	if err != nil {
		return "", err
	}
	sealed, err := seal(a, note{Kind: kindOffer, Session: session, SDP: sdp})
	if err != nil {
		return "", err
	}
	return session, board.Post(ctx, code.topic(), sealed)
}

func (l *Link) awaitAnswer(ctx context.Context, notes <-chan string, session string, code Code) error {
	a, err := code.aead()
	if err != nil {
		return err
	}
	for {
		select {
		case content, ok := <-notes:
			if !ok {
				return errors.New("lost the relays before the TV answered")
			}
			n, err := open(a, content, time.Now())
			if err != nil || n.Kind != kindAnswer || n.Session != session {
				continue
			}
			return l.pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: n.SDP})
		case <-ctx.Done():
			return errors.New("the TV did not answer; check tvlink is running and the code matches")
		}
	}
}

// Route says whether traffic goes straight to the TV or via a relay.
func (l *Link) Route() string { return route(l.pc) }

// Close ends the connection.
func (l *Link) Close() error { return l.pc.Close() }

// Forward carries every connection accepted on ln to the TV's adb until ctx ends
// or the link drops.
func (l *Link) Forward(ctx context.Context, ln net.Listener) error {
	go func() {
		select {
		case <-ctx.Done():
		case <-l.lost:
		}
		_ = ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-l.lost:
				return errors.New("connection to the TV was lost")
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}
		l.carry(conn)
	}
}

func (l *Link) carry(conn net.Conn) {
	dc, err := l.pc.CreateDataChannel(streamLabel, nil)
	if err != nil {
		_ = conn.Close()
		return
	}
	dc.OnOpen(func() {
		raw, err := dc.Detach()
		if err != nil {
			_ = conn.Close()
			return
		}
		go pipe(dc, raw, conn)
	})
}
