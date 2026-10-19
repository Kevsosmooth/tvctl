package uitree

import (
	"fmt"
	"strings"
)

// Query matches elements by visible text, content description or resource id.
type Query struct {
	Text  string // case-insensitive; exact match wins over substring
	ID    string // exact resource-id or its suffix after ":id/"
	Exact bool
}

// Find returns the best match: an exact text match first, then the first substring match.
func Find(elements []Element, q Query) (Element, error) {
	if q.Text == "" && q.ID == "" {
		return Element{}, fmt.Errorf("query needs text or id")
	}
	var partial *Element
	for i := range elements {
		e := elements[i]
		if q.ID != "" && !idMatches(e.ID, q.ID) {
			continue
		}
		if q.Text == "" {
			return e, nil
		}
		if equalFold(e.Text, q.Text) || equalFold(e.Desc, q.Text) {
			return e, nil
		}
		if !q.Exact && partial == nil && (containsFold(e.Text, q.Text) || containsFold(e.Desc, q.Text)) {
			partial = &elements[i]
		}
	}
	if partial != nil {
		return *partial, nil
	}
	return Element{}, fmt.Errorf("no element matching %s on screen", q.describe())
}

func (q Query) describe() string {
	switch {
	case q.Text != "" && q.ID != "":
		return fmt.Sprintf("text %q with id %q", q.Text, q.ID)
	case q.ID != "":
		return fmt.Sprintf("id %q", q.ID)
	}
	return fmt.Sprintf("text %q", q.Text)
}

func idMatches(id, want string) bool {
	return id == want || strings.HasSuffix(id, ":id/"+want)
}

func equalFold(a, b string) bool {
	return a != "" && strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func containsFold(s, sub string) bool {
	return s != "" && strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
