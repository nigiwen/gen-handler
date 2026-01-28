package main

import (
	"flag"
	"fmt"
	"os"
	
	"github.com/nigiwen/gen-handler/cmd"
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

func main() {
	// 定义命令行参数
	var (
		protoDir   = flag.String("proto-dir", "", "proto 生成的 grpc 文件目录（未指定时自动从 module 生成）")
		outputDir  = flag.String("output-dir", "./api/grpc", "handler 输出目录")
		coreDir    = flag.String("core-dir", "./core", "core service 输出目录")
		wireDir    = flag.String("wire-dir", "", "wire 命令执行目录（未指定时自动从 module 生成）")
		modulePath = flag.String("module", "", "Go 模块路径（用于生成 import 路径，未指定时自动从 go.mod 读取）")
		showHelp   = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Axis 开发工具集\n\n")
		fmt.Fprintf(os.Stderr, "用法: %s <命令> [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "可用命令:\n")
		fmt.Fprintf(os.Stderr, "  handler    生成 gRPC Handler 代码\n")
		fmt.Fprintf(os.Stderr, "  data       同步 Data 层 (Entity -> Repo & dbset)\n\n")
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s handler\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s data\n", os.Args[0])
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	command := os.Args[1]
	// 重新解析剩余的参数
	os.Args = append(os.Args[:1], os.Args[2:]...)
	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 如果未指定 module 参数，尝试从 go.mod 自动读取
	finalModulePath := *modulePath
	if finalModulePath == "" {
		// 获取当前工作目录
		workDir, err := os.Getwd()
		if err != nil {
			workDir = "."
		}

		// 尝试从 go.mod 读取
		if module, found := util.ReadModuleFromGoMod(workDir); found {
			finalModulePath = module
			fmt.Printf("📦 从 go.mod 自动读取 module: %s\n", finalModulePath)
		} else {
			fmt.Printf("❌ 未找到 go.mod 文件，且未指定 -module 参数\n")
			fmt.Printf("💡 请使用 -module 参数指定 Go 模块路径，或在项目根目录运行此工具\n")
			os.Exit(1)
		}
	}

	// 创建配置
	config := types.Config{
		ModulePath: finalModulePath,
		OutputDir:  *outputDir,
		CoreDir:    *coreDir,
	}

	// 如果未指定 wire-dir 参数，从 module 自动生成
	finalWireDir := *wireDir
	if finalWireDir == "" {
		finalWireDir = util.GenerateWireDirFromModule(config.ModulePath)
		fmt.Printf("🔧 从 module 自动生成 wire-dir: %s\n", finalWireDir)
	}
	config.WireDir = finalWireDir

	// 路由到对应的命令
	switch command {
	case "handler":
		cmd.RunHandlerCommand(config, *protoDir)
	case "data":
		cmd.RunDataCommand(config)
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}
