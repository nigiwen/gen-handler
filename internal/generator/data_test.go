package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateEntityFileWritesPackageOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "project.go")

	if err := GenerateEntityFile(path); err != nil {
		t.Fatalf("GenerateEntityFile error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	if string(content) != "package entity\n" {
		t.Fatalf("unexpected content: %q", string(content))
	}
}
