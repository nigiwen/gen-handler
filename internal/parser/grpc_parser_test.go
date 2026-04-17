package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGrpcFileUsesFilePackageAsDefaultProtoPackage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "order_grpc.pb.go")
	content := `package orders

import "context"

type OrderServer interface {
	Create(context.Context, *CreateOrderRequest) (*CreateOrderReply, error)
	Ping(context.Context, *basic.Empty) (*basic.Empty, error)
	mustEmbedUnimplementedOrderServer()
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write grpc file: %v", err)
	}

	services, err := ParseGrpcFile(path)
	if err != nil {
		t.Fatalf("ParseGrpcFile error: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	service := services[0]
	if service.ProtoPackage != "orders" {
		t.Fatalf("expected proto package orders, got %q", service.ProtoPackage)
	}
	if got := service.Methods[0].RequestPkg; got != "orders" {
		t.Fatalf("expected request pkg orders, got %q", got)
	}
	if got := service.Methods[0].ResponsePkg; got != "orders" {
		t.Fatalf("expected response pkg orders, got %q", got)
	}
	if got := service.Methods[1].RequestPkg; got != "basic" {
		t.Fatalf("expected basic request pkg, got %q", got)
	}
}
