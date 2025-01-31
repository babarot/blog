package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/babarot/blog/internal/config"
	"github.com/babarot/blog/internal/shell"
	"github.com/babarot/blog/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type editCmd struct {
	config config.Config

	tags    bool
	noTags  bool
	noDraft bool
}

func newEditCmd() *cobra.Command {
	c := &editCmd{}

	editCmd := &cobra.Command{
		Use:                   "edit",
		Short:                 "Edit existing articles",
		Aliases:               []string{},
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cmd.Context().Value(config.Key).(config.Config)
			c.config = cfg
			return c.run(args)
		},
	}

	f := editCmd.Flags()
	f.BoolVarP(&c.tags, "with-tags", "t", false, "with tags")
	f.BoolVarP(&c.noTags, "no-tags", "", false, "with no tags")
	f.BoolVarP(&c.noDraft, "no-draft", "", false, "with not in draft")

	return editCmd
}

func (c *editCmd) run(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hugo := shell.Shell{
		Command: "hugo",
		Args:    []string{"server", "-D"},
		Dir:     c.config.RootPath,
		Env:     map[string]string{},
		Stdin:   os.Stdin,
		Stdout:  c.config.LogWriter,
		Stderr:  c.config.LogWriter,
	}

	done := make(chan error)
	go func() {
		err := hugo.Run(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Debug("hugo canceled")
				done <- nil
				return
			}
			slog.Error("hugo failed", "error", err)
		} else {
			slog.Debug("hugo finished")
		}
		done <- err
	}()

	prog := tea.NewProgram(ui.Init(c.config.Posts))
	if _, err := prog.Run(); err != nil {
		return err
	}

	// stop hugo after UI stopped
	cancel()
	// wait for stopping hugo
	if err := <-done; err != nil {
		slog.Error("error failed")
		return err
	}

	return nil
}
