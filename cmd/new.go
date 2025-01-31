package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/babarot/blog/internal/config"
	"github.com/babarot/blog/internal/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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
	validate := func(input string) error {
		invalids := []string{"/", "_", " "}
		for _, invalid := range invalids {
			if strings.Contains(input, invalid) {
				return fmt.Errorf("%q cannot be used for filepath", invalid)
			}
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "New URL path",
		Validate: validate,
	}

	dirname, err := prompt.Run()
	if err != nil {
		return err
	}

	next := filepath.Join(c.config.PostDir, dirname, "index.md")
	hugo := shell.Shell{
		Command: "hugo",
		Args:    []string{"new", strings.TrimPrefix(next, "content/")},
		Dir:     c.config.RootPath,
		Env:     map[string]string{},
		Stdin:   os.Stdin,
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}
	_ = hugo

	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer stop()
	//
	// if err := hugo.Run(ctx); err != nil {
	// 	return err
	// }
	//
	// go c.runHugoServer(ctx)
	// defer log.Printf("[DEBUG] hugo: stopped server")
	//
	// article := blog.Article{
	// 	Path: filepath.Join(c.RootPath, c.PostDir, dirname, "index.md"),
	// }
	//
	// editor := shell.New(c.Editor, article.Path)
	// return editor.Run(ctx)
	return nil
}
