# Handler Completion Implementation Plan

> **Status:** 已执行完成。下面原始任务清单保留为历史计划记录，复选框未逐项回填；实际落地结果以本文顶部“Execution Summary”与“Verification”小节为准。

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `handler` 命令默认列出全部 gRPC 服务，并在执行时按“创建或补全”矩阵处理 `api/grpc/<service>.go` 与 `core/<service>.go`，只在本次创建过文件时更新 ProviderSet 和执行 `wire`。

**Architecture:** `cmd/handler.go` 不再按“缺失 handler”过滤服务，统一把全部解析结果交给 `HandlerWorkflow`。`internal/workflow/handler.go` 改成基于 `grpcCreated` / `coreCreated` 的分支执行器；`internal/generator/{handler,core}.go` 新增 `Ensure...` 入口，在文件不存在时创建完整文件，在文件存在时通过 AST 识别缺失方法并只追加缺失方法。共享的 Go 文件解析、import 补齐和方法收集逻辑放到一个小的 generator 辅助文件中，避免把 AST 处理塞进业务入口函数。

**Tech Stack:** Go、标准库 `go/ast` / `go/parser` / `go/token`、现有 `internal/tui` / `internal/util` / `internal/updater`、`go test`

---

## Execution Summary

- Task 1 已完成：`handler` 命令默认列出全部解析出的服务，不再按“缺失 handler”预过滤。
- Task 2 已完成：workflow 已按 `grpcCreated` / `coreCreated` 分支，只有在新建对应文件时才更新 provider，且只有创建了任意文件时才执行 `wire`。
- Task 3 已完成：`internal/generator/handler.go` 已实现 `EnsureHandlerFile`，已有 handler 文件只补缺失方法，不覆盖已有实现。
- Task 4 已完成：`internal/generator/core.go` 已实现 `EnsureCoreServiceFile`，已有 core 文件只补缺失方法，不覆盖已有实现。
- 实施过程补充完成：生成器已加入“同名方法签名漂移”检测，发现签名不一致时直接报错并保持文件不变。
- 实施过程补充完成：`handler` 命令错误已改为向上传播，grpc 扫描/解析失败、session 失败会导致 CLI 非零退出。
- 实施过程补充完成：主 proto 包名与主 proto import 已去掉对 `devopsx` 的硬编码，改为根据 grpc 文件 package 与 `ProtoDir` 推导；`grpc.go` 更新逻辑也已同步改成使用服务自身 proto 包。

## Actual File Map

- Modified: `main.go`
- Modified: `cmd/handler.go`
- Modified: `cmd/handler_test.go`
- Modified: `cmd/session.go`
- Modified: `internal/types/types.go`
- Modified: `internal/parser/grpc_parser.go`
- Added: `internal/parser/grpc_parser_test.go`
- Modified: `internal/workflow/handler.go`
- Modified: `internal/workflow/handler_test.go`
- Modified: `internal/generator/template.go`
- Modified: `internal/generator/handler.go`
- Modified: `internal/generator/core.go`
- Added: `internal/generator/go_file.go`
- Modified: `internal/generator/handler_test.go`
- Added: `internal/generator/core_test.go`
- Modified: `internal/updater/grpc.go`
- Added: `internal/updater/grpc_test.go`

## Verification

- Passed: `go test ./cmd ./internal/parser ./internal/generator ./internal/updater ./internal/workflow`
- Passed: `go test ./...`
- Passed: `go build .`

---

## File Map

- Modify: `cmd/handler.go`
  - 去掉“缺失 handler”过滤，直接把全部服务交给 workflow
- Modify: `cmd/handler_test.go`
  - 验证 `handler` 命令把全部服务交给 session，且不再调用 `findMissingHandlers`
- Modify: `cmd/session.go`
  - 删除不再需要的 `findMissingHandlers` 测试桩注入点
- Modify: `internal/workflow/handler.go`
  - 将固定 5 步改成按 `created` 标志执行的分支流程
- Modify: `internal/workflow/handler_test.go`
  - 增加四种矩阵场景的表驱动测试
- Modify: `internal/generator/template.go`
  - 新增 `HandlerMethodTemplate` / `CoreMethodTemplate`，并让整文件模板按方法包名决定 import
- Modify: `internal/generator/handler.go`
  - 新增 `EnsureHandlerFile`，已有文件走补全逻辑
- Modify: `internal/generator/core.go`
  - 新增 `EnsureCoreServiceFile`，已有文件走补全逻辑
- Create: `internal/generator/go_file.go`
  - 放共享的 AST 解析、缺失方法计算、import 补齐与追加写回逻辑
- Create: `internal/generator/handler_test.go`
  - 覆盖 handler 文件创建、补全、保留已有实现、错误路径
- Create: `internal/generator/core_test.go`
  - 覆盖 core 文件创建、补全、保留已有实现、错误路径

