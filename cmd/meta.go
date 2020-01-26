package cmd

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/b4b4r07/blog/pkg/blog"
	"github.com/b4b4r07/blog/pkg/shell"
	"github.com/dustin/go-humanize"
	"github.com/manifoldco/promptui"
)

type meta struct {
}

func (m *meta) init(args []string) error {
	return nil
}

func (m *meta) hugo(done <-chan bool) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-done:
			log.Printf("[DEBUG] [hugo process] finished")
			cancel()
		case <-ctx.Done():
		}
	}()

	hugo := shell.Shell{
		Stdin:   os.Stdin,
		Stdout:  ioutil.Discard,
		Stderr:  ioutil.Discard,
		Env:     map[string]string{},
		Command: "hugo",
		Args:    []string{"server", "-D"},
		Dir:     os.Getenv("BLOG_ROOT"),
	}

	go func() {
		log.Printf("[DEBUG] [hugo process] started as background process")
		hugo.Run(ctx)
	}()
}

func (m *meta) prompt() (blog.Article, error) {
	post := blog.Post{
		Path:  filepath.Join(os.Getenv("BLOG_ROOT"), os.Getenv("BLOG_POST_DIR")),
		Depth: 1,
	}
	err := post.Walk()
	if err != nil {
		return blog.Article{}, err
	}
	post.Articles.SortByDate()

	funcMap := promptui.FuncMap
	funcMap["time"] = humanize.Time
	funcMap["tags"] = func(tags []string) string {
		sort.Strings(tags)
		return strings.Join(tags, ", ")
	}
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   promptui.IconSelect + " {{ .Body.Title | cyan }}",
		Inactive: "  {{ .Body.Title | faint }}",
		Selected: promptui.IconGood + " {{ .Body.Title }}",
		Details: `
{{ "Date:" | faint }}	{{ .Date | time }}
{{ "Description:" | faint }}	{{ .Body.Description }}
{{ "Draft:" | faint }}	{{ .Body.Draft }}
{{ "Tags:" | faint }}	{{ .Body.Tags | tags }}
`,
		FuncMap: funcMap,
	}

	tagsContains := func(tags []string, input string) bool {
		for _, tag := range tags {
			if strings.ToLower(tag) == strings.ToLower(input) {
				return true
			}
		}
		return false
	}

	searcher := func(input string, index int) bool {
		article := post.Articles[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		title := strings.Replace(strings.ToLower(article.Body.Title), " ", "", -1)
		filename := strings.Replace(strings.ToLower(article.File), " ", "", -1)
		tagMatch := tagsContains(article.Body.Tags, input)
		return strings.Contains(title, input) || strings.Contains(filename, input) || tagMatch
	}

	prompt := promptui.Select{
		Label:             "Select:",
		Items:             post.Articles,
		Size:              10,
		Templates:         templates,
		Searcher:          searcher,
		StartInSearchMode: true,
		HideSelected:      true,
	}

	i, _, err := prompt.Run()
	return post.Articles[i], err
}
