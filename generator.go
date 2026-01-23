package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// generateHandlerFile 生成 handler 文件
func generateHandlerFile(service ServiceInfo, config Config, force bool) error {
	// 如果不强制生成，检查文件是否已存在
	if !force {
		if handlerExists(service, config.OutputDir) {
			return fmt.Errorf("文件已存在")
		}
	}

	// 创建输出目录
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成代码
	code, err := generateCode(service, config)
	if err != nil {
		return fmt.Errorf("生成代码失败: %v", err)
	}

	// 写入文件
	filePath := filepath.Join(config.OutputDir, service.FileName)
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// 更新 grpc.go 文件
	grpcFilePath := filepath.Join(config.OutputDir, "grpc.go")
	if err := updateGrpcFile(service, grpcFilePath); err != nil {
		// 更新失败不影响主流程，只打印警告
		fmt.Printf("⚠️  更新 grpc.go 失败: %v\n", err)
	}

	// 生成 core Service 文件
	if err := generateCoreServiceFile(service, config, force); err != nil {
		// 如果文件已存在，不报错
		if !strings.Contains(err.Error(), "已存在") {
			fmt.Printf("⚠️  生成 core Service 失败: %v\n", err)
		}
	} else {
		// 更新 core/core.go 的 ProviderSet
		coreGoPath := filepath.Join(config.CoreDir, "core.go")
		if err := updateCoreProviderSet(service, coreGoPath); err != nil {
			fmt.Printf("⚠️  更新 core/core.go 失败: %v\n", err)
		}
	}

	// 运行 wire 命令重新生成依赖注入代码
	if err := runWireCommand(config.WireDir); err != nil {
		fmt.Printf("⚠️  运行 wire 命令失败: %v\n", err)
		fmt.Printf("💡 请手动在 %s 目录下运行 wire 命令\n", config.WireDir)
	} else {
		fmt.Printf("✅ 已重新生成 wire 依赖注入代码\n")
	}

	return nil
}

// runWireCommand 在指定目录下运行 wire 命令
func runWireCommand(wireDir string) error {
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

// generateCode 生成 handler 代码
func generateCode(service ServiceInfo, config Config) (string, error) {
	tmpl := `package grpc

import (
	"context"

	"{{.ModulePath}}/core"
	"{{.ModulePath}}/internal/proto/axis/devopsx"
	"{{.ModulePath}}/internal/proto/basic"
)

type {{.HandlerName}} struct {
	devopsx.Unimplemented{{.ServerName}}
	{{.FieldName}} *core.{{.ServiceName}}
}

func New{{.HandlerName}}({{.FieldName}} *core.{{.ServiceName}}) *{{.HandlerName}} {
	return &{{.HandlerName}}{
		{{.FieldName}}: {{.FieldName}},
	}
}
{{range .Methods}}
{{if .Comment}}// {{.Comment}}{{else}}// {{.Name}}{{end}}
// @name devopsx.{{$.FieldName | trimSrv}}.{{.Name}}
// @desc {{if .Comment}}{{.Comment}}{{else}}{{.Name}}{{end}}
func ({{$.FieldName | firstChar}} *{{$.HandlerName}}) {{.Name}}(ctx context.Context, in *{{.RequestPkg}}.{{.RequestType}}) (*{{.ResponsePkg}}.{{.ResponseType}}, error) {
	return {{$.FieldName | firstChar}}.{{$.FieldName}}.{{.Name}}(ctx, in)
}
{{end}}
`
	
	// 创建模板数据，包含 service 和 config
	type templateData struct {
		ServiceInfo
		Config
	}
	
	data := templateData{
		ServiceInfo: service,
		Config:      config,
	}

	// 创建模板函数
	funcMap := template.FuncMap{
		"trimSrv": func(s string) string {
			return strings.TrimSuffix(s, "Srv")
		},
		"firstChar": func(s string) string {
			if len(s) == 0 {
				return "x"
			}
			return strings.ToLower(s[:1])
		},
	}

	t, err := template.New("handler").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
