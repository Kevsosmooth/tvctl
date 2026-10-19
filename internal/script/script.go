// Package script parses and runs YAML setup scripts that apply the same steps to many TVs.
package script

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"time"

	"go.yaml.in/yaml/v3"
)

// Step is one action. Exactly one field is set per step in YAML, e.g. `- key: [home, down]`.
type Step struct {
	Key        []string      `yaml:"key,omitempty"`
	Text       string        `yaml:"text,omitempty"`
	Click      string        `yaml:"click,omitempty"`
	WaitFor    string        `yaml:"wait_for,omitempty"`
	Wait       time.Duration `yaml:"wait,omitempty"`
	Launch     string        `yaml:"launch,omitempty"`
	Stop       string        `yaml:"stop,omitempty"`
	Install    string        `yaml:"install,omitempty"`
	Uninstall  string        `yaml:"uninstall,omitempty"`
	Shell      string        `yaml:"shell,omitempty"`
	Screenshot string        `yaml:"screenshot,omitempty"`
	Timeout    time.Duration `yaml:"timeout,omitempty"` // for wait_for, default 20s
}

// Script is a named list of steps.
type Script struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

// Load reads and validates a script file.
func Load(path string) (*Script, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Parse(data []byte) (*Script, error) {
	var s Script
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("script: %w", err)
	}
	if len(s.Steps) == 0 {
		return nil, fmt.Errorf("script has no steps")
	}
	for i, st := range s.Steps {
		if kinds := st.kinds(); len(kinds) != 1 {
			return nil, fmt.Errorf("step %d: needs exactly one action, got %v", i+1, kinds)
		}
	}
	return &s, nil
}

// Kind names the action a step performs.
func (st Step) Kind() string {
	k := st.kinds()
	if len(k) == 0 {
		return ""
	}
	return k[0]
}

func (st Step) kinds() []string {
	set := map[string]bool{
		"key":        len(st.Key) > 0,
		"text":       st.Text != "",
		"click":      st.Click != "",
		"wait_for":   st.WaitFor != "",
		"wait":       st.Wait > 0,
		"launch":     st.Launch != "",
		"stop":       st.Stop != "",
		"install":    st.Install != "",
		"uninstall":  st.Uninstall != "",
		"shell":      st.Shell != "",
		"screenshot": st.Screenshot != "",
	}
	var out []string
	for k, on := range set {
		if on {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
