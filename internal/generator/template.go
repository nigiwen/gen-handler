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

// HandlerTemplate handler 模板
const HandlerTemplate = `package grpc

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

// CoreServiceTemplate core service 模板
const CoreServiceTemplate = `package core

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

// DbsetTemplate dbset 模板
const DbsetTemplate = `package dbset

import (
	jujuerrors "github.com/juju/errors"
	"gorm.io/gorm"

	"{{.ModulePath}}/internal/model/entity"
	"bsi/kratos/micro/algo/snow"
)

type {{.EntityName}} struct {
	entity.{{.EntityName}}
}

func (t *{{.EntityName}}) BeforeCreate(tx *gorm.DB) error {
	if t.ID > 0 {
		return nil
	}

	id, err := snow.ID()
	if err != nil {
		return jujuerrors.Trace(err)
	}
	t.ID = id

	return nil
}
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
