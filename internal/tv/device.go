// Package tv holds the operations tvctl performs on one TV, built on adb.
package tv

import (
	"context"

	"tvremote/internal/adb"
)

// Device is one connected TV addressed by its adb serial (host:port).
type Device struct {
	client *adb.Client
	serial string
}

func New(client *adb.Client, serial string) *Device {
	return &Device{client: client, serial: serial}
}

func (d *Device) Serial() string { return d.serial }

func (d *Device) shell(ctx context.Context, args ...string) (string, error) {
	return d.client.Shell(ctx, d.serial, args...)
}

// Connect makes sure adb is attached and authorized before any command runs.
func (d *Device) Connect(ctx context.Context) error {
	return d.client.Connect(ctx, d.serial)
}

func (d *Device) Install(ctx context.Context, apkPath string) error {
	return d.client.Install(ctx, d.serial, apkPath)
}

func (d *Device) Reboot(ctx context.Context) error {
	return d.client.Reboot(ctx, d.serial)
}
