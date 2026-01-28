# Release v1.2.0

## 🎉 新功能

### 数据同步功能 (Data Sync)

工具现在支持 Data 层的自动同步，进一步提升开发效率：

- ✨ **自动扫描实体**：自动扫描 `internal/model/entity` 目录。
- ✨ **自动生成 dbset**：生成包含 `BeforeCreate` (雪花 ID) 的类型别名文件。
- ✨ **自动生成 repo**：生成标准的存储库结构和构造函数。
- ✨ **自动注册**：自动将新生成的 Repo 注册到 `data.go` 的 `ProviderSet`。
- ✨ **自动 wire**：注册完成后自动运行 `wire` 更新依赖注入。

### 使用示例

```bash
# 运行工具，选择 "Data Sync (Entity -> Repo)"
gen-handler
```

## 📦 安装方式

### 方式一：从源码安装（推荐）

```bash
go install github.com/nigiwen/gen-handler@v1.2.0
```

### 方式二：使用预编译二进制

下载对应平台的二进制文件：

- **Linux amd64**: [gen-handler_1.2.0_linux_amd64.tar.gz](gen-handler_1.2.0_linux_amd64.tar.gz)
- **Linux arm64**: [gen-handler_1.2.0_linux_arm64.tar.gz](gen-handler_1.2.0_linux_arm64.tar.gz)
- **macOS amd64**: [gen-handler_1.2.0_darwin_amd64.tar.gz](gen-handler_1.2.0_darwin_amd64.tar.gz)
- **macOS arm64**: [gen-handler_1.2.0_darwin_arm64.tar.gz](gen-handler_1.2.0_darwin_arm64.tar.gz)
- **Windows amd64**: [gen-handler_1.2.0_windows_amd64.zip](gen-handler_1.2.0_windows_amd64.zip)

## 📝 完整变更日志

详见 [CHANGELOG.md](CHANGELOG.md)
