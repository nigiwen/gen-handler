# Data Sync 基于 `*.gen.go` 的实体发现实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `data` 命令只从 `internal/model/entity/*.gen.go` 发现候选实体，按需补 `internal/model/entity/<name>.go` 占位文件，停止生成 `data/dbset/*.go`，同时保持现有 `data/<name>.go`、`ProviderSet` 与 `wire` 逻辑不变。

**Architecture:** 保留现有 `cmd/data.go` 的交互主流程，把变化集中在三个点：scanner 只识别 `*.gen.go`；generator 提供 entity 占位文件写入能力并移除 dbset 生成；`cmd/data.go` 改为只根据 repo 文件是否存在判定待同步，并在 repo 生成前补 entity 占位文件。为了让这条链路可测，命令层会增加少量内部辅助函数和可替换的函数变量，用于隔离选择器、`ProviderSet` 更新和 `wire` 调用。

**Tech Stack:** Go、标准库文件系统 API、`go test`、现有模板生成与 ProviderSet 更新工具

---

### Task 1: 只识别 `*.gen.go` 的实体扫描

**Files:**
- Modify: `internal/scanner/entity.go`
- Test: `internal/scanner/entity_test.go`

- [ ] **Step 1: 写失败的 scanner 测试**

```go
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
```

- [ ] **Step 2: 运行测试，确认当前实现失败**

Run: `go test ./internal/scanner -run TestScanEntitiesOnlyUsesGenFiles -v`

Expected: FAIL，原因是当前实现会把 `project.go` 也扫出来，且不会去掉 `.gen` 后缀。

- [ ] **Step 3: 写最小实现**

```go
for _, file := range files {
	if file.IsDir() || !strings.HasSuffix(file.Name(), ".gen.go") {
		continue
	}

	fileName := strings.TrimSuffix(file.Name(), ".gen.go")
	if fileName == "" {
		continue
	}

	entities = append(entities, types.EntityInfo{
		EntityName: util.ToUpperCamel(fileName),
		FileName:   fileName,
	})
}
```

- [ ] **Step 4: 再跑测试，确认转绿**

Run: `go test ./internal/scanner -run TestScanEntitiesOnlyUsesGenFiles -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add internal/scanner/entity.go internal/scanner/entity_test.go
git commit -m "test: limit entity scanning to gen files"
```

### Task 2: 用 entity 占位文件替代 dbset 生成

**Files:**
- Modify: `internal/generator/data.go`
- Modify: `internal/generator/template.go`
- Test: `internal/generator/data_test.go`

- [ ] **Step 1: 写失败的 generator 测试**

```go
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
```

- [ ] **Step 2: 运行测试，确认当前实现失败**

Run: `go test ./internal/generator -run TestGenerateEntityFileWritesPackageOnly -v`

Expected: FAIL，原因是 `GenerateEntityFile` 尚不存在。

- [ ] **Step 3: 写最小实现并移除 dbset 生成入口**

```go
func GenerateEntityFile(path string) error {
	return util.WriteFile(path, EntityTemplate)
}

func GenerateRepoFile(path string, entity types.EntityInfo) error {
	data := struct {
		EntityName string
	}{
		EntityName: entity.EntityName,
	}

	content, err := ExecuteTemplate(RepoTemplate, data)
	if err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	return util.WriteFile(path, content)
}
```

```go
const EntityTemplate = `package entity
`
```

同时删除 `GenerateDbsetFile` 与 `DbsetTemplate`。

- [ ] **Step 4: 再跑测试，确认转绿**

Run: `go test ./internal/generator -run TestGenerateEntityFileWritesPackageOnly -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add internal/generator/data.go internal/generator/template.go internal/generator/data_test.go
git commit -m "feat: generate handwritten entity stubs"
```

### Task 3: 重排 `data` 命令流程并保持 repo 逻辑不变

**Files:**
- Modify: `cmd/data.go`
- Test: `cmd/data_test.go`

- [ ] **Step 1: 写失败的命令层测试**

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"

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
```

- [ ] **Step 2: 运行测试，确认当前实现失败**

Run: `go test ./cmd -run "Test(FilterPendingEntitiesIgnoresExistingEntityStub|SyncSelectedEntityCreatesEntityStubAndRepo)" -v`

Expected: FAIL，原因是 `filterPendingEntities`、`syncSelectedEntity`、可替换的函数变量当前都不存在，且现有逻辑仍依赖 dbset。

- [ ] **Step 3: 写最小实现**

```go
var (
	selectEntities    = selector.SelectItems[types.EntityInfo]
	updateProviderSet = updater.UpdateProviderSet
	runWireCommand    = generator.RunWireCommand
)

func filterPendingEntities(entities []types.EntityInfo, dataDir string) []types.EntityInfo {
	var pending []types.EntityInfo
	for _, entity := range entities {
		repoPath := filepath.Join(dataDir, entity.FileName+".go")
		if !util.FileExists(repoPath) {
			pending = append(pending, entity)
		}
	}
	return pending
}

func syncSelectedEntity(projectRoot string, config types.Config, entity types.EntityInfo) error {
	entityDir := filepath.Join(projectRoot, "internal", "model", "entity")
	dataDir := filepath.Join(projectRoot, "data")

	entityPath := filepath.Join(entityDir, entity.FileName+".go")
	if !util.FileExists(entityPath) {
		if err := generator.GenerateEntityFile(entityPath); err != nil {
			return err
		}
	}

	repoPath := filepath.Join(dataDir, entity.FileName+".go")
	if util.FileExists(repoPath) {
		return nil
	}

	if err := generator.GenerateRepoFile(repoPath, entity); err != nil {
		return err
	}

	dataGoPath := filepath.Join(dataDir, "data.go")
	if err := updateProviderSet(dataGoPath, "New"+entity.EntityName+"Repo", "Repo"); err != nil {
		return err
	}

	return runWireCommand(config.WireDir)
}
```

同时把 `RunDataCommand` 改为复用这些辅助函数，并删除 `dbsetDir` 创建和 `GenerateDbsetFile` 调用。

- [ ] **Step 4: 再跑测试，确认转绿**

Run: `go test ./cmd -run "Test(FilterPendingEntitiesIgnoresExistingEntityStub|SyncSelectedEntityCreatesEntityStubAndRepo)" -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add cmd/data.go cmd/data_test.go
git commit -m "feat: sync data from generated entities"
```

### Task 4: 文档收尾与全量验证

**Files:**
- Modify: `README.md`
- Modify: `docs/CHANGELOG.md`
- Modify: `docs/RELEASE_CHECKLIST.md`

- [ ] **Step 1: 更新文档中的 dbset 描述**

把以下描述改为新规则：

```md
- 扫描 `internal/model/entity/*.gen.go`
- 按需生成 `internal/model/entity/<name>.go`
- 生成 `data/<name>.go`
- 更新 `data/data.go` 的 `ProviderSet`
```

同时删除或改写所有“自动生成 dbset”的表述。

- [ ] **Step 2: 运行全量测试**

Run: `go test ./...`

Expected: PASS

- [ ] **Step 3: 运行构建验证**

Run: `go build .`

Expected: PASS

- [ ] **Step 4: 提交文档与最终代码**

```bash
git add README.md docs/CHANGELOG.md docs/RELEASE_CHECKLIST.md
git commit -m "docs: update data sync behavior"
```
