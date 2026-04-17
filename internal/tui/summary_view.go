package tui

import (
	"fmt"
	"strings"
)

func renderSummaryView(m Model) string {
	rows := []string{
		titleStyle.Render(m.title),
		"",
		fmt.Sprintf("Selected: %d", len(m.runQueue)),
		fmt.Sprintf("Success: %d", m.succeeded),
		fmt.Sprintf("Failed: %d", m.failed),
		fmt.Sprintf("Skipped: %d", m.skipped),
		fmt.Sprintf("Stopped: %t", m.stopAfterCurrent),
	}

	successes := resultTitles(m.results, func(result RunResult) bool {
		return result.Success && !result.Skipped
	})
	failures := resultTitles(m.results, func(result RunResult) bool {
		return result.Err != nil
	})

	if len(successes) > 0 {
		rows = append(rows, "", "Successful items")
		for _, title := range successes {
			rows = append(rows, "- "+title)
		}
	}

	if len(failures) > 0 {
		rows = append(rows, "", "Failed items")
		for _, title := range failures {
			rows = append(rows, "- "+title)
		}
	}

	rows = append(rows, "", helpStyle.Render(summaryKeysHelp))
	return strings.Join(rows, "\n")
}

func resultTitles(results []RunResult, keep func(RunResult) bool) []string {
	titles := make([]string, 0, len(results))
	for _, result := range results {
		if !keep(result) {
			continue
		}
		title := result.Title
		if strings.TrimSpace(title) == "" {
			title = result.ItemID
		}
		if result.Err != nil {
			title = fmt.Sprintf("%s (%v)", title, result.Err)
		}
		titles = append(titles, title)
	}
	return titles
}
