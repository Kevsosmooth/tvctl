package tv

import (
	"context"
	"fmt"
	"strconv"

	"tvremote/internal/keys"
	"tvremote/internal/uitree"
)

// Keys presses remote buttons in order. Names are resolved and validated first so a bad
// name never reaches the device shell.
func (d *Device) Keys(ctx context.Context, names ...string) error {
	codes := make([]string, 0, len(names))
	for _, n := range names {
		code, err := keys.Resolve(n)
		if err != nil {
			return err
		}
		codes = append(codes, code)
	}
	for _, code := range codes {
		if _, err := d.shell(ctx, "input", "keyevent", code); err != nil {
			return err
		}
	}
	return nil
}

// Text types into the focused field.
func (d *Device) Text(ctx context.Context, text string) error {
	if text == "" {
		return fmt.Errorf("text is empty")
	}
	_, err := d.shell(ctx, "input", "text", inputTextArg(text))
	return err
}

func (d *Device) Tap(ctx context.Context, x, y int) error {
	if x < 0 || y < 0 {
		return fmt.Errorf("tap position must be positive, got %d,%d", x, y)
	}
	_, err := d.shell(ctx, "input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	return err
}

// Click finds an element on screen and taps its center, returning the element hit.
func (d *Device) Click(ctx context.Context, q uitree.Query) (uitree.Element, error) {
	elements, err := d.UI(ctx)
	if err != nil {
		return uitree.Element{}, err
	}
	el, err := uitree.Find(elements, q)
	if err != nil {
		return uitree.Element{}, err
	}
	x, y := el.Bounds.Center()
	return el, d.Tap(ctx, x, y)
}
