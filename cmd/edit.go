package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/b4b4r07/blog/pkg/blog"
	"github.com/b4b4r07/blog/pkg/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type editCmd struct {
	meta

	tags   bool
	noTags bool
}

// newEditCmd creates a new edit command
func newEditCmd() *cobra.Command {
	c := &editCmd{}

	editCmd := &cobra.Command{
		Use:                   "edit",
		Short:                 "Edit existing articles",
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

	f := editCmd.Flags()
	f.BoolVarP(&c.tags, "with-tags", "t", false, "with tags")
	f.BoolVarP(&c.noTags, "no-tags", "", false, "with no  tags")

	return editCmd
}

func (c *editCmd) run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go c.runHugoServer(ctx)
	defer log.Printf("[DEBUG] hugo: stopped server")

	switch {
	default:
		return c.withTitles(ctx, args)
	case c.tags:
		return c.withTags(ctx, args)
	case c.noTags:
		return c.withNoTags(ctx, args)
	}
}

func (c *editCmd) withTitles(ctx context.Context, args []string) error {
	article, err := c.prompt()
	if err != nil {
		return err
	}

	editor := shell.New(c.Editor, article.Path)
	return editor.Run(ctx)
}

func (c *editCmd) withTags(ctx context.Context, args []string) error {
	tt := map[string]Tag{}
	for _, article := range c.Post.Articles {
		for _, tag := range article.Tags {
			t := tt[tag]
			t.Titles = append(t.Titles, article.Title)
			t.Paths = append(t.Paths, article.Path)
			tt[tag] = t
		}
	}

	var tags []Tag
	for name, tag := range tt {
		tags = append(tags, Tag{
			Name:   name,
			Titles: tag.Titles,
			Paths:  tag.Paths,
		})
	}

	tag, err := selectTags(tags)
	if err != nil {
		return err
	}

	editor := shell.New(c.Editor, tag.Paths...)
	return editor.Run(ctx)
}

func (c *editCmd) withNoTags(ctx context.Context, args []string) error {
	c.Post.Articles.Filter(func(article blog.Article) bool {
		return len(article.Tags) == 0
	})

	article, err := c.prompt()
	if err != nil {
		return err
	}

	editor := shell.New(c.Editor, article.Path)
	return editor.Run(ctx)
}

type Tag struct {
	Name   string
	Titles []string
	Paths  []string
}

func selectTags(tags []Tag) (Tag, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ .Name }}",
		Active:   promptui.IconSelect + " {{ .Name | cyan }}",
		Inactive: "  {{ .Name | faint }}",
		Selected: promptui.IconGood + " {{ .Name }}",
		Details: `
{{ "Titles:" | faint }}
    {{ .Titles | len | faint }} {{ "article(s)" | faint }}
    {{- range .Titles }}
    - {{ . }}
    {{- end }}
`,
	}

	searcher := func(input string, index int) bool {
		tag := strings.Replace(strings.ToLower(tags[index].Name), " ", "", -1)
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

	i, _, err := prompt.Run()
	return tags[i], err
}
