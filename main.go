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
		protoDir  = flag.String("proto-dir", "./internal/proto/axis/devopsx", "proto 生成的 grpc 文件目录")
		outputDir = flag.String("output-dir", "./api/grpc", "handler 输出目录")
		coreDir   = flag.String("core-dir", "./core", "core service 输出目录")
		wireDir   = flag.String("wire-dir", "./cmd/devopsx", "wire 命令执行目录")
		modulePath = flag.String("module", "bsi/axis/devopsx", "Go 模块路径（用于生成 import 路径）")
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

	// 创建配置
	config := Config{
		ProtoDir:   *protoDir,
		OutputDir:  *outputDir,
		CoreDir:    *coreDir,
		WireDir:    *wireDir,
		ModulePath: *modulePath,
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
