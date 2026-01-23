# Release v1.1.0

## 🎉 新功能

### 自动配置功能

工具现在支持自动配置，大大简化了使用：

- ✨ **自动读取 go.mod**：自动从项目根目录的 `go.mod` 文件读取 module 路径
- ✨ **自动生成 proto-dir**：根据 module 路径自动生成 proto 目录路径
- ✨ **自动生成 wire-dir**：根据 module 路径自动生成 wire 目录路径

### 使用示例

现在使用工具更加简单：

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

## 📦 安装方式

### 方式一：从源码安装（推荐）

```bash
go install github.com/nigiwen/gen-handler@v1.1.0
```

### 方式二：使用预编译二进制

下载对应平台的二进制文件：

- **Linux amd64**: [gen-handler_1.1.0_linux_amd64.tar.gz](gen-handler_1.1.0_linux_amd64.tar.gz)
- **Linux arm64**: [gen-handler_1.1.0_linux_arm64.tar.gz](gen-handler_1.1.0_linux_arm64.tar.gz)
- **macOS amd64**: [gen-handler_1.1.0_darwin_amd64.tar.gz](gen-handler_1.1.0_darwin_amd64.tar.gz)
- **macOS arm64**: [gen-handler_1.1.0_darwin_arm64.tar.gz](gen-handler_1.1.0_darwin_arm64.tar.gz)
- **Windows amd64**: [gen-handler_1.1.0_windows_amd64.zip](gen-handler_1.1.0_windows_amd64.zip)

**Linux/macOS 解压使用**：
```bash
tar -xzf gen-handler_1.1.0_linux_amd64.tar.gz
./gen-handler -help
```

**Windows 解压使用**：
```powershell
# 解压 zip 文件
# 运行 gen-handler.exe
```

## 📝 完整变更日志

详见 [CHANGELOG.md](CHANGELOG.md)

## 🔗 相关链接

- [GitHub 仓库](https://github.com/nigiwen/gen-handler)
- [使用文档](README.md)
- [测试指南](TEST.md)
