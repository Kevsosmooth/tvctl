package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tvremote/internal/keys"
	"tvremote/internal/uitree"
)

func (a *app) keyCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "key <key>...",
		Short: "Press remote buttons in order (see `tvctl keys`)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return nil, d.Keys(ctx, args...)
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) keysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keys",
		Short: "List key names for `tvctl key`",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.jsonOut {
				return a.printJSON(cmd.OutOrStdout(), keys.Names())
			}
			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(keys.Names(), " "))
			return nil
		},
	}
}

func (a *app) textCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "text <text>",
		Short: "Type text into the focused field",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return nil, d.Text(ctx, args[0])
			})
		},
	}
	t.bind(cmd)
	return cmd
}

func (a *app) tapCmd() *cobra.Command {
	var t targetFlags
	cmd := &cobra.Command{
		Use:   "tap <x> <y>",
		Short: "Tap a screen position (from `tvctl ui`)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			x, errX := strconv.Atoi(args[0])
			y, errY := strconv.Atoi(args[1])
			if errX != nil || errY != nil {
				return fmt.Errorf("x and y must be whole numbers")
			}
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return nil, d.Tap(ctx, x, y)
			})
		},
	}
	t.bind(cmd)
	return cmd
}

// queryFlags are --id/--exact for commands that find an element by text.
type queryFlags struct {
	id    string
	exact bool
}

func (q *queryFlags) bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&q.id, "id", "", "match resource id instead of (or with) text")
	cmd.Flags().BoolVar(&q.exact, "exact", false, "require the whole text to match")
}

func (q queryFlags) query(args []string) uitree.Query {
	text := ""
	if len(args) > 0 {
		text = args[0]
	}
	return uitree.Query{Text: text, ID: q.id, Exact: q.exact}
}

func (a *app) clickCmd() *cobra.Command {
	var t targetFlags
	var q queryFlags
	cmd := &cobra.Command{
		Use:   "click [text]",
		Short: "Find an element on screen by text or --id and tap it",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return d.Click(ctx, q.query(args))
			})
		},
	}
	t.bind(cmd)
	q.bind(cmd)
	return cmd
}

func (a *app) waitForCmd() *cobra.Command {
	var t targetFlags
	var q queryFlags
	var wait time.Duration
	cmd := &cobra.Command{
		Use:   "wait-for [text]",
		Short: "Wait until an element appears on screen",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.onTargets(cmd, t, func(ctx context.Context, d target) (any, error) {
				return d.WaitFor(ctx, q.query(args), wait)
			})
		},
	}
	t.bind(cmd)
	q.bind(cmd)
	cmd.Flags().DurationVar(&wait, "wait", 20*time.Second, "how long to wait")
	return cmd
}
