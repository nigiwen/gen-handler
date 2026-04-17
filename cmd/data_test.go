package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nigiwen/gen-handler/internal/tui"
	"github.com/nigiwen/gen-handler/internal/types"
)

func TestFilterPendingEntitiesIgnoresExistingEntityStub(t *testing.T) {
	root := t.TempDir()
	entityDir := filepath.Join(root, "internal", "model", "entity")
	dataDir := filepath.Join(root, "data")

	if err := os.MkdirAll(entityDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(entityDir, "project.go"), []byte("package entity\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	entities := []types.EntityInfo{{FileName: "project", EntityName: "Project"}}
	pending := filterPendingEntities(entities, dataDir)

	if len(pending) != 1 || pending[0].FileName != "project" {
		t.Fatalf("expected project to stay pending, got %+v", pending)
	}
}

func TestSyncSelectedEntityCreatesEntityStubAndRepo(t *testing.T) {
	root := t.TempDir()
	entityDir := filepath.Join(root, "internal", "model", "entity")
	dataDir := filepath.Join(root, "data")

	if err := os.MkdirAll(entityDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "data.go"), []byte("package data\n\nvar ProviderSet = wire.NewSet()\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	calledProvider := false
	calledWire := false

	originalUpdateProviderSet := updateProviderSet
	originalRunWire := runWireCommand
	t.Cleanup(func() {
		updateProviderSet = originalUpdateProviderSet
		runWireCommand = originalRunWire
	})

	updateProviderSet = func(filePath, newItem, itemType string) error {
		calledProvider = true
		return nil
	}
	runWireCommand = func(wireDir string) error {
		calledWire = true
		return nil
	}

	err := syncSelectedEntity(root, types.Config{WireDir: filepath.Join(root, "cmd", "demo")}, types.EntityInfo{
		FileName:   "project",
		EntityName: "Project",
	})
	if err != nil {
		t.Fatalf("syncSelectedEntity error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(entityDir, "project.go")); err != nil {
		t.Fatalf("expected entity stub: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "project.go")); err != nil {
		t.Fatalf("expected repo file: %v", err)
	}
	if !calledProvider || !calledWire {
		t.Fatalf("expected provider update and wire run, got provider=%v wire=%v", calledProvider, calledWire)
	}
}

func TestSyncSelectedEntityKeepsExistingEntityStub(t *testing.T) {
	root := t.TempDir()
	entityDir := filepath.Join(root, "internal", "model", "entity")
	dataDir := filepath.Join(root, "data")

	if err := os.MkdirAll(entityDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(entityDir, "project.go"), []byte("package entity\n\n// keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "data.go"), []byte("package data\n\nvar ProviderSet = wire.NewSet()\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	originalUpdateProviderSet := updateProviderSet
	originalRunWire := runWireCommand
	t.Cleanup(func() {
		updateProviderSet = originalUpdateProviderSet
		runWireCommand = originalRunWire
	})

	updateProviderSet = func(filePath, newItem, itemType string) error { return nil }
	runWireCommand = func(wireDir string) error { return nil }

	err := syncSelectedEntity(root, types.Config{WireDir: filepath.Join(root, "cmd", "demo")}, types.EntityInfo{
		FileName:   "project",
		EntityName: "Project",
	})
	if err != nil {
		t.Fatalf("syncSelectedEntity error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(entityDir, "project.go"))
	if err != nil {
		t.Fatalf("read entity stub: %v", err)
	}
	if string(content) != "package entity\n\n// keep me\n" {
		t.Fatalf("expected existing entity stub to stay untouched, got %q", string(content))
	}
}

func TestRunDataCommandUsesUnifiedSession(t *testing.T) {
	root := t.TempDir()
	entityDir := filepath.Join(root, "internal", "model", "entity")
	dataDir := filepath.Join(root, "data")

	if err := os.MkdirAll(entityDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(entityDir, "project.gen.go"), []byte("package entity\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	originalRunSession := runSession
	t.Cleanup(func() {
		runSession = originalRunSession
	})

	var captured tui.SessionConfig
	runSession = func(cfg tui.SessionConfig) error {
		captured = cfg
		return nil
	}

	RunDataCommand(types.Config{WireDir: filepath.Join(root, "cmd", "demo")})

	if captured.Title != "Data Sync" {
		t.Fatalf("expected Data Sync title, got %q", captured.Title)
	}
	if len(captured.Items) != 1 || captured.Items[0].ID != "project" {
		t.Fatalf("unexpected session items: %+v", captured.Items)
	}
	if captured.Run == nil {
		t.Fatalf("expected session runner to be set")
	}
}
