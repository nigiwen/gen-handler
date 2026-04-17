package cmd

import (
	"fmt"
	"strings"

	"github.com/nigiwen/gen-handler/internal/generator"
	"github.com/nigiwen/gen-handler/internal/tui"
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
	"github.com/nigiwen/gen-handler/internal/workflow"
)

// RunHandlerCommand 执行 handler 命令
func RunHandlerCommand(config types.Config, protoDir string) error {
	// 如果未指定 proto-dir 参数，从 module 自动生成
	finalProtoDir := protoDir
	if finalProtoDir == "" {
		finalProtoDir = generateProtoDirFromModule(config.ModulePath)
		fmt.Printf("📁 从 module 自动生成 proto-dir: %s\n", finalProtoDir)
	}
	config.ProtoDir = finalProtoDir

	// 查找所有 *_grpc.pb.go 文件
	grpcFiles, err := globGrpcFiles(config.ProtoDir + "/*_grpc.pb.go")
	if err != nil {
		return fmt.Errorf("查找 grpc 文件失败: %w", err)
	}

	if len(grpcFiles) == 0 {
		return fmt.Errorf("未找到 grpc 文件在 %s", config.ProtoDir)
	}

	fmt.Printf("✅ 找到 %d 个 grpc 文件\n", len(grpcFiles))

	var services []types.ServiceInfo
	parseFailures := make([]string, 0)

	// 解析每个 grpc 文件
	for _, file := range grpcFiles {
		fileServices, err := parseGrpcFile(file)
		if err != nil {
			fmt.Printf("⚠️  解析文件 %s 失败: %v\n", file, err)
			parseFailures = append(parseFailures, fmt.Sprintf("%s: %v", file, err))
			continue
		}
		services = append(services, fileServices...)
	}

	fmt.Printf("✅ 找到 %d 个服务接口\n", len(services))

	if len(services) == 0 {
		if len(parseFailures) > 0 {
			return fmt.Errorf("解析 grpc 文件失败: %s", strings.Join(parseFailures, "; "))
		}
		fmt.Println("⚠️  未解析到任何服务接口，无需生成")
		return nil
	}

	wf := workflow.HandlerWorkflow{
		Config:                config,
		EnsureHandlerFile:     generator.EnsureHandlerFile,
		UpdateGrpcProvider:    generator.UpdateGrpcProvider,
		EnsureCoreServiceFile: generator.EnsureCoreServiceFile,
		UpdateCoreProvider:    generator.UpdateCoreProvider,
		RunWire:               generator.RunWireCommand,
	}

	if err := runSession(tui.SessionConfig{
		Title: "Handler Generate",
		Items: wf.BuildItems(services),
		Run:   wf.RunItem,
	}); err != nil {
		if len(parseFailures) > 0 {
			return fmt.Errorf("Handler Generate 失败: %w；另有 grpc 文件解析失败: %s", err, strings.Join(parseFailures, "; "))
		}
		return fmt.Errorf("Handler Generate 失败: %w", err)
	}
	if len(parseFailures) > 0 {
		return fmt.Errorf("部分 grpc 文件解析失败: %s", strings.Join(parseFailures, "; "))
	}
	return nil
}

// generateProtoDirFromModule 从 module 路径生成 proto-dir
func generateProtoDirFromModule(modulePath string) string {
	return util.GenerateProtoDirFromModule(modulePath)
}
