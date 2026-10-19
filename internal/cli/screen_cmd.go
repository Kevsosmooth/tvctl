package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"tvremote/internal/tv"
)

func (a *app) infoCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show model, Android version, screen state and the app in front",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return d.Info(ctx)
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) uiCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "List what is on screen (F=focused C=clickable x/o=checked) with tap points",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return d.UI(ctx)
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) screenshotCmd() *cobra.Command {
	var t targetFlags
	var out string
	var width int
	cmd := &cobra.Command{
		Use:   "screenshot",
		Short: "Save a PNG of the screen; {device} in -o becomes the device name",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(t.names) != 1 && !strings.Contains(out, "{device}") {
				return fmt.Errorf("with several devices, -o must contain {device}")
			}
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return saveScreenshot(ctx, d.Device, out, d.Name, width)
			})
		},
	}
	t.bind(cmd)
	cmd.Flags().StringVarP(&out, "out", "o", "shots/{device}.png", "output file")
	cmd.Flags().IntVar(&width, "width", tv.ShotWidth, "shrink to this width; 0 keeps full resolution")
	return cmd
}

func saveScreenshot(ctx context.Context, d *tv.Device, pattern, name string, width int) (string, error) {
	path := strings.ReplaceAll(pattern, "{device}", name)
	return path, d.SaveScreenshot(ctx, path, width)
}
