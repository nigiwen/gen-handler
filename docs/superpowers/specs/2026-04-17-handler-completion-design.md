# Handler 文件创建与补全设计

**日期：** 2026-04-17
**状态：** 已完成并验证通过

## 落地结果

本设计对应的实现已经完成，当前仓库中的 `handler` 生成链路已按本文方案落地，并补充了两项实现期新增约束：

- `handler` 命令默认列出全部解析到的服务，支持多选
- `api/grpc/<service>.go` 与 `core/<service>.go` 都已支持“文件不存在则创建，存在则只补缺失方法”
- 只有本次创建了对应文件时，才更新 `api/grpc/grpc.go` / `core/core.go`
- 只有本次创建了任意一个文件时，才执行 `wire`
- 若发现已有同名方法但签名与当前 proto 不一致，生成器会直接报错并保持文件不变
- 主 proto 包名和主 proto import 不再写死为 `devopsx`，而是根据 grpc 文件 package 与 `ProtoDir` 推导
- `handler` 命令在 grpc 扫描/解析失败、session 失败时会向上返回错误，由 CLI 以非零退出码结束

本次实现后的关键验证已通过：

- `go test ./cmd ./internal/parser ./internal/generator ./internal/updater ./internal/workflow`
- `go test ./...`
- `go build .`

## 目标

调整 `handler` 命令的默认行为，不再只列出“缺失 handler”的服务，而是默认列出从 `*_grpc.pb.go` 中解析出的全部服务，交由用户多选执行。

执行时按目标文件是否存在区分“创建”与“补全”两种模式：

- 文件不存在时，按现有模板创建完整文件
- 文件已存在时，只补全缺失的方法
- 已有同名方法实现必须原样保留，不能覆盖

同时，将 `grpc.go`、`core.go` ProviderSet 更新和 `wire` 执行改为由“本次是否创建过文件”驱动，而不是由“当前服务是否全新”驱动。

## 当前问题

当前 `handler` 流程存在以下限制：

1. 命令入口会先过滤出“缺失 handler 文件”的服务，已存在 handler 的服务不会出现在候选列表中。
2. 对于已存在的 `api/grpc/<service>.go` 和 `core/<service>.go`，当前流程没有“补全缺失方法”的能力。
3. 现有 workflow 固定执行：
   - 生成 handler 文件
   - 更新 `grpc.go`
   - 生成 core service
   - 更新 core ProviderSet
   - 执行 `wire`
4. 这套固定流程适合“整文件首次创建”，但不适合“已有文件仅追加缺失方法”的场景。

## 设计原则

- 保持现有命令入口与多选交互，不新增额外子命令
- 优先复用现有 workflow、generator、template 结构
- 新建与补全的边界要清晰，避免覆盖用户已有实现
- 后置步骤由“本次是否创建文件”决定，规则必须稳定且可测试
- 对已有 Go 文件的补全宁可严格失败，也不做危险的文本兜底写入
- 对同名但签名漂移的方法宁可显式报错，也不做静默跳过或危险覆盖
- 生成代码时不得将主 proto 包路径硬编码到 `devopsx`

## 用户可见行为

`handler` 命令的新行为如下：

1. 默认列出全部服务，不再按“缺失 handler”预过滤。
2. 用户可以像现在一样单选或多选。
3. 对每个选中的服务，分别处理：
   - `api/grpc/<service>.go`
   - `core/<service>.go`
4. 当目标文件不存在时：
   - 创建完整文件
   - 后续按规则更新 ProviderSet / 执行 `wire`
5. 当目标文件已存在时：
   - 仅补全缺失方法
   - 不覆盖已有方法
   - 不因“补全”触发对应的 ProviderSet 更新
6. 只要本次创建了 `grpc` 或 `core` 任意一个文件，就执行 `wire`
7. 只有在本次没有创建任何文件时，才不执行 `wire`

## 执行矩阵

对每个选中的服务，先判断两个目标文件是否存在：

- `grpc` 文件：`api/grpc/<service>.go`
- `core` 文件：`core/<service>.go`

### 情况 1：`grpc` 不存在，`core` 不存在

- 创建 `grpc` 文件，写入完整模板
- 更新 `api/grpc/grpc.go` 的 `ProviderSet` 和 `NewGRPCServer`
- 创建 `core` 文件，写入完整模板
- 更新 `core/core.go` 的 `ProviderSet`
- 执行 `wire`

### 情况 2：`grpc` 存在，`core` 不存在

