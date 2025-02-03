package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	PrimaryColor       = lipgloss.Color("#ea9d34")
	SecondaryColor     = lipgloss.Color("#d7827e")
	TertiaryColor      = lipgloss.Color("#c53b53")
	PrimaryGrayColor   = lipgloss.Color("#767676")
	SecondaryGrayColor = lipgloss.Color("#3a3b5b")
	AccentColor        = lipgloss.Color("#a7cb77")
	SuccessColor       = lipgloss.Color("#58b4ad")
	BaseColor          = lipgloss.Color("#853d8a") // #ad58b4
)

var (
	infoStatusStyle   = lipgloss.NewStyle().Foreground(PrimaryColor)
	warnStatusStyle   = lipgloss.NewStyle().Foreground(TertiaryColor)
	debugStatusStyle  = lipgloss.NewStyle().Foreground(PrimaryGrayColor)
	noticeStatusStyle = lipgloss.NewStyle().Foreground(SuccessColor)
)

type ShowToastMsg struct {
	Message string
	Toast   ToastType
}

const (
	ToastInfo = iota
	ToastWarn
	ToastNotice
	ToastDebug
)

type ToastType = int

func ShowToast(message string, toast ToastType) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{Message: message, Toast: toast}
	}
}

func NewToast() *ToastModel {
	return &ToastModel{}
}

type ToastModel struct {
	message string
	toast   ToastType
}

func (m *ToastModel) Init() tea.Cmd {
	return nil
}

func (m *ToastModel) View() string {
	if len(m.message) == 0 {
		return ""
	}
	switch m.toast {
	case ToastInfo:
		return infoStatusStyle.Render("  " + m.message)
	case ToastWarn:
		return warnStatusStyle.Render("  " + m.message)
	case ToastNotice:
		return noticeStatusStyle.Render("  " + m.message)
	case ToastDebug:
		return debugStatusStyle.Render("  " + m.message)
	}
	return ""
}

func (m *ToastModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ShowToastMsg:
		m.message = msg.Message
		m.toast = msg.Toast
		cmd = clearStatus()
	case clearStatusMsg:
		m.message = ""
	}

	return m, cmd
}

type clearStatusMsg struct{}

func clearStatus() tea.Cmd {
	return tea.Tick(time.Second*5, func(_ time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}
