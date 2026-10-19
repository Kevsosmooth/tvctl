package p2p

import (
	"context"
	"crypto/cipher"
	"net"
	"time"

	"github.com/pion/webrtc/v4"
)

// streamLabel marks data channels that carry an adb connection.
const streamLabel = "adb"

// Logf receives progress lines.
type Logf func(format string, args ...any)

// Serve runs on the TV. It answers helpers holding code and forwards each of
// their streams to target (the TV's own adb), until ctx ends.
func Serve(ctx context.Context, board Board, code Code, target string, opt Options, logf Logf) error {
	a, err := code.aead()
	if err != nil {
		return err
	}
	notes, err := board.Watch(ctx, code.topic())
	if err != nil {
		return err
	}
	answered := map[string]bool{}
	for content := range notes {
		n, err := open(a, content, time.Now())
		if err != nil || n.Kind != kindOffer || answered[n.Session] {
			continue
		}
		answered[n.Session] = true
		logf("helper %s asking to connect", n.Session)
		if err := answer(ctx, board, a, code, n, target, opt, logf); err != nil {
			logf("helper %s: %v", n.Session, err)
		}
	}
	return ctx.Err()
}

func answer(ctx context.Context, board Board, a cipher.AEAD, code Code, offer note, target string, opt Options, logf Logf) error {
	pc, err := newPeer(opt)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		logf("helper %s: %s", offer.Session, s)
		if s == webrtc.PeerConnectionStateFailed {
			_ = pc.Close()
		}
	})
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		if dc.Label() != streamLabel {
			return
		}
		dc.OnOpen(func() { forwardTo(dc, target, logf) })
	})
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offer.SDP}); err != nil {
		_ = pc.Close()
		return err
	}
	desc, err := pc.CreateAnswer(nil)
	if err != nil {
		_ = pc.Close()
		return err
	}
	gctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	sdp, err := gatheredSDP(gctx, pc, desc)
	if err != nil {
		_ = pc.Close()
		return err
	}
	sealed, err := seal(a, note{Kind: kindAnswer, Session: offer.Session, SDP: sdp})
	if err != nil {
		_ = pc.Close()
		return err
	}
	return board.Post(ctx, code.topic(), sealed)
}

func forwardTo(dc *webrtc.DataChannel, target string, logf Logf) {
	raw, err := dc.Detach()
	if err != nil {
		logf("stream: %v", err)
		return
	}
	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		logf("stream: %v", err)
		_ = raw.Close()
		return
	}
	go pipe(dc, raw, conn)
}
