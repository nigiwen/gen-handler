# 发布到 GitHub 指南

## 📦 发布步骤

### 1. 准备独立仓库

```bash
# 1. 在 GitHub 创建新仓库（如：gen-handler）
# 2. 克隆到本地
git clone https://github.com/nigiwen/gen-handler.git
cd gen-handler

# 3. 从当前项目复制文件
cp -r /workspace/bsi/axis/devopsx/tools/gen-handler/* .

# 4. 修改 go.mod 中的模块路径（已修改为 github.com/nigiwen/gen-handler）
# 5. 初始化依赖
go mod tidy

# 6. 测试编译
go build -o gen-handler

# 7. 提交并推送
git add .
git commit -m "feat: initial release of gen-handler tool"
git tag v1.0.0
git push origin main
git push origin v1.0.0
```

### 2. 更新 go.mod 模块路径

**重要**：在发布前，需要将 `go.mod` 中的模块路径改为你的 GitHub 仓库路径：

```go
module github.com/nigiwen/gen-handler  // 改为你的实际路径
```

### 3. 不需要编译打包

✅ **Go 工具不需要预编译**，用户可以直接从源码安装：

```bash
# 用户安装方式（自动编译）
go install github.com/nigiwen/gen-handler@latest
```

Go 会自动：
- 下载源码
- 编译二进制
- 安装到 `$GOPATH/bin` 或 `$GOBIN`

### 4. 可选：发布 Release（推荐）

虽然不需要编译，但可以发布 Release 提供：
- 版本说明
- 变更日志
- 预编译的二进制（可选，方便没有 Go 环境的用户）

```bash
# 编译不同平台的二进制（可选）
GOOS=linux GOARCH=amd64 go build -o gen-handler-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o gen-handler-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o gen-handler-windows-amd64.exe

# 在 GitHub 创建 Release，上传这些二进制文件
```

## 📥 用户安装和使用

### 安装

```bash
# 方式一：从 GitHub 安装（推荐）
go install github.com/nigiwen/gen-handler@latest

# 方式二：从特定版本安装
go install github.com/nigiwen/gen-handler@v1.0.0

# 方式三：从最新 main 分支安装
go install github.com/nigiwen/gen-handler@main
```

### 使用

```bash
# 查看帮助
gen-handler -help

# 使用默认配置
gen-handler

# 自定义配置
gen-handler \
  -proto-dir ./internal/proto/axis/devopsx \
  -output-dir ./api/grpc \
  -core-dir ./core \
  -wire-dir ./cmd/devopsx \
  -module bsi/axis/devopsx
```

### 验证安装

```bash
# 检查是否安装成功
which gen-handler
gen-handler -help
```

## 🔄 从当前项目删除

发布到 GitHub 后，可以从当前项目中删除：

```bash
# 1. 删除工具目录
rm -rf /workspace/bsi/axis/devopsx/tools/gen-handler

# 2. 更新 Makefile（如果引用了本地路径）
# 将：
# gen-handler:
# 	@go run ./tools/gen-handler
# 改为：
# gen-handler:
# 	@gen-handler

# 3. 或者直接使用全局安装的工具
# 在 Makefile 中直接调用 gen-handler 命令即可
```

## 📝 更新 Makefile

发布后，更新项目的 `Makefile`：

```makefile
.PHONY: gen-handler
# generate grpc handlers
gen-handler:
	@echo "生成 gRPC handler 文件..."
	@gen-handler
	@echo "✅ handler 文件生成完成"
```

**注意**：确保团队成员都已安装工具：
```bash
go install github.com/nigiwen/gen-handler@latest
```

## 🎯 最佳实践

1. **版本管理**：使用语义化版本（v1.0.0, v1.1.0 等）
2. **变更日志**：维护 CHANGELOG.md
3. **CI/CD**：可以设置 GitHub Actions 自动编译和发布
4. **文档**：README.md 要详细说明使用方法
5. **示例**：提供使用示例和配置说明

## ⚠️ 注意事项

1. **模块路径**：确保 `go.mod` 中的路径与 GitHub 仓库路径一致
2. **依赖管理**：运行 `go mod tidy` 确保依赖正确
3. **测试**：发布前在不同项目结构下测试
4. **向后兼容**：保持默认参数与当前项目兼容
