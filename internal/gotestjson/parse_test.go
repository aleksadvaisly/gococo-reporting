package gotestjson

import (
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	features, err := ParseFile(filepath.Join("..", "..", "testdata", "gobdd-report.jsonl"))
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(features))
	}

	feature := features[0]
	if feature.Name != "CSV processing in memory" {
		t.Fatalf("unexpected feature name: %q", feature.Name)
	}
	if len(feature.Elements) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(feature.Elements))
	}

	firstStep := feature.Elements[0].Steps[0]
	if firstStep.Keyword != "Given" {
		t.Fatalf("expected Given keyword, got %q", firstStep.Keyword)
	}
	if firstStep.Result.Status != "passed" {
		t.Fatalf("expected passed status, got %q", firstStep.Result.Status)
	}
}
