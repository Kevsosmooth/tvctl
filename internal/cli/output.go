package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"tvremote/internal/fanout"
	"tvremote/internal/uitree"
)

func (a *app) printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printResults prints one block per device: JSON for agents, short text for people.
func (a *app) printResults(w io.Writer, results []fanout.Result) error {
	if a.jsonOut {
		return a.printJSON(w, results)
	}
	for _, r := range results {
		if !r.OK {
			fmt.Fprintf(w, "%s: FAILED: %s\n", r.Device, r.Error)
			continue
		}
		text := formatValue(r.Value)
		if text == "" {
			fmt.Fprintf(w, "%s: ok (%s)\n", r.Device, r.Duration)
			continue
		}
		if len(results) == 1 {
			fmt.Fprintln(w, text)
			continue
		}
		fmt.Fprintf(w, "== %s\n%s\n", r.Device, text)
	}
	return nil
}

func formatValue(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case []string:
		return strings.Join(val, "\n")
	case []uitree.Element:
		return formatElements(val)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(data)
}

// formatElements prints one line per element: index, markers, label, center point.
func formatElements(els []uitree.Element) string {
	var b strings.Builder
	for _, e := range els {
		x, y := e.Bounds.Center()
		fmt.Fprintf(&b, "%3d %s %-40q %s @%d,%d\n", e.Index, markers(e), label(e), e.Class, x, y)
	}
	return strings.TrimRight(b.String(), "\n")
}

func markers(e uitree.Element) string {
	m := []byte("---")
	if e.Focused {
		m[0] = 'F'
	}
	if e.Clickable {
		m[1] = 'C'
	}
	if e.Checked != nil {
		m[2] = 'o'
		if *e.Checked {
			m[2] = 'x'
		}
	}
	return string(m)
}

func label(e uitree.Element) string {
	switch {
	case e.Text != "":
		return e.Text
	case e.Desc != "":
		return e.Desc
	}
	return e.ID
}
