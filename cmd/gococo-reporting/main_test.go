package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesReport(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "report.html")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{
		"-input", filepath.Join("..", "..", "testdata", "godog-report.json"),
		"-output", outputPath,
		"-title", "Command Report",
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run returned exit code %d, stderr=%s", code, stderr.String())
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(data), "Command Report") {
		t.Fatalf("expected generated report to include title")
	}
}

func TestRunFailsWithoutInput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "no input files provided") {
		t.Fatalf("expected missing input error, got %q", stderr.String())
	}
}
