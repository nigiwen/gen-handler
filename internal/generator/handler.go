package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nigiwen/gen-handler/internal/types"
	"github.com/nigiwen/gen-handler/internal/updater"
	"github.com/nigiwen/gen-handler/internal/util"
)

// EnsureHandlerFile 确保 handler 文件存在并补全缺失方法。
func EnsureHandlerFile(service types.ServiceInfo, config types.Config) (bool, error) {
	filePath := filepath.Join(config.OutputDir, service.FileName)
	if !util.FileExists(filePath) {
		if err := WriteHandlerFile(service, config, false); err != nil {
			return false, err
		}
		return true, nil
	}

	goFile, err := parseGoFile(filePath)
	if err != nil {
		return false, fmt.Errorf("解析 handler 文件失败: %w", err)
	}
	if !goFile.HasType(service.HandlerName) {
		return false, fmt.Errorf("未找到 handler 类型: %s", service.HandlerName)
	}

	protoPackage := serviceProtoPackage(service, config)
	missingMethods, mismatchedMethods := missingHandlerMethods(
		service.Methods,
		goFile.MethodSignaturesForReceiver(service.HandlerName),
		config,
		protoPackage,
	)
	if len(mismatchedMethods) > 0 {
		return false, newMethodSignatureMismatchError("handler 文件", mismatchedMethods)
	}
	if len(missingMethods) == 0 {
		return false, nil
	}

	goFile.EnsureImports(requiredHandlerImports(missingMethods, config, protoPackage))

	methodCode, err := generateHandlerMethodsCode(service, missingMethods, protoPackage)
	if err != nil {
		return false, err
	}
	updated, err := goFile.AppendCode(methodCode)
	if err != nil {
		return false, fmt.Errorf("格式化 handler 文件失败: %w", err)
	}

	if err := util.WriteFile(filePath, updated); err != nil {
		return false, fmt.Errorf("写入文件失败: %v", err)
	}

	return false, nil
}

// WriteHandlerFile 只生成 handler 文件本身。
func WriteHandlerFile(service types.ServiceInfo, config types.Config, force bool) error {
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

	return nil
}

// UpdateGrpcProvider 更新 grpc.go ProviderSet。
func UpdateGrpcProvider(service types.ServiceInfo, outputDir string) error {
	grpcFilePath := filepath.Join(outputDir, "grpc.go")
	return updater.UpdateGrpcFile(service, grpcFilePath)
}

// WriteCoreService 只生成 core service 文件。
func WriteCoreService(service types.ServiceInfo, config types.Config, force bool) error {
	return GenerateCoreServiceFile(service, config, force)
}

// UpdateCoreProvider 更新 core ProviderSet。
func UpdateCoreProvider(service types.ServiceInfo, coreDir string) error {
	coreGoPath := filepath.Join(coreDir, "core.go")
	return updater.UpdateProviderSet(coreGoPath, "New"+service.ServiceName, "Service")
}

// GenerateHandlerFile 生成 handler 文件。
func GenerateHandlerFile(service types.ServiceInfo, config types.Config, force bool) error {
	if err := WriteHandlerFile(service, config, force); err != nil {
		return err
	}

	if err := UpdateGrpcProvider(service, config.OutputDir); err != nil {
		// 更新失败不影响主流程，只打印警告
		fmt.Printf("⚠️  更新 grpc.go 失败: %v\n", err)
	}

	if err := WriteCoreService(service, config, force); err != nil {
		// 如果文件已存在，不报错
		if !strings.Contains(err.Error(), "已存在") {
			fmt.Printf("⚠️  生成 core Service 失败: %v\n", err)
		}
	} else {
		if err := UpdateCoreProvider(service, config.CoreDir); err != nil {
			fmt.Printf("⚠️  更新 core/core.go 失败: %v\n", err)
		}
	}

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
		Imports      []string
		ProtoPackage string
	}

	protoPackage := serviceProtoPackage(service, config)
	data := templateData{
		ServiceInfo:  service,
		Config:       config,
		Imports:      createHandlerImports(service.Methods, config, protoPackage),
		ProtoPackage: protoPackage,
	}

	return ExecuteTemplate(HandlerTemplate, data)
}

func generateHandlerMethodsCode(service types.ServiceInfo, methods []types.Method, protoPackage string) (string, error) {
	service.Methods = methods
	service.ProtoPackage = protoPackage
	return ExecuteTemplate(HandlerMethodsTemplate, service)
}

func missingHandlerMethods(methods []types.Method, existing map[string]methodSignature, config types.Config, protoPackage string) ([]types.Method, []methodSignatureMismatch) {
	return classifyExpectedMethods(methods, existing, config, protoPackage)
}

func requiredHandlerImports(methods []types.Method, config types.Config, protoPackage string) []string {
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

func createHandlerImports(methods []types.Method, config types.Config, protoPackage string) []string {
	mainProtoImport := defaultProtoImportPath(config)
	imports := map[string]struct{}{
		"context":                   {},
		config.ModulePath + "/core": {},
		mainProtoImport:             {},
	}
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

func protoImportPath(config types.Config, protoPackage, pkg string) string {
	switch pkg {
	case "", protoPackage:
		return defaultProtoImportPath(config)
	case "basic":
		return config.ModulePath + "/internal/proto/basic"
	case "zebra":
		return config.ModulePath + "/internal/proto/zebra"
	default:
		return ""
	}
}

func defaultProtoImportPath(config types.Config) string {
	protoDir := config.ProtoDir
	if protoDir == "" {
		protoDir = util.GenerateProtoDirFromModule(config.ModulePath)
	}

	cleaned := filepath.ToSlash(filepath.Clean(protoDir))
	cleaned = strings.TrimPrefix(cleaned, "./")
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" || cleaned == "." {
		return config.ModulePath
	}
	return config.ModulePath + "/" + cleaned
}

func serviceProtoPackage(service types.ServiceInfo, config types.Config) string {
	if service.ProtoPackage != "" {
		return service.ProtoPackage
	}
	if pkg := protoPackageFromMethods(service.Methods); pkg != "" {
		return pkg
	}
	importPath := defaultProtoImportPath(config)
	if slash := strings.LastIndex(importPath, "/"); slash >= 0 {
		return importPath[slash+1:]
	}
	return importPath
}

func protoPackageFromMethods(methods []types.Method) string {
	for _, method := range methods {
		for _, pkg := range []string{method.RequestPkg, method.ResponsePkg} {
			if pkg != "" && pkg != "basic" && pkg != "zebra" {
				return pkg
			}
		}
	}
	return ""
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
