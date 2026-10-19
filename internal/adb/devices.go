package adb

import "strings"

// DeviceState is one line of `adb devices -l`.
type DeviceState struct {
	Serial string `json:"serial"`
	State  string `json:"state"` // device, offline, unauthorized
	Model  string `json:"model,omitempty"`
}

// ParseDevices reads `adb devices -l` output.
func ParseDevices(out string) []DeviceState {
	var devices []DeviceState
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.HasPrefix(line, "List of devices") || strings.HasPrefix(line, "*") {
			continue
		}
		d := DeviceState{Serial: fields[0], State: fields[1]}
		for _, f := range fields[2:] {
			if model, ok := strings.CutPrefix(f, "model:"); ok {
				d.Model = model
			}
		}
		devices = append(devices, d)
	}
	return devices
}
