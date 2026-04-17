package cmd

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nigiwen/gen-handler/internal/tui"
	"github.com/nigiwen/gen-handler/internal/types"
)

func TestRunHandlerCommandUsesAllParsedServicesInSession(t *testing.T) {
	originalGlob := globGrpcFiles
	originalParseGrpcFile := parseGrpcFile
	originalRunSession := runSession
	t.Cleanup(func() {
		globGrpcFiles = originalGlob
		parseGrpcFile = originalParseGrpcFile
		runSession = originalRunSession
	})

	globGrpcFiles = func(pattern string) ([]string, error) {
		return []string{filepath.Join(t.TempDir(), "project_grpc.pb.go")}, nil
	}
	parseGrpcFile = func(filePath string) ([]types.ServiceInfo, error) {
		return []types.ServiceInfo{
			{
				FileName:    "project.go",
				HandlerName: "ProjectHandler",
				ServiceName: "ProjectService",
			},
			{
				FileName:    "task.go",
				HandlerName: "TaskHandler",
				ServiceName: "TaskService",
			},
		}, nil
	}
	var captured tui.SessionConfig
	runSession = func(cfg tui.SessionConfig) error {
		captured = cfg
		return nil
	}

	if err := RunHandlerCommand(types.Config{
		OutputDir: "./api/grpc",
		CoreDir:   "./core",
		WireDir:   "./cmd/devopsx",
	}, "./internal/proto/axis/devopsx"); err != nil {
		t.Fatalf("RunHandlerCommand error: %v", err)
	}

	if captured.Title != "Handler Generate" {
		t.Fatalf("expected Handler Generate title, got %q", captured.Title)
	}
	if len(captured.Items) != 2 {
		t.Fatalf("expected all parsed services to be passed through, got %+v", captured.Items)
	}
	if captured.Items[0].ID != "project.go" || captured.Items[1].ID != "task.go" {
		t.Fatalf("unexpected session items: %+v", captured.Items)
	}
	if captured.Run == nil {
		t.Fatalf("expected session runner to be set")
	}
}

func TestRunHandlerCommandSkipsSessionWhenNoParsedServices(t *testing.T) {
	originalGlob := globGrpcFiles
	originalParseGrpcFile := parseGrpcFile
	originalRunSession := runSession
	t.Cleanup(func() {
		globGrpcFiles = originalGlob
		parseGrpcFile = originalParseGrpcFile
		runSession = originalRunSession
	})

	globGrpcFiles = func(pattern string) ([]string, error) {
		return []string{filepath.Join(t.TempDir(), "empty_grpc.pb.go")}, nil
	}
	parseGrpcFile = func(filePath string) ([]types.ServiceInfo, error) {
		return []types.ServiceInfo{}, nil
	}

	sessionCalled := false
	runSession = func(cfg tui.SessionConfig) error {
		sessionCalled = true
		return nil
	}

	if err := RunHandlerCommand(types.Config{
		OutputDir: "./api/grpc",
		CoreDir:   "./core",
		WireDir:   "./cmd/devopsx",
	}, "./internal/proto/axis/devopsx"); err != nil {
		t.Fatalf("RunHandlerCommand error: %v", err)
	}

	if sessionCalled {
		t.Fatalf("expected runSession not to be called when no services are parsed")
	}
}

