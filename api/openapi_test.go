package api_test

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestOpenAPIParses(t *testing.T) {
	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(content, &document); err != nil {
		t.Fatalf("parse OpenAPI contract: %v", err)
	}
	if document["openapi"] != "3.1.0" {
		t.Fatalf("openapi = %v, want 3.1.0", document["openapi"])
	}
	paths, ok := document["paths"].(map[string]any)
	if !ok || len(paths) != 14 {
		t.Fatalf("paths = %d, want 14 Account paths", len(paths))
	}
}
