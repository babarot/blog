package ui

import (
	"log/slog"
	"os/exec"

	"github.com/babarot/blog/internal/blog"
	"github.com/babarot/blog/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	keymap   *keymap
	list     list.Model
	toast    tea.Model
	err      error
	quitting bool

	editor     string
	rootDir    string
	contentDir string
	showDraft  bool
}

type keymap struct {
	Quit  key.Binding
	Edit  key.Binding
	Draft key.Binding
}

func Init(c config.Config) Model {
	keymap := &keymap{
		Quit:  key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
		Edit:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("↵", "edit")),
		Draft: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "draft")),
	}
	l := list.New(nil, list.NewDefaultDelegate(), 10, 30)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.DisableQuitKeybindings()
	l.AdditionalFullHelpKeys = func() []key.Binding { return []key.Binding{keymap.Edit, keymap.Draft} }
	l.AdditionalShortHelpKeys = func() []key.Binding { return []key.Binding{keymap.Edit, keymap.Draft} }
	return Model{
		keymap:     keymap,
		list:       l,
		toast:      NewToast(),
		err:        nil,
		quitting:   false,
		editor:     c.Editor,
		rootDir:    c.Hugo.RootDir,
		contentDir: c.Hugo.ContentDir,
		showDraft:  false,
	}
}

var _ list.DefaultItem = (*blog.Article)(nil)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		ShowToast("hugo server is running in background!", ToastInfo),
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
		m.list.SetItems(msg.articles)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Draft):
			m.showDraft = !m.showDraft
			return m, m.loadArticles

		case key.Matches(msg, m.keymap.Edit):
			if m.list.FilterState() != list.Filtering {
				if selected := m.list.SelectedItem(); selected != nil {
					article := selected.(blog.Article)
					slog.Debug("edit", "file", article.Meta.Title)
					return m, m.openEditor(article.Path)
				}
			}
		}

	case editorFinishedMsg:
		slog.Debug("editorFinishedMsg")
		if msg.err != nil {
			slog.Error("editorFinishedMsg", "error", msg.err)
			m.err = msg.err
			return m, tea.Quit
		}
		return m, m.loadArticles

	case errMsg:
		if msg.error != nil {
			slog.Error("errMsg", "error", msg.error)
			m.err = msg.error
			return m, tea.Quit
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

// msgs

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

type articlesLoadedMsg struct{ articles []list.Item }

type editorFinishedMsg struct{ err error }

// cmds

func (m Model) loadArticles() tea.Msg {
	var items []list.Item

	articles, err := blog.Posts(m.rootDir, m.contentDir)
	if err != nil {
		return errMsg{err}
	}

	for _, article := range articles {
		if !m.showDraft {
			if article.Draft {
				continue
			}
		}
		items = append(items, article)
	}

	return articlesLoadedMsg{articles: items}
}

func (m Model) openEditor(path string) tea.Cmd {
	if m.editor == "" {
		return ShowToast("editor not set", ToastWarn)
	}
	c := exec.Command(m.editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}
