package cmd

import (
	"fmt"
	"io"

	"github.com/babarot/blog/internal/config"
	"github.com/babarot/blog/internal/env"
	"github.com/nxadm/tail"
	"github.com/spf13/cobra"
)

type logsCmd struct {
	config config.Config

	followNew bool
}

func newLogsCmd() *cobra.Command {
	c := &logsCmd{}

	logsCmd := &cobra.Command{
		Use:                   "logs",
		Short:                 "Stream logs in real-time, like tail -f",
		Aliases:               []string{},
		GroupID:               "sub",
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

	f := logsCmd.Flags()
	f.BoolVarP(&c.followNew, "follow-new", "n", false, "Stream only new logs, ignoring existing ones.")

	return logsCmd
}

func (c *logsCmd) run(args []string) error {
	tailConfig := tail.Config{
		ReOpen: true,
		Poll:   true,
		Follow: true,
		Logger: tail.DiscardingLogger,
	}
	if c.followNew {
		tailConfig.Location = &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		}
	}
	t, err := tail.TailFile(env.BLOG_LOG_PATH, tailConfig)
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
	if err != nil {
		return err
	}
	return nil
}
