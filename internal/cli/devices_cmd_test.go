package cli

import "testing"

func TestNormalizeAddress(t *testing.T) {
	good := map[string]string{
		"192.168.50.83":       "192.168.50.83:5555",
		"192.168.50.83:5556":  "192.168.50.83:5556",
		"shield.tail1234.net": "shield.tail1234.net:5555",
		"[fd7a:115c::1]:5555": "[fd7a:115c::1]:5555",
		"fd7a:115c::1":        "[fd7a:115c::1]:5555",
	}
	for in, want := range good {
		got, err := normalizeAddress(in)
		if err != nil || got != want {
			t.Errorf("normalizeAddress(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", ":5555", "-host", "tv;reboot", "tv/x", "tv:abc", "tv:0", "tv:70000", "tv name"} {
		if got, err := normalizeAddress(in); err == nil {
			t.Errorf("normalizeAddress(%q) = %q, want error", in, got)
		}
	}
}