### Task 1: 让 handler 命令默认列出全部服务

**Files:**
- Modify: `cmd/handler.go`
- Modify: `cmd/handler_test.go`
- Modify: `cmd/session.go`

- [ ] **Step 1: 写失败测试，锁定“不过滤服务”的命令行为**

在 `cmd/handler_test.go` 中将现有测试替换为：

```go
package cmd

import (
	"path/filepath"
	"testing"

	"github.com/nigiwen/gen-handler/internal/tui"
	"github.com/nigiwen/gen-handler/internal/types"
)

func TestRunHandlerCommandUsesAllParsedServicesInSession(t *testing.T) {
	originalGlob := globGrpcFiles
	originalParseGrpcFile := parseGrpcFile
	originalFindMissingHandlers := findMissingHandlers
	originalRunSession := runSession
	t.Cleanup(func() {
		globGrpcFiles = originalGlob
		parseGrpcFile = originalParseGrpcFile
		findMissingHandlers = originalFindMissingHandlers
		runSession = originalRunSession
	})

	globGrpcFiles = func(pattern string) ([]string, error) {
		return []string{filepath.Join(t.TempDir(), "project_grpc.pb.go")}, nil
	}
	parseGrpcFile = func(filePath string) ([]types.ServiceInfo, error) {
		return []types.ServiceInfo{
			{
				FileName:    "project.go",
				HandlerName: "ProjectHandler",
				ServiceName: "ProjectService",
			},
			{
				FileName:    "task.go",
				HandlerName: "TaskHandler",
				ServiceName: "TaskService",
			},
		}, nil
	}
	findMissingHandlers = func(services []types.ServiceInfo, outputDir string) []types.ServiceInfo {
		t.Fatalf("findMissingHandlers should not be called")
		return nil
	}

	var captured tui.SessionConfig
	runSession = func(cfg tui.SessionConfig) error {
		captured = cfg
		return nil
	}

	RunHandlerCommand(types.Config{
		OutputDir: "./api/grpc",
		CoreDir:   "./core",
		WireDir:   "./cmd/devopsx",
	}, "./internal/proto/axis/devopsx")

	if captured.Title != "Handler Generate" {
		t.Fatalf("expected Handler Generate title, got %q", captured.Title)
	}
	if len(captured.Items) != 2 {
		t.Fatalf("expected all parsed services to be passed through, got %+v", captured.Items)
	}
	if captured.Items[0].ID != "project.go" || captured.Items[1].ID != "task.go" {
		t.Fatalf("unexpected session items: %+v", captured.Items)
	}
	if captured.Run == nil {
		t.Fatalf("expected session runner to be set")
	}
}
```

- [ ] **Step 2: 运行测试，确认它先红**

Run: `go test ./cmd -run TestRunHandlerCommandUsesAllParsedServicesInSession -v`

Expected: FAIL，原因是当前 `RunHandlerCommand` 仍会调用 `findMissingHandlers`，并且只把过滤后的结果传给 session。

- [ ] **Step 3: 写最小实现，去掉缺失过滤**

在 `cmd/handler.go` 中把 `missingServices` 分支改成直接使用 `services`：

```go
	if len(services) == 0 {
		fmt.Println("⚠️  未解析到任何服务接口")
		return
	}

	wf := workflow.HandlerWorkflow{
		Config:                config,
		EnsureHandlerFile:     generator.EnsureHandlerFile,
		UpdateGrpcProvider:    generator.UpdateGrpcProvider,
		EnsureCoreServiceFile: generator.EnsureCoreServiceFile,
		UpdateCoreProvider:    generator.UpdateCoreProvider,
		RunWire:               generator.RunWireCommand,
	}

	if err := runSession(tui.SessionConfig{
		Title: "Handler Generate",
		Items: wf.BuildItems(services),
		Run:   wf.RunItem,
	}); err != nil {
		fmt.Printf("❌ Handler Generate 失败: %v\n", err)
	}
```

在 `cmd/session.go` 中删除 `findMissingHandlers` 变量，只保留：

```go
var (
	scanEntities  = scanner.ScanEntities
	runSession    = tui.RunSession
	globGrpcFiles = filepath.Glob
	parseGrpcFile = parser.ParseGrpcFile
)
```

并把 `cmd/handler_test.go` 中的清理逻辑同步改为不再保存/恢复 `findMissingHandlers`。

- [ ] **Step 4: 再跑命令层测试，确认转绿**

Run: `go test ./cmd -run TestRunHandlerCommandUsesAllParsedServicesInSession -v`

Expected: PASS

- [ ] **Step 5: 记录这一小步**

在这个仓库里，任何 `git` 命令都必须先征得用户同意。先问用户是否要执行下面这组命令；只有在用户明确同意后再运行：

```bash
git add cmd/handler.go cmd/handler_test.go cmd/session.go
git commit -m "feat: list all handler services in session"
```

