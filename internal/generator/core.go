package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

// GenerateCoreServiceFile 生成 core 层 Service 文件
func GenerateCoreServiceFile(service types.ServiceInfo, config types.Config, force bool) error {
	// 如果不强制生成，检查文件是否已存在
	if !force {
		if coreServiceExists(service, config.CoreDir) {
			return fmt.Errorf("core service 文件已存在")
		}
	}

	// 创建输出目录
	if err := os.MkdirAll(config.CoreDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成代码
	code, err := generateCoreServiceCode(service, config)
	if err != nil {
		return fmt.Errorf("生成代码失败: %v", err)
	}

	// 写入文件（去掉 Service 后缀）
	baseName := strings.TrimSuffix(service.ServiceName, "Service")
	fileName := util.CamelToSnake(baseName) + ".go"
	filePath := filepath.Join(config.CoreDir, fileName)
	if err := util.WriteFile(filePath, code); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// coreServiceExists 检查 core service 文件是否已存在
func coreServiceExists(service types.ServiceInfo, coreDir string) bool {
	// 去掉 Service 后缀
	baseName := strings.TrimSuffix(service.ServiceName, "Service")
	fileName := util.CamelToSnake(baseName) + ".go"
	filePath := filepath.Join(coreDir, fileName)
	return util.FileExists(filePath)
}

// generateCoreServiceCode 生成 core Service 代码
func generateCoreServiceCode(service types.ServiceInfo, config types.Config) (string, error) {
	// 创建模板数据，包含 service 和 config
	type templateData struct {
		types.ServiceInfo
		types.Config
	}
	
	data := templateData{
		ServiceInfo: service,
		Config:      config,
	}

	return ExecuteTemplate(CoreServiceTemplate, data)
}
