package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSwaggerPathsCoverRegisteredRoutes(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine caller path")
	}

	apiRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	swaggerPath := filepath.Join(apiRoot, "docs", "swagger.json")
	raw, err := os.ReadFile(swaggerPath)
	if err != nil {
		t.Fatalf("failed reading swagger file: %v", err)
	}

	var document struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("failed parsing swagger json: %v", err)
	}

	expected := []string{
		"/import/csv",
		"/accounts",
		"/vendors",
		"/marketdata/listing",
		"/marketdata/listings",
		"/marketdata/listings/search",
		"/portfolio/rebuild",
		"/portfolio/snapshots",
		"/portfolio/positions",
		"/portfolio/transactions",
		"/portfolio/transactions/manual",
		"/marketdata/dailies",
		"/marketdata/dailies/upload",
		"/marketdata/dailies/uploads/{upload_id}",
		"/cashflow/transactions",
		"/cashflow/transactions/manual",
		"/cashflow/analytics/monthly",
		"/cashflow/analytics/tags",
		"/cashflow/transactions/tag",
		"/cashflow/transactions/tag/selection",
		"/cashflow/transactions/tag/filter",
		"/cashflow/transactions/ignore/selection",
		"/cashflow/transactions/ignore/filter",
		"/assets/classes",
		"/assets/classes/{class_id}",
		"/assets/items",
		"/assets/items/worth/set",
		"/assets/items/worth/adjust",
		"/health",
	}

	for _, path := range expected {
		if _, exists := document.Paths[path]; !exists {
			t.Fatalf("swagger path missing: %s", path)
		}
	}
}