### Task 2: 让 workflow 按创建结果决定 ProviderSet 和 wire

**Files:**
- Modify: `internal/workflow/handler.go`
- Modify: `internal/workflow/handler_test.go`
- Modify: `cmd/handler.go`

- [ ] **Step 1: 写失败测试，锁定四种执行矩阵**

将 `internal/workflow/handler_test.go` 改为表驱动测试：

```go
package workflow

import (
	"reflect"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestHandlerWorkflowRunItemBranchesByFileCreation(t *testing.T) {
	cases := []struct {
		name           string
		grpcCreated    bool
		coreCreated    bool
		wantSteps      []string
		wantUpdateGrpc bool
		wantUpdateCore bool
		wantWire       bool
	}{
		{
			name:           "create both files",
			grpcCreated:    true,
			coreCreated:    true,
			wantSteps:      []string{"处理 handler 文件", "更新 grpc.go", "处理 core service", "更新 core ProviderSet", "运行 wire"},
			wantUpdateGrpc: true,
			wantUpdateCore: true,
			wantWire:       true,
		},
		{
			name:           "create core only",
			grpcCreated:    false,
			coreCreated:    true,
			wantSteps:      []string{"处理 handler 文件", "处理 core service", "更新 core ProviderSet", "运行 wire"},
			wantUpdateGrpc: false,
			wantUpdateCore: true,
			wantWire:       true,
		},
		{
			name:           "create grpc only",
			grpcCreated:    true,
			coreCreated:    false,
			wantSteps:      []string{"处理 handler 文件", "更新 grpc.go", "处理 core service", "运行 wire"},
			wantUpdateGrpc: true,
			wantUpdateCore: false,
			wantWire:       true,
		},
		{
			name:           "complete existing files only",
			grpcCreated:    false,
			coreCreated:    false,
			wantSteps:      []string{"处理 handler 文件", "处理 core service"},
			wantUpdateGrpc: false,
			wantUpdateCore: false,
			wantWire:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotSteps []string
			var updatedGrpc bool
			var updatedCore bool
			var ranWire bool

			wf := HandlerWorkflow{
				Config: types.Config{
					OutputDir: "./api/grpc",
					CoreDir:   "./core",
					WireDir:   "./cmd/devopsx",
				},
				EnsureHandlerFile: func(service types.ServiceInfo, config types.Config) (bool, error) {
					return tc.grpcCreated, nil
				},
				UpdateGrpcProvider: func(service types.ServiceInfo, outputDir string) error {
					updatedGrpc = true
					return nil
				},
				EnsureCoreServiceFile: func(service types.ServiceInfo, config types.Config) (bool, error) {
					return tc.coreCreated, nil
				},
				UpdateCoreProvider: func(service types.ServiceInfo, coreDir string) error {
					updatedCore = true
					return nil
				},
				RunWire: func(wireDir string) error {
					ranWire = true
					return nil
				},
			}

			item := Item{
				ID:    "project.go",
				Title: "project.go",
				Payload: types.ServiceInfo{
					FileName:    "project.go",
					HandlerName: "ProjectHandler",
					ServiceName: "ProjectService",
				},
			}

			result := wf.RunItem(item, func(ev ProgressEvent) {
				gotSteps = append(gotSteps, ev.Step)
			})

			if !result.Success {
				t.Fatalf("expected success result, got %+v", result)
			}
			if updatedGrpc != tc.wantUpdateGrpc {
				t.Fatalf("unexpected UpdateGrpcProvider call: got %v want %v", updatedGrpc, tc.wantUpdateGrpc)
			}
			if updatedCore != tc.wantUpdateCore {
				t.Fatalf("unexpected UpdateCoreProvider call: got %v want %v", updatedCore, tc.wantUpdateCore)
			}
			if ranWire != tc.wantWire {
				t.Fatalf("unexpected RunWire call: got %v want %v", ranWire, tc.wantWire)
			}
			if !reflect.DeepEqual(gotSteps, tc.wantSteps) {
				t.Fatalf("unexpected steps: got %v want %v", gotSteps, tc.wantSteps)
			}
		})
	}
}
```

- [ ] **Step 2: 运行 workflow 测试，确认它先红**

Run: `go test ./internal/workflow -run TestHandlerWorkflowRunItemBranchesByFileCreation -v`

Expected: FAIL，原因是当前 `HandlerWorkflow` 仍使用固定步骤函数签名，并且 `RunItem` 无条件更新 `grpc.go`、`core.go` 和执行 `wire`。

- [ ] **Step 3: 写最小实现，把 workflow 改成 `Ensure + created` 语义**

在 `internal/workflow/handler.go` 中改成下面这套接口与执行顺序：

