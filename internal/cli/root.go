// Package cli wires tvctl's commands to the tv, registry and fanout packages.
package cli

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"tvremote/internal/adb"
	"tvremote/internal/registry"
)

// app holds settings shared by every command.
type app struct {
	jsonOut  bool
	adbPath  string
	regPath  string
	parallel int
	timeout  time.Duration
}

// NewRoot builds the tvctl command tree.
func NewRoot() *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:           "tvctl",
		Short:         "Control Android TV, Google TV and Fire TV devices over adb",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	f := root.PersistentFlags()
	f.BoolVar(&a.jsonOut, "json", false, "print JSON (for scripts and AI agents)")
	f.StringVar(&a.adbPath, "adb", envOr("TVCTL_ADB", "adb"), "adb binary path (env TVCTL_ADB)")
	f.StringVar(&a.regPath, "registry", "", "device registry file (env TVCTL_REGISTRY)")
	f.IntVar(&a.parallel, "parallel", 8, "max devices worked on at once")
	f.DurationVar(&a.timeout, "timeout", 2*time.Minute, "per-device time limit")

	root.AddCommand(
		a.devicesCmd(), a.addCmd(), a.removeCmd(), a.connectCmd(), a.linkCmd(),
		a.infoCmd(), a.uiCmd(), a.screenshotCmd(),
		a.keyCmd(), a.textCmd(), a.tapCmd(), a.clickCmd(), a.waitForCmd(),
		a.appsCmd(), a.installCmd(), a.uninstallCmd(), a.launchCmd(), a.stopCmd(),
		a.shellCmd(), a.rebootCmd(), a.runCmd(), a.viewCmd(), a.keysCmd(),
	)
	return root
}

func (a *app) client() *adb.Client {
	return adb.NewClient(adb.ExecRunner{Path: a.adbPath})
}

func (a *app) registry() (*registry.Registry, error) {
	path := a.regPath
	if path == "" {
		p, err := registry.DefaultPath()
		if err != nil {
			return nil, err
		}
		path = p
	}
	return registry.Load(path)
}

func (a *app) deviceContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, a.timeout)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
