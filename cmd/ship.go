package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Songmu/gitconfig"
	"github.com/b4b4r07/blog/pkg/blog"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type shipCmd struct {
	meta

	tags   bool
	noTags bool
}

// newShipCmd creates a new ship command
func newShipCmd() *cobra.Command {
	c := &shipCmd{}

	shipCmd := &cobra.Command{
		Use:                   "ship",
		Short:                 "Ship articles",
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

	return shipCmd
}

func (c *shipCmd) run(args []string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return errors.New("GITHUB_TOKEN is missing")
	}

	r, err := git.PlainOpen(c.RootPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}

	var articles blog.Articles
	for file := range status {
		// ref: https://github.com/src-d/go-git/issues/773
		path := filepath.Join(c.RootPath, file)
		article, err := blog.NewArticle(path)
		if err != nil {
			continue
		}
		articles = append(articles, article)
	}

	if len(articles) == 0 {
		return errors.New("no changed articles found")
	}

	article, err := c.prompt(articles...)
	if err != nil {
		return err
	}

	path, err := filepath.Rel(c.RootPath, article.Path)
	if err != nil {
		return err
	}

	_, err = w.Add(path)
	if err != nil {
		return err
	}

	user, err := gitconfig.User()
	if err != nil {
		return err
	}

	email, err := gitconfig.Email()
	if err != nil {
		return err
	}

	_, err = w.Commit(path, &git.CommitOptions{
		Author: &object.Signature{
			Name:  user,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	return r.Push(&git.PushOptions{
		// https://github.com/src-d/go-git/issues/637
		Auth: &http.BasicAuth{
			Username: user,
			Password: token,
		},
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
}