- 补全已有 `grpc` 文件中的缺失方法
- 不更新 `api/grpc/grpc.go`
- 创建 `core` 文件，写入完整模板
- 更新 `core/core.go` 的 `ProviderSet`
- 执行 `wire`

### 情况 3：`grpc` 不存在，`core` 存在

- 创建 `grpc` 文件，写入完整模板
- 更新 `api/grpc/grpc.go` 的 `ProviderSet` 和 `NewGRPCServer`
- 补全已有 `core` 文件中的缺失方法
- 不更新 `core/core.go`
- 执行 `wire`

### 情况 4：`grpc` 存在，`core` 存在

- 只补全两个文件里各自缺失的方法
- 不更新 `api/grpc/grpc.go`
- 不更新 `core/core.go`
- 不执行 `wire`

## 补全的定义

补全操作必须满足以下约束：

- 只追加缺失的方法
- 已有同名且签名一致的方法一律保留
- 已有同名但签名不一致的方法直接报错，且文件保持不变
- 不覆盖已有方法体
- 不重排已有代码
- 不清理已有 import
- 由格式化步骤负责最终的 Go 代码格式整理

## 架构设计

本次改动保持现有命令入口和 UI，会将变更集中在 `workflow` 与 `generator` 两层。

### 1. 命令层

涉及文件：

- `cmd/handler.go`

调整方向：

- 移除“只保留缺失 handler 服务”的前置过滤
- 直接将全部解析结果交给 `workflow.HandlerWorkflow`
- 保持现有 TTY / fallback 多选体验不变

### 2. Workflow 层

涉及文件：

- `internal/workflow/handler.go`

职责调整：

- 从“固定 5 步”改为“按文件状态执行分支”
- 在单个服务执行过程中记录：
  - `grpcCreated`
  - `coreCreated`
- 执行规则：
  - `grpcCreated == true` 时才更新 `api/grpc/grpc.go`
  - `coreCreated == true` 时才更新 `core/core.go`
  - `grpcCreated || coreCreated` 时才执行 `wire`

建议的步骤语义：

- `处理 handler 文件`
- `更新 grpc.go`
- `处理 core service`
- `更新 core ProviderSet`
- `运行 wire`

其中：

- “处理 handler 文件”既可能是创建，也可能是补全
- “处理 core service”既可能是创建，也可能是补全
- 如果后续步骤因未创建文件而不执行，应明确记为 `skip` 或不发该步骤事件，但行为必须在测试中固定下来

### 3. Generator 层

涉及文件：

- `internal/generator/handler.go`
- `internal/generator/core.go`
- `internal/generator/template.go`

职责拆分：

- 保留现有“整文件创建”的能力
- 新增“补全已有文件缺失方法”的能力
- 对外暴露“确保文件存在且方法齐全”的统一入口

建议接口方向：

- `EnsureHandlerFile(service, config) (created bool, err error)`
- `EnsureCoreServiceFile(service, config) (created bool, err error)`

语义：

- 文件不存在：创建完整文件并返回 `created=true`
- 文件存在：仅补全缺失方法并返回 `created=false`

### 4. 模板层

`internal/generator/template.go` 在保留现有整文件模板的同时，新增方法级模板：

- `HandlerMethodTemplate`
- `CoreMethodTemplate`

使用方式：

- 创建新文件时：继续使用整文件模板
- 补全已有文件时：只使用方法级模板生成缺失的方法片段

## 补全实现策略

补全已有文件时，不采用整文件重写，也不采用脆弱的纯文本字符串匹配追加。

采用的策略是：

1. 读取已有文件内容
2. 解析 Go AST
3. 确认目标类型存在：
   - handler 文件中应存在 `type XxxHandler struct`
   - core 文件中应存在 `type XxxService struct`
4. 收集当前文件中该 receiver 已实现的方法名
5. 根据 `ServiceInfo.Methods` 计算缺失方法集合
6. 若缺失集合为空：
   - 不修改文件
   - 返回 `created=false`
7. 若存在缺失方法：
   - 通过方法级模板生成缺失方法代码片段
   - 追加到文件末尾
   - 统一走格式化

这样可以保证：

- 不覆盖已有实现
- 不依赖现有方法顺序
- 不强行重写整个文件
- 能较稳妥地补齐新增 RPC 对应的方法

## 数据流

单个服务的执行流如下：

1. workflow 收到用户选中的 `ServiceInfo`
2. 调用 handler 生成器：
   - 创建完整文件或补全缺失方法
   - 返回 `grpcCreated`
3. 若 `grpcCreated == true`：
   - 更新 `api/grpc/grpc.go`
