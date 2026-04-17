package tui

import (
	"fmt"
	"strings"
)

func renderRunView(m Model) string {
	total := len(m.runQueue)
	current := m.completed + 1
	if total == 0 {
		current = 0
	} else if current > total {
		current = total
	}

	rows := []string{
		titleStyle.Render(m.title),
		subtitleStyle.Render(fmt.Sprintf("Running %d/%d", current, total)),
		"",
		fmt.Sprintf("%s Processing", m.spinner.View()),
		fmt.Sprintf("Current item: %s", displayCurrentValue(m.currentItem)),
		fmt.Sprintf("Current step: %s", displayCurrentValue(m.currentStep)),
		"",
		fmt.Sprintf("Success: %d  Failed: %d  Skipped: %d", m.succeeded, m.failed, m.skipped),
	}

	if len(m.recent) > 0 {
		rows = append(rows, "", "Recent")
		for _, event := range m.recent {
			rows = append(rows, fmt.Sprintf("- [%s] %s: %s", event.Status, event.ItemID, event.Step))
		}
	}

	if m.stopAfterCurrent {
		rows = append(rows, "", warningStyle.Render("Soft stop requested. The current item will finish before exit."))
	} else {
		rows = append(rows, "", mutedStyle.Render("Press q to stop after the current item."))
	}

	rows = append(rows, helpStyle.Render(runKeysHelp))
	return strings.Join(rows, "\n")
}

func displayCurrentValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
