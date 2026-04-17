# 发布检查清单 v1.4.0

## 发布前检查

- [ ] 代码改动已确认完整
- [ ] `docs/CHANGELOG.md` 已更新到 `1.4.0`
- [ ] `docs/RELEASE.md` 已更新为 `v1.4.0` 发布说明
- [ ] 已确认本次 Git 标签使用 `v1.4.0`
- [ ] 已确认预编译产物已生成

## 预编译产物

当前 `dist/` 目录中的发布归档为：

- [ ] `gen-handler_1.4.0_linux_amd64.tar.gz`
- [ ] `gen-handler_1.4.0_linux_arm64.tar.gz`
- [ ] `gen-handler_1.4.0_darwin_amd64.tar.gz`
- [ ] `gen-handler_1.4.0_darwin_arm64.tar.gz`
- [ ] `gen-handler_1.4.0_windows_amd64.tar.gz`

## 版本说明

- Git 标签：`v1.4.0`
- 当前已打包归档名：`1.4.0`
- 当前已打包二进制 `-version` 输出：`1.4.0`

如果你希望 Git 标签、归档名和程序 `-version` 全部都带 `v` 前缀，需要重新按 `v1.4.0` 重新打包；当前这批产物保持 `1.4.0` 更稳妥。

## GitHub Release 步骤

1. 创建或选择标签：`v1.4.0`
2. 打开 Release 页面：`https://github.com/nigiwen/gen-handler/releases/new`
3. Release 标题建议：`v1.4.0: handler 补全与统一 TUI`
4. Release 正文：直接粘贴 `docs/RELEASE.md`
5. 上传 `dist/` 下的 5 个归档文件

## 发布后验证

- [ ] `go install github.com/nigiwen/gen-handler@v1.4.0` 可正常安装
- [ ] 二进制 `-version` 输出符合预期
- [ ] Release 页面中的归档文件可正常下载
