package tv

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"tvremote/internal/uitree"
)

// UI dumps the current screen as a compact element list.
func (d *Device) UI(ctx context.Context) ([]uitree.Element, error) {
	raw, err := d.client.ExecOut(ctx, d.serial, "uiautomator", "dump", "/dev/tty")
	if err != nil {
		return nil, err
	}
	xmlDoc, err := extractHierarchy(raw)
	if err != nil {
		return nil, err
	}
	return uitree.Parse(xmlDoc)
}

// extractHierarchy strips the "UI hierchary dumped to" trailer uiautomator prints.
func extractHierarchy(raw []byte) ([]byte, error) {
	start := bytes.Index(raw, []byte("<?xml"))
	end := bytes.LastIndex(raw, []byte("</hierarchy>"))
	if start < 0 || end < 0 {
		msg := bytes.TrimSpace(firstLine(raw))
		if bytes.Contains(raw, []byte("null root node")) {
			return nil, fmt.Errorf("ui dump failed: nothing on screen (TV asleep? try `key wake`): %s", msg)
		}
		return nil, fmt.Errorf("ui dump failed: %s", msg)
	}
	return raw[start : end+len("</hierarchy>")], nil
}

func firstLine(b []byte) []byte {
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		return b[:i]
	}
	return b
}

var focusPattern = regexp.MustCompile(`mCurrentFocus=Window\{\S+ \S+ ([^\s}]+)\}`)

// FocusedWindow returns the package/activity in front, or "" if none is focused.
func (d *Device) FocusedWindow(ctx context.Context) (string, error) {
	out, err := d.shell(ctx, "dumpsys", "window")
	if err != nil {
		return "", err
	}
	return parseFocus(out), nil
}

func parseFocus(dumpsys string) string {
	m := focusPattern.FindStringSubmatch(dumpsys)
	if m == nil {
		return ""
	}
	return m[1]
}
