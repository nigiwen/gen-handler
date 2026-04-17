# gRPC Handler 生成工具

这是一个用于自动生成 gRPC Handler 和 Core Service 代码的 Go 工具插件。

## 功能特性

- 🔍 自动扫描 proto 生成的 `*_grpc.pb.go` 文件
- 📝 支持创建或补全 gRPC Handler 文件
- 🏗️ 支持创建或补全 Core Service 文件
- 🔄 仅在新建文件时更新 `grpc.go` 和 `core.go` 的 ProviderSet
- 📦 支持 Data 层同步：从 `internal/model/entity/*.gen.go` 发现实体，按需补 `internal/model/entity/<name>.go` 并生成 `data/<name>.go`
- ⚡ 仅在本次创建了任意文件时自动运行 `wire`
- 🎯 统一 Bubble Tea 终端 UI：支持 `Space` 多选、`/` 搜索、执行进度和结果汇总
- 🏷️ 支持 `-version` 直接输出当前版本

## 安装方式

这是一个独立的 Go 工具，需要先安装才能使用。

### 从 GitHub 安装（推荐）

```bash
# 安装最新版本
go install github.com/nigiwen/gen-handler@latest

# 或安装特定版本
go install github.com/nigiwen/gen-handler@v1.0.0
```

### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/nigiwen/gen-handler.git
cd gen-handler

# 安装
go install
```

### 验证安装

```bash
# 检查是否安装成功
which gen-handler
gen-handler -version
gen-handler -help
```

安装后，工具会被安装到 `$GOPATH/bin` 或 `$GOBIN` 目录，确保该目录在 `$PATH` 环境变量中。

## 使用方法

### 基本使用（推荐 ⭐）

工具支持**自动配置**，会自动从项目的 `go.mod` 读取 module 路径，并自动生成 `proto-dir`。

```bash
# 在项目根目录运行（会自动读取 go.mod 并生成相关路径）
gen-handler \
  -output-dir ./api/grpc \
  -core-dir ./core
```

**自动配置规则**：
- `module`: 自动从 `go.mod` 读取（如：`bsi/axis/devopsx`）
- `proto-dir`: 自动从 module 生成（规则：`./internal/proto` + module去掉第一个`/`前面的部分）
  - 例如：`bsi/axis/devopsx` → `./internal/proto/axis/devopsx`
- `wire-dir`: 自动从 module 生成（规则：`./cmd` + module最后一个`/`后面的部分）
  - 例如：`bsi/axis/devopsx` → `./cmd/devopsx`

### 自定义配置

如果需要覆盖自动配置，可以手动指定参数：

```bash
gen-handler \
  -proto-dir ./internal/proto/your-project \
  -output-dir ./api/grpc \
  -core-dir ./core \
  -wire-dir ./cmd/your-service \
  -module your-module/path