4. 调用 core 生成器：
   - 创建完整文件或补全缺失方法
   - 返回 `coreCreated`
5. 若 `coreCreated == true`：
   - 更新 `core/core.go`
6. 若 `grpcCreated || coreCreated`：
   - 在 `wire-dir` 下执行 `wire`
7. 返回该服务的执行结果

## 错误处理

### 补全阶段

- 若已有文件无法被正常解析为 Go 文件，直接报错
- 若无法识别预期类型：
  - handler 文件缺少 `type XxxHandler struct`
  - core 文件缺少 `type XxxService struct`
  则直接报错
- 若存在同名方法但签名与当前 proto 不一致，直接报错
- 不提供字符串级兜底补全，避免破坏用户现有代码

### 后置步骤

- `grpc.go` 更新失败：该服务执行失败
- `core.go` ProviderSet 更新失败：该服务执行失败
- `wire` 执行失败：该服务执行失败

### 命令层

- 扫描 grpc 文件失败：命令直接返回错误
- 未找到 grpc 文件：命令直接返回错误
- 全部 grpc 文件解析失败：命令直接返回错误
- 部分 grpc 文件解析失败但仍有可执行服务时：
  - session 仍会执行成功解析出的服务
  - 命令结束时返回聚合错误，避免脚本误判为成功
- session 执行失败：命令直接返回错误，由 CLI 以非零退出码结束

### 执行边界

- “补全已有文件”本身不触发 ProviderSet 更新
- ProviderSet 更新只和“本次是否创建对应文件”有关
- `wire` 执行只和“本次是否创建了任意文件”有关

## 测试策略

本次改动必须遵循 TDD，先写失败测试，再写实现。

### 1. 命令层测试

涉及文件：

- `cmd/handler_test.go`

新增验证点：

- `handler` 命令现在默认将全部服务交给 session
- 不再调用“缺失 handler 过滤”来裁剪候选项

### 2. Workflow 测试

涉及文件：

- `internal/workflow/handler_test.go`

重点覆盖四种执行矩阵：

1. `grpc` 不存在，`core` 不存在
2. `grpc` 存在，`core` 不存在
3. `grpc` 不存在，`core` 存在
4. `grpc` 存在，`core` 存在

每种情况都要断言：

- 哪些步骤执行
- 哪些步骤不执行
- `grpc.go` 是否更新
- `core.go` 是否更新
- `wire` 是否执行

其中必须明确验证：

- 只要创建过任意文件，就执行 `wire`
- 没有创建任何文件时，不执行 `wire`

### 3. Handler 生成器测试

建议新增文件：

- `internal/generator/handler_test.go`

覆盖行为：

- 文件不存在时，创建完整 handler 文件并返回 `created=true`
- 文件存在但缺少部分 RPC 方法时，只补全缺失方法并返回 `created=false`
- 已有 handler 方法实现不被覆盖
- 没有缺失方法时文件内容不变
- 非法 Go 文件时报错
- 缺少 `type XxxHandler struct` 时报错

### 4. Core 生成器测试

建议新增文件：

- `internal/generator/core_test.go`

覆盖行为：

- 文件不存在时，创建完整 core 文件并返回 `created=true`
- 文件存在但缺少部分 service 方法时，只补全缺失方法并返回 `created=false`
- 已有 core 方法实现不被覆盖
- 没有缺失方法时文件内容不变
- 非法 Go 文件时报错
- 缺少 `type XxxService struct` 时报错

### 5. 回归验证

实现完成后至少执行：

- 相关包定向测试
- `go test ./...`

## 范围边界

本次设计明确包含：

- `handler` 默认列出全部服务
- 现有文件的缺失方法补全
- 由“创建过文件”驱动 ProviderSet 更新与 `wire`
- 同名方法签名漂移检测
- 主 proto 包与主 proto import 的动态推导
- `handler` 命令错误向上传播

本次设计明确不包含：

- 自动删除多余方法
- 自动清理 import
- 自动重排已有方法顺序
- 对用户已有实现做语义级合并
- 对异常文件结构做宽松容错修复

## 结论

对当前仓库来说，最合适的方案是：

- 保留现有命令入口与多选执行框架
- 将核心变更集中在 `workflow` 和 `generator`
- 用“创建或补全”的统一入口取代当前的固定生成步骤
- 用 AST 识别已有方法并只追加缺失方法
- 将 `grpc.go`、`core.go` 更新和 `wire` 执行明确绑定到“本次是否创建过文件”

这套设计在复用现有结构、避免覆盖用户代码、以及满足新增补全需求之间取得了最稳妥的平衡。
