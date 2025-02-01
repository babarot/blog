package blog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Article struct {
	Meta

	Date     time.Time
	Filename string
	Dirname  string
	Path     string
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
		suffix = " ::Draft"
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

type Articles []Article

func (as *Articles) Filter(f func(Article) bool) {
	articles := make(Articles, 0)
	for _, a := range *as {
		if f(a) {
			articles = append(articles, a)
		}
	}
	*as = articles
}

type Blog struct {
	Path     string
	Depth    int
	Articles Articles
}

func Posts(root, dir string, depth int) ([]Article, error) {
	b := Blog{
		Path:  filepath.Join(root, dir),
		Depth: depth,
	}
	if err := b.Walk(); err != nil {
		return []Article{}, err
	}
	b.Articles.SortByDate()
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
		date, _ := time.Parse("2006-01-02T15:04:05-07:00", meta.Date)
		p.Articles = append(p.Articles, Article{
			Date:     date,
			Filename: filepath.Base(path),
			Dirname:  filepath.Base(filepath.Dir(path)),
			Path:     path,
			Meta:     meta,
		})
		return nil
	})
}

func (as *Articles) SortByDate() {
	sort.Slice(*as, func(i, j int) bool {
		return (*as)[i].Date.After((*as)[j].Date)
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
