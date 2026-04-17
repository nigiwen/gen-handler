# Packaging Version Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 `build.bat` / `build.sh` 并为 CLI 增加 `-version` 输出入口，使版本注入和打包产物都能按当前代码正常工作。

**Architecture:** 在 `main.go` 中增加包级 `version = "dev"`，并抽出一个可测试的 `run(args, stdout, stderr) int` 入口，让 `-version` 行为可以先测试再实现。打包脚本保留现有入口和产物命名，但移除对 `git describe` 的依赖，统一默认版本为 `dev`，Windows 脚本改为纯 ASCII 以避免编码问题。

**Tech Stack:** Go、标准库 `flag` / `os` / `fmt`、Windows batch、bash、`go test`、`go build`、PowerShell `Compress-Archive`

---

## File Map

- Modify: `main.go`
  - 增加包级 `version`、`-version` 标志、可测试的 `run` 入口
- Create: `main_test.go`
  - 覆盖 `-version` 路径和注入版本值的输出行为
- Modify: `build.bat`
  - 改为 ASCII 脚本，移除 `git describe`，默认版本改为 `dev`
- Modify: `build.sh`
  - 移除 `git describe`，默认版本改为 `dev`

> Note: 不包含任何 `git add` / `git commit` 步骤。当前仓库规则要求所有 `git` 操作都必须由用户明确授权。

### Task 1: 为 CLI 增加可测试的 `-version` 输出

**Files:**
- Modify: `main.go`
- Create: `main_test.go`

- [ ] **Step 1: 写失败测试，锁定默认版本和注入版本的输出行为**

在 `main_test.go` 中加入：

```go
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsVersionAndExitsSuccess(t *testing.T) {
	originalVersion := version
	t.Cleanup(func() {
		version = originalVersion
	})

	version = "dev"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"gen-handler", "-version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if got := strings.TrimSpace(stdout.String()); got != "dev" {
		t.Fatalf("expected version output dev, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunPrintsInjectedVersion(t *testing.T) {
	originalVersion := version
	t.Cleanup(func() {
		version = originalVersion
	})

	version = "v1.2.3"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"gen-handler", "-version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if got := strings.TrimSpace(stdout.String()); got != "v1.2.3" {
		t.Fatalf("expected version output v1.2.3, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}
```

- [ ] **Step 2: 跑测试，确认先红**

Run: `go test . -run "TestRunPrints(VersionAndExitsSuccess|InjectedVersion)" -v`

Expected: FAIL，原因是当前没有包级 `version` 变量，也没有 `run` 函数。

- [ ] **Step 3: 写最小实现，让 `-version` 可用且可测**

将 `main.go` 重构为下面的结构：

```go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nigiwen/gen-handler/cmd"
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.SetOutput(stderr)

	var (
		protoDir   = flagSet.String("proto-dir", "", "proto 生成的 grpc 文件目录（未指定时自动从 module 生成）")
		outputDir  = flagSet.String("output-dir", "./api/grpc", "handler 输出目录")
		coreDir    = flagSet.String("core-dir", "./core", "core service 输出目录")
		wireDir    = flagSet.String("wire-dir", "", "wire 命令执行目录（未指定时自动从 module 生成）")
		modulePath = flagSet.String("module", "", "Go 模块路径（用于生成 import 路径，未指定时自动从 go.mod 读取）")
		showHelp   = flagSet.Bool("help", false, "显示帮助信息")
		showVer    = flagSet.Bool("version", false, "显示版本信息")
	)

	flagSet.Usage = func() {
		fmt.Fprintf(stderr, "Axis 开发工具集\n\n")
		fmt.Fprintf(stderr, "用法: %s <命令> [选项]\n\n", args[0])
		fmt.Fprintf(stderr, "可用命令:\n")
		fmt.Fprintf(stderr, "  handler    生成 gRPC Handler / Core / wire\n")
		fmt.Fprintf(stderr, "  data       同步 Data 层 (*.gen.go -> entity & repo)\n\n")
		fmt.Fprintf(stderr, "选项:\n")
		flagSet.PrintDefaults()
		fmt.Fprintf(stderr, "\n示例:\n")
		fmt.Fprintf(stderr, "  %s handler\n", args[0])
		fmt.Fprintf(stderr, "  %s data\n", args[0])
	}

	if len(args) < 2 {
		flagSet.Usage()
		return 1
	}

	if args[1] == "-version" || args[1] == "--version" {
		fmt.Fprintln(stdout, version)
		return 0
	}

	command := args[1]
	parsedArgs := append([]string{args[0]}, args[2:]...)
	if err := flagSet.Parse(parsedArgs[1:]); err != nil {
		return 1
	}

	if *showVer {
		fmt.Fprintln(stdout, version)
		return 0
	}
	if *showHelp {
		flagSet.Usage()
		return 0
	}

	finalModulePath := *modulePath
	if finalModulePath == "" {
		workDir, err := os.Getwd()
		if err != nil {
			workDir = "."
		}
		if module, found := util.ReadModuleFromGoMod(workDir); found {
			finalModulePath = module
			fmt.Fprintf(stdout, "📦 从 go.mod 自动读取 module: %s\n", finalModulePath)
		} else {
			fmt.Fprintln(stdout, "❌ 未找到 go.mod 文件，且未指定 -module 参数")
			fmt.Fprintln(stdout, "💡 请使用 -module 参数指定 Go 模块路径，或在项目根目录运行此工具")
			return 1
		}
	}

	config := types.Config{
		ModulePath: finalModulePath,
		OutputDir:  *outputDir,
		CoreDir:    *coreDir,
	}

	finalWireDir := *wireDir
	if finalWireDir == "" {
		finalWireDir = util.GenerateWireDirFromModule(config.ModulePath)
		fmt.Fprintf(stdout, "🔧 从 module 自动生成 wire-dir: %s\n", finalWireDir)
	}
	config.WireDir = finalWireDir

	switch command {
	case "handler":
		if err := cmd.RunHandlerCommand(config, *protoDir); err != nil {
			fmt.Fprintf(stdout, "❌ %v\n", err)
			return 1
		}
	case "data":
		cmd.RunDataCommand(config)
	default:
		fmt.Fprintf(stdout, "❌ 未知命令: %s\n", command)
		flagSet.Usage()
		return 1
	}

	return 0
}
```

