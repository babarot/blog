package cmd

import (
	"strings"

	"github.com/k0kubun/pp"
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
	// tags := Tags{}
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

	tag, err := prompt(tags)
	if err != nil {
		return err
	}

	pp.Println(tag)
	return nil

	// articles := c.Post.Articles
	// articles.Filter(func(article blog.Article) bool {
	// 	tagsContains := func(tags []string, input string) bool {
	// 		for _, tag := range tags {
	// 			if strings.ToLower(tag) == strings.ToLower(input) {
	// 				return true
	// 			}
	// 		}
	// 		return false
	// 	}
	// 	return tagsContains(article.Tags, tag)
	// })
	//
	// var paths []string
	// for _, article := range articles {
	// 	paths = append(paths, article.Path)
	// }
	//
	// editor := shell.New(c.Editor, paths...)
	// return editor.Run(context.Background())
}

// type Tags map[string]Tag

type Tag struct {
	Name   string
	Titles []string
	Paths  []string
}

func prompt(tags []Tag) (Tag, error) {
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
