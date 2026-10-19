package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func (a *app) appsCmd() *cobra.Command {
	var t targetFlags
	var system bool
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "List installed apps (user-installed only unless --system)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return d.Apps(ctx, !system)
			})
		},
	}
	t.bind(cmd)
	cmd.Flags().BoolVar(&system, "system", false, "include system apps")
	return cmd
}

func (a *app) installCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "install <file.apk>",
		Short: "Install or update an APK, keeping app data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apk := args[0]
			if err := checkAPK(apk); err != nil {
				return err
			}
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return nil, d.Install(ctx, apk)
			})
		},
	}
	t.bind(cmd)
	return cmd
}

// checkAPK fails fast on a bad path before contacting any TV.
func checkAPK(path string) error {
	if !strings.HasSuffix(strings.ToLower(path), ".apk") {
		return fmt.Errorf("%s is not an .apk file", path)
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	return nil
}

// packageCmd builds the simple "<verb> <package>" commands.
func (a *app) packageCmd(use, short string, run func(ctx context.Context, d target, pkg string) error) *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   use + " <package>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return nil, run(ctx, d, args[0])
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) uninstallCmd() *cobra.Command {
	return a.packageCmd("uninstall", "Remove an app", func(ctx context.Context, d target, pkg string) error {
		return d.Uninstall(ctx, pkg)
	})
}

func (a *app) launchCmd() *cobra.Command {
	return a.packageCmd("launch", "Open an app", func(ctx context.Context, d target, pkg string) error {
		return d.Launch(ctx, pkg)
	})
}

func (a *app) stopCmd() *cobra.Command {
	return a.packageCmd("stop", "Force-stop an app", func(ctx context.Context, d target, pkg string) error {
		return d.Stop(ctx, pkg)
	})
}
