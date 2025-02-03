package ui

import (
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/babarot/blog/internal/blog"
	"github.com/babarot/blog/internal/config"
	"github.com/babarot/blog/internal/shell"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	keymap   *keymap
	list     list.Model
	toast    tea.Model
	err      error
	quitting bool

	editorCmd  string
	openCmd    string
	rootDir    string
	contentDir string
	showDraft  bool
}

type keymap struct {
	Quit  key.Binding
	Edit  key.Binding
	Open  key.Binding
	Draft key.Binding
}

func Init(c config.Config) Model {
	keymap := &keymap{
		Quit:  key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
		Edit:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("↵", "edit")),
		Open:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
		Draft: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "draft")),
	}
	l := list.New(nil, list.NewDefaultDelegate(), 10, 30)

	l.SetShowTitle(false)
	l.Title = c.Site.Name
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("#ee6ff8")). // #ee6ff8, #ad58b4, (#a743fd, #22222e, #706f8e)
		Foreground(lipgloss.Color("#22222e")).
		Padding(0, 1)
	l.SetShowStatusBar(true)
	l.StatusMessageLifetime = time.Second * 3

	l.DisableQuitKeybindings()
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keymap.Edit, keymap.Draft}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keymap.Edit, keymap.Open, keymap.Draft}
	}
	return Model{
		keymap:     keymap,
		list:       l,
		toast:      NewToast(),
		err:        nil,
		quitting:   false,
		editorCmd:  c.Editor,
		openCmd:    c.Open,
		rootDir:    c.Hugo.RootDir,
		contentDir: c.Hugo.ContentDir,
		showDraft:  false,
	}
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
		m.list.SetItems(msg.articles)

	case HugoServerMsg:
		cmds = append(cmds, ShowToast(" "+msg.Text, msg.Type))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Draft):
			m.showDraft = !m.showDraft
			msg := "hide draft posts!"
			if m.showDraft {
				msg = "show draft posts!"
			}
			cmds = append(cmds, m.list.NewStatusMessage(msg), ShowToast(msg, ToastNotice), m.loadArticles)
			// do not call m.list.Update
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keymap.Edit):
			if m.list.FilterState() != list.Filtering {
				if selected := m.list.SelectedItem(); selected != nil {
					article := selected.(blog.Article)
					slog.Debug("edit", "file", article.Meta.Title)
					return m, m.openEditor(article.Path)
				}
			}

		case key.Matches(msg, m.keymap.Open):
			if m.list.FilterState() != list.Filtering {
				if selected := m.list.SelectedItem(); selected != nil {
					article := selected.(blog.Article)
					slog.Debug("open", "folder", article.Dirname)
					return m, m.openFolder(article.Path)
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
		cmds = append(cmds, m.loadArticles)

	case openFinishedMsg:
		slog.Debug("openFinishedMsg")
		if msg.err != nil {
			slog.Error("openFinishedMsg", "error", msg.err)
			return m, ShowToast("failed to open", ToastWarn)
		}
		cmds = append(cmds, ShowToast("open "+strings.TrimPrefix(msg.target, m.rootDir+"/"), ToastNotice))

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

type openFinishedMsg struct {
	target string
	err    error
}

type HugoServerMsg struct {
	Text string
	Type ToastType
}

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
	if m.editorCmd == "" {
		return ShowToast("editor not set", ToastWarn)
	}
	c := shell.Command(m.editorCmd, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m Model) openFolder(path string) tea.Cmd {
	if m.openCmd == "" {
		return ShowToast("open command not set", ToastWarn)
	}
	dir := filepath.Dir(path)
	c := shell.Command(m.openCmd, dir)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return openFinishedMsg{target: dir, err: err}
	})
}
