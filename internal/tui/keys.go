package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/key"
)

var (
	keySelect = tea.KeySpace
	keySubmit = tea.KeyEnter
)

type selectKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	All    key.Binding
	Search key.Binding
	Submit key.Binding
	Quit   key.Binding
}

func (k selectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Search, k.Submit}
}

func (k selectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle, k.All},
		{k.Search, k.Submit, k.Quit},
	}
}

var selectKeys = selectKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "move up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "move down")),
	Toggle: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle")),
	All:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle visible")),
	Search: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "run")),
	Quit:   key.NewBinding(key.WithKeys("esc", "q"), key.WithHelp("esc/q", "quit")),
}

const (
	runKeysHelp     = "q stop after current  ctrl+c force quit"
	summaryKeysHelp = "enter/q quit"
)
