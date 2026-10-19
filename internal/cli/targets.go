package cli

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"tvremote/internal/fanout"
	"tvremote/internal/registry"
	"tvremote/internal/tv"
)

// targetFlags are the -t/--group/--all flags shared by every device command.
type targetFlags struct {
	names []string
	group string
	all   bool
}

func (t *targetFlags) bind(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringSliceVarP(&t.names, "tv", "t", nil, "device name (repeat or comma-separate)")
	f.StringVarP(&t.group, "group", "g", "", "every device in a group")
	f.BoolVar(&t.all, "all", false, "every registered device")
}

// target is one connected TV plus its registry name.
type target struct {
	*tv.Device
	Name string
}

// deviceAction runs on one connected TV.
type deviceAction func(ctx context.Context, t target) (any, error)

// onTargets connects to each selected TV and runs action in parallel, then prints the
// results. It returns an error when any device failed so the exit code reflects it.
func (a *app) onTargets(cmd *cobra.Command, t targetFlags, action deviceAction) error {
	reg, err := a.registry()
	if err != nil {
		return err
	}
	devices, err := reg.Select(registry.Selector{Names: t.names, Group: t.group, All: t.all})
	if err != nil {
		return err
	}
	client := a.client()
	results := fanout.Run(cmd.Context(), devices, a.parallel, func(ctx context.Context, rd registry.Device) (any, error) {
		ctx, cancel := a.deviceContext(ctx)
		defer cancel()
		d := tv.New(client, rd.Address)
		if err := d.Connect(ctx); err != nil {
			return nil, err
		}
		return action(ctx, target{Device: d, Name: rd.Name})
	})
	if err := a.printResults(cmd.OutOrStdout(), results); err != nil {
		return err
	}
	if !fanout.AllOK(results) {
		return errors.New("one or more devices failed")
	}
	return nil
}
