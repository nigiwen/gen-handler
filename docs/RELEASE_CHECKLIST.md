# 发布检查清单 v1.2.0

## ✅ 发布前检查

- [ ] 代码已提交到 git
- [ ] 已创建版本标签 v1.2.0
- [ ] 已编译所有平台的二进制文件
- [ ] 已更新 CHANGELOG.md
- [ ] 已创建 RELEASE.md 发布说明

## 📦 打包文件清单

以下文件已生成在 `dist/` 目录：

### Linux
- [ ] `gen-handler_v1.2.0_linux_amd64.tar.gz`
- [ ] `gen-handler_v1.2.0_linux_arm64.tar.gz`

### macOS
- [ ] `gen-handler_v1.2.0_darwin_amd64.tar.gz`
- [ ] `gen-handler_v1.2.0_darwin_arm64.tar.gz`

### Windows
- [ ] `gen-handler_v1.2.0_windows_amd64.tar.gz`

**注意**：Windows 文件是 tar.gz 格式（因为系统没有 zip 命令）。如果需要 zip 格式，可以在有 zip 命令的系统上重新打包，或使用 Windows 系统运行 `build.bat`。

## 🚀 发布步骤

### 1. 推送代码和标签到 GitHub

```bash
cd /workspace/bsi/axis/tools/gen-handler

# 推送代码
git push origin main

# 推送标签
git push origin v1.2.0
```

### 2. 在 GitHub 创建 Release

1. 访问：https://github.com/nigiwen/gen-handler/releases/new
2. 选择标签：`v1.2.0`
3. 标题：`v1.2.0: 数据同步功能`
4. 描述：复制 `RELEASE.md` 的内容
5. 上传文件：上传 `dist/` 目录下的所有 `.tar.gz` 文件

### 3. 发布说明模板

```markdown
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

下载对应平台的二进制文件并解压使用。

## 📝 完整变更日志

详见 [CHANGELOG.md](CHANGELOG.md)
```

## ✅ 发布后验证

发布成功后，验证安装：

```bash
# 测试安装
go install github.com/nigiwen/gen-handler@v1.2.0

# 验证版本
gen-handler -help
```

## 📋 文件位置

- 打包文件：`dist/` 目录
- 发布说明：`RELEASE.md`
- 变更日志：`CHANGELOG.md`
- 编译脚本：`build.sh` (Linux/macOS) 和 `build.bat` (Windows)
