// Command tvctl controls Android TV, Google TV and Fire TV devices over adb.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"tvremote/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := cli.NewRoot().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
