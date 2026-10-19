package tv

import (
	"context"
	"strings"
)

// Info is a quick summary of a TV.
type Info struct {
	Serial       string `json:"serial"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Android      string `json:"android"`
	SDK          string `json:"sdk"`
	FireOS       string `json:"fire_os,omitempty"`
	Focused      string `json:"focused,omitempty"`
	ScreenOn     bool   `json:"screen_on"`
}

var infoProps = []string{
	"ro.product.manufacturer",
	"ro.product.model",
	"ro.build.version.release",
	"ro.build.version.sdk",
	"ro.build.version.name",
}

// Info reads device properties in one shell round trip.
func (d *Device) Info(ctx context.Context) (Info, error) {
	cmd := make([]string, 0, len(infoProps))
	for _, p := range infoProps {
		cmd = append(cmd, "getprop "+p)
	}
	out, err := d.shell(ctx, strings.Join(cmd, ";"))
	if err != nil {
		return Info{}, err
	}
	v := strings.Split(strings.ReplaceAll(out, "\r", ""), "\n")
	for len(v) < len(infoProps) {
		v = append(v, "")
	}
	info := Info{Serial: d.serial, Manufacturer: v[0], Model: v[1], Android: v[2], SDK: v[3]}
	if info.Manufacturer == "Amazon" {
		info.FireOS = v[4]
	}
	if info.Focused, err = d.FocusedWindow(ctx); err != nil {
		return info, err
	}
	info.ScreenOn, err = d.ScreenOn(ctx)
	return info, err
}

// ScreenOn reports whether the display is awake.
func (d *Device) ScreenOn(ctx context.Context) (bool, error) {
	out, err := d.shell(ctx, "dumpsys", "power")
	if err != nil {
		return false, err
	}
	return parseScreenOn(out), nil
}

func parseScreenOn(dumpsys string) bool {
	return strings.Contains(dumpsys, "mWakefulness=Awake") ||
		strings.Contains(dumpsys, "Display Power: state=ON")
}
