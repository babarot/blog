package cmd

import (
	"context"

	"github.com/b4b4r07/blog/pkg/shell"
	"github.com/spf13/cobra"
)

type editCmd struct {
	meta
}

// newEditCmd creates a new edit command
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

	return editCmd
}

func (c *editCmd) run(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.runHugoServer(ctx)

	article, err := c.prompt()
	if err != nil {
		return err
	}

	editor := shell.New(c.Editor, article.Path)
	return editor.Run(context.Background())
}
