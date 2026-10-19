// Command tvlink runs on the TV. It lets a helper holding the TV's pairing code
// reach the TV's own adb over a direct connection. It stands in for the TV app
// until that is built.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tvremote/internal/p2p"
)

func main() {
	codeFile := flag.String("code-file", "/data/local/tmp/tvlink.code", "pairing code file, created if missing")
	target := flag.String("target", "127.0.0.1:5555", "the TV's adb address")
	flag.Parse()
	p2p.UseResolverFallback()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	code, err := p2p.LoadOrCreateCode(*codeFile)
	if err != nil {
		log.Fatalf("pairing code: %v", err)
	}
	board, err := p2p.NewNostr(ctx, p2p.DefaultRelays)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("tvlink ready, pairing code in %s", *codeFile)
	if err := p2p.Serve(ctx, board, code, *target, p2p.DefaultOptions(), log.Printf); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
