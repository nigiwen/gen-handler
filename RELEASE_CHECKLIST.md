# 发布检查清单 v1.1.0

## ✅ 发布前检查

- [x] 代码已提交到 git
- [x] 已创建版本标签 v1.1.0
- [x] 已编译所有平台的二进制文件
- [x] 已更新 CHANGELOG.md
- [x] 已创建 RELEASE.md 发布说明

## 📦 打包文件清单

以下文件已生成在 `dist/` 目录：

### Linux
- [x] `gen-handler_v1.1.0_linux_amd64.tar.gz` (3.0M)
- [x] `gen-handler_v1.1.0_linux_arm64.tar.gz` (2.8M)

### macOS
- [x] `gen-handler_v1.1.0_darwin_amd64.tar.gz` (2.9M)
- [x] `gen-handler_v1.1.0_darwin_arm64.tar.gz` (2.7M)

### Windows
- [x] `gen-handler_v1.1.0_windows_amd64.tar.gz` (3.2M)

**注意**：Windows 文件是 tar.gz 格式（因为系统没有 zip 命令）。如果需要 zip 格式，可以在有 zip 命令的系统上重新打包，或使用 Windows 系统运行 `build.bat`。

## 🚀 发布步骤

### 1. 推送代码和标签到 GitHub

```bash
cd /workspace/bsi/axis/tools/gen-handler

# 推送代码
git push origin main

# 推送标签
git push origin v1.1.0
```

### 2. 在 GitHub 创建 Release

1. 访问：https://github.com/nigiwen/gen-handler/releases/new
2. 选择标签：`v1.1.0`
3. 标题：`v1.1.0: 自动配置功能`
4. 描述：复制 `RELEASE.md` 的内容
5. 上传文件：上传 `dist/` 目录下的所有 `.tar.gz` 文件

### 3. 发布说明模板

```markdown
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

## 📦 安装方式

### 方式一：从源码安装（推荐）

```bash
go install github.com/nigiwen/gen-handler@v1.1.0
```

### 方式二：使用预编译二进制

下载对应平台的二进制文件并解压使用。

## 📝 完整变更日志

详见 [CHANGELOG.md](CHANGELOG.md)
```

## ✅ 发布后验证

发布成功后，验证安装：

```bash
# 测试安装
go install github.com/nigiwen/gen-handler@v1.1.0

# 验证版本
gen-handler -help
```

## 📋 文件位置

- 打包文件：`dist/` 目录
- 发布说明：`RELEASE.md`
- 变更日志：`CHANGELOG.md`
- 编译脚本：`build.sh` (Linux/macOS) 和 `build.bat` (Windows)
