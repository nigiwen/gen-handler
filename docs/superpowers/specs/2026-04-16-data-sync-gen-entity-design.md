# Data Sync 基于 `*.gen.go` 的实体发现设计

**日期：** 2026-04-16

## 目标

调整 `data` 命令的行为，使其只从 `internal/model/entity` 下的 `*.gen.go` 文件发现候选实体；在用户选择后，按需补一个对应的非生成 entity 占位文件；停止生成 `data/dbset/*.go`；同时保持现有 `data/<name>.go`、`ProviderSet` 和 `wire` 的生成链路不变。

## 当前行为

当前的 `data` 命令会扫描 `internal/model/entity` 下所有 `.go` 文件，仅排除 `entity.go` 和 `*_test.go`。对每个扫描到的实体，命令可能生成：

- `data/dbset/<name>.go`
- `data/<name>.go`
- `data/data.go` 中的 `ProviderSet` 注册项
- repo 生成成功后的 `wire` 刷新

这已经不符合当前目标项目的实际工作流，因为现在的实体来源已经变成 `internal/model/entity` 下由 GORM 生成的 `*.gen.go` 文件。

## 需求

### 发现规则

- 只扫描 `internal/model/entity` 目录。
- 只有文件名匹配 `*.gen.go` 的文件才算候选实体。
- 目录下其他文件一律忽略，包括手写的普通 `*.go` 文件。
- 候选文件名转换规则如下：
  - `project.gen.go` 映射为 `FileName=project`，`EntityName=Project`
  - `project_member.gen.go` 映射为 `FileName=project_member`，`EntityName=ProjectMember`

### 同步规则

- 保留现有的交互式选择流程。
- 用户选中某个实体后：
  - 如果 `internal/model/entity/<name>.go` 不存在，则创建该文件，内容固定为：

```go
package entity
```

  - 如果 `internal/model/entity/<name>.go` 已存在，则跳过，不覆盖。
- 不再生成 `data/dbset/<name>.go`。
- 继续按现有 repo 模板逻辑生成 `data/<name>.go`。
- 只有在本次真正生成了新的 repo 文件时，才继续更新 `data/data.go` 的 `ProviderSet`。
- 只有在 repo 文件生成成功且 `ProviderSet` 更新成功后，才继续运行 `wire`。

### 待同步判定

- 一个实体是否属于“待同步”，只由 `data/<name>.go` 是否存在决定。
- `internal/model/entity/<name>.go` 是否存在，不影响它是否出现在待同步列表中。
- 因此，如果 entity 占位文件已经存在，但 `data/<name>.go` 不存在，该实体仍然必须显示为待同步，并继续执行 repo 生成。

## 采用方案

采用方案 1：保留现有 `data` 命令主流程，只调整实体发现方式，并在 repo 生成前新增一个“补 entity 占位文件”的步骤。

### 选择这个方案的原因

- 可以保持 repo 生成、`ProviderSet` 更新和 `wire` 执行逻辑不变。
- 改动范围集中在真正变化的部分，风险更低。
- 不需要为了这次需求做更大范围的命令层重构，同时仍然可以把新行为补上测试。

## 详细设计

### Scanner 改动

调整实体扫描逻辑，使 scanner：

- 只接受 `*.gen.go`
- 去掉 `.gen.go` 后缀，得到逻辑基础文件名
- 基于这个基础文件名继续使用现有的 snake_case 转 UpperCamel 规则生成 `EntityName`

scanner 的输出结构保持不变：

- `EntityInfo.FileName` 仍然保存去掉 `.gen` 之后的基础文件名
- `EntityInfo.EntityName` 仍然保存 repo 生成所需的 UpperCamel 名称

这样可以保证下游对 `data/<name>.go` 和 `internal/model/entity/<name>.go` 的路径拼接逻辑基本不用改。

### Generator 改动

用 entity 占位文件生成替换原来的 dbset 生成：

- 删除 dbset 模板和 dbset 生成函数
- 新增一个 generator 函数，用于生成 `internal/model/entity/<name>.go`
- 这个文件内容固定为 `package entity`
- repo 生成逻辑保持不变

### 命令流改动

`RunDataCommand` 仍然保持以下主流程：

1. 扫描候选实体
2. 计算待同步实体
3. 让用户交互选择
4. 逐个处理被选中的实体

每个被选中实体的新处理顺序如下：

1. 确保 `internal/model/entity/<name>.go` 存在，不存在则创建
2. 确保 `data/<name>.go` 存在，不存在则创建
3. 如果本次生成了新的 repo 文件：
   - 更新 `data/data.go`
   - 运行 `wire`

### 停止生成 dbset

命令不再创建：

- `data/dbset` 目录
- 任何 `data/dbset/*.go` 文件

已经存在于消费项目中的 `data/dbset` 文件和目录不在本次需求处理范围内。工具只是停止继续生成新的 dbset 文件，不负责清理旧文件。

## 测试策略

补充最小但完整的行为测试，重点覆盖发生变化的部分。

### Scanner 测试

增加 scanner 测试，验证：

- `project.gen.go` 会被识别为 `project`
- `project_member.gen.go` 会被识别为 `project_member`
- `project.go`、`entity.go`、`project_test.go` 等非目标文件不会被识别为候选实体

### Generator 测试

增加 generator 测试，验证新的 entity 占位文件生成函数写出的内容严格等于：

```go
package entity
```

并且不包含其他内容。

### Command 级测试

增加围绕待同步判定和单实体处理流程的命令级测试，验证：

- 当 `data/<name>.go` 缺失时，即使 `internal/model/entity/<name>.go` 已经存在，该实体仍然属于待同步
- 当用户选择一个实体且没有手写占位文件时，会同时补出 entity 占位文件和 repo 文件
- 当用户选择一个实体且手写占位文件已存在时，会跳过占位文件生成，但仍然继续生成 repo 文件

测试应在临时目录中隔离文件副作用，并避免真正调用系统里的 `wire` 命令。

## 预计改动文件

- `cmd/data.go`
- `internal/scanner/entity.go`
- `internal/generator/data.go`
- `internal/generator/template.go`
- 若干新的 `*_test.go`
- 当前文档中所有仍描述 dbset 生成功能的位置

## 不在本次范围内

- 修改 repo 模板内容
- 修改 `ProviderSet` 的插入规则
- 修改 `wire` 的执行方式
- 清理消费项目里已经存在的 `data/dbset` 文件
- 扩展到 `internal/model/entity` 以外的扫描目录
