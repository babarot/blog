package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"unicode"

	"github.com/babarot/blog/internal/blog"
	"github.com/babarot/blog/internal/config"
	"github.com/babarot/blog/internal/shell"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type newCmd struct {
	config config.Config
}

func newNewCmd() *cobra.Command {
	c := &newCmd{}

	newCmd := &cobra.Command{
		Use:                   "new",
		Short:                 "Create new article",
		Aliases:               []string{},
		GroupID:               "main",
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

	return newCmd
}

func (c *newCmd) run(args []string) error {
	var (
		slug  string
		title string
		toc   bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What’s for slug?").
				Prompt("? ").
				Validate(func(s string) error {
					re := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
					if re.MatchString(s) {
						return nil
					}
					return errors.New("invalid chars included for slug")
				}).
				Value(&slug),
			huh.NewInput().
				Title("What’s for title?").
				Prompt("? ").
				Validate(func(s string) error {
					isAllowed := func(s string) bool {
						for _, ch := range s {
							if !(unicode.IsLetter(ch) ||
								unicode.IsDigit(ch) ||
								unicode.IsSpace(ch) ||
								unicode.IsPunct(ch) ||
								unicode.IsSymbol(ch)) {
								return false
							}
						}
						return true
					}
					if isAllowed(s) {
						return nil
					}
					return errors.New("invalid chars included for title")
				}).
				Value(&title),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Show table of contents?").
				Affirmative("Yes!").
				Negative("No.").
				Value(&toc),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	date := time.Now().Format("2006-01-02T15:04:05-07:00")
	year := time.Now().Year()
	meta := blog.Meta{
		Title: title,
		Toc:   toc,
		Date:  date,
	}
	mdFile := fmt.Sprintf("%s/%d/%s/index.md", c.config.Hugo.ContentDir, year, slug)

	hugo := shell.Shell{
		Command: "hugo new " + mdFile,
		Dir:     c.config.Hugo.RootDir,
		Stdout:  c.config.LogWriter,
		Stderr:  c.config.LogWriter,
	}

	if err := hugo.Run(context.Background()); err != nil {
		return fmt.Errorf("failed to run hugo new: %w", err)
	}

	data, err := yaml.Marshal(&meta)
	if err != nil {
		return fmt.Errorf("error marshalling to YAML: %w", err)
	}

	mdPath := filepath.Join(c.config.Hugo.RootDir, mdFile)
	file, err := os.Create(mdPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	frontMatter := fmt.Sprintf("---\n%s---\n", string(data))
	_, err = file.Write([]byte(frontMatter))
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		hugoServer := shell.Shell{
			Command: c.config.Hugo.Command,
			Dir:     c.config.Hugo.RootDir,
			Stdout:  c.config.LogWriter,
			Stderr:  c.config.LogWriter,
		}
		err := hugoServer.Run(ctx)
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

	slog.Debug("running", "editor", c.config.Editor)
	if err := shell.RunCommand(fmt.Sprintf("%s %s", c.config.Editor, mdPath)); err != nil {
		return fmt.Errorf("failed to run %s: %w", c.config.Editor, err)
	}

	// stop hugo after editing
	cancel()
	// wait for stopping hugo
	if err := <-done; err != nil {
		slog.Error("error failed")
		return err
	}

	return nil
}