```go
package workflow

import appTypes "github.com/nigiwen/gen-handler/internal/types"

type HandlerWorkflow struct {
	Config                appTypes.Config
	EnsureHandlerFile     func(service appTypes.ServiceInfo, config appTypes.Config) (bool, error)
	UpdateGrpcProvider    func(service appTypes.ServiceInfo, outputDir string) error
	EnsureCoreServiceFile func(service appTypes.ServiceInfo, config appTypes.Config) (bool, error)
	UpdateCoreProvider    func(service appTypes.ServiceInfo, coreDir string) error
	RunWire               func(wireDir string) error
}

func (wf HandlerWorkflow) BuildItems(services []appTypes.ServiceInfo) []Item {
	items := make([]Item, 0, len(services))
	for _, service := range services {
		items = append(items, Item{
			ID:          service.FileName,
			Title:       service.FileName,
			Description: "Handler: " + service.HandlerName,
			Keywords:    []string{service.FileName, service.HandlerName, service.ServiceName},
			Payload:     service,
		})
	}
	return items
}

func (wf HandlerWorkflow) RunItem(item Item, emit func(ProgressEvent)) RunResult {
	service := item.Payload.(appTypes.ServiceInfo)

	emit(ProgressEvent{ItemID: item.ID, Step: "处理 handler 文件"})
	grpcCreated, err := wf.EnsureHandlerFile(service, wf.Config)
	if err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	if grpcCreated {
		emit(ProgressEvent{ItemID: item.ID, Step: "更新 grpc.go"})
		if err := wf.UpdateGrpcProvider(service, wf.Config.OutputDir); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "处理 core service"})
	coreCreated, err := wf.EnsureCoreServiceFile(service, wf.Config)
	if err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	if coreCreated {
		emit(ProgressEvent{ItemID: item.ID, Step: "更新 core ProviderSet"})
		if err := wf.UpdateCoreProvider(service, wf.Config.CoreDir); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	if grpcCreated || coreCreated {
		emit(ProgressEvent{ItemID: item.ID, Step: "运行 wire"})
		if err := wf.RunWire(wf.Config.WireDir); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	return RunResult{ItemID: item.ID, Title: item.Title, Success: true}
}
```

同时在 `cmd/handler.go` 中把 workflow 字段名改成新的 `Ensure...` 版本：

```go
	wf := workflow.HandlerWorkflow{
		Config:                config,
		EnsureHandlerFile:     generator.EnsureHandlerFile,
		UpdateGrpcProvider:    generator.UpdateGrpcProvider,
		EnsureCoreServiceFile: generator.EnsureCoreServiceFile,
		UpdateCoreProvider:    generator.UpdateCoreProvider,
		RunWire:               generator.RunWireCommand,
	}
```

这里故意不为“跳过的步骤”发送 `ProgressEvent`，这样测试和 UI 行为都更稳定。

- [ ] **Step 4: 再跑 workflow 测试，确认转绿**

Run: `go test ./internal/workflow -run TestHandlerWorkflowRunItemBranchesByFileCreation -v`

Expected: PASS

- [ ] **Step 5: 记录这一小步**

在这个仓库里，任何 `git` 命令都必须先征得用户同意。先问用户是否要执行下面这组命令；只有在用户明确同意后再运行：

```bash
git add internal/workflow/handler.go internal/workflow/handler_test.go cmd/handler.go
git commit -m "feat: branch handler workflow by file creation"
```

### Task 3: 为 handler 文件实现“创建或补全”

**Files:**
- Modify: `internal/generator/template.go`
- Modify: `internal/generator/handler.go`
- Create: `internal/generator/go_file.go`
- Create: `internal/generator/handler_test.go`

- [ ] **Step 1: 写失败测试，锁定 handler 创建与补全语义**

在 `internal/generator/handler_test.go` 中加入：

```go
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func testProjectService() types.ServiceInfo {
	return types.ServiceInfo{
		ServerName:  "ProjectServer",
		HandlerName: "ProjectHandler",
		FileName:    "project.go",
		FieldName:   "projectSrv",
		ServiceName: "ProjectService",
		Methods: []types.Method{
			{
				Name:         "List",
				RequestPkg:   "devopsx",
				RequestType:  "ListProjectRequest",
				ResponsePkg:  "devopsx",
				ResponseType: "ListProjectReply",
			},
			{
				Name:         "Ping",
				RequestPkg:   "basic",
				RequestType:  "Empty",
				ResponsePkg:  "basic",
				ResponseType: "Empty",
			},
		},
	}
}

func TestEnsureHandlerFileCreatesFullFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "type ProjectHandler struct") {
		t.Fatalf("expected handler struct in generated file, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectHandler) List(") {
		t.Fatalf("expected List method in generated file, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected basic import in generated file, got %q", got)
	}
}

func TestEnsureHandlerFileAppendsMissingMethodsWithoutOverwritingExistingBody(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	path := filepath.Join(dir, service.FileName)
	content := `package grpc

