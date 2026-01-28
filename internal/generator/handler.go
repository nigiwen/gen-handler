package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/updater"
	"github.com/nigiwen/gen-handler/internal/util"
)

// GenerateHandlerFile 生成 handler 文件
func GenerateHandlerFile(service types.ServiceInfo, config types.Config, force bool) error {
	// 如果不强制生成，检查文件是否已存在
	if !force {
		filePath := filepath.Join(config.OutputDir, service.FileName)
		if util.FileExists(filePath) {
			return fmt.Errorf("文件已存在")
		}
	}

	// 创建输出目录
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成代码
	code, err := generateHandlerCode(service, config)
	if err != nil {
		return fmt.Errorf("生成代码失败: %v", err)
	}

	// 写入文件
	filePath := filepath.Join(config.OutputDir, service.FileName)
	if err := util.WriteFile(filePath, code); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// 更新 grpc.go 文件
	grpcFilePath := filepath.Join(config.OutputDir, "grpc.go")
	if err := updater.UpdateGrpcFile(service, grpcFilePath); err != nil {
		// 更新失败不影响主流程，只打印警告
		fmt.Printf("⚠️  更新 grpc.go 失败: %v\n", err)
	}

	// 生成 core Service 文件
	if err := GenerateCoreServiceFile(service, config, force); err != nil {
		// 如果文件已存在，不报错
		if !strings.Contains(err.Error(), "已存在") {
			fmt.Printf("⚠️  生成 core Service 失败: %v\n", err)
		}
	} else {
		// 更新 core/core.go 的 ProviderSet
		coreGoPath := filepath.Join(config.CoreDir, "core.go")
		if err := updater.UpdateProviderSet(coreGoPath, "New"+service.ServiceName, "Service"); err != nil {
			fmt.Printf("⚠️  更新 core/core.go 失败: %v\n", err)
		}
	}

	// 运行 wire 命令重新生成依赖注入代码
	if err := RunWireCommand(config.WireDir); err != nil {
		fmt.Printf("⚠️  运行 wire 命令失败: %v\n", err)
		fmt.Printf("💡 请手动在 %s 目录下运行 wire 命令\n", config.WireDir)
	} else {
		fmt.Printf("✅ 已重新生成 wire 依赖注入代码\n")
	}

	return nil
}

// generateHandlerCode 生成 handler 代码
func generateHandlerCode(service types.ServiceInfo, config types.Config) (string, error) {
	// 创建模板数据，包含 service 和 config
	type templateData struct {
		types.ServiceInfo
		types.Config
	}
	
	data := templateData{
		ServiceInfo: service,
		Config:      config,
	}

	return ExecuteTemplate(HandlerTemplate, data)
}

// RunWireCommand 在指定目录下运行 wire 命令
func RunWireCommand(wireDir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(wireDir); os.IsNotExist(err) {
		return fmt.Errorf("目录不存在: %s", wireDir)
	}

	// 执行 wire 命令
	cmd := exec.Command("wire", ".")
	cmd.Dir = wireDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wire 命令执行失败: %v", err)
	}

	return nil
}
