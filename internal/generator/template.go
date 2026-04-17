package generator

import (
	"bytes"
	"text/template"

	"github.com/nigiwen/gen-handler/internal/util"
)

// GetTemplateFuncs 返回所有模板函数
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"trimSrv":   util.TrimSrv,
		"firstChar": util.FirstChar,
	}
}

// ExecuteTemplate 执行模板
func ExecuteTemplate(tmplStr string, data interface{}) (string, error) {
	t, err := template.New("tmpl").Funcs(GetTemplateFuncs()).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const handlerMethodsTemplate = `{{range .Methods}}
{{if .Comment}}// {{.Comment}}{{else}}// {{.Name}}{{end}}
// @name {{$.ProtoPackage}}.{{$.FieldName | trimSrv}}.{{.Name}}
// @desc {{if .Comment}}{{.Comment}}{{else}}{{.Name}}{{end}}
func ({{$.FieldName | firstChar}} *{{$.HandlerName}}) {{.Name}}(ctx context.Context, in *{{.RequestPkg}}.{{.RequestType}}) (*{{.ResponsePkg}}.{{.ResponseType}}, error) {
	return {{$.FieldName | firstChar}}.{{$.FieldName}}.{{.Name}}(ctx, in)
}
{{end}}
`

// HandlerTemplate handler 模板
const HandlerTemplate = `package grpc

import (
{{range .Imports}}    "{{.}}"
{{end}}
)

type {{.HandlerName}} struct {
	{{.ProtoPackage}}.Unimplemented{{.ServerName}}
	{{.FieldName}} *core.{{.ServiceName}}
}

func New{{.HandlerName}}({{.FieldName}} *core.{{.ServiceName}}) *{{.HandlerName}} {
    return &{{.HandlerName}}{
        {{.FieldName}}: {{.FieldName}},
    }
}
` + handlerMethodsTemplate

// HandlerMethodsTemplate handler 方法模板
const HandlerMethodsTemplate = handlerMethodsTemplate

const coreServiceMethodsTemplate = `{{range .Methods}}
{{if .Comment}}// {{.Comment}}{{else}}// {{.Name}}{{end}}
func ({{$.FieldName | firstChar}} *{{$.ServiceName}}) {{.Name}}(ctx context.Context, in *{{.RequestPkg}}.{{.RequestType}}) (*{{.ResponsePkg}}.{{.ResponseType}}, error) {
    {{$.FieldName | firstChar}}.log.Debug("not implement")
    {{if eq .ResponseType "Empty"}}return &{{.ResponsePkg}}.Empty{}, nil{{else}}return nil, nil{{end}}
}
{{end}}
`

// CoreServiceTemplate core service 模板
const CoreServiceTemplate = `package core

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"{{.ModulePath}}/internal/micro/client"
	"{{.ProtoImportPath}}"
{{range .ProtoImports}}    "{{.}}"
{{end}}
	mgorm "bsi/kratos/micro/gorm"
)

type {{.ServiceName}} struct {
	srvClient        *client.Client
	log              *log.Helper
	bs               *{{.ProtoPackage}}.Bootstrap
	transactionScope *mgorm.TransactionScope
}

//nolint:revive
func New{{.ServiceName}}(
	srvClient *client.Client,
	logger log.Logger,
	bs *{{.ProtoPackage}}.Bootstrap,
	transactionScope *mgorm.TransactionScope,
) *{{.ServiceName}} {
	return &{{.ServiceName}}{
		srvClient:        srvClient,
		log:              log.NewHelper(log.With(logger, "module", "{{.FieldName | trimSrv}}")),
		bs:               bs,
        transactionScope: transactionScope,
    }
}
` + coreServiceMethodsTemplate

// CoreServiceMethodsTemplate core service 方法模板
const CoreServiceMethodsTemplate = coreServiceMethodsTemplate

// EntityTemplate entity 占位模板
const EntityTemplate = `package entity
`

// RepoTemplate repo 模板
const RepoTemplate = `package data

import (
	"github.com/go-kratos/kratos/v2/log"

	mgorm "bsi/kratos/micro/gorm"
)

type {{.EntityName}}Repo struct {
	log *log.Helper
	*mgorm.GormDB
}

func New{{.EntityName}}Repo(logger log.Logger, db *mgorm.GormDB) *{{.EntityName}}Repo {
	return &{{.EntityName}}Repo{
		log:    log.NewHelper(logger),
		GormDB: db,
	}
}
`
