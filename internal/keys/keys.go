// Package keys maps friendly remote-control names to Android keycodes.
package keys

import (
	"fmt"
	"sort"
	"strings"
)

var codes = map[string]string{
	"home":      "KEYCODE_HOME",
	"back":      "KEYCODE_BACK",
	"up":        "KEYCODE_DPAD_UP",
	"down":      "KEYCODE_DPAD_DOWN",
	"left":      "KEYCODE_DPAD_LEFT",
	"right":     "KEYCODE_DPAD_RIGHT",
	"ok":        "KEYCODE_DPAD_CENTER",
	"enter":     "KEYCODE_ENTER",
	"menu":      "KEYCODE_MENU",
	"settings":  "KEYCODE_SETTINGS",
	"search":    "KEYCODE_SEARCH",
	"play":      "KEYCODE_MEDIA_PLAY_PAUSE",
	"rewind":    "KEYCODE_MEDIA_REWIND",
	"forward":   "KEYCODE_MEDIA_FAST_FORWARD",
	"volup":     "KEYCODE_VOLUME_UP",
	"voldown":   "KEYCODE_VOLUME_DOWN",
	"mute":      "KEYCODE_VOLUME_MUTE",
	"power":     "KEYCODE_POWER",
	"wake":      "KEYCODE_WAKEUP",
	"sleep":     "KEYCODE_SLEEP",
	"delete":    "KEYCODE_DEL",
	"recents":   "KEYCODE_APP_SWITCH",
	"tab":       "KEYCODE_TAB",
	"pageup":    "KEYCODE_PAGE_UP",
	"pagedown":  "KEYCODE_PAGE_DOWN",
	"guide":     "KEYCODE_GUIDE",
	"channelup": "KEYCODE_CHANNEL_UP",
}

// Resolve accepts a friendly name ("ok"), a KEYCODE_ name, or a numeric keycode.
func Resolve(name string) (string, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	if code, ok := codes[n]; ok {
		return code, nil
	}
	upper := strings.ToUpper(name)
	if strings.HasPrefix(upper, "KEYCODE_") && isKeycodeName(upper) {
		return upper, nil
	}
	if isDigits(n) {
		return n, nil
	}
	return "", fmt.Errorf("unknown key %q (known: %s)", name, strings.Join(Names(), ", "))
}

// Names lists the friendly key names.
func Names() []string {
	out := make([]string, 0, len(codes))
	for k := range codes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func isKeycodeName(s string) bool {
	for _, r := range s {
		if !(r == '_' || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func isDigits(s string) bool {
	if s == "" || len(s) > 4 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
