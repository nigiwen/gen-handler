package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// generateCoreServiceFile 生成 core 层 Service 文件
func generateCoreServiceFile(service ServiceInfo, config Config, force bool) error {
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
	fileName := camelToSnake(baseName) + ".go"
	filePath := filepath.Join(config.CoreDir, fileName)
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// coreServiceExists 检查 core service 文件是否已存在
func coreServiceExists(service ServiceInfo, coreDir string) bool {
	// 去掉 Service 后缀
	baseName := strings.TrimSuffix(service.ServiceName, "Service")
	fileName := camelToSnake(baseName) + ".go"
	filePath := filepath.Join(coreDir, fileName)
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

// generateCoreServiceCode 生成 core Service 代码
func generateCoreServiceCode(service ServiceInfo, config Config) (string, error) {
	tmpl := `package core

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"{{.ModulePath}}/internal/micro/client"
	"{{.ModulePath}}/internal/proto/axis/devopsx"
	"{{.ModulePath}}/internal/proto/basic"
	mgorm "bsi/kratos/micro/gorm"
)

type {{.ServiceName}} struct {
	srvClient        *client.Client
	log              *log.Helper
	bs               *devopsx.Bootstrap
	transactionScope *mgorm.TransactionScope
}

//nolint:revive
func New{{.ServiceName}}(
	srvClient *client.Client,
	logger log.Logger,
	bs *devopsx.Bootstrap,
	transactionScope *mgorm.TransactionScope,
) *{{.ServiceName}} {
	return &{{.ServiceName}}{
		srvClient:        srvClient,
		log:              log.NewHelper(log.With(logger, "module", "{{.FieldName | trimSrv}}")),
		bs:               bs,
		transactionScope: transactionScope,
	}
}
{{range .Methods}}
{{if .Comment}}// {{.Comment}}{{else}}// {{.Name}}{{end}}
func ({{$.FieldName | firstChar}} *{{$.ServiceName}}) {{.Name}}(ctx context.Context, in *{{.RequestPkg}}.{{.RequestType}}) (*{{.ResponsePkg}}.{{.ResponseType}}, error) {
	{{$.FieldName | firstChar}}.log.Debug("not implement")
	{{if eq .ResponseType "Empty"}}return &{{.ResponsePkg}}.Empty{}, nil{{else}}return nil, nil{{end}}
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
				return "s"
			}
			return strings.ToLower(s[:1])
		},
	}

	t, err := template.New("coreService").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
