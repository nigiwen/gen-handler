# Unified Bubble Tea UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `handler` 和 `data` 命令落地统一的 Bubble Tea + Bubbles 终端 UI，支持多选、`/` 搜索、执行进度和结果汇总，并移除旧的 selector 入口。

**Architecture:** 新增 `internal/tui/` 负责统一状态机和界面渲染，新增 `internal/workflow/` 负责命令级候选项发现与逐项执行。`cmd/handler.go` 和 `cmd/data.go` 只保留参数解析与调用 workflow/UI 的职责；底层 generator/updater/scanner 继续提供安静的业务能力，通过 workflow 发出结构化进度事件给 UI。

**Tech Stack:** Go、Bubble Tea、Bubbles、Lip Gloss、标准库测试、现有 scanner/generator/updater

---

## File Map

- Create: `internal/tui/types.go`
  - 定义 `Item`、`ProgressEvent`、`RunResult`、`SessionConfig`
- Create: `internal/tui/keys.go`
  - 定义选择页、运行页、结果页的键位映射
- Create: `internal/tui/styles.go`
  - 定义统一样式
- Create: `internal/tui/model.go`
  - 统一 `tea.Model`、状态切换、命令派发
- Create: `internal/tui/select_view.go`
  - 多选、搜索、过滤、列表渲染
- Create: `internal/tui/run_view.go`
  - spinner/progress/最近结果渲染
- Create: `internal/tui/summary_view.go`
  - summary 视图渲染
- Create: `internal/tui/fallback.go`
  - 非 TTY 环境的编号输入 fallback
- Create: `internal/tui/model_test.go`
  - 选择/搜索/软停止/汇总行为测试
- Create: `internal/workflow/types.go`
  - workflow 通用接口和结果类型
- Create: `internal/workflow/data.go`
  - Data Sync 的候选发现与逐项执行
- Create: `internal/workflow/data_test.go`
  - Data workflow 进度事件与错误行为测试
- Create: `internal/workflow/handler.go`
  - Handler Generate 的候选发现与逐项执行
- Create: `internal/workflow/handler_test.go`
  - Handler workflow 进度事件与错误行为测试
- Modify: `internal/generator/handler.go`
  - 拆出安静的 handler/core/provider/wire 步骤函数，去掉直接 `fmt.Printf`
- Modify: `cmd/data.go`
  - 从直接执行改为调用 workflow + TUI session
- Modify: `cmd/handler.go`
  - 从旧 selector 改为调用 workflow + TUI session，并支持多选批处理
- Delete: `internal/selector/selector.go`
  - 旧交互入口删除
- Modify: `go.mod`
  - 添加 Bubble Tea / Bubbles / Lip Gloss 依赖
- Modify: `README.md`
  - 更新统一 UI 和操作说明
- Modify: `main.go`
  - 更新 `handler` / `data` 的帮助文案

### Task 1: 建立统一选择模型与基础依赖

