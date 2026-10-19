package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tvremote/internal/registry"
	"tvremote/internal/script"
)

func (a *app) shellCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "shell -- <command...>",
		Short: "Run a shell command on the TV and print its output",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			line := strings.Join(args, " ")
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return d.Shell(ctx, line)
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) rebootCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "reboot",
		Short: "Restart TVs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return nil, d.Reboot(ctx)
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) runCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "run <script.yaml>",
		Short: "Run a setup script on TVs in parallel (see docs/scripts.md)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := script.Load(args[0])
			if err != nil {
				return err
			}
			dir := filepath.Dir(args[0])
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				done, err := s.Run(ctx, d, script.Env{Device: d.Name, Dir: dir})
				return fmt.Sprintf("%d/%d steps", done, len(s.Steps)), err
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) viewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <name>",
		Short: "Open a live view with mouse and keyboard control (needs scrcpy)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := a.registry()
			if err != nil {
				return err
			}
			dev, err := reg.Get(args[0])
			if err != nil {
				return err
			}
			return runScrcpy(cmd, a.adbPath, dev)
		},
	}
}

func runScrcpy(cmd *cobra.Command, adbPath string, dev registry.Device) error {
	path, err := exec.LookPath("scrcpy")
	if err != nil {
		return fmt.Errorf("scrcpy is not installed: https://github.com/Genymobile/scrcpy")
	}
	c := exec.CommandContext(cmd.Context(), path, "--serial", dev.Address, "--window-title", dev.Name, "--max-size", "1280")
	c.Env = append(os.Environ(), "ADB="+adbPath)
	c.Stdout, c.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
	return c.Run()
}
