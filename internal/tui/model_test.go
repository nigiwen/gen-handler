package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestSelectViewSpaceTogglesSelectionAndEnterReturnsCheckedItems(t *testing.T) {
	items := []Item{{ID: "field_module", Title: "field_module", Description: "Entity: FieldModule"}, {ID: "process_node", Title: "process_node", Description: "Entity: ProcessNode"}}
	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})
	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = modelIface.(Model)
	if !model.selected["field_module"] {
		t.Fatalf("expected field_module to be selected")
	}
	modelIface, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = modelIface.(Model)
	if model.state != stateRunning {
		t.Fatalf("expected stateRunning, got %v", model.state)
	}
	if len(model.runQueue) != 1 || model.runQueue[0].ID != "field_module" {
		t.Fatalf("unexpected runQueue: %+v", model.runQueue)
	}
}

func TestSelectViewSpaceRuneTogglesSelection(t *testing.T) {
	items := []Item{{ID: "field_module", Title: "field_module", Description: "Entity: FieldModule"}}
	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})

	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = modelIface.(Model)

	if !model.selected["field_module"] {
		t.Fatalf("expected field_module to be selected when space is reported as KeyRunes")
	}
}

func TestSelectViewEnterDefaultsToCurrentItemWhenNothingChecked(t *testing.T) {
	items := []Item{{ID: "test_case", Title: "test_case.go", Description: "Handler: TestCaseHandler, 12 个方法"}, {ID: "project", Title: "project.go", Description: "Handler: ProjectHandler, 8 个方法"}}
	model := NewModel(SessionConfig{Title: "Handler Generate", Items: items})
	model.cursor = 1
	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = modelIface.(Model)
	if len(model.runQueue) != 1 || model.runQueue[0].ID != "project" {
		t.Fatalf("expected current item to be queued, got %+v", model.runQueue)
	}
}

func TestSelectViewSlashEntersSearchAndFiltersVisibleItems(t *testing.T) {
	items := []Item{
		{ID: "field_module", Title: "field_module", Description: "Entity: FieldModule"},
		{ID: "process_node", Title: "process_node", Description: "Entity: ProcessNode"},
	}

	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})

	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = modelIface.(Model)

	if model.state != stateSearching {
		t.Fatalf("expected stateSearching, got %v", model.state)
	}

	modelIface, _ = model.Update(searchChangedMsg("field"))
	model = modelIface.(Model)

	if len(model.visible) != 1 || model.visible[0].ID != "field_module" {
		t.Fatalf("unexpected visible items: %+v", model.visible)
	}
}

func TestSelectViewKeyASelectsAllVisibleItems(t *testing.T) {
	items := []Item{
		{ID: "field_module", Title: "field_module", Description: "Entity: FieldModule"},
		{ID: "process_node", Title: "process_node", Description: "Entity: ProcessNode"},
	}

	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})
	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = modelIface.(Model)

	if !model.selected["field_module"] || !model.selected["process_node"] {
		t.Fatalf("expected all visible items to be selected, got %+v", model.selected)
	}
}

func TestRunViewProcessesProgressEventsAndMovesToSummary(t *testing.T) {
	items := []Item{{ID: "field_module", Title: "field_module", Description: "Entity: FieldModule"}}
	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})
	model.runQueue = items
	model.state = stateRunning

	modelIface, _ := model.Update(ProgressEvent{
		ItemID:  "field_module",
		Step:    "生成 repo",
		Status:  StatusSuccess,
		Message: "repo written",
	})
	model = modelIface.(Model)

	modelIface, _ = model.Update(runItemFinishedMsg{
		Result: RunResult{ItemID: "field_module", Success: true},
	})
	model = modelIface.(Model)

	if model.completed != 1 {
		t.Fatalf("expected completed=1, got %d", model.completed)
	}

	modelIface, _ = model.Update(runFinishedMsg{})
	model = modelIface.(Model)

	if model.state != stateSummary {
		t.Fatalf("expected stateSummary, got %v", model.state)
	}
}

func TestRunViewQRequestsSoftStop(t *testing.T) {
	model := NewModel(SessionConfig{Title: "Data Sync", Items: nil})
	model.state = stateRunning

	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = modelIface.(Model)

	if !model.stopAfterCurrent {
		t.Fatalf("expected soft stop to be requested")
	}
}

func TestSelectViewUpDownMovesCursorWithinBounds(t *testing.T) {
	items := []Item{
		{ID: "first", Title: "first", Description: "first"},
		{ID: "second", Title: "second", Description: "second"},
	}
	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})

	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = modelIface.(Model)
	if model.cursor != 1 {
		t.Fatalf("expected cursor to move down to 1, got %d", model.cursor)
	}

	modelIface, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = modelIface.(Model)
	if model.cursor != 1 {
		t.Fatalf("expected cursor to stay at 1, got %d", model.cursor)
	}

	modelIface, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = modelIface.(Model)
	if model.cursor != 0 {
		t.Fatalf("expected cursor to move up to 0, got %d", model.cursor)
	}

	modelIface, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = modelIface.(Model)
	if model.cursor != 0 {
		t.Fatalf("expected cursor to stay at 0, got %d", model.cursor)
	}
}

func TestSelectViewScrollsVisibleWindowWithCursor(t *testing.T) {
	items := []Item{
		{ID: "item-1", Title: "item-1", Description: "item-1"},
		{ID: "item-2", Title: "item-2", Description: "item-2"},
		{ID: "item-3", Title: "item-3", Description: "item-3"},
		{ID: "item-4", Title: "item-4", Description: "item-4"},
		{ID: "item-5", Title: "item-5", Description: "item-5"},
	}
	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})

	modelIface, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	model = modelIface.(Model)

	for i := 0; i < 4; i++ {
		modelIface, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = modelIface.(Model)
	}

	if model.cursor != 4 {
		t.Fatalf("expected cursor to move to 4, got %d", model.cursor)
	}

	view := model.View()
	if !strings.Contains(view, "item-5") {
		t.Fatalf("expected current item to be visible, got view:\n%s", view)
	}
	if strings.Contains(view, "item-1") {
		t.Fatalf("expected view to scroll past item-1, got view:\n%s", view)
	}
}

func TestSelectViewHighlightsCurrentItemTitleAndDescription(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	previousDarkBackground := lipgloss.HasDarkBackground()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(previousProfile)
		lipgloss.SetHasDarkBackground(previousDarkBackground)
	})

	items := []Item{
		{ID: "first", Title: "first", Description: "first description"},
		{ID: "second", Title: "second", Description: "second description"},
	}
	model := NewModel(SessionConfig{Title: "Data Sync", Items: items})
	model.cursor = 1

	view := model.View()

	highlightedTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("[ ] second")
	highlightedDescription := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Render("  second description")
	mutedCurrentDescription := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("  second description")
	mutedOtherDescription := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("  first description")

	if !strings.Contains(view, highlightedTitle) {
		t.Fatalf("expected current title to use highlighted color, got view:\n%s", view)
	}
	if !strings.Contains(view, highlightedDescription) {
		t.Fatalf("expected current description to use highlighted color, got view:\n%s", view)
	}
	if strings.Contains(view, mutedCurrentDescription) {
		t.Fatalf("expected current description not to use muted color, got view:\n%s", view)
	}
	if !strings.Contains(view, mutedOtherDescription) {
		t.Fatalf("expected non-current description to stay muted, got view:\n%s", view)
	}
}
