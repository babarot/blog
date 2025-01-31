package ui

import (
	"os"
	"os/exec"

	"github.com/babarot/blog/internal/blog"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/exp/slog"
)

type Model struct {
	editor     string
	rootDir    string
	contentDir string
	articles   []blog.Article
	list       list.Model
	showDraft  bool
	err        error
}

type errMsg struct {
	err error
}

func NewModel(editor, rootDir, contentDir string) Model {
	m := Model{
		editor:     editor,
		rootDir:    rootDir,
		contentDir: contentDir,
		articles:   []blog.Article{},
		list:       list.New(nil, list.NewDefaultDelegate(), 10, 30),
		showDraft:  false,
		err:        nil,
	}
	return m
}

var _ list.DefaultItem = (*blog.Article)(nil)

type postsLoadedMsg struct {
	pages []list.Item
}

func (m Model) loadArticles() tea.Msg {
	var items []list.Item

	articles, err := blog.Posts(m.rootDir, m.contentDir, 1)
	if err != nil {
		return errMsg{err}
	}

	m.articles = articles // TODO
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

	case errMsg:
		if msg.err != nil {
			slog.Warn("got an error", "error", msg.err)
			m.err = msg.err
			return m, nil
		}

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
