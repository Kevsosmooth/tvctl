package tv

import "strings"

// shellQuote wraps s in single quotes for the device's /system/bin/sh, because adb shell
// joins its arguments into one command line that the device shell re-parses.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// inputTextArg encodes text for `input text`, which treats %s as a space and splits on
// real spaces.
func inputTextArg(s string) string {
	return shellQuote(strings.ReplaceAll(s, " ", "%s"))
}
