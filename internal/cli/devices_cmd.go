package cli

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"tvremote/internal/adb"
	"tvremote/internal/registry"
	"tvremote/internal/tv"
)

// deviceRow is one line of `tvctl devices`.
type deviceRow struct {
	registry.Device
	State string `json:"state"`
}

func (a *app) devicesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devices",
		Short: "List registered TVs and whether adb can reach them",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			reg, err := a.registry()
			if err != nil {
				return err
			}
			states, err := a.client().Devices(cmd.Context())
			if err != nil {
				return err
			}
			rows := buildRows(reg.Devices, states)
			if a.jsonOut {
				return a.printJSON(cmd.OutOrStdout(), rows)
			}
			return printRows(cmd, rows)
		},
	}
}

func buildRows(devices []registry.Device, states []adb.DeviceState) []deviceRow {
	byAddr := map[string]string{}
	for _, s := range states {
		byAddr[s.Serial] = s.State
	}
	rows := make([]deviceRow, 0, len(devices))
	for _, d := range devices {
		state := byAddr[d.Address]
		if state == "" {
			state = "not connected"
		}
		rows = append(rows, deviceRow{Device: d, State: state})
	}
	return rows
}

func printRows(cmd *cobra.Command, rows []deviceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No devices yet. Add one: tvctl add <name> <host[:port]>")
		return nil
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tADDRESS\tSTATE\tMODEL\tGROUPS")
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.Address, r.State, r.Model, strings.Join(r.Groups, ","))
	}
	return w.Flush()
}

func (a *app) addCmd() *cobra.Command {
	var groups []string
	cmd := &cobra.Command{
		Use:   "add <name> <host[:port]>",
		Short: "Register a TV, connect to it and record its model",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := normalizeAddress(args[1])
			if err != nil {
				return err
			}
			reg, err := a.registry()
			if err != nil {
				return err
			}
			dev := registry.Device{Name: args[0], Address: addr, Groups: groups}
			if err := reg.Upsert(dev); err != nil {
				return err
			}
			ctx, cancel := a.deviceContext(cmd.Context())
			defer cancel()
			if dev.Model, err = probeModel(ctx, tv.New(a.client(), addr)); err != nil {
				return fmt.Errorf("%w (if the TV shows an \"Allow debugging?\" prompt, accept it and run add again)", err)
			}
			if err := reg.Upsert(dev); err != nil {
				return err
			}
			if err := reg.Save(); err != nil {
				return err
			}
			if a.jsonOut {
				return a.printJSON(cmd.OutOrStdout(), dev)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added %s (%s) at %s\n", dev.Name, dev.Model, dev.Address)
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&groups, "group", "g", nil, "groups for bulk commands (repeatable)")
	return cmd
}

func probeModel(ctx context.Context, d *tv.Device) (string, error) {
	if err := d.Connect(ctx); err != nil {
		return "", err
	}
	info, err := d.Info(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(info.Manufacturer + " " + info.Model), nil
}

// normalizeAddress accepts host or host:port and defaults to adb's port 5555.
var hostPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.:-]*$`)

// normalizeAddress accepts host or host:port (adb's default port 5555 when missing)
// and allows only hostname, IPv4 and IPv6 characters, since it ends up in adb's arguments.
func normalizeAddress(s string) (string, error) {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		host, port = s, "5555"
	}
	if !hostPattern.MatchString(host) {
		return "", fmt.Errorf("invalid address %q", s)
	}
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("invalid port in %q", s)
	}
	return net.JoinHostPort(host, port), nil
}

func (a *app) removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Forget a TV and disconnect adb from it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := a.registry()
			if err != nil {
				return err
			}
			dev, err := reg.Get(args[0])
			if err != nil {
				return err
			}
			// Disconnect is best effort: the TV may already be offline.
			_ = a.client().Disconnect(cmd.Context(), dev.Address)
			if err := reg.Remove(dev.Name); err != nil {
				return err
			}
			if err := reg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", dev.Name)
			return nil
		},
	}
}

func (a *app) connectCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect adb to TVs (other commands do this automatically)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.onTargets(cmd, t, func(context.Context, target) (any, error) { return nil, nil })
		},
	}
	t.bind(cmd)
	return cmd
}
