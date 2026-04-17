# 基于 Bubble Tea + Bubbles 的统一 CLI UI 设计

**日期：** 2026-04-17

## 目标

为当前仓库的 `handler` 和 `data` 命令提供一套统一的终端交互界面，使用 `Bubble Tea + Bubbles` 实现多选、搜索、执行进度和结果汇总，替换现有的简易选择器与命令行打印输出。

本次设计明确以下交互方向：

- `handler` 与 `data` 共用一套 UI 流程
- 选择页支持多选
- 使用 `Space` 勾选/取消
- 使用 `/` 进入搜索模式
- 执行阶段展示 spinner 与 progress，避免“卡住感”
- 结果页统一展示成功与失败项

## 当前问题

当前交互存在以下问题：

1. `handler` 和 `data` 虽然共用 `internal/selector/selector.go`，但整体体验不统一。
2. 选择器依赖当前终端是否支持 raw mode；一旦回退到非终端模式，就只能使用编号输入。
3. 当前终端交互模式只支持单选。
4. `data` 和 `handler` 在执行阶段大量直接 `fmt.Printf`，运行 `wire` 或批量处理时容易给人“终端卡住”的感觉。
5. `handler` 当前即便未来支持多选，执行逻辑也仍然只处理第一个选项。

## 设计原则

- 保持命令语义不变：`handler` 仍然生成 handler/core/wire，`data` 仍然同步 entity/repo/provider/wire
- 统一体验而不是拼接两套交互
- UI 负责展示，workflow 负责执行，底层 generator/scanner/updater 尽量保持“安静”
- 长列表下必须可用，不能只依赖上下键逐行翻找
- 对非 TTY 环境保留 fallback
- 第一版只做必要能力，不引入并发执行、鼠标支持、主题切换等额外复杂度

## 统一流程

两条命令统一为如下三阶段流程：

1. `Select View`
2. `Run View`
3. `Summary View`

命令层在完成参数解析和候选项发现后，将候选项和执行函数交给统一 UI 会话，UI 负责后续交互和展示。

## 页面结构

### 1. Select View

用于候选项浏览、搜索与多选。

页面结构：

- 标题区
  - 命令标题：`Handler Generate` 或 `Data Sync`
  - 副标题：当前命令的关键路径信息
- 搜索区
  - 显示当前过滤词
  - 搜索模式下显示可编辑输入框
- 列表区
  - 多选列表
  - 每项两行展示：标题 + 描述
- 状态区
  - `Total / Visible / Selected`
- 帮助区
  - 键位说明

显示规则：

- 当前光标行高亮
- 已选项前缀使用 `[x]`
- 未选项前缀使用 `[ ]`
- 搜索无结果时显示空状态提示

### 2. Run View

用于展示批处理执行进度。

页面结构：

- 标题区
  - 命令标题
  - 当前进度：例如 `Running 3/8`
- 主状态区
  - spinner
  - 当前项名称
  - 当前步骤名称
- 进度区
  - progress bar
  - success / failed / skipped 统计
- 最近结果区
  - 最近几条完成记录
- 提示区
  - 显示软停止提示

### 3. Summary View

用于展示批处理结果汇总。

页面结构：

- 汇总标题
- 统计区
  - Selected / Success / Failed / Stopped
- 明细区
  - 成功项列表
  - 失败项列表
  - 每个失败项的错误摘要
- 退出区
  - 提示用户按 `Enter` 或 `q` 退出

## 键位设计

### Select View 普通模式

- `↑` / `k`：上移
- `↓` / `j`：下移
- `Space`：勾选/取消当前项
- `a`：全选/取消全选当前可见项
- `/`：进入搜索模式
- `Enter`：确认执行
- `?`：展开/折叠帮助
- `Esc` / `q`：取消退出

兼容行为：

- 如果用户按 `Enter` 时没有显式勾选任何项，则默认执行当前光标所在项
- 这样兼容现有单选的使用习惯

### 搜索模式

- 任意字符：更新过滤词
- `Backspace`：删除字符
- `Enter`：保留过滤词并返回普通模式
- `Esc`：清空过滤词并返回普通模式

搜索行为：

- 实时过滤
- 大小写不敏感
- 使用子串匹配

匹配字段：

- 标题
- 描述
- 额外关键词

### Run View

- `q`：请求在当前项完成后停止后续执行
- `Ctrl+C`：强制退出

### Summary View

- `↑` / `↓`：滚动结果
- `Enter` / `q`：退出
- `?`：显示帮助

## 命令级展示规则

### Data Sync

选择页项目展示：

- 标题：基础文件名，例如 `field_module`
- 描述：`Entity: FieldModule`

执行步骤建议统一为：

- `检查 entity 占位文件`
- `生成 entity 占位文件`
- `生成 repo`
- `更新 ProviderSet`
- `运行 wire`

如果某步不需要执行，例如 entity 已存在，则记录为 `skip`。

### Handler Generate

选择页项目展示：

- 标题：文件名，例如 `test_case.go`
- 描述：`Handler: TestCaseHandler, 12 个方法`

执行步骤建议统一为：

