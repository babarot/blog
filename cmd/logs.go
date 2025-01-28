package cmd

import (
	"fmt"
	"io"

	"github.com/babarot/blog/internal/env"
	"github.com/nxadm/tail"
	"github.com/spf13/cobra"
)

type logsCmd struct {
	meta

	follow bool
}

func newLogsCmd() *cobra.Command {
	c := &logsCmd{}

	logsCmd := &cobra.Command{
		Use:                   "logs",
		Short:                 "Show logs",
		Aliases:               []string{},
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(args)
		},
	}

	f := logsCmd.Flags()
	f.BoolVarP(&c.follow, "follow", "f", false, "follow logs")

	return logsCmd
}

func (c *logsCmd) run(args []string) error {
	cfg := tail.Config{
		ReOpen: true,
		Poll:   true,
		Follow: true,
		Logger: tail.DiscardingLogger,
	}
	if c.follow {
		seekinfo := &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		}
		cfg.Location = seekinfo
	}
	t, err := tail.TailFile(env.BLOG_LOG_PATH, cfg)
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
	if err != nil {
		return err
	}
	return nil
}
