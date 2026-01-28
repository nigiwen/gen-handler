# 发布检查清单

## ✅ 发布前检查

- [ ] 修改 `go.mod` 中的模块路径为你的 GitHub 仓库路径
- [ ] 运行 `go mod tidy` 确保依赖正确
- [ ] 运行 `go build` 确保编译成功
- [ ] 测试工具功能是否正常
- [ ] 更新 `README.md` 中的安装路径（如果已修改）
- [ ] 检查所有文件是否已提交

## 📋 发布步骤

### 1. 在 GitHub 创建仓库

```bash
# 在 GitHub 网页创建新仓库：gen-handler
# 设置为 Public（如果希望公开）或 Private
```

### 2. 准备本地仓库

```bash
# 进入工具目录
cd /workspace/bsi/axis/devopsx/tools/gen-handler

# 初始化 git（如果还没有）
git init

# 修改 go.mod 中的模块路径
# 将 github.com/nigiwen/gen-handler 改为你的实际路径

# 初始化依赖
go mod tidy

# 测试编译
go build -o gen-handler
./gen-handler -help  # 测试是否正常

# 添加文件
git add .
git commit -m "feat: initial release"
```

### 3. 推送到 GitHub

```bash
# 添加远程仓库
git remote add origin https://github.com/nigiwen/gen-handler.git

# 推送代码
git branch -M main
git push -u origin main

# 创建标签（版本）
git tag v1.0.0
git push origin v1.0.0
```

### 4. 验证安装

```bash
# 在另一个目录测试安装
go install github.com/nigiwen/gen-handler@latest

# 验证
gen-handler -help
```

## 🗑️ 从当前项目删除

发布成功后，可以从当前项目中删除：

```bash
# 1. 删除工具目录
rm -rf /workspace/bsi/axis/devopsx/tools/gen-handler

# 2. 更新 Makefile
# 编辑 Makefile，将：
#   @go run ./tools/gen-handler
# 改为：
#   @gen-handler

# 3. 确保团队成员安装工具
# 在项目 README 或文档中说明：
#   安装工具：go install github.com/nigiwen/gen-handler@latest
```

## 📝 更新项目文档

在项目主 README 中添加：

```markdown
## 开发工具

### gen-handler

用于生成 gRPC Handler 和 Core Service 代码的工具。

**安装：**
```bash
go install github.com/nigiwen/gen-handler@latest
```

**使用：**
```bash
make gen-handler
# 或
gen-handler
```
```
