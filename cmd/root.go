package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/babarot/blog/internal/blog"
	"github.com/babarot/blog/internal/config"
	"github.com/babarot/blog/internal/env"
	"github.com/charmbracelet/log"
	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

var runID = sync.OnceValue(func() string {
	id := xid.New().String()
	return id
})

var (
	// Version is the version number
	Version = "unset"

	// BuildTag set during build to git tag, if any
	BuildTag = "unset"

	// BuildSHA is the git sha set during build
	BuildSHA = "unset"
)

var (
	ConfigPath string
)

// newRootCmd returns the root command
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                "blog",
		Short:              "A CLI tool for editing blog built by hugo etc",
		SilenceErrors:      true,
		DisableSuggestions: false,
		Version:            fmt.Sprintf("%s (%s/%s)", Version, BuildTag, BuildSHA),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			logDir := filepath.Dir(env.BLOG_LOG_PATH)
			if _, err := os.Stat(logDir); os.IsNotExist(err) {
				err := os.MkdirAll(logDir, 0755)
				if err != nil {
					return err
				}
			}

			var w io.Writer
			if file, err := os.OpenFile(env.BLOG_LOG_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				w = file
			} else {
				w = os.Stderr
			}

			logger := log.NewWithOptions(os.Stderr, log.Options{
				ReportCaller:    true,
				ReportTimestamp: true,
				TimeFormat:      time.Kitchen,
				Level:           log.DebugLevel,
				Formatter: func() log.Formatter {
					// if strings.ToLower(opt.Debug) == "json" {
					// 	return log.JSONFormatter
					// }
					return log.TextFormatter
				}(),
			})
			logger.SetOutput(w)
			logger.With("run_id", runID())
			slog.SetDefault(slog.New(logger))

			defer slog.Debug("root command finished")
			slog.Debug("root command started",
				"version", Version,
				"GoVersion", runtime.Version(),
				"buildTag/SHA", BuildTag+"/"+BuildSHA,
				"args", os.Args,
			)

			c, err := config.Parse("")
			if err != nil {
				return err
			}

			articles, err := blog.Posts(c.Hugo.RootDir, c.Hugo.ContentDir, 1)
			if err != nil {
				return err
			}

			c.Posts = articles
			c.LogWriter = w

			ctx := context.WithValue(cmd.Context(), config.Key, c)
			cmd.SetContext(ctx)

			return nil
		},
	}

	rootCmd.AddCommand(
		newEditCmd(),
		newNewCmd(),
		newLogsCmd(),
	)
	rootCmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", env.BLOG_CONFIG_PATH, "path to config")

	return rootCmd
}

func Execute() error {
	return newRootCmd().Execute()
}
