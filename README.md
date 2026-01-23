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

### 方式一：作为项目工具使用（当前方式）

```bash
# 在项目根目录运行
go run ./tools/gen-handler

# 或使用 Makefile
make gen-handler
```

### 方式二：安装为全局工具（推荐）

```bash
# 从当前项目安装
go install ./tools/gen-handler

# 或从 GitHub 安装（发布后）
go install github.com/nigiwen/gen-handler@latest

# 使用
gen-handler
```

**注意**：如果从 GitHub 安装，需要先发布到 GitHub，详见 `PUBLISH.md`

## 使用方法

### 基本使用（使用默认配置）

```bash
gen-handler
```

默认配置：
- `proto-dir`: `./internal/proto/axis/devopsx`
- `output-dir`: `./api/grpc`
- `core-dir`: `./core`
- `wire-dir`: `./cmd/devopsx`
- `module`: `bsi/axis/devopsx`

### 自定义配置

```bash
gen-handler \
  -proto-dir ./internal/proto/axis/devopsx \
  -output-dir ./api/grpc \
  -core-dir ./core \
  -wire-dir ./cmd/devopsx \
  -module bsi/axis/devopsx
```

### 查看帮助

```bash
gen-handler -help
```

## 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-proto-dir` | proto 生成的 grpc 文件目录 | `./internal/proto/axis/devopsx` |
| `-output-dir` | handler 输出目录 | `./api/grpc` |
| `-core-dir` | core service 输出目录 | `./core` |
| `-wire-dir` | wire 命令执行目录 | `./cmd/devopsx` |
| `-module` | Go 模块路径（用于生成 import 路径） | `bsi/axis/devopsx` |
| `-help` | 显示帮助信息 | - |

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

1. **路径配置**：确保所有路径参数都是相对于项目根目录的
2. **模块路径**：`-module` 参数应该与项目的 `go.mod` 中的模块路径一致
3. **Wire 命令**：确保系统已安装 `wire` 工具，且 `wire-dir` 目录存在
4. **文件覆盖**：如果文件已存在，工具会跳过生成（除非使用强制模式）

## 开发说明

### 项目结构

```
tools/gen-handler/
├── main.go              # 主入口，命令行参数解析
├── types.go              # 类型定义（ServiceInfo, Method, Config）
├── parser.go             # 解析 grpc 文件
├── finder.go             # 查找缺失的 handler
├── selector.go           # 交互式选择服务
├── generator.go          # 生成 handler 代码
├── core_generator.go     # 生成 core service 代码
├── updater.go            # 更新 grpc.go
├── core_updater.go       # 更新 core.go
└── utils.go              # 工具函数
```

### 扩展开发

如果需要支持其他项目结构，可以：
1. 修改默认配置值
2. 添加配置文件支持（如 YAML/JSON）
3. 支持模板自定义

## 许可证

与主项目保持一致。
