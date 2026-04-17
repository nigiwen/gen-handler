package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/util"
)

// EnsureCoreServiceFile 确保 core service 文件存在并补全缺失方法。
func EnsureCoreServiceFile(service types.ServiceInfo, config types.Config) (bool, error) {
	filePath := filepath.Join(config.CoreDir, service.FileName)
	if !util.FileExists(filePath) {
		if err := GenerateCoreServiceFile(service, config, false); err != nil {
			return false, err
		}
		return true, nil
	}

	goFile, err := parseGoFile(filePath)
	if err != nil {
		return false, fmt.Errorf("解析 core 文件失败: %w", err)
	}
	if !goFile.HasType(service.ServiceName) {
		return false, fmt.Errorf("未找到 core service 类型: %s", service.ServiceName)
	}

	protoPackage := serviceProtoPackage(service, config)
	missingMethods, mismatchedMethods := missingCoreMethods(
		service.Methods,
		goFile.MethodSignaturesForReceiver(service.ServiceName),
		config,
		protoPackage,
	)
	if len(mismatchedMethods) > 0 {
		return false, newMethodSignatureMismatchError("core 文件", mismatchedMethods)
	}
	if len(missingMethods) == 0 {
		return false, nil
	}

	goFile.EnsureImports(requiredCoreImports(missingMethods, config, protoPackage))

	methodCode, err := generateCoreMethodsCode(service, missingMethods)
	if err != nil {
		return false, err
	}
	updated, err := goFile.AppendCode(methodCode)
	if err != nil {
		return false, fmt.Errorf("格式化 core 文件失败: %w", err)
	}

	if err := util.WriteFile(filePath, updated); err != nil {
		return false, fmt.Errorf("写入文件失败: %v", err)
	}

	return false, nil
}

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

	// 写入文件
	filePath := filepath.Join(config.CoreDir, service.FileName)
	if err := util.WriteFile(filePath, code); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// coreServiceExists 检查 core service 文件是否已存在
func coreServiceExists(service types.ServiceInfo, coreDir string) bool {
	filePath := filepath.Join(coreDir, service.FileName)
	return util.FileExists(filePath)
}

// generateCoreServiceCode 生成 core Service 代码
func generateCoreServiceCode(service types.ServiceInfo, config types.Config) (string, error) {
	// 创建模板数据，包含 service 和 config
	type templateData struct {
		types.ServiceInfo
		types.Config
		ProtoImports    []string
		ProtoImportPath string
		ProtoPackage    string
	}

	protoPackage := serviceProtoPackage(service, config)
	data := templateData{
		ServiceInfo:     service,
		Config:          config,
		ProtoImports:    createCoreProtoImports(service.Methods, config, protoPackage),
		ProtoImportPath: defaultProtoImportPath(config),
		ProtoPackage:    protoPackage,
	}

	return ExecuteTemplate(CoreServiceTemplate, data)
}

func generateCoreMethodsCode(service types.ServiceInfo, methods []types.Method) (string, error) {
	service.Methods = methods
	return ExecuteTemplate(CoreServiceMethodsTemplate, service)
}

func missingCoreMethods(methods []types.Method, existing map[string]methodSignature, config types.Config, protoPackage string) ([]types.Method, []methodSignatureMismatch) {
	return classifyExpectedMethods(methods, existing, config, protoPackage)
}

func requiredCoreImports(methods []types.Method, config types.Config, protoPackage string) []string {
	imports := map[string]struct{}{
		"context": {},
	}
	for _, method := range methods {
		for _, pkg := range []string{method.RequestPkg, method.ResponsePkg} {
			importPath := protoImportPath(config, protoPackage, pkg)
			if importPath == "" {
				continue
			}
			imports[importPath] = struct{}{}
		}
	}

	result := make([]string, 0, len(imports))
	for importPath := range imports {
		result = append(result, importPath)
	}
	sort.Strings(result)
	return result
}

func createCoreProtoImports(methods []types.Method, config types.Config, protoPackage string) []string {
	mainProtoImport := defaultProtoImportPath(config)
	imports := make(map[string]struct{})
	for _, method := range methods {
		for _, pkg := range []string{method.RequestPkg, method.ResponsePkg} {
			importPath := protoImportPath(config, protoPackage, pkg)
			if importPath == "" || importPath == mainProtoImport {
				continue
			}
			imports[importPath] = struct{}{}
		}
	}

	result := make([]string, 0, len(imports))
	for importPath := range imports {
		result = append(result, importPath)
	}
	sort.Strings(result)
	return result
}
