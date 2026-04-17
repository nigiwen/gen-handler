package tui

import "testing"

func TestFallbackDefaultsToFirstItemWhenInputEmpty(t *testing.T) {
	items := []Item{
		{ID: "field_module", Title: "field_module"},
		{ID: "process_node", Title: "process_node"},
	}

	selected := fallbackSelect(items, "")
	if len(selected) != 1 || selected[0].ID != "field_module" {
		t.Fatalf("unexpected fallback selection: %+v", selected)
	}
}
