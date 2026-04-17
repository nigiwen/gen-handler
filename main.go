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
	if err := flagSet.Parse(args[2:]); err != nil {
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
