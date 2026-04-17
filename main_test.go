package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsVersionAndExitsSuccess(t *testing.T) {
	originalVersion := version
	t.Cleanup(func() {
		version = originalVersion
	})

	version = "dev"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"gen-handler", "-version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if got := strings.TrimSpace(stdout.String()); got != "dev" {
		t.Fatalf("expected version output dev, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunPrintsInjectedVersion(t *testing.T) {
	originalVersion := version
	t.Cleanup(func() {
		version = originalVersion
	})

	version = "v1.2.3"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"gen-handler", "-version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if got := strings.TrimSpace(stdout.String()); got != "v1.2.3" {
		t.Fatalf("expected version output v1.2.3, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}
