package registry

import (
	"errors"
	"fmt"
)

// Selector picks which devices a command runs on.
type Selector struct {
	Names []string
	Group string
	All   bool
}

// Select resolves a selector. Exactly one of names, group or all must be given, so a
// typo never silently widens a command to every TV.
func (r *Registry) Select(s Selector) ([]Device, error) {
	set := 0
	if len(s.Names) > 0 {
		set++
	}
	if s.Group != "" {
		set++
	}
	if s.All {
		set++
	}
	if set != 1 {
		return nil, errors.New("choose devices with names, --group, or --all (exactly one)")
	}
	switch {
	case s.All:
		if len(r.Devices) == 0 {
			return nil, errors.New("no devices registered (see `tvctl add`)")
		}
		return r.Devices, nil
	case s.Group != "":
		devices := r.InGroup(s.Group)
		if len(devices) == 0 {
			return nil, fmt.Errorf("group %q has no devices", s.Group)
		}
		return devices, nil
	}
	out := make([]Device, 0, len(s.Names))
	for _, n := range s.Names {
		d, err := r.Get(n)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}
