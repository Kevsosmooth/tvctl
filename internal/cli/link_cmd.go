package cli

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"

	"tvremote/internal/p2p"
)

func (a *app) linkCmd() *cobra.Command {
	var codeFile, listen string
	var wait time.Duration
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Connect directly to a TV running tvlink, from any network",
		Long: "Finds the TV through public Nostr relays, connects straight to it and serves its adb\n" +
			"on --listen until stopped. In another terminal: tvctl add <name> <listen address>.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireLoopback(listen); err != nil {
				return err
			}
			code, err := p2p.ReadCode(codeFile)
			if err != nil {
				return fmt.Errorf("pairing code: %w", err)
			}
			ln, err := net.Listen("tcp", listen)
			if err != nil {
				return err
			}
			defer ln.Close()
			ctx := cmd.Context()
			board, err := p2p.NewNostr(ctx, p2p.DefaultRelays)
			if err != nil {
				return err
			}
			start := time.Now()
			dialCtx, cancel := context.WithTimeout(ctx, wait)
			link, err := p2p.Dial(dialCtx, board, code, p2p.DefaultOptions())
			cancel()
			if err != nil {
				return err
			}
			defer link.Close()
			fmt.Fprintf(cmd.OutOrStdout(), "linked in %s, %s; TV adb at %s\n",
				time.Since(start).Round(100*time.Millisecond), link.Route(), listen)
			return link.Forward(ctx, ln)
		},
	}
	f := cmd.Flags()
	f.StringVar(&codeFile, "code-file", "", "file holding the TV's pairing code")
	f.StringVar(&listen, "listen", "127.0.0.1:15555", "local address to serve the TV's adb on")
	f.DurationVar(&wait, "wait", 45*time.Second, "how long to wait for the TV")
	_ = cmd.MarkFlagRequired("code-file")
	return cmd
}

// requireLoopback keeps the TV's adb off the network: anyone who can reach the
// listen address controls the TV.
func requireLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("listen address %q: %w", addr, err)
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("listen address must be on this computer only (127.0.0.1), got %q", host)
	}
	return nil
}
