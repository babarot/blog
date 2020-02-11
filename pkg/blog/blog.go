package blog

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v2"
)

// Article represents the article information
type Article struct {
	Meta

	Date time.Time
	File string
	Path string
}

// Meta represents article contents
type Meta struct {
	Title       string   `yaml:"title"`
	Date        string   `yaml:"date"`
	Description string   `yaml:"description"`
	Categories  []string `yaml:"categories"`
	Draft       bool     `yaml:"draft"`
	Author      string   `yaml:"author"`
	Oldlink     string   `yaml:"oldlink"`
	Tags        []string `yaml:"tags"`
}

// Articles is a collection of articles
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

// Post represents
type Post struct {
	Path     string
	Depth    int
	Articles Articles
}

// Walk walks post directory and search markdow files
func (p *Post) Walk() error {
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
		article, err := NewArticle(path)
		if err != nil {
			return err
		}
		p.Articles = append(p.Articles, article)
		return nil
	})
}

func NewArticle(path string) (Article, error) {
	content, err := readFrontMatter(path)
	if err != nil {
		return Article{}, err
	}
	var meta Meta
	if err = yaml.Unmarshal(content, &meta); err != nil {
		return Article{}, err
	}
	date, _ := time.Parse("2006-01-02T15:04:05-07:00", meta.Date)
	return Article{
		Date: date,
		File: filepath.Base(path),
		Path: path,
		Meta: meta,
	}, nil
	// return nil
}

// SortByDate sorts by the date of the article
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
