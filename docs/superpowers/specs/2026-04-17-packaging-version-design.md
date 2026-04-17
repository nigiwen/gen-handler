# 打包脚本与版本输出设计

**日期：** 2026-04-17
**状态：** 待评审

## 目标

修复当前仓库的打包链路，使 Windows 和 Unix 打包脚本都能基于当前代码正常产出构建包，并让注入的版本号可以被 CLI 直接输出验证。

本次设计覆盖两件事：

- 修复 `build.bat` / `build.sh`
- 为 CLI 增加 `-version` 输出入口

## 当前问题

当前打包能力存在三个明确问题：

1. `build.bat` 文件编码异常，直接在当前 Windows 环境执行时会被 `cmd` 误解析，导致脚本无法完成打包。
2. `build.bat` 与 `build.sh` 都使用 `-ldflags "-X main.version=..."` 注入版本，但当前 `main.go` 中没有对应的包级 `version` 变量，导致版本注入没有实际效果。
3. CLI 没有版本查看入口，用户即使成功注入版本，也无法直接通过命令确认最终二进制携带的版本值。

## 设计原则

- 保持现有发布入口不变，继续使用 `build.bat <version>` 和 `build.sh <version>`
- 未传版本号时，统一回退到 `dev`，不依赖 `git describe`
- Windows 脚本使用 ASCII 内容，避免编码再次破坏批处理执行
- 版本输出入口要轻量，不引入新的复杂子命令体系
- 优先通过自动化测试锁定 CLI 行为，再修改实现

## 用户可见行为

实现完成后，用户可见行为如下：

1. 执行 `gen-handler -version` 时，CLI 直接输出当前版本并退出，退出码为 `0`
2. 未注入版本号时，默认输出 `dev`
3. 执行 `build.bat v1.2.3` 时，生成：
   - `dist/windows_amd64/gen-handler.exe`
   - `dist/gen-handler_v1.2.3_windows_amd64.zip`
4. 执行 `build.sh v1.2.3` 时，继续生成各平台归档文件
5. 未传版本号时：
   - `build.bat` 默认打包 `dev`
   - `build.sh` 默认打包 `dev`

## 方案选择

本次采用“轻量版本标志 + 修复现有脚本”的方案，而不是新增 `version` 子命令。

推荐原因：

- 与当前 CLI 的 flag 风格一致，改动面更小
- 不需要新增命令路由逻辑
- 能最直接验证 `ldflags` 注入是否生效

不采用 `version` 子命令的原因：

- 对当前工具来说收益很小
- 会让命令分发层比需求本身更重

## 架构设计

### 1. CLI 版本输出

涉及文件：

- `main.go`

设计：

- 增加包级变量：
  - `var version = "dev"`
- 增加 `-version` 布尔参数
- 在解析参数后，若传入 `-version`：
  - 输出 `version`
  - 直接退出 `0`

这样可保证：

- 本地直接 `go build .` 时，默认版本是 `dev`
- 发布脚本通过 `-ldflags "-X main.version=..."` 注入时，`-version` 能反映最终值

### 2. Windows 打包脚本

涉及文件：

- `build.bat`

设计：

- 用纯 ASCII 重写脚本
- 保留现有入口：`build.bat <version>`
- 若未传版本号，直接使用 `dev`
- 清理并重建 `dist/`
- 构建 `dist/windows_amd64/gen-handler.exe`
- 使用 PowerShell `Compress-Archive` 生成 `dist/gen-handler_<version>_windows_amd64.zip`

不再从脚本里调用 `git describe`，避免：

- 违反当前仓库“未经用户许可不执行 git”的约束
- 在无 tag 或无 git 环境下出现不可预期差异

### 3. Unix 打包脚本

涉及文件：

- `build.sh`

设计：

- 保留现有多平台打包结构
- 若未传版本号，直接使用 `dev`
- 移除 `git describe` 默认分支
- 保留现有平台矩阵：
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- 继续按平台生成 `.tar.gz` 或 `.zip`

### 4. 测试

涉及文件：

- `main_test.go`（建议新增）

测试重点：

- 默认版本变量为 `dev`
- 当命令收到 `-version` 时，只输出版本并成功退出

如果当前 `main.go` 结构不便于直接测试，可接受的小重构是：

- 抽一个可测试的 `run(args []string, stdout, stderr)` 或等价辅助函数

重构边界仅限于提升 `-version` 行为可测试性，不扩展到其他命令结构整理。

## 错误处理

- 打包脚本中的任何 `go build` 失败都应使脚本直接失败
- 压缩归档失败应使脚本直接失败
- `-version` 路径不依赖 `go.mod`、proto 路径、wire 路径等其他运行时配置

## 验证策略

本次改动的验证分两层：

### 自动化验证

- `go test ./...`
- `go build .`

### 脚本验证

- 在 Windows 环境执行：`build.bat dev`
- 如当前环境具备可执行 bash，再执行：`build.sh dev`
- 之后运行打出的二进制：
  - `dist/windows_amd64/gen-handler.exe -version`
  - 预期输出：`dev`

## 范围边界

本次设计明确包含：

- `-version` 标志
- 默认版本值 `dev`
- `build.bat` 修复
- `build.sh` 修复
- 构建产物命名与版本号注入联动

本次设计明确不包含：

- 新增 `version` 子命令
- 自动从 `git tag` 推导版本
- 发布流程、Release 文档或 changelog 的同步更新
- 多平台构建矩阵的扩展或裁剪

## 结论

最合适的方案是：

- 在 `main.go` 中增加包级 `version` 变量和 `-version` 标志
- 重写 `build.bat`，修复编码与默认版本逻辑
- 修正 `build.sh` 的默认版本逻辑
- 用测试和实际打包命令共同验证版本注入与产物生成

这套方案改动面小，但能把“脚本可执行”“版本可注入”“版本可验证”三个问题一起闭环。
