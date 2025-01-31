package ui

import (
	"log/slog"
	"os/exec"

	"github.com/babarot/blog/internal/blog"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	editor     string
	rootDir    string
	contentDir string
	articles   []blog.Article
	list       list.Model
	toast      tea.Model
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
		toast:      NewToast(),
		showDraft:  false,
		err:        nil,
	}
	return m
}

var _ list.DefaultItem = (*blog.Article)(nil)

type articlesLoadedMsg struct {
	pages []list.Item
	err   error
}

func (m Model) loadArticles() tea.Msg {
	var items []list.Item
	var err error

	m.articles, err = blog.Posts(m.rootDir, m.contentDir, 1)
	if err != nil {
		// return errMsg{err}
		return articlesLoadedMsg{err: err}
	}

	for _, article := range m.articles {
		if !m.showDraft {
			if article.Draft {
				continue
			}
		}
		items = append(items, article)
	}

	return articlesLoadedMsg{pages: items}
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
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)
	// global listeners
	m.toast, cmd = m.toast.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)

	case articlesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
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
					return m, m.openEditor(article.Path)
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
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}
	return m.list.View() + "\n" + m.toast.View()
}

type editorFinishedMsg struct{ err error }

func (m Model) openEditor(path string) tea.Cmd {
	if m.editor == "" {
		return ShowToast("editor not set", ToastWarn)
	}
	c := exec.Command(m.editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}
