package p2p

import (
	"context"
	"net"
	"os"
	"time"

	// Android keeps its trusted certificates where Go cannot find them, so bundle
	// Mozilla's list for when the system list comes back empty.
	_ "golang.org/x/crypto/x509roots/fallback"
)

// publicDNS is used only where /etc/resolv.conf is missing, as on Android: there
// Go's built-in resolver would otherwise ask 127.0.0.1, where nothing answers.
var publicDNS = []string{"1.1.1.1:53", "8.8.8.8:53"}

// UseResolverFallback points Go at public DNS when the system gives it none.
func UseResolverFallback() {
	if _, err := os.Stat("/etc/resolv.conf"); err == nil {
		return
	}
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			d := net.Dialer{Timeout: 3 * time.Second}
			var lastErr error
			for _, addr := range publicDNS {
				conn, err := d.DialContext(ctx, network, addr)
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			return nil, lastErr
		},
	}
}
