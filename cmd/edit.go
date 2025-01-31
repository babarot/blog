package cmd

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/babarot/blog/internal/shell"
	"github.com/babarot/blog/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type editCmd struct {
	meta

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
			if err := c.meta.init(args); err != nil {
				return err
			}
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
		Command:     "hugo",
		Args:        []string{"server", "-D"},
		Dir:         c.RootPath,
		Env:         map[string]string{},
		Stdin:       os.Stdin,
		Stdout:      io.Discard,
		Stderr:      io.Discard,
		StartingMsg: "hugo serving",
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

	prog := tea.NewProgram(ui.Init(c.Post.Articles))
	_, err := prog.Run()
	if err != nil {
		return err
	}

	// stop hugo
	cancel()

	if err := <-done; err != nil {
		slog.Error("error failed")
		return err
	}

	return nil
}
