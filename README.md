# gRPC Handler 生成工具

这是一个用于自动生成 gRPC Handler 和 Core Service 代码的 Go 工具插件。

## 功能特性

- 🔍 自动扫描 proto 生成的 `*_grpc.pb.go` 文件
- 📝 自动生成 gRPC Handler 文件
- 🏗️ 自动生成 Core Service 文件
- 🔄 自动更新 `grpc.go` 和 `core.go` 的 ProviderSet
- ⚡ 自动运行 `wire` 命令生成依赖注入代码
- 🎯 交互式选择要生成的服务

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

## 命令行参数

| 参数 | 说明 | 默认值/自动生成 |
|------|------|--------|
| `-proto-dir` | proto 生成的 grpc 文件目录 | 未指定时自动从 module 生成：`./internal/proto/{module去掉第一个/前面的部分}` |
| `-output-dir` | handler 输出目录 | `./api/grpc` |
| `-core-dir` | core service 输出目录 | `./core` |
| `-wire-dir` | wire 命令执行目录 | 未指定时自动从 module 生成：`./cmd/{module最后一个/后面的部分}` |
| `-module` | Go 模块路径（用于生成 import 路径） | 未指定时自动从 `go.mod` 读取 |
| `-help` | 显示帮助信息 | - |

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
3. **检查缺失**：检查哪些服务还没有生成 handler 文件
4. **交互选择**：使用上下键选择要生成的服务（支持交互式界面）
5. **生成代码**：
   - 生成 Handler 文件到 `output-dir`
   - 生成 Core Service 文件到 `core-dir`
   - 更新 `grpc.go` 的 ProviderSet
   - 更新 `core.go` 的 ProviderSet
6. **运行 Wire**：在 `wire-dir` 目录下运行 `wire` 命令

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
5. **文件覆盖**：如果文件已存在，工具会跳过生成（保护已有代码）
6. **项目结构**：工具假设你的项目遵循标准的 Go 项目结构，proto 文件已编译生成 `*_grpc.pb.go` 文件

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
│   ├── selector/             # 交互式选择服务
│   ├── generator/            # 代码生成（handler、core、data、模板）
│   └── updater/              # 代码更新（grpc、provider）
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
./gen-handler -help
```

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
