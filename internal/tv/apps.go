package tv

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var packagePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z0-9_]+)+$`)

func validPackage(pkg string) error {
	if !packagePattern.MatchString(pkg) {
		return fmt.Errorf("invalid package name %q", pkg)
	}
	return nil
}

// Apps lists installed packages; thirdParty limits it to user-installed apps.
func (d *Device) Apps(ctx context.Context, thirdParty bool) ([]string, error) {
	args := []string{"pm", "list", "packages"}
	if thirdParty {
		args = append(args, "-3")
	}
	out, err := d.shell(ctx, args...)
	if err != nil {
		return nil, err
	}
	return parsePackages(out), nil
}

func parsePackages(out string) []string {
	var pkgs []string
	for _, line := range strings.Split(out, "\n") {
		if p, ok := strings.CutPrefix(strings.TrimSpace(line), "package:"); ok && p != "" {
			pkgs = append(pkgs, p)
		}
	}
	sort.Strings(pkgs)
	return pkgs
}

// Launch opens an app's TV launcher entry, falling back to its phone launcher entry
// for apps that only declare one.
func (d *Device) Launch(ctx context.Context, pkg string) error {
	if err := validPackage(pkg); err != nil {
		return err
	}
	for _, category := range []string{"android.intent.category.LEANBACK_LAUNCHER", "android.intent.category.LAUNCHER"} {
		out, err := d.shell(ctx, "cmd", "package", "resolve-activity", "--brief", "-c", category, pkg)
		if err != nil {
			return err
		}
		component := lastLine(out)
		if !componentPattern.MatchString(component) {
			continue
		}
		// A screensaver would hide the app while am start still reports success.
		if _, err := d.shell(ctx, "input", "keyevent", "KEYCODE_WAKEUP"); err != nil {
			return err
		}
		out, err = d.shell(ctx, "am", "start", "-n", component)
		if err != nil {
			return err
		}
		if strings.Contains(out, "Error") {
			return fmt.Errorf("launch %s: %s", pkg, lastLine(out))
		}
		return nil
	}
	return fmt.Errorf("%s has no launchable activity", pkg)
}

var componentPattern = regexp.MustCompile(`^[a-zA-Z][\w.]*/[\w.$]+$`)

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(strings.ReplaceAll(s, "\r", "")), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

func (d *Device) Uninstall(ctx context.Context, pkg string) error {
	if err := validPackage(pkg); err != nil {
		return err
	}
	out, err := d.shell(ctx, "pm", "uninstall", pkg)
	if err != nil {
		return err
	}
	if !strings.Contains(out, "Success") {
		return fmt.Errorf("uninstall %s: %s", pkg, out)
	}
	return nil
}

func (d *Device) Stop(ctx context.Context, pkg string) error {
	if err := validPackage(pkg); err != nil {
		return err
	}
	_, err := d.shell(ctx, "am", "force-stop", pkg)
	return err
}