import (
	"context"

	"bsi/axis/devopsx/core"
	"bsi/axis/devopsx/internal/proto/axis/devopsx"
)

type ProjectHandler struct {
	devopsx.UnimplementedProjectServer
	projectSrv *core.ProjectService
}

func (p *ProjectHandler) List(ctx context.Context, in *devopsx.ListProjectRequest) (*devopsx.ListProjectReply, error) {
	return &devopsx.ListProjectReply{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing handler file: %v", err)
	}

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated handler file: %v", err)
	}

	got := string(updated)
	if strings.Count(got, "func (p *ProjectHandler) List(") != 1 {
		t.Fatalf("expected existing List method to stay single, got %q", got)
	}
	if !strings.Contains(got, "return &devopsx.ListProjectReply{}, nil") {
		t.Fatalf("expected existing List body to be preserved, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectHandler) Ping(") {
		t.Fatalf("expected missing Ping method to be appended, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected missing basic import to be added, got %q", got)
	}
}

func TestEnsureHandlerFileReturnsErrorWhenHandlerTypeMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	if err := os.WriteFile(path, []byte("package grpc\n"), 0o644); err != nil {
		t.Fatalf("write invalid handler file: %v", err)
	}

	_, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err == nil || !strings.Contains(err.Error(), "ProjectHandler") {
		t.Fatalf("expected missing type error, got %v", err)
	}
}
```

- [ ] **Step 2: 运行 generator 测试，确认它先红**

Run: `go test ./internal/generator -run "TestEnsureHandlerFile(CreateFullFileWhenMissing|AppendsMissingMethodsWithoutOverwritingExistingBody|ReturnsErrorWhenHandlerTypeMissing)" -v`

Expected: FAIL，原因是 `EnsureHandlerFile`、方法级模板和共享 AST helper 还不存在。

- [ ] **Step 3: 写最小实现，引入 `EnsureHandlerFile` 和共享 Go 文件 helper**

在 `internal/generator/template.go` 中新增方法级模板，并让整文件模板按方法包名决定 import：

```go
const HandlerTemplate = `package grpc

import (
	"context"

	"{{.ModulePath}}/core"
	"{{.ModulePath}}/internal/proto/axis/devopsx"{{if .UsesBasic}}
	"{{.ModulePath}}/internal/proto/basic"{{end}}{{if .UsesZebra}}
	"{{.ModulePath}}/internal/proto/zebra"{{end}}
)

type {{.HandlerName}} struct {
	devopsx.Unimplemented{{.ServerName}}
	{{.FieldName}} *core.{{.ServiceName}}
}

func New{{.HandlerName}}({{.FieldName}} *core.{{.ServiceName}}) *{{.HandlerName}} {
	return &{{.HandlerName}}{
		{{.FieldName}}: {{.FieldName}},
	}
}
{{range .Methods}}
{{if .Comment}}// {{.Comment}}{{else}}// {{.Name}}{{end}}
// @name devopsx.{{$.FieldName | trimSrv}}.{{.Name}}
// @desc {{if .Comment}}{{.Comment}}{{else}}{{.Name}}{{end}}
func ({{$.FieldName | firstChar}} *{{$.HandlerName}}) {{.Name}}(ctx context.Context, in *{{.RequestPkg}}.{{.RequestType}}) (*{{.ResponsePkg}}.{{.ResponseType}}, error) {
	return {{$.FieldName | firstChar}}.{{$.FieldName}}.{{.Name}}(ctx, in)
}
{{end}}
`

const HandlerMethodTemplate = `{{if .Method.Comment}}// {{.Method.Comment}}{{else}}// {{.Method.Name}}{{end}}
// @name devopsx.{{.Service.FieldName | trimSrv}}.{{.Method.Name}}
// @desc {{if .Method.Comment}}{{.Method.Comment}}{{else}}{{.Method.Name}}{{end}}
func ({{.Service.FieldName | firstChar}} *{{.Service.HandlerName}}) {{.Method.Name}}(ctx context.Context, in *{{.Method.RequestPkg}}.{{.Method.RequestType}}) (*{{.Method.ResponsePkg}}.{{.Method.ResponseType}}, error) {
	return {{.Service.FieldName | firstChar}}.{{.Service.FieldName}}.{{.Method.Name}}(ctx, in)
}
`
```

在 `internal/generator/go_file.go` 中加入共享 helper：

```go
package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

type fileImport struct {
	Alias string
	Path  string
}

func usesPkg(methods []types.Method, pkg string) bool {
	for _, method := range methods {
		if method.RequestPkg == pkg || method.ResponsePkg == pkg {
			return true
		}
	}
	return false
}

