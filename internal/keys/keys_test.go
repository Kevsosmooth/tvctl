package keys

import "testing"

func TestResolve(t *testing.T) {
	cases := map[string]string{
		"ok":                 "KEYCODE_DPAD_CENTER",
		"HOME":               "KEYCODE_HOME",
		"keycode_dpad_up":    "KEYCODE_DPAD_UP",
		"KEYCODE_MEDIA_STOP": "KEYCODE_MEDIA_STOP",
		"66":                 "66",
	}
	for in, want := range cases {
		got, err := Resolve(in)
		if err != nil || got != want {
			t.Errorf("Resolve(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

func TestResolveRejectsShellInjection(t *testing.T) {
	for _, in := range []string{"", "nope", "KEYCODE_HOME;reboot", "1; rm -rf /", "123456"} {
		if _, err := Resolve(in); err == nil {
			t.Errorf("Resolve(%q) should fail", in)
		}
	}
}
