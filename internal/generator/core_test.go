package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestEnsureCoreServiceFileCreatesFullFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated core file: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "type ProjectService struct") {
		t.Fatalf("expected service struct in generated file, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectService) List(") {
		t.Fatalf("expected List method in generated file, got %q", got)
	}
	if !strings.Contains(got, "\"github.com/go-kratos/kratos/v2/log\"") {
		t.Fatalf("expected log import in generated file, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/micro/client\"") {
		t.Fatalf("expected client import in generated file, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/axis/devopsx\"") {
		t.Fatalf("expected devopsx import in generated file, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/kratos/micro/gorm\"") {
		t.Fatalf("expected mgorm import in generated file, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected basic import in generated file, got %q", got)
	}
}

func TestEnsureCoreServiceFileCreatePathSkipsUnusedBasicImport(t *testing.T) {
	dir := t.TempDir()
	service := testProjectServiceDevopsOnly()

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated core file: %v", err)
	}

	got := string(content)
	if strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected no basic import in generated file, got %q", got)
	}
	if strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/zebra\"") {
		t.Fatalf("expected no zebra import in generated file, got %q", got)
	}
}

func TestEnsureCoreServiceFileUsesConfiguredProtoImportForDefaultPackage(t *testing.T) {
	dir := t.TempDir()
	service := testOrderService()

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ProtoDir:   "./internal/proto/acme/orders",
		ModulePath: "example.com/acme/ordersapp",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}

	content, err := os.ReadFile(filepath.Join(dir, service.FileName))
	if err != nil {
		t.Fatalf("read generated core file: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "\"example.com/acme/ordersapp/internal/proto/acme/orders\"") {
		t.Fatalf("expected configured proto import, got %q", got)
	}
	if !strings.Contains(got, "bs               *orders.Bootstrap") {
		t.Fatalf("expected orders proto package in bootstrap field, got %q", got)
	}
	if strings.Contains(got, "devopsx") {
		t.Fatalf("expected no hard-coded devopsx reference, got %q", got)
	}
}

func TestEnsureCoreServiceFileAppendsMissingMethodsWithoutOverwritingExistingBody(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	content := `package core

import (
    "context"

    "github.com/go-kratos/kratos/v2/log"

    "bsi/axis/devopsx/internal/micro/client"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
    mgorm "bsi/kratos/micro/gorm"
)

type ProjectService struct {
    srvClient        *client.Client
    log              *log.Helper
    transactionScope *mgorm.TransactionScope
}

func (p *ProjectService) List(ctx context.Context, in *devopsx.ListProjectRequest) (*devopsx.ListProjectReply, error) {
    return &devopsx.ListProjectReply{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing core file: %v", err)
	}

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated core file: %v", err)
	}

	got := string(updated)
	if strings.Count(got, "func (p *ProjectService) List(") != 1 {
		t.Fatalf("expected existing List method to stay single, got %q", got)
	}
	if !strings.Contains(got, "return &devopsx.ListProjectReply{}, nil") {
		t.Fatalf("expected existing List body to be preserved, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectService) Ping(") {
		t.Fatalf("expected missing Ping method to be appended, got %q", got)
	}
	if !strings.Contains(got, "\"bsi/axis/devopsx/internal/proto/basic\"") {
		t.Fatalf("expected missing basic import to be added, got %q", got)
	}
}

func TestEnsureCoreServiceFileAppendsContextImportWhenMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	content := `package core

import (
    "github.com/go-kratos/kratos/v2/log"

    "bsi/axis/devopsx/internal/micro/client"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
    mgorm "bsi/kratos/micro/gorm"
)

type ProjectService struct {
    srvClient        *client.Client
    log              *log.Helper
    transactionScope *mgorm.TransactionScope
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing core file: %v", err)
	}

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated core file: %v", err)
	}

	got := string(updated)
	if !strings.Contains(got, "\"context\"") {
		t.Fatalf("expected context import to be added, got %q", got)
	}
	if !strings.Contains(got, "func (p *ProjectService) Ping(") {
		t.Fatalf("expected missing method to be appended, got %q", got)
	}
}

func TestEnsureCoreServiceFileNoOpWhenAllMethodsExist(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	content := `package core

import (
    "context"

    "github.com/go-kratos/kratos/v2/log"

    "bsi/axis/devopsx/internal/micro/client"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
    "bsi/axis/devopsx/internal/proto/basic"
    mgorm "bsi/kratos/micro/gorm"
)

type ProjectService struct {
    srvClient        *client.Client
    log              *log.Helper
    transactionScope *mgorm.TransactionScope
}

func (p *ProjectService) List(ctx context.Context, in *devopsx.ListProjectRequest) (*devopsx.ListProjectReply, error) {
    return &devopsx.ListProjectReply{}, nil
}

func (p *ProjectService) Ping(ctx context.Context, in *basic.Empty) (*basic.Empty, error) {
    return &basic.Empty{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write existing core file: %v", err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read original core file: %v", err)
	}

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err != nil {
		t.Fatalf("EnsureCoreServiceFile error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unchanged core file: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("expected file content to stay unchanged\nbefore:\n%s\nafter:\n%s", string(before), string(after))
	}
}

func TestEnsureCoreServiceFileReturnsErrorWhenServiceTypeMissing(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	if err := os.WriteFile(path, []byte("package core\n"), 0o644); err != nil {
		t.Fatalf("write invalid core file: %v", err)
	}

	_, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
		ModulePath: "bsi/axis/devopsx",
	})
	if err == nil || !strings.Contains(err.Error(), "ProjectService") {
		t.Fatalf("expected missing type error, got %v", err)
	}
}

func TestEnsureCoreServiceFileReturnsErrorWhenMethodSignatureDrifts(t *testing.T) {
	dir := t.TempDir()
	service := testProjectService()
	path := filepath.Join(dir, service.FileName)

	content := `package core

import (
    "context"

    "github.com/go-kratos/kratos/v2/log"

    "bsi/axis/devopsx/internal/micro/client"
    "bsi/axis/devopsx/internal/proto/axis/devopsx"
    mgorm "bsi/kratos/micro/gorm"
)

type ProjectService struct {
    srvClient        *client.Client
    log              *log.Helper
    transactionScope *mgorm.TransactionScope
}

func (p *ProjectService) List(ctx context.Context, in *devopsx.OldListProjectRequest) (*devopsx.ListProjectReply, error) {
    return &devopsx.ListProjectReply{}, nil
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write drifted core file: %v", err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read original core file: %v", err)
	}

	created, err := EnsureCoreServiceFile(service, types.Config{
		CoreDir:    dir,
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
		t.Fatalf("read unchanged core file: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("expected file content to stay unchanged\nbefore:\n%s\nafter:\n%s", string(before), string(after))
	}
}
