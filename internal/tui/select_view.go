package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
)

func (m Model) View() string {
	switch m.state {
	case stateRunning:
		return renderRunView(m)
	case stateSummary:
		return renderSummaryView(m)
	default:
		return renderSelectView(m)
	}
}

func renderSelectView(m Model) string {
	var rows []string

	rows = append(rows, titleStyle.Render(m.title))
	rows = append(rows, subtitleStyle.Render(fmt.Sprintf("Filter: %s", displayFilter(m))))
	rows = append(rows, "")

	window := m.visibleWindow()
	for i, item := range window {
		prefix := "[ ]"
		if m.selected[item.ID] {
			prefix = "[x]"
		}

		line := fmt.Sprintf("%s %s", prefix, item.Title)
		description := mutedStyle.Render("  " + item.Description)
		if m.listOffset+i == m.cursor {
			line = selectedStyle.Render(line)
			description = selectedHintStyle.Render("  " + item.Description)
		}

		rows = append(rows, line)
		rows = append(rows, description)
	}

	rows = append(rows, "")
	rows = append(rows, mutedStyle.Render(fmt.Sprintf("Total: %d  Visible: %d  Selected: %d", len(m.allItems), len(m.visible), m.selectedCount())))
	rows = append(rows, help.New().View(selectKeys))

	return strings.Join(rows, "\n")
}

func displayFilter(m Model) string {
	if m.state == stateSearching {
		return m.searchInput.View()
	}
	if strings.TrimSpace(m.filterQuery) == "" {
		return "(none)"
	}
	return m.filterQuery
}
