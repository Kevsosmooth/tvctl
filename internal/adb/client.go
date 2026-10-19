package adb

import (
	"context"
	"fmt"
	"strings"
)

// Client issues adb commands against one device serial (host:port for network devices).
type Client struct {
	runner Runner
}

func NewClient(r Runner) *Client {
	return &Client{runner: r}
}

// Connect attaches adb to a network device. adb exits 0 even when the connection fails,
// so the output text decides success.
func (c *Client) Connect(ctx context.Context, serial string) error {
	out, err := c.runner.Run(ctx, "connect", serial)
	if err != nil {
		return err
	}
	text := strings.TrimSpace(string(out))
	if strings.HasPrefix(text, "connected to") || strings.HasPrefix(text, "already connected") {
		return nil
	}
	return fmt.Errorf("connect %s: %s", serial, text)
}

func (c *Client) Disconnect(ctx context.Context, serial string) error {
	_, err := c.runner.Run(ctx, "disconnect", serial)
	return err
}

// Devices lists every device adb currently knows about.
func (c *Client) Devices(ctx context.Context) ([]DeviceState, error) {
	out, err := c.runner.Run(ctx, "devices", "-l")
	if err != nil {
		return nil, err
	}
	return ParseDevices(string(out)), nil
}

// Shell runs a shell command on the device and returns its trimmed output.
func (c *Client) Shell(ctx context.Context, serial string, args ...string) (string, error) {
	out, err := c.onDevice(ctx, serial, append([]string{"shell"}, args...)...)
	return strings.TrimRight(string(out), "\r\n"), err
}

// ExecOut returns raw binary output (screenshots).
func (c *Client) ExecOut(ctx context.Context, serial string, args ...string) ([]byte, error) {
	return c.onDevice(ctx, serial, append([]string{"exec-out"}, args...)...)
}

// Install installs or updates an APK, keeping app data.
func (c *Client) Install(ctx context.Context, serial, apkPath string) error {
	out, err := c.onDevice(ctx, serial, "install", "-r", apkPath)
	if err != nil {
		return err
	}
	if !strings.Contains(string(out), "Success") {
		return fmt.Errorf("install failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func (c *Client) Reboot(ctx context.Context, serial string) error {
	_, err := c.onDevice(ctx, serial, "reboot")
	return err
}
