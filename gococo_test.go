package gococo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aleksadvaisly/gococo-reporting"
)

func TestGenerateHTMLFromFiles(t *testing.T) {
	html, err := gococo.GenerateHTMLFromFiles([]string{
		filepath.Join("testdata", "godog-report.json"),
	}, gococo.Options{
		Title:       "Acceptance Report",
		GeneratedAt: time.Date(2026, 3, 18, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("GenerateHTMLFromFiles returned error: %v", err)
	}

	output := string(html)
	assertContains(t, output, "<title>Acceptance Report</title>")
	assertContains(t, output, "Single-file cucumber report")
	assertContains(t, output, "Checkout flow")
	assertContains(t, output, "Search catalogue")
	assertContains(t, output, "card declined")
	assertContains(t, output, "grid-template-columns")
	assertContains(t, output, "border-radius")
	assertNotContains(t, output, "<table")
}

func TestWriteHTMLReport(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "nested", "report.html")

	err := gococo.WriteHTMLReport(outputPath, []string{
		filepath.Join("testdata", "godog-report.json"),
	}, gococo.Options{Title: "CLI Smoke"})
	if err != nil {
		t.Fatalf("WriteHTMLReport returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	assertContains(t, string(data), "CLI Smoke")
}

func TestGenerateHTMLFromGoTestJSON(t *testing.T) {
	html, err := gococo.GenerateHTMLFromFiles([]string{
		filepath.Join("testdata", "gobdd-report.jsonl"),
	}, gococo.Options{
		Title: "BDD Report",
	})
	if err != nil {
		t.Fatalf("GenerateHTMLFromFiles returned error: %v", err)
	}

	output := string(html)
	assertContains(t, output, "CSV processing in memory")
	assertContains(t, output, "Sum order amounts")
	assertContains(t, output, "Given an orders dataset")
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected output to contain %q", want)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("expected output not to contain %q", want)
	}
}
