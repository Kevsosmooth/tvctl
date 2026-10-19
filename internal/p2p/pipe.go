package p2p

import (
	"io"
	"net"
	"time"

	"github.com/pion/webrtc/v4"
)

const (
	chunkSize = 16 << 10 // largest message every WebRTC stack accepts
	readSize  = 64 << 10
	// An APK upload reads from disk far faster than the link sends, so stop
	// reading once this much is queued instead of buffering the whole file.
	highWater = 1 << 20
	lowWater  = 256 << 10
)

// pipe copies both ways between a TCP connection and a data channel until either
// side closes, then closes both.
func pipe(dc *webrtc.DataChannel, raw io.ReadWriteCloser, conn net.Conn) {
	drained := make(chan struct{}, 1)
	dc.SetBufferedAmountLowThreshold(lowWater)
	dc.OnBufferedAmountLow(func() {
		select {
		case drained <- struct{}{}:
		default:
		}
	})
	done := make(chan struct{}, 2)
	go func() {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, chunkSize)
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				if !waitForRoom(dc, drained) {
					return
				}
				if _, werr := raw.Write(buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	go func() {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, readSize)
		for {
			n, err := raw.Read(buf)
			if n > 0 {
				if _, werr := conn.Write(buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	<-done
	_ = conn.Close()
	_ = raw.Close()
}

// waitForRoom blocks while the send queue is full. False means the channel closed.
func waitForRoom(dc *webrtc.DataChannel, drained <-chan struct{}) bool {
	for dc.BufferedAmount() > highWater {
		if dc.ReadyState() != webrtc.DataChannelStateOpen {
			return false
		}
		select {
		case <-drained:
		case <-time.After(time.Second):
		}
	}
	return true
}
