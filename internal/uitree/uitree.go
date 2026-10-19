// Package uitree turns a uiautomator XML dump into a compact element list an agent can read.
package uitree

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Element is one on-screen node worth showing: it has text, a description, or can be acted on.
type Element struct {
	Index     int    `json:"i"`
	Text      string `json:"text,omitempty"`
	Desc      string `json:"desc,omitempty"`
	ID        string `json:"id,omitempty"`
	Class     string `json:"class,omitempty"`
	Package   string `json:"pkg,omitempty"`
	Bounds    Rect   `json:"bounds"`
	Focused   bool   `json:"focused,omitempty"`
	Clickable bool   `json:"clickable,omitempty"`
	Focusable bool   `json:"focusable,omitempty"`
	Selected  bool   `json:"selected,omitempty"`
	Checked   *bool  `json:"checked,omitempty"`
}

type Rect struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

// Center is where a tap lands.
func (r Rect) Center() (int, int) {
	return (r.X1 + r.X2) / 2, (r.Y1 + r.Y2) / 2
}

type xmlNode struct {
	Text      string    `xml:"text,attr"`
	ID        string    `xml:"resource-id,attr"`
	Class     string    `xml:"class,attr"`
	Package   string    `xml:"package,attr"`
	Desc      string    `xml:"content-desc,attr"`
	Checkable string    `xml:"checkable,attr"`
	Checked   string    `xml:"checked,attr"`
	Clickable string    `xml:"clickable,attr"`
	Focusable string    `xml:"focusable,attr"`
	Focused   string    `xml:"focused,attr"`
	Selected  string    `xml:"selected,attr"`
	Bounds    string    `xml:"bounds,attr"`
	Children  []xmlNode `xml:"node"`
}

var boundsPattern = regexp.MustCompile(`^\[(-?\d+),(-?\d+)\]\[(-?\d+),(-?\d+)\]$`)

// Parse reads a uiautomator dump and keeps only meaningful elements, in screen order.
func Parse(data []byte) ([]Element, error) {
	var root struct {
		Nodes []xmlNode `xml:"node"`
	}
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse ui dump: %w", err)
	}
	var out []Element
	for _, n := range root.Nodes {
		collect(n, &out, false)
	}
	return out, nil
}

// collect walks the tree. A clickable or focusable row with no label of its own takes
// the text of its children, and those plain text children are then skipped, so each
// row shows up once with a readable label.
func collect(n xmlNode, out *[]Element, absorbed bool) {
	e, keep := toElement(n)
	interactive := e.Clickable || e.Focusable
	absorbing := false
	if interactive && e.Text == "" && e.Desc == "" {
		if label := descendantText(n); label != "" {
			e.Text, keep, absorbing = label, true, true
		}
	}
	if absorbed && !interactive && e.Checked == nil {
		keep = false
	}
	if keep {
		e.Index = len(*out)
		*out = append(*out, e)
	}
	for _, c := range n.Children {
		collect(c, out, absorbed || absorbing)
	}
}

const maxLabel = 120

func descendantText(n xmlNode) string {
	var parts []string
	var walk func(xmlNode)
	walk = func(x xmlNode) {
		for _, c := range x.Children {
			switch {
			case c.Text != "":
				parts = append(parts, c.Text)
			case c.Desc != "":
				parts = append(parts, c.Desc)
			}
			walk(c)
		}
	}
	walk(n)
	label := []rune(strings.Join(parts, " · "))
	if len(label) > maxLabel {
		label = label[:maxLabel]
	}
	return string(label)
}

func toElement(n xmlNode) (Element, bool) {
	e := Element{
		Text:      n.Text,
		Desc:      n.Desc,
		ID:        n.ID,
		Class:     shortClass(n.Class),
		Package:   n.Package,
		Bounds:    parseBounds(n.Bounds),
		Focused:   n.Focused == "true",
		Clickable: n.Clickable == "true",
		Focusable: n.Focusable == "true",
		Selected:  n.Selected == "true",
	}
	if n.Checkable == "true" {
		checked := n.Checked == "true"
		e.Checked = &checked
	}
	keep := e.Text != "" || e.Desc != "" || e.Focused || e.Clickable || e.Checked != nil
	return e, keep
}

func parseBounds(s string) Rect {
	m := boundsPattern.FindStringSubmatch(s)
	if m == nil {
		return Rect{}
	}
	v := make([]int, 4)
	for i := range v {
		v[i], _ = strconv.Atoi(m[i+1])
	}
	return Rect{X1: v[0], Y1: v[1], X2: v[2], Y2: v[3]}
}

// shortClass drops the package prefix: android.widget.TextView -> TextView.
func shortClass(c string) string {
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == '.' {
			return c[i+1:]
		}
	}
	return c
}
