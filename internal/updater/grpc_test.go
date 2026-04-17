package updater

import (
	"strings"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestAddToNewGRPCServerBodyUsesServiceProtoPackage(t *testing.T) {
	content := `package grpc

func NewGRPCServer(
	orderHandler *OrderHandler,
) *Server {
	srv := NewServer()
	return srv
}
`

	updated := addToNewGRPCServerBody(content, types.ServiceInfo{
		ServerName:   "OrderServer",
		HandlerName:  "OrderHandler",
		ProtoPackage: "orders",
	})

	if !strings.Contains(updated, "orders.RegisterOrderServer(srv, orderHandler)") {
		t.Fatalf("expected orders Register call, got %q", updated)
	}
}
