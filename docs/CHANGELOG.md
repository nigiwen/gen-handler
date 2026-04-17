# Changelog

所有重要的变更都会记录在这个文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### 变更

- 暂无

## [1.4.0] - 2026-04-17

### 新增功能

- ✨ **统一终端交互**：`handler` 与 `data` 命令统一切换到 Bubble Tea TUI，支持多选、搜索、执行进度和结果汇总。
- ✨ **版本输出**：新增 `-version` 参数，可直接输出构建注入的版本号。

### 改进

- ✨ **Handler 默认展示全部服务**：`handler` 命令不再按“缺失文件”预过滤，默认列出全部解析出的 gRPC 服务供用户选择。
- ✨ **按需补全 Handler/Core 文件**：目标文件不存在时创建完整文件；文件已存在时只补缺失方法，保留已有实现。
- 🔄 **按创建结果更新 ProviderSet / wire**：仅在本次新建了 `api/grpc/<service>.go` 时更新 `api/grpc/grpc.go`，仅在本次新建了 `core/<service>.go` 时更新 `core/core.go`，且只有创建了任意文件时才执行 `wire`。
- 🧠 **去掉主 proto 硬编码**：主 proto 包名和 import 路径改为根据 `proto-dir` 与 grpc 文件 package 自动推导，不再写死为 `devopsx`。
- 📦 **调整 Data 同步来源**：`data` 命令现在只扫描 `internal/model/entity/*.gen.go` 作为候选实体来源。
- ✨ **补手写 entity 占位文件**：选择实体后，如果 `internal/model/entity/<name>.go` 不存在，会自动生成仅包含 `package entity` 的占位文件。
- 🗑️ **停止生成 dbset 文件**：不再生成 `data/dbset/*.go`，repo、`ProviderSet` 和 `wire` 逻辑保持不变。

### 修复

- 🛡️ **签名漂移保护**：如果已存在同名方法但签名与当前 proto 不一致，工具会直接报错并保持文件不变。
- 📤 **命令失败状态更准确**：`handler` 命令现在会把关键错误向上传播，便于 CLI 返回非零退出码。
- 📜 **长列表滚动体验修复**：选择列表在条目较多时会跟随光标自动滚动。
- 🎨 **当前项高亮增强**：当前光标所在项的标题和说明文字使用更明显的高亮颜色显示。

## [1.3.0] - 2026-01-28

### 重构

- 🏗️ **项目结构模块化**：将扁平结构重构为模块化的 `internal/` 目录结构，提高代码可维护性
  - `internal/types/` - 类型定义
  - `internal/util/` - 工具函数（命名转换、文件操作、AST 操作、模块路径）
  - `internal/parser/` - gRPC 文件解析
  - `internal/scanner/` - 文件扫描（handler、entity）
  - `internal/selector/` - 交互式选择
  - `internal/generator/` - 代码生成（handler、core、data、模板）
  - `internal/updater/` - 代码更新（grpc、provider）
  - `cmd/` - 命令层（handler、data）
  - `docs/` - 文档

### 修复

- 🐛 **修复 gRPC 服务注册问题**：修复 `RegisterXXXServer` 调用未正确添加到 `NewGRPCServer` 函数体的问题
- 🐛 **修复花括号匹配**：新增 `FindMatchingCurlyBrace` 函数正确处理函数体的花括号匹配

## [1.2.0] - 2026-01-28

### 新增功能

- ✨ **数据同步功能 (Data Sync)**：支持自动扫描 `internal/model/entity` 目录下的实体，并生成对应的 `dbset` 类型别名和 `repo` 存储库代码。
- ✨ **自动注册 ProviderSet**：生成的 `repo` 会自动注册到 `data/data.go` 的 `ProviderSet` 中，并自动触发 `wire` 命令重新生成依赖注入代码。
- ✨ **交互式同步选择**：在执行数据同步时，支持交互式选择需要同步的实体。

### 改进

- 🚀 增强了主程序的子命令支持，现在可以通过交互式菜单选择执行“服务生成”或“数据同步”。

## [1.1.0] - 2026-01-23

### 新增功能

- ✨ **自动读取 go.mod**：工具现在会自动从项目根目录的 `go.mod` 文件中读取 module 路径，无需手动指定 `-module` 参数
- ✨ **自动生成 proto-dir**：根据 module 路径自动生成 `proto-dir`，规则为 `./internal/proto` + (module 去掉第一个 `/` 前面的部分)
- ✨ **自动生成 wire-dir**：根据 module 路径自动生成 `wire-dir`，规则为 `./cmd` + (module 最后一个 `/` 后面的部分)
- 📝 添加了 `TEST.md` 测试指南文档

### 改进

- 🔧 优化了命令行参数，`-module`、`-proto-dir` 和 `-wire-dir` 现在都支持自动生成
- 📚 更新了 README 文档，详细说明了自动配置功能和使用示例

### 使用示例

现在使用工具更加简单，只需要指定必要的参数：

```bash
# 在项目根目录运行（自动读取 go.mod 并生成相关路径）
gen-handler \
  -output-dir ./api/grpc \
  -core-dir ./core
```

工具会自动：
- 从 `go.mod` 读取 module（如：`bsi/axis/devopsx`）
- 生成 `proto-dir`（如：`./internal/proto/axis/devopsx`）
- 生成 `wire-dir`（如：`./cmd/devopsx`）

## [1.0.0] - 2026-01-23

### 初始版本

- 🎉 首次发布
- ✨ 支持自动扫描和解析 `*_grpc.pb.go` 文件
- ✨ 支持自动生成 gRPC Handler 文件
- ✨ 支持自动生成 Core Service 文件
- ✨ 支持自动更新 `grpc.go` 和 `core.go` 的 ProviderSet
- ✨ 支持自动运行 `wire` 命令生成依赖注入代码
- ✨ 支持交互式选择要生成的服务