- `生成 handler 文件`
- `更新 grpc.go`
- `生成 core service`
- `更新 core ProviderSet`
- `运行 wire`

## 状态流

统一状态机建议包含以下状态：

- `selecting`
- `searching`
- `running`
- `summary`
- `aborted`

状态流：

1. 命令层完成候选项发现
2. 进入 `selecting`
3. 用户可进入 `searching`
4. 用户确认后进入 `running`
5. 批处理完成后进入 `summary`
6. 用户退出程序

## 执行策略

执行采用串行模式，不做并发。

原因：

- `handler` 和 `data` 都涉及文件写入
- 两者都可能更新 `ProviderSet`
- 两者都可能调用 `wire`
- 并发执行会增加顺序和冲突风险，不适合当前工具

停止策略：

- `q` 在 `Run View` 中表示软停止
- 当前项执行完后停止处理后续项
- `Ctrl+C` 保持为强制退出

错误处理策略：

- 单项失败不终止整个批次
- 失败项记录到 summary
- 后续项继续执行
- 致命初始化错误直接返回，不进入 `Run View`

## 代码结构拆分

### 新增 `internal/tui/`

职责：纯 UI 状态机与渲染逻辑。

建议文件：

- `internal/tui/types.go`
  - 定义 UI item、运行结果、进度事件
- `internal/tui/model.go`
  - 主 `tea.Model`
- `internal/tui/select_view.go`
  - 多选、搜索、列表渲染
- `internal/tui/run_view.go`
  - spinner、progress、最近结果展示
- `internal/tui/summary_view.go`
  - 汇总视图
- `internal/tui/keys.go`
  - 键位定义
- `internal/tui/styles.go`
  - lipgloss 样式
- `internal/tui/fallback.go`
  - 非 TTY 环境下的简化 fallback

使用组件：

- `bubbletea`
- `bubbles/spinner`
- `bubbles/progress`
- `bubbles/help`
- `bubbles/key`
- `bubbles/textinput`
- `bubbles/viewport`
- `lipgloss`

明确不使用 `bubbles/list` 作为核心多选模型。

原因是本仓库的核心需求是“批量勾选”，不是“浏览型列表组件”。

### 新增 `internal/workflow/`

职责：命令级候选项发现与执行编排，向 UI 提供统一接口。

建议文件：

- `internal/workflow/handler.go`
- `internal/workflow/data.go`

每个 workflow 提供两类能力：

- `Discover(...) ([]tui.Item, error)`
- `RunItem(..., emit func(ProgressEvent)) RunResult`

### 调整现有命令层

- `cmd/handler.go`
- `cmd/data.go`

命令层职责收敛为：

- 解析参数
- 发现候选项
- 调用统一 UI 会话

### 调整现有选择器

`internal/selector/selector.go` 不再作为主交互入口。

本次改造完成后，非 TTY fallback 也迁移到 `internal/tui/fallback.go`，因此 `internal/selector/selector.go` 应删除，避免仓库中长期同时维护两套交互入口。

### 调整现有执行函数

当前 `generator.GenerateHandlerFile`、`RunDataCommand` 等函数含有直接输出行为。

统一 UI 后需要调整为：

- 底层函数返回结构化结果
- workflow 将结构化步骤转换为进度事件
- UI 负责展示，不再由业务逻辑直接打印

## 兼容策略

### 非 TTY 环境

在非 TTY 环境中，不启动 Bubble Tea UI，回退到简化模式：

- 仍支持编号输入选择
- 保持原有脚本场景可用

### 当前单选行为兼容

即便支持多选，直接按回车而不显式勾选时，仍默认执行当前项，保证老用户不需要重新学习“单项执行”。

## 测试策略

### `internal/tui/` 测试

重点覆盖：

- 光标移动
- `Space` 勾选/取消
- `a` 全选/取消全选
- `/` 进入搜索模式
- 搜索结果过滤
- `Enter` 无勾选时默认选择当前项
- `q` 在运行页触发软停止

### `internal/workflow/` 测试

重点覆盖：

- `data` workflow 的步骤事件顺序
- `handler` workflow 的步骤事件顺序
- 单项失败后后续项继续
- `wire` 失败时结果记录正确

### 回归测试

保留并扩展当前逻辑测试：

- scanner
- generator
- `data` 同步规则
- handler 批量执行行为
- fallback 选择逻辑

## 第一版范围

第一版明确包含：

- 统一选择页
- 多选
- `/` 搜索
- spinner
- progress
- summary
- `handler` 与 `data` 统一接入
- 非 TTY fallback

第一版明确不包含：

- 模糊排序
- 鼠标支持
- 多列布局
- 并发执行
- 主题切换
- 历史命令记忆

## 结论

对当前仓库来说，最合适的实现方式不是硬套现成列表组件，而是：

- 用 `Bubble Tea` 建立统一状态机
- 用 `Bubbles` 提供 spinner/help/key/progress/textinput/viewport 等交互部件
- 选择页采用自定义多选模型
- 执行页与结果页统一承接 `handler` 与 `data` 两条命令

这套方案在交互复杂度、维护成本和仓库适配度之间取得了最合理的平衡。