func protoImports(modulePath string, methods []types.Method) []fileImport {
	imports := []fileImport{
		{Path: modulePath + "/internal/proto/axis/devopsx"},
	}
	if usesPkg(methods, "basic") {
		imports = append(imports, fileImport{Path: modulePath + "/internal/proto/basic"})
	}
	if usesPkg(methods, "zebra") {
		imports = append(imports, fileImport{Path: modulePath + "/internal/proto/zebra"})
	}
	return imports
}

func collectExistingMethods(content, typeName string) (map[string]struct{}, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析 Go 文件失败: %w", err)
	}
	if !hasType(node, typeName) {
		return nil, fmt.Errorf("未找到类型定义: %s", typeName)
	}
	methods := make(map[string]struct{})
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		if receiverTypeName(fn.Recv.List[0].Type) == typeName {
			methods[fn.Name.Name] = struct{}{}
		}
	}
	return methods, nil
}

func hasType(node *ast.File, typeName string) bool {
	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if ok && ts.Name.Name == typeName {
				return true
			}
		}
	}
	return false
}

func receiverTypeName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		if ident, ok := x.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func missingMethods(all []types.Method, existing map[string]struct{}) []types.Method {
	missing := make([]types.Method, 0, len(all))
	for _, method := range all {
		if _, ok := existing[method.Name]; ok {
			continue
		}
		missing = append(missing, method)
	}
	return missing
}

func ensureImports(content string, imports []fileImport) string {
	for _, imp := range imports {
		quoted := strconv.Quote(imp.Path)
		if strings.Contains(content, quoted) {
			continue
		}
		block := "import (\n"
		insert := "\t" + quoted + "\n"
		if idx := strings.Index(content, block); idx >= 0 {
			content = content[:idx+len(block)] + insert + content[idx+len(block):]
		}
	}
	return content
}

func appendGoFragments(filePath string, imports []fileImport, fragments []string) error {
	content, err := util.ReadFile(filePath)
	if err != nil {
		return err
	}
	updated := ensureImports(content, imports)
	updated = strings.TrimRight(updated, "\n") + "\n\n" + strings.Join(fragments, "\n\n") + "\n"
	formatted, err := util.FormatGoFile(updated)
	if err != nil {
		return fmt.Errorf("格式化 %s 失败: %w", filepath.Base(filePath), err)
	}
	return util.WriteFile(filePath, formatted)
}
```

在 `internal/generator/handler.go` 中加入 `EnsureHandlerFile`：

```go
func EnsureHandlerFile(service types.ServiceInfo, config types.Config) (bool, error) {
	filePath := filepath.Join(config.OutputDir, service.FileName)
	if !util.FileExists(filePath) {
		return true, WriteHandlerFile(service, config, true)
	}
	if err := completeHandlerFile(filePath, service, config); err != nil {
		return false, err
	}
	return false, nil
}

func completeHandlerFile(filePath string, service types.ServiceInfo, config types.Config) error {
	content, err := util.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取 handler 文件失败: %w", err)
	}

	existing, err := collectExistingMethods(content, service.HandlerName)
	if err != nil {
		return err
	}

	missing := missingMethods(service.Methods, existing)
	if len(missing) == 0 {
		return nil
	}

	fragments := make([]string, 0, len(missing))
	for _, method := range missing {
		code, err := executeHandlerMethodTemplate(service, method)
		if err != nil {
			return fmt.Errorf("生成 handler 方法失败: %w", err)
		}
		fragments = append(fragments, code)
	}

	return appendGoFragments(filePath, protoImports(config.ModulePath, missing), fragments)
}

func executeHandlerMethodTemplate(service types.ServiceInfo, method types.Method) (string, error) {
	data := map[string]any{
		"Service": service,
		"Method":  method,
	}
	return ExecuteTemplate(HandlerMethodTemplate, data)
}
```

把 `generateHandlerCode` 的模板数据补成：

```go
type templateData struct {
	types.ServiceInfo
	types.Config
	UsesBasic bool
	UsesZebra bool
}

data := templateData{
	ServiceInfo: service,
	Config:      config,
	UsesBasic:   usesPkg(service.Methods, "basic"),
	UsesZebra:   usesPkg(service.Methods, "zebra"),
}
```

- [ ] **Step 4: 再跑 handler generator 测试，确认转绿**

Run: `go test ./internal/generator -run "TestEnsureHandlerFile(CreateFullFileWhenMissing|AppendsMissingMethodsWithoutOverwritingExistingBody|ReturnsErrorWhenHandlerTypeMissing)" -v`

Expected: PASS

- [ ] **Step 5: 记录这一小步**

在这个仓库里，任何 `git` 命令都必须先征得用户同意。先问用户是否要执行下面这组命令；只有在用户明确同意后再运行：

```bash
git add internal/generator/template.go internal/generator/handler.go internal/generator/go_file.go internal/generator/handler_test.go
git commit -m "feat: complete existing handler files"
```

### Task 4: 为 core 文件实现“创建或补全”

**Files:**
- Modify: `internal/generator/template.go`
- Modify: `internal/generator/core.go`
- Modify: `internal/generator/go_file.go`
- Create: `internal/generator/core_test.go`

- [ ] **Step 1: 写失败测试，锁定 core 创建与补全语义**

在 `internal/generator/core_test.go` 中加入：

```go
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestEnsureCoreServiceFileCreatesFullFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated core file: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "type ProjectService struct") {
		t.Fatalf("expected service struct in generated file, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectService) List(") {
		t.Fatalf("expected List method in generated file, got %q", got)
	}
}