```

**注意**：
- 如果不指定 `-module`，工具会自动从 `go.mod` 读取
- 如果不指定 `-proto-dir`，工具会根据 module 自动生成
- 如果找不到 `go.mod` 且未指定 `-module`，工具会报错并提示

### 查看帮助

```bash
gen-handler -help
```

### 查看版本

```bash
gen-handler -version
```

### Data 同步命令

`data` 子命令用于同步 Data 层，当前行为如下：

- 扫描 `internal/model/entity/*.gen.go`
- 通过统一 TUI 选择待同步实体
- 如果 `internal/model/entity/<name>.go` 不存在，则补一个只包含 `package entity` 的占位文件
- 生成 `data/<name>.go`
- 更新 `data/data.go` 的 `ProviderSet`
- 运行 `wire`

示例：

```bash
gen-handler data
```

### 统一交互

- `↑/↓` 或 `j/k`：移动光标
- `Space`：勾选或取消当前项
- `a`：全选或取消全选当前可见项
- `/`：进入搜索模式
- `Enter`：执行已勾选项；如果没有勾选，则默认执行当前项
- `q`：退出；在运行页表示“当前项完成后停止”
- 非 TTY 环境会自动切换到编号输入 fallback

## 命令行参数

| 参数 | 说明 | 默认值/自动生成 |
|------|------|--------|
| `-proto-dir` | proto 生成的 grpc 文件目录 | 未指定时自动从 module 生成：`./internal/proto/{module去掉第一个/前面的部分}` |
| `-output-dir` | handler 输出目录 | `./api/grpc` |
| `-core-dir` | core service 输出目录 | `./core` |
| `-wire-dir` | wire 命令执行目录 | 未指定时自动从 module 生成：`./cmd/{module最后一个/后面的部分}` |
| `-module` | Go 模块路径（用于生成 import 路径） | 未指定时自动从 `go.mod` 读取 |
| `-help` | 显示帮助信息 | - |
| `-version` | 显示版本信息并退出 | 默认构建值为 `dev` |

**自动生成规则示例**：
- Module: `bsi/axis/devopsx`
  - Proto-dir: `./internal/proto/axis/devopsx`
  - Wire-dir: `./cmd/devopsx`
- Module: `github.com/user/project`
  - Proto-dir: `./internal/proto/user/project`
  - Wire-dir: `./cmd/project`
- Module: `example.com/service`
  - Proto-dir: `./internal/proto/service`
  - Wire-dir: `./cmd/service`

## 工作流程

1. **扫描文件**：在 `proto-dir` 目录下查找所有 `*_grpc.pb.go` 文件
2. **解析服务**：解析每个文件中的 gRPC 服务接口定义
3. **交互选择**：默认展示全部解析出的服务，使用统一 TUI 进行多选、搜索和确认
4. **处理代码**：对每个选中的服务分别处理 `output-dir/<service>.go` 和 `core-dir/<service>.go`
   - 文件不存在：创建完整文件
   - 文件已存在：只补全缺失方法，保留已有实现
   - 若发现同名方法但签名与当前 proto 不一致：直接报错，不修改文件
5. **更新 ProviderSet**：
   - 只有本次新建了 Handler 文件时，才更新 `grpc.go`
   - 只有本次新建了 Core 文件时，才更新 `core.go`
6. **运行 Wire**：只有本次创建了任意一个文件时，才在 `wire-dir` 目录下运行 `wire`

## 生成的文件结构

### Handler 文件示例

```go
package grpc

import (
	"context"
	"bsi/axis/devopsx/core"
	"bsi/axis/devopsx/internal/proto/axis/devopsx"
	"bsi/axis/devopsx/internal/proto/basic"
)

type TestCaseHandler struct {
	devopsx.UnimplementedTestCaseServer
	testCaseSrv *core.TestCaseService
}

func NewTestCaseHandler(testCaseSrv *core.TestCaseService) *TestCaseHandler {
	return &TestCaseHandler{
		testCaseSrv: testCaseSrv,
	}
}

// CreateTestCase 创建测试用例
func (t *TestCaseHandler) CreateTestCase(ctx context.Context, in *devopsx.CreateTestCaseRequest) (*devopsx.CreateTestCaseResponse, error) {
	return t.testCaseSrv.CreateTestCase(ctx, in)
}
```

### Core Service 文件示例

```go
package core

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"bsi/axis/devopsx/internal/micro/client"
	"bsi/axis/devopsx/internal/proto/axis/devopsx"
	"bsi/axis/devopsx/internal/proto/basic"
	mgorm "bsi/kratos/micro/gorm"
)

type TestCaseService struct {
	srvClient        *client.Client
	log              *log.Helper
	bs               *devopsx.Bootstrap
	transactionScope *mgorm.TransactionScope
}

func NewTestCaseService(
	srvClient *client.Client,
	logger log.Logger,
	bs *devopsx.Bootstrap,
	transactionScope *mgorm.TransactionScope,
) *TestCaseService {
	return &TestCaseService{
		srvClient:        srvClient,
		log:              log.NewHelper(log.With(logger, "module", "testCase")),
		bs:               bs,
		transactionScope: transactionScope,
	}
}

// CreateTestCase 创建测试用例
func (t *TestCaseService) CreateTestCase(ctx context.Context, in *devopsx.CreateTestCaseRequest) (*devopsx.CreateTestCaseResponse, error) {
	t.log.Debug("not implement")
	return nil, nil
}
```

## 注意事项

1. **独立工具**：这是一个独立的工具，需要先安装才能使用（见安装方式）
2. **路径配置**：所有路径参数都是相对于**运行命令时的当前工作目录**（通常是项目根目录）
3. **模块路径**：`-module` 参数必须与你的项目 `go.mod` 中的模块路径完全一致，否则生成的 import 路径会错误
4. **Wire 命令**：确保系统已安装 `wire` 工具（`go install github.com/google/wire/cmd/wire@latest`），且 `wire-dir` 目录存在
5. **文件处理**：如果文件已存在，工具不会整文件覆盖，而是只补缺失方法
6. **签名漂移**：如果已有同名方法但签名与当前 proto 不一致，工具会报错并保持文件不变
7. **主 proto 包**：主 proto import 与包名会根据 `proto-dir` 和 grpc 文件自身 package 推导，不再写死为 `devopsx`
8. **项目结构**：工具假设你的项目遵循标准的 Go 项目结构，proto 文件已编译生成 `*_grpc.pb.go` 文件

## 开发说明

这是一个独立的 Go 工具项目，可以用于任何使用 gRPC 和 Wire 的 Go 项目。

### 项目结构

```
gen-handler/
├── main.go                    # 主入口，命令行参数解析和路由
├── cmd/                       # 命令层
│   ├── handler.go            # handler 生成命令
│   └── data.go               # data 同步命令
├── internal/                  # 内部实现
│   ├── types/                # 类型定义（ServiceInfo, Method, Config）
│   ├── util/                 # 工具函数（命名转换、文件操作、AST、模块路径）
│   ├── parser/               # gRPC 文件解析
│   ├── scanner/              # 文件扫描（handler、entity）
│   ├── generator/            # 代码生成（handler、core、data、模板）
│   ├── updater/              # 代码更新（grpc、provider）
│   ├── workflow/             # 命令级候选项发现与逐项执行
│   └── tui/                  # 统一 Bubble Tea 终端 UI
├── docs/                      # 文档（CHANGELOG、RELEASE 等）
├── go.mod                     # Go 模块定义
└── README.md                  # 本文档
```

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/nigiwen/gen-handler.git
cd gen-handler

# 安装依赖
go mod tidy

# 本地测试（在目标项目中）
go run . -proto-dir ./internal/proto/your-project ...

# 或编译后测试
go build -o gen-handler
./gen-handler -version
./gen-handler -help
```

### 打包与版本

```bash
# Windows
build.bat v1.2.3

# Linux/macOS
./build.sh v1.2.3
```

说明：

- `build.bat` 和 `build.sh` 未传版本号时，都会默认使用 `dev`
- 版本通过 `-ldflags "-X main.version=<VERSION>"` 注入
- 可用 `gen-handler -version` 或打包后的二进制 `-version` 验证最终版本值
- 构建产物会输出到 `dist/`

### 扩展开发

如果需要支持其他项目结构或功能，可以：
1. 修改默认配置值（在 `main.go` 中）
2. 添加配置文件支持（如 YAML/JSON）
3. 支持模板自定义
4. 添加更多代码生成选项

### 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

本项目采用与原始项目相同的许可证。详见 LICENSE 文件（如有）。
