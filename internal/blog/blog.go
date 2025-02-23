package blog

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/babarot/blog/internal/config"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v2"
)

const LocalHost = "http://localhost"

type Article struct {
	Meta

	config config.Blog

	Date     time.Time
	Filename string
	Dirname  string
	Path     string
}

// Article implements list.Item
var _ list.Item = (*Article)(nil)

func (p Article) URL() string {
	return path.Join(p.config.URL, "post", p.Date.Format("2006/01/02"), p.Slug())
}

func (p Article) DevURL() string {
	localhost := fmt.Sprintf("%s:%d", LocalHost, p.config.DevPort)
	return path.Join(localhost, "post", p.Date.Format("2006/01/02"), p.Slug())
}

func (p Article) Slug() string {
	slug := p.Dirname
	if regexp.MustCompile(`^20\d{2}$`).MatchString(slug) {
		slug = strings.TrimSuffix(p.Filename, filepath.Ext(p.Filename))
	}
	return slug
}

func (p Article) Description() string {
	const bullet = "•"
	return fmt.Sprintf("%s %s %s", p.Date.Format("2006-01-02"), bullet, p.Slug())
}

func (p Article) Title() string {
	var suffix string

	if p.Meta.Draft {
		draftSuffix := p.config.Draft.Suffix
		draftStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.config.Draft.Color))
		suffix = draftStyle.Render(" " + draftSuffix)
	}

	return p.Meta.Title + suffix
}

func (p Article) FilterValue() string {
	return p.Meta.Title + p.Slug()
}

type Meta struct {
	Title       string   `yaml:"title"`
	Date        string   `yaml:"date"`
	Description string   `yaml:"description"`
	Categories  []string `yaml:"categories"`
	Draft       bool     `yaml:"draft"`
	Author      string   `yaml:"author"`
	Oldlink     string   `yaml:"oldlink"`
	Image       string   `yaml:"image"`
	Tags        []string `yaml:"tags"`
	Aliases     []string `yaml:"aliases"`
	Toc         bool     `yaml:"toc"`
}

type Blog struct {
	Config   config.Blog
	Path     string
	Articles []Article
}

func Posts(c config.Config) ([]Article, error) {
	b := Blog{
		Config: c.Blog,
		Path:   filepath.Join(c.Hugo.RootDir, c.Hugo.ContentDir),
	}
	if err := b.Walk(); err != nil {
		return []Article{}, err
	}
	sort.Slice(b.Articles, func(i, j int) bool {
		return b.Articles[i].Date.After(b.Articles[j].Date)
	})
	return b.Articles, nil
}

func (p *Blog) Walk() error {
	return filepath.Walk(p.Path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if path == p.Path {
			return nil
		}
		switch filepath.Ext(path) {
		case ".md", ".mkd", ".markdown":
			// ok
		default:
			return nil
		}
		content, err := readFrontMatter(path)
		if err != nil {
			return err
		}
		var meta Meta
		if err = yaml.Unmarshal(content, &meta); err != nil {
			return err
		}

		formats := []string{
			"2006-01-02T15:04:05-07:00",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		var date time.Time
		for _, format := range formats {
			date, err = time.Parse(format, meta.Date)
			if err == nil {
				break
			}
		}
		if err != nil {
			slog.Warn("failed to parse datetime with all formats",
				"error", err,
				"input", meta.Date)
		}

		p.Articles = append(p.Articles, Article{
			config:   p.Config,
			Date:     date,
			Filename: filepath.Base(path),
			Dirname:  filepath.Base(filepath.Dir(path)),
			Path:     path,
			Meta:     meta,
		})
		return nil
	})
}

func readFrontMatter(path string) ([]byte, error) {
	var encount int
	var content string
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == "---" {
			encount++
		}
		if encount == 2 {
			break
		}
		content += scanner.Text() + "\n"
	}
	return []byte(content), scanner.Err()
}