func TestEnsureCoreServiceFileAppendsMissingMethodsWithoutOverwritingExistingBody(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	content := `package core

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"bsi/axis/devopsx/internal/micro/client"
	"bsi/axis/devopsx/internal/proto/axis/devopsx"
	mgorm "bsi/kratos/micro/gorm"
)

type ProjectService struct {
	srvClient        *client.Client
	log              *log.Helper
	transactionScope *mgorm.TransactionScope
}

func (p *ProjectService) List(ctx context.Context, in *devopsx.ListProjectRequest) (*devopsx.ListProjectReply, error) {
	return &devopsx.ListProjectReply{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing core file: %v", err)
	}

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated core file: %v", err)
	}

	got := string(updated)
	if strings.Count(got, "func (p *ProjectService) List(") != 1 {
		t.Fatalf("expected existing List method to stay single, got %q", got)
	}
	if !strings.Contains(got, "return &devopsx.ListProjectReply{}, nil") {
		t.Fatalf("expected existing List body to be preserved, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectService) Ping(") {
		t.Fatalf("expected missing Ping method to be appended, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected missing basic import to be added, got %q", got)
	}
}

func TestEnsureCoreServiceFileReturnsErrorWhenServiceTypeMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	if err := os.WriteFile(path, []byte("package core\n"), 0o644); err != nil {
		t.Fatalf("write invalid core file: %v", err)
	}

	_, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err == nil || !strings.Contains(err.Error(), "ProjectService") {
		t.Fatalf("expected missing type error, got %v", err)
	}
}
```

- [ ] **Step 2: 运行 core generator 测试，确认它先红**

Run: `go test ./internal/generator -run "TestEnsureCoreServiceFile(CreateFullFileWhenMissing|AppendsMissingMethodsWithoutOverwritingExistingBody|ReturnsErrorWhenServiceTypeMissing)" -v`

Expected: FAIL，原因是 `EnsureCoreServiceFile` 和 `CoreMethodTemplate` 还不存在。

- [ ] **Step 3: 写最小实现，让 core 也复用共享 helper**

在 `internal/generator/template.go` 中新增 `CoreMethodTemplate`，并让 `CoreServiceTemplate` 也按方法包名决定 import：

```go
const CoreServiceTemplate = `package core

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"{{.ModulePath}}/internal/micro/client"
	"{{.ModulePath}}/internal/proto/axis/devopsx"{{if .UsesBasic}}
	"{{.ModulePath}}/internal/proto/basic"{{end}}{{if .UsesZebra}}
	"{{.ModulePath}}/internal/proto/zebra"{{end}}
	mgorm "bsi/kratos/micro/gorm"
)

type {{.ServiceName}} struct {
	srvClient        *client.Client
	log              *log.Helper
	bs               *devopsx.Bootstrap
	transactionScope *mgorm.TransactionScope
}

//nolint:revive
func New{{.ServiceName}}(
	srvClient *client.Client,
	logger log.Logger,
	bs *devopsx.Bootstrap,
	transactionScope *mgorm.TransactionScope,
) *{{.ServiceName}} {
	return &{{.ServiceName}}{
		srvClient:        srvClient,
		log:              log.NewHelper(log.With(logger, "module", "{{.FieldName | trimSrv}}")),
		bs:               bs,
		transactionScope: transactionScope,
	}
}
{{range .Methods}}
{{if .Comment}}// {{.Comment}}{{else}}// {{.Name}}{{end}}
func ({{$.FieldName | firstChar}} *{{$.ServiceName}}) {{.Name}}(ctx context.Context, in *{{.RequestPkg}}.{{.RequestType}}) (*{{.ResponsePkg}}.{{.ResponseType}}, error) {
	{{$.FieldName | firstChar}}.log.Debug("not implement")
	{{if eq .ResponseType "Empty"}}return &{{.ResponsePkg}}.Empty{}, nil{{else}}return nil, nil{{end}}
}
{{end}}
`

const CoreMethodTemplate = `{{if .Method.Comment}}// {{.Method.Comment}}{{else}}// {{.Method.Name}}{{end}}
func ({{.Service.FieldName | firstChar}} *{{.Service.ServiceName}}) {{.Method.Name}}(ctx context.Context, in *{{.Method.RequestPkg}}.{{.Method.RequestType}}) (*{{.Method.ResponsePkg}}.{{.Method.ResponseType}}, error) {
	{{.Service.FieldName | firstChar}}.log.Debug("not implement")
	{{if eq .Method.ResponseType "Empty"}}return &{{.Method.ResponsePkg}}.Empty{}, nil{{else}}return nil, nil{{end}}
}
`
```