**Files:**
- Modify: `go.mod`
- Create: `internal/tui/types.go`
- Create: `internal/tui/keys.go`
- Create: `internal/tui/styles.go`
- Create: `internal/tui/model.go`
- Create: `internal/tui/select_view.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: 写选择页失败测试**

```go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectViewSpaceTogglesSelectionAndEnterReturnsCheckedItems(t *testing.T) {
	items := []Item{
		{ID: "field_module", Title: "field_module", Description: "Entity: FieldModule"},
		{ID: "process_node", Title: "process_node", Description: "Entity: ProcessNode"},
	}

	model := NewModel(SessionConfig{
		Title: "Data Sync",
		Items: items,
	})

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

func TestSelectViewEnterDefaultsToCurrentItemWhenNothingChecked(t *testing.T) {
	items := []Item{
		{ID: "test_case", Title: "test_case.go", Description: "Handler: TestCaseHandler, 12 个方法"},
		{ID: "project", Title: "project.go", Description: "Handler: ProjectHandler, 8 个方法"},
	}

	model := NewModel(SessionConfig{
		Title: "Handler Generate",
		Items: items,
	})

	model.cursor = 1
	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = modelIface.(Model)

	if len(model.runQueue) != 1 || model.runQueue[0].ID != "project" {
		t.Fatalf("expected current item to be queued, got %+v", model.runQueue)
	}
}
```

- [ ] **Step 2: 运行测试，确认当前失败**

Run: `go test ./internal/tui -run "TestSelectView(SpaceTogglesSelectionAndEnterReturnsCheckedItems|EnterDefaultsToCurrentItemWhenNothingChecked)" -v`

Expected: FAIL，原因是 `internal/tui` 包和 `NewModel` 尚不存在。

- [ ] **Step 3: 写最小实现并加依赖**

在 `go.mod` 中加入：

```go
require (
	github.com/charmbracelet/bubbles v0.20.0
	github.com/charmbracelet/bubbletea v1.3.4
	github.com/charmbracelet/lipgloss v1.1.0
)
```

在 `internal/tui/types.go` 中先加入最小定义：

```go
package tui

type Item struct {
	ID          string
	Title       string
	Description string
	Keywords    []string
	Payload     any
}

type SessionConfig struct {
	Title string
	Items []Item
}
```

在 `internal/tui/model.go` 中加入：

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateSelecting state = iota
	stateSearching
	stateRunning
	stateSummary
)

type Model struct {
	state     state
	title     string
	allItems  []Item
	visible   []Item
	selected  map[string]bool
	cursor    int
	runQueue  []Item
}

func NewModel(cfg SessionConfig) Model {
	return Model{
		state:    stateSelecting,
		title:    cfg.Title,
		allItems: cfg.Items,
		visible:  append([]Item(nil), cfg.Items...),
		selected: make(map[string]bool),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateSelecting:
			return m.updateSelecting(msg)
		}
	}
	return m, nil
}

func (m Model) updateSelecting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
	case tea.KeySpace:
		if len(m.visible) > 0 {
			id := m.visible[m.cursor].ID
			m.selected[id] = !m.selected[id]
		}
	case tea.KeyEnter:
		m.runQueue = m.selectedItemsOrCurrent()
		m.state = stateRunning
	}
	return m, nil
}

func (m Model) selectedItemsOrCurrent() []Item {
	var items []Item
	for _, item := range m.visible {
		if m.selected[item.ID] {
			items = append(items, item)
		}
	}
	if len(items) == 0 && len(m.visible) > 0 {
		return []Item{m.visible[m.cursor]}
	}
	return items
}
```

在 `internal/tui/select_view.go` 中加入：

```go
package tui

func (m Model) View() string {
	return renderSelectView(m)
}

func renderSelectView(m Model) string {
	return m.title
}
```

- [ ] **Step 4: 运行测试，确认转绿**

Run: `go test ./internal/tui -run "TestSelectView(SpaceTogglesSelectionAndEnterReturnsCheckedItems|EnterDefaultsToCurrentItemWhenNothingChecked)" -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add go.mod go.sum internal/tui/types.go internal/tui/keys.go internal/tui/styles.go internal/tui/model.go internal/tui/select_view.go internal/tui/model_test.go
git commit -m "feat: add unified tui selection model"
```

### Task 2: 实现 `/` 搜索、多选列表渲染和帮助栏

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/select_view.go`
- Modify: `internal/tui/keys.go`
- Modify: `internal/tui/styles.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: 写失败的搜索与全选测试**

```go
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
```

- [ ] **Step 2: 运行测试，确认当前失败**

Run: `go test ./internal/tui -run "TestSelectView(SlashEntersSearchAndFiltersVisibleItems|KeyASelectsAllVisibleItems)" -v`

Expected: FAIL，原因是搜索状态、`searchChangedMsg` 和 `a` 全选逻辑尚不存在。

- [ ] **Step 3: 写最小实现**

在 `internal/tui/model.go` 中扩展：

```go
import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type searchChangedMsg string

type Model struct {
	state       state
	title       string
	allItems    []Item
	visible     []Item
	selected    map[string]bool
	cursor      int
	runQueue    []Item
	searchInput textinput.Model
	filterQuery string
}

func NewModel(cfg SessionConfig) Model {
	input := textinput.New()
	input.Prompt = "/ "
	return Model{
		state:       stateSelecting,
		title:       cfg.Title,
		allItems:    cfg.Items,
		visible:     append([]Item(nil), cfg.Items...),
		selected:    make(map[string]bool),
		searchInput: input,
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchChangedMsg:
		m.filterQuery = string(msg)
		m.applyFilter()
		return m, nil
	case tea.KeyMsg:
		switch m.state {
		case stateSelecting:
			return m.updateSelecting(msg)
		case stateSearching:
			return m.updateSearching(msg)
		}
	}
	return m, nil
}

func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.filterQuery))
	if query == "" {
		m.visible = append([]Item(nil), m.allItems...)
		m.cursor = 0
		return
	}

	var filtered []Item
	for _, item := range m.allItems {
		blob := strings.ToLower(item.Title + " " + item.Description + " " + strings.Join(item.Keywords, " "))
		if strings.Contains(blob, query) {
			filtered = append(filtered, item)
		}
	}
	m.visible = filtered
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
}

func (m Model) updateSelecting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '/':
		m.state = stateSearching
		return m, textinput.Blink
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'a':
		allSelected := true
		for _, item := range m.visible {
			if !m.selected[item.ID] {
				allSelected = false
				break
			}
		}
		for _, item := range m.visible {
			m.selected[item.ID] = !allSelected
		}
		return m, nil
	}
	// 保留 Task 1 中已有逻辑
	return m.updateSelectingKeys(msg)
}

func (m Model) updateSearching(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchInput.SetValue("")
		m.filterQuery = ""
		m.applyFilter()
		m.state = stateSelecting
		return m, nil
	case tea.KeyEnter:
		m.state = stateSelecting
		return m, nil
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.filterQuery = m.searchInput.Value()
	m.applyFilter()
	return m, cmd
}
```

在 `internal/tui/select_view.go` 中渲染：

```go
import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
)

func renderSelectView(m Model) string {
	var rows []string
	rows = append(rows, titleStyle.Render(m.title))
	rows = append(rows, subtitleStyle.Render(fmt.Sprintf("Filter: %s", displayFilter(m))))
	rows = append(rows, "")
	for i, item := range m.visible {
		prefix := "[ ]"
		if m.selected[item.ID] {
			prefix = "[x]"
		}
		line := fmt.Sprintf("%s %s", prefix, item.Title)
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		rows = append(rows, line)
		rows = append(rows, mutedStyle.Render("  "+item.Description))
	}
	rows = append(rows, "")
	rows = append(rows, mutedStyle.Render(fmt.Sprintf("Total: %d  Visible: %d  Selected: %d", len(m.allItems), len(m.visible), m.selectedCount())))
	rows = append(rows, help.New().View(selectKeys))
	return strings.Join(rows, "\n")
}
```

- [ ] **Step 4: 运行测试，确认转绿**

Run: `go test ./internal/tui -run "TestSelectView(SlashEntersSearchAndFiltersVisibleItems|KeyASelectsAllVisibleItems)" -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add internal/tui/model.go internal/tui/select_view.go internal/tui/keys.go internal/tui/styles.go internal/tui/model_test.go
git commit -m "feat: add search and multi-select tui interactions"
```

### Task 3: 实现运行页、结果页和软停止

**Files:**
- Modify: `internal/tui/types.go`
- Modify: `internal/tui/model.go`
- Create: `internal/tui/run_view.go`
- Create: `internal/tui/summary_view.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: 写失败的运行与汇总测试**

```go
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
```

- [ ] **Step 2: 运行测试，确认当前失败**

Run: `go test ./internal/tui -run "Test(RunViewProcessesProgressEventsAndMovesToSummary|RunViewQRequestsSoftStop)" -v`

Expected: FAIL，原因是 `ProgressEvent`、`RunResult`、运行状态和汇总状态尚未实现。

- [ ] **Step 3: 写最小实现**

在 `internal/tui/types.go` 中补充运行态类型，并把 `SessionConfig` 扩展为可携带 `Run`：

```go
package tui

type Status string

const (
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

type Runner func(item Item, emit func(ProgressEvent)) RunResult

type ProgressEvent struct {
	ItemID   string
	Step     string
	Status   Status
	Message  string
	Err      error
}

type RunResult struct {
	ItemID   string
	Title    string
	Success  bool
	Skipped  bool
	Err      error
}

type runItemFinishedMsg struct {
	Result RunResult
}

type runFinishedMsg struct{}

type SessionConfig struct {
	Title string
	Items []Item
	Run   Runner
}
```

在 `internal/tui/model.go` 中扩展：

```go
import "github.com/charmbracelet/bubbles/spinner"

type recentEvent struct {
	ItemID   string
	Step     string
	Status   Status
	Message  string
}

type Model struct {
	// ... 保留前面字段
	spinner          spinner.Model
	completed        int
	succeeded        int
	failed           int
	skipped          int
	currentItem      string
	currentStep      string
	recent           []recentEvent
	results          []RunResult
	stopAfterCurrent bool
}

func NewModel(cfg SessionConfig) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	// 其余字段保留
	return Model{spinner: sp}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressEvent:
		m.currentItem = msg.ItemID
		m.currentStep = msg.Step
		m.recent = append([]recentEvent{{
			ItemID:  msg.ItemID,
			Step:    msg.Step,
			Status:  msg.Status,
			Message: msg.Message,
		}}, m.recent...)
		if len(m.recent) > 3 {
			m.recent = m.recent[:3]
		}
		return m, nil
	case runItemFinishedMsg:
		m.completed++
		m.results = append(m.results, msg.Result)
		switch {
		case msg.Result.Skipped:
			m.skipped++
		case msg.Result.Success:
			m.succeeded++
		default:
			m.failed++
		}
		return m, nil
	case runFinishedMsg:
		m.state = stateSummary
		return m, nil
	case tea.KeyMsg:
		if m.state == stateRunning && len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
			m.stopAfterCurrent = true
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}
```

在 `internal/tui/run_view.go` 中加入：

```go
package tui

import "fmt"

func renderRunView(m Model) string {
	return fmt.Sprintf(
		"%s\n\n%s 正在处理 %d/%d\n当前项: %s\n当前步骤: %s\n成功: %d  失败: %d  跳过: %d",
		m.title,
		m.spinner.View(),
		m.completed+1,
		len(m.runQueue),
		m.currentItem,
		m.currentStep,
		m.succeeded,
		m.failed,
		m.skipped,
	)
}
```

在 `internal/tui/summary_view.go` 中加入：

```go
package tui

import "fmt"

func renderSummaryView(m Model) string {
	return fmt.Sprintf(
		"%s\n\nSelected: %d\nSuccess: %d\nFailed: %d\nStopped: %t\n\n按 Enter 或 q 退出",
		m.title,
		len(m.runQueue),
		m.succeeded,
		m.failed,
		m.stopAfterCurrent,
	)
}
```

- [ ] **Step 4: 运行测试，确认转绿**

Run: `go test ./internal/tui -run "Test(RunViewProcessesProgressEventsAndMovesToSummary|RunViewQRequestsSoftStop)" -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add internal/tui/types.go internal/tui/model.go internal/tui/run_view.go internal/tui/summary_view.go internal/tui/model_test.go
git commit -m "feat: add tui run and summary states"
```

### Task 4: 落地 Data workflow 并接入进度事件

**Files:**
- Create: `internal/workflow/types.go`
- Create: `internal/workflow/data.go`
- Create: `internal/workflow/data_test.go`
- Modify: `cmd/data.go`

- [ ] **Step 1: 写失败的 Data workflow 测试**

```go
package workflow

import (
	"path/filepath"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestDataWorkflowRunItemEmitsStepsInOrder(t *testing.T) {
	var events []ProgressEvent

	wf := DataWorkflow{
		Config: types.Config{WireDir: filepath.Join(t.TempDir(), "cmd", "demo")},
		EntityDir: filepath.Join(t.TempDir(), "internal", "model", "entity"),
		DataDir: filepath.Join(t.TempDir(), "data"),
		GenerateEntity: func(path string) error { return nil },
		GenerateRepo: func(path string, entity types.EntityInfo) error { return nil },
		UpdateProviderSet: func(filePath, newItem, itemType string) error { return nil },
		RunWire: func(wireDir string) error { return nil },
		FileExists: func(path string) bool { return false },
	}

	item := Item{
		ID: "field_module",
		Payload: types.EntityInfo{
			FileName: "field_module",
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
```

- [ ] **Step 2: 运行测试，确认当前失败**

Run: `go test ./internal/workflow -run TestDataWorkflowRunItemEmitsStepsInOrder -v`

Expected: FAIL，原因是 `internal/workflow` 包尚不存在。

- [ ] **Step 3: 写最小实现并接入 `cmd/data.go`**

在 `internal/workflow/types.go` 中加入：

```go
package workflow

import "github.com/nigiwen/gen-handler/internal/tui"

type ProgressEvent = tui.ProgressEvent
type RunResult = tui.RunResult
type Item = tui.Item
```

在 `internal/workflow/data.go` 中加入：

```go
package workflow

import (
	"path/filepath"

	"github.com/nigiwen/gen-handler/internal/types"
)

type DataWorkflow struct {
	Config            types.Config
	EntityDir         string
	DataDir           string
	GenerateEntity    func(path string) error
	GenerateRepo      func(path string, entity types.EntityInfo) error
	UpdateProviderSet func(filePath, newItem, itemType string) error
	RunWire           func(wireDir string) error
	FileExists        func(path string) bool
}

func (wf DataWorkflow) BuildItems(entities []types.EntityInfo) []Item {
	items := make([]Item, 0, len(entities))
	for _, entity := range entities {
		items = append(items, Item{
			ID:          entity.FileName,
			Title:       entity.FileName,
			Description: "Entity: " + entity.EntityName,
			Keywords:    []string{entity.EntityName, entity.FileName},
			Payload:     entity,
		})
	}
	return items
}

func (wf DataWorkflow) RunItem(item Item, emit func(ProgressEvent)) RunResult {
	entity := item.Payload.(types.EntityInfo)
	entityPath := filepath.Join(wf.EntityDir, entity.FileName+".go")
	repoPath := filepath.Join(wf.DataDir, entity.FileName+".go")

	emit(ProgressEvent{ItemID: item.ID, Step: "检查 entity 占位文件"})
	if !wf.FileExists(entityPath) {
		emit(ProgressEvent{ItemID: item.ID, Step: "生成 entity 占位文件"})
		if err := wf.GenerateEntity(entityPath); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "生成 repo"})
	if !wf.FileExists(repoPath) {
		if err := wf.GenerateRepo(repoPath, entity); err != nil {
			return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
		}
	} else {
		return RunResult{ItemID: item.ID, Title: item.Title, Success: true, Skipped: true}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "更新 ProviderSet"})
	if err := wf.UpdateProviderSet(filepath.Join(wf.DataDir, "data.go"), "New"+entity.EntityName+"Repo", "Repo"); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "运行 wire"})
	if err := wf.RunWire(wf.Config.WireDir); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	return RunResult{ItemID: item.ID, Title: item.Title, Success: true}
}
```

在 `cmd/data.go` 中改为：

```go
wf := workflow.DataWorkflow{
	Config:            config,
	EntityDir:         entityDir,
	DataDir:           dataDir,
	GenerateEntity:    generator.GenerateEntityFile,
	GenerateRepo:      generator.GenerateRepoFile,
	UpdateProviderSet: updater.UpdateProviderSet,
	RunWire:           generator.RunWireCommand,
	FileExists:        util.FileExists,
}

items := wf.BuildItems(missingEntities)
if err := tui.RunSession(tui.SessionConfig{
	Title: "Data Sync",
	Items: items,
	Run:   wf.RunItem,
}); err != nil {
	fmt.Printf("❌ Data Sync 失败: %v\n", err)
}
```

- [ ] **Step 4: 运行测试，确认转绿**

Run: `go test ./internal/workflow -run TestDataWorkflowRunItemEmitsStepsInOrder -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add internal/workflow/types.go internal/workflow/data.go internal/workflow/data_test.go cmd/data.go
git commit -m "feat: add data workflow for unified tui"
```

### Task 5: 落地 Handler workflow，并把 generator 输出改成安静步骤

**Files:**
- Create: `internal/workflow/handler.go`
- Create: `internal/workflow/handler_test.go`
- Modify: `internal/generator/handler.go`
- Modify: `cmd/handler.go`

- [ ] **Step 1: 写失败的 Handler workflow 测试**

```go
package workflow

import (
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestHandlerWorkflowRunItemEmitsStepsInOrder(t *testing.T) {
	var events []ProgressEvent

	wf := HandlerWorkflow{
		Config: types.Config{},
		RunGenerateHandler: func(service types.ServiceInfo, config types.Config) error { return nil },
		RunUpdateGrpc: func(service types.ServiceInfo, outputDir string) error { return nil },
		RunGenerateCore: func(service types.ServiceInfo, config types.Config) error { return nil },
		RunUpdateCoreProvider: func(service types.ServiceInfo, coreDir string) error { return nil },
		RunWire: func(wireDir string) error { return nil },
	}

	item := Item{
		ID: "test_case.go",
		Title: "test_case.go",
		Payload: types.ServiceInfo{
			FileName: "test_case.go",
			HandlerName: "TestCaseHandler",
			ServiceName: "TestCaseService",
		},
	}

	result := wf.RunItem(item, func(ev ProgressEvent) {
		events = append(events, ev)
	})

	if !result.Success {
		t.Fatalf("expected success result, got %+v", result)
	}

	got := []string{events[0].Step, events[1].Step, events[2].Step, events[3].Step, events[4].Step}
	want := []string{"生成 handler 文件", "更新 grpc.go", "生成 core service", "更新 core ProviderSet", "运行 wire"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected step order: got %v want %v", got, want)
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认当前失败**

Run: `go test ./internal/workflow -run TestHandlerWorkflowRunItemEmitsStepsInOrder -v`

Expected: FAIL，原因是 `HandlerWorkflow` 尚不存在，且 generator 仍然直接打印。

- [ ] **Step 3: 写最小实现**

在 `internal/generator/handler.go` 中拆出安静函数：

```go
func WriteHandlerFile(service types.ServiceInfo, config types.Config, force bool) error {
	if !force {
		filePath := filepath.Join(config.OutputDir, service.FileName)
		if util.FileExists(filePath) {
			return fmt.Errorf("文件已存在")
		}
	}
	if err := os.MkdirAll(config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	code, err := generateHandlerCode(service, config)
	if err != nil {
		return fmt.Errorf("生成代码失败: %v", err)
	}
	filePath := filepath.Join(config.OutputDir, service.FileName)
	if err := util.WriteFile(filePath, code); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	return nil
}

func UpdateGrpcProvider(service types.ServiceInfo, outputDir string) error {
	grpcFilePath := filepath.Join(outputDir, "grpc.go")
	return updater.UpdateGrpcFile(service, grpcFilePath)
}

func WriteCoreService(service types.ServiceInfo, config types.Config, force bool) error {
	return GenerateCoreServiceFile(service, config, force)
}

func UpdateCoreProvider(service types.ServiceInfo, coreDir string) error {
	coreGoPath := filepath.Join(coreDir, "core.go")
	return updater.UpdateProviderSet(coreGoPath, "New"+service.ServiceName, "Service")
}
```

在 `internal/workflow/handler.go` 中加入：

```go
package workflow

import "github.com/nigiwen/gen-handler/internal/types"

type HandlerWorkflow struct {
	Config                types.Config
	RunGenerateHandler    func(service types.ServiceInfo, config types.Config, force bool) error
	RunUpdateGrpc         func(service types.ServiceInfo, outputDir string) error
	RunGenerateCore       func(service types.ServiceInfo, config types.Config, force bool) error
	RunUpdateCoreProvider func(service types.ServiceInfo, coreDir string) error
	RunWire               func(wireDir string) error
}

func (wf HandlerWorkflow) BuildItems(services []types.ServiceInfo) []Item {
	items := make([]Item, 0, len(services))
	for _, service := range services {
		items = append(items, Item{
			ID:          service.FileName,
			Title:       service.FileName,
			Description: "Handler: " + service.HandlerName,
			Keywords:    []string{service.HandlerName, service.ServiceName, service.FileName},
			Payload:     service,
		})
	}
	return items
}

func (wf HandlerWorkflow) RunItem(item Item, emit func(ProgressEvent)) RunResult {
	service := item.Payload.(types.ServiceInfo)

	emit(ProgressEvent{ItemID: item.ID, Step: "生成 handler 文件"})
	if err := wf.RunGenerateHandler(service, wf.Config, true); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "更新 grpc.go"})
	if err := wf.RunUpdateGrpc(service, wf.Config.OutputDir); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "生成 core service"})
	if err := wf.RunGenerateCore(service, wf.Config, true); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "更新 core ProviderSet"})
	if err := wf.RunUpdateCoreProvider(service, wf.Config.CoreDir); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	emit(ProgressEvent{ItemID: item.ID, Step: "运行 wire"})
	if err := wf.RunWire(wf.Config.WireDir); err != nil {
		return RunResult{ItemID: item.ID, Title: item.Title, Err: err}
	}

	return RunResult{ItemID: item.ID, Title: item.Title, Success: true}
}
```

- [ ] **Step 4: 运行测试，确认转绿**

Run: `go test ./internal/workflow -run TestHandlerWorkflowRunItemEmitsStepsInOrder -v`

Expected: PASS

- [ ] **Step 5: 提交本任务**

```bash
git add internal/generator/handler.go internal/workflow/handler.go internal/workflow/handler_test.go cmd/handler.go
git commit -m "feat: add handler workflow for unified tui"
```

### Task 6: 集成统一 UI、fallback、删除旧 selector，并完成文档与验证

**Files:**
- Create: `internal/tui/fallback.go`
- Modify: `internal/tui/model.go`
- Modify: `cmd/data.go`
- Modify: `cmd/handler.go`
- Delete: `internal/selector/selector.go`
- Modify: `README.md`
- Modify: `main.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: 写失败的 fallback 与集成测试**

```go
func TestRunViewStopsAfterCurrentWhenQPressed(t *testing.T) {
	model := NewModel(SessionConfig{
		Title: "Data Sync",
		Items: []Item{
			{ID: "field_module", Title: "field_module"},
			{ID: "process_node", Title: "process_node"},
		},
	})

	model.runQueue = []Item{
		{ID: "field_module", Title: "field_module"},
		{ID: "process_node", Title: "process_node"},
	}
	model.state = stateRunning

	modelIface, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = modelIface.(Model)
	if !model.stopAfterCurrent {
		t.Fatalf("expected stopAfterCurrent to be true")
	}
}
```

新增 `internal/tui/fallback.go` 的测试：

```go
func TestFallbackDefaultsToCurrentItemWhenInputEmpty(t *testing.T) {
	items := []Item{
		{ID: "field_module", Title: "field_module"},
		{ID: "process_node", Title: "process_node"},
	}

	selected := fallbackSelect(items, "2")
	if len(selected) != 1 || selected[0].ID != "process_node" {
		t.Fatalf("unexpected fallback selection: %+v", selected)
	}
}
```

- [ ] **Step 2: 运行测试，确认当前失败**

Run: `go test ./internal/tui -run "Test(RunViewStopsAfterCurrentWhenQPressed|FallbackDefaultsToCurrentItemWhenInputEmpty)" -v`

Expected: FAIL，原因是 fallback 和 stop-after-current 的最终集成逻辑尚未完成。

- [ ] **Step 3: 写最小实现并删除旧 selector**

在 `internal/tui/fallback.go` 中加入：

```go
package tui

import (
	"strconv"
	"strings"
)

func fallbackSelect(items []Item, input string) []Item {
	input = strings.TrimSpace(input)
	if input == "" && len(items) > 0 {
		return []Item{items[0]}
	}
	if strings.EqualFold(input, "all") {
		return items
	}
	var selected []Item
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil || n <= 0 || n > len(items) {
			continue
		}
		selected = append(selected, items[n-1])
	}
	return selected
}
```

在 `internal/tui/model.go` 中补 `RunSession`：

```go
func RunSession(cfg SessionConfig) error {
	program := tea.NewProgram(NewModel(cfg), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
```

在 `cmd/data.go` 中改为：

```go
wf := workflow.DataWorkflow{
	Config:            config,
	EntityDir:         entityDir,
	DataDir:           dataDir,
	GenerateEntity:    generator.GenerateEntityFile,
	GenerateRepo:      generator.GenerateRepoFile,
	UpdateProviderSet: updater.UpdateProviderSet,
	RunWire:           generator.RunWireCommand,
	FileExists:        util.FileExists,
}

items := wf.BuildItems(missingEntities)
if err := tui.RunSession(tui.SessionConfig{
	Title: "Data Sync",
	Items: items,
	Run:   wf.RunItem,
}); err != nil {
	fmt.Printf("❌ Data Sync 失败: %v\n", err)
}
```

在 `cmd/handler.go` 中改为：

```go
wf := workflow.HandlerWorkflow{
	Config:                config,
	RunGenerateHandler:    generator.WriteHandlerFile,
	RunUpdateGrpc:         generator.UpdateGrpcProvider,
	RunGenerateCore:       generator.WriteCoreService,
	RunUpdateCoreProvider: generator.UpdateCoreProvider,
	RunWire:               generator.RunWireCommand,
}

items := wf.BuildItems(missingServices)
if err := tui.RunSession(tui.SessionConfig{
	Title: "Handler Generate",
	Items: items,
	Run:   wf.RunItem,
}); err != nil {
	fmt.Printf("❌ Handler Generate 失败: %v\n", err)
}
```

在 `main.go` 中帮助文案改为：

```go
fmt.Fprintf(os.Stderr, "  data       同步 Data 层 (*.gen.go -> entity & repo)\n\n")
fmt.Fprintf(os.Stderr, "  handler    生成 gRPC Handler / Core / wire\n")
```

删除文件：

```text
internal/selector/selector.go
```

更新 `README.md` 加入统一操作说明：

```md
### 统一交互

- `↑/↓` 或 `j/k`：移动
- `Space`：勾选/取消
- `/`：搜索
- `a`：全选当前可见项
- `Enter`：执行
- `q`：退出或在运行页请求软停止
```

- [ ] **Step 4: 运行全量验证**

Run: `go test ./...`

Expected: PASS

Run: `go build .`

Expected: PASS

- [ ] **Step 5: 提交最终集成**

```bash
git add go.mod go.sum main.go cmd/data.go cmd/handler.go internal/tui internal/workflow README.md
git rm internal/selector/selector.go
git commit -m "feat: add unified bubble tea ui"
```
