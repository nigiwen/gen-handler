package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// 定义命令行参数
	var (
		protoDir  = flag.String("proto-dir", "", "proto 生成的 grpc 文件目录（未指定时自动从 module 生成）")
		outputDir = flag.String("output-dir", "./api/grpc", "handler 输出目录")
		coreDir   = flag.String("core-dir", "./core", "core service 输出目录")
		wireDir   = flag.String("wire-dir", "", "wire 命令执行目录（未指定时自动从 module 生成）")
		modulePath = flag.String("module", "", "Go 模块路径（用于生成 import 路径，未指定时自动从 go.mod 读取）")
		showHelp  = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gRPC Handler 生成工具\n\n")
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -proto-dir ./proto -output-dir ./handlers\n", os.Args[0])
	}

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
		if module, found := readModuleFromGoMod(workDir); found {
			finalModulePath = module
			fmt.Printf("📦 从 go.mod 自动读取 module: %s\n", finalModulePath)
		} else {
			fmt.Printf("❌ 未找到 go.mod 文件，且未指定 -module 参数\n")
			fmt.Printf("💡 请使用 -module 参数指定 Go 模块路径，或在项目根目录运行此工具\n")
			os.Exit(1)
		}
	}

	// 如果未指定 proto-dir 参数，从 module 自动生成
	finalProtoDir := *protoDir
	if finalProtoDir == "" {
		finalProtoDir = generateProtoDirFromModule(finalModulePath)
		fmt.Printf("📁 从 module 自动生成 proto-dir: %s\n", finalProtoDir)
	}

	// 如果未指定 wire-dir 参数，从 module 自动生成
	finalWireDir := *wireDir
	if finalWireDir == "" {
		finalWireDir = generateWireDirFromModule(finalModulePath)
		fmt.Printf("🔧 从 module 自动生成 wire-dir: %s\n", finalWireDir)
	}

	// 创建配置
	config := Config{
		ProtoDir:   finalProtoDir,
		OutputDir:  *outputDir,
		CoreDir:    *coreDir,
		WireDir:    finalWireDir,
		ModulePath: finalModulePath,
	}

	// 查找所有 *_grpc.pb.go 文件
	grpcFiles, err := filepath.Glob(filepath.Join(config.ProtoDir, "*_grpc.pb.go"))
	if err != nil {
		fmt.Printf("❌ 查找 grpc 文件失败: %v\n", err)
		os.Exit(1)
	}

	if len(grpcFiles) == 0 {
		fmt.Printf("⚠️  未找到 grpc 文件在 %s\n", config.ProtoDir)
		os.Exit(1)
	}

	fmt.Printf("✅ 找到 %d 个 grpc 文件\n", len(grpcFiles))

	var services []ServiceInfo

	// 解析每个 grpc 文件
	for _, file := range grpcFiles {
		fileServices, err := parseGrpcFile(file)
		if err != nil {
			fmt.Printf("⚠️  解析文件 %s 失败: %v\n", file, err)
			continue
		}
		services = append(services, fileServices...)
	}

	fmt.Printf("✅ 找到 %d 个服务接口\n", len(services))

	// 检查哪些服务还没有生成 handler 文件
	missingServices := findMissingHandlers(services, config.OutputDir)

	if len(missingServices) == 0 {
		fmt.Println("✅ 所有 handler 文件都已存在，无需生成")
		return
	}

	// 交互式选择
	selected := selectServices(missingServices)
	if len(selected) == 0 {
		fmt.Println("\n⏭️  已取消")
		return
	}

	// 生成选中的文件
	service := selected[0]
	if err := generateHandlerFile(service, config, true); err != nil {
		fmt.Printf("\n⚠️  生成 %s 失败: %v\n", service.FileName, err)
		return
	}
	fmt.Printf("\n✅ 已生成 %s\n", service.FileName)
}
