package workflow

import (
	"reflect"
	"testing"

	"github.com/nigiwen/gen-handler/internal/types"
)

func TestHandlerWorkflowRunItemBranchesByFileCreation(t *testing.T) {
	cases := []struct {
		name           string
		grpcCreated    bool
		coreCreated    bool
		wantSteps      []string
		wantUpdateGrpc bool
		wantUpdateCore bool
		wantWire       bool
	}{
		{
			name:           "create both files",
			grpcCreated:    true,
			coreCreated:    true,
			wantSteps:      []string{"处理 handler 文件", "更新 grpc.go", "处理 core service", "更新 core ProviderSet", "运行 wire"},
			wantUpdateGrpc: true,
			wantUpdateCore: true,
			wantWire:       true,
		},
		{
			name:           "create core only",
			grpcCreated:    false,
			coreCreated:    true,
			wantSteps:      []string{"处理 handler 文件", "处理 core service", "更新 core ProviderSet", "运行 wire"},
			wantUpdateGrpc: false,
			wantUpdateCore: true,
			wantWire:       true,
		},
		{
			name:           "create grpc only",
			grpcCreated:    true,
			coreCreated:    false,
			wantSteps:      []string{"处理 handler 文件", "更新 grpc.go", "处理 core service", "运行 wire"},
			wantUpdateGrpc: true,
			wantUpdateCore: false,
			wantWire:       true,
		},
		{
			name:           "complete existing files only",
			grpcCreated:    false,
			coreCreated:    false,
			wantSteps:      []string{"处理 handler 文件", "处理 core service"},
			wantUpdateGrpc: false,
			wantUpdateCore: false,
			wantWire:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotSteps []string
			var updatedGrpc bool
			var updatedCore bool
			var ranWire bool

			wf := HandlerWorkflow{
				Config: types.Config{
					OutputDir: "./api/grpc",
					CoreDir:   "./core",
					WireDir:   "./cmd/devopsx",
				},
				EnsureHandlerFile: func(service types.ServiceInfo, config types.Config) (bool, error) {
					return tc.grpcCreated, nil
				},
				UpdateGrpcProvider: func(service types.ServiceInfo, outputDir string) error {
					updatedGrpc = true
					return nil
				},
				EnsureCoreServiceFile: func(service types.ServiceInfo, config types.Config) (bool, error) {
					return tc.coreCreated, nil
				},
				UpdateCoreProvider: func(service types.ServiceInfo, coreDir string) error {
					updatedCore = true
					return nil
				},
				RunWire: func(wireDir string) error {
					ranWire = true
					return nil
				},
			}

			item := Item{
				ID:    "project.go",
				Title: "project.go",
				Payload: types.ServiceInfo{
					FileName:    "project.go",
					HandlerName: "ProjectHandler",
					ServiceName: "ProjectService",
				},
			}

			result := wf.RunItem(item, func(ev ProgressEvent) {
				gotSteps = append(gotSteps, ev.Step)
			})

			if !result.Success {
				t.Fatalf("expected success result, got %+v", result)
			}
			if updatedGrpc != tc.wantUpdateGrpc {
				t.Fatalf("unexpected UpdateGrpcProvider call: got %v want %v", updatedGrpc, tc.wantUpdateGrpc)
			}
			if updatedCore != tc.wantUpdateCore {
				t.Fatalf("unexpected UpdateCoreProvider call: got %v want %v", updatedCore, tc.wantUpdateCore)
			}
			if ranWire != tc.wantWire {
				t.Fatalf("unexpected RunWire call: got %v want %v", ranWire, tc.wantWire)
			}
			if !reflect.DeepEqual(gotSteps, tc.wantSteps) {
				t.Fatalf("unexpected steps: got %v want %v", gotSteps, tc.wantSteps)
			}
		})
	}
}
