package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

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

// newRootCmd returns the root command
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                "blog",
		Short:              "A CLI tool for editing blog built by hugo etc",
		SilenceErrors:      true,
		DisableSuggestions: false,
		Version:            fmt.Sprintf("%s (%s/%s)", Version, BuildTag, BuildSHA),
	}

	rootCmd.AddCommand(
		newEditCmd(),
		newNewCmd(),
		newLogsCmd(),
	)
	return rootCmd
}

func Execute() error {
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

	return newRootCmd().Execute()
}
