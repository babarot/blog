package config

import (
	"io"

	"github.com/babarot/blog/internal/blog"
)

type KeyType string

const Key KeyType = "config"

type Config struct {
	Editor   string
	RootPath string
	PostDir  string
	// Blog      blog.Blog
	Posts     []blog.Article
	LogWriter io.Writer
}
