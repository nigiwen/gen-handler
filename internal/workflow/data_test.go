package workflow

import (
	"path/filepath"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestDataWorkflowRunItemEmitsStepsInOrder(t *testing.T) {
	var events []ProgressEvent

	wf := DataWorkflow{
		Config:            types.Config{WireDir: filepath.Join(t.TempDir(), "cmd", "demo")},
		EntityDir:         filepath.Join(t.TempDir(), "internal", "model", "entity"),
		DataDir:           filepath.Join(t.TempDir(), "data"),
		GenerateEntity:    func(path string) error { return nil },
		GenerateRepo:      func(path string, entity types.EntityInfo) error { return nil },
		UpdateProviderSet: func(filePath, newItem, itemType string) error { return nil },
		RunWire:           func(wireDir string) error { return nil },
		FileExists:        func(path string) bool { return false },
	}

	item := Item{
		ID: "field_module",
		Payload: types.EntityInfo{
			FileName:   "field_module",
			EntityName: "FieldModule",
		},
	}

	result := wf.RunItem(item, func(ev ProgressEvent) {
		events = append(events, ev)
	})

	if !result.Success {
		t.Fatalf("expected success result, got %+v", result)
	}

	got := []string{events[0].Step, events[1].Step, events[2].Step, events[3].Step}
	want := []string{"检查 entity 占位文件", "生成 entity 占位文件", "生成 repo", "更新 ProviderSet"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected step order: got %v want %v", got, want)
		}
	}
}
