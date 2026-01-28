package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/nigiwen/gen-handler/internal/generator"
	"github.com/nigiwen/gen-handler/internal/parser"
	"github.com/nigiwen/gen-handler/internal/scanner"
	"github.com/nigiwen/gen-handler/internal/selector"
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

// RunHandlerCommand 执行 handler 命令
func RunHandlerCommand(config types.Config, protoDir string) {
	// 如果未指定 proto-dir 参数，从 module 自动生成
	finalProtoDir := protoDir
	if finalProtoDir == "" {
		finalProtoDir = generateProtoDirFromModule(config.ModulePath)
		fmt.Printf("📁 从 module 自动生成 proto-dir: %s\n", finalProtoDir)
	}
	config.ProtoDir = finalProtoDir

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

	var services []types.ServiceInfo

	// 解析每个 grpc 文件
	for _, file := range grpcFiles {
		fileServices, err := parser.ParseGrpcFile(file)
		if err != nil {
			fmt.Printf("⚠️  解析文件 %s 失败: %v\n", file, err)
			continue
		}
		services = append(services, fileServices...)
	}

	fmt.Printf("✅ 找到 %d 个服务接口\n", len(services))

	// 检查哪些服务还没有生成 handler 文件
	missingServices := scanner.FindMissingHandlers(services, config.OutputDir)

	if len(missingServices) == 0 {
		fmt.Println("✅ 所有 handler 文件都已存在，无需生成")
		return
	}

	// 交互式选择
	selected := selector.SelectItems(missingServices)
	if len(selected) == 0 {
		fmt.Println("\n⏭️  已取消")
		return
	}

	// 生成选中的文件
	service := selected[0]
	if err := generator.GenerateHandlerFile(service, config, true); err != nil {
		fmt.Printf("\n⚠️  生成 %s 失败: %v\n", service.FileName, err)
		return
	}
	fmt.Printf("\n✅ 已生成 %s\n", service.FileName)
}

// generateProtoDirFromModule 从 module 路径生成 proto-dir
func generateProtoDirFromModule(modulePath string) string {
	return util.GenerateProtoDirFromModule(modulePath)
}
