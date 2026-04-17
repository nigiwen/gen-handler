package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanEntitiesOnlyUsesGenFiles(t *testing.T) {
	dir := t.TempDir()

	files := []string{
		"project.gen.go",
		"project_member.gen.go",
		"project.go",
		"entity.go",
		"project_test.go",
	}

	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("package entity\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	entities, err := ScanEntities(dir)
	if err != nil {
		t.Fatalf("ScanEntities error: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(entities))
	}

	if entities[0].FileName != "project" || entities[0].EntityName != "Project" {
		t.Fatalf("unexpected first entity: %+v", entities[0])
	}

	if entities[1].FileName != "project_member" || entities[1].EntityName != "ProjectMember" {
		t.Fatalf("unexpected second entity: %+v", entities[1])
	}
}