- [ ] **Step 4: 再跑测试，确认转绿**

Run: `go test . -run "TestRunPrints(VersionAndExitsSuccess|InjectedVersion)" -v`

Expected: PASS

### Task 2: 修复 Windows 打包脚本

**Files:**
- Modify: `build.bat`

- [ ] **Step 1: 运行当前脚本，确认问题可复现**

Run: `.\build.bat dev`

Expected: FAIL，出现批处理解析错误，证明当前脚本确实被编码问题破坏。

- [ ] **Step 2: 用 ASCII 重写 `build.bat`**

将 `build.bat` 全量替换为：

```bat
@echo off
setlocal

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=dev"

set "APP_NAME=gen-handler"
set "BUILD_DIR=dist"
set "PLATFORM_DIR=%BUILD_DIR%\windows_amd64"
set "OUTPUT_PATH=%PLATFORM_DIR%\%APP_NAME%.exe"
set "ARCHIVE_PATH=%BUILD_DIR%\%APP_NAME%_%VERSION%_windows_amd64.zip"

echo [INFO] Building %APP_NAME% %VERSION%

if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
mkdir "%PLATFORM_DIR%"

go build -ldflags "-X main.version=%VERSION%" -o "%OUTPUT_PATH%" .
if errorlevel 1 exit /b 1

powershell -NoProfile -Command "Compress-Archive -Path '%OUTPUT_PATH%' -DestinationPath '%ARCHIVE_PATH%' -Force"
if errorlevel 1 exit /b 1

echo [INFO] Created %ARCHIVE_PATH%
exit /b 0
```

- [ ] **Step 3: 运行脚本，确认 Windows 包能正常生成**

Run: `.\build.bat dev`

Expected: PASS，并生成：

- `dist\windows_amd64\gen-handler.exe`
- `dist\gen-handler_dev_windows_amd64.zip`

### Task 3: 修复 Unix 打包脚本

**Files:**
- Modify: `build.sh`

- [ ] **Step 1: 写最小改动，去掉 `git describe` 依赖**

将 `build.sh` 顶部的版本获取逻辑改成：

```bash
#!/bin/bash

set -euo pipefail

VERSION="${1:-dev}"
APP_NAME="gen-handler"
BUILD_DIR="dist"
```

保留其余多平台构建逻辑不变，只确保所有 `go build` 继续使用：

```bash
-ldflags "-X main.version=$VERSION"
```

- [ ] **Step 2: 如当前环境可用 bash，则验证 Unix 脚本**

Run: `bash ./build.sh dev`

Expected: PASS，并在 `dist/` 下生成各平台归档文件。

If bash is unavailable in the current environment:

- 记录“未执行，原因是本机无 bash 环境”
- 不要猜测其结果

### Task 4: 做端到端验证

**Files:**
- Test only

- [ ] **Step 1: 跑全量单测**

Run: `go test ./...`

Expected: PASS

- [ ] **Step 2: 跑构建验证**

Run: `go build .`

Expected: PASS

- [ ] **Step 3: 验证 Windows 产物里的版本输出**

Run: `.\dist\windows_amd64\gen-handler.exe -version`

Expected: 输出 `dev`

- [ ] **Step 4: 验证带版本参数的产物命名**

Run: `.\build.bat v1.2.3`

Expected: PASS，并生成 `dist\gen-handler_v1.2.3_windows_amd64.zip`

- [ ] **Step 5: 验证带版本参数的二进制输出**

Run: `.\dist\windows_amd64\gen-handler.exe -version`

Expected: 输出 `v1.2.3`
