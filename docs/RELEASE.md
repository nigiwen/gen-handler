# Release v1.3.0

## 🏗️ 项目结构重构

本版本对项目进行了全面的模块化重构，提高代码可维护性：

```
gen-handler/
├── main.go                    # 主入口
├── cmd/                       # 命令层
│   ├── handler.go            # handler 生成命令
│   └── data.go               # data 同步命令
├── internal/                  # 内部实现
│   ├── types/                # 类型定义
│   ├── util/                 # 工具函数
│   ├── parser/               # gRPC 文件解析
│   ├── scanner/              # 文件扫描
│   ├── selector/             # 交互式选择
│   ├── generator/            # 代码生成
│   └── updater/              # 代码更新
└── docs/                      # 文档
```

## 🐛 Bug 修复

- **修复 gRPC 服务注册问题**：修复 `RegisterXXXServer` 调用未正确添加到 `NewGRPCServer` 函数体的问题
- **修复花括号匹配**：新增 `FindMatchingCurlyBrace` 函数正确处理函数体边界

## 📦 安装方式

### 方式一：从源码安装（推荐）

```bash
go install github.com/nigiwen/gen-handler@v1.3.0
```

### 方式二：使用预编译二进制

下载对应平台的二进制文件：

- **Linux amd64**: [gen-handler_v1.3.0_linux_amd64.tar.gz](gen-handler_v1.3.0_linux_amd64.tar.gz)
- **Linux arm64**: [gen-handler_v1.3.0_linux_arm64.tar.gz](gen-handler_v1.3.0_linux_arm64.tar.gz)
- **macOS amd64**: [gen-handler_v1.3.0_darwin_amd64.tar.gz](gen-handler_v1.3.0_darwin_amd64.tar.gz)
- **macOS arm64**: [gen-handler_v1.3.0_darwin_arm64.tar.gz](gen-handler_v1.3.0_darwin_arm64.tar.gz)
- **Windows amd64**: [gen-handler_v1.3.0_windows_amd64.zip](gen-handler_v1.3.0_windows_amd64.zip)

## 📝 完整变更日志

详见 [CHANGELOG.md](CHANGELOG.md)
