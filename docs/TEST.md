# 本地测试指南

## 🧪 测试自动读取 go.mod 功能

### 方式一：在工具目录直接测试（快速验证）

```bash
# 1. 进入工具目录
cd /workspace/bsi/axis/tools/gen-handler

# 2. 编译工具
go build -o gen-handler

# 3. 测试帮助信息
./gen-handler -help

# 4. 测试自动读取 go.mod（在工具目录本身）
# 工具会读取当前目录的 go.mod，应该显示：github.com/nigiwen/gen-handler
./gen-handler -proto-dir ./test-proto -output-dir ./test-output
```

### 方式二：在实际项目中使用（推荐 ⭐）

```bash
# 1. 编译工具（在工具目录）
cd /workspace/bsi/axis/tools/gen-handler
go build -o gen-handler

# 2. 进入你的实际项目（比如 devopsx）
cd /workspace/bsi/axis/devopsx  # 或你的项目根目录

# 3. 使用绝对路径运行编译后的工具（不指定 -module 参数）
# 工具会自动从当前目录的 go.mod 读取 module 路径
/workspace/bsi/axis/tools/gen-handler/gen-handler \
  -proto-dir ./internal/proto/axis/devopsx \
  -output-dir ./api/grpc \
  -core-dir ./core \
  -wire-dir ./cmd/devopsx

# 应该会看到输出：
# 📦 从 go.mod 自动读取 module: bsi/axis/devopsx
```

**或者使用相对路径**（如果在项目根目录）：
```bash
../tools/gen-handler/gen-handler \
  -proto-dir ./internal/proto/axis/devopsx \
  -output-dir ./api/grpc \
  -core-dir ./core \
  -wire-dir ./cmd/devopsx
```

### 方式三：使用 go run（需要在工具目录内）

⚠️ **注意**：`go run` 不能直接运行项目外的目录，会报错 "directory outside main module"。

**正确的做法**：

```bash
# 1. 进入工具目录
cd /workspace/bsi/axis/tools/gen-handler

# 2. 使用 go run，但需要指定项目目录作为工作目录
# 使用 -C 参数（Go 1.20+）或者先 cd 到项目目录
cd /workspace/bsi/axis/devopsx
go run /workspace/bsi/axis/tools/gen-handler \
  -proto-dir ./internal/proto/axis/devopsx \
  -output-dir ./api/grpc \
  -core-dir ./core \
  -wire-dir ./cmd/devopsx

# 或者使用环境变量指定工作目录（如果工具支持）
```

**推荐**：直接使用方法二（先编译），更简单可靠。

### 方式四：安装到系统后测试

```bash
# 1. 在工具目录安装到系统
cd /workspace/bsi/axis/tools/gen-handler
go install

# 2. 进入你的项目目录
cd /workspace/bsi/axis/devopsx

# 3. 直接使用命令（工具会自动从 go.mod 读取）
gen-handler \
  -proto-dir ./internal/proto/axis/devopsx \
  -output-dir ./api/grpc \
  -core-dir ./core \
  -wire-dir ./cmd/devopsx
```

## ✅ 验证测试点

### 1. 测试自动读取功能

```bash
# 在项目根目录运行（不指定 -module）
cd /workspace/bsi/axis/devopsx
go run /workspace/bsi/axis/tools/gen-handler -proto-dir ./internal/proto/axis/devopsx

# 预期输出应该包含：
# 📦 从 go.mod 自动读取 module: bsi/axis/devopsx
```

### 2. 测试手动指定 module（覆盖自动读取）

```bash
# 手动指定 module，应该使用指定的值
go run /workspace/bsi/axis/tools/gen-handler \
  -module custom/module/path \
  -proto-dir ./internal/proto/axis/devopsx

# 预期：不会显示"从 go.mod 自动读取"的提示，直接使用 custom/module/path
```

### 3. 测试找不到 go.mod 的情况

```bash
# 在一个没有 go.mod 的目录运行
cd /tmp
go run /workspace/bsi/axis/tools/gen-handler -proto-dir ./test

# 预期输出：
# ❌ 未找到 go.mod 文件，且未指定 -module 参数
# 💡 请使用 -module 参数指定 Go 模块路径，或在项目根目录运行此工具
```

### 4. 测试向上查找 go.mod

```bash
# 在项目子目录运行（工具会向上查找 go.mod）
cd /workspace/bsi/axis/devopsx/internal/proto
go run /workspace/bsi/axis/tools/gen-handler -proto-dir ./axis/devopsx

# 预期：工具会向上查找到项目根目录的 go.mod 并读取
```

## 🔍 调试技巧

### 查看工具是否找到 go.mod

如果工具没有自动读取，可以：

1. **检查当前目录**：
   ```bash
   pwd
   ls -la go.mod  # 确认 go.mod 存在
   ```

2. **检查 go.mod 内容**：
   ```bash
   head -1 go.mod  # 应该看到：module bsi/axis/devopsx
   ```

3. **手动测试读取函数**：
   可以在 `main.go` 中添加调试输出，查看查找过程

### 常见问题

1. **问题**：工具提示找不到 go.mod
   - **解决**：确保在项目根目录运行，或确保项目根目录有 go.mod 文件

2. **问题**：读取的 module 路径不对
   - **解决**：检查 go.mod 第一行的 module 声明是否正确

3. **问题**：工具没有自动读取
   - **解决**：检查是否手动指定了 `-module` 参数（即使为空字符串也会覆盖自动读取）

## 📝 测试清单

- [ ] 在项目根目录测试自动读取
- [ ] 在子目录测试向上查找
- [ ] 测试手动指定 module 参数
- [ ] 测试找不到 go.mod 的错误提示
- [ ] 验证生成的代码中 import 路径正确
