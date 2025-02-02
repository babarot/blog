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
	keys       AdditionalKeys
	editor     string
	rootDir    string
	contentDir string
	articles   []blog.Article
	list       list.Model
	toast      tea.Model
	showDraft  bool
	err        error
	quitting   bool
}

type errMsg struct {
	err error
}

func NewModel(editor, rootDir, contentDir string) Model {
	keys := AdditionalKeyMap()
	l := list.NewDefaultDelegate()
	l.ShortHelpFunc = keys.ShortHelp
	l.FullHelpFunc = keys.FullHelp
	m := Model{
		keys:       keys,
		editor:     editor,
		rootDir:    rootDir,
		contentDir: contentDir,
		articles:   []blog.Article{},
		list:       list.New(nil, l, 10, 30),
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

	m.articles, err = blog.Posts(m.rootDir, m.contentDir)
	if err != nil {
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
			m.quitting = true
			return m, tea.Quit
		}
		m.list.SetItems(msg.pages)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Draft):
			m.showDraft = !m.showDraft
			return m, m.loadArticles

		case key.Matches(msg, m.keys.Edit):
			if m.list.FilterState() != list.Filtering {
				if selected := m.list.SelectedItem(); selected != nil {
					article := selected.(blog.Article)
					slog.Debug("edit", "file", article.Meta.Title)
					return m, m.openEditor(article.Path)
				}
			}
		}

	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.quitting = true
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
	if m.quitting {
		return ""
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