func TestRunHandlerCommandReturnsErrorWhenAllGrpcParsesFail(t *testing.T) {
	originalGlob := globGrpcFiles
	originalParseGrpcFile := parseGrpcFile
	originalRunSession := runSession
	t.Cleanup(func() {
		globGrpcFiles = originalGlob
		parseGrpcFile = originalParseGrpcFile
		runSession = originalRunSession
	})

	grpcPath := filepath.Join(t.TempDir(), "broken_grpc.pb.go")
	globGrpcFiles = func(pattern string) ([]string, error) {
		return []string{grpcPath}, nil
	}
	parseGrpcFile = func(filePath string) ([]types.ServiceInfo, error) {
		return nil, errors.New("boom")
	}

	sessionCalled := false
	runSession = func(cfg tui.SessionConfig) error {
		sessionCalled = true
		return nil
	}

	err := RunHandlerCommand(types.Config{
		OutputDir: "./api/grpc",
		CoreDir:   "./core",
		WireDir:   "./cmd/devopsx",
	}, "./internal/proto/axis/devopsx")
	if err == nil || !strings.Contains(err.Error(), "解析 grpc 文件失败") || !strings.Contains(err.Error(), grpcPath) {
		t.Fatalf("expected parse failure error, got %v", err)
	}
	if sessionCalled {
		t.Fatalf("expected runSession not to be called when all grpc parsing fails")
	}
}

func TestRunHandlerCommandReturnsErrorWhenSessionFails(t *testing.T) {
	originalGlob := globGrpcFiles
	originalParseGrpcFile := parseGrpcFile
	originalRunSession := runSession
	t.Cleanup(func() {
		globGrpcFiles = originalGlob
		parseGrpcFile = originalParseGrpcFile
		runSession = originalRunSession
	})

	globGrpcFiles = func(pattern string) ([]string, error) {
		return []string{filepath.Join(t.TempDir(), "project_grpc.pb.go")}, nil
	}
	parseGrpcFile = func(filePath string) ([]types.ServiceInfo, error) {
		return []types.ServiceInfo{{FileName: "project.go", HandlerName: "ProjectHandler", ServiceName: "ProjectService"}}, nil
	}
	runSession = func(cfg tui.SessionConfig) error {
		return errors.New("session boom")
	}

	err := RunHandlerCommand(types.Config{
		OutputDir: "./api/grpc",
		CoreDir:   "./core",
		WireDir:   "./cmd/devopsx",
	}, "./internal/proto/axis/devopsx")
	if err == nil || !strings.Contains(err.Error(), "session boom") {
		t.Fatalf("expected session error, got %v", err)
	}
}

func TestRunHandlerCommandReturnsErrorAfterPartialParseFailure(t *testing.T) {
	originalGlob := globGrpcFiles
	originalParseGrpcFile := parseGrpcFile
	originalRunSession := runSession
	t.Cleanup(func() {
		globGrpcFiles = originalGlob
		parseGrpcFile = originalParseGrpcFile
		runSession = originalRunSession
	})

	okPath := filepath.Join(t.TempDir(), "project_grpc.pb.go")
	badPath := filepath.Join(t.TempDir(), "broken_grpc.pb.go")
	globGrpcFiles = func(pattern string) ([]string, error) {
		return []string{okPath, badPath}, nil
	}
	parseGrpcFile = func(filePath string) ([]types.ServiceInfo, error) {
		if filePath == badPath {
			return nil, errors.New("broken")
		}
		return []types.ServiceInfo{{FileName: "project.go", HandlerName: "ProjectHandler", ServiceName: "ProjectService"}}, nil
	}

	sessionCalled := false
	runSession = func(cfg tui.SessionConfig) error {
		sessionCalled = true
		if len(cfg.Items) != 1 || cfg.Items[0].ID != "project.go" {
			t.Fatalf("unexpected session items: %+v", cfg.Items)
		}
		return nil
	}

	err := RunHandlerCommand(types.Config{
		OutputDir: "./api/grpc",
		CoreDir:   "./core",
		WireDir:   "./cmd/devopsx",
	}, "./internal/proto/axis/devopsx")
	if err == nil || !strings.Contains(err.Error(), "部分 grpc 文件解析失败") || !strings.Contains(err.Error(), badPath) {
		t.Fatalf("expected partial parse failure error, got %v", err)
	}
	if !sessionCalled {
		t.Fatalf("expected runSession to execute for successfully parsed services")
	}
}
