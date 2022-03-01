package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/b4b4r07/blog/pkg/blog"
	"github.com/b4b4r07/blog/pkg/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type newCmd struct {
	meta
}

// newNewCmd creates a new new command
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
			if err := c.meta.init(args); err != nil {
				return err
			}
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

	next := filepath.Join(c.PostDir, dirname, "index.md")
	hugo := shell.Shell{
		Command: "hugo",
		Args:    []string{"new", strings.TrimPrefix(next, "content/")},
		Dir:     c.RootPath,
		Env:     map[string]string{},
		Stdin:   os.Stdin,
		Stdout:  ioutil.Discard,
		Stderr:  ioutil.Discard,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := hugo.Run(ctx); err != nil {
		return err
	}

	go c.runHugoServer(ctx)
	defer log.Printf("[DEBUG] hugo: stopped server")

	article := blog.Article{
		Path: filepath.Join(c.RootPath, c.PostDir, dirname, "index.md"),
	}

	editor := shell.New(c.Editor, article.Path)
	return editor.Run(ctx)
}
