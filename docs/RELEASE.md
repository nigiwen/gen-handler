# Release v1.4.0

## 重点更新

### Handler 增量补全

本版本将 `handler` 命令从“只处理缺失文件”调整为“默认列出全部服务，按选择结果创建或补全代码”：

- 默认展示全部解析出的 gRPC 服务
- `api/grpc/<service>.go` 不存在时创建完整文件，存在时只补缺失方法
- `core/<service>.go` 不存在时创建完整文件，存在时只补缺失方法
- 已有实现保持不变，不做整文件覆盖
- 如果发现同名方法但签名与当前 proto 不一致，会直接报错并停止修改

### ProviderSet 与 wire 行为收敛

生成流程现在按“是否新建文件”决定后续步骤：

- 只有本次新建了 handler 文件，才更新 `api/grpc/grpc.go`
- 只有本次新建了 core 文件，才更新 `core/core.go`
- 只有本次创建了任意文件，才执行 `wire`

这让补全已有服务时不会再引入无意义的 ProviderSet 变更和额外 `wire` 执行。

### 统一终端 UI

`handler` 与 `data` 命令现在共享统一的 Bubble Tea TUI：

- `Space` 多选
- `/` 搜索过滤
- 执行进度与结果汇总
- 非 TTY 环境自动回退到编号输入
- 长列表会跟随光标自动滚动
- 当前项高亮更明显

### Data 与版本能力

- `data` 命令现在只从 `internal/model/entity/*.gen.go` 发现候选实体
- 如缺少 `internal/model/entity/<name>.go`，会自动补占位文件
- 不再生成 `data/dbset/*.go`
- 新增 `-version` 参数，支持输出构建注入版本

## 安装方式

### 从源码安装

```bash
go install github.com/nigiwen/gen-handler@v1.4.0
```

### 使用预编译二进制

当前已生成的发布归档如下：

- `gen-handler_1.4.0_linux_amd64.tar.gz`
- `gen-handler_1.4.0_linux_arm64.tar.gz`
- `gen-handler_1.4.0_darwin_amd64.tar.gz`
- `gen-handler_1.4.0_darwin_arm64.tar.gz`
- `gen-handler_1.4.0_windows_amd64.tar.gz`

## 发布备注

- Git 标签建议使用：`v1.4.0`
- 当前已打包产物的归档名与二进制内置版本串为：`1.4.0`
- 发布 GitHub Release 时，正文可直接使用本文内容

## 完整变更日志

详见 [CHANGELOG.md](CHANGELOG.md)
