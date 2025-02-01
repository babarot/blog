package ui

import "github.com/charmbracelet/bubbles/key"

type AdditionalKeys struct {
	Edit  key.Binding
	Draft key.Binding
}

func AdditionalKeyMap() AdditionalKeys {
	return AdditionalKeys{
		Edit:  key.NewBinding(key.WithKeys("enter", "e"), key.WithHelp("enter/e", "edit")),
		Draft: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "toggle draft")),
	}
}

func (k AdditionalKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Edit, k.Draft}
}

func (k AdditionalKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Edit, k.Draft},
	}
}
