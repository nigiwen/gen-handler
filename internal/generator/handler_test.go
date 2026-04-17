package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func testProjectService() types.ServiceInfo {
	return types.ServiceInfo{
		ServerName:  "ProjectServer",
		HandlerName: "ProjectHandler",
		FileName:    "project.go",
		FieldName:   "projectSrv",
		ServiceName: "ProjectService",
		Methods: []types.Method{
			{
				Name:         "List",
				RequestPkg:   "devopsx",
				RequestType:  "ListProjectRequest",
				ResponsePkg:  "devopsx",
				ResponseType: "ListProjectReply",
			},
			{
				Name:         "Ping",
				RequestPkg:   "basic",
				RequestType:  "Empty",
				ResponsePkg:  "basic",
				ResponseType: "Empty",
			},
		},
	}
}

func testProjectServiceDevopsOnly() types.ServiceInfo {
	return types.ServiceInfo{
		ServerName:  "ProjectServer",
		HandlerName: "ProjectHandler",
		FileName:    "project.go",
		FieldName:   "projectSrv",
		ServiceName: "ProjectService",
		Methods: []types.Method{
			{
				Name:         "List",
				RequestPkg:   "devopsx",
				RequestType:  "ListProjectRequest",
				ResponsePkg:  "devopsx",
				ResponseType: "ListProjectReply",
			},
		},
	}
}

func testOrderService() types.ServiceInfo {
	return types.ServiceInfo{
		ServerName:   "OrderServer",
		ProtoPackage: "orders",
		HandlerName:  "OrderHandler",
		FileName:     "order.go",
		FieldName:    "orderSrv",
		ServiceName:  "OrderService",
		Methods: []types.Method{
			{
				Name:         "Create",
				RequestPkg:   "orders",
				RequestType:  "CreateOrderRequest",
				ResponsePkg:  "orders",
				ResponseType: "CreateOrderReply",
			},
		},
	}
}

func TestEnsureHandlerFileCreatesFullFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "type ProjectHandler struct") {
		t.Fatalf("expected handler struct in generated file, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectHandler) List(") {
		t.Fatalf("expected List method in generated file, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected basic import in generated file, got %q", got)
	}
}

func TestEnsureHandlerFileCreatePathSkipsUnusedBasicImport(t *testing.T) {
	dir := t.TempDir()
	service := testProjectServiceDevopsOnly()

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	got := string(content)
	if strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected no basic import in generated file, got %q", got)
	}
}

func TestEnsureHandlerFileUsesConfiguredProtoImportForDefaultPackage(t *testing.T) {
	dir := t.TempDir()
	service := testOrderService()

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ProtoDir:   "./internal/proto/acme/orders",
		ModulePath: "example.com/acme/ordersapp",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "\"example.com/acme/ordersapp/internal/proto/acme/orders\"") {
		t.Fatalf("expected configured proto import, got %q", got)
	}
	if !strings.Contains(got, "orders.UnimplementedOrderServer") {
		t.Fatalf("expected orders proto package in handler struct, got %q", got)
	}
	if strings.Contains(got, "devopsx") {
		t.Fatalf("expected no hard-coded devopsx reference, got %q", got)
	}
}

func TestEnsureHandlerFileAppendsMissingMethodsWithoutOverwritingExistingBody(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	path := filepath.Join(dir, service.FileName)
	content := `package grpc

import (
    "context"

    "bsi/axis/devopsx/core"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
)

type ProjectHandler struct {
    devopsx.UnimplementedProjectServer
    projectSrv *core.ProjectService
}

func (p *ProjectHandler) List(ctx context.Context, in *devopsx.ListProjectRequest) (*devopsx.ListProjectReply, error) {
    return &devopsx.ListProjectReply{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing handler file: %v", err)
	}

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated handler file: %v", err)
	}

	got := string(updated)
	if strings.Count(got, "func (p *ProjectHandler) List(") != 1 {
		t.Fatalf("expected existing List method to stay single, got %q", got)
	}
	if !strings.Contains(got, "return &devopsx.ListProjectReply{}, nil") {
		t.Fatalf("expected existing List body to be preserved, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectHandler) Ping(") {
		t.Fatalf("expected missing Ping method to be appended, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected missing basic import to be added, got %q", got)
	}
}

func TestEnsureHandlerFileAppendsContextImportWhenMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	path := filepath.Join(dir, service.FileName)
	content := `package grpc

import (
    "bsi/axis/devopsx/core"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
)

type ProjectHandler struct {
    devopsx.UnimplementedProjectServer
    projectSrv *core.ProjectService
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing handler file: %v", err)
	}

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated handler file: %v", err)
	}

	got := string(updated)
	if !strings.Contains(got, "\"context\"") {
		t.Fatalf("expected context import to be added, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectHandler) Ping(") {
		t.Fatalf("expected missing method to be appended, got %q", got)
	}
}

func TestEnsureHandlerFileNoOpWhenAllMethodsExist(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	path := filepath.Join(dir, service.FileName)
	content := `package grpc

import (
    "context"

    "bsi/axis/devopsx/core"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
    "bsi/axis/devopsx/internal/proto/basic"
)

type ProjectHandler struct {
    devopsx.UnimplementedProjectServer
    projectSrv *core.ProjectService
}

func (p *ProjectHandler) List(ctx context.Context, in *devopsx.ListProjectRequest) (*devopsx.ListProjectReply, error) {
    return &devopsx.ListProjectReply{}, nil
}

func (p *ProjectHandler) Ping(ctx context.Context, in *basic.Empty) (*basic.Empty, error) {
    return &basic.Empty{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing handler file: %v", err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read original handler file: %v", err)
	}

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureHandlerFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unchanged handler file: %v", err)
	}

	if string(after) != string(before) {
		t.Fatalf("expected file content to stay unchanged\nbefore:\n%s\nafter:\n%s", string(before), string(after))
	}
}

func TestEnsureHandlerFileReturnsErrorWhenHandlerTypeMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	if err := os.WriteFile(path, []byte("package grpc\n"), 0o644); err != nil {
		t.Fatalf("write invalid handler file: %v", err)
	}

	_, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err == nil || !strings.Contains(err.Error(), "ProjectHandler") {
		t.Fatalf("expected missing type error, got %v", err)
	}
}

func TestEnsureHandlerFileReturnsErrorWhenMethodSignatureDrifts(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	content := `package grpc

import (
    "context"

    "bsi/axis/devopsx/core"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
)

type ProjectHandler struct {
    devopsx.UnimplementedProjectServer
    projectSrv *core.ProjectService
}

func (p *ProjectHandler) List(ctx context.Context, in *devopsx.OldListProjectRequest) (*devopsx.ListProjectReply, error) {
    return &devopsx.ListProjectReply{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write drifted handler file: %v", err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read original handler file: %v", err)
	}

	created, err := EnsureHandlerFile(service, types.Config{
		OutputDir:  dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err == nil {
		t.Fatalf("expected signature drift error")
	}
	if created {
		t.Fatalf("expected created=false")
	}
	if !strings.Contains(err.Error(), "List") || !strings.Contains(err.Error(), "签名") {
		t.Fatalf("expected signature drift details, got %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unchanged handler file: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("expected file content to stay unchanged\nbefore:\n%s\nafter:\n%s", string(before), string(after))
	}
}
