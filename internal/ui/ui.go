package ui

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/babarot/blog/internal/blog"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	articles  []blog.Article
	list      list.Model
	showDraft bool
	err       error
}

func Init(articles []blog.Article) Model {
	m := Model{
		articles:  articles,
		list:      list.New(nil, list.NewDefaultDelegate(), 10, 30),
		showDraft: false,
	}
	return m
}

var _ list.DefaultItem = (*blog.Article)(nil)

type postsLoadedMsg struct {
	pages []list.Item
}

func getArticles() (blog.Post, error) {
	var post blog.Post
	rootPath := os.Getenv("BLOG_ROOT")
	if rootPath == "" {
		return post, errors.New("BLOG_ROOT is missing")
	}
	postDir := os.Getenv("BLOG_POST_DIR")
	if postDir == "" {
		return post, errors.New("BLOG_POST_DIR is missing")
	}
	post = blog.Post{
		Path:  filepath.Join(rootPath, postDir),
		Depth: 1,
	}

	if err := post.Walk(); err != nil {
		return post, err
	}

	post.Articles.SortByDate()
	return post, nil
}

func (m Model) loadArticles() tea.Msg {
	var items []list.Item
	post, err := getArticles()
	if err != nil {
		// TODO
	}
	m.articles = post.Articles // TODO
	for _, article := range m.articles {
		if !m.showDraft {
			if article.Draft {
				continue
			}
		}
		items = append(items, article)
	}
	return postsLoadedMsg{pages: items}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadArticles,
	)
}

var KeyDraft = key.NewBinding(
	key.WithKeys("d"),
	key.WithHelp("d", "toggle draft"),
)
var Enter = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "edit"),
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)

	case postsLoadedMsg:
		m.list.SetItems(msg.pages)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, KeyDraft):
			m.showDraft = !m.showDraft
			return m, m.loadArticles
		case key.Matches(msg, Enter):
			if m.list.FilterState() != list.Filtering {
				if selected := m.list.SelectedItem(); selected != nil {
					article := selected.(blog.Article)
					return m, openEditor(article.Path)
				}
			}
		}

	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		return m, m.loadArticles
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}
	return m.list.View()
}

type editorFinishedMsg struct{ err error }

func openEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}
