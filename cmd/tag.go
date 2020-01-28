package cmd

import (
	"context"
	"strings"

	"github.com/b4b4r07/blog/pkg/blog"
	"github.com/b4b4r07/blog/pkg/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type tagCmd struct {
	meta
}

// newTagCmd creates a new tag command
func newTagCmd() *cobra.Command {
	c := &tagCmd{}

	tagCmd := &cobra.Command{
		Use:                   "tag",
		Short:                 "List tags",
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

	return tagCmd
}

func (c *tagCmd) run(args []string) error {
	var tags []string
	for _, article := range c.Post.Articles {
		tags = append(tags, article.Tags...)
	}

	tag, err := prompt(tags)
	if err != nil {
		return err
	}

	articles := c.Post.Articles

	articles.Filter(func(article blog.Article) bool {
		tagsContains := func(tags []string, input string) bool {
			for _, tag := range tags {
				if strings.ToLower(tag) == strings.ToLower(input) {
					return true
				}
			}
			return false
		}
		return tagsContains(article.Tags, tag)
	})

	var paths []string
	for _, article := range articles {
		paths = append(paths, article.Path)
	}

	editor := shell.New(c.Editor, paths...)
	return editor.Run(context.Background())
}

type Tag struct {
	Name  string
	Paths []string
	Num   int
}

func prompt(tags []string) (string, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   promptui.IconSelect + " {{ . | cyan }}",
		Inactive: "  {{ . | faint }}",
		Selected: promptui.IconGood + " {{ . }}",
		// 			Details: `
		// {{ "Date:" | faint }}	{{ .Date | time }}
		// {{ "Description:" | faint }}	{{ .Description }}
		// {{ "Draft:" | faint }}	{{ .Draft }}
		// {{ "Tags:" | faint }}	{{ .Tags | tags }}
		// `,
	}

	m := make(map[string]bool)
	uniq := []string{}
	for _, tag := range tags {
		if !m[tag] {
			m[tag] = true
			uniq = append(uniq, tag)
		}
	}
	tags = uniq

	searcher := func(input string, index int) bool {
		tag := strings.Replace(strings.ToLower(tags[index]), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(tag, input)
	}

	prompt := promptui.Select{
		Label:             "Select:",
		Items:             tags,
		Size:              10,
		Templates:         templates,
		Searcher:          searcher,
		StartInSearchMode: true,
		HideSelected:      true,
	}

	_, tag, err := prompt.Run()
	return tag, err
}
