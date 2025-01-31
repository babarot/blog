package cmd

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/babarot/blog/internal/blog"
	"github.com/babarot/blog/internal/shell"
)

type meta struct {
	Editor   string
	RootPath string
	PostDir  string

	Post blog.Post
}

func (m *meta) init(args []string) error {
	rootPath := os.Getenv("BLOG_ROOT")
	if rootPath == "" {
		return errors.New("BLOG_ROOT is missing")
	}
	m.RootPath = rootPath

	postDir := os.Getenv("BLOG_POST_DIR")
	if postDir == "" {
		return errors.New("BLOG_POST_DIR is missing")
	}
	m.PostDir = postDir

	editor := os.Getenv("BLOG_EDITOR")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	m.Editor = editor

	post := blog.Post{
		Path:  filepath.Join(m.RootPath, m.PostDir),
		Depth: 1,
	}

	err := post.Walk()
	if err != nil {
		return err
	}

	post.Articles.SortByDate()
	m.Post = post

	return nil
}

func (m *meta) runHugoServer(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	hugo := shell.Shell{
		Command: "hugo",
		Args:    []string{"server", "-D"},
		Dir:     m.RootPath,
		Env:     map[string]string{},
		Stdin:   os.Stdin,
		Stdout:  ioutil.Discard,
		Stderr:  ioutil.Discard,
	}

	log.Printf("[DEBUG] hugo: run server")
	_ = hugo.Run(ctx)
}
