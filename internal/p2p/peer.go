package p2p

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/pion/webrtc/v4"
)

// Options sets how the two ends reach each other.
type Options struct {
	// ICE lists STUN servers (find our public address) and TURN servers (relay
	// for networks that block direct connections).
	ICE []webrtc.ICEServer

	loopbackOnly bool // tests: both ends on this machine
}

// virtualPrefixes name container and VM bridges. A Docker host can have dozens,
// and offering each one as a route makes connecting slow or time out.
var virtualPrefixes = []string{"docker", "br-", "veth", "virbr", "vmnet", "vboxnet", "cni", "flannel"}

func physicalInterface(name string) bool {
	for _, p := range virtualPrefixes {
		if strings.HasPrefix(name, p) {
			return false
		}
	}
	return true
}

// DefaultOptions uses free public STUN servers and no relay.
func DefaultOptions() Options {
	return Options{ICE: []webrtc.ICEServer{{URLs: []string{
		"stun:stun.l.google.com:19302",
		"stun:stun.cloudflare.com:3478",
	}}}}
}

func newPeer(opt Options) (*webrtc.PeerConnection, error) {
	var se webrtc.SettingEngine
	se.DetachDataChannels()
	se.SetInterfaceFilter(physicalInterface)
	if opt.loopbackOnly {
		se.SetIncludeLoopbackCandidate(true)
		se.SetIPFilter(func(ip net.IP) bool { return ip.IsLoopback() })
	}
	api := webrtc.NewAPI(webrtc.WithSettingEngine(se))
	return api.NewPeerConnection(webrtc.Configuration{ICEServers: opt.ICE})
}

// gatheredSDP sets desc as the local description and returns it with every
// candidate included, because the notes are swapped only once.
func gatheredSDP(ctx context.Context, pc *webrtc.PeerConnection, desc webrtc.SessionDescription) (string, error) {
	done := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(desc); err != nil {
		return "", err
	}
	select {
	case <-done:
	case <-ctx.Done():
		return "", fmt.Errorf("finding network addresses: %w", ctx.Err())
	}
	return pc.LocalDescription().SDP, nil
}

// route says whether traffic goes straight between the two ends or via a relay.
func route(pc *webrtc.PeerConnection) string {
	sctp := pc.SCTP()
	if sctp == nil {
		return "unknown"
	}
	pair, err := sctp.Transport().ICETransport().GetSelectedCandidatePair()
	if err != nil || pair == nil {
		return "unknown"
	}
	if pair.Local.Typ == webrtc.ICECandidateTypeRelay || pair.Remote.Typ == webrtc.ICECandidateTypeRelay {
		return "relay"
	}
	return fmt.Sprintf("direct (%s to %s)", pair.Local.Typ, pair.Remote.Typ)
}