在 `internal/generator/core.go` 中加入：

```go
func EnsureCoreServiceFile(service types.ServiceInfo, config types.Config) (bool, error) {
	baseName := strings.TrimSuffix(service.ServiceName, "Service")
	fileName := util.CamelToSnake(baseName) + ".go"
	filePath := filepath.Join(config.CoreDir, fileName)

	if !util.FileExists(filePath) {
		return true, GenerateCoreServiceFile(service, config, true)
	}
	if err := completeCoreServiceFile(filePath, service, config); err != nil {
		return false, err
	}
	return false, nil
}

func completeCoreServiceFile(filePath string, service types.ServiceInfo, config types.Config) error {
	content, err := util.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取 core 文件失败: %w", err)
	}

	existing, err := collectExistingMethods(content, service.ServiceName)
	if err != nil {
		return err
	}

	missing := missingMethods(service.Methods, existing)
	if len(missing) == 0 {
		return nil
	}

	fragments := make([]string, 0, len(missing))
	for _, method := range missing {
		code, err := executeCoreMethodTemplate(service, method)
		if err != nil {
			return fmt.Errorf("生成 core 方法失败: %w", err)
		}
		fragments = append(fragments, code)
	}

	return appendGoFragments(filePath, protoImports(config.ModulePath, missing), fragments)
}

func executeCoreMethodTemplate(service types.ServiceInfo, method types.Method) (string, error) {
	data := map[string]any{
		"Service": service,
		"Method":  method,
	}
	return ExecuteTemplate(CoreMethodTemplate, data)
}
```

把 `generateCoreServiceCode` 的模板数据补成和 handler 一样的 `UsesBasic` / `UsesZebra` 布尔值。

- [ ] **Step 4: 再跑 core generator 测试，确认转绿**

Run: `go test ./internal/generator -run "TestEnsureCoreServiceFile(CreateFullFileWhenMissing|AppendsMissingMethodsWithoutOverwritingExistingBody|ReturnsErrorWhenServiceTypeMissing)" -v`

Expected: PASS

- [ ] **Step 5: 记录这一小步**

在这个仓库里，任何 `git` 命令都必须先征得用户同意。先问用户是否要执行下面这组命令；只有在用户明确同意后再运行：

```bash
git add internal/generator/template.go internal/generator/core.go internal/generator/go_file.go internal/generator/core_test.go
git commit -m "feat: complete existing core service files"
```

### Task 5: 做回归验证并收口

**Files:**
- Modify: `cmd/handler.go`
- Modify: `internal/workflow/handler.go`
- Modify: `internal/generator/handler.go`
- Modify: `internal/generator/core.go`
- Test: `cmd/handler_test.go`
- Test: `internal/workflow/handler_test.go`
- Test: `internal/generator/handler_test.go`
- Test: `internal/generator/core_test.go`

- [ ] **Step 1: 跑定向测试，确认命令层、workflow、generator 都覆盖到**

Run: `go test ./cmd -run TestRunHandlerCommandUsesAllParsedServicesInSession -v`

Expected: PASS

Run: `go test ./internal/workflow -run TestHandlerWorkflowRunItemBranchesByFileCreation -v`

Expected: PASS

Run: `go test ./internal/generator -run "TestEnsure(HandlerFile|CoreServiceFile)" -v`

Expected: PASS

- [ ] **Step 2: 跑全量单测，确认没有回归**

Run: `go test ./...`

Expected: PASS

- [ ] **Step 3: 跑构建验证，确认 CLI 仍可编译**

Run: `go build .`

Expected: PASS

- [ ] **Step 4: 检查关键行为是否与 spec 一致**

人工核对以下 4 点：

```text
1. handler 命令默认列出全部服务
2. 已有 grpc/core 文件只补缺失方法，不覆盖已有实现
3. 只有本次创建了 grpc 文件才更新 api/grpc/grpc.go
4. 只有本次创建了 core 文件才更新 core/core.go；只要创建了任意文件就执行 wire
```

- [ ] **Step 5: 如需提交，先征得用户许可**

在这个仓库里，任何 `git` 命令都必须先征得用户同意。先问用户是否要执行下面这组命令；只有在用户明确同意后再运行：

```bash
git add cmd/handler.go cmd/handler_test.go cmd/session.go internal/workflow/handler.go internal/workflow/handler_test.go internal/generator/template.go internal/generator/go_file.go internal/generator/handler.go internal/generator/core.go internal/generator/handler_test.go internal/generator/core_test.go
git commit -m "feat: complete existing handler and core files"
```
