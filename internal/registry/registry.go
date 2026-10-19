// Package registry stores the operator's list of customer TVs.
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
)

// Device is one customer TV. Address is host:port reachable over the tailnet or LAN.
type Device struct {
	Name    string   `json:"name"`
	Address string   `json:"address"`
	Groups  []string `json:"groups,omitempty"`
	Model   string   `json:"model,omitempty"`
}

// Registry is the on-disk device list.
type Registry struct {
	path    string
	Devices []Device `json:"devices"`
}

var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

// DefaultPath is ~/.config/tvctl/devices.json, overridable with TVCTL_REGISTRY.
func DefaultPath() (string, error) {
	if p := os.Getenv("TVCTL_REGISTRY"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tvctl", "devices.json"), nil
}

// Load reads the registry; a missing file is an empty registry.
func Load(path string) (*Registry, error) {
	r := &Registry{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return r, nil
}

// Save writes atomically so a crash never leaves a half-written registry.
func (r *Registry) Save() error {
	sort.Slice(r.Devices, func(i, j int) bool { return r.Devices[i].Name < r.Devices[j].Name })
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o700); err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

// Upsert adds a device or replaces the one with the same name.
func (r *Registry) Upsert(d Device) error {
	if !namePattern.MatchString(d.Name) {
		return fmt.Errorf("invalid name %q: use lowercase letters, digits and dashes", d.Name)
	}
	for _, g := range d.Groups {
		if !namePattern.MatchString(g) {
			return fmt.Errorf("invalid group %q", g)
		}
	}
	for i := range r.Devices {
		if r.Devices[i].Name == d.Name {
			r.Devices[i] = d
			return nil
		}
	}
	r.Devices = append(r.Devices, d)
	return nil
}

func (r *Registry) Remove(name string) error {
	for i := range r.Devices {
		if r.Devices[i].Name == name {
			r.Devices = slices.Delete(r.Devices, i, i+1)
			return nil
		}
	}
	return fmt.Errorf("no device named %q", name)
}

func (r *Registry) Get(name string) (Device, error) {
	for _, d := range r.Devices {
		if d.Name == name {
			return d, nil
		}
	}
	return Device{}, fmt.Errorf("no device named %q (see `tvctl devices`)", name)
}

func (r *Registry) InGroup(group string) []Device {
	var out []Device
	for _, d := range r.Devices {
		if slices.Contains(d.Groups, group) {
			out = append(out, d)
		}
	}
	return out
}
