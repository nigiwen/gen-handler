package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle        = lipgloss.NewStyle().Bold(true)
	subtitleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedHintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	mutedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	warningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)
